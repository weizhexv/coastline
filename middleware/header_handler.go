package middleware

import (
	"coastline/consts"
	"coastline/ctx"
	"github.com/gin-gonic/gin"
)

func HeaderHandler() gin.HandlerFunc {
	return func(gc *gin.Context) {
		c := ctx.DetachFrom(gc)

		for k, v := range gc.Request.Header {
			if len(v) > 0 {
				c.AddHeader(k, v[0])
			}
		}

		c.AddHeader(consts.HeaderRemoteIP, gc.ClientIP())
		c.Infof("client headers are:%v/n", gc.Request.Header)
	}
}
