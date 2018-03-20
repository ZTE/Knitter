package context

import (
	"fmt"
	"github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestRemoveNetFromFlowMgrActionPanic(t *testing.T) {
	action := RemoveNetFromFlowMgrAction{}
	err := action.Exec(nil)
	fmt.Println("recover err:", err)

	convey.Convey("TestRemoveNetFromFlowMgrActionPanic\n", t, func() {
		convey.So(err, convey.ShouldNotBeNil)
	})
}
