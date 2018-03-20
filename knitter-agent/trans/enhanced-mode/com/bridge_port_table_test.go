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

package com

import (
	"errors"
	"strings"
	"testing"

	"github.com/ZTE/Knitter/knitter-agent/domain/adapter"
	"github.com/ZTE/Knitter/knitter-agent/err-obj"

	"github.com/ZTE/Knitter/knitter-agent/infra/os-encap"
	"github.com/golang/gostub"
	. "github.com/smartystreets/goconvey/convey"
)

func TestBr0PortTableGet(t *testing.T) {
	portTle := GetBr0PortTableSingleton()
	Convey("TestBr0PortTableGetGet\n", t, func() {
		Convey("Get() test\n", func() {
			vethNameAPI := "veth123456"
			value, err := portTle.Get(vethNameAPI)
			So(err, ShouldEqual, errobj.ErrRecordNtExist)
			So(value, ShouldBeNil)

			outputs := []gostub.Output{{StubVals: gostub.Values{nil}, Times: 2}}
			stub := gostub.StubFuncSeq(&adapter.DataPersisterSaveToMemFile, outputs)
			portValueNetapi := Br0PortValue{
				TenantNetID: TenantNetworkID{
					TenantID:     "test-tenant-uuid",
					NetworkPlane: "std",
					NetworkID:    "net_api",
				},
				PortID: "test-port-uuid",
			}
			err = portTle.Insert(vethNameAPI, portValueNetapi)
			So(err, ShouldBeNil)

			value, err = portTle.Get(vethNameAPI)
			So(err, ShouldBeNil)
			So(value.(Br0PortValue), ShouldResemble, portValueNetapi)

			err = portTle.Delete(vethNameAPI)
			So(err, ShouldBeNil)
			stub.Reset()
		})
		Convey("Get() test failed\n", func() {
			errStr := "save to memory file failed"
			stub := gostub.StubFunc(&adapter.DataPersisterSaveToMemFile, errors.New(errStr))
			vethNameAPI := "veth123456"
			portValueNetapi := Br0PortValue{
				TenantNetID: TenantNetworkID{
					TenantID:     "test-tenant-uuid",
					NetworkPlane: "std",
					NetworkID:    "net_api",
				},
				PortID: "test-port-uuid",
			}
			err := portTle.Insert(vethNameAPI, portValueNetapi)
			So(strings.Contains(err.Error(), errStr), ShouldBeTrue)
			stub.Reset()
		})
	})
}

func TestBr0PortTableGetAll(t *testing.T) {
	portTbl := GetBr0PortTableSingleton()
	Convey("TestBr0PortTableGetAll\n", t, func() {
		Convey("GetAll() test-empty map\n", func() {
			portMaps, err := portTbl.GetAll()
			So(err, ShouldBeNil)
			So(len(portMaps), ShouldBeZeroValue)
		})
		Convey("GetAll() test\n", func() {
			vethNameAPI := "veth123456"
			portValueNetapi := Br0PortValue{
				TenantNetID: TenantNetworkID{
					TenantID:     "test-tenant-uuid",
					NetworkPlane: "std",
					NetworkID:    "net_api",
				},
				PortID: "test-port-uuid",
			}
			stubs := gostub.StubFunc(&osencap.OsStat, nil, nil)
			stubs.StubFunc(&adapter.DataPersisterSaveToMemFile, nil)

			err := portTbl.Insert(vethNameAPI, portValueNetapi)
			So(err, ShouldBeNil)

			tnMaps, err := portTbl.GetAll()
			So(err, ShouldBeNil)
			So(len(tnMaps), ShouldEqual, 1)
			So(tnMaps[vethNameAPI], ShouldResemble, portValueNetapi)
			err = portTbl.Delete(vethNameAPI)
			So(err, ShouldBeNil)
			stubs.Reset()
		})
	})
}

func TestBr0PortTableLoad(t *testing.T) {
	portTbl := GetBr0PortTableSingleton()
	Convey("TestBr0PortTableLoad\n", t, func() {
		Convey("test SUCC", func() {
			stubs := gostub.StubFunc(&osencap.OsStat, nil, nil)
			stubs.StubFunc(&adapter.DataPersisterLoadFromMemFile, nil)
			err := portTbl.Load()
			stubs.Reset()
			So(err, ShouldBeNil)
		})
		Convey("test FAIL- os.Stat error", func() {
			errStr := "os.Stat file failed"
			stubs := gostub.StubFunc(&osencap.OsStat, nil, errors.New(errStr))
			err := portTbl.Load()
			stubs.Reset()
			So(strings.Contains(err.Error(), errStr), ShouldBeTrue)
		})
		Convey("test FAIL- save to memory file error", func() {
			errStr := "dp.DataPersisterLoadFromMemFile file failed"
			stubs := gostub.StubFunc(&osencap.OsStat, nil, errors.New(errStr))
			stubs.StubFunc(&adapter.DataPersisterLoadFromMemFile, errors.New(errStr))
			err := portTbl.Load()
			stubs.Reset()
			So(strings.Contains(err.Error(), errStr), ShouldBeTrue)
		})
	})
}
