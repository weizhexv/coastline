package middleware

import (
	"coastline/ctx"
	"coastline/safeutil"
	"coastline/upstream"
	"dghire.com/libs/go-monitor"
	"github.com/gin-gonic/gin"
	"net/http"
)

const KB = 1 << 10

func UpstreamHandler() gin.HandlerFunc {
	return func(gc *gin.Context) {
		c := ctx.DetachFrom(gc)
		c.RefreshUpstreamStartAt()

		resp, code := upstream.Invoke(gc)
		c.Infof("invoke upstream resp code: %d\n", code)

		if code == upstream.OK {
			ok(resp, gc)
		} else {
			if code == upstream.BadUpstream {
				upstream.NewResult().RenderSysErr(gc)
			} else {
				upstream.NewResult().RenderGatewayErr(gc)
			}
			monitorFailureResp(ctx.DetachFrom(gc))
		}
	}
}

func ok(resp *http.Response, gc *gin.Context) {
	c := ctx.DetachFrom(gc)

	//set status
	gc.Status(resp.StatusCode)

	//set headers
	if len(resp.Header) > 0 {
		for k, v := range resp.Header {
			gc.Header(k, v[0])
		}
	}

	//abort HTTP code != 200
	if resp.StatusCode != http.StatusOK {
		c.Warnf("abort with bad upstream status: %d", resp.StatusCode)
		gc.AbortWithStatus(resp.StatusCode)
		return
	}

	//set body
	bs, err := safeutil.Read(resp)
	if err != nil {
		c.Errorln("read body err:", err)
		gc.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	if len(bs) == 0 {
		c.Warnf("blank resp body")
		return
	}

	_, err = gc.Writer.Write(bs)
	if err != nil {
		c.Errorln("write resp err:", err)
		gc.AbortWithStatus(http.StatusBadGateway)
		return
	}

	//monitor success
	monitorSuccessResp(c, bs)
}

func monitorSuccessResp(c *ctx.Ctx, bs []byte) {
	defer func() {
		if r := recover(); r != nil {
			c.Errorln("monitor panic: ", r)
		}
	}()

	clientCost := c.CurrentClientCost()
	serverCost := c.CurrentServerCost()
	monitor.HttpClientDuration(c.Route.UpstreamUrl, "false", clientCost)
	monitor.HttpServerDuration(c.Route.UpstreamUrl, "false", serverCost)

	if len(bs) <= 4*KB {
		c.Infof("resp successfully, body:%s, client cost[%d]ms, server cost[%d]ms", bs, clientCost, serverCost)
	} else {
		c.Infof("resp successfully, body len:%d, cost[%d]ms, server cost[%d]ms", len(bs), clientCost, serverCost)
	}
}

func monitorFailureResp(c *ctx.Ctx) {
	defer func() {
		if r := recover(); r != nil {
			c.Errorln("monitor panic: ", r)
		}
	}()

	monitor.HttpClientDuration(c.Route.UpstreamUrl, "true", c.CurrentClientCost())
}
