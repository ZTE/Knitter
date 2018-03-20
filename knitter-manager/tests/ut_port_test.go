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

package test

import (
	"testing"

	"github.com/coreos/etcd/client"
	"github.com/golang/mock/gomock"
	. "github.com/smartystreets/goconvey/convey"

	"errors"
	"github.com/ZTE/Knitter/knitter-manager/iaas"
	"github.com/ZTE/Knitter/knitter-manager/public"
	_ "github.com/ZTE/Knitter/knitter-manager/routers"
	"github.com/ZTE/Knitter/pkg/iaas-accessor"
	"github.com/golang/gostub"
)

func TestApiCreatPortERR401(t *testing.T) {
	//cfgMock := gomock.NewController(t)
	//defer cfgMock.Finish()
	//mockDB := NewMockDbAccessor(cfgMock)
	//common.SetDataBase(mockDB)
	stubs := gostub.StubFunc(&iaas.GetIaaS, nil)
	defer stubs.Reset()
	resp := APICreatePort("the-uuid-of-port")
	Convey("TestApiCreatPortERR401\n", t, func() {
		So(resp.Code, ShouldEqual, 401)
	})
}

func TestApiCreatPortERR403A(t *testing.T) {
	cfgMock := gomock.NewController(t)
	defer cfgMock.Finish()
	//mockDB := NewMockDbAccessor(cfgMock)
	mockIaaS := NewMockIaaS(cfgMock)
	//common.SetDataBase(mockDB)
	stubs := gostub.StubFunc(&iaas.GetIaaS, mockIaaS)
	defer stubs.Reset()

	resp := APICreatePort("844d0d23-2d53-454a-93cf-73c8253f94d6")
	Convey("TestApiCreatPortERR403A\n", t, func() {
		So(resp.Code, ShouldEqual, 403)
	})
}

func TestApiCreatPortERR406B(t *testing.T) {
	cfgMock := gomock.NewController(t)
	defer cfgMock.Finish()
	//mockDB := NewMockDbAccessor(cfgMock)
	mockIaaS := NewMockIaaS(cfgMock)
	//common.SetDataBase(mockDB)
	stubs := gostub.StubFunc(&iaas.GetIaaS, mockIaaS)
	defer stubs.Reset()

	cfg := string(`{"network_name":"want-to-attach-network",
		"vnic_type":"normal"}`)
	resp := APICreatePort(cfg)
	Convey("TestApiCreatPortERR406B\n", t, func() {
		So(resp.Code, ShouldEqual, 406)
	})
}

func listAllNetwork(mockDB *MockDbAccessor) {
	var list []*client.Node
	node := client.Node{Key: "network-uuid", Value: "network-info"}
	list = append(list, &node)
	list = append(list, &node)
	mockDB.EXPECT().ReadDir(gomock.Any()).Return(list, nil)
	netInfo0 := string(`{
    "network": {
        "cidr": "123.124.125.0/24",
        "gateway": "123.124.125.1",
        "name": "want-to-attach-network",
        "network_id": "9a4bf247-0cdd-4649-a5aa-4255856d3da2"
    }}`)
	netInfo1 := string(`{
    "network": {
        "cidr": "123.124.125.0/24",
        "gateway": "123.124.125.1",
        "name": "network_not_show",
        "network_id": "9a4bf247-0cdd-4649-a5aa-4255856d3da2"
    }}`)
	mockDB.EXPECT().ReadLeaf(gomock.Any()).Return(netInfo0, nil)
	mockDB.EXPECT().ReadLeaf(gomock.Any()).Return(netInfo1, nil)
	mockDB.EXPECT().ReadDir(gomock.Any()).Return(list, nil)
	mockDB.EXPECT().ReadLeaf(gomock.Any()).Return(netInfo0, nil)
	mockDB.EXPECT().ReadLeaf(gomock.Any()).Return(netInfo1, nil)
}

func listAllNetwork2(mockDB *MockDbAccessor) {
	var list []*client.Node
	node := client.Node{Key: "network-uuid", Value: "network-info"}
	list = append(list, &node)
	list = append(list, &node)
	mockDB.EXPECT().ReadDir(gomock.Any()).Return(list, nil)
	netInfo0 := string(`{
    "network": {
        "cidr": "123.124.125.0/24",
        "gateway": "123.124.125.1",
        "name": "want-to-attach-network",
        "network_id": "9a4bf247-0cdd-4649-a5aa-4255856d3da2"
    }}`)
	netInfo1 := string(`{
    "network": {
        "cidr": "123.124.125.0/24",
        "gateway": "123.124.125.1",
        "name": "network_not_show",
        "network_id": "9a4bf247-0cdd-4649-a5aa-4255856d3da2"
    }}`)
	mockDB.EXPECT().ReadLeaf(gomock.Any()).Return(netInfo0, nil)
	mockDB.EXPECT().ReadLeaf(gomock.Any()).Return(netInfo1, nil)
	mockDB.EXPECT().ReadDir(gomock.Any()).Return(list, nil)
	mockDB.EXPECT().ReadLeaf(gomock.Any()).Return(netInfo0, nil)
	mockDB.EXPECT().ReadLeaf(gomock.Any()).Return(netInfo1, nil)

	mockDB.EXPECT().ReadDir(gomock.Any()).Return(list, nil)
	mockDB.EXPECT().ReadLeaf(gomock.Any()).Return(netInfo0, nil)
	mockDB.EXPECT().ReadLeaf(gomock.Any()).Return(netInfo1, nil)
	mockDB.EXPECT().ReadDir(gomock.Any()).Return(list, nil)
	mockDB.EXPECT().ReadLeaf(gomock.Any()).Return(netInfo0, nil)
	mockDB.EXPECT().ReadLeaf(gomock.Any()).Return(netInfo1, nil)

	mockDB.EXPECT().ReadDir(gomock.Any()).Return(list, nil)
	mockDB.EXPECT().ReadLeaf(gomock.Any()).Return(netInfo0, nil)
	mockDB.EXPECT().ReadLeaf(gomock.Any()).Return(netInfo1, nil)
	mockDB.EXPECT().ReadDir(gomock.Any()).Return(list, nil)
	mockDB.EXPECT().ReadLeaf(gomock.Any()).Return(netInfo0, nil)
	mockDB.EXPECT().ReadLeaf(gomock.Any()).Return(netInfo1, nil)
}

func TestApiDeletePortERR401(t *testing.T) {
	cfgMock := gomock.NewController(t)
	defer cfgMock.Finish()
	//mockDB := NewMockDbAccessor(cfgMock)
	//common.SetDataBase(mockDB)
	stubs := gostub.StubFunc(&iaas.GetIaaS, nil)
	defer stubs.Reset()

	resp := APIDeletePort("the-name-of-network")
	Convey("TestApiDeletePortERR401\n", t, func() {
		So(resp.Code, ShouldEqual, 401)
	})
}

func TestApiAttachERR401(t *testing.T) {
	cfgMock := gomock.NewController(t)
	defer cfgMock.Finish()
	//mockDB := NewMockDbAccessor(cfgMock)
	//common.SetDataBase(mockDB)
	stubs := gostub.StubFunc(&iaas.GetIaaS, nil)
	defer stubs.Reset()

	resp := APIAttach("port-uuid", "vm-uuid")
	Convey("TestApiDeletePortERR401\n", t, func() {
		So(resp.Code, ShouldEqual, 401)
	})
}

func TestApiAttachERR404A(t *testing.T) {
	cfgMock := gomock.NewController(t)
	defer cfgMock.Finish()
	mockDB := NewMockDbAccessor(cfgMock)
	mockIaaS := NewMockIaaS(cfgMock)
	common.SetDataBase(mockDB)
	stubs := gostub.StubFunc(&iaas.GetIaaS, mockIaaS)
	defer stubs.Reset()

	mockIaaS.EXPECT().AttachPortToVM(
		gomock.Any(), gomock.Any()).Return(nil, errors.New("error"))
	mockDB.EXPECT().SaveLeaf(gomock.Any(), gomock.Any())

	resp := APIAttach("port-uuid", "vm-uuid")
	Convey("TestApiAttachERR404A\n", t, func() {
		So(resp.Code, ShouldEqual, 406)
	})
}

func TestApiAttachERR404B(t *testing.T) {
	cfgMock := gomock.NewController(t)
	defer cfgMock.Finish()
	mockDB := NewMockDbAccessor(cfgMock)
	mockIaaS := NewMockIaaS(cfgMock)
	common.SetDataBase(mockDB)
	stubs := gostub.StubFunc(&iaas.GetIaaS, mockIaaS)
	defer stubs.Reset()

	mockIaaS.EXPECT().AttachPortToVM(
		gomock.Any(), gomock.Any()).Return(nil, errors.New("error"))
	mockDB.EXPECT().SaveLeaf(gomock.Any(), gomock.Any()).Return(nil)

	resp := APIAttach("port-uuid", "vm-uuid")
	Convey("TestApiAttachERR404A\n", t, func() {
		So(resp.Code, ShouldEqual, 406)
	})
}

func TestApiAttachERR404C(t *testing.T) {
	cfgMock := gomock.NewController(t)
	defer cfgMock.Finish()
	mockDB := NewMockDbAccessor(cfgMock)
	mockIaaS := NewMockIaaS(cfgMock)
	common.SetDataBase(mockDB)
	stubs := gostub.StubFunc(&iaas.GetIaaS, mockIaaS)
	defer stubs.Reset()

	mockIaaS.EXPECT().AttachPortToVM(gomock.Any(), gomock.Any()).Return(nil, nil)
	mockIaaS.EXPECT().GetType().Return("TECS")
	mockIaaS.EXPECT().GetPort(gomock.Any()).Return(nil, errors.New("error"))
	mockIaaS.EXPECT().DetachPortFromVM(gomock.Any(), gomock.Any()).Return(nil)

	resp := APIAttach("port-uuid", "vm-uuid")
	Convey("TestApiAttachERR404A\n", t, func() {
		So(resp.Code, ShouldEqual, 406)
	})
}

func TestApiAttachERR404D(t *testing.T) {
	cfgMock := gomock.NewController(t)
	defer cfgMock.Finish()
	mockDB := NewMockDbAccessor(cfgMock)
	mockIaaS := NewMockIaaS(cfgMock)
	common.SetDataBase(mockDB)
	stubs := gostub.StubFunc(&iaas.GetIaaS, mockIaaS)
	defer stubs.Reset()

	mockIaaS.EXPECT().AttachPortToVM(gomock.Any(), gomock.Any()).Return(nil, nil)
	mockIaaS.EXPECT().GetType().Return("TECS")
	mockIaaS.EXPECT().GetPort(gomock.Any()).Return(nil, errors.New("error"))
	mockIaaS.EXPECT().DetachPortFromVM(gomock.Any(), gomock.Any()).Return(errors.New("error"))
	mockDB.EXPECT().SaveLeaf(gomock.Any(), gomock.Any()).Return(nil)

	resp := APIAttach("port-uuid", "vm-uuid")
	Convey("TestApiAttachERR404A\n", t, func() {
		So(resp.Code, ShouldEqual, 406)
	})
}

func TestApiAttachOK(t *testing.T) {
	cfgMock := gomock.NewController(t)
	defer cfgMock.Finish()
	mockDB := NewMockDbAccessor(cfgMock)
	mockIaaS := NewMockIaaS(cfgMock)
	common.SetDataBase(mockDB)
	stubs := gostub.StubFunc(&iaas.GetIaaS, mockIaaS)
	defer stubs.Reset()

	mockIaaS.EXPECT().AttachPortToVM(gomock.Any(), gomock.Any()).Return(nil, nil)
	mockIaaS.EXPECT().GetType().Return("TECS")
	port := iaasaccessor.Interface{Status: "ACTIVE"}
	mockIaaS.EXPECT().GetPort(gomock.Any()).Return(&port, nil)

	resp := APIAttach("port-uuid", "vm-uuid")
	Convey("TestApiAttachERR404A\n", t, func() {
		So(resp.Code, ShouldEqual, 200)
	})
}

func TestApiAttachVNFMOK(t *testing.T) {
	cfgMock := gomock.NewController(t)
	defer cfgMock.Finish()
	mockDB := NewMockDbAccessor(cfgMock)
	mockIaaS := NewMockIaaS(cfgMock)
	common.SetDataBase(mockDB)
	stubs := gostub.StubFunc(&iaas.GetIaaS, mockIaaS)
	defer stubs.Reset()

	mockIaaS.EXPECT().AttachPortToVM(gomock.Any(), gomock.Any()).Return(nil, nil)
	mockIaaS.EXPECT().GetType().Return("VNFM")

	resp := APIAttach("port-uuid", "vm-uuid")
	Convey("TestApiAttachVNFMOK\n", t, func() {
		So(resp.Code, ShouldEqual, 200)
	})
}

func TestApiDetachERR401(t *testing.T) {
	cfgMock := gomock.NewController(t)
	defer cfgMock.Finish()
	//mockDB := NewMockDbAccessor(cfgMock)
	//common.SetDataBase(mockDB)
	stubs := gostub.StubFunc(&iaas.GetIaaS, nil)
	defer stubs.Reset()

	resp := APIDetach("port-uuid", "vm-uuid")
	Convey("TestApiDetachERR401\n", t, func() {
		So(resp.Code, ShouldEqual, 401)
	})
}

func TestApiDetachERR404A(t *testing.T) {
	cfgMock := gomock.NewController(t)
	defer cfgMock.Finish()
	mockDB := NewMockDbAccessor(cfgMock)
	mockIaaS := NewMockIaaS(cfgMock)
	common.SetDataBase(mockDB)
	stubs := gostub.StubFunc(&iaas.GetIaaS, mockIaaS)
	defer stubs.Reset()

	mockIaaS.EXPECT().DetachPortFromVM(
		gomock.Any(), gomock.Any()).Return(errors.New("error")).Times(3)
	mockDB.EXPECT().SaveLeaf(gomock.Any(), gomock.Any()).Return(nil)

	resp := APIDetach("port-uuid", "vm-uuid")
	Convey("TestApiAttachERR404A\n", t, func() {
		So(resp.Code, ShouldEqual, 406)
	})
}

func TestApiDetachERR404B(t *testing.T) {
	cfgMock := gomock.NewController(t)
	defer cfgMock.Finish()
	mockDB := NewMockDbAccessor(cfgMock)
	mockIaaS := NewMockIaaS(cfgMock)
	common.SetDataBase(mockDB)
	stubs := gostub.StubFunc(&iaas.GetIaaS, mockIaaS)
	defer stubs.Reset()

	mockIaaS.EXPECT().DetachPortFromVM(gomock.Any(), gomock.Any()).Return(nil)
	mockIaaS.EXPECT().GetType().Return("TECS")
	mockIaaS.EXPECT().GetPort(gomock.Any()).Return(nil, errors.New("error"))
	mockDB.EXPECT().SaveLeaf(gomock.Any(), gomock.Any()).Return(nil)

	resp := APIDetach("port-uuid", "vm-uuid")
	Convey("TestApiAttachERR404A\n", t, func() {
		So(resp.Code, ShouldEqual, 406)
	})
}

func TestApiDetachOK(t *testing.T) {
	cfgMock := gomock.NewController(t)
	defer cfgMock.Finish()
	mockDB := NewMockDbAccessor(cfgMock)
	mockIaaS := NewMockIaaS(cfgMock)
	common.SetDataBase(mockDB)
	stubs := gostub.StubFunc(&iaas.GetIaaS, mockIaaS)
	defer stubs.Reset()

	mockIaaS.EXPECT().DetachPortFromVM(gomock.Any(), gomock.Any()).Return(nil)
	mockIaaS.EXPECT().GetType().Return("TECS")
	port2 := iaasaccessor.Interface{Status: "DOWN"}
	mockIaaS.EXPECT().GetPort(gomock.Any()).Return(&port2, nil)
	resp := APIDetach("port-uuid", "vm-uuid")
	Convey("TestApiAttachERR404A\n", t, func() {
		So(resp.Code, ShouldEqual, 200)
	})
}

func TestApiDetachVNFMOK(t *testing.T) {
	cfgMock := gomock.NewController(t)
	defer cfgMock.Finish()
	mockDB := NewMockDbAccessor(cfgMock)
	mockIaaS := NewMockIaaS(cfgMock)
	common.SetDataBase(mockDB)
	stubs := gostub.StubFunc(&iaas.GetIaaS, mockIaaS)
	defer stubs.Reset()

	mockIaaS.EXPECT().DetachPortFromVM(gomock.Any(), gomock.Any()).Return(nil)
	mockIaaS.EXPECT().GetType().Return("VNFM")
	resp := APIDetach("port-uuid", "vm-uuid")
	Convey("TestApiAttachERR404A\n", t, func() {
		So(resp.Code, ShouldEqual, 200)
	})
}
