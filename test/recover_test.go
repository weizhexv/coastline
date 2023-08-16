package test

import (
	"log"
	"testing"
	"time"
)

func TestPanic(t *testing.T) {
	log.Println("Hello before")
	panicFunc()
	time.Sleep(time.Second * 10)
	log.Println("hello stop")
}

func panicFunc() {
	go func() {
		panic("panic in panicFunc")
	}()
}

func myGo(f func()) {
	defer func() {
		log.Println("inside defer")
		if r := recover(); r != nil {
			log.Println("inside recover", r)
		}
	}()
	f()
}
