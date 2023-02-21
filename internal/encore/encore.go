package encore

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	//"os"
	//"os/signal"
	//"syscall"

	"github.com/gin-gonic/gin"
	//"github.com/superseriousbusiness/gotosocial/cmd/gotosocial/action"
	"github.com/superseriousbusiness/gotosocial/internal/api"
	apiutil "github.com/superseriousbusiness/gotosocial/internal/api/util"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/middleware"

	"github.com/superseriousbusiness/gotosocial/internal/concurrency"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/db/bundb"
	"github.com/superseriousbusiness/gotosocial/internal/email"
	"github.com/superseriousbusiness/gotosocial/internal/federation"
	"github.com/superseriousbusiness/gotosocial/internal/federation/federatingdb"
	//"github.com/superseriousbusiness/gotosocial/internal/gotosocial"
	"github.com/superseriousbusiness/gotosocial/internal/httpclient"
	//"github.com/superseriousbusiness/gotosocial/internal/log"
	"github.com/superseriousbusiness/gotosocial/internal/media"
	"github.com/superseriousbusiness/gotosocial/internal/messages"
	"github.com/superseriousbusiness/gotosocial/internal/oauth"
	"github.com/superseriousbusiness/gotosocial/internal/oidc"
	"github.com/superseriousbusiness/gotosocial/internal/processing"
	"github.com/superseriousbusiness/gotosocial/internal/router"
	"github.com/superseriousbusiness/gotosocial/internal/state"
	gtsstorage "github.com/superseriousbusiness/gotosocial/internal/storage"
	"github.com/superseriousbusiness/gotosocial/internal/transport"
	"github.com/superseriousbusiness/gotosocial/internal/typeutils"
	"github.com/superseriousbusiness/gotosocial/internal/web"

	"encore.dev/storage/sqldb"
	"encore.dev/rlog"
)

// Start creates and starts a gotosocial server
//encore: service
type Service struct {
	engine *gin.Engine
}

var encoreDB *sqldb.Database = sqldb.Named("encore")
var encoreRouter *Service

func initService() (*Service, error) {
	ctx := context.WithValue(context.Background(), "encoreDB", encoreDB)

	var state state.State

	// Initialize caches
	state.Caches.Init()
	state.Caches.Start()
	defer state.Caches.Stop()

	// Open connection to the database
	dbService, err := bundb.NewBunDBService(ctx, &state)
	if err != nil {
		return nil, fmt.Errorf("error creating dbservice: %s", err)
	}

	// Set the state DB connection
	state.DB = dbService

	if err := dbService.CreateInstanceAccount(ctx); err != nil {
		return nil, fmt.Errorf("error creating instance account: %s", err)
	}

	if err := dbService.CreateInstanceInstance(ctx); err != nil {
		return nil, fmt.Errorf("error creating instance instance: %s", err)
	}

	// Create the client API and federator worker pools
	// NOTE: these MUST NOT be used until they are passed to the
	// processor and it is started. The reason being that the processor
	// sets the Worker process functions and start the underlying pools
	clientWorker := concurrency.NewWorkerPool[messages.FromClientAPI](-1, -1)
	fedWorker := concurrency.NewWorkerPool[messages.FromFederator](-1, -1)

	federatingDB := federatingdb.New(dbService, fedWorker)

	// build converters and util
	typeConverter := typeutils.NewConverter(dbService)

	// Open the storage backend
	storage, err := gtsstorage.AutoConfig()
	if err != nil {
		return nil, fmt.Errorf("error creating storage backend: %w", err)
	}

	// Build HTTP client (TODO: add configurables here)
	client := httpclient.New(httpclient.Config{})

	// build backend handlers
	mediaManager, err := media.NewManager(dbService, storage)
	if err != nil {
		return nil, fmt.Errorf("error creating media manager: %s", err)
	}
	oauthServer := oauth.New(ctx, dbService)
	transportController := transport.NewController(dbService, federatingDB, &federation.Clock{}, client)
	federator := federation.NewFederator(dbService, federatingDB, transportController, typeConverter, mediaManager)

	// decide whether to create a noop email sender (won't send emails) or a real one
	var emailSender email.Sender
	if smtpHost := config.GetSMTPHost(); smtpHost != "" {
		// host is defined so create a proper sender
		emailSender, err = email.NewSender()
		if err != nil {
			return nil, fmt.Errorf("error creating email sender: %s", err)
		}
	} else {
		// no host is defined so create a noop sender
		emailSender, err = email.NewNoopSender(nil)
		if err != nil {
			return nil, fmt.Errorf("error creating noop email sender: %s", err)
		}
	}

	// create the message processor using the other services we've created so far
	processor := processing.NewProcessor(typeConverter, federator, oauthServer, mediaManager, storage, dbService, emailSender, clientWorker, fedWorker)
	if err := processor.Start(); err != nil {
		return nil, fmt.Errorf("error creating processor: %s", err)
	}

	/*
		HTTP router initialization
	*/

	router_, err := router.NewRouter(ctx)
	if err != nil {
		return nil, fmt.Errorf("error creating router: %s", err)
	}

	// attach global middlewares which are used for every request
	router_.AttachGlobalMiddleware(
		middleware.Logger(),
		middleware.UserAgent(),
		middleware.CORS(),
		middleware.ExtraHeaders(),
	)

	// attach global no route / 404 handler to the router
	router_.AttachNoRouteHandler(func(c *gin.Context) {
		apiutil.ErrorHandler(c, gtserror.NewErrorNotFound(errors.New(http.StatusText(http.StatusNotFound))), processor.InstanceGetV1)
	})

	// build router modules
	var idp oidc.IDP
	if config.GetOIDCEnabled() {
		idp, err = oidc.NewIDP(ctx)
		if err != nil {
			return nil, fmt.Errorf("error creating oidc idp: %w", err)
		}
	}

	routerSession, err := dbService.GetSession(ctx)
	if err != nil {
		return nil, fmt.Errorf("error retrieving router session for session middleware: %w", err)
	}

	sessionName, err := middleware.SessionName()
	if err != nil {
		return nil, fmt.Errorf("error generating session name for session middleware: %w", err)
	}

	var (
		authModule        = api.NewAuth(dbService, processor, idp, routerSession, sessionName) // auth/oauth paths
		clientModule      = api.NewClient(dbService, processor)                                // api client endpoints
		fileserverModule  = api.NewFileserver(processor)                                       // fileserver endpoints
		wellKnownModule   = api.NewWellKnown(processor)                                        // .well-known endpoints
		nodeInfoModule    = api.NewNodeInfo(processor)                                         // nodeinfo endpoint
		activityPubModule = api.NewActivityPub(dbService, processor)                           // ActivityPub endpoints
		webModule         = web.New(processor)                                                 // web pages + user profiles + settings panels etc
	)

	// create required middleware
	// rate limiting
	limit := config.GetAdvancedRateLimitRequests()
	clLimit := middleware.RateLimit(limit)  // client api
	s2sLimit := middleware.RateLimit(limit) // server-to-server (AP)
	fsLimit := middleware.RateLimit(limit)  // fileserver / web templates

	// throttling
	cpuMultiplier := config.GetAdvancedThrottlingMultiplier()
	clThrottle := middleware.Throttle(cpuMultiplier)  // client api
	s2sThrottle := middleware.Throttle(cpuMultiplier) // server-to-server (AP)
	fsThrottle := middleware.Throttle(cpuMultiplier)  // fileserver / web templates

	gzip := middleware.Gzip() // applied to all except fileserver

	// these should be routed in order;
	// apply throttling *after* rate limiting
	authModule.Route(router_, clLimit, clThrottle, gzip)
	clientModule.Route(router_, clLimit, clThrottle, gzip)
	fileserverModule.Route(router_, fsLimit, fsThrottle)
	wellKnownModule.Route(router_, gzip, s2sLimit, s2sThrottle)
	nodeInfoModule.Route(router_, s2sLimit, s2sThrottle, gzip)
	activityPubModule.Route(router_, s2sLimit, s2sThrottle, gzip)
	webModule.Route(router_, fsLimit, fsThrottle, gzip)

	return &Service{engine: router_.(*router.RouterType).Engine}, nil
}

//encore:api public raw path=/*gtsPath
func (s *Service) gtsMain(w http.ResponseWriter, req *http.Request) {
	rlog.Info(
		"raw header",
		"host", req.Host,
		"get host header", req.Header.Get("host"),
		"Host header", req.Header["Host"],
		"host header", req.Header["host"],
	)
	s.engine.ServeHTTP(w, req)
}

/*
var Start action.GTSAction = func(ctx context.Context) error {

	gts, err := gotosocial.NewServer(dbService, router, federator, mediaManager)
	if err != nil {
		return fmt.Errorf("error creating gotosocial service: %s", err)
	}

	if err := gts.Start(ctx); err != nil {
		return fmt.Errorf("error starting gotosocial service: %s", err)
	}

	// perform initial media prune in case value of MediaRemoteCacheDays changed
	if err := processor.AdminMediaPrune(ctx, config.GetMediaRemoteCacheDays()); err != nil {
		return fmt.Errorf("error during initial media prune: %s", err)
	}

	// catch shutdown signals from the operating system
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	sig := <-sigs // block until signal received
	log.Infof("received signal %s, shutting down", sig)

	// close down all running services in order
	if err := gts.Stop(ctx); err != nil {
		return fmt.Errorf("error closing gotosocial service: %s", err)
	}

	log.Info("done! exiting...")
	return nil
}
*/
