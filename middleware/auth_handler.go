package middleware

import (
	"coastline/consts"
	"coastline/ctx"
	"coastline/route"
	"coastline/upstream"
	"github.com/gin-gonic/gin"
	"strconv"
	"strings"
)

func AuthHandler() gin.HandlerFunc {
	return func(gc *gin.Context) {
		c := ctx.DetachFrom(gc)

		if c.SkipLogin() {
			c.Infoln("non login request skip auth")
			return
		}

		token, ok := parseToken(gc)
		if !ok {
			c.Errorln("token not found")
			upstream.NewResult().RenderAuthErr(gc)
			return
		}
		c.Infof("parse token: %s\n", token)

		resp, code := upstream.Auth(token, c)
		c.Infof("auth resp code:%d\n", code)

		if code != upstream.OK {
			if code == upstream.BadUpstream {
				upstream.NewResult().RenderSysErr(gc)
			} else {
				upstream.NewResult().RenderGatewayErr(gc)
			}
			return
		}

		c.Infof("auth result: %s\n", resp)
		if !resp.Success {
			c.Warnln("user need login")
			upstream.NewResult().RenderAuthErr(gc)
			return
		}

		if forbidden(c.Route, resp.Data.UserInfo) {
			c.Errorln("request is forbidden")
			upstream.NewResult().RenderForbiddenErr(gc)
			return
		}

		c.AddHeader(consts.HeaderDomain, resp.Data.TokenInfo.Domain)
		c.AddHeader(consts.HeaderUid, strconv.Itoa(resp.Data.TokenInfo.Uid))
		c.AddHeader(consts.HeaderDomainId, strconv.Itoa(resp.Data.UserInfo.DomainId))
		c.AddHeader(consts.HeaderRole, resp.Data.UserInfo.Role)

		c.Debugf("ctx in auth handler: %s\n", c)
	}
}

func forbidden(route *route.Route, info *upstream.UserInfo) bool {
	return strings.HasPrefix(route.Path, "/a") && !strings.EqualFold(info.Domain, consts.DomainPlatform)
}

func parseToken(gc *gin.Context) (string, bool) {
	token, _ := gc.Cookie(consts.HeaderToken)
	if len(token) == 0 {
		token = gc.GetHeader(consts.HeaderToken)
	}
	if len(token) == 0 {
		return "", false
	}
	return token, true
}
