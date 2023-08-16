package main

import (
	"coastline/middleware"
	"coastline/route"
	"coastline/vconfig"
	"dghire.com/libs/go-monitor"
	"github.com/gin-gonic/gin"
	"log"
)

var app *gin.Engine

func run() {
	app = gin.Default()
	app.MaxMultipartMemory = 8 << 20

	app.Use(middleware.ErrorHandler())
	app.Use(middleware.CorsHandler())
	app.Use(middleware.HealthHandler())
	app.Use(middleware.TraceHandler())
	app.Use(middleware.LoggerHandler())
	app.Use(middleware.HeaderHandler())
	app.Use(middleware.RouteHandler())
	app.Use(middleware.RunAsHandler())
	app.Use(middleware.AuthHandler())
	app.Use(middleware.UpstreamHandler())

	log.Fatalln(app.Run(":" + vconfig.ServerPort()))
}

func main() {
	//拉取路由表
	go route.StartPull()
	//启动监控
	monitor.Start(vconfig.AppName(), vconfig.MonitorPort())
	//启动web服务
	run()
}
