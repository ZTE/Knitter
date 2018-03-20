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
	"github.com/ZTE/Knitter/knitter-manager/iaas"
	"github.com/ZTE/Knitter/knitter-manager/public"
	"github.com/ZTE/Knitter/knitter-manager/tests"
	"github.com/ZTE/Knitter/knitter-manager/tests/mock/db-mock"
	"github.com/ZTE/Knitter/pkg/iaas-accessor"
	"github.com/antonholmquist/jason"
	"github.com/bouk/monkey"
	"github.com/coreos/etcd/client"
	"github.com/golang/gostub"
	"github.com/golang/mock/gomock"
	"github.com/rackspace/gophercloud"
	"github.com/rackspace/gophercloud/openstack/networking/v2/subnets"
	"github.com/smartystreets/goconvey/convey"
	"reflect"
	"testing"
)

func TestSaveInitConfOK(t *testing.T) {
	bodyContent := `{"init_configuration": "xxx"}`
	initConf := []byte(bodyContent)
	mockCtl := gomock.NewController(t)
	defer mockCtl.Finish()
	mockDB := mockdbaccessor.NewMockDbAccessor(mockCtl)
	stubs := gostub.StubFunc(&common.GetDataBase, mockDB)
	defer stubs.Reset()
	mockDB.EXPECT().SaveLeaf(gomock.Any(), gomock.Any()).Return(nil)
	convey.Convey("Test_SaveInitConf_OK\n", t, func() {
		err := SaveInitConf(initConf)
		convey.So(err, convey.ShouldEqual, nil)
	})
}

func TestSaveInitConfErr(t *testing.T) {
	bodyContent := `{"init_configuration": "xxx"}`
	initConf := []byte(bodyContent)
	mockCtl := gomock.NewController(t)
	defer mockCtl.Finish()
	mockDB := mockdbaccessor.NewMockDbAccessor(mockCtl)
	stubs := gostub.StubFunc(&common.GetDataBase, mockDB)
	defer stubs.Reset()
	mockDB.EXPECT().SaveLeaf(gomock.Any(), gomock.Any()).Return(errors.New("saveErr"))
	convey.Convey("Test_SaveInitConf_Err\n", t, func() {
		err := SaveInitConf(initConf)
		convey.So(err, convey.ShouldNotEqual, nil)
	})
}

func TestIsLegalInitAllocationPoolsOK(t *testing.T) {
	convey.Convey("Test_IsLegalInitAllocationPools_OK_Pool_Len0\n", t, func() {
		pools := make([]subnets.AllocationPool, 0)
		cidr := "100.100.100.0/24"
		gw := "100.100.100.1"
		err := IsLegalInitAllocationPools(pools, cidr, gw)
		convey.So(err, convey.ShouldEqual, true)
	})
	convey.Convey("Test_IsLegalInitAllocationPools_OK_Nomal\n", t, func() {
		pools := []subnets.AllocationPool{
			{
				Start: "100.100.100.100",
				End:   "100.100.100.101",
			},
		}
		cidr := "100.100.100.0/24"
		gw := "100.100.100.1"
		stubs := gostub.StubFunc(&IsCidrLegal, true)
		defer stubs.Reset()
		stubs.StubFunc(&IsFixIPInIPRange, false)
		stubs.StubFunc(&IsAllocationPoolsLegal, true)
		err := IsLegalInitAllocationPools(pools, cidr, gw)
		convey.So(err, convey.ShouldEqual, true)
	})
}

func TestIsLegalInitAllocationPoolsErr(t *testing.T) {
	convey.Convey("Test_IsLegalInitAllocationPools_Err_IsCidrLegal\n", t, func() {
		pools := []subnets.AllocationPool{
			{
				Start: "100.100.100.100",
				End:   "100.100.100.101",
			},
		}
		cidr := "100.100.100.0/37"
		gw := "100.100.100.1"
		stubs := gostub.StubFunc(&IsCidrLegal, false)
		defer stubs.Reset()
		err := IsLegalInitAllocationPools(pools, cidr, gw)
		convey.So(err, convey.ShouldEqual, false)
	})
	convey.Convey("Test_IsLegalInitAllocationPools_Err_IsFixIPInIPRange\n", t, func() {
		pools := []subnets.AllocationPool{
			{
				Start: "100.100.100.1",
				End:   "100.100.100.101",
			},
		}
		cidr := "100.100.100.0/24"
		gw := "100.100.100.1"
		stubs := gostub.StubFunc(&IsCidrLegal, true)
		defer stubs.Reset()
		stubs.StubFunc(&IsFixIPInIPRange, true)
		err := IsLegalInitAllocationPools(pools, cidr, gw)
		convey.So(err, convey.ShouldEqual, false)
	})
	convey.Convey("Test_IsLegalInitAllocationPools_Err_IsAllocationPoolsLegal\n", t, func() {
		pools := []subnets.AllocationPool{
			{
				Start: "100.100.100.102",
				End:   "100.100.100.101",
			},
		}
		cidr := "100.100.100.0/24"
		gw := "100.100.100.1"
		stubs := gostub.StubFunc(&IsCidrLegal, true)
		defer stubs.Reset()
		stubs.StubFunc(&IsFixIPInIPRange, false)
		stubs.StubFunc(&IsAllocationPoolsLegal, false)
		err := IsLegalInitAllocationPools(pools, cidr, gw)
		convey.So(err, convey.ShouldEqual, false)
	})
}

func TestRollBackInitNetsOK(t *testing.T) {
	nets := []string{"netid1", "netid2"}
	mockCtl := gomock.NewController(t)
	defer mockCtl.Finish()
	stubs := gostub.StubFunc(&DeleteNetwork, nil)
	defer stubs.Reset()
	convey.Convey("Test_RollBackInitNets_OK\n", t, func() {
		err := RollBackInitNets(nets)
		convey.So(err, convey.ShouldEqual, nil)
	})
}

func TestRollBackInitConfOK(t *testing.T) {
	monkey.Patch(DelAdminTenantInfoFromDB, func() error {
		return nil
	})
	defer monkey.UnpatchAll()
	monkey.Patch(iaas.DelIaasTenantInfoFromDB, func(_ string) error {
		return nil
	})
	mockCtl := gomock.NewController(t)
	defer mockCtl.Finish()
	stubs := gostub.StubFunc(&DelOpenStackConfg, nil)
	defer stubs.Reset()
	convey.Convey("Test_RollBackInitConf_OK\n", t, func() {
		err := RollBackInitConf("tenantid")
		convey.So(err, convey.ShouldEqual, nil)
	})
}

func TestCrtInitNetworkOK(t *testing.T) {
	crtNet := InitCNetwork{
		Name:            "net1",
		Cidr:            "199.199.199.0/24",
		Desc:            "desc1",
		Public:          true,
		Gw:              "199.199.199.1",
		AllocationPool:  make([]subnets.AllocationPool, 0),
		NetworksType:    "",
		PhysicalNetwork: "",
		SegmentationID:  "",
	}
	mockCtl := gomock.NewController(t)
	defer mockCtl.Finish()
	stubs := gostub.StubFunc(&IsGatewayValid, true)
	defer stubs.Reset()
	stubs.StubFunc(&IsLegalInitAllocationPools, true)
	stubs.StubFunc(&GetNetNumOfTenant, 1)
	mockIaas := test.NewMockIaaS(mockCtl)
	stubs.StubFunc(&iaas.GetIaaS, mockIaas)
	mockIaas.EXPECT().CreateNetwork(gomock.Any()).Return(&iaasaccessor.Network{Id: "netid1"}, nil)
	mockIaas.EXPECT().CreateSubnet("netid1", "199.199.199.0/24", "199.199.199.1", gomock.Any()).
		Return(&iaasaccessor.Subnet{Id: "subnetid1"}, nil)
	mockIaas.EXPECT().GetNetworkExtenAttrs(gomock.Any()).Return(&iaasaccessor.NetworkExtenAttrs{}, nil)
	mockDb := mockdbaccessor.NewMockDbAccessor(mockCtl)
	stubs.StubFunc(&common.GetDataBase, mockDb)
	mockDb.EXPECT().SaveLeaf(gomock.Any(), gomock.Any()).Return(nil)
	mockDb.EXPECT().SaveLeaf(gomock.Any(), gomock.Any()).Return(nil)
	convey.Convey("Test_CrtInitNetwork_OK\n", t, func() {
		_, err := CrtInitNetwork(crtNet)
		convey.So(err, convey.ShouldEqual, nil)
	})
}

func TestCrtInitNetworkErr(t *testing.T) {
	convey.Convey("Test_CrtInitNetwork_Err_IsGatewayValid\n", t, func() {
		crtNet := InitCNetwork{
			Name:            "net1",
			Cidr:            "199.199.199.0/24",
			Desc:            "desc1",
			Public:          true,
			Gw:              "199.199.198.1",
			AllocationPool:  make([]subnets.AllocationPool, 0),
			NetworksType:    "",
			PhysicalNetwork: "",
			SegmentationID:  "",
		}
		stubs := gostub.StubFunc(&IsGatewayValid, false)
		defer stubs.Reset()
		id, err := CrtInitNetwork(crtNet)
		convey.So(err, convey.ShouldNotEqual, nil)
		convey.So(id, convey.ShouldEqual, "")
	})
	convey.Convey("Test_CrtInitNetwork_Err_IsLegalInitAllocationPools\n", t, func() {
		crtNet := InitCNetwork{
			Name:   "net1",
			Cidr:   "199.199.199.0/24",
			Desc:   "desc1",
			Public: true,
			Gw:     "199.199.199.1",
			AllocationPool: []subnets.AllocationPool{
				{
					Start: "199.199.198.2",
					End:   "199.199.198.3",
				},
			},
			NetworksType:    "",
			PhysicalNetwork: "",
			SegmentationID:  "",
		}
		stubs := gostub.StubFunc(&IsGatewayValid, true)
		defer stubs.Reset()
		stubs.StubFunc(&IsLegalInitAllocationPools, false)
		id, err := CrtInitNetwork(crtNet)
		convey.So(err, convey.ShouldNotEqual, nil)
		convey.So(id, convey.ShouldEqual, "")
	})
	convey.Convey("Test_CrtInitNetwork_Err_CheckQuota\n", t, func() {
		crtNet := InitCNetwork{
			Name:            "net1",
			Cidr:            "199.199.199.0/24",
			Desc:            "desc1",
			Public:          true,
			Gw:              "",
			AllocationPool:  make([]subnets.AllocationPool, 0),
			NetworksType:    "",
			PhysicalNetwork: "",
			SegmentationID:  "",
		}
		stubs := gostub.StubFunc(&GetNetNumOfTenant, 500)
		defer stubs.Reset()
		id, err := CrtInitNetwork(crtNet)
		convey.So(err, convey.ShouldNotEqual, nil)
		convey.So(id, convey.ShouldEqual, "")
	})
	convey.Convey("Test_CrtInitNetwork_Err_Create\n", t, func() {
		crtNet := InitCNetwork{
			Name:            "net1",
			Cidr:            "199.199.199.0/24",
			Desc:            "desc1",
			Public:          true,
			Gw:              "",
			AllocationPool:  make([]subnets.AllocationPool, 0),
			NetworksType:    "",
			PhysicalNetwork: "",
			SegmentationID:  "",
		}
		stubs := gostub.StubFunc(&GetNetNumOfTenant, 1)
		defer stubs.Reset()
		mockCtl := gomock.NewController(t)
		defer mockCtl.Finish()
		mockIaas := test.NewMockIaaS(mockCtl)
		stubs.StubFunc(&iaas.GetIaaS, mockIaas)
		mockIaas.EXPECT().CreateNetwork(gomock.Any()).Return(nil, errors.New("create err"))
		id, err := CrtInitNetwork(crtNet)
		convey.So(err, convey.ShouldNotEqual, nil)
		convey.So(id, convey.ShouldEqual, "")
	})
}

func TestGetNeedCrtNets(t *testing.T) {
	crtNets := []*InitCNetwork{
		{
			Name:            "net1",
			Cidr:            "199.199.199.0/24",
			Desc:            "desc1",
			Public:          true,
			Gw:              "199.199.199.1",
			AllocationPool:  make([]subnets.AllocationPool, 0),
			NetworksType:    "",
			PhysicalNetwork: "",
			SegmentationID:  "",
		},
		{
			Name:            "net2",
			Cidr:            "199.199.198.0/24",
			Desc:            "desc2",
			Public:          true,
			Gw:              "199.199.198.1",
			AllocationPool:  make([]subnets.AllocationPool, 0),
			NetworksType:    "",
			PhysicalNetwork: "",
			SegmentationID:  "",
		},
	}
	netsAdmin := []*PaasNetwork{
		{
			Name:        "net1",
			Cidr:        "199.199.199.0/24",
			ExternalNet: false,
		},
	}
	convey.Convey("Test_GetNeedCrtNets\n", t, func() {
		nets := GetNeedCrtNets(crtNets, netsAdmin)
		convey.So(nets[0].Name, convey.ShouldEqual, "net2")
		convey.So(nets[0].Cidr, convey.ShouldEqual, "199.199.198.0/24")
	})
}

func TestCreateInitNetworksOK(t *testing.T) {
	crtNets := []*InitCNetwork{
		{
			Name:            "net1",
			Cidr:            "199.199.199.0/24",
			Desc:            "desc1",
			Public:          true,
			Gw:              "199.199.199.1",
			AllocationPool:  make([]subnets.AllocationPool, 0),
			NetworksType:    "",
			PhysicalNetwork: "",
			SegmentationID:  "",
		},
	}
	mockCtl := gomock.NewController(t)
	defer mockCtl.Finish()
	mockDB := mockdbaccessor.NewMockDbAccessor(mockCtl)
	stubs := gostub.StubFunc(&common.GetDataBase, mockDB)
	defer stubs.Reset()
	stubs.StubFunc(&CrtInitNetwork, "netid1", nil)
	convey.Convey("Test_CreateInitNetworks_OK\n", t, func() {
		err := CreateInitNetworks(crtNets)
		convey.So(err, convey.ShouldEqual, nil)
	})
}

func TestCreateInitNetworksErr(t *testing.T) {
	crtNets := []*InitCNetwork{
		{
			Name:            "net1",
			Cidr:            "199.199.199.0/24",
			Desc:            "desc1",
			Public:          true,
			Gw:              "199.199.199.1",
			AllocationPool:  make([]subnets.AllocationPool, 0),
			NetworksType:    "",
			PhysicalNetwork: "",
			SegmentationID:  "",
		},
	}
	mockCtl := gomock.NewController(t)
	defer mockCtl.Finish()
	mockDB := mockdbaccessor.NewMockDbAccessor(mockCtl)
	stubs := gostub.StubFunc(&common.GetDataBase, mockDB)
	defer stubs.Reset()
	var list []*client.Node
	node1 := client.Node{Key: "netid1", Value: "netinfo1"}
	list = append(list, &node1)
	stubs.StubFunc(&GetNeedCrtNets, crtNets)
	stubs.StubFunc(&CrtInitNetwork, "", errors.New("CrtInitNetwork err"))
	convey.Convey("Test_CreateInitNetworks_Err\n", t, func() {
		err := CreateInitNetworks(crtNets)
		convey.So(err, convey.ShouldNotEqual, nil)
	})
}

func TestCheckCrtNetsOK(t *testing.T) {
	convey.Convey("Test_CreateInitNetworks_OK_TECS\n", t, func() {
		Scene = "TECS"
		crtNets := []*InitCNetwork{
			{
				Name:            "control",
				Cidr:            "199.199.199.0/24",
				Desc:            "desc1",
				Public:          true,
				Gw:              "199.199.199.1",
				AllocationPool:  make([]subnets.AllocationPool, 0),
				NetworksType:    "",
				PhysicalNetwork: "",
				SegmentationID:  "",
			},
			{
				Name:            "media",
				Cidr:            "199.199.198.0/24",
				Desc:            "desc2",
				Public:          true,
				Gw:              "199.199.198.1",
				AllocationPool:  make([]subnets.AllocationPool, 0),
				NetworksType:    "",
				PhysicalNetwork: "",
				SegmentationID:  "",
			},
		}
		nets, err := CheckCrtNets(crtNets)
		convey.So(err, convey.ShouldEqual, nil)
		convey.So(len(nets), convey.ShouldEqual, 2)
	})
	convey.Convey("Test_CreateInitNetworks_OK_VNM\n", t, func() {
		Scene = "vNM"
		crtNets := []*InitCNetwork{
			{
				Name:            "control",
				Cidr:            "199.199.199.0/24",
				Desc:            "desc1",
				Public:          true,
				Gw:              "199.199.199.1",
				AllocationPool:  make([]subnets.AllocationPool, 0),
				NetworksType:    "",
				PhysicalNetwork: "",
				SegmentationID:  "",
			},
			{
				Name:            "media",
				Cidr:            "199.199.198.0/24",
				Desc:            "desc2",
				Public:          true,
				Gw:              "199.199.198.1",
				AllocationPool:  make([]subnets.AllocationPool, 0),
				NetworksType:    "",
				PhysicalNetwork: "",
				SegmentationID:  "",
			},
		}
		nets, err := CheckCrtNets(crtNets)
		convey.So(err, convey.ShouldEqual, nil)
		convey.So(len(nets), convey.ShouldEqual, 2)
	})
	convey.Convey("Test_CreateInitNetworks_OK_EMBEDDED\n", t, func() {
		Scene = "EMBEDDED"
		crtNets := []*InitCNetwork{
			{
				Name:            "control",
				Cidr:            "199.199.199.0/24",
				Desc:            "desc1",
				Public:          true,
				Gw:              "199.199.199.1",
				AllocationPool:  make([]subnets.AllocationPool, 0),
				NetworksType:    "",
				PhysicalNetwork: "",
				SegmentationID:  "",
			},
			{
				Name:            "media",
				Cidr:            "199.199.198.0/24",
				Desc:            "desc2",
				Public:          true,
				Gw:              "199.199.198.1",
				AllocationPool:  make([]subnets.AllocationPool, 0),
				NetworksType:    "",
				PhysicalNetwork: "",
				SegmentationID:  "",
			},
			{
				Name:            "net_api",
				Cidr:            "199.199.197.0/24",
				Desc:            "desc3",
				Public:          true,
				Gw:              "199.199.197.1",
				AllocationPool:  make([]subnets.AllocationPool, 0),
				NetworksType:    "",
				PhysicalNetwork: "",
				SegmentationID:  "",
			},
		}
		nets, err := CheckCrtNets(crtNets)
		convey.So(err, convey.ShouldEqual, nil)
		convey.So(len(nets), convey.ShouldEqual, 3)
	})
}

func TestCheckCrtNetsErr(t *testing.T) {
	convey.Convey("Test_CreateInitNetworks_ERrr_Scene\n", t, func() {
		Scene = "xxx"
		crtNets := []*InitCNetwork{
			{
				Name:            "control",
				Cidr:            "199.199.199.0/24",
				Desc:            "desc1",
				Public:          true,
				Gw:              "199.199.199.1",
				AllocationPool:  make([]subnets.AllocationPool, 0),
				NetworksType:    "",
				PhysicalNetwork: "",
				SegmentationID:  "",
			},
			{
				Name:            "media",
				Cidr:            "199.199.198.0/24",
				Desc:            "desc2",
				Public:          true,
				Gw:              "199.199.198.1",
				AllocationPool:  make([]subnets.AllocationPool, 0),
				NetworksType:    "",
				PhysicalNetwork: "",
				SegmentationID:  "",
			},
		}
		nets, err := CheckCrtNets(crtNets)
		convey.So(err, convey.ShouldNotEqual, nil)
		convey.So(len(nets), convey.ShouldEqual, 0)
	})
	convey.Convey("Test_CreateInitNetworks_ERR_EMBEDDED\n", t, func() {
		Scene = "EMBEDDED"
		crtNets := []*InitCNetwork{
			{
				Name:            "control",
				Cidr:            "199.199.199.0/24",
				Desc:            "desc1",
				Public:          true,
				Gw:              "199.199.199.1",
				AllocationPool:  make([]subnets.AllocationPool, 0),
				NetworksType:    "",
				PhysicalNetwork: "",
				SegmentationID:  "",
			},
			{
				Name:            "media",
				Cidr:            "199.199.198.0/24",
				Desc:            "desc2",
				Public:          true,
				Gw:              "199.199.198.1",
				AllocationPool:  make([]subnets.AllocationPool, 0),
				NetworksType:    "",
				PhysicalNetwork: "",
				SegmentationID:  "",
			},
		}
		nets, err := CheckCrtNets(crtNets)
		convey.So(err, convey.ShouldNotEqual, nil)
		convey.So(len(nets), convey.ShouldEqual, 0)
	})
}

func TestAnalyseCrtNetsOK(t *testing.T) {
	netObjs := make([]*jason.Object, 0)
	crtNets := []*InitCNetwork{
		{
			Name:   "net1",
			Cidr:   "199.199.199.0/24",
			Desc:   "desc1",
			Public: true,
			Gw:     "199.199.199.1",
			AllocationPool: []subnets.AllocationPool{
				{
					Start: "199.199.199.5",
					End:   "199.199.199.7",
				},
			},
			NetworksType:    "",
			PhysicalNetwork: "",
			SegmentationID:  "",
		},
	}
	for _, net := range crtNets {
		netsByte, _ := json.Marshal(net)
		netObj, _ := jason.NewObjectFromBytes(netsByte)
		netObjs = append(netObjs, netObj)
	}
	convey.Convey("Test_AnalyseCrtNets_OK\n", t, func() {
		nets, err := AnalyseCrtNets(netObjs)
		convey.So(err, convey.ShouldEqual, nil)
		convey.So(nets[0].Name, convey.ShouldEqual, "net1")
	})
}

func TestAnalyseCrtNetsErr(t *testing.T) {
	convey.Convey("Test_AnalyseCrtNets_ERR_GetString_name\n", t, func() {
		netObjs := make([]*jason.Object, 0)
		crtNets := []*InitCNetwork{
			{
				//Name: "net1",
				Cidr:   "199.199.199.0/24",
				Desc:   "desc1",
				Public: true,
				Gw:     "199.199.199.1",
				AllocationPool: []subnets.AllocationPool{
					{
						Start: "199.199.199.5",
						End:   "199.199.199.7",
					},
				},
				NetworksType:    "",
				PhysicalNetwork: "",
				SegmentationID:  "",
			},
		}
		for _, net := range crtNets {
			netsByte, _ := json.Marshal(net)
			netObj, _ := jason.NewObjectFromBytes(netsByte)
			netObjs = append(netObjs, netObj)
		}
		nets, err := AnalyseCrtNets(netObjs)
		convey.So(err, convey.ShouldNotEqual, nil)
		convey.So(nets, convey.ShouldEqual, nil)
	})
	convey.Convey("Test_AnalyseCrtNets_ERR_GetString_cidr\n", t, func() {
		netObjs := make([]*jason.Object, 0)
		crtNets := []*InitCNetwork{
			{
				Name: "net1",
				//Cidr: "199.199.199.0/24",
				Desc:   "desc1",
				Public: true,
				Gw:     "199.199.199.1",
				AllocationPool: []subnets.AllocationPool{
					{
						Start: "199.199.199.5",
						End:   "199.199.199.7",
					},
				},
				NetworksType:    "",
				PhysicalNetwork: "",
				SegmentationID:  "",
			},
		}
		for _, net := range crtNets {
			netsByte, _ := json.Marshal(net)
			netObj, _ := jason.NewObjectFromBytes(netsByte)
			netObjs = append(netObjs, netObj)
		}
		nets, err := AnalyseCrtNets(netObjs)
		convey.So(err, convey.ShouldNotEqual, nil)
		convey.So(nets, convey.ShouldEqual, nil)
	})
	convey.Convey("Test_AnalyseCrtNets_ERR_GetString_start\n", t, func() {
		netObjs := make([]*jason.Object, 0)
		crtNets := []*InitCNetwork{
			{
				Name:   "net1",
				Cidr:   "199.199.199.0/24",
				Desc:   "desc1",
				Public: true,
				Gw:     "199.199.199.1",
				AllocationPool: []subnets.AllocationPool{
					{
						//Start: "199.199.199.5",
						End: "199.199.199.7",
					},
				},
				NetworksType:    "",
				PhysicalNetwork: "",
				SegmentationID:  "",
			},
		}
		for _, net := range crtNets {
			netsByte, _ := json.Marshal(net)
			netObj, _ := jason.NewObjectFromBytes(netsByte)
			netObjs = append(netObjs, netObj)
		}
		nets, err := AnalyseCrtNets(netObjs)
		convey.So(err, convey.ShouldNotEqual, nil)
		convey.So(nets, convey.ShouldEqual, nil)
	})
	convey.Convey("Test_AnalyseCrtNets_ERR_GetString_end\n", t, func() {
		netObjs := make([]*jason.Object, 0)
		crtNets := []*InitCNetwork{
			{
				Name:   "net1",
				Cidr:   "199.199.199.0/24",
				Desc:   "desc1",
				Public: true,
				Gw:     "199.199.199.1",
				AllocationPool: []subnets.AllocationPool{
					{
						Start: "199.199.199.5",
						//End: "199.199.199.7",
					},
				},
				NetworksType:    "",
				PhysicalNetwork: "",
				SegmentationID:  "",
			},
		}
		for _, net := range crtNets {
			netsByte, _ := json.Marshal(net)
			netObj, _ := jason.NewObjectFromBytes(netsByte)
			netObjs = append(netObjs, netObj)
		}
		nets, err := AnalyseCrtNets(netObjs)
		convey.So(err, convey.ShouldNotEqual, nil)
		convey.So(nets, convey.ShouldEqual, nil)
	})
}

func TestInitCNetworksOK(t *testing.T) {
	crtNets := []*InitCNetwork{
		{
			Name:   "net1",
			Cidr:   "199.199.199.0/24",
			Desc:   "desc1",
			Public: true,
			Gw:     "199.199.199.1",
			AllocationPool: []subnets.AllocationPool{
				{
					Start: "199.199.199.5",
					End:   "199.199.199.7",
				},
			},
			NetworksType:    "",
			PhysicalNetwork: "",
			SegmentationID:  "",
		},
	}
	initNets := InitNetworks{
		InitRegNetworks:    make([]*InitRNetwork, 0),
		InitCreateNetworks: crtNets,
	}
	netsByte, _ := json.Marshal(initNets)
	netObj, _ := jason.NewObjectFromBytes(netsByte)
	mockCtl := gomock.NewController(t)
	defer mockCtl.Finish()
	stubs := gostub.StubFunc(&AnalyseCrtNets, crtNets, nil)
	defer stubs.Reset()
	stubs.StubFunc(&CheckCrtNets, crtNets, nil)
	stubs.StubFunc(&CreateInitNetworks, nil)
	convey.Convey("Test_InitCNetworks_OK\n", t, func() {
		err := InitCNetworks(netObj)
		convey.So(err, convey.ShouldEqual, nil)
	})
}

func TestInitCNetworksErr(t *testing.T) {
	convey.Convey("Test_InitCNetworks_ERR_AnalyseCrtNets\n", t, func() {
		crtNets := []*InitCNetwork{
			{
				Name:   "net1",
				Cidr:   "199.199.199.0/24",
				Desc:   "desc1",
				Public: true,
				Gw:     "199.199.199.1",
				AllocationPool: []subnets.AllocationPool{
					{
						//Start: "199.199.199.5",
						End: "199.199.199.7",
					},
				},
				NetworksType:    "",
				PhysicalNetwork: "",
				SegmentationID:  "",
			},
		}
		initNets := InitNetworks{
			InitRegNetworks:    make([]*InitRNetwork, 0),
			InitCreateNetworks: crtNets,
		}
		netsByte, _ := json.Marshal(initNets)
		netObj, _ := jason.NewObjectFromBytes(netsByte)
		mockCtl := gomock.NewController(t)
		defer mockCtl.Finish()
		stubs := gostub.StubFunc(&AnalyseCrtNets, nil, errors.New("AnalyseCrtNets err"))
		defer stubs.Reset()
		err := InitCNetworks(netObj)
		convey.So(err, convey.ShouldNotEqual, nil)
	})
	convey.Convey("Test_InitCNetworks_ERR_CheckCrtNets\n", t, func() {
		crtNets := []*InitCNetwork{
			{
				Name:   "net1",
				Cidr:   "199.199.199.0/24",
				Desc:   "desc1",
				Public: true,
				Gw:     "199.199.199.1",
				AllocationPool: []subnets.AllocationPool{
					{
						Start: "199.199.199.5",
						End:   "199.199.199.7",
					},
				},
				NetworksType:    "",
				PhysicalNetwork: "",
				SegmentationID:  "",
			},
		}
		initNets := InitNetworks{
			InitRegNetworks:    make([]*InitRNetwork, 0),
			InitCreateNetworks: crtNets,
		}
		netsByte, _ := json.Marshal(initNets)
		netObj, _ := jason.NewObjectFromBytes(netsByte)
		mockCtl := gomock.NewController(t)
		defer mockCtl.Finish()
		stubs := gostub.StubFunc(&AnalyseCrtNets, crtNets, nil)
		defer stubs.Reset()
		stubs.StubFunc(&CheckCrtNets, make([]*InitCNetwork, 0), errors.New("CheckCrtNets err"))
		err := InitCNetworks(netObj)
		convey.So(err, convey.ShouldNotEqual, nil)
	})
	convey.Convey("Test_InitCNetworks_ERR_CreateInitNetworks\n", t, func() {
		crtNets := []*InitCNetwork{
			{
				Name:   "net1",
				Cidr:   "199.199.199.0/24",
				Desc:   "desc1",
				Public: true,
				Gw:     "199.199.199.1",
				AllocationPool: []subnets.AllocationPool{
					{
						Start: "199.199.199.5",
						End:   "199.199.199.7",
					},
				},
				NetworksType:    "",
				PhysicalNetwork: "",
				SegmentationID:  "",
			},
		}
		initNets := InitNetworks{
			InitRegNetworks:    make([]*InitRNetwork, 0),
			InitCreateNetworks: crtNets,
		}
		netsByte, _ := json.Marshal(initNets)
		netObj, _ := jason.NewObjectFromBytes(netsByte)
		mockCtl := gomock.NewController(t)
		defer mockCtl.Finish()
		stubs := gostub.StubFunc(&AnalyseCrtNets, crtNets, nil)
		defer stubs.Reset()
		stubs.StubFunc(&CheckCrtNets, crtNets, nil)
		stubs.StubFunc(&CreateInitNetworks, errors.New("CreateInitNetworks err"))
		err := InitCNetworks(netObj)
		convey.So(err, convey.ShouldNotEqual, nil)
	})
}

func TestRegInitNetworkOK(t *testing.T) {
	crtNet := InitRNetwork{
		Name:   "net1",
		UUID:   "netid1",
		Desc:   "desc1",
		Public: true,
	}
	mockCtl := gomock.NewController(t)
	defer mockCtl.Finish()
	stubs := gostub.StubFunc(&GetNetNumOfTenant, 1)
	defer stubs.Reset()
	mockIaas := test.NewMockIaaS(mockCtl)
	stubs.StubFunc(&iaas.GetIaaS, mockIaas)
	mockIaas.EXPECT().GetSubnetID(crtNet.UUID).Return("subid1", nil)
	mockIaas.EXPECT().GetNetwork(crtNet.UUID).Return(&iaasaccessor.Network{}, nil)
	mockIaas.EXPECT().GetSubnet("subid1").Return(&iaasaccessor.Subnet{}, nil)
	mockIaas.EXPECT().GetNetworkExtenAttrs(gomock.Any()).Return(&iaasaccessor.NetworkExtenAttrs{}, nil)
	mockDb := mockdbaccessor.NewMockDbAccessor(mockCtl)
	stubs.StubFunc(&common.GetDataBase, mockDb)
	mockDb.EXPECT().SaveLeaf(gomock.Any(), gomock.Any()).Return(nil)
	mockDb.EXPECT().SaveLeaf(gomock.Any(), gomock.Any()).Return(nil)
	convey.Convey("Test_RegInitNetwork_OK\n", t, func() {
		_, err := RegInitNetwork(crtNet)
		convey.So(err, convey.ShouldEqual, nil)
	})
}

func TestRegInitNetworkErr(t *testing.T) {
	convey.Convey("Test_RegInitNetwork_ERR_CheckQuota\n", t, func() {
		crtNet := InitRNetwork{
			Name:   "net1",
			UUID:   "netid1",
			Desc:   "desc1",
			Public: true,
		}
		mockCtl := gomock.NewController(t)
		defer mockCtl.Finish()
		stubs := gostub.StubFunc(&GetNetNumOfTenant, 500)
		defer stubs.Reset()
		_, err := RegInitNetwork(crtNet)
		convey.So(err, convey.ShouldNotEqual, nil)
	})
	convey.Convey("Test_RegInitNetwork_ERR_GetSubnetID\n", t, func() {
		crtNet := InitRNetwork{
			Name:   "net1",
			UUID:   "netid1",
			Desc:   "desc1",
			Public: true,
		}
		mockCtl := gomock.NewController(t)
		defer mockCtl.Finish()
		stubs := gostub.StubFunc(&GetNetNumOfTenant, 1)
		defer stubs.Reset()
		mockIaas := test.NewMockIaaS(mockCtl)
		stubs.StubFunc(&iaas.GetIaaS, mockIaas)
		mockIaas.EXPECT().GetSubnetID(crtNet.UUID).Return("", errors.New("GetSubnetID err"))
		_, err := RegInitNetwork(crtNet)
		convey.So(err, convey.ShouldNotEqual, nil)
	})
	convey.Convey("Test_RegInitNetwork_ERR_getNetWorkInfo\n", t, func() {
		crtNet := InitRNetwork{
			Name:   "net1",
			UUID:   "netid1",
			Desc:   "desc1",
			Public: true,
		}
		mockCtl := gomock.NewController(t)
		defer mockCtl.Finish()
		stubs := gostub.StubFunc(&GetNetNumOfTenant, 1)
		defer stubs.Reset()
		mockIaas := test.NewMockIaaS(mockCtl)
		stubs.StubFunc(&iaas.GetIaaS, mockIaas)
		mockIaas.EXPECT().GetSubnetID(crtNet.UUID).Return("subid1", nil)
		mockIaas.EXPECT().GetNetwork(crtNet.UUID).Return(nil, errors.New("GetNetwork err"))
		_, err := RegInitNetwork(crtNet)
		convey.So(err, convey.ShouldNotEqual, nil)
	})
	convey.Convey("Test_RegInitNetwork_ERR_saveNetworkToEtcd\n", t, func() {
		crtNet := InitRNetwork{
			Name:   "net1",
			UUID:   "netid1",
			Desc:   "desc1",
			Public: true,
		}
		mockCtl := gomock.NewController(t)
		defer mockCtl.Finish()
		stubs := gostub.StubFunc(&GetNetNumOfTenant, 1)
		defer stubs.Reset()
		mockIaas := test.NewMockIaaS(mockCtl)
		stubs.StubFunc(&iaas.GetIaaS, mockIaas)
		mockIaas.EXPECT().GetSubnetID(crtNet.UUID).Return("subid1", nil)
		mockIaas.EXPECT().GetNetwork(crtNet.UUID).Return(&iaasaccessor.Network{}, nil)
		mockIaas.EXPECT().GetSubnet("subid1").Return(&iaasaccessor.Subnet{}, nil)
		mockIaas.EXPECT().GetNetworkExtenAttrs(gomock.Any()).Return(&iaasaccessor.NetworkExtenAttrs{}, nil)
		mockDb := mockdbaccessor.NewMockDbAccessor(mockCtl)
		stubs.StubFunc(&common.GetDataBase, mockDb)
		mockDb.EXPECT().SaveLeaf(gomock.Any(), gomock.Any()).Return(errors.New("SaveLeaf err"))
		_, err := RegInitNetwork(crtNet)
		convey.So(err, convey.ShouldNotEqual, nil)
	})
}

func TestGetNeedRegNets(t *testing.T) {
	regNets := []*InitRNetwork{
		{
			Name:   "net1",
			UUID:   "netid1",
			Desc:   "desc1",
			Public: true,
		},
		{
			Name:   "net2",
			UUID:   "netid2",
			Desc:   "desc2",
			Public: true,
		},
	}
	netsAdmin := []*PaasNetwork{
		{
			Name:        "net1",
			ID:          "netid1",
			ExternalNet: true,
		},
	}
	convey.Convey("Test_GetNeedRegNets\n", t, func() {
		nets := GetNeedRegNets(regNets, netsAdmin)
		convey.So(nets[0].Name, convey.ShouldEqual, "net2")
		convey.So(nets[0].UUID, convey.ShouldEqual, "netid2")
	})
}

func TestRegisterInitNetworksOK(t *testing.T) {
	regNets := []*InitRNetwork{
		{
			Name:   "net1",
			UUID:   "netid1",
			Desc:   "desc1",
			Public: true,
		},
	}
	mockCtl := gomock.NewController(t)
	defer mockCtl.Finish()
	mockDB := mockdbaccessor.NewMockDbAccessor(mockCtl)
	stubs := gostub.StubFunc(&common.GetDataBase, mockDB)
	defer stubs.Reset()
	stubs.StubFunc(&RegInitNetwork, "netid1", nil)
	convey.Convey("Test_RegisterInitNetworks_OK\n", t, func() {
		err := RegisterInitNetworks(regNets)
		convey.So(err, convey.ShouldEqual, nil)
	})
}

func TestRegisterInitNetworksErr(t *testing.T) {
	regNets := []*InitRNetwork{
		{
			Name:   "net1",
			UUID:   "netid1",
			Desc:   "desc1",
			Public: true,
		},
	}
	mockCtl := gomock.NewController(t)
	defer mockCtl.Finish()
	mockDB := mockdbaccessor.NewMockDbAccessor(mockCtl)
	stubs := gostub.StubFunc(&common.GetDataBase, mockDB)
	defer stubs.Reset()
	var list []*client.Node
	node1 := client.Node{Key: "netid1", Value: "netinfo1"}
	list = append(list, &node1)
	stubs.StubFunc(&GetNeedRegNets, regNets)
	stubs.StubFunc(&RegInitNetwork, "", errors.New("RegInitNetwork err"))
	convey.Convey("Test_RegisterInitNetworks_Err_RegInitNetwork\n", t, func() {
		err := RegisterInitNetworks(regNets)
		convey.So(err, convey.ShouldNotEqual, nil)
	})
}

func TestCheckRegNetsOK(t *testing.T) {
	regNets := []*InitRNetwork{
		{
			Name:   "net_api",
			UUID:   "net_api_id",
			Desc:   "desc1",
			Public: true,
		},
		{
			Name:   "net_mgt",
			UUID:   "net_mgt_id",
			Desc:   "desc1",
			Public: true,
		},
	}
	convey.Convey("Test_CheckRegNets_OK\n", t, func() {
		err := CheckRegNets(regNets)
		convey.So(err, convey.ShouldEqual, nil)
	})
}

func TestCheckRegNetsErr(t *testing.T) {
	regNets := []*InitRNetwork{
		{
			Name:   "net_api",
			UUID:   "net_api_id",
			Desc:   "desc1",
			Public: true,
		},
	}
	convey.Convey("Test_CheckRegNets_ERR\n", t, func() {
		err := CheckRegNets(regNets)
		convey.So(err, convey.ShouldNotEqual, nil)
	})
}

func TestAnalyseRegNetsOK(t *testing.T) {
	netObjs := make([]*jason.Object, 0)
	regNets := []*InitRNetwork{
		{
			Name:   "net1",
			UUID:   "netid1",
			Desc:   "desc1",
			Public: true,
		},
	}
	for _, net := range regNets {
		netsByte, _ := json.Marshal(net)
		netObj, _ := jason.NewObjectFromBytes(netsByte)
		netObjs = append(netObjs, netObj)
	}
	convey.Convey("Test_AnalyseRegNets_OK\n", t, func() {
		nets, err := AnalyseRegNets(netObjs)
		convey.So(err, convey.ShouldEqual, nil)
		convey.So(nets[0].Name, convey.ShouldEqual, "net1")
	})
}

func TestAnalyseRegNetsErr(t *testing.T) {
	convey.Convey("Test_AnalyseRegNets_ERR_GetString_uuid\n", t, func() {
		netObjs := make([]*jason.Object, 0)
		regNets := []*InitRNetwork{
			{
				Name: "net1",
				//UUID: "netid1",
				Desc:   "desc1",
				Public: true,
			},
		}
		for _, net := range regNets {
			netsByte, _ := json.Marshal(net)
			netObj, _ := jason.NewObjectFromBytes(netsByte)
			netObjs = append(netObjs, netObj)
		}
		nets, err := AnalyseRegNets(netObjs)
		convey.So(err, convey.ShouldNotEqual, nil)
		convey.So(nets, convey.ShouldEqual, nil)
	})
}

func TestInitRNetworksOK(t *testing.T) {
	regNets := []*InitRNetwork{
		{
			Name:   "net1",
			UUID:   "netid1",
			Desc:   "desc1",
			Public: true,
		},
	}
	initNets := InitNetworks{
		InitRegNetworks:    regNets,
		InitCreateNetworks: make([]*InitCNetwork, 0),
	}
	netsByte, _ := json.Marshal(initNets)
	netObj, _ := jason.NewObjectFromBytes(netsByte)
	mockCtl := gomock.NewController(t)
	defer mockCtl.Finish()
	stubs := gostub.StubFunc(&AnalyseRegNets, regNets, nil)
	defer stubs.Reset()
	stubs.StubFunc(&CheckRegNets, nil)
	stubs.StubFunc(&RegisterInitNetworks, nil)
	convey.Convey("Test_InitRNetworks_OK\n", t, func() {
		err := InitRNetworks(netObj)
		convey.So(err, convey.ShouldEqual, nil)
	})
}

func TestInitRNetworksErr(t *testing.T) {
	convey.Convey("Test_InitRNetworks_ERR_GetObjectArray_registered_networks\n", t, func() {
		initNets := InitNetworks{
			InitRegNetworks:    make([]*InitRNetwork, 0),
			InitCreateNetworks: make([]*InitCNetwork, 0),
		}
		netsByte, _ := json.Marshal(initNets)
		netObj, _ := jason.NewObjectFromBytes(netsByte)
		err := InitRNetworks(netObj)
		convey.So(err, convey.ShouldNotEqual, nil)
	})
	convey.Convey("Test_InitRNetworks_ERR_AnalyseRegNets\n", t, func() {
		regNets := []*InitRNetwork{
			{
				Name:   "net1",
				UUID:   "netid1",
				Desc:   "desc1",
				Public: true,
			},
		}
		initNets := InitNetworks{
			InitRegNetworks:    regNets,
			InitCreateNetworks: make([]*InitCNetwork, 0),
		}
		netsByte, _ := json.Marshal(initNets)
		netObj, _ := jason.NewObjectFromBytes(netsByte)
		mockCtl := gomock.NewController(t)
		defer mockCtl.Finish()
		stubs := gostub.StubFunc(&AnalyseRegNets, nil, errors.New("AnalyseRegNets err"))
		defer stubs.Reset()
		err := InitRNetworks(netObj)
		convey.So(err, convey.ShouldNotEqual, nil)
	})
	convey.Convey("Test_InitRNetworks_ERR_CheckRegNets\n", t, func() {
		regNets := []*InitRNetwork{
			{
				Name:   "net1",
				UUID:   "netid1",
				Desc:   "desc1",
				Public: true,
			},
		}
		initNets := InitNetworks{
			InitRegNetworks:    regNets,
			InitCreateNetworks: make([]*InitCNetwork, 0),
		}
		netsByte, _ := json.Marshal(initNets)
		netObj, _ := jason.NewObjectFromBytes(netsByte)
		mockCtl := gomock.NewController(t)
		defer mockCtl.Finish()
		stubs := gostub.StubFunc(&AnalyseRegNets, regNets, nil)
		defer stubs.Reset()
		stubs.StubFunc(&CheckRegNets, errors.New("CheckRegNets err"))
		err := InitRNetworks(netObj)
		convey.So(err, convey.ShouldNotEqual, nil)
	})
	convey.Convey("Test_InitRNetworks_ERR_RegisterInitNetworks\n", t, func() {
		regNets := []*InitRNetwork{
			{
				Name:   "net1",
				UUID:   "netid1",
				Desc:   "desc1",
				Public: true,
			},
		}
		initNets := InitNetworks{
			InitRegNetworks:    regNets,
			InitCreateNetworks: make([]*InitCNetwork, 0),
		}
		netsByte, _ := json.Marshal(initNets)
		netObj, _ := jason.NewObjectFromBytes(netsByte)
		mockCtl := gomock.NewController(t)
		defer mockCtl.Finish()
		stubs := gostub.StubFunc(&AnalyseRegNets, regNets, nil)
		defer stubs.Reset()
		stubs.StubFunc(&CheckRegNets, nil)
		stubs.StubFunc(&RegisterInitNetworks, errors.New("RegisterInitNetworks err"))
		err := InitRNetworks(netObj)
		convey.So(err, convey.ShouldNotEqual, nil)
	})
}

func TestHandleInitNetworksOK(t *testing.T) {
	initNets := InitNetworks{
		InitRegNetworks:    make([]*InitRNetwork, 0),
		InitCreateNetworks: make([]*InitCNetwork, 0),
	}
	initcfg := InitCfg{
		Networks: initNets,
	}
	initByte, _ := json.Marshal(initcfg)
	initObj, _ := jason.NewObjectFromBytes(initByte)
	mockCtl := gomock.NewController(t)
	defer mockCtl.Finish()
	stubs := gostub.StubFunc(&iaas.GetSceneByKnitterJSON, "TECS", nil)
	defer stubs.Reset()
	stubs.StubFunc(&InitRNetworks, nil)
	stubs.StubFunc(&InitCNetworks, nil)
	convey.Convey("Test_HandleInitNetworks_OK\n", t, func() {
		err := HandleInitNetworks(initObj)
		convey.So(err, convey.ShouldEqual, nil)
	})
}

func TestAuthInitCfgOK(t *testing.T) {
	initConf := &InitConfiguration{
		EndPoint:   "endpoint1",
		User:       "user1",
		Password:   "password1",
		TenantName: "tenantname1",
		TenantID:   "tenantid1",
	}
	stubs := gostub.StubFunc(&iaas.CheckOpenstackConfig, &gophercloud.ProviderClient{TenantID: "tenantid1", TenantName: "tenantname1"}, nil)
	defer stubs.Reset()
	convey.Convey("Test_AuthInitCfg_OK\n", t, func() {
		_, err := AuthInitCfg(initConf)
		convey.So(err, convey.ShouldEqual, nil)
	})
}

func TestAuthInitCfgErr(t *testing.T) {
	initConf := &InitConfiguration{
		EndPoint:   "endpoint1",
		User:       "user1",
		Password:   "password1",
		TenantName: "tenantname1",
		TenantID:   "tenantid1",
	}
	stubs := gostub.StubFunc(&iaas.CheckOpenstackConfig, nil, errors.New("CheckOpenstackConfig err"))
	defer stubs.Reset()
	convey.Convey("Test_AuthInitCfg_ERR_CheckOpenstackConfig\n", t, func() {
		_, err := AuthInitCfg(initConf)
		convey.So(err, convey.ShouldNotEqual, nil)
	})
}

func TestAnalyseInitConfigurationOK(t *testing.T) {
	initConf := &InitConfiguration{
		EndPoint:   "endpoint1",
		User:       "user1",
		Password:   "password1",
		TenantName: "tenantname1",
		TenantID:   "tenantid1",
	}
	cfgByte, _ := json.Marshal(initConf)
	cfgObj, _ := jason.NewObjectFromBytes(cfgByte)
	convey.Convey("Test_AnalyseInitConfiguration_OK\n", t, func() {
		cfg, err := AnalyseInitConfiguration(cfgObj)
		convey.So(err, convey.ShouldEqual, nil)
		convey.So(cfg.EndPoint, convey.ShouldEqual, "endpoint1")
	})
}

func TestAnalyseInitConfigurationErr(t *testing.T) {
	convey.Convey("Test_AnalyseInitConfiguration_ERR_GetString_endpoint\n", t, func() {
		initConf := &InitConfiguration{
			//EndPoint: "endpoint1",
			User:       "user1",
			Password:   "password1",
			TenantName: "tenantname1",
			TenantID:   "tenantid1",
		}
		cfgByte, _ := json.Marshal(initConf)
		cfgObj, _ := jason.NewObjectFromBytes(cfgByte)
		cfg, err := AnalyseInitConfiguration(cfgObj)
		convey.So(err, convey.ShouldNotEqual, nil)
		convey.So(cfg, convey.ShouldEqual, nil)
	})
	convey.Convey("Test_AnalyseInitConfiguration_ERR_GetString_user\n", t, func() {
		initConf := &InitConfiguration{
			EndPoint: "endpoint1",
			//User: "user1",
			Password:   "password1",
			TenantName: "tenantname1",
			TenantID:   "tenantid1",
		}
		cfgByte, _ := json.Marshal(initConf)
		cfgObj, _ := jason.NewObjectFromBytes(cfgByte)
		cfg, err := AnalyseInitConfiguration(cfgObj)
		convey.So(err, convey.ShouldNotEqual, nil)
		convey.So(cfg, convey.ShouldEqual, nil)
	})
	convey.Convey("Test_AnalyseInitConfiguration_ERR_GetString_password\n", t, func() {
		initConf := &InitConfiguration{
			EndPoint: "endpoint1",
			User:     "user1",
			//Password: "password1",
			TenantName: "tenantname1",
			TenantID:   "tenantid1",
		}
		cfgByte, _ := json.Marshal(initConf)
		cfgObj, _ := jason.NewObjectFromBytes(cfgByte)
		cfg, err := AnalyseInitConfiguration(cfgObj)
		convey.So(err, convey.ShouldNotEqual, nil)
		convey.So(cfg, convey.ShouldEqual, nil)
	})
}

func TestGetInitConfiguratonOK(t *testing.T) {
	initConfiguration := InitConfiguration{
		EndPoint:   "endpoint1",
		User:       "user1",
		Password:   "password1",
		TenantName: "tenantname1",
		TenantID:   "tenantid1",
	}
	initcfg := InitCfg{
		Configuration: initConfiguration,
	}
	initByte, _ := json.Marshal(initcfg)
	initObj, _ := jason.NewObjectFromBytes(initByte)
	stubs := gostub.StubFunc(&AnalyseInitConfiguration, &initConfiguration, nil)
	defer stubs.Reset()
	stubs.StubFunc(&AuthInitCfg, &initConfiguration, nil)
	convey.Convey("Test_GetInitConfiguraton_OK\n", t, func() {
		cfg, err := GetInitConfiguraton(initObj)
		convey.So(err, convey.ShouldEqual, nil)
		convey.So(cfg.EndPoint, convey.ShouldEqual, "endpoint1")
	})
}

func TestGetInitConfiguratonErr(t *testing.T) {
	convey.Convey("Test_GetInitConfiguraton_ERR_GetObject_configuration\n", t, func() {
		type InitBadCfg struct {
			BadConfiguration InitConfiguration `json:"bad_configuration"`
			Networks         InitNetworks      `json:"networks"`
		}
		initConfiguration := InitConfiguration{
			EndPoint:   "endpoint1",
			User:       "user1",
			Password:   "password1",
			TenantName: "tenantname1",
			TenantID:   "tenantid1",
		}
		initcfg := InitBadCfg{
			BadConfiguration: initConfiguration,
		}
		initByte, _ := json.Marshal(initcfg)
		initObj, _ := jason.NewObjectFromBytes(initByte)
		cfg, err := GetInitConfiguraton(initObj)
		convey.So(err, convey.ShouldNotEqual, nil)
		convey.So(cfg, convey.ShouldEqual, nil)
	})
	convey.Convey("Test_GetInitConfiguraton_ERR_AnalyseInitConfiguration\n", t, func() {
		initConfiguration := InitConfiguration{
			EndPoint:   "endpoint1",
			User:       "user1",
			Password:   "password1",
			TenantName: "tenantname1",
			TenantID:   "tenantid1",
		}
		initcfg := InitCfg{
			Configuration: initConfiguration,
		}
		initByte, _ := json.Marshal(initcfg)
		initObj, _ := jason.NewObjectFromBytes(initByte)
		stubs := gostub.StubFunc(&AnalyseInitConfiguration, nil, errors.New("AuthInitCfg err"))
		defer stubs.Reset()
		cfg, err := GetInitConfiguraton(initObj)
		convey.So(err, convey.ShouldNotEqual, nil)
		convey.So(cfg, convey.ShouldEqual, nil)
	})
	convey.Convey("Test_GetInitConfiguraton_ERR_AuthInitCfg\n", t, func() {
		initConfiguration := InitConfiguration{
			EndPoint:   "endpoint1",
			User:       "user1",
			Password:   "password1",
			TenantName: "tenantname1",
			TenantID:   "tenantid1",
		}
		initcfg := InitCfg{
			Configuration: initConfiguration,
		}
		initByte, _ := json.Marshal(initcfg)
		initObj, _ := jason.NewObjectFromBytes(initByte)
		stubs := gostub.StubFunc(&AnalyseInitConfiguration, &initConfiguration, nil)
		defer stubs.Reset()
		stubs.StubFunc(&AuthInitCfg, nil, errors.New("AuthInitCfg err"))
		cfg, err := GetInitConfiguraton(initObj)
		convey.So(err, convey.ShouldNotEqual, nil)
		convey.So(cfg, convey.ShouldEqual, nil)
	})
}

func TestRecoverInitNetworkErr(t *testing.T) {
	convey.Convey("Test_RecoverInitNetwork_ERR_ReadInitConf\n", t, func() {
		stubs := gostub.StubFunc(&ReadInitConf, "", errors.New("ReadInitConf err"))
		defer stubs.Reset()
		err := RecoverInitNetwork()
		convey.So(err, convey.ShouldNotEqual, nil)
	})
	convey.Convey("Test_RecoverInitNetwork_ERR_NewObjectFromBytes\n", t, func() {
		bodyContent := string(`{"init_configuration": "xxx}`)
		stubs := gostub.StubFunc(&ReadInitConf, bodyContent, nil)
		defer stubs.Reset()
		err := RecoverInitNetwork()
		convey.So(err, convey.ShouldNotEqual, nil)
	})
	convey.Convey("Test_RecoverInitNetwork_ERR_NewObjectFromBytes\n", t, func() {
		bodyContent := string(`{"init_configuration": "xxx"}`)
		stubs := gostub.StubFunc(&ReadInitConf, bodyContent, nil)
		defer stubs.Reset()
		err := RecoverInitNetwork()
		convey.So(err, convey.ShouldNotEqual, nil)
	})
	convey.Convey("Test_RecoverInitNetwork_ERR_HandleInitNetworks\n", t, func() {
		bodyContent := string(`{
		"init_configuration": {
		        "configuration": "configuration_obj",
		        "Networks": "networks_obj"
			}
		}`)
		stubs := gostub.StubFunc(&ReadInitConf, bodyContent, nil)
		defer stubs.Reset()
		stubs.StubFunc(&HandleInitNetworks, errors.New("HandleInitNetworks err"))
		err := RecoverInitNetwork()
		convey.So(err, convey.ShouldNotEqual, nil)
	})
}

func TestReadInitConfOK(t *testing.T) {
	bodyContent := string(`{"init_configuration": "xxx"}`)
	mockCtl := gomock.NewController(t)
	defer mockCtl.Finish()
	mockDB := mockdbaccessor.NewMockDbAccessor(mockCtl)
	stubs := gostub.StubFunc(&common.GetDataBase, mockDB)
	defer stubs.Reset()
	mockDB.EXPECT().ReadLeaf(gomock.Any()).Return(bodyContent, nil)
	convey.Convey("Test_ReadInitConf_OK\n", t, func() {
		_, err := ReadInitConf()
		convey.So(err, convey.ShouldEqual, nil)
	})
}

func TestReadInitConfErr(t *testing.T) {
	mockCtl := gomock.NewController(t)
	defer mockCtl.Finish()
	mockDB := mockdbaccessor.NewMockDbAccessor(mockCtl)
	stubs := gostub.StubFunc(&common.GetDataBase, mockDB)
	defer stubs.Reset()
	mockDB.EXPECT().ReadLeaf(gomock.Any()).Return("", errors.New("ReadLeaf err"))
	convey.Convey("Test_ReadInitConf_Err\n", t, func() {
		_, err := ReadInitConf()
		convey.So(err, convey.ShouldNotEqual, nil)
	})
}

func TestGetDefaultProviderNetworkByInitConfig(t *testing.T) {
	bodyContent := `{
		"configuration": {
			"default_physnet": "physnet1"
		}
	}`
	initJSON, _ := jason.NewObjectFromBytes([]byte(bodyContent))
	convey.Convey("Test GetDefaultProviderNetworkByInitConfig OK\n", t, func() {
		phy, err := GetDefaultProviderNetworkByInitConfig(initJSON)
		convey.So(err, convey.ShouldEqual, nil)
		convey.So(phy, convey.ShouldEqual, "physnet1")
	})
}

func TestGetDefaultProviderNetworkByInitConfigErr(t *testing.T) {
	bodyContent := `{
		"configuration": {
			"default_physne": "physnet1"
		}
	}`
	initJSON, _ := jason.NewObjectFromBytes([]byte(bodyContent))
	convey.Convey("Test GetDefaultProviderNetworkByInitConfig OK\n", t, func() {
		_, err := GetDefaultProviderNetworkByInitConfig(initJSON)
		convey.So(err, convey.ShouldNotEqual, nil)
	})
}

func Test_SaveAdminTenantInfoToDB_nil(t *testing.T) {
	iaasID := "iaasTenentID1"
	iaasName := "iaasName"
	tenant := &Tenant{}
	monkey.PatchInstanceMethod(reflect.TypeOf(tenant), "SaveTenantToEtcd",
		func(_ *Tenant) error {
			return nil
		})
	defer monkey.UnpatchAll()
	convey.Convey("Test SaveAdminTenantInfoToDB OK\n", t, func() {
		err := SaveAdminTenantInfoToDB(iaasID, iaasName)
		convey.So(err, convey.ShouldBeNil)
	})
}

func Test_DelAdminTenantInfoFromDB_nil(t *testing.T) {
	mockCtl := gomock.NewController(t)
	defer mockCtl.Finish()
	mockDB := mockdbaccessor.NewMockDbAccessor(mockCtl)
	stubs := gostub.StubFunc(&common.GetDataBase, mockDB)
	defer stubs.Reset()
	mockDB.EXPECT().DeleteLeaf("/paasnet/tenants/admin/self").Return(nil)
	convey.Convey("Test DelAdminTenantInfoFromDB OK\n", t, func() {
		err := DelAdminTenantInfoFromDB()
		convey.So(err, convey.ShouldBeNil)
	})
}
