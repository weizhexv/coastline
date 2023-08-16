package ctx

import (
	"coastline/route"
	"coastline/tlog"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"time"
)

const Key = "Key"

type logEntry struct {
	*logrus.Entry
}

type Ctx struct {
	RunAs           bool               `json:"runAs"`
	Headers         map[string]*Header `json:"headers"`
	Route           *route.Route       `json:"route"`
	StartAt         time.Time          `json:"startAt"`
	UpstreamStartAt time.Time          `json:"UpstreamStartAt"`
	TraceId         string             `json:"traceId"`
	logEntry        `json:"-"`
}

type Header struct {
	Key string `json:"key"`
	Val string `json:"val"`
}

func New(traceId string) *Ctx {
	return &Ctx{
		RunAs:           false,
		Headers:         make(map[string]*Header),
		Route:           nil,
		StartAt:         time.Now(),
		UpstreamStartAt: time.Now(),
		TraceId:         traceId,
		logEntry:        logEntry{tlog.NewEntry(traceId)},
	}
}

func (c *Ctx) AttachTo(gc *gin.Context) {
	gc.Set(Key, c)
}

func DetachFrom(gc *gin.Context) *Ctx {
	c, ok := gc.Get(Key)
	if ok {
		return c.(*Ctx)
	}
	return nil
}

func (c *Ctx) String() string {
	if c == nil {
		return ""
	}
	if bs, err := json.Marshal(&c); err != nil {
		tlog.Entry().Errorln("to string err ", err)
		return ""
	} else {
		return string(bs)
	}
}

func (c *Ctx) AddHeader(key string, val string) {
	c.Headers[key] = &Header{
		Key: key,
		Val: val,
	}
}

func (c *Ctx) AddRoute(route *route.Route) {
	c.Route = route
}

func (c *Ctx) SkipLogin() bool {
	return !c.NeedLogin()
}

func (c *Ctx) NeedLogin() bool {
	if c == nil {
		return true
	}
	if c.RunAs {
		return false
	}
	if c.Route == nil {
		return true
	}
	return c.Route.Login
}

func (c *Ctx) RefreshUpstreamStartAt() {
	c.UpstreamStartAt = time.Now()
}

func (c *Ctx) CurrentServerCost() int64 {
	return time.Now().UnixMilli() - c.StartAt.UnixMilli()
}

func (c *Ctx) CurrentClientCost() int64 {
	return time.Now().UnixMilli() - c.UpstreamStartAt.UnixMilli()
}
