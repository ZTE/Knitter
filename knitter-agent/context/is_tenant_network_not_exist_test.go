package context

import (
	"github.com/ZTE/Knitter/knitter-agent/domain/cni"
	"github.com/ZTE/Knitter/knitter-agent/domain/object/knitter-obj"
	"github.com/ZTE/Knitter/knitter-agent/domain/object/pod-obj"
	"github.com/ZTE/Knitter/knitter-agent/domain/object/port-obj"
	"github.com/ZTE/Knitter/knitter-agent/infra"
	"github.com/ZTE/Knitter/pkg/trans-dsl"
	"github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestIsTenantNetworkNotExist(t *testing.T) {
	action := &IsTenantNetworkNotExist{}

	cniParam := &cni.CniParam{PodNs: "nw001"}
	knitterObj := &knitterobj.KnitterObj{CniParam: cniParam}

	var portObjs []*portobj.PortObj
	portObjs = append(portObjs,
		&portobj.PortObj{LazyAttr: portobj.PortLazyAttr{NetAttr: portobj.NetworkAttrs{ID: "id1"}}},
	)
	podObj := &podobj.PodObj{PortObjs: portObjs}
	knitterInfo := &KnitterInfo{KnitterObj: knitterObj, podObj: podObj}
	transInfo := &transdsl.TransInfo{AppInfo: knitterInfo}

	infra.Init()

	convey.Convey("TestIsTenantNetworkNotExist for succ!\n", t, func() {
		ok := action.Ok(transInfo)
		convey.So(ok, convey.ShouldEqual, true)
	})
}
