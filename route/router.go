package route

import (
	"coastline/db"
	"coastline/tlog"
	"encoding/json"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"strconv"
	"strings"
	"sync"
	"time"
)

var checkpoints = make(map[string]time.Time)
var routes = make(map[string]*Route)
var mutex sync.RWMutex

func StartPull() {
	tlog.Entry().Info("start pull routes")
	for {
		refreshCheckpoints()
		time.Sleep(time.Duration(5) * time.Second)
	}
}

func refreshCheckpoints() {
	logEntry := tlog.NewEntry(uuid.NewString())
	logEntry.Debugln("start refresh route checkpoint")

	var cps []Checkpoint
	err := db.Mysql().Select(&cps, "select id, service_name, refreshed_at from route_checkpoint where id > ?", 0)
	if err != nil {
		logEntry.Errorln("select checkpoint error", err)
		return
	}

	if len(cps) > 0 {
		for _, cp := range cps {
			cpAt := checkpoints[cp.ServiceName]
			if cpAt.Before(cp.RefreshedAt) {
				logEntry.Infof("need refresh routes, serviceName %s, before %s, current %s\n",
					cp.ServiceName, cpAt, cp.RefreshedAt)

				err = refreshRoutes(cp.ServiceName, logEntry)
				if err != nil {
					logEntry.Errorln("refresh routes error:", err)
					continue
				}

				checkpoints[cp.ServiceName] = cp.RefreshedAt
			}
		}
	}
}

func refreshRoutes(serviceName string, logEntry *logrus.Entry) error {
	logEntry.Infof("refresh routes by service name %s", serviceName)

	var rts []*Route
	err := db.Mysql().Select(&rts, "select service_name, port, path, login from route where service_name = ?", serviceName)
	if err != nil {
		logEntry.Errorln("select routes error:", err)
		return err
	}

	if len(rts) > 0 {
		safeLoadRoutes(rts)
	}

	bs, err := json.Marshal(routes)
	if err != nil {
		return err
	}

	logEntry.Infof("refresh routes success %s\n", bs)
	return nil
}

func safeLoadRoutes(rts []*Route) {
	mutex.Lock()
	defer mutex.Unlock()
	for _, rt := range rts {
		if rt.Port == 0 {
			rt.Port = 80
		}

		var builder strings.Builder
		builder.WriteString("http://")
		builder.WriteString(rt.ServiceName)
		builder.WriteString(":")
		builder.WriteString(strconv.Itoa(rt.Port))
		builder.WriteString(rt.Path)

		rt.UpstreamUrl = builder.String()
		routes[rt.Path] = rt
	}
}

func Lookup(path string) (*Route, bool) {
	mutex.RLock()
	defer mutex.RUnlock()
	r, ok := routes[path]
	return r, ok
}
