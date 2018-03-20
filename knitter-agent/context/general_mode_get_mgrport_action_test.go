package context

import (
	"testing"
	//"github.com/ZTE/Knitter/knitter-agent/domain/cni"
	//"github.com/ZTE/Knitter/knitter-agent/domain/object/knitter-obj"
	//"github.com/ZTE/Knitter/knitter-agent/domain/object/port-obj"
	//"github.com/ZTE/Knitter/knitter-agent/domain/object/pod-obj"
	//"github.com/ZTE/Knitter/pkg/trans-dsl"
	//"github.com/smartystreets/goconvey/convey"
	//"github.com/rackspace/gophercloud/openstack/networking/v2/ports"
	"github.com/ZTE/Knitter/knitter-agent/domain/cni"
	"github.com/ZTE/Knitter/knitter-agent/domain/object/knitter-obj"
	"github.com/ZTE/Knitter/knitter-agent/domain/object/pod-obj"
	"github.com/ZTE/Knitter/knitter-agent/domain/object/port-obj"
	"github.com/ZTE/Knitter/pkg/trans-dsl"
	"github.com/golang/gostub"
	"github.com/rackspace/gophercloud/openstack/networking/v2/ports"
	"github.com/smartystreets/goconvey/convey"
)

func TestGeneralModeGetMgrPortActionSucc(t *testing.T) {
	action := &GeneralModeGetMgrPortAction{}
	stubs := gostub.StubFunc(&cni.GetGlobalContext, &cni.AgentContext{Mtu: "1500"})
	defer stubs.Reset()
	cniParam := &cni.CniParam{PodNs: "nw001"}
	knitterObj := &knitterobj.KnitterObj{CniParam: cniParam}
	var portObjs []*portobj.PortObj
	portObjs = append(portObjs,
		&portobj.PortObj{
			EagerAttr: portobj.PortEagerAttr{
				PortName:     "ethapi",
				NetworkName:  "api",
				Accelerate:   "false",
				NetworkPlane: "std"},
			LazyAttr: portobj.PortLazyAttr{
				NetAttr:  portobj.NetworkAttrs{Cidr: "cidrapi", GateWay: "gwapi"},
				FixedIps: []ports.IP{{SubnetID: "subid1", IPAddress: "ip1"}},
				Name:     "ethapi"}})
	podObj := &podobj.PodObj{PortObjs: portObjs}
	knitterInfo := &KnitterInfo{KnitterObj: knitterObj, podObj: podObj}
	transInfo := &transdsl.TransInfo{AppInfo: knitterInfo, RepeatIdx: 0}
	convey.Convey("Test GeneralModeGetMgrPortAction for Succ\n", t, func() {
		ok := action.Exec(transInfo)
		convey.So(ok, convey.ShouldBeNil)
		convey.So(knitterInfo.mgrPort.Name, convey.ShouldEqual, "ethapi")
	})
}

func TestGeneralModeGetMgrPortActionRollBack(t *testing.T) {
	action := &GeneralModeGetMgrPortAction{}
	cniParam := &cni.CniParam{PodNs: "nw001", TenantID: "tenant1"}
	knitterObj := &knitterobj.KnitterObj{CniParam: cniParam}
	var portObjs []*portobj.PortObj
	portObjs = append(portObjs,
		&portobj.PortObj{EagerAttr: portobj.PortEagerAttr{PortName: "etheio1", NetworkName: "eio1", Accelerate: "false", NetworkPlane: "eio"},
			LazyAttr: portobj.PortLazyAttr{NetAttr: portobj.NetworkAttrs{Cidr: "cidreio1", GateWay: "gweio1"}}})
	podObj := &podobj.PodObj{PortObjs: portObjs}
	knitterInfo := &KnitterInfo{KnitterObj: knitterObj, podObj: podObj}
	transInfo := &transdsl.TransInfo{AppInfo: knitterInfo}
	convey.Convey("Test GeneralModeGetMgrPortAction RollBack\n", t, func() {
		action.RollBack(transInfo)
	})
}

func TestGeneralModeGetMgrPortActionPanic(t *testing.T) {
	action := GeneralModeGetMgrPortAction{}
	err := action.Exec(nil)
	convey.Convey("Test GeneralModeGetMgrPortAction Panic\n", t, func() {
		convey.So(err, convey.ShouldNotBeNil)
	})
}
