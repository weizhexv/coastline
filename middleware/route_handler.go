package middleware

import (
	"coastline/ctx"
	"coastline/route"
	"dghire.com/libs/go-monitor"
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
)

const apiPrefix = "/api"

func RouteHandler() gin.HandlerFunc {
	return func(gc *gin.Context) {
		c := ctx.DetachFrom(gc)

		c.Infof("client request url: %s", gc.Request.URL.String())
		//trim path if start with /api
		path := gc.Request.URL.Path
		if strings.HasPrefix(path, apiPrefix) {
			path = strings.TrimPrefix(path, apiPrefix)
		}

		//find target route
		rt, ok := route.Lookup(path)
		if !ok {
			c.Warnln("not found path:", path)
			gc.AbortWithStatus(http.StatusNotFound)
			return
		}
		c.Infof("found route %s\n", rt)

		//save route to ctx
		c.AddRoute(rt)

		//count monitor
		monitorRequest(c.Route.UpstreamUrl)
	}
}

func monitorRequest(upstreamUrl string) {
	monitor.HttpServerCounter(upstreamUrl)
	monitor.HttpClientCounter(upstreamUrl)
}
