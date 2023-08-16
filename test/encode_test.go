package test

import (
	"fmt"
	"testing"
	"time"
)

type Widget string

type WrappedWidget struct {
	Widget          // this is the promoted field
	time.Time       // this is another anonymous field that has a runtime name of Time
	price     int64 // normal field
}

func TestEncode(t *testing.T) {
	wrappedWidget := WrappedWidget{"string", time.Now(), 1234}

	fmt.Printf("Widget named %s, created at %s, has price %d\n",
		wrappedWidget.Widget,
		//wrappedWidget.name, // name is passed on to the wrapped Widget since it's
		// the promoted field

		wrappedWidget.GoString(), // We access the anonymous time.Time as Time
		wrappedWidget.price)

	fmt.Printf("Widget named %s, created at %s, has price %d\n",
		wrappedWidget.Widget, // We can also access the Widget directly
		// via Widget
		wrappedWidget.Time,
		wrappedWidget.price)
}
