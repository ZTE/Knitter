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

package controllers

import (
	"encoding/json"
	"testing"

	"github.com/ZTE/Knitter/knitter-manager/err-obj"
	"github.com/ZTE/Knitter/knitter-manager/models"
	"github.com/ZTE/Knitter/pkg/adapter"
	"github.com/ZTE/Knitter/pkg/inter-cmpt/agt-mgr"

	. "github.com/golang/gostub"
	"github.com/smartystreets/goconvey/convey"
)

func TestIsBulk(t *testing.T) {
	convey.Convey("TestIsBulk for true", t, func() {
		agtBulkPortsReq := agtmgr.AgtBulkPortsReq{Ports: make([]agtmgr.AgtPortReq, 0)}
		body, _ := json.Marshal(agtBulkPortsReq)
		convey.So(isBulk(body), convey.ShouldEqual, true)
	})
	convey.Convey("TestIsBulk for false", t, func() {
		stubs := StubFunc(&adapter.NewObjectFromBytes, nil, errobj.ErrAny)
		defer stubs.Reset()
		agtBulkPortsReq := agtmgr.AgtBulkPortsReq{Ports: make([]agtmgr.AgtPortReq, 0)}
		body, _ := json.Marshal(agtBulkPortsReq)
		convey.So(isBulk(body), convey.ShouldEqual, false)

		body, _ = json.Marshal("1")
		convey.So(isBulk(body), convey.ShouldEqual, false)
	})
}

func TestBuildReqObj(t *testing.T) {
	agtBulkPortsReq := agtmgr.AgtBulkPortsReq{
		Ports: []agtmgr.AgtPortReq{
			{
				TenantID:    "tenant_id",
				NetworkName: "networkName1",
				PortName:    "portName1",
				NodeID:      "node_id",
				PodNs:       "pod_ns",
				PodName:     "pod_name1",
				FixIP:       "",
				ClusterID:   "cluster_id"},
			{
				TenantID:    "tenant_id",
				NetworkName: "networkName2",
				PortName:    "portName2",
				NodeID:      "node_id",
				PodNs:       "pod_ns",
				PodName:     "pod_name2",
				FixIP:       "",
				ClusterID:   "cluster_id"},
		}}
	body, _ := json.Marshal(agtBulkPortsReq)
	tranID := models.TranID("asdf")
	tenantID := "tenantUser"

	convey.Convey("TestBuildReqObj for succ\n", t, func() {
		reqBody, err := buildReqObj(body, tranID, tenantID)
		convey.So(err, convey.ShouldEqual, nil)
		convey.So(reqBody.Ports[0].NetworkName, convey.ShouldEqual, "networkName1")
		convey.So(reqBody.Ports[0].PortName, convey.ShouldEqual, "portName1")
		convey.So(reqBody.Ports[0].TenantId, convey.ShouldEqual, "tenantUser")
		convey.So(reqBody.Ports[1].NetworkName, convey.ShouldEqual, "networkName2")
		convey.So(reqBody.Ports[1].PortName, convey.ShouldEqual, "portName2")
		convey.So(reqBody.Ports[1].TenantId, convey.ShouldEqual, "tenantUser")
	})
	convey.Convey("TestBuildReqObj for err403\n", t, func() {
		stubs := StubFunc(&adapter.Unmarshal, errobj.Err403)
		defer stubs.Reset()
		_, err := buildReqObj(body, models.TranID("1"), tenantID)
		convey.So(err, convey.ShouldEqual, errobj.Err403)
	})
}
