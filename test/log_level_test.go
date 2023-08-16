package test

import (
	"coastline/tlog"
	"fmt"
	"testing"
)

func TestLogLevel(t *testing.T) {
	v := tlog.IsDebug()
	fmt.Println("v====|", v)
}
