package middleware

import (
	"coastline/consts"
	"coastline/ctx"
	"coastline/upstream"
	"github.com/gin-gonic/gin"
	"strconv"
	"strings"
)

const (
	runAs = "runAs"
)

func RunAsHandler() gin.HandlerFunc {
	return func(gc *gin.Context) {
		c := ctx.DetachFrom(gc)
		runAsUid := gc.Query(runAs)
		token := gc.Query(consts.HeaderToken)

		if len(runAsUid) == 0 || len(token) == 0 {
			c.RunAs = false
			return
		} else {
			c.RunAs = true
		}

		auth, code := upstream.Auth(token, c)
		if code != upstream.OK {
			c.Warnln("auth err code:", code)
			if code == upstream.BadUpstream {
				upstream.NewResult().RenderSysErr(gc)
			} else {
				upstream.NewResult().RenderGatewayErr(gc)
			}
			return
		}

		c.Infof("run as auth result: %s\n", auth)
		if !auth.Success {
			c.Errorln("run as auth failed")
			upstream.NewResult().RenderAuthErr(gc)
			return
		}

		dm := auth.Data.TokenInfo.Domain
		if !strings.EqualFold(dm, consts.DomainPlatform) {
			c.Errorln("run as not a platform user:", dm)
			upstream.NewResult().RenderForbiddenErr(gc)
			return
		}

		_, err := strconv.ParseInt(runAsUid, 10, 64)
		if err != nil {
			c.Errorln("run as uid not a number:", dm)
			upstream.NewResult().RenderForbiddenErr(gc)
			return
		}

		opId := auth.Data.UserInfo.Uid
		c.Infof("run as uid:[%s], opid:[%d]", runAsUid, opId)

		info, ok := upstream.GetUserInfo(runAsUid, c)
		if ok != upstream.OK {
			c.Errorln("get user info err code:", code)
			if code == upstream.BadUpstream {
				upstream.NewResult().RenderSysErr(gc)
			} else {
				upstream.NewResult().RenderGatewayErr(gc)
			}
			return
		}

		c.Infof("get runAs user info: %v", info)
		if !info.Success {
			upstream.NewResult().RenderAuthErr(gc)
			return
		}

		c.AddHeader(consts.HeaderDomain, info.Data.Domain)
		c.AddHeader(consts.HeaderUid, strconv.Itoa(info.Data.Uid))
		c.AddHeader(consts.HeaderDomainId, strconv.Itoa(info.Data.DomainId))
		c.AddHeader(consts.HeaderRole, info.Data.Role)
		c.AddHeader(consts.HeaderOpId, strconv.Itoa(opId))

		c.Debugf("ctx in run as handler: %s\n", c)
	}
}
