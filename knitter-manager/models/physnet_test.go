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
	"errors"
	"github.com/ZTE/Knitter/knitter-manager/iaas"
	"github.com/ZTE/Knitter/pkg/noauth_openstack"
	"github.com/ZTE/Knitter/pkg/openstack"
	"github.com/golang/gostub"
	"github.com/golang/mock/gomock"
	"github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestUpdateDefaultPhysnet(t *testing.T) {
	cfgMock := gomock.NewController(t)
	defer cfgMock.Finish()
	var phy string = "physnet2"
	noAuth := noauth_openstack.NoauthOpenStack{
		NoauthOpenStackConf: noauth_openstack.NoauthOpenStackConf{
			IP:       "10.92.247.10",
			Tenantid: "admin",
			URL:      "hettp://10.92.247:10080/vnm/v2.0",
		},
		NeutronConf: noauth_openstack.NoauthNeutronConf{
			Port:   "9696",
			ApiVer: "v2.0",
			ProviderConf: noauth_openstack.DefaultProviderConf{
				PhyscialNetwork: "physnet1",
				NetworkType:     "vlan",
			},
		},
	}
	stubs := gostub.StubFunc(&iaas.GetIaaS, &noAuth)
	defer stubs.Reset()
	stubs.StubFunc(&CheckPhysnet, nil)
	stubs.StubFunc(&iaas.SaveDefaultPhysnet, nil)

	convey.Convey("Test_UpdateDefaultPhysnet_OK\n", t, func() {
		errU := UpdateDefaultPhysnet(phy)
		convey.So(errU, convey.ShouldEqual, nil)
	})
}

func TestUpdateDefaultPhysnetErr(t *testing.T) {
	cfgMock := gomock.NewController(t)
	defer cfgMock.Finish()
	var phy string = "physnet2"

	convey.Convey("Test_UpdateDefaultPhysnet_Err_CheckPhysnet\n", t, func() {
		noAuth := noauth_openstack.NoauthOpenStack{
			NoauthOpenStackConf: noauth_openstack.NoauthOpenStackConf{
				IP:       "10.92.247.10",
				Tenantid: "admin",
				URL:      "hettp://10.92.247:10080/vnm/v2.0",
			},
			NeutronConf: noauth_openstack.NoauthNeutronConf{
				Port:   "9696",
				ApiVer: "v2.0",
				ProviderConf: noauth_openstack.DefaultProviderConf{
					PhyscialNetwork: "physnet1",
					NetworkType:     "vlan",
				},
			},
		}
		stubs := gostub.StubFunc(&iaas.GetIaaS, &noAuth)
		defer stubs.Reset()
		stubs.StubFunc(&CheckPhysnet, errors.New("checkPhysnet err"))
		errU := UpdateDefaultPhysnet(phy)
		convey.So(errU, convey.ShouldNotEqual, nil)
	})

	convey.Convey("Test_UpdateDefaultPhysnet_Err_value\n", t, func() {
		opn := openstack.OpenStack{}
		stubs := gostub.StubFunc(&iaas.GetIaaS, &opn)
		defer stubs.Reset()
		errU := UpdateDefaultPhysnet(phy)
		convey.So(errU, convey.ShouldNotEqual, nil)
	})
}
