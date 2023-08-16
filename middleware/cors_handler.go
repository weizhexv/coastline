package middleware

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
)

func CorsHandler() gin.HandlerFunc {
	return func(gc *gin.Context) {
		origin := gc.Request.Header.Get("Origin")
		if len(origin) == 0 {
			// request is not a CORS request
			return
		}
		host := gc.Request.Host

		if origin == "http://"+host || origin == "https://"+host {
			// request is not a CORS request but have origin header.
			// for example, use fetch api
			return
		}

		gc.Header("Access-Control-Allow-Origin", origin)
		gc.Header("Access-Control-Allow-Credentials", "true")
		gc.Header("Access-Control-Allow-Headers", "DNT,User-Agent,X-Requested-With,If-Modified-Since,Cache-Control,Content-Type,Range,platform,token,os,trace-id,app-version,s-token,run-as,lang")
		gc.Header("Access-Control-Allow-Methods", strings.Join([]string{http.MethodPost, http.MethodGet, http.MethodOptions, http.MethodHead}, ","))

		if gc.Request.Method == "OPTIONS" {
			gc.AbortWithStatus(http.StatusNoContent)
			return
		}
	}
}
