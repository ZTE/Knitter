package context

import (
	"fmt"
	"github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestGetAllPortsOfPodActionPanic(t *testing.T) {
	action := GetAllPortsOfPodAction{}
	err := action.Exec(nil)
	fmt.Println("recover err:", err)

	convey.Convey("TestGetAllPortsOfPodActionPanic\n", t, func() {
		convey.So(err, convey.ShouldNotBeNil)
	})
}
