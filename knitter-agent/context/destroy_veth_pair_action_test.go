package context

import (
	"fmt"
	"github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestDestroyVethPairActionPanic(t *testing.T) {
	action := DestroyVethPairAction{}
	err := action.Exec(nil)
	fmt.Println("recover err:", err)

	convey.Convey("TestDestroyVethPairActionPanic\n", t, func() {
		convey.So(err, convey.ShouldNotBeNil)
	})
}
