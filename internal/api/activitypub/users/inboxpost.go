/*
   GoToSocial
   Copyright (C) 2021-2023 GoToSocial Authors admin@gotosocial.org

   This program is free software: you can redistribute it and/or modify
   it under the terms of the GNU Affero General Public License as published by
   the Free Software Foundation, either version 3 of the License, or
   (at your option) any later version.

   This program is distributed in the hope that it will be useful,
   but WITHOUT ANY WARRANTY; without even the implied warranty of
   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
   GNU Affero General Public License for more details.

   You should have received a copy of the GNU Affero General Public License
   along with this program.  If not, see <http://www.gnu.org/licenses/>.
*/

package users

import (
	"errors"
	"strings"

	"github.com/gin-gonic/gin"
	apiutil "github.com/superseriousbusiness/gotosocial/internal/api/util"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror" //nolint:typecheck

	"encore.dev/rlog"
)

// InboxPOSTHandler deals with incoming POST requests to an actor's inbox.
// Eg., POST to https://example.org/users/whatever/inbox.
func (m *Module) InboxPOSTHandler(c *gin.Context) {
	// usernames on our instance are always lowercase
	requestedUsername := strings.ToLower(c.Param(UsernameKey))
	if requestedUsername == "" {
		err := errors.New("no username specified in request")
		rlog.Error("no username specified in request", "usernameKey", c.Param(UsernameKey))
		apiutil.ErrorHandler(c, gtserror.NewErrorBadRequest(err, err.Error()), m.processor.InstanceGetV1)
		return
	}

	if posted, err := m.processor.InboxPost(apiutil.TransferSignatureContext(c), c.Writer, c.Request); err != nil {
		if withCode, ok := err.(gtserror.WithCode); ok {
			rlog.Error("api/inboxpost A", "err", err, "withCode", withCode)
			apiutil.ErrorHandler(c, withCode, m.processor.InstanceGetV1)
		} else {
			rlog.Error("api/inboxpost B", "err", err)
			apiutil.ErrorHandler(c, gtserror.NewErrorBadRequest(err, err.Error()), m.processor.InstanceGetV1)
		}
	} else if !posted {
		err := errors.New("unable to process request")
		rlog.Error("api/inboxpost C", "err", err, "posted", posted)
		apiutil.ErrorHandler(c, gtserror.NewErrorBadRequest(err, err.Error()), m.processor.InstanceGetV1)
	}
}
