package route

import (
	"coastline/tlog"
	"encoding/json"
	"time"
)

type Checkpoint struct {
	Id          int64     `json:"id" db:"id"`
	ServiceName string    `json:"serviceName" db:"service_name"`
	RefreshedAt time.Time `json:"refreshedAt" db:"refreshed_at"`
}

type Route struct {
	ServiceName string `json:"serviceName" db:"service_name"`
	Port        int    `json:"port" db:"port"`
	Path        string `json:"path" db:"path"`
	Login       bool   `json:"login" db:"login"`
	UpstreamUrl string `json:"upstreamUrl"`
}

func (r *Route) String() string {
	if r == nil {
		return ""
	}
	if bs, err := json.Marshal(r); err != nil {
		tlog.Entry().Errorln("to string err ", err)
		return ""
	} else {
		return string(bs)
	}
}
