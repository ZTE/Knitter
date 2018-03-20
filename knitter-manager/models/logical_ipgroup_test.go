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
	"github.com/ZTE/Knitter/knitter-manager/err-obj"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestIPGroupObjRepo(t *testing.T) {
	igObject1 := IPGroupObject{ID: "id1", NetworkID: "netid1", TenantID: "tid1"}

	Convey("TestIPGroupObjRepo:", t, func() {
		err := GetIPGroupObjRepoSingleton().Add(&igObject1)
		So(err, ShouldBeNil)

		obj1, err := GetIPGroupObjRepoSingleton().Get(igObject1.ID)
		So(err, ShouldBeNil)
		So(obj1, ShouldPointTo, &igObject1)

		obj1, err = GetIPGroupObjRepoSingleton().Get("invalid-id")
		So(err, ShouldEqual, errobj.ErrRecordNotExist)
		So(obj1, ShouldBeNil)

		tmpObjs, err := GetIPGroupObjRepoSingleton().ListByNetworkID("")
		So(err, ShouldBeNil)
		So(len(tmpObjs), ShouldEqual, 0)

		tmpObjs, err = GetIPGroupObjRepoSingleton().ListByTenantID("tid1")
		So(err, ShouldBeNil)
		So(len(tmpObjs), ShouldEqual, 1)

	})

}
