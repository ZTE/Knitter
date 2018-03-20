package context

import (
	"github.com/ZTE/Knitter/knitter-agent/domain/object/pod-obj"
	"github.com/ZTE/Knitter/knitter-agent/domain/object/port-obj"
	"github.com/ZTE/Knitter/pkg/trans-dsl"
	"github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestIsOvs(t *testing.T) {
	action := &IsOvs{}
	porattr := portobj.PortEagerAttr{VnicType: "normal"}
	var portObjs = make([]*portobj.PortObj, 0)
	portObjs = append(portObjs,
		&portobj.PortObj{EagerAttr: porattr},
	)
	podObj := &podobj.PodObj{PortObjs: portObjs}
	knitterInfo := &KnitterInfo{podObj: podObj}
	transInfo := &transdsl.TransInfo{AppInfo: knitterInfo}

	ok := action.Ok(transInfo)
	convey.Convey("TestIsOvs for success!\n", t, func() {
		convey.So(ok, convey.ShouldEqual, true)
	})
}

func TestIsOvs_false(t *testing.T) {
	action := &IsOvs{}
	porattr := portobj.PortEagerAttr{VnicType: "direct"}
	var portObjs = make([]*portobj.PortObj, 0)
	portObjs = append(portObjs,
		&portobj.PortObj{EagerAttr: porattr},
	)
	podObj := &podobj.PodObj{PortObjs: portObjs}
	knitterInfo := &KnitterInfo{podObj: podObj}
	transInfo := &transdsl.TransInfo{AppInfo: knitterInfo}

	ok := action.Ok(transInfo)
	convey.Convey("TestIsOvs_EX for return false!\n", t, func() {
		convey.So(ok, convey.ShouldEqual, false)
	})
}
