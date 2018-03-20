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
	"github.com/ZTE/Knitter/knitter-manager/iaas"
	"github.com/ZTE/Knitter/knitter-manager/public"
	"github.com/ZTE/Knitter/knitter-manager/tests/mock/db-mock"
	"github.com/ZTE/Knitter/pkg/db-accessor"
	"github.com/ZTE/Knitter/pkg/etcd"
	"github.com/antonholmquist/jason"
	"github.com/bouk/monkey"
	"github.com/golang/mock/gomock"
	"github.com/smartystreets/goconvey/convey"
	"reflect"
	"testing"
)

func Test_InitEnv4Manger_nil(t *testing.T) {
	conf := `{
			"etcd": {
				"api_version": 2,
				"urls": "etcd_url",
				"etcd_service_query_url": "http://192.169.0.14:10081/api/microservices/v1/services/etcd/version/v2"
			},
			"self_service": {
				"url": "knitter_manager_url"
			},
			"interval": {
				"senconds": "15"
			},
			"multiple_iaas_tenants": false,
			"max_req_attach": "5",
			"event_url": {
				"platform": "platform_event_url",
				"wiki": "http://wiki.zte.com.cn/pages/viewpage.action?pageId=17786624"
			}
		}`
	confObj, _ := jason.NewObjectFromBytes([]byte(conf))

	mockCtl := gomock.NewController(t)
	defer mockCtl.Finish()
	mockDb := mockdbaccessor.NewMockDbAccessor(mockCtl)
	monkey.Patch(etcd.NewEtcd,
		func(ver int, url string) dbaccessor.DbAccessor {
			return mockDb
		})
	monkey.Patch(common.SetDataBase,
		func(i dbaccessor.DbAccessor) error {
			return nil
		})
	monkey.Patch(common.CheckDB,
		func() error {
			return nil
		})
	monkey.Patch(iaas.SetMultipleIaasTenantsFlag,
		func(_ bool) {
			return
		})
	monkey.Patch(common.RegisterSelfToDb,
		func(serviceURL string) error {
			return nil
		})
	monkey.Patch(initIaas,
		func(cfg *jason.Object) error {
			return nil
		})
	monkey.Patch(GetSyncMgt,
		func() *SyncMgt {
			return &SyncMgt{Active: false}
		})
	syncMgt := &SyncMgt{}
	monkey.PatchInstanceMethod(reflect.TypeOf(syncMgt), "SetInterval",
		func(_ *SyncMgt, interval string) error {
			return nil
		})
	monkey.Patch(SetNetQuota,
		func(cfg *jason.Object) {
			return
		})
	monkey.Patch(UpdateEtcd4NetQuota,
		func() error {
			return nil
		})
	monkey.Patch(LoadAllResourcesToCache,
		func() {
			return
		})
	monkey.Patch(CancelResidualTenants,
		func() {
			return
		})
	convey.Convey("Test InitEnv4Manger nil:", t, func() {
		err := InitEnv4Manger(confObj)
		convey.So(err, convey.ShouldBeNil)
	})

}
