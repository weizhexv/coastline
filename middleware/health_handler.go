package middleware

import (
	"coastline/tlog"
	"github.com/gin-gonic/gin"
	"net/http"
)

const hPath = "/health"

func HealthHandler() gin.HandlerFunc {
	return func(gc *gin.Context) {
		if gc.Request.RequestURI == hPath {
			tlog.Entry().Debugln("Health check from ELB")
			gc.AbortWithStatus(http.StatusOK)
		}
	}
}
