package middleware

import (
	"coastline/consts"
	"coastline/ctx"
	"coastline/errs"
	"coastline/tlog"
	"coastline/upstream"
	"coastline/vconfig"
	"dghire.com/libs/go-monitor"
	"errors"
	"github.com/gin-gonic/gin"
	"net/http"
)

func ErrorHandler() gin.HandlerFunc {
	return func(gc *gin.Context) {
		gc.Header(consts.HeaderPoweredBy, vconfig.AppName())
		gc.Next()

		err := gc.Errors.Last()
		if err == nil {
			return
		}

		monitorErrorResp(gc)

		var apiErr *errs.ApiErr
		if errors.As(err, &apiErr) {
			upstream.NewResult().RenderErr(gc, apiErr)
		}
		gc.AbortWithStatus(http.StatusBadGateway)
	}
}

func monitorErrorResp(gc *gin.Context) {
	c := ctx.DetachFrom(gc)
	if c == nil {
		tlog.Entry().Errorln("response failed:", gc.Errors.JSON())
	} else {
		monitor.HttpServerDuration(c.Route.UpstreamUrl, "true", c.CurrentServerCost())
		c.Errorln("response failed:", gc.Errors.JSON())
	}
}
