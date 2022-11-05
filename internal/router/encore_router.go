package router

import (
	"context"
	"fmt"
	"net"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"golang.org/x/crypto/acme/autocert"
)

type RouterType router

// New returns a new Router with the specified configuration.
//
// The given DB is only used in the New function for parsing config values, and is not otherwise
// pinned to the router.
func NewRouter(ctx context.Context, db db.DB) (Router, error) {
	gin.SetMode(gin.ReleaseMode)

	// create the actual engine here -- this is the core request routing handler for gts
	engine := gin.New()
	engine.Use(loggingMiddleware)

	// 8 MiB
	engine.MaxMultipartMemory = 8 << 20

	// set up IP forwarding via x-forward-* headers.
	trustedProxies := config.GetTrustedProxies()
	if err := engine.SetTrustedProxies(trustedProxies); err != nil {
		return nil, err
	}

	// enable cors on the engine
	if err := useCors(engine); err != nil {
		return nil, err
	}

	// enable gzip compression on the engine
	if err := useGzip(engine); err != nil {
		return nil, err
	}

	// enable session store middleware on the engine
	if err := useSession(ctx, db, engine); err != nil {
		return nil, err
	}

	// set template functions
	LoadTemplateFunctions(engine)

	// load templates onto the engine
	if err := LoadTemplates(engine); err != nil {
		return nil, err
	}

	// use the passed-in command context as the base context for the server,
	// since we'll never want the server to live past the command anyway
	baseCtx := func(_ net.Listener) context.Context {
		return ctx
	}

	bindAddress := config.GetBindAddress()
	port := config.GetPort()
	addr := fmt.Sprintf("%s:%d", bindAddress, port)

	s := &http.Server{
		Addr:              addr,
		Handler:           engine, // use gin engine as handler
		ReadTimeout:       readTimeout,
		ReadHeaderTimeout: readHeaderTimeout,
		WriteTimeout:      writeTimeout,
		IdleTimeout:       idleTimeout,
		BaseContext:       baseCtx,
	}

	// We need to spawn the underlying server slightly differently depending on whether lets encrypt is enabled or not.
	// In either case, the gin engine will still be used for routing requests.
	leEnabled := config.GetLetsEncryptEnabled()

	var m *autocert.Manager
	if leEnabled {
		// le IS enabled, so roll up an autocert manager for handling letsencrypt requests
		host := config.GetHost()
		leCertDir := config.GetLetsEncryptCertDir()
		leEmailAddress := config.GetLetsEncryptEmailAddress()
		m = &autocert.Manager{
			Prompt:     autocert.AcceptTOS,
			HostPolicy: autocert.HostWhitelist(host),
			Cache:      autocert.DirCache(leCertDir),
			Email:      leEmailAddress,
		}
		s.TLSConfig = m.TLSConfig()
	}

	return &router{
		engine:      engine,
		srv:         s,
		certManager: m,
	}, nil
}
