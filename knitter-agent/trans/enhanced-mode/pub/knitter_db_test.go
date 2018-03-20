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

package modelsext

import (
	"testing"

	"errors"
	"github.com/ZTE/Knitter/knitter-agent/domain/cni"
	"github.com/ZTE/Knitter/knitter-agent/tests/mock/db-mock"
	"github.com/ZTE/Knitter/pkg/db-accessor"
	"github.com/ZTE/Knitter/pkg/iaas-accessor"
	"github.com/golang/mock/gomock"
	"github.com/smartystreets/goconvey/convey"
)

//todo: storeSaveInterface

func TestDeletePodFromLocalDB(t *testing.T) {
	mockCtl := gomock.NewController(t)
	defer mockCtl.Finish()
	mockDB := mockdbaccessor.NewMockDbAccessor(mockCtl)
	agtObj := &cni.AgentContext{
		DB:        mockDB,
		ClusterID: "cluster_id",
		HostIP:    "192.168.1.100"}

	cniParam := &cni.CniParam{
		TenantID: "paas-tenant-id",
		PodNs:    "pod-ns",
		PodName:  "pod-name"}

	keyPod := dbaccessor.GetKeyOfPod(cniParam.TenantID, cniParam.PodNs, cniParam.PodName)
	mockDB.EXPECT().DeleteDir(keyPod).Return(nil)

	keyCluster := dbaccessor.GetKeyOfPodForNode(agtObj.ClusterID, agtObj.HostIP, cniParam.PodNs, cniParam.PodName)
	mockDB.EXPECT().DeleteLeaf(keyCluster).Return(nil)

	convey.Convey("TestDeletePodFromLocalDB", t, func() {
		err := DeletePodFromLocalDB(agtObj, cniParam)
		convey.So(err, convey.ShouldBeNil)
	})
}

func TestDeletePodFromLocalDB_Failed(t *testing.T) {
	mockCtl := gomock.NewController(t)
	defer mockCtl.Finish()
	mockDB := mockdbaccessor.NewMockDbAccessor(mockCtl)
	agtObj := &cni.AgentContext{
		DB:        mockDB,
		ClusterID: "cluster_id",
		HostIP:    "192.168.1.100"}

	cniParam := &cni.CniParam{
		TenantID: "paas-tenant-id",
		PodNs:    "pod-ns",
		PodName:  "pod-name"}

	errObj := errors.New("file broken")
	keyPod := dbaccessor.GetKeyOfPod(cniParam.TenantID, cniParam.PodNs, cniParam.PodName)
	mockDB.EXPECT().DeleteDir(keyPod).Return(errObj)

	keyCluster := dbaccessor.GetKeyOfPodForNode(agtObj.ClusterID, agtObj.HostIP, cniParam.PodNs, cniParam.PodName)
	mockDB.EXPECT().DeleteLeaf(keyCluster).Return(errObj)

	convey.Convey("TestDeletePodFromLocalDB_Failed", t, func() {
		err := DeletePodFromLocalDB(agtObj, cniParam)
		convey.So(err, convey.ShouldBeNil)
	})
}

func TestDeletePortFromLocalDB(t *testing.T) {
	mockCtl := gomock.NewController(t)
	defer mockCtl.Finish()
	mockRemoteDB := mockdbaccessor.NewMockDbAccessor(mockCtl)
	mockDB := mockdbaccessor.NewMockDbAccessor(mockCtl)
	agtObj := &cni.AgentContext{DB: mockDB, RemoteDB: mockRemoteDB, ClusterID: "cluster_id", HostIP: "192.168.1.100"}

	port := iaasaccessor.Interface{
		Id:        "port-id",
		PodNs:     "pod-ns",
		PodName:   "pod-name",
		NetworkId: "network-id",
		TenantID:  "paas-tenant-id",
		NetPlane:  "eio",
	}

	interfaceID := port.Id + port.PodNs + port.PodName
	keyOfInterfaceInNetwork := dbaccessor.GetKeyOfInterfaceInNetwork(port.TenantID, port.NetworkId, interfaceID)
	mockDB.EXPECT().DeleteLeaf(keyOfInterfaceInNetwork).Return(nil)

	keyOfInterface := dbaccessor.GetKeyOfInterface(port.TenantID, interfaceID)
	mockDB.EXPECT().DeleteDir(keyOfInterface).Return(nil)

	urlPaasInterfaceForNode := dbaccessor.GetKeyOfPaasInterfaceForNode(agtObj.ClusterID, agtObj.HostIP, interfaceID)
	mockDB.EXPECT().DeleteLeaf(urlPaasInterfaceForNode).Return(nil)

	urlIaasEioInterfaceForNode := dbaccessor.GetKeyOfIaasEioInterfaceForNode(agtObj.ClusterID, agtObj.HostIP, port.Id)
	mockDB.EXPECT().DeleteLeaf(urlIaasEioInterfaceForNode).Return(nil)

	keyPortInPod := dbaccessor.GetKeyOfInterfaceInPod(port.TenantID,
		port.Id, port.PodNs, port.PodName)
	mockDB.EXPECT().DeleteLeaf(keyPortInPod).Return(nil)

	convey.Convey("TestDeletePortFromLocalDB", t, func() {
		err := DeletePortFromLocalDB(agtObj, port)
		convey.So(err, convey.ShouldBeNil)
	})
}

func TestDeletePortFromLocalDB_Failed(t *testing.T) {
	mockCtl := gomock.NewController(t)
	defer mockCtl.Finish()
	mockRemoteDB := mockdbaccessor.NewMockDbAccessor(mockCtl)
	mockDB := mockdbaccessor.NewMockDbAccessor(mockCtl)
	agtObj := &cni.AgentContext{DB: mockDB, RemoteDB: mockRemoteDB, ClusterID: "cluster_id", HostIP: "192.168.1.100"}

	port := iaasaccessor.Interface{
		Id:        "port-id",
		PodNs:     "pod-ns",
		PodName:   "pod-name",
		NetworkId: "network-id",
		TenantID:  "paas-tenant-id",
		NetPlane:  "eio",
	}

	errStr := "etcd cluster misconfig"
	errObj := errors.New(errStr)
	interfaceID := port.Id + port.PodNs + port.PodName
	keyOfInterfaceInNetwork := dbaccessor.GetKeyOfInterfaceInNetwork(port.TenantID, port.NetworkId, interfaceID)
	mockDB.EXPECT().DeleteLeaf(keyOfInterfaceInNetwork).Return(errObj)

	keyOfInterface := dbaccessor.GetKeyOfInterface(port.TenantID, interfaceID)
	mockDB.EXPECT().DeleteDir(keyOfInterface).Return(errObj)

	urlPaasInterfaceForNode := dbaccessor.GetKeyOfPaasInterfaceForNode(agtObj.ClusterID, agtObj.HostIP, interfaceID)
	mockDB.EXPECT().DeleteLeaf(urlPaasInterfaceForNode).Return(errObj)

	urlIaasEioInterfaceForNode := dbaccessor.GetKeyOfIaasEioInterfaceForNode(agtObj.ClusterID, agtObj.HostIP, port.Id)
	mockDB.EXPECT().DeleteLeaf(urlIaasEioInterfaceForNode).Return(errObj)

	keyPortInPod := dbaccessor.GetKeyOfInterfaceInPod(port.TenantID,
		port.Id, port.PodNs, port.PodName)
	mockDB.EXPECT().DeleteLeaf(keyPortInPod).Return(errObj)

	convey.Convey("TestDeletePortFromLocalDB_Failed", t, func() {
		err := DeletePortFromLocalDB(agtObj, port)
		convey.So(err, convey.ShouldBeNil)
	})
}
