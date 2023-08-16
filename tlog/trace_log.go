package tlog

import (
	"coastline/consts"
	"coastline/safeutil"
	"coastline/vconfig"
	"github.com/sirupsen/logrus"
	"os"
	"sort"
)

var traceLog = initTLog()

func initTLog() *logrus.Logger {
	level, err := logrus.ParseLevel(vconfig.LoggingLevel())
	if err != nil {
		panic(err)
	}
	return &logrus.Logger{
		Out:       os.Stdout,
		Formatter: setFormatter(),
		Level:     level,
	}
}

func setFormatter() logrus.Formatter {
	return &logrus.TextFormatter{
		DisableQuote:           true,
		FullTimestamp:          true,
		TimestampFormat:        "2006-01-02 15:04:05.999999",
		DisableSorting:         true,
		DisableLevelTruncation: true,
		PadLevelText:           false,
		SortingFunc: func(strings []string) {
			sort.Slice(strings, func(i, j int) bool {
				if strings[i] == "level" {
					return true
				}
				return false
			})
		},
	}
}

func Entry() *logrus.Logger {
	return traceLog
}

func NewEntry(traceId string) *logrus.Entry {
	return traceLog.WithFields(logrus.Fields{
		consts.HeaderTraceId: traceId,
		"goid":               safeutil.QuickGetGoRoutineId(),
	})
}

func IsDebug() bool {
	return traceLog.IsLevelEnabled(logrus.DebugLevel)
}
