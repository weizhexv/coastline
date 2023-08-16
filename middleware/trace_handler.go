package middleware

import (
	"coastline/consts"
	"coastline/ctx"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func TraceHandler() gin.HandlerFunc {
	return func(gc *gin.Context) {
		//get trace id first
		traceId := getTraceId(gc)
		//init api ctx
		c := ctx.New(traceId)
		//attach to fiberCtx
		c.AttachTo(gc)
		//set header trace id
		c.AddHeader(consts.HeaderTraceId, traceId)

		c.Infof("request trace id:[%s]", traceId)
	}
}

func getTraceId(gc *gin.Context) string {
	var traceId string

	bs := gc.GetHeader(consts.HeaderTraceId)
	if len(bs) == 0 {
		traceId = uuid.NewString()
	} else {
		traceId = bs
	}

	return traceId
}
