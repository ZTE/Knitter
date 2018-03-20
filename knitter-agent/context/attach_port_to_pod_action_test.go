package context

import (
	"fmt"
	"github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestAttachPortToPodActionPanic(t *testing.T) {
	action := AttachPortToPodAction{}
	err := action.Exec(nil)
	fmt.Println("recover err:", err)

	convey.Convey("TestAttachPortToPodActionPanic\n", t, func() {
		convey.So(err, convey.ShouldNotBeNil)
	})
}
