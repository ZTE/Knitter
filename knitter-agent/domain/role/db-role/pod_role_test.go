package dbrole

import (
	"encoding/json"
	"testing"

	"github.com/coreos/etcd/client"
	"github.com/golang/mock/gomock"
	"github.com/smartystreets/goconvey/convey"

	"errors"
	"github.com/ZTE/Knitter/knitter-agent/domain/cni"
	"github.com/ZTE/Knitter/knitter-agent/domain/object/port-obj"
	"github.com/ZTE/Knitter/knitter-agent/mock/mock-pkg/mock-dbaccessor"
	"github.com/ZTE/Knitter/pkg/db-accessor"
	"github.com/ZTE/Knitter/pkg/iaas-accessor"
	"github.com/golang/gostub"
)

func TestPodRole_GetAllPortJsonList_SUCC(t *testing.T) {
	mockCtl := gomock.NewController(t)
	defer mockCtl.Finish()
	mockDB := mockdbaccessor.NewMockDbAccessor(mockCtl)

	tenantID := "test_tenant_id"
	podName := "test_pod_name"

	eth0PortID := "eth0-port-id"
	eth1PortID := "eth1-port-id"
	eth2PortID := "eth2-port-id"

	netAPIID := "net-api-id"
	netControlID := "net-control-id"
	netMediaID := "net-media-id"

	podInterfaces := []*iaasaccessor.Interface{
		{
			NetPlane:     "std",
			NetPlaneName: "net_api",
			Name:         "eth0",
			NicType:      "normal",
			PodNs:        tenantID,
			PodName:      podName,
			Accelerate:   "false",
			Id:           eth0PortID,
			NetworkId:    netAPIID,
		},
		{
			NetPlane:     "control",
			NetPlaneName: "net_control",
			Name:         "eth1",
			NicType:      "normal",
			PodNs:        tenantID,
			PodName:      podName,
			Accelerate:   "false",
			Id:           eth1PortID,
			NetworkId:    netControlID,
		},
		{
			NetPlane:     "media",
			NetPlaneName: "net_media",
			Name:         "eth2",
			NicType:      "normal",
			PodNs:        tenantID,
			PodName:      podName,
			Accelerate:   "false",
			Id:           eth2PortID,
			NetworkId:    netMediaID,
		},
	}
	keyDir := dbaccessor.GetKeyOfInterfaceGroupInPod(tenantID, tenantID, podName)
	eth0PortKey := dbaccessor.GetKeyOfInterfaceSelf(tenantID, eth0PortID)
	eth1PortKey := dbaccessor.GetKeyOfInterfaceSelf(tenantID, eth1PortID)
	eth2PortKey := dbaccessor.GetKeyOfInterfaceSelf(tenantID, eth2PortID)

	nodes := []*client.Node{{Value: eth0PortKey}, {Value: eth1PortKey}, {Value: eth2PortKey}}

	mockDB.EXPECT().ReadDir(keyDir).Return(nodes, nil)
	eth0PortJSON, _ := json.Marshal(podInterfaces[0])
	eth1PortJSON, _ := json.Marshal(podInterfaces[1])
	eth2PortJSON, _ := json.Marshal(podInterfaces[2])
	gomock.InOrder(
		mockDB.EXPECT().ReadLeaf(eth0PortKey).Return(string(eth0PortJSON), nil),
		mockDB.EXPECT().ReadLeaf(eth1PortKey).Return(string(eth1PortJSON), nil),
		mockDB.EXPECT().ReadLeaf(eth2PortKey).Return(string(eth2PortJSON), nil),
	)

	portJSONList, err := PodRole{}.GetAllPortJSONList(tenantID, tenantID, podName, mockDB)

	convey.Convey("TestPodRole_GetAllPortJsonList_SUCC", t, func() {
		convey.So(err, convey.ShouldBeNil)
		convey.So(portJSONList, convey.ShouldResemble, []string{string(eth0PortJSON), string(eth1PortJSON), string(eth2PortJSON)})
	})
}

func TestPodRole_GetAllPortJsonList_ReadDirFailed(t *testing.T) {
	mockCtl := gomock.NewController(t)
	defer mockCtl.Finish()
	mockDB := mockdbaccessor.NewMockDbAccessor(mockCtl)

	tenantID := "test_tenant_id"
	podName := "test_pod_name"

	keyDir := dbaccessor.GetKeyOfInterfaceGroupInPod(tenantID, tenantID, podName)

	errReadDir := "read dir failed"
	mockDB.EXPECT().ReadDir(keyDir).Return(nil, errors.New(errReadDir))

	portJSONList, err := PodRole{}.GetAllPortJSONList(tenantID, tenantID, podName, mockDB)

	convey.Convey("TestPodRole_GetAllPortJsonList_ReadDirFailed", t, func() {
		convey.So(err.Error(), convey.ShouldEqual, errReadDir)
		convey.So(portJSONList, convey.ShouldBeNil)
	})
}

func TestPodRole_GetAllPortJsonList_ReadLeafFailed(t *testing.T) {
	mockCtl := gomock.NewController(t)
	defer mockCtl.Finish()
	mockDB := mockdbaccessor.NewMockDbAccessor(mockCtl)

	tenantID := "test_tenant_id"
	podName := "test_pod_name"

	eth0PortID := "eth0-port-id"
	eth1PortID := "eth1-port-id"
	eth2PortID := "eth2-port-id"

	netAPIID := "net-api-id"
	netControlID := "net-control-id"
	netMediaID := "net-media-id"

	podInterfaces := []*iaasaccessor.Interface{
		{
			NetPlane:     "std",
			NetPlaneName: "net_api",
			Name:         "eth0",
			NicType:      "normal",
			PodNs:        tenantID,
			PodName:      podName,
			Accelerate:   "false",
			Id:           eth0PortID,
			NetworkId:    netAPIID,
		},
		{
			NetPlane:     "control",
			NetPlaneName: "net_control",
			Name:         "eth1",
			NicType:      "normal",
			PodNs:        tenantID,
			PodName:      podName,
			Accelerate:   "false",
			Id:           eth1PortID,
			NetworkId:    netControlID,
		},
		{
			NetPlane:     "media",
			NetPlaneName: "net_media",
			Name:         "eth2",
			NicType:      "normal",
			PodNs:        tenantID,
			PodName:      podName,
			Accelerate:   "false",
			Id:           eth2PortID,
			NetworkId:    netMediaID,
		},
	}
	keyDir := dbaccessor.GetKeyOfInterfaceGroupInPod(tenantID, tenantID, podName)
	eth0PortKey := dbaccessor.GetKeyOfInterfaceSelf(tenantID, eth0PortID)
	eth1PortKey := dbaccessor.GetKeyOfInterfaceSelf(tenantID, eth1PortID)
	eth2PortKey := dbaccessor.GetKeyOfInterfaceSelf(tenantID, eth2PortID)

	nodes := []*client.Node{{Value: eth0PortKey}, {Value: eth1PortKey}, {Value: eth2PortKey}}

	mockDB.EXPECT().ReadDir(keyDir).Return(nodes, nil)
	eth0PortJSON, _ := json.Marshal(podInterfaces[0])

	errReadLeaf := "read leaf failed"

	mockDB.EXPECT().ReadLeaf(eth0PortKey).Return(string(eth0PortJSON), errors.New(errReadLeaf))

	portJSONList, err := PodRole{}.GetAllPortJSONList(tenantID, tenantID, podName, mockDB)

	convey.Convey("TestPodRole_GetAllPortJsonList_ReadLeafFailed", t, func() {
		convey.So(err.Error(), convey.ShouldEqual, errReadLeaf)
		convey.So(portJSONList, convey.ShouldBeNil)
	})
}

func TestPodRole_SaveLogicPortInfoForPod(t *testing.T) {
	tanantID := "tanantID1"
	podNs := "podNs1"
	podName := "podName1"
	ports := []*portobj.LogicPortObj{
		{
			Accelerate:   "true",
			NetworkPlane: "control",
		},
		{
			Accelerate:   "false",
			NetworkPlane: "std",
			ID:           "id1",
		},
	}
	mockCtl := gomock.NewController(t)
	defer mockCtl.Finish()
	mockDB := mockdbaccessor.NewMockDbAccessor(mockCtl)
	key := dbaccessor.GetKeyOfLogicPort(tanantID, podNs, podName, "id1")
	mockDB.EXPECT().SaveLeaf(key, gomock.Any()).Return(nil)
	stubs := gostub.StubFunc(&cni.GetGlobalContext, &cni.AgentContext{RemoteDB: mockDB})
	defer stubs.Reset()

	err := PodRole{}.SaveLogicPortInfoForPod(tanantID, podNs, podName, ports)
	convey.Convey("TestPodRole_SaveLogicPortInfoForPod", t, func() {
		convey.So(err, convey.ShouldEqual, nil)
	})
}
