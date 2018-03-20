package context

import (
	"fmt"
	"github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestDetachPortFromBrintActionPanic(t *testing.T) {
	action := DetachPortFromBrintAction{}
	err := action.Exec(nil)
	fmt.Println("recover err:", err)

	convey.Convey("TestDetachPortFromBrintActionPanic\n", t, func() {
		convey.So(err, convey.ShouldNotBeNil)
	})
}
