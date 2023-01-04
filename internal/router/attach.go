/*
   GoToSocial
   Copyright (C) 2021-2022 GoToSocial Authors admin@gotosocial.org

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

package router

import "github.com/gin-gonic/gin"

func (r *RouterType) AttachGlobalMiddleware(handlers ...gin.HandlerFunc) gin.IRoutes {
	return r.Engine.Use(handlers...)
}

func (r *RouterType) AttachNoRouteHandler(handler gin.HandlerFunc) {
	r.Engine.NoRoute(handler)
}

func (r *RouterType) AttachGroup(relativePath string, handlers ...gin.HandlerFunc) *gin.RouterGroup {
	return r.Engine.Group(relativePath, handlers...)
}

func (r *RouterType) AttachHandler(method string, path string, handler gin.HandlerFunc) {
	r.Engine.Handle(method, path, handler)
}
