package context

import (
	"fmt"
	"github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestGetVniActionPanic(t *testing.T) {
	action := GetVniAction{}
	err := action.Exec(nil)
	fmt.Println("recover err:", err)

	convey.Convey("TestGetVniActionPanic\n", t, func() {
		convey.So(err, convey.ShouldNotBeNil)
	})
}
