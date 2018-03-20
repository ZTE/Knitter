package context

import (
	"testing"

	"errors"
	"github.com/ZTE/Knitter/knitter-agent/domain/adapter"
	"github.com/ZTE/Knitter/knitter-agent/domain/cni"
	"github.com/ZTE/Knitter/knitter-agent/domain/object/knitter-obj"
	"github.com/ZTE/Knitter/knitter-agent/mock/mock-pkg/mock-dbaccessor"
	"github.com/ZTE/Knitter/pkg/db-accessor"
	"github.com/ZTE/Knitter/pkg/trans-dsl"
	"github.com/coreos/etcd/client"
	"github.com/golang/gostub"
	"github.com/golang/mock/gomock"
	"github.com/smartystreets/goconvey/convey"
)

func TestDeleteLogicPortsForPodAction_Exec(t *testing.T) {
	mockCtl := gomock.NewController(t)
	defer mockCtl.Finish()
	mockRemoteDB := mockdbaccessor.NewMockDbAccessor(mockCtl)
	agtCtx := &cni.AgentContext{RemoteDB: mockRemoteDB}
	stubs := gostub.StubFunc(&cni.GetGlobalContext, agtCtx)
	defer stubs.Reset()
	cniParam := &cni.CniParam{TenantID: "tenantid1", PodNs: "podns1", PodName: "podname1"}
	knitterObj := &knitterobj.KnitterObj{CniParam: cniParam}
	knitterInfo := &KnitterInfo{KnitterObj: knitterObj}
	transInfo := &transdsl.TransInfo{AppInfo: knitterInfo}
	portInfo := string(`{
		"id": "id1",
		"tenant_id": "tenant1"
	}`)
	nodes := []*client.Node{{Value: portInfo}}
	keyPorts := dbaccessor.GetKeyOfLogicPortsInPod(cniParam.TenantID, cniParam.PodNs, cniParam.PodName)
	mockRemoteDB.EXPECT().ReadDir(keyPorts).Return(nodes, nil)

	stubs.StubFunc(&adapter.DestroyPort, nil)
	keyPort := dbaccessor.GetKeyOfLogicPort(cniParam.TenantID, cniParam.PodNs, cniParam.PodName, "id1")
	mockRemoteDB.EXPECT().DeleteLeaf(keyPort).Return(errors.New("DeleteLeaf err"))
	mockRemoteDB.EXPECT().ReadDir(keyPorts).Return([]*client.Node{}, nil)
	mockRemoteDB.EXPECT().DeleteDir(gomock.Any()).Return(nil)

	err := (&DeleteLogicPortsForPodAction{}).Exec(transInfo)
	convey.Convey("TestDeleteLogicPortsForPodAction_Exec", t, func() {
		convey.So(err, convey.ShouldBeNil)
	})
}

func TestDeleteLogicPortsForPodAction_Err_ReadDir(t *testing.T) {
	mockCtl := gomock.NewController(t)
	defer mockCtl.Finish()
	mockRemoteDB := mockdbaccessor.NewMockDbAccessor(mockCtl)
	agtCtx := &cni.AgentContext{RemoteDB: mockRemoteDB}
	stubs := gostub.StubFunc(&cni.GetGlobalContext, agtCtx)
	defer stubs.Reset()
	cniParam := &cni.CniParam{TenantID: "tenantid1", PodNs: "podns1", PodName: "podname1"}
	knitterObj := &knitterobj.KnitterObj{CniParam: cniParam}
	knitterInfo := &KnitterInfo{KnitterObj: knitterObj}
	transInfo := &transdsl.TransInfo{AppInfo: knitterInfo}

	keyPorts := dbaccessor.GetKeyOfLogicPortsInPod(cniParam.TenantID, cniParam.PodNs, cniParam.PodName)
	mockRemoteDB.EXPECT().ReadDir(keyPorts).Return(nil, errors.New("ReadDir err"))

	err := (&DeleteLogicPortsForPodAction{}).Exec(transInfo)
	convey.Convey("TestDeleteLogicPortsForPodAction_Exec", t, func() {
		convey.So(err, convey.ShouldBeNil)
	})
}

func TestDeleteLogicPortsForPodAction_Err_DestroyPort(t *testing.T) {
	mockCtl := gomock.NewController(t)
	defer mockCtl.Finish()
	mockRemoteDB := mockdbaccessor.NewMockDbAccessor(mockCtl)
	agtCtx := &cni.AgentContext{RemoteDB: mockRemoteDB}
	stubs := gostub.StubFunc(&cni.GetGlobalContext, agtCtx)
	defer stubs.Reset()
	cniParam := &cni.CniParam{TenantID: "tenantid1", PodNs: "podns1", PodName: "podname1"}
	knitterObj := &knitterobj.KnitterObj{CniParam: cniParam}
	knitterInfo := &KnitterInfo{KnitterObj: knitterObj}
	transInfo := &transdsl.TransInfo{AppInfo: knitterInfo}
	portInfo := string(`{
		"id": "id1"
	}`)
	nodes := []*client.Node{{Value: portInfo}}
	keyPorts := dbaccessor.GetKeyOfLogicPortsInPod(cniParam.TenantID, cniParam.PodNs, cniParam.PodName)
	mockRemoteDB.EXPECT().ReadDir(keyPorts).Return(nodes, nil)

	stubs.StubFunc(&adapter.DestroyPort, errors.New("DestroyPort err"))
	mockRemoteDB.EXPECT().ReadDir(keyPorts).Return(nodes, nil)

	err := (&DeleteLogicPortsForPodAction{}).Exec(transInfo)
	convey.Convey("TestDeleteLogicPortsForPodAction_Exec", t, func() {
		convey.So(err, convey.ShouldBeNil)
	})
}

func TestDeleteLogicPortsForPodAction_Panic(t *testing.T) {
	action := DeleteLogicPortsForPodAction{}
	err := action.Exec(nil)
	convey.Convey("TestDeleteLogicPortsForPodAction_Panic\n", t, func() {
		convey.So(err, convey.ShouldNotBeNil)
	})
}

func TestDeleteLogicPortsForPodAction_RollBack(t *testing.T) {
	action := &DeleteLogicPortsForPodAction{}
	transInfo := &transdsl.TransInfo{}
	convey.Convey("Test DeleteLogicPortsForPodAction RollBack\n", t, func() {
		action.RollBack(transInfo)
	})
}
