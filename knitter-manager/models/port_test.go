/*
Copyright 2018 ZTE Corporation. All rights reserved.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package models

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/ZTE/Knitter/knitter-manager/const-value"
	"github.com/ZTE/Knitter/knitter-manager/err-obj"
	"github.com/ZTE/Knitter/knitter-manager/public"
	"github.com/ZTE/Knitter/knitter-manager/tests"
	"github.com/ZTE/Knitter/knitter-manager/tests/mock/db-mock"
	"github.com/ZTE/Knitter/pkg/iaas-accessor"
	"github.com/ZTE/Knitter/pkg/inter-cmpt/agt-mgr"
	"github.com/ZTE/Knitter/pkg/inter-cmpt/mgr-agt"
	"github.com/ZTE/Knitter/pkg/inter-cmpt/mgr-iaas"

	"github.com/bouk/monkey"
	"github.com/coreos/etcd/client"
	"github.com/golang/gostub"
	"github.com/golang/mock/gomock"
	. "github.com/smartystreets/goconvey/convey"

	"github.com/ZTE/Knitter/knitter-manager/iaas"
)

func TestGetLogicalPort(t *testing.T) {
	portID := "port-id"
	mockCtl := gomock.NewController(t)
	defer mockCtl.Finish()
	mockDB := mockdbaccessor.NewMockDbAccessor(mockCtl)
	portStr := `{"ID":"port-id"}`
	mockDB.EXPECT().ReadLeaf("/knitter/manager/ports/port-id").Return(portStr, nil)
	stub := gostub.StubFunc(&common.GetDataBase, mockDB)
	defer stub.Reset()

	Convey("TestGetLogicalPort", t, func() {
		logicPort, err := GetLogicalPort(portID)
		So(err, ShouldBeNil)
		So(logicPort.ID, ShouldEqual, portID)
	})
}

func TestGetLogicalPort_DBFailed(t *testing.T) {
	portID := "port-id"
	mockCtl := gomock.NewController(t)
	defer mockCtl.Finish()
	mockDB := mockdbaccessor.NewMockDbAccessor(mockCtl)
	errStr := "etcd cluster misconfig"
	mockDB.EXPECT().ReadLeaf("/knitter/manager/ports/port-id").Return("", errors.New(errStr))
	stub := gostub.StubFunc(&common.GetDataBase, mockDB)
	defer stub.Reset()

	Convey("TestGetLogicalPort_DBFailed", t, func() {
		logicPort, err := GetLogicalPort(portID)
		So(err.Error(), ShouldEqual, errStr)
		So(logicPort, ShouldBeNil)
	})
}

func TestGetLogicalPort_UnmarshalFailed(t *testing.T) {
	portID := "port-id"
	mockCtl := gomock.NewController(t)
	defer mockCtl.Finish()
	mockDB := mockdbaccessor.NewMockDbAccessor(mockCtl)
	portStr := `{"ID":"port-id"}`
	mockDB.EXPECT().ReadLeaf("/knitter/manager/ports/port-id").Return(portStr, nil)
	stub := gostub.StubFunc(&common.GetDataBase, mockDB)
	defer stub.Reset()

	guard := monkey.Patch(json.Unmarshal, func(_ []byte, _ interface{}) error {
		return errors.New("unexpected content")
	})
	defer guard.Unpatch()

	Convey("TestGetLogicalPort_UnmarshalFailed", t, func() {
		logicPort, err := GetLogicalPort(portID)
		So(err, ShouldEqual, errobj.ErrUnmarshalFailed)
		So(logicPort, ShouldBeNil)
	})
}

func TestUnmarshalLogicPort(t *testing.T) {
	Convey("TestUnmarshalLogicPort", t, func() {
		logicPort, err := UnmarshalLogicPort([]byte(`{"ID":"port-id1"}`))
		So(err, ShouldBeNil)
		So(logicPort.ID, ShouldEqual, "port-id1")
	})
}

func TestUnmarshalLogicPort_Fail(t *testing.T) {
	errObj := errors.New("invalid content")
	guard := monkey.Patch(json.Unmarshal, func(_ []byte, _ interface{}) error {
		return errObj
	})
	defer guard.Unpatch()

	Convey("TestUnmarshalLogicPort_Fail", t, func() {
		logicPort, err := UnmarshalLogicPort([]byte(`{"ID":"port-id1"}`))
		So(err, ShouldEqual, errobj.ErrUnmarshalFailed)
		So(logicPort, ShouldBeNil)
	})
}

func TestGetAllLogicalPorts(t *testing.T) {
	mockCtl := gomock.NewController(t)
	defer mockCtl.Finish()
	mockDB := mockdbaccessor.NewMockDbAccessor(mockCtl)

	nodes := []*client.Node{
		{
			Key:   "/knitter/manager/ports/port-id1",
			Value: `{"ID":"port-id1"}`,
		},
		{
			Key:   "/knitter/manager/ports/port-id2",
			Value: `{"ID":"port-id2"}`,
		},
		{
			Key:   "/knitter/manager/ports/port-id3",
			Value: `{"ID":"port-id3"}`,
		},
	}
	mockDB.EXPECT().ReadDir("/knitter/manager/ports").Return(nodes, nil)
	stub := gostub.StubFunc(&common.GetDataBase, mockDB)

	outputs := []gostub.Output{
		{StubVals: gostub.Values{&LogicalPort{ID: "port-id1"}, nil}},
		{StubVals: gostub.Values{&LogicalPort{ID: "port-id2"}, nil}},
		{StubVals: gostub.Values{&LogicalPort{ID: "port-id3"}, nil}},
	}
	stub.StubFuncSeq(&UnmarshalLogicPort, outputs)
	defer stub.Reset()

	Convey("TestGetAllLogicalPorts", t, func() {
		logicPorts, err := GetAllLogicalPorts()
		So(err, ShouldBeNil)
		So(len(logicPorts), ShouldEqual, 3)
		So(logicPorts[0].ID, ShouldEqual, "port-id1")
		So(logicPorts[1].ID, ShouldEqual, "port-id2")
		So(logicPorts[2].ID, ShouldEqual, "port-id3")
	})
}

func TestGetAllLogicalPorts_DBFailed(t *testing.T) {
	mockCtl := gomock.NewController(t)
	defer mockCtl.Finish()
	mockDB := mockdbaccessor.NewMockDbAccessor(mockCtl)

	errStr := "etcd cluster misconfig"
	mockDB.EXPECT().ReadDir("/knitter/manager/ports").Return(nil, errors.New(errStr))

	stub := gostub.StubFunc(&common.GetDataBase, mockDB)
	defer stub.Reset()

	Convey("TestGetAllLogicalPorts_DBFailed", t, func() {
		LogicPorts, err := GetAllLogicalPorts()
		So(err.Error(), ShouldEqual, errStr)
		So(LogicPorts, ShouldBeNil)
	})
}

func TestGetAllLogicalPorts_UnmarshalFail(t *testing.T) {
	mockCtl := gomock.NewController(t)
	defer mockCtl.Finish()
	mockDB := mockdbaccessor.NewMockDbAccessor(mockCtl)

	nodes := []*client.Node{
		{
			Key:   "/knitter/manager/ports/port-id1",
			Value: `{"ID":"port-id1"}`,
		},
		{
			Key:   "/knitter/manager/ports/port-id2",
			Value: `{"ID":"port-id2"}`,
		},
		{
			Key:   "/knitter/manager/ports/port-id3",
			Value: `{"ID":"port-id3"}`,
		},
	}
	mockDB.EXPECT().ReadDir("/knitter/manager/ports").Return(nodes, nil)
	stub := gostub.StubFunc(&common.GetDataBase, mockDB)

	errStr := "etcd cluster misconfig"
	outputs := []gostub.Output{
		{StubVals: gostub.Values{&LogicalPort{ID: "port-id1"}, nil}},
		{StubVals: gostub.Values{&LogicalPort{ID: "port-id2"}, nil}},
		{StubVals: gostub.Values{nil, errors.New(errStr)}},
	}
	stub.StubFuncSeq(&UnmarshalLogicPort, outputs)
	defer stub.Reset()

	Convey("TestGetAllLogicalPorts_UnmarshalFail", t, func() {
		logicPorts, err := GetAllLogicalPorts()
		So(err.Error(), ShouldEqual, errStr)
		So(logicPorts, ShouldBeNil)
	})
}

func TestCreateBulkPortsOK(t *testing.T) {
	Convey("TestCreateBulkPortsOK", t, func() {
		mgrBulkPortsReq := mgriaas.MgrBulkPortsReq{
			TranId: "123",
			Ports: []*mgriaas.MgrPortReq{
				{
					AgtPortReq: agtmgr.AgtPortReq{
						TenantID:    "tenant_id",
						NetworkName: "networkName1",
						PortName:    "portName1",
						NodeID:      "node_id",
						PodNs:       "pod_ns",
						PodName:     "pod_name1",
						ClusterID:   "cluster_id",
					},
					TenantId:  "tenant_id",
					NetworkId: "network-id-1",
					SubnetId:  "subnet-id-1",
				},
				{
					AgtPortReq: agtmgr.AgtPortReq{
						TenantID:    "tenant_id",
						NetworkName: "networkName2",
						PortName:    "portName2",
						NodeID:      "node_id",
						PodNs:       "pod_ns",
						PodName:     "pod_name1",
						ClusterID:   "cluster_id",
					},
					TenantId:  "tenant_id",
					NetworkId: "network-id-2",
					SubnetId:  "subnet-id-2",
				},
			},
		}

		inters := []*iaasaccessor.Interface{
			{
				Id:        "int-id1",
				Name:      "portName1",
				Ip:        "10.92.247.5",
				NetworkId: "network-id-1",
				SubnetId:  "subnet-id-1",
			},
			{
				Id:        "int-id2",
				Name:      "portName2",
				Ip:        "10.92.247.6",
				NetworkId: "network-id-2",
				SubnetId:  "subnet-id-2",
			},
		}

		cfgMock := gomock.NewController(t)
		defer cfgMock.Finish()
		mockIaas := test.NewMockIaaS(cfgMock)
		stubs := gostub.StubFunc(&iaas.GetIaaS, mockIaas)
		defer stubs.Reset()

		mockIaas.EXPECT().CreateBulkPorts(gomock.Any()).Return(inters, nil)

		nics, err := GetPortServiceObj().CreateBulkPorts(&mgrBulkPortsReq)
		So(err, ShouldEqual, nil)
		So(inters[0].SubnetId, ShouldEqual, "subnet-id-1")
		So(inters[1].SubnetId, ShouldEqual, "subnet-id-2")
		So(inters[0].NetworkId, ShouldEqual, "network-id-1")
		So(inters[1].NetworkId, ShouldEqual, "network-id-2")
		So(len(nics), ShouldEqual, 2)
		So(nics[0].Name, ShouldEqual, "portName1")
		So(nics[1].Name, ShouldEqual, "portName2")
		So(nics[0].Id, ShouldEqual, "int-id1")
		So(nics[1].Id, ShouldEqual, "int-id2")
	})
}

func TestCreateBulkPortsOKWithRetry(t *testing.T) {
	Convey("TestCreateBulkPortsOKWithRetry", t, func() {
		mgrBulkPortsReq := mgriaas.MgrBulkPortsReq{
			TranId: "123",
			Ports: []*mgriaas.MgrPortReq{
				{
					AgtPortReq: agtmgr.AgtPortReq{
						TenantID:    "tenant_id",
						NetworkName: "networkName1",
						PortName:    "portName1",
						NodeID:      "node_id",
						PodNs:       "pod_ns",
						PodName:     "pod_name1",
						ClusterID:   "cluster_id",
					},
					TenantId:  "tenant_id",
					NetworkId: "network-id-1",
					SubnetId:  "subnet-id-1",
				},
				{
					AgtPortReq: agtmgr.AgtPortReq{
						TenantID:    "tenant_id",
						NetworkName: "networkName2",
						PortName:    "portName2",
						NodeID:      "node_id",
						PodNs:       "pod_ns",
						PodName:     "pod_name1",
						ClusterID:   "cluster_id",
					},
					TenantId:  "tenant_id",
					NetworkId: "network-id-2",
					SubnetId:  "subnet-id-2",
				},
			},
		}

		inters := []*iaasaccessor.Interface{
			{
				Id:        "int-id1",
				Name:      "portName1",
				Ip:        "10.92.247.5",
				NetworkId: "network-id-1",
				SubnetId:  "subnet-id-1",
			},
			{
				Id:        "int-id2",
				Name:      "portName2",
				Ip:        "10.92.247.6",
				NetworkId: "network-id-2",
				SubnetId:  "subnet-id-2",
			},
		}

		cfgMock := gomock.NewController(t)
		defer cfgMock.Finish()
		mockIaas := test.NewMockIaaS(cfgMock)
		stubs := gostub.StubFunc(&iaas.GetIaaS, mockIaas)
		defer stubs.Reset()

		gomock.InOrder(
			mockIaas.EXPECT().CreateBulkPorts(gomock.Any()).Return(nil, errors.New("connection timeout")),
			mockIaas.EXPECT().CreateBulkPorts(gomock.Any()).Return(inters, nil),
		)

		nics, err := GetPortServiceObj().CreateBulkPorts(&mgrBulkPortsReq)
		So(err, ShouldEqual, nil)
		So(inters[0].SubnetId, ShouldEqual, "subnet-id-1")
		So(inters[1].SubnetId, ShouldEqual, "subnet-id-2")
		So(inters[0].NetworkId, ShouldEqual, "network-id-1")
		So(inters[1].NetworkId, ShouldEqual, "network-id-2")
		So(len(nics), ShouldEqual, 2)
		So(nics[0].Name, ShouldEqual, "portName1")
		So(nics[1].Name, ShouldEqual, "portName2")
		So(nics[0].Id, ShouldEqual, "int-id1")
		So(nics[1].Id, ShouldEqual, "int-id2")
	})
}

func TestCreateBulkPortsOKWithBlankReq(t *testing.T) {
	Convey("TestCreateBulkPortsOKWithBlankReq", t, func() {
		mgrBulkPortsReq := mgriaas.MgrBulkPortsReq{
			TranId: "123",
			Ports:  []*mgriaas.MgrPortReq{},
		}

		nics, err := GetPortServiceObj().CreateBulkPorts(&mgrBulkPortsReq)
		So(err, ShouldBeNil)
		So(nics, ShouldBeNil)
	})
}

func TestCreateBulkPortsError(t *testing.T) {
	Convey("TestCreateBulkPortsError", t, func() {
		mgrBulkPortsReq := mgriaas.MgrBulkPortsReq{
			TranId: "123",
			Ports: []*mgriaas.MgrPortReq{
				{
					AgtPortReq: agtmgr.AgtPortReq{
						TenantID:    "tenant_id",
						NetworkName: "networkName1",
						PortName:    "portName1",
						NodeID:      "node_id",
						PodNs:       "pod_ns",
						PodName:     "pod_name1",
						ClusterID:   "cluster_id",
					},
					TenantId:  "tenant_id",
					NetworkId: "network-id-1",
					SubnetId:  "subnet-id-1",
				},
				{
					AgtPortReq: agtmgr.AgtPortReq{
						TenantID:    "tenant_id",
						NetworkName: "networkName2",
						PortName:    "portName2",
						NodeID:      "node_id",
						PodNs:       "pod_ns",
						PodName:     "pod_name1",
						ClusterID:   "cluster_id",
					},
					TenantId:  "tenant_id",
					NetworkId: "network-id-2",
					SubnetId:  "subnet-id-2",
				},
			},
		}

		cfgMock := gomock.NewController(t)
		defer cfgMock.Finish()
		mockIaas := test.NewMockIaaS(cfgMock)
		stubs := gostub.StubFunc(&iaas.GetIaaS, mockIaas)
		defer stubs.Reset()

		errStr := "connection timeout"
		mockIaas.EXPECT().CreateBulkPorts(gomock.Any()).Return(nil, errors.New(errStr)).Times(3)

		nics, err := GetPortServiceObj().CreateBulkPorts(&mgrBulkPortsReq)
		So(err.Error(), ShouldEqual, errStr)
		So(nics, ShouldBeNil)
	})
}

func TestCreatePortOK0(t *testing.T) {
	Convey("TestCreatePortOK0", t, func() {
		mockCtl := gomock.NewController(t)
		defer mockCtl.Finish()
		mockIaas := test.NewMockIaaS(mockCtl)
		mockDB := mockdbaccessor.NewMockDbAccessor(mockCtl)
		common.SetDataBase(mockDB)
		iaas.SetIaaS(constvalue.DefaultIaasTenantID, mockIaas)

		stubs := gostub.StubFunc(&common.GetDataBase, mockDB)
		defer stubs.Reset()
		stubs.StubFunc(&iaas.GetIaaS, mockIaas)

		//nodes := []*client.Node{
		//	{Key: "/paasnet/tenants/TenantId/ipgroups/gid0"},
		//}
		//net := Net{Network:iaasaccessor.Network{Name:"net0"}, Subnet:iaasaccessor.Subnet{Id:"subnet0"}}
		//netByte, _ := json.Marshal(net)
		//igInDb := IPGroupInDB{Name:"ig0", IPs:[]IPInDB{
		//	{IPAddr:"1.1.1.1", PortID:"id0"},
		//	{IPAddr:"1.1.1.2", PortID:"id0"}}}
		//igByte, _ := json.Marshal(igInDb)
		networkID := "network-id"
		networkName := "network-name"
		subnetID := "subnet-id"
		portName := "port-name"
		portID := "port-id"

		monkey.Patch(GetNetworkByName, func(tid, netName string) (*PaasNetwork, error) {
			return &PaasNetwork{
				ID:       networkID,
				SubnetID: subnetID,
			}, nil
		})
		defer monkey.UnpatchAll()
		monkey.Patch(iaas.GetIaasTenantIDByPaasTenantID, func(paasTenantID string) (string, error) {
			return "iaasTenantID", nil
		})

		iaasPort := iaasaccessor.Interface{
			Id:        portID,
			Name:      portName,
			NetworkId: networkID,
			SubnetId:  subnetID,
		}
		mockIaas.EXPECT().CreatePort(networkID, subnetID, portName, "", "", "normal").Return(&iaasPort, nil)

		expResp := &mgragt.CreatePortResp{
			Port: mgragt.CreatePortInfo{
				PortID:    portID,
				Name:      portName,
				NetworkID: networkID,
			}}
		stubs.StubFunc(&AssembleResponse, expResp, nil)

		expPortObj := &PortObj{
			ID:        portID,
			Name:      portName,
			NetworkID: networkID,
			SubnetID:  subnetID,
			Status:    CreatedOK,
			OwnerType: constvalue.OwnerTypePod,
		}

		req := &CreatePortReq{
			AgtPortReq: agtmgr.AgtPortReq{
				NetworkName: networkName,
				PortName:    portName,
				VnicType:    "normal"}}
		portObj, resp, err := GetPortServiceObj().CreatePort("", req)
		So(err, ShouldBeNil)
		So(portObj, ShouldResemble, expPortObj)
		So(resp, ShouldResemble, expResp)
	})
}

func TestAttachPortToVMErr(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockDB := test.NewMockDbAccessor(ctrl)
	stubs := gostub.StubFunc(&common.GetDataBase, mockDB)
	defer stubs.Reset()
	mockIaas := test.NewMockIaaS(ctrl)
	stubs.StubFunc(&iaas.GetIaaS, mockIaas)
	iaas.SetIaaS(constvalue.DefaultIaasTenantID, mockIaas)
	req := &PortVMOpsReq{
		VMID:   "222222",
		PortID: "333333333333",
	}
	mockIaas.EXPECT().AttachPortToVM(gomock.Any(), gomock.Any()).Return(nil, errors.New("iaas attach err"))

	Convey("TestAttachPortToVMErr", t, func() {
		port := &PortOps{}
		_, err := port.AttachPortToVM("111111", mockIaas, req)
		So(err.Error(), ShouldContainSubstring, "iaas attach err")
	})
}
