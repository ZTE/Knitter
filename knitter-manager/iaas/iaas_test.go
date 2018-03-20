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

package iaas

import (
	"errors"
	"github.com/ZTE/Knitter/knitter-manager/public"
	"github.com/ZTE/Knitter/knitter-manager/tests/mock/db-mock"
	"github.com/ZTE/Knitter/pkg/uuid"
	"github.com/golang/gostub"
	"github.com/golang/mock/gomock"
	"github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestGetOriginalPortName(t *testing.T) {
	convey.Convey("TestGetOriginalPortName", t, func() {
		outputName := "eth0"
		inputName := uuid.NewUUID() + "_" + outputName
		resultName := GetOriginalPortName(inputName)
		convey.Println("INPUT:", inputName)
		convey.Println("OUTPUT:", resultName)
		convey.So(resultName, convey.ShouldNotEqual, outputName)
	})
}

func TestGetOriginalPortName1(t *testing.T) {
	convey.Convey("TestGetOriginalPortName", t, func() {
		outputName := "eth0"
		inputName := uuid.GetUUID8Byte(uuid.NewUUID()) +
			uuid.GetUUID8Byte(uuid.NewUUID()) +
			uuid.GetUUID8Byte(uuid.NewUUID()) + "_" + outputName
		resultName := GetOriginalPortName(inputName)
		convey.Println("INPUT:", inputName)
		convey.Println("OUTPUT:", resultName)
		convey.So(resultName, convey.ShouldEqual, outputName)
	})
}

func TestGetOriginalPortName2(t *testing.T) {
	convey.Convey("TestGetOriginalPortName", t, func() {
		outputName := "eth0"
		inputName := uuid.GetUUID8Byte(uuid.NewUUID()) +
			uuid.GetUUID8Byte(uuid.NewUUID()) +
			uuid.GetUUID8Byte(uuid.NewUUID()) +
			uuid.GetUUID8Byte(uuid.NewUUID()) + "_" + outputName
		resultName := GetOriginalPortName(inputName)
		convey.Println("INPUT:", inputName)
		convey.Println("OUTPUT:", resultName)
		convey.So(resultName, convey.ShouldEqual, outputName)
	})
}

func TestGetOriginalPortName3(t *testing.T) {
	convey.Convey("TestGetOriginalPortName", t, func() {
		outputName := "eth0"
		inputName := outputName
		resultName := GetOriginalPortName(inputName)
		convey.Println("INPUT:", inputName)
		convey.Println("OUTPUT:", resultName)
		convey.So(resultName, convey.ShouldEqual, outputName)
	})
}

func TestSaveDefaultPhysnetOK(t *testing.T) {
	var defaultPhysnet string = "physnet1"
	mockCtl := gomock.NewController(t)
	defer mockCtl.Finish()
	mockDB := mockdbaccessor.NewMockDbAccessor(mockCtl)
	stubs := gostub.StubFunc(&common.GetDataBase, mockDB)
	defer stubs.Reset()

	gomock.InOrder(
		mockDB.EXPECT().SaveLeaf("/paasnet/defaultphysnet", defaultPhysnet).Return(nil),
	)

	convey.Convey("Test_SaveDefaultPhysnet_OK\n", t, func() {
		errS := SaveDefaultPhysnet(defaultPhysnet)
		convey.So(errS, convey.ShouldEqual, nil)
	})
}

func TestSaveDefaultPhysnetErr(t *testing.T) {
	var defaultPhysnet string = "physnet1"
	mockCtl := gomock.NewController(t)
	defer mockCtl.Finish()
	mockDB := mockdbaccessor.NewMockDbAccessor(mockCtl)
	stubs := gostub.StubFunc(&common.GetDataBase, mockDB)
	defer stubs.Reset()

	gomock.InOrder(
		mockDB.EXPECT().SaveLeaf("/paasnet/defaultphysnet", defaultPhysnet).Return(errors.New("save leaf err")),
	)

	convey.Convey("Test_SaveDefaultPhysnet_OK\n", t, func() {
		errS := SaveDefaultPhysnet(defaultPhysnet)
		convey.So(errS, convey.ShouldNotEqual, nil)
	})
}
