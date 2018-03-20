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
	"github.com/ZTE/Knitter/knitter-agent/err-obj"
	"github.com/ZTE/Knitter/knitter-manager/public"
	"github.com/ZTE/Knitter/knitter-manager/tests/mock/db-mock"
	"github.com/ZTE/Knitter/pkg/db-accessor"
	. "github.com/bouk/monkey"
	"github.com/coreos/etcd/client"
	"github.com/golang/mock/gomock"
	. "github.com/smartystreets/goconvey/convey"
	"reflect"
	"testing"
)

//init input method
func MakeLogicPod1() *LogicPod {
	port1 := PortInfo{"port1", "control", "control", "10.92.247.1"}
	port2 := PortInfo{"port2", "media", "media", "10.92.247.2"}
	ports1 := []PortInfo{port1, port2}
	pod1 := &LogicPod{"9854396d-8e7a-43e5-9666-f95bdaa04793-1-11111", "admin", "admin", ports1}
	return pod1
}

func MakeLogicPod2() *LogicPod {
	port3 := PortInfo{"port3", "media", "media", "10.92.247.3"}
	port4 := PortInfo{"port4", "media", "media", "10.92.247.4"}
	ports2 := []PortInfo{port3, port4}

	pod2 := &LogicPod{"1d521358-f523-4f5b-970d-416aece0a2fc-1-22222", "admin", "admin", ports2}
	return pod2
}

func MakeLogicPod1TenantIDIsNil() *LogicPod {
	pod1TenantIDIsNil := MakeLogicPod1()
	pod1TenantIDIsNil.TenantID = ""
	return pod1TenantIDIsNil
}

func MakePodForResponse1() *PodForResponse {
	portForResponse1 := PortForResponse{"control", "control", "10.92.247.1"}
	portForResponse2 := PortForResponse{"media", "media", "10.92.247.2"}
	portsForResponse1 := []*PortForResponse{&portForResponse1, &portForResponse2}
	podForResponse1 := &PodForResponse{
		PodIps:     portsForResponse1,
		Name:       "9854396d-8e7a-43e5-9666-f95bdaa04793-1-11111",
		tenantUUID: "admin",
	}
	return podForResponse1
}

func MakePodForResponse2() *PodForResponse {
	portForResponse3 := PortForResponse{"media", "media", "10.92.247.3"}
	portForResponse4 := PortForResponse{"media", "media", "10.92.247.4"}

	portsForResponse2 := []*PortForResponse{&portForResponse3, &portForResponse4}

	podForResponse2 := &PodForResponse{
		PodIps:     portsForResponse2,
		Name:       "1d521358-f523-4f5b-970d-416aece0a2fc-1-22222",
		tenantUUID: "admin",
	}
	return podForResponse2
}

func TestSaveSucc(t *testing.T) {

	pod1 := MakeLogicPod1()

	ctl := gomock.NewController(t)
	defer ctl.Finish()
	dbmock := mockdbaccessor.NewMockDbAccessor(ctl)

	guard := Patch(common.GetDataBase, func() dbaccessor.DbAccessor {
		return dbmock
	})
	defer guard.Unpatch()

	podBodyExpect, _ := json.Marshal(pod1)
	key := "/knitter/manager/pods/admin/9854396d-8e7a-43e5-9666-f95bdaa04793-1-11111"
	dbmock.EXPECT().SaveLeaf(key, string(podBodyExpect)).Return(nil)

	Convey("TestSaveSucc ", t, func() {
		err := pod1.Save()
		So(err, ShouldBeNil)
	})
}

func TestSaveTenantIDNilFail(t *testing.T) {
	pod1TenantIDIsNil := MakeLogicPod1TenantIDIsNil()
	Convey("TestSaveTenantIDNilFail ", t, func() {
		err := pod1TenantIDIsNil.Save()
		So(err, ShouldEqual, errobj.ErrTenantsIDOrPodNameIsNil)
	})
}

func TestSaveMarshalFail(t *testing.T) {

	pod1 := MakeLogicPod1()
	Convey("TestGetEncapPodsForResponse SUCC", t, func() {
		guard := Patch(json.Marshal, func(_ interface{}) ([]byte, error) {
			return nil, errobj.ErrUnmarshalFailed
		})
		defer guard.Unpatch()
		err := pod1.Save()
		So(err, ShouldResemble, errobj.ErrUnmarshalFailed)
	})
}

func TestSaveLeafFail(t *testing.T) {
	pod1 := MakeLogicPod1()

	ctl := gomock.NewController(t)
	defer ctl.Finish()
	dbmock := mockdbaccessor.NewMockDbAccessor(ctl)

	guard := Patch(common.GetDataBase, func() dbaccessor.DbAccessor {
		return dbmock
	})
	defer guard.Unpatch()

	podBodyExpect, _ := json.Marshal(pod1)
	key := "/knitter/manager/pods/admin/9854396d-8e7a-43e5-9666-f95bdaa04793-1-11111"
	errStr := "save leaf error"
	dbmock.EXPECT().SaveLeaf(key, string(podBodyExpect)).Return(errors.New(errStr))

	Convey("TestSaveLeafFail ", t, func() {
		err := pod1.Save()
		So(err.Error(), ShouldEqual, errStr)
	})
}

func TestDeleteSucc(t *testing.T) {

	pod1 := MakeLogicPod1()

	ctl := gomock.NewController(t)
	defer ctl.Finish()
	dbmock := mockdbaccessor.NewMockDbAccessor(ctl)

	guard := Patch(common.GetDataBase, func() dbaccessor.DbAccessor {
		return dbmock
	})
	defer guard.Unpatch()

	key := "/knitter/manager/pods/admin/9854396d-8e7a-43e5-9666-f95bdaa04793-1-11111"
	dbmock.EXPECT().DeleteLeaf(key).Return(nil)

	Convey("TestDeleteSucc", t, func() {
		err := pod1.Delete()
		So(err, ShouldBeNil)
	})
}

func TestDeleteTenantIdNilFail(t *testing.T) {

	pod1TenantIDIsNil := MakeLogicPod1TenantIDIsNil()
	Convey("TestDeleteTenantIdNilFail ", t, func() {
		err := pod1TenantIDIsNil.Delete()
		So(err, ShouldEqual, errobj.ErrTenantsIDOrPodNameIsNil)
	})
}

func TestDeleteLeafFail(t *testing.T) {

	pod1 := MakeLogicPod1()
	ctl := gomock.NewController(t)
	defer ctl.Finish()
	dbmock := mockdbaccessor.NewMockDbAccessor(ctl)

	guard := Patch(common.GetDataBase, func() dbaccessor.DbAccessor {
		return dbmock
	})
	defer guard.Unpatch()

	key := "/knitter/manager/pods/admin/9854396d-8e7a-43e5-9666-f95bdaa04793-1-11111"
	errStr := "delete leaf error"
	dbmock.EXPECT().DeleteLeaf(key).Return(errors.New(errStr))

	Convey("TestDeleteLeafFail ", t, func() {
		err := pod1.Delete()
		So(err.Error(), ShouldEqual, errStr)
	})
}

func TestGetSucc(t *testing.T) {
	pod1 := MakeLogicPod1()
	ctl := gomock.NewController(t)
	defer ctl.Finish()
	dbmock := mockdbaccessor.NewMockDbAccessor(ctl)

	guard := Patch(common.GetDataBase, func() dbaccessor.DbAccessor {
		return dbmock
	})
	defer guard.Unpatch()

	podBodyExpect, _ := json.Marshal(pod1)
	key := "/knitter/manager/pods/admin/9854396d-8e7a-43e5-9666-f95bdaa04793-1-11111"
	dbmock.EXPECT().ReadLeaf(key).Return(string(podBodyExpect), nil)

	Convey("TestGetSucc ", t, func() {
		pod := &LogicPod{
			TenantID: "admin",
			PodName:  "9854396d-8e7a-43e5-9666-f95bdaa04793-1-11111",
		}
		err := pod.Get()
		So(pod, ShouldResemble, pod1)
		So(err, ShouldBeNil)
	})
}

func TestGetTenantIdNilFail(t *testing.T) {
	pod1TenantIDIsNil := MakeLogicPod1TenantIDIsNil()
	Convey("TestGetTenantIdNilFail ", t, func() {
		err := pod1TenantIDIsNil.Get()
		So(err, ShouldEqual, errobj.ErrTenantsIDOrPodNameIsNil)
	})
}

func TestGetReadLeafFail(t *testing.T) {
	pod1 := MakeLogicPod1()
	ctl := gomock.NewController(t)
	defer ctl.Finish()
	dbmock := mockdbaccessor.NewMockDbAccessor(ctl)

	guard := Patch(common.GetDataBase, func() dbaccessor.DbAccessor {
		return dbmock
	})
	defer guard.Unpatch()

	key := "/knitter/manager/pods/admin/9854396d-8e7a-43e5-9666-f95bdaa04793-1-11111"
	errStr := "read leaf error"
	dbmock.EXPECT().ReadLeaf(key).Return("", errors.New(errStr))

	Convey("TestGetReadLeafFail ", t, func() {
		err := pod1.Get()
		So(err.Error(), ShouldEqual, errStr)
	})
}

func TestGetUnmarshalFail(t *testing.T) {
	pod1 := MakeLogicPod1()

	ctl := gomock.NewController(t)
	defer ctl.Finish()
	dbmock := mockdbaccessor.NewMockDbAccessor(ctl)

	guard := Patch(common.GetDataBase, func() dbaccessor.DbAccessor {
		return dbmock
	})
	defer guard.Unpatch()

	key := "/knitter/manager/pods/admin/9854396d-8e7a-43e5-9666-f95bdaa04793-1-11111"
	errMashalStr := string("{pod:,,}")
	dbmock.EXPECT().ReadLeaf(key).Return(errMashalStr, nil)

	Convey("TestGetUnmarshalFail ", t, func() {
		err := pod1.Get()
		So(err.Error(), ShouldContainSubstring, "invalid character")
	})
}

func TestGetEncapPodForResponseSucc(t *testing.T) {

	pod1 := MakeLogicPod1()

	podForResponse1 := MakePodForResponse1()

	encapPodExpect := &EncapPodForResponse{Pod: podForResponse1}

	var logicPod *LogicPod
	guard := PatchInstanceMethod(reflect.TypeOf(logicPod),
		"Get", func(_ *LogicPod) error {
			return nil
		})
	defer guard.Unpatch()

	Convey("TestGetEncapPodForResponseSucc", t, func() {
		encapPod, err := pod1.GetEncapPodForResponse()
		So(encapPod, ShouldResemble, encapPodExpect)
		So(err, ShouldBeNil)
	})
}

func TestGetEncapPodForResponseGetFail(t *testing.T) {
	pod1 := MakeLogicPod1()

	getErr := errors.New("get err")
	var logicPod *LogicPod
	guard := PatchInstanceMethod(reflect.TypeOf(logicPod),
		"Get", func(_ *LogicPod) error {
			return getErr
		})
	defer guard.Unpatch()

	Convey("TestGetEncapPodForResponseGetFail", t, func() {
		_, err := pod1.GetEncapPodForResponse()

		So(err.Error(), ShouldEqual, "get err")
	})
}

func TestGetEncapPodForResponseTenantIdNilFail(t *testing.T) {
	pod1TenantIDIsNil := MakeLogicPod1TenantIDIsNil()

	Convey("TestGetEncapPodForResponseTenantIdNilFail", t, func() {
		_, err := pod1TenantIDIsNil.GetEncapPodForResponse()

		So(err, ShouldResemble, errobj.ErrTenantsIDOrPodNameIsNil)
	})
}

func TestTransformToPodForResponseSUCC(t *testing.T) {

	pod1 := MakeLogicPod1()
	podForResponse1Expect := MakePodForResponse1()

	Convey("TestTransformToPodForResponseSUCC", t, func() {
		podForResponse1 := pod1.TransformToPodForResponse()

		So(podForResponse1, ShouldResemble, podForResponse1Expect)
	})

}

func TestGetEncapPodsForResponseSUCC(t *testing.T) {

	pod1 := MakeLogicPod1()
	pod2 := MakeLogicPod2()
	pods := []*LogicPod{pod1, pod2}

	podForResponse1 := MakePodForResponse1()
	podForResponse2 := MakePodForResponse2()
	podsForResponse := []*PodForResponse{podForResponse1, podForResponse2}
	encapPodsForResponseExpect := &EncapPodsForResponse{Pods: podsForResponse}
	encapPodsForResponseExpectBody, _ := json.Marshal(encapPodsForResponseExpect)

	var logicPodManager *LogicPodManager
	guard := PatchInstanceMethod(reflect.TypeOf(logicPodManager), "GetAll", func(self *LogicPodManager, _ string) ([]*LogicPod, error) {
		return pods, nil
	})
	defer guard.Unpatch()

	Convey("TestGetEncapPodsForResponse_SUCC", t, func() {

		podManager := &LogicPodManager{}
		encapPod, err := podManager.GetEncapPodsForResponse("admin")
		encapPodBody, _ := json.Marshal(encapPod)

		So(string(encapPodBody), ShouldEqual, string(encapPodsForResponseExpectBody))
		So(err, ShouldBeNil)
	})
}

func TestGetEncapPodsForResponseTenantIdNilFaiL(t *testing.T) {
	podManager := &LogicPodManager{}

	Convey("TestGetEncapPodsForResponseFAIL", t, func() {
		_, err := podManager.GetEncapPodsForResponse("")

		So(err, ShouldResemble, errobj.ErrTenantsIDIsNil)
	})
}

func TestGetEncapPodsForResponseGetAllFail(t *testing.T) {

	pod1 := MakeLogicPod1()
	pod2 := MakeLogicPod2()
	pods := []*LogicPod{pod1, pod2}

	var logicPodManager *LogicPodManager
	errGetAllExpect := errors.New("GetAll err")
	guard := PatchInstanceMethod(reflect.TypeOf(logicPodManager), "GetAll", func(self *LogicPodManager, _ string) ([]*LogicPod, error) {
		return pods, errGetAllExpect
	})
	defer guard.Unpatch()

	Convey("TestGetEncapPodsForResponseGetAllFail", t, func() {

		podManager := &LogicPodManager{}
		_, err := podManager.GetEncapPodsForResponse("admin")

		So(err, ShouldResemble, errGetAllExpect)
	})
}

func TestGetAllSucc(t *testing.T) {
	logicPod1 := MakeLogicPod1()
	logicPod2 := MakeLogicPod2()
	logicPodsExpect := []*LogicPod{logicPod1, logicPod2}

	pod1Body, _ := json.Marshal(logicPod1)
	pod2Body, _ := json.Marshal(logicPod2)
	keyPod1 := "/knitter/manager/pods/admin/9854396d-8e7a-43e5-9666-f95bdaa04793-1-11111"
	keyPod2 := "/knitter/manager/pods/admin/1d521358-f523-4f5b-970d-416aece0a2fc-1-22222"

	node1 := &client.Node{
		Key:   keyPod1,
		Value: string(pod1Body),
	}
	node2 := &client.Node{
		Key:   keyPod2,
		Value: string(pod2Body),
	}
	nodes := make([]*client.Node, 0)
	nodes = append(nodes, node1)
	nodes = append(nodes, node2)

	ctl := gomock.NewController(t)
	defer ctl.Finish()
	dbmock := mockdbaccessor.NewMockDbAccessor(ctl)
	guard := Patch(common.GetDataBase, func() dbaccessor.DbAccessor {
		return dbmock
	})
	defer guard.Unpatch()
	key := "/knitter/manager/pods/admin"
	dbmock.EXPECT().ReadDir(key).Return(nodes, nil)

	Convey("TestGetAllSucc", t, func() {
		logicPodManager := LogicPodManager{}
		logicPods, err := logicPodManager.GetAll("admin")
		So(logicPods, ShouldResemble, logicPodsExpect)
		So(err, ShouldBeNil)

	})

}

func TestGetAllTenantIdNilFaiL(t *testing.T) {
	podManager := &LogicPodManager{}

	Convey("TestGetAllTenantIdNilFaiL", t, func() {
		_, err := podManager.GetEncapPodsForResponse("")

		So(err, ShouldResemble, errobj.ErrTenantsIDIsNil)
	})
}

func TestGetAllReadDirFaiL(t *testing.T) {

	logicPod1 := MakeLogicPod1()
	logicPod2 := MakeLogicPod2()

	pod1Body, _ := json.Marshal(logicPod1)
	pod2Body, _ := json.Marshal(logicPod2)
	keyPod1 := "/knitter/manager/pods/admin/9854396d-8e7a-43e5-9666-f95bdaa04793-1-11111"
	keyPod2 := "/knitter/manager/pods/admin/1d521358-f523-4f5b-970d-416aece0a2fc-1-22222"

	node1 := &client.Node{
		Key:   keyPod1,
		Value: string(pod1Body),
	}
	node2 := &client.Node{
		Key:   keyPod2,
		Value: string(pod2Body),
	}
	nodes := make([]*client.Node, 0)
	nodes = append(nodes, node1)
	nodes = append(nodes, node2)

	ctl := gomock.NewController(t)
	defer ctl.Finish()
	dbmock := mockdbaccessor.NewMockDbAccessor(ctl)
	guard := Patch(common.GetDataBase, func() dbaccessor.DbAccessor {
		return dbmock
	})
	defer guard.Unpatch()
	key := "/knitter/manager/pods/admin"
	errReadDir := errors.New("read dir error")
	dbmock.EXPECT().ReadDir(key).Return(nodes, errReadDir)

	Convey("TestGetAllReadDirFaiL", t, func() {
		logicPodManager := LogicPodManager{}
		_, err := logicPodManager.GetAll("admin")
		So(err, ShouldResemble, errReadDir)

	})

}

func TestGetAllUnmarshalFail(t *testing.T) {

	logicPod1 := MakeLogicPod1()
	logicPod2 := MakeLogicPod2()

	pod1Body, _ := json.Marshal(logicPod1)
	pod2Body, _ := json.Marshal(logicPod2)
	keyPod1 := "/knitter/manager/pods/admin/9854396d-8e7a-43e5-9666-f95bdaa04793-1-11111"
	keyPod2 := "/knitter/manager/pods/admin/1d521358-f523-4f5b-970d-416aece0a2fc-1-22222"

	node1 := &client.Node{
		Key:   keyPod1,
		Value: string(pod1Body) + ",",
	}
	node2 := &client.Node{
		Key:   keyPod2,
		Value: string(pod2Body) + "??",
	}
	nodes := make([]*client.Node, 0)
	nodes = append(nodes, node1)
	nodes = append(nodes, node2)

	ctl := gomock.NewController(t)
	defer ctl.Finish()
	dbmock := mockdbaccessor.NewMockDbAccessor(ctl)
	guard := Patch(common.GetDataBase, func() dbaccessor.DbAccessor {
		return dbmock
	})
	defer guard.Unpatch()
	key := "/knitter/manager/pods/admin"
	dbmock.EXPECT().ReadDir(key).Return(nodes, nil)

	Convey("TestGetAllUnmarshalFail", t, func() {
		logicPodManager := LogicPodManager{}
		_, err := logicPodManager.GetAll("admin")
		So(err.Error(), ShouldContainSubstring, "invalid character")

	})

}

//func TestGetEncapPodsForResponse(t *testing.T) {
//
//	port1 := PortInfo{"port1", "control", "control", "10.92.247.1"}
//	port2 := PortInfo{"port2", "media", "media", "10.92.247.2"}
//	port3 := PortInfo{"port3", "media", "media", "10.92.247.3"}
//	port4 := PortInfo{"port4", "media", "media", "10.92.247.4"}
//	ports1 := []PortInfo{port1, port2}
//	ports2 := []PortInfo{port3, port4}
//
//	pod1 := LogicPod{"9854396d-8e7a-43e5-9666-f95bdaa04793-1-11111", "admin", "admin", ports1}
//	pod2 := LogicPod{"1d521358-f523-4f5b-970d-416aece0a2fc-1-22222", "admin", "admin", ports2}
//
//	pods := []*LogicPod{&pod1, &pod2}
//
//	portForResponse1 := PortForResponse{"control", "control", "10.92.247.1"}
//	portForResponse2 := PortForResponse{"media", "media", "10.92.247.2"}
//	portForResponse3 := PortForResponse{"media", "media", "10.92.247.3"}
//	portForResponse4 := PortForResponse{"media", "media", "10.92.247.4"}
//
//	portsForResponse1 := []*PortForResponse{&portForResponse1, &portForResponse2}
//	portsForResponse2 := []*PortForResponse{&portForResponse3, &portForResponse4}
//	podForResponse1 := &PodForResponse{
//		PodIps:     portsForResponse1,
//		Name:       "9854396d-8e7a-43e5-9666-f95bdaa04793-1-11111",
//		tenantUUID: "admin",
//	}
//
//	podForResponse2 := &PodForResponse{
//		PodIps:     portsForResponse2,
//		Name:       "1d521358-f523-4f5b-970d-416aece0a2fc-1-22222",
//		tenantUUID: "admin",
//	}
//	podsForREsponse := []*PodForResponse{podForResponse1, podForResponse2}
//	encapPodsForResponseExpect := &EncapPodsForResponse{Pods: podsForREsponse}
//	var logicPodManager *LogicPodManager
//	guard := PatchInstanceMethod(reflect.TypeOf(logicPodManager), "GetAll", func(self *LogicPodManager, _ string) ([]*LogicPod, error) {
//		return pods, nil
//	})
//	defer guard.Unpatch()
//
//	Convey("TestGetEncapPodsForResponse_SUCC", t, func() {
//
//		podManager := &LogicPodManager{}
//		encapPod, err := podManager.GetEncapPodsForResponse("admin")
//
//		encapPodBody, _ := json.Marshal(encapPod)
//		encapPodsForResponseExpect, _ := json.Marshal(encapPodsForResponseExpect)
//
//		So(string(encapPodBody), ShouldEqual, string(encapPodsForResponseExpect))
//		So(err, ShouldBeNil)
//	})
//}
