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
	_ "encoding/json"
	"errors"
	_ "net/http"
	"testing"
	"time"

	"github.com/ZTE/Knitter/knitter-manager/err-obj"
	"github.com/ZTE/Knitter/knitter-manager/iaas"
	"github.com/ZTE/Knitter/knitter-manager/public"
	"github.com/ZTE/Knitter/knitter-manager/tests"
	_ "github.com/ZTE/Knitter/knitter-manager/tests"
	"github.com/ZTE/Knitter/knitter-manager/tests/mock/db-mock"
	"github.com/ZTE/Knitter/pkg/db-accessor"
	_ "github.com/ZTE/Knitter/pkg/db-accessor"
	"github.com/ZTE/Knitter/pkg/iaas-accessor"
	_ "github.com/ZTE/Knitter/pkg/iaas-accessor"
	"github.com/bouk/monkey"
	_ "github.com/coreos/etcd/client"
	"github.com/golang/gostub"
	"github.com/golang/mock/gomock"
	. "github.com/smartystreets/goconvey/convey"
)

func TestGetCancellingTenantsInfo(t *testing.T) {
	Convey("TestGetCancellingTenantsInfo", t, func() {
		Convey("succ test", func() {
			ctl := gomock.NewController(t)
			defer ctl.Finish()
			dbmock := mockdbaccessor.NewMockDbAccessor(ctl)

			stubs := gostub.StubFunc(&common.GetDataBase, dbmock)

			Paas1TID := "paas1"
			Paas2TID := "paas2"
			Paas3TID := "paas3"

			tids := []string{"admin", Paas1TID, Paas2TID, Paas3TID}
			stubs.StubFunc(&getAllTenantIds, tids, nil)
			defer stubs.Reset()

			Paas1TenantInStr := `{
				"tenant_name": "paas1",
				"tenant_uuid": "paas1",
				"is_cancelling": true}`
			Paas2TenantInStr := `{
				"tenant_name": "paas2",
				"tenant_uuid": "paas2",
				"is_cancelling": true}`
			Paas3TenantInStr := `{
				"tenant_name": "paas3",
				"tenant_uuid": "paas3",
				"is_cancelling": false}`
			cancelTIDs := []string{Paas1TID, Paas2TID}

			gomock.InOrder(
				dbmock.EXPECT().ReadLeaf("/paasnet/tenants/paas1/self").Return(Paas1TenantInStr, nil),
				dbmock.EXPECT().ReadLeaf("/paasnet/tenants/paas2/self").Return(Paas2TenantInStr, nil),
				dbmock.EXPECT().ReadLeaf("/paasnet/tenants/paas3/self").Return(Paas3TenantInStr, nil),
			)

			tenantsInfo := GetCancellingTenantsInfo()
			So(tenantsInfo, ShouldResemble, cancelTIDs)
		})

		Convey("fail test", func() {
			Convey("getAllTenantIds error", func() {
				ctl := gomock.NewController(t)
				defer ctl.Finish()
				dbmock := mockdbaccessor.NewMockDbAccessor(ctl)

				stubs := gostub.StubFunc(&common.GetDataBase, dbmock)
				errStr := "500: etcd cluster is unavailable or misconfig"
				stubs.StubFunc(&getAllTenantIds, nil, errors.New(errStr))
				defer stubs.Reset()

				tenantsInfo := GetCancellingTenantsInfo()
				So(tenantsInfo, ShouldBeNil)
			})

			Convey("getTenantInfo error", func() {
				ctl := gomock.NewController(t)
				defer ctl.Finish()
				dbmock := mockdbaccessor.NewMockDbAccessor(ctl)

				stubs := gostub.StubFunc(&common.GetDataBase, dbmock)

				Paas1TID := "paas1"
				Paas2TID := "paas2"
				Paas3TID := "paas3"

				tids := []string{"admin", Paas1TID, Paas2TID, Paas3TID}
				stubs.StubFunc(&getAllTenantIds, tids, nil)
				defer stubs.Reset()

				Paas1TenantInStr := `{
				"tenant_name": "paas1",
				"tenant_uuid": "paas1",
				"is_cancelling": true}`

				Paas3TenantInStr := `{
				"tenant_name": "paas3",
				"tenant_uuid": "paas3",
				"is_cancelling": false}`
				cancelTIDs := []string{Paas1TID}

				errStr := "500: etcd cluster is unavailable or misconfig"
				gomock.InOrder(
					dbmock.EXPECT().ReadLeaf("/paasnet/tenants/paas1/self").Return(Paas1TenantInStr, nil),
					dbmock.EXPECT().ReadLeaf("/paasnet/tenants/paas2/self").Return("", errors.New(errStr)),
					dbmock.EXPECT().ReadLeaf("/paasnet/tenants/paas3/self").Return(Paas3TenantInStr, nil),
				)

				tenantsInfo := GetCancellingTenantsInfo()
				So(tenantsInfo, ShouldResemble, cancelTIDs)
			})

			Convey("getTenantInfo-json.Unmarshal error", func() {
				ctl := gomock.NewController(t)
				defer ctl.Finish()
				dbmock := mockdbaccessor.NewMockDbAccessor(ctl)

				stubs := gostub.StubFunc(&common.GetDataBase, dbmock)

				Paas1TID := "paas1"
				Paas2TID := "paas2"
				Paas3TID := "paas3"

				tids := []string{"admin", Paas1TID, Paas2TID, Paas3TID}
				stubs.StubFunc(&getAllTenantIds, tids, nil)
				defer stubs.Reset()

				Paas1TenantInStr := `{
				"tenant_name": "paas1",
				"tenant_uuid": "paas1",
				"is_cancelling": true}`

				Paas2TenantInStr := `{
				"tenant_name": "paas2",
				"tenant_uuid": "paas2",
				"is_cancelling": true,}`

				Paas3TenantInStr := `{
				"tenant_name": "paas3",
				"tenant_uuid": "paas3",
				"is_cancelling": false}`
				cancelTIDs := []string{Paas1TID}

				gomock.InOrder(
					dbmock.EXPECT().ReadLeaf("/paasnet/tenants/paas1/self").Return(Paas1TenantInStr, nil),
					dbmock.EXPECT().ReadLeaf("/paasnet/tenants/paas2/self").Return(Paas2TenantInStr, nil),
					dbmock.EXPECT().ReadLeaf("/paasnet/tenants/paas3/self").Return(Paas3TenantInStr, nil),
				)

				tenantsInfo := GetCancellingTenantsInfo()
				So(tenantsInfo, ShouldResemble, cancelTIDs)
			})
		})
	})
}

func TestGetAllNormalTenantIds(t *testing.T) {
	Convey("TestGetAllNormalTenantIds", t, func() {
		Convey("succ test", func() {
			ctl := gomock.NewController(t)
			defer ctl.Finish()
			dbmock := mockdbaccessor.NewMockDbAccessor(ctl)

			stubs := gostub.StubFunc(&common.GetDataBase, dbmock)

			Paas1TID := "paas1"
			Paas2TID := "paas2"
			Paas3TID := "paas3"

			tids := []string{"admin", Paas1TID, Paas2TID, Paas3TID}
			stubs.StubFunc(&getAllTenantIds, tids, nil)
			defer stubs.Reset()

			Paas1TenantInStr := `{
				"tenant_name": "paas1",
				"tenant_uuid": "paas1",
				"is_cancelling": true}`
			Paas2TenantInStr := `{
				"tenant_name": "paas2",
				"tenant_uuid": "paas2",
				"is_cancelling": true}`
			Paas3TenantInStr := `{
				"tenant_name": "paas3",
				"tenant_uuid": "paas3",
				"is_cancelling": false}`
			normalTIDs := []string{"admin", Paas3TID}

			gomock.InOrder(
				dbmock.EXPECT().ReadLeaf("/paasnet/tenants/paas1/self").Return(Paas1TenantInStr, nil),
				dbmock.EXPECT().ReadLeaf("/paasnet/tenants/paas2/self").Return(Paas2TenantInStr, nil),
				dbmock.EXPECT().ReadLeaf("/paasnet/tenants/paas3/self").Return(Paas3TenantInStr, nil),
			)

			tenantsInfo, _ := GetAllNormalTenantIds()
			So(tenantsInfo, ShouldResemble, normalTIDs)
		})

		Convey("fail test", func() {
			Convey("getAllTenantIds error", func() {
				ctl := gomock.NewController(t)
				defer ctl.Finish()
				dbmock := mockdbaccessor.NewMockDbAccessor(ctl)

				stubs := gostub.StubFunc(&common.GetDataBase, dbmock)
				errStr := "500: etcd cluster is unavailable or misconfig"
				stubs.StubFunc(&getAllTenantIds, nil, errors.New(errStr))
				defer stubs.Reset()

				tenantsInfo, _ := GetAllNormalTenantIds()
				So(tenantsInfo, ShouldBeNil)
			})

			Convey("getTenantInfo error", func() {
				ctl := gomock.NewController(t)
				defer ctl.Finish()
				dbmock := mockdbaccessor.NewMockDbAccessor(ctl)

				stubs := gostub.StubFunc(&common.GetDataBase, dbmock)

				Paas1TID := "paas1"
				Paas2TID := "paas2"
				Paas3TID := "paas3"

				tids := []string{"admin", Paas1TID, Paas2TID, Paas3TID}
				stubs.StubFunc(&getAllTenantIds, tids, nil)
				defer stubs.Reset()

				Paas1TenantInStr := `{
				"tenant_name": "paas1",
				"tenant_uuid": "paas1",
				"is_cancelling": true}`

				errStr := "500: etcd cluster is unavailable or misconfig"
				gomock.InOrder(
					dbmock.EXPECT().ReadLeaf("/paasnet/tenants/paas1/self").Return(Paas1TenantInStr, nil),
					dbmock.EXPECT().ReadLeaf("/paasnet/tenants/paas2/self").Return("", errors.New(errStr)),
				)

				_, err := GetAllNormalTenantIds()
				So(err, ShouldNotBeNil)
			})
		})
	})
}

func TestCancelTenant(t *testing.T) {
	Convey("TestCancelTenant", t, func() {
		Convey("succ test", func() {
			ctl := gomock.NewController(t)
			defer ctl.Finish()
			dbmock := mockdbaccessor.NewMockDbAccessor(ctl)

			tenantID := "tenant-paas-id"

			stubs := gostub.StubFunc(&ClearTenantPods, nil)
			stubs.StubFunc(&ClearIPGroups, nil)
			stubs.StubFunc(&ClearLogicalPorts, nil)
			stubs.StubFunc(&ClearPhysicalPorts, nil)
			stubs.StubFunc(&ClearNetworks, nil)
			stubs.StubFunc(&common.GetDataBase, dbmock)
			defer stubs.Reset()

			dbmock.EXPECT().DeleteDir("/paasnet/tenants/" + tenantID).Return(nil)
			err := CancelTenant(tenantID)
			So(err, ShouldBeNil)
		})

		Convey("failded test", func() {
			Convey("ClearTenantPods error", func() {
				ctl := gomock.NewController(t)
				defer ctl.Finish()

				tenantID := "tenant-paas-id"
				errStr := "etcd misconfig"
				stubs := gostub.StubFunc(&ClearTenantPods, errors.New(errStr))
				defer stubs.Reset()

				err := CancelTenant(tenantID)
				So(err.Error(), ShouldEqual, errStr)
			})

			Convey("ClearLogicalPorts error", func() {
				ctl := gomock.NewController(t)
				defer ctl.Finish()

				tenantID := "tenant-paas-id"
				errStr := "openstack access error"
				stubs := gostub.StubFunc(&ClearTenantPods, nil)
				stubs.StubFunc(&ClearIPGroups, nil)
				stubs.StubFunc(&ClearLogicalPorts, errors.New(errStr))
				defer stubs.Reset()

				err := CancelTenant(tenantID)
				So(err.Error(), ShouldEqual, errStr)
			})

			Convey("DeletePublicNetworks error", func() {
				ctl := gomock.NewController(t)
				defer ctl.Finish()

				tenantID := "tenant-paas-id"
				errStr := "401: etcd access error"
				stubs := gostub.StubFunc(&ClearTenantPods, nil)
				stubs.StubFunc(&ClearIPGroups, nil)
				stubs.StubFunc(&ClearLogicalPorts, nil)
				stubs.StubFunc(&ClearPhysicalPorts, errors.New(errStr))
				defer stubs.Reset()

				err := CancelTenant(tenantID)
				So(err.Error(), ShouldEqual, errStr)
			})

			Convey("ClearPrivateNetworks error", func() {
				ctl := gomock.NewController(t)
				defer ctl.Finish()

				tenantID := "tenant-paas-id"
				errStr := "502: openstack access error"
				stubs := gostub.StubFunc(&ClearTenantPods, nil)
				stubs.StubFunc(&ClearIPGroups, nil)
				stubs.StubFunc(&ClearLogicalPorts, nil)
				stubs.StubFunc(&ClearPhysicalPorts, nil)
				stubs.StubFunc(&ClearNetworks, errors.New(errStr))
				defer stubs.Reset()

				err := CancelTenant(tenantID)
				So(err.Error(), ShouldEqual, errStr)
			})

			Convey("DeleteDir for tenant error", func() {
				ctl := gomock.NewController(t)
				defer ctl.Finish()
				dbmock := mockdbaccessor.NewMockDbAccessor(ctl)

				tenantID := "tenant-paas-id"
				errStr := "502: openstack access error"
				stubs := gostub.StubFunc(&ClearTenantPods, nil)
				stubs.StubFunc(&ClearIPGroups, nil)
				stubs.StubFunc(&ClearLogicalPorts, nil)
				stubs.StubFunc(&ClearPhysicalPorts, nil)
				stubs.StubFunc(&ClearNetworks, nil)
				stubs.StubFunc(&common.GetDataBase, dbmock)
				defer stubs.Reset()

				dbmock.EXPECT().DeleteDir("/paasnet/tenants/" + tenantID).Return(errors.New(errStr))
				err := CancelTenant(tenantID)
				So(err.Error(), ShouldEqual, errStr)
			})
		})
	})
}

func TestClearTenantPods(t *testing.T) {
	Convey("TestClearTenantPods", t, func() {
		Convey("succ test", func() {
			ctl := gomock.NewController(t)
			defer ctl.Finish()
			dbmock := mockdbaccessor.NewMockDbAccessor(ctl)

			stubs := gostub.StubFunc(&waitTenantPodsDeleted, nil)
			stubs.StubFunc(&common.GetDataBase, dbmock)
			defer stubs.Reset()

			tenantID := "tenant-paas-id"
			err := ClearTenantPods(tenantID)
			So(err, ShouldBeNil)
		})
	})
}

func TestGetAllNormalTenants(t *testing.T) {
	Convey("TestGetAllNormalTenants", t, func() {
		Convey("succ test", func() {
			ctl := gomock.NewController(t)
			defer ctl.Finish()
			dbmock := mockdbaccessor.NewMockDbAccessor(ctl)

			stubs := gostub.StubFunc(&common.GetDataBase, dbmock)

			Paas1TID := "paas1"
			Paas2TID := "paas2"
			Paas3TID := "paas3"

			tids := []string{Paas1TID, Paas2TID, Paas3TID}
			stubs.StubFunc(&getAllTenantIds, tids, nil)
			//stubs.StubFunc(&GetNetNumOfTenant, 4)
			defer stubs.Reset()

			Paas1TenantInStr := `{
				"tenant_name": "paas1",
				"tenant_uuid": "paas1",
				"net_number": 1,
				"net_quota": 10,
				"is_cancelling": false}`
			Paas2TenantInStr := `{
				"tenant_name": "paas2",
				"tenant_uuid": "paas2",
				"net_number": 1,
				"net_quota": 10,
				"is_cancelling": false}`
			Paas3TenantInStr := `{
				"tenant_name": "paas3",
				"tenant_uuid": "paas3",
				"net_number": 1,
				"net_quota": 10,
				"is_cancelling": false}`
			paasTenants := []*Tenant{
				{TenantName: "paas1", TenantUUID: "paas1", NetNum: 1, Quota: 10, IsCancelling: false},
				{TenantName: "paas2", TenantUUID: "paas2", NetNum: 1, Quota: 10, IsCancelling: false},
				{TenantName: "paas3", TenantUUID: "paas3", NetNum: 1, Quota: 10, IsCancelling: false},
			}

			gomock.InOrder(
				dbmock.EXPECT().ReadLeaf("/paasnet/tenants/paas1/self").Return(Paas1TenantInStr, nil),
				dbmock.EXPECT().ReadLeaf("/paasnet/tenants/paas2/self").Return(Paas2TenantInStr, nil),
				dbmock.EXPECT().ReadLeaf("/paasnet/tenants/paas3/self").Return(Paas3TenantInStr, nil),
			)

			tenantsInfo := GetAllTenants()
			So(tenantsInfo, ShouldResemble, paasTenants)
		})

		Convey("fail test", func() {
			Convey("getAllTenantIds error", func() {
				ctl := gomock.NewController(t)
				defer ctl.Finish()
				dbmock := mockdbaccessor.NewMockDbAccessor(ctl)

				stubs := gostub.StubFunc(&common.GetDataBase, dbmock)
				errStr := "500: etcd cluster is unavailable or misconfig"
				stubs.StubFunc(&getAllTenantIds, nil, errors.New(errStr))
				defer stubs.Reset()

				tenantsInfo := GetAllTenants()
				So(tenantsInfo, ShouldBeNil)
			})

			Convey("getTenantInfo error", func() {
				ctl := gomock.NewController(t)
				defer ctl.Finish()
				dbmock := mockdbaccessor.NewMockDbAccessor(ctl)

				stubs := gostub.StubFunc(&common.GetDataBase, dbmock)

				Paas1TID := "paas1"
				Paas2TID := "paas2"
				Paas3TID := "paas3"

				tids := []string{Paas1TID, Paas2TID, Paas3TID}
				stubs.StubFunc(&getAllTenantIds, tids, nil)
				//stubs.StubFunc(&GetNetNumOfTenant, 4)
				defer stubs.Reset()

				Paas1TenantInStr := `{
				"tenant_name": "paas1",
				"tenant_uuid": "paas1",
				"net_number": 1,
				"net_quota": 10,
				"is_cancelling": false}`
				Paas3TenantInStr := `{
				"tenant_name": "paas3",
				"tenant_uuid": "paas3",
				"net_number": 1,
				"net_quota": 10,
				"is_cancelling": false}`
				paasTenants := []*Tenant{
					{TenantName: "paas1", TenantUUID: "paas1", NetNum: 1, Quota: 10, IsCancelling: false},
					{TenantName: "paas3", TenantUUID: "paas3", NetNum: 1, Quota: 10, IsCancelling: false},
				}

				errStr := "500: etcd cluster is unavailable or misconfig"
				gomock.InOrder(
					dbmock.EXPECT().ReadLeaf("/paasnet/tenants/paas1/self").Return(Paas1TenantInStr, nil),
					dbmock.EXPECT().ReadLeaf("/paasnet/tenants/paas2/self").Return("", errors.New(errStr)),
					dbmock.EXPECT().ReadLeaf("/paasnet/tenants/paas3/self").Return(Paas3TenantInStr, nil),
				)

				tenantsInfo := GetAllTenants()
				So(tenantsInfo, ShouldResemble, paasTenants)
			})
		})
	})
}

func Test_getWaitTenantPodsDeletedTimout(t *testing.T) {
	timeoutStr := "200"
	timeout := 200
	ctl := gomock.NewController(t)
	defer ctl.Finish()
	dbmock := mockdbaccessor.NewMockDbAccessor(ctl)
	dbmock.EXPECT().ReadLeaf(dbaccessor.GetKeyOfCancelWaitPodsDeletedTimeout()).Return(timeoutStr, nil)

	stub := gostub.StubFunc(&common.GetDataBase, dbmock)
	defer stub.Reset()

	Convey("Test_getWaitTenantPodsDeletedTimout", t, func() {
		retval := getWaitTenantPodsDeletedTimout()
		So(retval, ShouldEqual, timeout)
	})
}

func Test_getWaitTenantPodsDeletedTimout_EtcdErr(t *testing.T) {
	ctl := gomock.NewController(t)
	defer ctl.Finish()
	dbmock := mockdbaccessor.NewMockDbAccessor(ctl)
	dbmock.EXPECT().ReadLeaf(dbaccessor.GetKeyOfCancelWaitPodsDeletedTimeout()).Return("", errors.New("etcd misconfig"))

	stub := gostub.StubFunc(&common.GetDataBase, dbmock)
	defer stub.Reset()

	Convey("Test_getWaitTenantPodsDeletedTimout", t, func() {
		retval := getWaitTenantPodsDeletedTimout()
		So(retval, ShouldEqual, CancelWaitPodsDeletedTimeout)
	})
}

func Test_getWaitTenantPodsDeletedTimout_StrconvErr(t *testing.T) {
	timeoutStr := "1ab0"
	ctl := gomock.NewController(t)
	defer ctl.Finish()
	dbmock := mockdbaccessor.NewMockDbAccessor(ctl)
	dbmock.EXPECT().ReadLeaf(dbaccessor.GetKeyOfCancelWaitPodsDeletedTimeout()).Return(timeoutStr, nil)

	stub := gostub.StubFunc(&common.GetDataBase, dbmock)
	defer stub.Reset()

	Convey("Test_getWaitTenantPodsDeletedTimout", t, func() {
		retval := getWaitTenantPodsDeletedTimout()
		So(retval, ShouldEqual, CancelWaitPodsDeletedTimeout)
	})
}

func Test_waitTenantPodsDeleted(t *testing.T) {
	stubs := gostub.StubFunc(&getWaitTenantPodsDeletedTimout, 30)
	stubs.StubFunc(&GetPodsOfTenant, make([]*Pod, 0))
	defer stubs.Reset()

	Convey("Test_waitTenantPodsDeleted", t, func() {
		pods := waitTenantPodsDeleted("test-paas-tenant-id")
		So(pods, ShouldBeNil)
	})
}

func Test_waitTenantPodsDeleted_RetryOK(t *testing.T) {
	guard := monkey.Patch(time.Sleep, func(d time.Duration) {})
	defer guard.Unpatch()
	stubs := gostub.StubFunc(&getWaitTenantPodsDeletedTimout, 30)
	outputs := []gostub.Output{
		{StubVals: gostub.Values{make([]*Pod, 1)}},
		{StubVals: gostub.Values{nil}},
	}
	stubs.StubFuncSeq(&GetPodsOfTenant, outputs)
	defer stubs.Reset()

	Convey("Test_waitTenantPodsDeleted_RetryOK", t, func() {
		waitTenantPodsDeleted("test-paas-tenant-id")
	})
}

func Test_waitTenantPodsDeleted_RetryTimeout(t *testing.T) {
	guard := monkey.Patch(time.Sleep, func(d time.Duration) {})
	defer guard.Unpatch()
	stubs := gostub.StubFunc(&getWaitTenantPodsDeletedTimout, 20)
	outputs := []gostub.Output{
		{StubVals: gostub.Values{make([]*Pod, 1)}, Times: 2},
		{StubVals: gostub.Values{nil}},
	}
	stubs.StubFuncSeq(&GetPodsOfTenant, outputs)
	defer stubs.Reset()

	Convey("Test_waitTenantPodsDeleted_RetryTimeout", t, func() {
		waitTenantPodsDeleted("test-paas-tenant-id")
	})
}

func TestCancelTenantLoop(t *testing.T) {
	guard := monkey.Patch(time.Sleep, func(d time.Duration) {})
	defer guard.Unpatch()
	outputs := []gostub.Output{
		{StubVals: gostub.Values{errors.New("etcd misconfig")}},
		{StubVals: gostub.Values{nil}}}
	stub := gostub.StubFuncSeq(&CancelTenant, outputs)
	defer stub.Reset()

	Convey("TestCancelTenantLoop", t, func() {
		CancelTenantLoop("tenant-id")
	})
}

func TestCancelResidualTenants(t *testing.T) {
	stubs := gostub.StubFunc(&GetCancellingTenantsInfo, []string{"t1", "t2"})
	stubs.StubFunc(&CancelTenantLoop)
	defer stubs.Reset()

	Convey("TestCancelResidualTenants", t, func() {
		CancelResidualTenants()
	})
}

func Test_forceDeleteNetwork(t *testing.T) {
	mockCtl := gomock.NewController(t)
	defer mockCtl.Finish()
	iaasMock := test.NewMockIaaS(mockCtl)

	stubs := gostub.StubFunc(&deleteResidualPorts, nil)
	stubs.StubFunc(&iaas.GetIaaS, iaasMock)
	defer stubs.Reset()

	tenantID := "test-tenant-id"
	networkID := "test-network-id"
	iaasMock.EXPECT().DeleteNetwork(networkID).Return(nil)

	Convey("Test_forceDeleteNetwork", t, func() {
		err := forceDeleteNetwork(tenantID, networkID)
		So(err, ShouldBeNil)
	})
}

func Test_forceDeleteNetwork_DeletePortFailed(t *testing.T) {
	errStr := "etcd misconfig"
	stubs := gostub.StubFunc(&deleteResidualPorts, errors.New(errStr))
	defer stubs.Reset()

	tenantID := "test-tenant-id"
	networkID := "test-network-id"
	Convey("Test_forceDeleteNetwork_DeletePortFailed", t, func() {
		err := forceDeleteNetwork(tenantID, networkID)
		So(err.Error(), ShouldEqual, errStr)
	})
}

func Test_forceDeleteNetwork_DeleteNetworkFailed(t *testing.T) {
	mockCtl := gomock.NewController(t)
	defer mockCtl.Finish()
	iaasMock := test.NewMockIaaS(mockCtl)

	stubs := gostub.StubFunc(&deleteResidualPorts, nil)
	stubs.StubFunc(&iaas.GetIaaS, iaasMock)
	defer stubs.Reset()

	tenantID := "test-tenant-id"
	networkID := "test-network-id"
	errStr := "500 internal error"
	iaasMock.EXPECT().DeleteNetwork(networkID).Return(errors.New(errStr))

	Convey("Test_forceDeleteNetwork_DeleteNetworkFailed", t, func() {
		err := forceDeleteNetwork(tenantID, networkID)
		So(err.Error(), ShouldEqual, errStr)
	})
}

func Test_deleteResidualPorts(t *testing.T) {
	ctl := gomock.NewController(t)
	defer ctl.Finish()
	iaasMock := test.NewMockIaaS(ctl)
	stubs := gostub.StubFunc(&iaas.GetIaaS, iaasMock)
	stubs.StubFunc(&clearResidualPort, nil)
	defer stubs.Reset()

	tenantID := "test-tenant-id"
	networkID := "net_control_id"
	ports := make([]*iaasaccessor.Interface, 2)

	iaasMock.EXPECT().ListPorts(networkID).Return(ports, nil)

	Convey("Test_deleteResidualPorts", t, func() {
		err := deleteResidualPorts(tenantID, networkID)
		So(err, ShouldBeNil)
	})
}

func Test_deleteResidualPorts_ListPortsFailed(t *testing.T) {
	ctl := gomock.NewController(t)
	defer ctl.Finish()
	iaasMock := test.NewMockIaaS(ctl)
	stubs := gostub.StubFunc(&iaas.GetIaaS, iaasMock)
	defer stubs.Reset()

	tenantID := "test-tenant-id"
	networkID := "net_control_id"

	errStr := "iaas error"
	iaasMock.EXPECT().ListPorts(networkID).Return(nil, errors.New(errStr))

	Convey("Test_deleteResidualPorts_ListPortsFailed", t, func() {
		err := deleteResidualPorts(tenantID, networkID)
		So(err.Error(), ShouldEqual, errStr)
	})
}

func Test_deleteResidualPorts_DeletePortFailed(t *testing.T) {
	ctl := gomock.NewController(t)
	defer ctl.Finish()
	iaasMock := test.NewMockIaaS(ctl)
	stubs := gostub.StubFunc(&iaas.GetIaaS, iaasMock)
	errStr := "delete port error"
	outputs := []gostub.Output{
		{StubVals: gostub.Values{nil}},
		{StubVals: gostub.Values{errors.New(errStr)}},
	}
	stubs.StubFuncSeq(&clearResidualPort, outputs)
	defer stubs.Reset()

	tenantID := "test-tenant-id"
	networkID := "net_control_id"
	ports := []*iaasaccessor.Interface{
		{Id: "port-id-1"},
		{Id: "port-id-2"},
	}

	iaasMock.EXPECT().ListPorts(networkID).Return(ports, nil)

	Convey("Test_deleteResidualPorts_DeletePortFailed", t, func() {
		err := deleteResidualPorts(tenantID, networkID)
		So(err, ShouldEqual, errobj.ErrNetworkHasPortsInUse)
	})
}

func Test_clearResidualPort_DetachedPort(t *testing.T) {
	ctl := gomock.NewController(t)
	defer ctl.Finish()
	iaasMock := test.NewMockIaaS(ctl)
	stubs := gostub.StubFunc(&iaas.GetIaaS, iaasMock)
	defer stubs.Reset()

	tenantID := "test-tenant-id"
	port := iaasaccessor.Interface{DeviceId: "123-456-789-0", Id: "port-id"}

	iaasMock.EXPECT().DetachPortFromVM(port.DeviceId, port.Id).Return(nil)
	iaasMock.EXPECT().DeletePort(port.Id).Return(nil)
	Convey("Test_clearResidualPort_DetachedPort", t, func() {
		err := clearResidualPort(tenantID, &port)
		So(err, ShouldBeNil)
	})
}

func Test_clearResidualPort_DetachedPortFailed(t *testing.T) {
	ctl := gomock.NewController(t)
	defer ctl.Finish()
	iaasMock := test.NewMockIaaS(ctl)
	stubs := gostub.StubFunc(&iaas.GetIaaS, iaasMock)
	defer stubs.Reset()

	tenantID := "test-tenant-id"
	port := iaasaccessor.Interface{DeviceId: "123-456-789-0", Id: "port-id"}

	errStr := "iaas operation failed"
	iaasMock.EXPECT().DetachPortFromVM(port.DeviceId, port.Id).Return(errors.New(errStr))
	iaasMock.EXPECT().DeletePort(port.Id).Return(nil)
	Convey("Test_clearResidualPort_DetachedPortFailed", t, func() {
		err := clearResidualPort(tenantID, &port)
		So(err, ShouldBeNil)
	})
}

func Test_clearResidualPort_DeletePortFailed(t *testing.T) {
	ctl := gomock.NewController(t)
	defer ctl.Finish()
	iaasMock := test.NewMockIaaS(ctl)
	stubs := gostub.StubFunc(&iaas.GetIaaS, iaasMock)
	defer stubs.Reset()

	tenantID := "test-tenant-id"
	port := iaasaccessor.Interface{DeviceId: "123-456-789-0", Id: "port-id"}

	errStr := "iaas operation failed"
	iaasMock.EXPECT().DetachPortFromVM(port.DeviceId, port.Id).Return(nil)
	iaasMock.EXPECT().DeletePort(port.Id).Return(errors.New(errStr))
	Convey("Test_clearResidualPort_DeletePortFailed", t, func() {
		err := clearResidualPort(tenantID, &port)
		So(err.Error(), ShouldEqual, errStr)
	})
}

func Test_clearResidualPort_DHCPPort(t *testing.T) {
	tenantID := "test-tenant-id"
	port := iaasaccessor.Interface{DeviceId: "dhcpq-123", Id: "port-id"}
	Convey("Test_clearResidualPort_DetachedPort", t, func() {
		err := clearResidualPort(tenantID, &port)
		So(err, ShouldBeNil)
	})
}

func Test_clearResidualPort_NotDetachedPort(t *testing.T) {
	ctl := gomock.NewController(t)
	defer ctl.Finish()
	iaasMock := test.NewMockIaaS(ctl)
	stubs := gostub.StubFunc(&iaas.GetIaaS, iaasMock)
	defer stubs.Reset()

	tenantID := "test-tenant-id"
	port := iaasaccessor.Interface{DeviceId: "", Id: "port-id"}

	iaasMock.EXPECT().DeletePort(port.Id).Return(nil)
	Convey("Test_clearResidualPort_DetachedPort", t, func() {
		err := clearResidualPort(tenantID, &port)
		So(err, ShouldBeNil)
	})
}
