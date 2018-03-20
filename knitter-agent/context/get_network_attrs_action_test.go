package context

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ZTE/Knitter/knitter-agent/domain/cni"
	"github.com/ZTE/Knitter/knitter-agent/domain/manager"
	"github.com/ZTE/Knitter/knitter-agent/domain/object/knitter-obj"
	"github.com/ZTE/Knitter/knitter-agent/domain/object/pod-obj"
	"github.com/ZTE/Knitter/knitter-agent/domain/object/port-obj"
	"github.com/ZTE/Knitter/knitter-agent/domain/ovs"
	. "github.com/ZTE/Knitter/knitter-agent/err-obj"
	"github.com/ZTE/Knitter/pkg/iaas-accessor"
	"github.com/ZTE/Knitter/pkg/trans-dsl"
	"github.com/golang/gostub"
	. "github.com/golang/gostub"
	"github.com/smartystreets/goconvey/convey"
	"net/http"
	"testing"
)

func TestGetNetworkAttrsActionNoNeedProvider(t *testing.T) {

	action := GetNetworkAttrsAction{}
	cniParam := &cni.CniParam{PodNs: "nw001", HostType: "virtual_machine"}
	knitterObj := &knitterobj.KnitterObj{CniParam: cniParam}

	var portObjs []*portobj.PortObj
	portObjs = append(portObjs,
		&portobj.PortObj{EagerAttr: portobj.PortEagerAttr{NetworkName: "lan"}},
		&portobj.PortObj{EagerAttr: portobj.PortEagerAttr{NetworkName: "net_api"}},
		&portobj.PortObj{EagerAttr: portobj.PortEagerAttr{NetworkName: "control"}})
	podObj := &podobj.PodObj{PortObjs: portObjs}
	knitterInfo := &KnitterInfo{KnitterObj: knitterObj, podObj: podObj}

	mc := manager.ManagerClient{URLKnitterManager: "manager-url", VMID: "200"}
	stubs := gostub.StubFunc(&cni.GetGlobalContext, &cni.AgentContext{Mc: mc})
	defer stubs.Reset()

	stubs.StubFunc(&manager.HTTPPost, &http.Response{StatusCode: 200}, nil)
	networkAttrs := []*portobj.NetworkAttrs{{Name: "lan", ID: "1111111"}}
	networkAttrs = append(networkAttrs, &portobj.NetworkAttrs{Name: "net_api", ID: "22222"})
	networkAttrs = append(networkAttrs, &portobj.NetworkAttrs{Name: "control", ID: "33333"})
	networkBytes, _ := json.Marshal(networkAttrs)
	stubs.StubFunc(&manager.HTTPReadAll, networkBytes, nil)
	stubs.StubFunc(&manager.HTTPClose, nil)
	transInfo := &transdsl.TransInfo{AppInfo: knitterInfo}

	convey.Convey("TestGetNetworkAttrsAction for succ\n", t, func() {
		err := action.Exec(transInfo)
		convey.So(err, convey.ShouldEqual, nil)
	})
}

func TestGetNetworkAttrsActionNeedProvider(t *testing.T) {

	action := GetNetworkAttrsAction{}
	cniParam := &cni.CniParam{PodNs: "nw001", HostType: "bare_metal"}
	knitterObj := &knitterobj.KnitterObj{CniParam: cniParam}

	var portObjs []*portobj.PortObj
	portObjs = append(portObjs,
		&portobj.PortObj{EagerAttr: portobj.PortEagerAttr{NetworkName: "lan"}},
		&portobj.PortObj{EagerAttr: portobj.PortEagerAttr{NetworkName: "net_api"}},
		&portobj.PortObj{EagerAttr: portobj.PortEagerAttr{NetworkName: "control"}})
	podObj := &podobj.PodObj{PortObjs: portObjs}
	knitterInfo := &KnitterInfo{KnitterObj: knitterObj, podObj: podObj}

	mc := manager.ManagerClient{URLKnitterManager: "manager-url", VMID: "200"}
	stubs := gostub.StubFunc(&cni.GetGlobalContext, &cni.AgentContext{Mc: mc})
	defer stubs.Reset()

	stubs.StubFunc(&manager.HTTPPost, &http.Response{StatusCode: 200}, nil)
	networkAttrs := []*portobj.NetworkAttrs{
		{Name: "lan", ID: "1111111",
			Provider: iaasaccessor.NetworkExtenAttrs{NetworkType: "vlan",
				PhysicalNetwork: "physnet1",
				SegmentationID:  "100"}}}
	networkAttrs = append(networkAttrs, &portobj.NetworkAttrs{Name: "net_api", ID: "22222",
		Provider: iaasaccessor.NetworkExtenAttrs{NetworkType: "vxlan",
			PhysicalNetwork: "physnet1",
			SegmentationID:  "101"}})
	networkAttrs = append(networkAttrs, &portobj.NetworkAttrs{Name: "control", ID: "33333",
		Provider: iaasaccessor.NetworkExtenAttrs{NetworkType: "flat",
			PhysicalNetwork: "physnetex",
			SegmentationID:  "0"}})
	networkBytes, _ := json.Marshal(networkAttrs)
	stubs.StubFunc(&manager.HTTPReadAll, networkBytes, nil)
	stubs.StubFunc(&manager.HTTPClose, nil)

	GetNwMechDriverOutputs := []Output{
		{StubVals: Values{"ovs", nil}},
		{StubVals: Values{"sriov", nil}},
		{StubVals: Values{"physical", nil}},
	}
	stubs.StubFuncSeq(&ovs.GetNwMechDriver, GetNwMechDriverOutputs)
	transInfo := &transdsl.TransInfo{AppInfo: knitterInfo}

	convey.Convey("TestGetNetworkAttrsAction for succ\n", t, func() {
		err := action.Exec(transInfo)
		convey.So(err, convey.ShouldEqual, nil)
	})
}

func TestGetNetworkAttrsActionNeedProvider_CheckErr(t *testing.T) {

	action := GetNetworkAttrsAction{}
	cniParam := &cni.CniParam{PodNs: "nw001", HostType: "bare_metal"}
	knitterObj := &knitterobj.KnitterObj{CniParam: cniParam}

	var portObjs []*portobj.PortObj
	portObjs = append(portObjs,
		&portobj.PortObj{EagerAttr: portobj.PortEagerAttr{NetworkName: "lan"}},
		&portobj.PortObj{EagerAttr: portobj.PortEagerAttr{NetworkName: "net_api"}},
		&portobj.PortObj{EagerAttr: portobj.PortEagerAttr{NetworkName: "control"}})
	podObj := &podobj.PodObj{PortObjs: portObjs}
	knitterInfo := &KnitterInfo{KnitterObj: knitterObj, podObj: podObj}

	mc := manager.ManagerClient{URLKnitterManager: "manager-url", VMID: "200"}
	stubs := gostub.StubFunc(&cni.GetGlobalContext, &cni.AgentContext{Mc: mc})
	defer stubs.Reset()

	stubs.StubFunc(&manager.HTTPPost, &http.Response{StatusCode: 200}, nil)
	networkAttrs := []*portobj.NetworkAttrs{
		{Name: "lan", ID: "1111111",
			Provider: iaasaccessor.NetworkExtenAttrs{NetworkType: "vlan",
				PhysicalNetwork: "physnet1",
				SegmentationID:  "100"}}}
	networkAttrs = append(networkAttrs, &portobj.NetworkAttrs{Name: "net_api", ID: "22222",
		Provider: iaasaccessor.NetworkExtenAttrs{NetworkType: "vxlan",
			PhysicalNetwork: "physnet1",
			SegmentationID:  "101"}})
	networkAttrs = append(networkAttrs, &portobj.NetworkAttrs{Name: "control", ID: "33333",
		Provider: iaasaccessor.NetworkExtenAttrs{NetworkType: "flat",
			PhysicalNetwork: "physnetex",
			SegmentationID:  "0"}})
	networkBytes, _ := json.Marshal(networkAttrs)
	stubs.StubFunc(&manager.HTTPReadAll, networkBytes, nil)
	stubs.StubFunc(&manager.HTTPClose, nil)

	GetNwMechDriverOutputs := []Output{
		{StubVals: Values{"ovs", nil}},
		{StubVals: Values{"", ErrAny}},
		{StubVals: Values{"physical", nil}},
	}
	stubs.StubFuncSeq(&ovs.GetNwMechDriver, GetNwMechDriverOutputs)
	transInfo := &transdsl.TransInfo{AppInfo: knitterInfo}

	convey.Convey("TestGetNetworkAttrsAction check error\n", t, func() {
		err := action.Exec(transInfo)
		convey.So(err, convey.ShouldNotEqual, nil)
	})
}

func TestGetNetworkAttrsActionNeedProvider_CheckErr1(t *testing.T) {

	action := GetNetworkAttrsAction{}
	cniParam := &cni.CniParam{PodNs: "nw001", HostType: "bare_metal"}
	knitterObj := &knitterobj.KnitterObj{CniParam: cniParam}

	var portObjs []*portobj.PortObj
	portObjs = append(portObjs,
		&portobj.PortObj{EagerAttr: portobj.PortEagerAttr{NetworkName: "lan"}},
		&portobj.PortObj{EagerAttr: portobj.PortEagerAttr{NetworkName: "net_api"}},
		&portobj.PortObj{EagerAttr: portobj.PortEagerAttr{NetworkName: "control"}})
	podObj := &podobj.PodObj{PortObjs: portObjs}
	knitterInfo := &KnitterInfo{KnitterObj: knitterObj, podObj: podObj}

	mc := manager.ManagerClient{URLKnitterManager: "manager-url", VMID: "200"}
	stubs := gostub.StubFunc(&cni.GetGlobalContext, &cni.AgentContext{Mc: mc})
	defer stubs.Reset()

	stubs.StubFunc(&manager.HTTPPost, &http.Response{StatusCode: 200}, nil)
	networkAttrs := []*portobj.NetworkAttrs{
		{Name: "lan", ID: "1111111",
			Provider: iaasaccessor.NetworkExtenAttrs{NetworkType: "vlan",
				PhysicalNetwork: "physnet1",
				SegmentationID:  "100"}}}
	networkAttrs = append(networkAttrs, &portobj.NetworkAttrs{Name: "net_api", ID: "22222",
		Provider: iaasaccessor.NetworkExtenAttrs{NetworkType: "vxlan",
			PhysicalNetwork: "physnet1",
			SegmentationID:  "101"}})
	networkAttrs = append(networkAttrs, &portobj.NetworkAttrs{Name: "control", ID: "33333",
		Provider: iaasaccessor.NetworkExtenAttrs{NetworkType: "flat",
			PhysicalNetwork: "physnetex",
			SegmentationID:  "0"}})
	networkBytes, _ := json.Marshal(networkAttrs)
	stubs.StubFunc(&manager.HTTPReadAll, networkBytes, nil)
	stubs.StubFunc(&manager.HTTPClose, nil)

	GetNwMechDriverOutputs := []Output{
		{StubVals: Values{"ovs", nil}},
		{StubVals: Values{"other", nil}},
		{StubVals: Values{"physical", nil}},
	}
	stubs.StubFuncSeq(&ovs.GetNwMechDriver, GetNwMechDriverOutputs)
	transInfo := &transdsl.TransInfo{AppInfo: knitterInfo}

	convey.Convey("TestGetNetworkAttrsAction check error\n", t, func() {
		err := action.Exec(transInfo)
		convey.So(err, convey.ShouldNotEqual, nil)
	})
}

func TestGetNetworkAttrsActionErr406(t *testing.T) {

	action := GetNetworkAttrsAction{}
	cniParam := &cni.CniParam{PodNs: "nw001", HostType: "bare_metal"}
	knitterObj := &knitterobj.KnitterObj{CniParam: cniParam}

	var portObjs []*portobj.PortObj
	portObjs = append(portObjs,
		&portobj.PortObj{EagerAttr: portobj.PortEagerAttr{NetworkName: "lan"}},
		&portobj.PortObj{EagerAttr: portobj.PortEagerAttr{NetworkName: "net_api"}},
		&portobj.PortObj{EagerAttr: portobj.PortEagerAttr{NetworkName: "control"}})
	podObj := &podobj.PodObj{PortObjs: portObjs}
	knitterInfo := &KnitterInfo{KnitterObj: knitterObj, podObj: podObj}

	mc := manager.ManagerClient{URLKnitterManager: "manager-url", VMID: "200"}
	stubs := gostub.StubFunc(&cni.GetGlobalContext, &cni.AgentContext{Mc: mc})
	defer stubs.Reset()

	stubs.StubFunc(&manager.HTTPPost, &http.Response{StatusCode: 406}, errors.New("post error"))
	transInfo := &transdsl.TransInfo{AppInfo: knitterInfo}

	convey.Convey("TestGetNetworkAttrsAction for succ\n", t, func() {
		err := action.Exec(transInfo)
		convey.So(err, convey.ShouldNotEqual, nil)
	})
}

func TestGetNetworkAttrsActionPanic(t *testing.T) {
	action := GetNetworkAttrsAction{}
	err := action.Exec(nil)
	fmt.Println("recover err:", err)

	convey.Convey("TestGetNetworkAttrsActionPanic\n", t, func() {
		convey.So(err, convey.ShouldNotBeNil)
	})
}
