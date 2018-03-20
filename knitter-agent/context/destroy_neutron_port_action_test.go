package context

import (
	"fmt"
	"github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestDestroyNeutronPortActionPanic(t *testing.T) {
	action := DestroyNeutronPortAction{}
	err := action.Exec(nil)
	fmt.Println("recover err:", err)

	convey.Convey("TestDestroyNeutronPortActionPanic\n", t, func() {
		convey.So(err, convey.ShouldNotBeNil)
	})
}
