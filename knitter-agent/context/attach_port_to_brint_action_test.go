package context

import (
	"fmt"
	"github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestAttachPortToBrIntActionPanic(t *testing.T) {
	action := AttachPortToBrIntAction{}
	err := action.Exec(nil)
	fmt.Println("recover err:", err)

	convey.Convey("TestAttachPortToBrIntActionPanic\n", t, func() {
		convey.So(err, convey.ShouldNotBeNil)
	})
}
