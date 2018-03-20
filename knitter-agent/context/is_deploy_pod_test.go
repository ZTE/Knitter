package context

import (
	"github.com/ZTE/Knitter/knitter-agent/domain/cni"
	"github.com/ZTE/Knitter/knitter-agent/domain/object/knitter-obj"
	"github.com/ZTE/Knitter/pkg/trans-dsl"
	"github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestIsDeployPod(t *testing.T) {
	action := &IsDeployPod{}
	cniParam := &cni.CniParam{PodNs: "nw001", PodName: "pod_name-deploy"}
	knitterObj := &knitterobj.KnitterObj{CniParam: cniParam}

	knitterInfo := &KnitterInfo{KnitterObj: knitterObj}
	transInfo := &transdsl.TransInfo{AppInfo: knitterInfo}

	convey.Convey("TestIsDeployPod for succ\n", t, func() {
		ok := action.Ok(transInfo)
		convey.So(ok, convey.ShouldEqual, true)
	})
}

func TestIsDeployPod_false(t *testing.T) {
	action := &IsDeployPod{}
	cniParam := &cni.CniParam{PodNs: "nw001", PodName: "pod_name"}
	knitterObj := &knitterobj.KnitterObj{CniParam: cniParam}

	knitterInfo := &KnitterInfo{KnitterObj: knitterObj}
	transInfo := &transdsl.TransInfo{AppInfo: knitterInfo}

	convey.Convey("TestIsDeployPod for return false\n", t, func() {
		ok := action.Ok(transInfo)
		convey.So(ok, convey.ShouldEqual, false)
	})
}
