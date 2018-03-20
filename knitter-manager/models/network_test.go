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
	"testing"

	"encoding/json"
	"github.com/antonholmquist/jason"
	"github.com/rackspace/gophercloud/openstack/networking/v2/subnets"
	. "github.com/smartystreets/goconvey/convey"
)

type Network4Test struct {
	Cidr            string                   `mapstructure:"cidr" json:"cidr"`
	CreateTime      string                   `mapstructure:"create_time" json:"create_time"`
	Description     string                   `mapstructure:"description" json:"description"`
	Gateway         string                   `mapstructure:"gateway" json:"gateway"`
	Name            string                   `mapstructure:"name" json:"name"`
	Public          bool                     `mapstructure:"public" json:"public"`
	NetworkID       string                   `mapstructure:"network_id" json:"network_id"`
	State           string                   `mapstructure:"state" json:"state"`
	AllocationPools []subnets.AllocationPool `mapstructure:"allocation_pools" json:"allocation_pools"`
}

func TestGetCidrIpRangeOK(t *testing.T) {
	Convey("TestGetCidrIpRangeOK1\n", t, func() {
		cidr24 := "129.128.127.0/24"
		minIP, maxIP := GetCidrIPRange(cidr24)
		So(minIP, ShouldEqual, "129.128.127.1")
		So(maxIP, ShouldEqual, "129.128.127.254")
	})
	Convey("TestGetCidrIpRangeOK2\n", t, func() {
		cidr26 := "129.128.127.0/26"
		minIP, maxIP := GetCidrIPRange(cidr26)
		So(minIP, ShouldEqual, "129.128.127.1")
		So(maxIP, ShouldEqual, "129.128.127.62")
	})
}

func TestGetCidrIpRangeErr(t *testing.T) {
	Convey("TestGetCidrIpRangeErr\n", t, func() {
		cidr24 := ""
		minIP, maxIP := GetCidrIPRange(cidr24)
		So(minIP, ShouldEqual, "")
		So(maxIP, ShouldEqual, "")
	})
}

func TestIsFixIpInCidrOK(t *testing.T) {
	Convey("TestIsFixIpInCidrOK\n", t, func() {
		ip := "129.128.127.44"
		cidr := "129.128.127.0/26"
		bool := IsFixIPInCidr(ip, cidr)
		So(bool, ShouldEqual, true)
	})
}

func TestIsFixIpInCidrErr(t *testing.T) {
	Convey("TestIsFixIpInCidrErr\n", t, func() {
		ip := "129.128.127.100"
		cidr := "129.128.127.0/26"
		bool := IsFixIPInCidr(ip, cidr)
		So(bool, ShouldEqual, false)
	})
}

func TestIsAllocationPoolsInCidrOK(t *testing.T) {
	allocationPools := []subnets.AllocationPool{
		{
			Start: "10.92.124.40",
			End:   "10.92.124.100",
		},
		{
			Start: "10.92.124.101",
			End:   "10.92.124.200",
		},
	}
	Convey("TestIsAllocationPoolsInCidrOK\n", t, func() {
		cidr := "129.128.124.0/24"
		bool := IsAllocationPoolsInCidr(allocationPools, cidr)
		So(bool, ShouldEqual, false)
	})
}

func TestIsAllocationPoolsInCidrErr(t *testing.T) {
	Convey("TestIsAllocationPoolsInCidrErr1\n", t, func() {
		allocationPools := []subnets.AllocationPool{
			{
				Start: "10.92.124.30",
				End:   "10.92.124.45",
			},
		}
		cidr := "129.128.124.40/28"
		bool := IsAllocationPoolsInCidr(allocationPools, cidr)
		So(bool, ShouldEqual, false)
	})
	Convey("TestIsAllocationPoolsInCidrErr2\n", t, func() {
		allocationPools := []subnets.AllocationPool{
			{
				Start: "10.92.124.35",
				End:   "10.92.124.50",
			},
		}
		cidr := "129.128.124.40/28"
		bool := IsAllocationPoolsInCidr(allocationPools, cidr)
		So(bool, ShouldEqual, false)
	})
}

func TestIsFixIPInIpRangeOk(t *testing.T) {
	Convey("TestIsFixIpInIpRangeOk-fixip-equal-pool.start\n", t, func() {
		allocationPool := subnets.AllocationPool{
			Start: "10.92.124.35",
			End:   "10.92.124.50",
		}
		fixIP := "10.92.124.35"
		bool := IsFixIPInIPRange(fixIP, allocationPool)
		So(bool, ShouldEqual, true)
	})
	Convey("TestIsFixIpInIpRangeOk-fixip-equal-pool.end\n", t, func() {
		allocationPool := subnets.AllocationPool{
			Start: "10.92.124.35",
			End:   "10.92.124.50",
		}
		fixIP := "10.92.124.50"
		bool := IsFixIPInIPRange(fixIP, allocationPool)
		So(bool, ShouldEqual, true)
	})
	Convey("TestIsFixIpInIpRangeOk-fixip-in-pool\n", t, func() {
		allocationPool := subnets.AllocationPool{
			Start: "10.92.124.35",
			End:   "10.92.124.50",
		}
		fixIP := "10.92.124.45"
		bool := IsFixIPInIPRange(fixIP, allocationPool)
		So(bool, ShouldEqual, true)
	})
}

func TestIsFixIpInIPRangeErr(t *testing.T) {
	Convey("TestIsFixIpInIpRangeErr-input-param-fixip-error\n", t, func() {
		allocationPool := subnets.AllocationPool{
			Start: "10.92.124.35",
			End:   "10.92.124.50",
		}
		fixIP := ""
		bool := IsFixIPInIPRange(fixIP, allocationPool)
		So(bool, ShouldEqual, false)
	})
	Convey("TestIsFixIpInIpRangeErr-input-param-pool.stat-error\n", t, func() {
		allocationPool := subnets.AllocationPool{
			Start: "",
			End:   "10.92.124.50",
		}
		fixIP := "129.128.124.36"
		bool := IsFixIPInIPRange(fixIP, allocationPool)
		So(bool, ShouldEqual, false)
	})
	Convey("TestIsFixIpInIpRangeErr-input-param-pool.end-error\n", t, func() {
		allocationPool := subnets.AllocationPool{
			Start: "10.92.124.35",
			End:   "",
		}
		fixIP := "129.128.124.50"
		bool := IsFixIPInIPRange(fixIP, allocationPool)
		So(bool, ShouldEqual, false)
	})
	Convey("TestIsFixIpInIpRangeErr-fixip<pool.start\n", t, func() {
		allocationPool := subnets.AllocationPool{
			Start: "10.92.124.35",
			End:   "10.92.124.50",
		}
		fixIP := "129.128.124.34"
		bool := IsFixIPInIPRange(fixIP, allocationPool)
		So(bool, ShouldEqual, false)
	})
	Convey("TestIsFixIpInIpRangeErr-fixip>pool.end\n", t, func() {
		allocationPool := subnets.AllocationPool{
			Start: "10.92.124.35",
			End:   "10.92.124.50",
		}
		fixIP := "129.128.124.51"
		bool := IsFixIPInIPRange(fixIP, allocationPool)
		So(bool, ShouldEqual, false)
	})
}

func TestIsAllocationPoolsCoverdOk(t *testing.T) {
	Convey("TestIsAllocationPoolsCoverdOk\n", t, func() {
		allocationPools := []subnets.AllocationPool{
			{
				Start: "10.92.124.40",
				End:   "10.92.124.100",
			},
			{
				Start: "10.92.124.101",
				End:   "10.92.124.200",
			},
		}
		bool := IsAllocationPoolsCoverd(allocationPools)
		So(bool, ShouldEqual, false)
	})
}

func TestIsAllocationPoolsCoverdErr(t *testing.T) {
	Convey("TestIsAllocationPoolsCoverdErr\n", t, func() {
		allocationPools := []subnets.AllocationPool{
			{
				Start: "10.92.124.40",
				End:   "10.92.124.100",
			},
			{
				Start: "10.92.124.100",
				End:   "10.92.124.200",
			},
		}
		bool := IsAllocationPoolsCoverd(allocationPools)
		So(bool, ShouldEqual, true)
	})
	Convey("TestIsAllocationPoolsCoverdErr2\n", t, func() {
		allocationPools := []subnets.AllocationPool{
			{
				Start: "10.92.124.110",
				End:   "10.92.124.120",
			},
			{
				Start: "10.92.124.100",
				End:   "10.92.124.200",
			},
		}
		bool := IsAllocationPoolsCoverd(allocationPools)
		So(bool, ShouldEqual, true)
	})
	Convey("TestIsAllocationPoolsCoverdErr2\n", t, func() {
		allocationPools := []subnets.AllocationPool{}
		bool := IsAllocationPoolsCoverd(allocationPools)
		So(bool, ShouldEqual, true)
	})
}

func TestCheckAllocationPoolsOK(t *testing.T) {
	Convey("TestCheckAllocationPoolsOK\n", t, func() {
		allocationPools := []subnets.AllocationPool{
			{
				Start: "10.92.124.10",
				End:   "10.92.124.20",
			},
			{
				Start: "10.92.124.100",
				End:   "10.92.124.200",
			},
		}
		cidr := "10.92.124.0/24"
		error := CheckAllocationPools(allocationPools, cidr)
		So(error, ShouldEqual, nil)
	})
}

func TestCheckAllocationPoolsErr(t *testing.T) {
	Convey("TestCheckAllocationPoolsErr\n", t, func() {
		allocationPools := []subnets.AllocationPool{}
		cidr := "10.92.124.0/24"
		error := CheckAllocationPools(allocationPools, cidr)
		So(error, ShouldNotBeNil)
	})
	Convey("TestCheckAllocationPoolsErr1\n", t, func() {
		allocationPools := []subnets.AllocationPool{
			{
				Start: "10.92.123.10",
				End:   "10.92.123.20",
			},
			{
				Start: "10.92.124.100",
				End:   "10.92.124.200",
			},
		}
		cidr := ""
		error := CheckAllocationPools(allocationPools, cidr)
		So(error, ShouldNotBeNil)
	})
	Convey("TestCheckAllocationPoolsErr2\n", t, func() {
		allocationPools := []subnets.AllocationPool{
			{
				Start: "10.92.123.10",
				End:   "10.92.123.20",
			},
			{
				Start: "10.92.124.100",
				End:   "10.92.124.200",
			},
		}
		cidr := "10.92.124.0/24"
		error := CheckAllocationPools(allocationPools, cidr)
		So(error, ShouldNotBeNil)
	})
	Convey("TestCheckAllocationPoolsErr2\n", t, func() {
		allocationPools := []subnets.AllocationPool{
			{
				Start: "10.92.124.110",
				End:   "10.92.124.120",
			},
			{
				Start: "10.92.124.100",
				End:   "10.92.124.200",
			},
		}
		cidr := "10.92.124.0/24"
		error := CheckAllocationPools(allocationPools, cidr)
		So(error, ShouldNotBeNil)
	})
}

func TestIsAllocationPoolsLegalOK(t *testing.T) {
	Convey("TestIsAllocationPoolsLegalOK\n", t, func() {
		allocationPools := []subnets.AllocationPool{
			{
				Start: "10.92.124.10",
				End:   "10.92.124.20",
			},
			{
				Start: "10.92.124.100",
				End:   "10.92.124.200",
			},
		}
		cidr := "10.92.124.0/24"
		bool := IsAllocationPoolsLegal(allocationPools, cidr)
		So(bool, ShouldEqual, true)
	})
}

func TestIsAllocationPoolsLegalErr(t *testing.T) {
	Convey("TestIsAllocationPoolsLegalErr1\n", t, func() {
		allocationPools := []subnets.AllocationPool{}
		cidr := "10.92.124.0/24"
		bool := IsAllocationPoolsLegal(allocationPools, cidr)
		So(bool, ShouldEqual, false)
	})
	Convey("TestIsAllocationPoolsLegalErr2\n", t, func() {
		allocationPools := []subnets.AllocationPool{
			{
				Start: "10.92.124.10",
				End:   "10.92.124.20",
			},
			{
				Start: "10.92.124.100",
				End:   "10.92.124.200",
			},
		}
		cidr := ""
		bool := IsAllocationPoolsLegal(allocationPools, cidr)
		So(bool, ShouldEqual, false)
	})
	Convey("TestIsAllocationPoolsLegalErr3\n", t, func() {
		allocationPools := []subnets.AllocationPool{
			{
				Start: "10.92.124.40",
				End:   "10.92.124.20",
			},
			{
				Start: "10.92.124.100",
				End:   "10.92.124.200",
			},
		}
		cidr := "10.92.124.0/24"
		bool := IsAllocationPoolsLegal(allocationPools, cidr)
		So(bool, ShouldEqual, false)
	})
	Convey("TestIsAllocationPoolsLegalErr4\n", t, func() {
		allocationPools := []subnets.AllocationPool{
			{
				Start: "10.92.124.110",
				End:   "10.92.124.120",
			},
			{
				Start: "10.92.124.100",
				End:   "10.92.124.200",
			},
		}
		cidr := "10.92.124.0/24"
		bool := IsAllocationPoolsLegal(allocationPools, cidr)
		So(bool, ShouldEqual, false)
	})
}

func TestGetAllocationPoolsOk(t *testing.T) {
	networkInfo := Network4Test{
		Cidr:    "10.92.123.0/24",
		Gateway: "10.92.123.1",
		AllocationPools: []subnets.AllocationPool{
			{
				Start: "10.92.123.10",
				End:   "10.92.123.100",
			},
		},
	}
	body, _ := json.Marshal(networkInfo)
	networkObject, _ := jason.NewObjectFromBytes(body)
	allocationPools, _ := networkObject.GetObjectArray("allocation_pools")
	Convey("TestGetAllocationPoolsOk\n", t, func() {
		pools, err := GetAllocationPools(allocationPools, networkInfo.Cidr, networkInfo.Gateway)
		So(err, ShouldEqual, nil)
		So(pools[0].Start, ShouldEqual, "10.92.123.10")
		So(pools[0].End, ShouldEqual, "10.92.123.100")
	})
}

func TestGetAllocationPoolsERR(t *testing.T) {

	Convey("TestGetAllocationPoolsERR-allocation-pools-null\n", t, func() {
		networkInfo := Network4Test{
			Cidr:            "10.92.123.0/24",
			Gateway:         "10.92.123.1",
			AllocationPools: []subnets.AllocationPool{},
		}
		body, _ := json.Marshal(networkInfo)
		networkObject, _ := jason.NewObjectFromBytes(body)
		allocationPools, _ := networkObject.GetObjectArray("allocation_pools")
		pools, err := GetAllocationPools(allocationPools, networkInfo.Cidr, networkInfo.Gateway)
		So(err, ShouldEqual, nil)
		So(len(pools), ShouldEqual, 0)
	})
	Convey("TestGetAllocationPoolsERR-cidr-illegal\n", t, func() {
		networkInfo := Network4Test{
			Cidr:    "10.92.123.0",
			Gateway: "10.92.123.1",
			AllocationPools: []subnets.AllocationPool{
				{
					Start: "10.92.123.10",
					End:   "10.92.123.100",
				},
			},
		}
		body, _ := json.Marshal(networkInfo)
		networkObject, _ := jason.NewObjectFromBytes(body)
		allocationPools, _ := networkObject.GetObjectArray("allocation_pools")
		pools, err := GetAllocationPools(allocationPools, networkInfo.Cidr, networkInfo.Gateway)
		So(err, ShouldNotBeNil)
		So(len(pools), ShouldEqual, 0)
	})
	Convey("TestGetAllocationPoolsERR-gw-in-ip-pools\n", t, func() {
		networkInfo := Network4Test{
			Cidr:    "10.92.123.0/24",
			Gateway: "10.92.123.10",
			AllocationPools: []subnets.AllocationPool{
				{
					Start: "10.92.123.10",
					End:   "10.92.123.100",
				},
			},
		}
		body, _ := json.Marshal(networkInfo)
		networkObject, _ := jason.NewObjectFromBytes(body)
		allocationPools, _ := networkObject.GetObjectArray("allocation_pools")
		pools, err := GetAllocationPools(allocationPools, networkInfo.Cidr, networkInfo.Gateway)
		So(err, ShouldNotBeNil)
		So(len(pools), ShouldEqual, 0)
	})
	Convey("TestGetAllocationPoolsERR-ip-pools-illegal\n", t, func() {
		networkInfo := Network4Test{
			Cidr:    "10.92.123.0/24",
			Gateway: "10.92.123.1",
			AllocationPools: []subnets.AllocationPool{
				{
					Start: "10.92.123.10",
					End:   "10.92.123.100",
				},
				{
					Start: "10.92.123.100",
					End:   "10.92.123.200",
				},
			},
		}
		body, _ := json.Marshal(networkInfo)
		networkObject, _ := jason.NewObjectFromBytes(body)
		allocationPools, _ := networkObject.GetObjectArray("allocation_pools")
		pools, err := GetAllocationPools(allocationPools, networkInfo.Cidr, networkInfo.Gateway)
		So(err, ShouldNotBeNil)
		So(len(pools), ShouldEqual, 0)
	})
}

func TestGetAllocationPoolsByCidrOK(t *testing.T) {
	Convey("TestGetAllocationPoolsByCidrOK\n", t, func() {
		cidr := "12.12.12.0/24"
		pools := GetAllocationPoolsByCidr(cidr)
		So(pools[0].Start, ShouldEqual, "12.12.12.1")
		So(pools[0].End, ShouldEqual, "12.12.12.254")
	})
}

func TestGetAllocationPoolsByCidrERR(t *testing.T) {
	Convey("TestGetAllocationPoolsByCidrERR\n", t, func() {
		cidr := "12.12.12.0"
		pools := GetAllocationPoolsByCidr(cidr)
		So(len(pools), ShouldEqual, 0)
	})
}

func TestIsCidrLegalOK(t *testing.T) {
	Convey("TestIsCidrLegalOK\n", t, func() {
		cidr := "12.12.12.0/24"
		bl := IsCidrLegal(cidr)
		So(bl, ShouldEqual, true)
	})
}

func TestIsCidrLegalERR(t *testing.T) {
	Convey("TestIsCidrLegalERR1\n", t, func() {
		cidr := "12.12.12.0"
		bl := IsCidrLegal(cidr)
		So(bl, ShouldEqual, false)
	})
	Convey("TestIsCidrLegalERR2\n", t, func() {
		cidr := ""
		bl := IsCidrLegal(cidr)
		So(bl, ShouldEqual, false)
	})
	Convey("TestIsCidrLegalERR3\n", t, func() {
		cidr := "12.1a.12.0/24"
		bl := IsCidrLegal(cidr)
		So(bl, ShouldEqual, false)
	})
}
