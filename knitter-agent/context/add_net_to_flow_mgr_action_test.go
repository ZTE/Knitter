package context

import (
	"fmt"
	"github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestAddNetToFlowMgrActionPanic(t *testing.T) {
	action := AddNetToFlowMgrAction{}
	err := action.Exec(nil)
	fmt.Println("recover err:", err)

	convey.Convey("TestAddNetToFlowMgrActionPanic\n", t, func() {
		convey.So(err, convey.ShouldNotBeNil)
	})
}
