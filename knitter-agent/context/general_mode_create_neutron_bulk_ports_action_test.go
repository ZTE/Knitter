package context

import (
	"encoding/json"
	"errors"
	"github.com/ZTE/Knitter/knitter-agent/domain/cni"
	"github.com/ZTE/Knitter/knitter-agent/domain/manager"
	"github.com/ZTE/Knitter/knitter-agent/domain/object/knitter-obj"
	"github.com/ZTE/Knitter/knitter-agent/domain/object/pod-obj"
	"github.com/ZTE/Knitter/knitter-agent/domain/object/port-obj"
	"github.com/ZTE/Knitter/knitter-agent/domain/role/db-role"
	"github.com/ZTE/Knitter/pkg/inter-cmpt/mgr-agt"
	"github.com/ZTE/Knitter/pkg/trans-dsl"
	"github.com/bouk/monkey"
	"github.com/golang/gostub"
	"github.com/rackspace/gophercloud/openstack/networking/v2/ports"
	"github.com/smartystreets/goconvey/convey"
	"net/http"
	"reflect"
	"testing"
)

func TestGeneralModeCreateNeutronBulkPortsActionForSucc(t *testing.T) {
	action := GeneralModeCreateNeutronBulkPortsAction{}

	cniParam := &cni.CniParam{TenantID: "test-tenantid"}
	knitterObj := &knitterobj.KnitterObj{CniParam: cniParam}
	knitterInfo := &KnitterInfo{KnitterObj: knitterObj}
	eagerAttr := portobj.PortEagerAttr{
		NetworkName:  "control",
		NetworkPlane: "control",
		PortName:     "eth0",
		VnicType:     "direct",
	}
	portObjs := []*portobj.PortObj{}
	portObj := &portobj.PortObj{
		EagerAttr: eagerAttr,
	}
	portObjs = append(portObjs, portObj)
	podObj := &podobj.PodObj{
		PortObjs: portObjs,
	}
	podObj.PodID = "test-pod-id"
	knitterInfo.podObj = podObj

	fixIps := []ports.IP{}
	ip := ports.IP{
		SubnetID:  "right subnet id",
		IPAddress: "127.0.0.1",
	}
	fixIps = append(fixIps, ip)

	info := mgragt.CreatePortInfo{
		Name:     "test-port-name",
		FixedIps: fixIps,
	}
	resp := mgragt.CreatePortResp{Port: info}

	stubs := gostub.StubFunc(&manager.HTTPPost, &http.Response{StatusCode: 200}, nil)
	defer stubs.Reset()
	portBytes, _ := json.Marshal(resp)
	stubs.StubFunc(&manager.HTTPReadAll, portBytes, nil)
	stubs.StubFunc(&manager.HTTPClose, nil)
	transInfo := &transdsl.TransInfo{AppInfo: knitterInfo}

	var podRole dbrole.PodRole
	guard := monkey.PatchInstanceMethod(reflect.TypeOf(podRole), "SaveLogicPortInfoForPod", func(_ dbrole.PodRole, _, _, _ string, _ []*portobj.LogicPortObj) error {
		return nil
	})
	defer guard.Unpatch()

	convey.Convey("TestGeneralModeCreateNeutronBulkPortsActionForSucc\n", t, func() {
		err := action.Exec(transInfo)
		convey.So(err, convey.ShouldEqual, nil)
		convey.So(knitterInfo.podObj.PortObjs[0].LazyAttr.Name, convey.ShouldEqual, "test-port-name")
	})
}

func TestGeneralModeCreateNeutronBulkPortsActionForErr_httpPost(t *testing.T) {
	action := GeneralModeCreateNeutronBulkPortsAction{}

	cniParam := &cni.CniParam{TenantID: "test-tenantid"}
	knitterObj := &knitterobj.KnitterObj{CniParam: cniParam}
	knitterInfo := &KnitterInfo{KnitterObj: knitterObj}
	eagerAttr := portobj.PortEagerAttr{
		NetworkName:  "control",
		NetworkPlane: "control",
		PortName:     "eth0",
		VnicType:     "direct",
	}
	portObjs := []*portobj.PortObj{}
	portObj := &portobj.PortObj{
		EagerAttr: eagerAttr,
	}
	portObjs = append(portObjs, portObj)
	podObj := &podobj.PodObj{
		PortObjs: portObjs,
	}
	podObj.PodID = "test-pod-id"
	knitterInfo.podObj = podObj

	fixIps := []ports.IP{}
	ip := ports.IP{
		SubnetID:  "right subnet id",
		IPAddress: "127.0.0.1",
	}
	fixIps = append(fixIps, ip)

	info := mgragt.CreatePortInfo{
		Name:     "test-port-name",
		FixedIps: fixIps,
	}
	resp := mgragt.CreatePortResp{Port: info}

	stubs := gostub.StubFunc(&manager.HTTPPost, &http.Response{StatusCode: 409}, errors.New("http post err"))
	defer stubs.Reset()
	portBytes, _ := json.Marshal(resp)
	stubs.StubFunc(&manager.HTTPReadAll, portBytes, nil)
	stubs.StubFunc(&manager.HTTPClose, nil)
	transInfo := &transdsl.TransInfo{AppInfo: knitterInfo}

	convey.Convey("TestGeneralModeCreateNeutronBulkPortsAction For Err httpPost\n", t, func() {
		err := action.Exec(transInfo)
		convey.So(err, convey.ShouldNotEqual, nil)
	})
}

func TestGeneralModeCreateNeutronBulkPortsActionPanic(t *testing.T) {
	action := GeneralModeCreateNeutronBulkPortsAction{}
	err := action.Exec(nil)
	convey.Convey("Test GeneralModeCreateNeutronBulkPortsAction Panic\n", t, func() {
		convey.So(err, convey.ShouldNotBeNil)
	})
}
