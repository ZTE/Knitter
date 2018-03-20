package context

import (
	"fmt"
	"github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestDelPodActionPanic(t *testing.T) {
	action := DelPodAction{}
	err := action.Exec(nil)
	fmt.Println("recover err:", err)

	convey.Convey("TestDelPodActionPanic\n", t, func() {
		convey.So(err, convey.ShouldNotBeNil)
	})
}
