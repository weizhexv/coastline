package upstream

import (
	"coastline/errs"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
)

type Result[T any] struct {
	Success bool   `json:"success"`
	Code    int    `json:"code"`
	Message string `json:"message,omitempty"`
	Data    T      `json:"data,omitempty"`
}

type AuthInfo struct {
	TokenInfo *TokenInfo `json:"tokenInfo,omitempty"`
	UserInfo  *UserInfo  `json:"userInfo,omitempty"`
}

type TokenInfo struct {
	Uid      int    `json:"uid,omitempty"`
	Domain   string `json:"domain,omitempty"`
	Platform string `json:"platform,omitempty"`
}

type UserInfo struct {
	Uid      int    `json:"uid,omitempty"`
	Name     string `json:"name,omitempty"`
	Mobile   string `json:"mobile,omitempty"`
	Email    string `json:"email,omitempty"`
	Domain   string `json:"domain,omitempty"`
	DomainId int    `json:"domainId,omitempty"`
	Role     string `json:"role,omitempty"`
	Enabled  bool   `json:"enabled,omitempty"`
	Deleted  bool   `json:"deleted,omitempty"`
	Verified bool   `json:"verified,omitempty"`
}

func NewResult() *Result[any] {
	return new(Result[any])
}

func NewAuthInfo() *Result[AuthInfo] {
	return new(Result[AuthInfo])
}

func NewUserInfo() *Result[UserInfo] {
	return new(Result[UserInfo])
}

func (r *Result[T]) RenderSysErr(gc *gin.Context) {
	r.RenderErr(gc, &errs.ErrSystem)
}

func (r *Result[T]) RenderAuthErr(gc *gin.Context) {
	r.RenderErr(gc, &errs.ErrAuth)
}

func (r *Result[T]) RenderGatewayErr(gc *gin.Context) {
	r.RenderErr(gc, &errs.ErrGateway)
}

func (r *Result[T]) RenderForbiddenErr(gc *gin.Context) {
	r.RenderErr(gc, &errs.ErrForbidden)
}

func (r *Result[T]) RenderErr(gc *gin.Context, e *errs.ApiErr) {
	r.Success = false
	r.Code = e.Code
	r.Message = e.Message

	gc.AbortWithStatusJSON(http.StatusOK, r)
}

func (r *Result[T]) String() string {
	if r == nil {
		return "<nil>"
	}
	if bs, err := json.Marshal(&r); err != nil {
		fmt.Printf("Result[T] String() err: %s", err.Error())
		return err.Error()
	} else {
		return string(bs)
	}
}
