package context

import (
	"errors"
	"github.com/ZTE/Knitter/knitter-agent/domain/cni"
	"github.com/ZTE/Knitter/knitter-agent/domain/object/knitter-obj"
	"github.com/ZTE/Knitter/knitter-agent/domain/object/pod-obj"
	"github.com/ZTE/Knitter/knitter-agent/domain/object/port-obj"
	"github.com/ZTE/Knitter/knitter-agent/mock/mock-pkg/mock-dbaccessor"
	"github.com/ZTE/Knitter/pkg/trans-dsl"
	"github.com/coreos/etcd/client"
	"github.com/golang/gostub"
	"github.com/golang/mock/gomock"
	"github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestReusePortInfoActionSucc(t *testing.T) {
	action := &ReusePortInfoAction{}

	cniParam := &cni.CniParam{PodNs: "nw001"}
	knitterObj := &knitterobj.KnitterObj{CniParam: cniParam}

	var portObjs []*portobj.PortObj
	portObjs = append(portObjs,
		&portobj.PortObj{EagerAttr: portobj.PortEagerAttr{PortName: "ethcontrol", NetworkName: "control", Accelerate: "true", NetworkPlane: "control"},
			LazyAttr: portobj.PortLazyAttr{NetAttr: portobj.NetworkAttrs{Cidr: "cidrcontrol", GateWay: "gwcontrol"}}},
		&portobj.PortObj{EagerAttr: portobj.PortEagerAttr{PortName: "ethapi", NetworkName: "api", Accelerate: "false", NetworkPlane: "std"},
			LazyAttr: portobj.PortLazyAttr{NetAttr: portobj.NetworkAttrs{Cidr: "cidrapi", GateWay: "gwapi"}}})
	podObj := &podobj.PodObj{PortObjs: portObjs}
	knitterInfo := &KnitterInfo{KnitterObj: knitterObj, podObj: podObj}
	transInfo := &transdsl.TransInfo{AppInfo: knitterInfo}
	cfgMock := gomock.NewController(t)
	defer cfgMock.Finish()
	mockDB := mockdbaccessor.NewMockDbAccessor(cfgMock)
	stubs := gostub.StubFunc(&cni.GetGlobalContext, &cni.AgentContext{RemoteDB: mockDB, RunMode: "underlay"})
	defer stubs.Reset()
	portsInEtcd := []*portobj.LogicPortObj{
		{Name: "ethcontrol", ID: "idcontrol", MacAddress: "maccontrol",
			TenantID: "tenant1", NetworkPlane: "control", NetworkID: "netidcontrol"},
		{Name: "ethapi", ID: "idapi", MacAddress: "macapi",
			TenantID: "tenant1", NetworkPlane: "std", NetworkID: "netidapi"},
	}
	stubs.StubFunc(&getPortListOfPod, portsInEtcd, nil)
	stubs.StubFunc(&cleanPortsRecordInETCD, nil)
	stubs.StubFunc(&cleanPortsRecordInLocalDB, nil)

	convey.Convey("Test VMReusePortInfoAction for Succ\n", t, func() {
		ok := action.Exec(transInfo)
		convey.So(ok, convey.ShouldBeNil)
	})
}

func TestReusePortInfoActionErr_getPortListOfPod(t *testing.T) {
	action := &ReusePortInfoAction{}

	cniParam := &cni.CniParam{PodNs: "nw001"}
	knitterObj := &knitterobj.KnitterObj{CniParam: cniParam}

	var portObjs []*portobj.PortObj
	portObjs = append(portObjs,
		&portobj.PortObj{EagerAttr: portobj.PortEagerAttr{PortName: "ethcontrol", NetworkName: "control", Accelerate: "true", NetworkPlane: "control"},
			LazyAttr: portobj.PortLazyAttr{NetAttr: portobj.NetworkAttrs{Cidr: "cidrcontrol", GateWay: "gwcontrol"}}},
		&portobj.PortObj{EagerAttr: portobj.PortEagerAttr{PortName: "ethapi", NetworkName: "api", Accelerate: "false", NetworkPlane: "std"},
			LazyAttr: portobj.PortLazyAttr{NetAttr: portobj.NetworkAttrs{Cidr: "cidrapi", GateWay: "gwapi"}}})
	podObj := &podobj.PodObj{PortObjs: portObjs}
	knitterInfo := &KnitterInfo{KnitterObj: knitterObj, podObj: podObj}
	transInfo := &transdsl.TransInfo{AppInfo: knitterInfo}
	cfgMock := gomock.NewController(t)
	defer cfgMock.Finish()
	mockDB := mockdbaccessor.NewMockDbAccessor(cfgMock)
	stubs := gostub.StubFunc(&cni.GetGlobalContext, &cni.AgentContext{RemoteDB: mockDB, RunMode: "underlay"})
	defer stubs.Reset()
	stubs.StubFunc(&getPortListOfPod, nil, errors.New("getPortListOfPod err"))

	convey.Convey("Test VMReusePortInfoAction for ERR getPortListOfPod\n", t, func() {
		ok := action.Exec(transInfo)
		convey.So(ok, convey.ShouldNotBeNil)
	})
}

func TestReusePortInfoActionErr_findPortInfoInEtcd(t *testing.T) {
	action := &ReusePortInfoAction{}

	cniParam := &cni.CniParam{PodNs: "nw001"}
	knitterObj := &knitterobj.KnitterObj{CniParam: cniParam}

	var portObjs []*portobj.PortObj
	portObjs = append(portObjs,
		&portobj.PortObj{EagerAttr: portobj.PortEagerAttr{PortName: "ethapi", NetworkName: "api", Accelerate: "false", NetworkPlane: "std"},
			LazyAttr: portobj.PortLazyAttr{NetAttr: portobj.NetworkAttrs{Cidr: "cidrapi", GateWay: "gwapi"}}})
	podObj := &podobj.PodObj{PortObjs: portObjs}
	knitterInfo := &KnitterInfo{KnitterObj: knitterObj, podObj: podObj}
	transInfo := &transdsl.TransInfo{AppInfo: knitterInfo}
	cfgMock := gomock.NewController(t)
	defer cfgMock.Finish()
	mockDB := mockdbaccessor.NewMockDbAccessor(cfgMock)
	stubs := gostub.StubFunc(&cni.GetGlobalContext, &cni.AgentContext{DB: mockDB, RunMode: "underlay"})
	defer stubs.Reset()
	portsInEtcd := []*portobj.LogicPortObj{
		{Name: "etheio2", ID: "ideio1", MacAddress: "maceio1",
			TenantID: "tenant1", NetworkPlane: "eio", NetworkID: "netideio1"},
	}
	stubs.StubFunc(&getPortListOfPod, portsInEtcd, nil)

	convey.Convey("Test VMReusePortInfoAction for ERR findPortInfoInEtcd\n", t, func() {
		ok := action.Exec(transInfo)
		convey.So(ok, convey.ShouldNotBeNil)
	})
}

func TestVMReusePortInfoActionRollBack(t *testing.T) {
	action := &ReusePortInfoAction{}
	transInfo := &transdsl.TransInfo{}
	convey.Convey("Test VMReusePortInfoAction RollBack\n", t, func() {
		action.RollBack(transInfo)
	})
}

func TestGetPortListOfPodSucc(t *testing.T) {
	cfgMock := gomock.NewController(t)
	defer cfgMock.Finish()
	mockDB := mockdbaccessor.NewMockDbAccessor(cfgMock)
	portInfo := `{
		"name": "name1"
	}`
	var nodes []*client.Node
	node := &client.Node{Key: "etcd_key", Value: string(portInfo)}
	nodes = append(nodes, node)
	mockDB.EXPECT().ReadDir(gomock.Any()).Return(nodes, nil)
	agtCtx := &cni.AgentContext{RemoteDB: mockDB}
	convey.Convey("Test getPortListOfPod Succ\n", t, func() {
		ports, err := getPortListOfPod(agtCtx, "tenant1", "podns1", "podname1")
		convey.So(err, convey.ShouldEqual, nil)
		convey.So(ports[0].Name, convey.ShouldEqual, "name1")
	})
}

func TestGetPortListOfPodErr(t *testing.T) {
	convey.Convey("Test getPortListOfPod Err ReadDir\n", t, func() {
		cfgMock := gomock.NewController(t)
		defer cfgMock.Finish()
		mockDB := mockdbaccessor.NewMockDbAccessor(cfgMock)
		mockDB.EXPECT().ReadDir(gomock.Any()).Return(nil, errors.New("ReadDir err"))
		agtCtx := &cni.AgentContext{RemoteDB: mockDB}
		ports, err := getPortListOfPod(agtCtx, "tenant1", "podns1", "podname1")
		convey.So(err, convey.ShouldNotEqual, nil)
		convey.So(ports, convey.ShouldEqual, nil)
	})
	convey.Convey("Test getPortListOfPod Err Unmarshal\n", t, func() {
		cfgMock := gomock.NewController(t)
		defer cfgMock.Finish()
		mockDB := mockdbaccessor.NewMockDbAccessor(cfgMock)
		portInfo := `{
		    "name": "name1",
	    }`
		var nodes []*client.Node
		node := &client.Node{Key: "etcd_key", Value: string(portInfo)}
		nodes = append(nodes, node)
		mockDB.EXPECT().ReadDir(gomock.Any()).Return(nodes, nil)
		agtCtx := &cni.AgentContext{RemoteDB: mockDB}
		ports, err := getPortListOfPod(agtCtx, "tenant1", "podns1", "podname1")
		convey.So(err, convey.ShouldNotEqual, nil)
		convey.So(ports, convey.ShouldEqual, nil)
	})
}
