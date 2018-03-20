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
	"errors"
	"testing"

	"github.com/coreos/etcd/client"
	"github.com/golang/mock/gomock"
	. "github.com/smartystreets/goconvey/convey"

	"github.com/ZTE/Knitter/knitter-manager/iaas"
	"github.com/ZTE/Knitter/knitter-manager/models"
	"github.com/ZTE/Knitter/knitter-manager/public"
	_ "github.com/ZTE/Knitter/knitter-manager/routers"
	"github.com/bouk/monkey"
	"github.com/golang/gostub"
	"reflect"
)

func TestGetPodERR401(t *testing.T) {
	cfgMock := gomock.NewController(t)
	defer cfgMock.Finish()
	mockDB := NewMockDbAccessor(cfgMock)
	common.SetDataBase(mockDB)
	//common.SetIaaS(constvalue.DefaultIaasTenantID, nil)
	//monkey.Patch(iaas.GetIaasTenantIDByPaasTenantID,
	//	func(_ string) (string, error) {
	//		return constvalue.DefaultIaasTenantID, nil
	//	})
	//defer monkey.UnpatchAll()

	MockPaasAdminCheck(mockDB)
	stubs := gostub.StubFunc(&iaas.GetIaaS, nil)
	defer stubs.Reset()
	resp := GetPod("the-uuid-of-port")
	Convey("TestGetPodERR401\n", t, func() {
		So(resp.Code, ShouldEqual, 401)
	})
}

func TestGetPodERR404A(t *testing.T) {
	cfgMock := gomock.NewController(t)
	defer cfgMock.Finish()
	mockDB := NewMockDbAccessor(cfgMock)
	mockIaaS := NewMockIaaS(cfgMock)
	common.SetDataBase(mockDB)
	//common.SetIaaS(constvalue.DefaultIaasTenantID, mockIaaS)

	MockPaasAdminCheck(mockDB)
	stubs := gostub.StubFunc(&iaas.GetIaaS, mockIaaS)
	defer stubs.Reset()
	mockDB.EXPECT().ReadLeaf(gomock.Any()).Return(
		"", errors.New("read-network-from-etcd-error"))

	resp := GetPod("844d0d23-2d53-454a-93cf-73c8253f94d6")
	Convey("TestGetInterfaceERR404B\n", t, func() {
		So(resp.Code, ShouldEqual, 404)
	})
}

func TestGetPodOK200B(t *testing.T) {
	cfgMock := gomock.NewController(t)
	defer cfgMock.Finish()
	mockDB := NewMockDbAccessor(cfgMock)
	mockIaaS := NewMockIaaS(cfgMock)
	common.SetDataBase(mockDB)
	//common.SetIaaS(constvalue.DefaultIaasTenantID, mockIaaS)

	pod1 := &models.PodForResponse{
		Name: "11111111",
	}

	encapPodForResponse := &models.EncapPodForResponse{
		Pod: pod1,
	}
	var logicPod *models.LogicPod
	guard := monkey.PatchInstanceMethod(reflect.TypeOf(logicPod), "GetEncapPodForResponse",
		func(_ *models.LogicPod) (*models.EncapPodForResponse, error) {
			return encapPodForResponse, nil

		})
	defer guard.Unpatch()
	MockPaasAdminCheck(mockDB)
	stubs := gostub.StubFunc(&iaas.GetIaaS, mockIaaS)
	defer stubs.Reset()
	cfg := string(`{"name":"right-pod-name"}`)
	mockDB.EXPECT().ReadLeaf(gomock.Any()).Return(cfg, nil)
	mockDB.EXPECT().ReadDir(gomock.Any()).Return(nil, errors.New("error"))

	resp := GetPod("844d0d23-2d53-454a-93cf-73c8253f94d6")
	Convey("TestGetPodOK200B\n", t, func() {
		So(resp.Code, ShouldEqual, 200)
	})
}

func TestGetPodOK(t *testing.T) {
	cfgMock := gomock.NewController(t)
	defer cfgMock.Finish()
	mockDB := NewMockDbAccessor(cfgMock)
	mockIaaS := NewMockIaaS(cfgMock)
	common.SetDataBase(mockDB)
	//common.SetIaaS(constvalue.DefaultIaasTenantID , mockIaaS)

	MockPaasAdminCheck(mockDB)
	stubs := gostub.StubFunc(&iaas.GetIaaS, mockIaaS)
	defer stubs.Reset()
	cfg := string(`{"name":"right-pod-name"}`)
	mockDB.EXPECT().ReadLeaf(gomock.Any()).Return(cfg, nil)
	podCfg := string(`{"net_plane_type":"std","net_plane_name":"wan",
		"ip":"192.168.1.1"}`)
	node1 := client.Node{Key: "network-uuid", Value: "network-info"}
	node2 := client.Node{Key: "network-uuid", Value: "self-key-for-port-in-pod"}
	var list []*client.Node
	list = append(list, &node1)
	list = append(list, &node2)
	mockDB.EXPECT().ReadDir(gomock.Any()).Return(list, nil)
	mockDB.EXPECT().ReadLeaf(gomock.Any()).Return(podCfg, nil)
	mockDB.EXPECT().ReadLeaf(gomock.Any()).Return(podCfg, nil)

	pod1 := &models.PodForResponse{
		Name: "11111111",
	}

	encapPodForResponse := &models.EncapPodForResponse{
		Pod: pod1,
	}
	var logicPod *models.LogicPod
	guard := monkey.PatchInstanceMethod(reflect.TypeOf(logicPod), "GetEncapPodForResponse",
		func(_ *models.LogicPod) (*models.EncapPodForResponse, error) {
			return encapPodForResponse, nil

		})
	defer guard.Unpatch()
	resp := GetPod("844d0d23-2d53-454a-93cf-73c8253f94d6")
	Convey("TestGetInterfaceERR404B\n", t, func() {
		So(resp.Code, ShouldEqual, 200)
	})
}
func TestGetAllPodERR401(t *testing.T) {
	cfgMock := gomock.NewController(t)
	defer cfgMock.Finish()
	mockDB := NewMockDbAccessor(cfgMock)
	common.SetDataBase(mockDB)
	//common.SetIaaS(constvalue.DefaultIaasTenantID, nil)
	//pod1 := &models.PodForResponse{
	//	Name:"11111111",
	//}
	//pods := []*models.PodForResponse{pod1}
	//encapPodsForResponse := &models.EncapPodsForResponse{
	//	Pods:pods,
	//}
	//var logicPodManager *models.LogicPodManager
	//guard := monkey.PatchInstanceMethod(reflect.TypeOf(logicPodManager), "GetEncapPodsForResponse",
	//	func(_ *models.LogicPodManager, _ string)(*models.EncapPodsForResponse, error) {
	//		return encapPodsForResponse, nil
	//
	//	})
	//defer guard.Unpatch()
	MockPaasAdminCheck(mockDB)
	stubs := gostub.StubFunc(&iaas.GetIaaS, nil)
	defer stubs.Reset()
	resp := GetAllPod()
	Convey("TestGetAllPodERR401\n", t, func() {
		So(resp.Code, ShouldEqual, 401)
	})
}

func TestGetAllPodOK1(t *testing.T) {
	cfgMock := gomock.NewController(t)
	defer cfgMock.Finish()
	mockDB := NewMockDbAccessor(cfgMock)
	mockIaaS := NewMockIaaS(cfgMock)
	common.SetDataBase(mockDB)
	//common.SetIaaS(constvalue.DefaultIaasTenantID, mockIaaS)

	MockPaasAdminCheck(mockDB)
	stubs := gostub.StubFunc(&iaas.GetIaaS, mockIaaS)
	defer stubs.Reset()
	mockDB.EXPECT().ReadDir(gomock.Any()).Return(nil, errors.New("error"))

	resp := GetAllPod()
	Convey("TestGetAllPodOK1\n", t, func() {
		So(resp.Code, ShouldEqual, 200)
	})
}

func TestGetAllPodOK4(t *testing.T) {
	cfgMock := gomock.NewController(t)
	defer cfgMock.Finish()
	mockDB := NewMockDbAccessor(cfgMock)
	mockIaaS := NewMockIaaS(cfgMock)
	common.SetDataBase(mockDB)
	//common.SetIaaS(constvalue.DefaultIaasTenantID, mockIaaS)

	MockPaasAdminCheck(mockDB)
	stubs := gostub.StubFunc(&iaas.GetIaaS, mockIaaS)
	defer stubs.Reset()
	var nsList []*client.Node
	ns0 := client.Node{Key: "ns-uuid0", Value: "info0"}
	nsList = append(nsList, &ns0)
	gomock.InOrder(
		mockDB.EXPECT().ReadDir(gomock.Any()).Return(nsList, nil),
		mockDB.EXPECT().ReadDir(gomock.Any()).Return(nil, errors.New("error0")),
	)

	resp := GetAllPod()
	Convey("TestGetAllNetworkOK3\n", t, func() {
		So(resp.Code, ShouldEqual, 200)
	})
}

func TestGetAllPodOK5(t *testing.T) {
	cfgMock := gomock.NewController(t)
	defer cfgMock.Finish()
	mockDB := NewMockDbAccessor(cfgMock)
	mockIaaS := NewMockIaaS(cfgMock)
	common.SetDataBase(mockDB)
	//common.SetIaaS(constvalue.DefaultIaasTenantID, mockIaaS)

	MockPaasAdminCheck(mockDB)
	stubs := gostub.StubFunc(&iaas.GetIaaS, mockIaaS)
	defer stubs.Reset()
	var nsList, nameList []*client.Node
	ns0 := client.Node{Key: "ns-uuid0", Value: "info0"}
	nsList = append(nsList, &ns0)
	mockDB.EXPECT().ReadDir(gomock.Any()).Return(nsList, nil)
	name0 := client.Node{Key: "name-uuid0", Value: "info0"}
	nameList = append(nameList, &name0)
	mockDB.EXPECT().ReadDir(gomock.Any()).Return(nameList, nil)
	mockDB.EXPECT().ReadLeaf(gomock.Any()).Return("", errors.New("error1"))

	resp := GetAllPod()
	Convey("TestGetAllNetworkOK3\n", t, func() {
		So(resp.Code, ShouldEqual, 200)
	})
}

func TestGetAllPodOK6(t *testing.T) {
	cfgMock := gomock.NewController(t)
	defer cfgMock.Finish()
	mockDB := NewMockDbAccessor(cfgMock)
	mockIaaS := NewMockIaaS(cfgMock)
	common.SetDataBase(mockDB)
	//common.SetIaaS(constvalue.DefaultIaasTenantID, mockIaaS)

	MockPaasAdminCheck(mockDB)
	stubs := gostub.StubFunc(&iaas.GetIaaS, mockIaaS)
	defer stubs.Reset()
	var nsList, nameList []*client.Node
	ns0 := client.Node{Key: "ns-uuid0", Value: "info0"}
	nsList = append(nsList, &ns0)
	mockDB.EXPECT().ReadDir(gomock.Any()).Return(nsList, nil)

	name0 := client.Node{Key: "name-uuid0", Value: "info0"}
	nameList = append(nameList, &name0)
	mockDB.EXPECT().ReadDir(gomock.Any()).Return(nameList, nil)

	cfg := string(`{"name":"right-pod-name"}`)
	mockDB.EXPECT().ReadLeaf(gomock.Any()).Return(cfg, nil)
	podCfg := string(`{"net_plane_type":"std","net_plane_name":"wan",
		"ip":"192.168.1.1"}`)
	node2 := client.Node{Key: "port-uuid2", Value: "self-key-for-port-in-pod"}
	var list1 []*client.Node
	list1 = append(list1, &node2)
	mockDB.EXPECT().ReadDir(gomock.Any()).Return(list1, nil)
	mockDB.EXPECT().ReadLeaf(gomock.Any()).Return(podCfg, nil)

	resp := GetAllPod()
	Convey("TestGetAllPodOK2\n", t, func() {
		So(resp.Code, ShouldEqual, 200)
	})
}

func TestGetAllPodOK2(t *testing.T) {
	cfgMock := gomock.NewController(t)
	defer cfgMock.Finish()
	mockDB := NewMockDbAccessor(cfgMock)
	mockIaaS := NewMockIaaS(cfgMock)
	common.SetDataBase(mockDB)
	//common.SetIaaS(constvalue.DefaultIaasTenantID, mockIaaS)

	MockPaasAdminCheck(mockDB)
	stubs := gostub.StubFunc(&iaas.GetIaaS, mockIaaS)
	defer stubs.Reset()
	var nsList, nameList []*client.Node
	ns0 := client.Node{Key: "ns-uuid0", Value: "info0"}
	nsList = append(nsList, &ns0)
	name0 := client.Node{Key: "name-uuid0", Value: "info0"}
	nameList = append(nameList, &name0)
	cfg := string(`{"name":"right-pod-name"}`)
	node2 := client.Node{Key: "port-uuid2", Value: "self-key-for-port-in-pod"}
	var list1 []*client.Node
	list1 = append(list1, &node2)

	gomock.InOrder(
		mockDB.EXPECT().ReadDir(gomock.Any()).Return(nsList, nil),
		mockDB.EXPECT().ReadDir(gomock.Any()).Return(nameList, nil),
		mockDB.EXPECT().ReadLeaf(gomock.Any()).Return(cfg, nil),
		mockDB.EXPECT().ReadDir(gomock.Any()).Return(list1, nil),
		mockDB.EXPECT().ReadLeaf(gomock.Any()).Return("", errors.New("read-prot-error")),
	)
	pod1 := &models.PodForResponse{
		Name: "11111111",
	}
	pods := []*models.PodForResponse{pod1}
	encapPodsForResponse := &models.EncapPodsForResponse{
		Pods: pods,
	}
	var logicPodManager *models.LogicPodManager
	guard := monkey.PatchInstanceMethod(reflect.TypeOf(logicPodManager), "GetEncapPodsForResponse",
		func(_ *models.LogicPodManager, _ string) (*models.EncapPodsForResponse, error) {
			return encapPodsForResponse, nil

		})
	defer guard.Unpatch()
	resp := GetAllPod()
	Convey("TestGetAllNetworkOK3\n", t, func() {
		So(resp.Code, ShouldEqual, 200)
	})
}
