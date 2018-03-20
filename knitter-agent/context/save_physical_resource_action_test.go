package context

import (
	"errors"
	"reflect"
	"testing"

	"github.com/bouk/monkey"
	"github.com/smartystreets/goconvey/convey"

	"github.com/ZTE/Knitter/knitter-agent/domain/cni"
	"github.com/ZTE/Knitter/knitter-agent/domain/object/knitter-obj"
	"github.com/ZTE/Knitter/knitter-agent/domain/object/physical-resource-obj"
	"github.com/ZTE/Knitter/knitter-agent/domain/role/physical-resource-role"
	"github.com/ZTE/Knitter/pkg/inter-cmpt/mgr-agt"
	"github.com/ZTE/Knitter/pkg/trans-dsl"
)

func TestSavePhysicalResourceAction_Exec(t *testing.T) {
	knitterObj := &knitterobj.KnitterObj{
		CniParam: &cni.CniParam{
			TenantID:    "test-tenant-id",
			ContainerID: "test-container-id",
		},
	}
	knitterInfo := &KnitterInfo{
		KnitterObj: knitterObj,
	}
	repeatIdx := 0
	transInfo := &transdsl.TransInfo{
		RepeatIdx: repeatIdx,
		AppInfo:   knitterInfo,
	}

	southIf := &SouthInterface{
		port: &mgragt.CreatePortInfo{
			PortID: "test-eio-port-id",
		},
		vnicRole: &physicalresourcerole.VnicRole{
			NicType: "direct",
		},
	}
	knitterInfo.southIfs = map[int]*SouthInterface{transInfo.RepeatIdx: southIf}
	monkey.PatchInstanceMethod(reflect.TypeOf(&physicalresourcerole.VnicRole{}), "SaveResourceToLocalDB",
		func(_ *physicalresourcerole.VnicRole) error {
			return nil
		})
	monkey.PatchInstanceMethod(reflect.TypeOf(&physicalresourceobj.NouthInterfaceObj{}), "SaveInterface",
		func(_ *physicalresourceobj.NouthInterfaceObj) error {
			return nil
		})
	defer monkey.UnpatchAll()

	convey.Convey("TestSavePhysicalResourceAction_Exec", t, func() {
		err := (&SavePhysicalResourceAction{}).Exec(transInfo)
		convey.So(err, convey.ShouldBeNil)
	})
}

func TestSavePhysicalResourceAction_Exec_SaveSouthToLocalFail(t *testing.T) {
	knitterObj := &knitterobj.KnitterObj{
		CniParam: &cni.CniParam{
			TenantID:    "test-tenant-id",
			ContainerID: "test-container-id",
		},
	}
	knitterInfo := &KnitterInfo{
		KnitterObj: knitterObj,
	}
	repeatIdx := 0
	transInfo := &transdsl.TransInfo{
		RepeatIdx: repeatIdx,
		AppInfo:   knitterInfo,
	}

	southIf := &SouthInterface{
		port: &mgragt.CreatePortInfo{
			PortID: "test-eio-port-id",
		},
		vnicRole: &physicalresourcerole.VnicRole{
			NicType: "direct",
		},
	}
	knitterInfo.southIfs = map[int]*SouthInterface{transInfo.RepeatIdx: southIf}
	errStr := "level db not inited"
	monkey.PatchInstanceMethod(reflect.TypeOf(&physicalresourcerole.VnicRole{}), "SaveResourceToLocalDB",
		func(_ *physicalresourcerole.VnicRole) error {
			return errors.New(errStr)
		})
	defer monkey.UnpatchAll()

	convey.Convey("TestSavePhysicalResourceAction_Exec_SaveSouthToLocalFail", t, func() {
		err := (&SavePhysicalResourceAction{}).Exec(transInfo)
		convey.So(err.Error(), convey.ShouldEqual, errStr)
	})
}

func TestSavePhysicalResourceAction_ExecSaveNorthFail(t *testing.T) {
	knitterObj := &knitterobj.KnitterObj{
		CniParam: &cni.CniParam{
			TenantID:    "test-tenant-id",
			ContainerID: "test-container-id",
		},
	}
	knitterInfo := &KnitterInfo{
		KnitterObj: knitterObj,
	}
	repeatIdx := 0
	transInfo := &transdsl.TransInfo{
		RepeatIdx: repeatIdx,
		AppInfo:   knitterInfo,
	}

	southIf := &SouthInterface{
		port: &mgragt.CreatePortInfo{
			PortID: "test-eio-port-id",
		},
		vnicRole: &physicalresourcerole.VnicRole{
			NicType: "direct",
		},
	}
	knitterInfo.southIfs = map[int]*SouthInterface{transInfo.RepeatIdx: southIf}
	monkey.PatchInstanceMethod(reflect.TypeOf(&physicalresourcerole.VnicRole{}), "SaveResourceToLocalDB",
		func(_ *physicalresourcerole.VnicRole) error {
			return nil
		})
	errStr := "level db not inited"
	monkey.PatchInstanceMethod(reflect.TypeOf(&physicalresourceobj.NouthInterfaceObj{}), "SaveInterface",
		func(_ *physicalresourceobj.NouthInterfaceObj) error {
			return errors.New(errStr)
		})
	errStr2 := "disk busy"
	monkey.PatchInstanceMethod(reflect.TypeOf(&physicalresourcerole.VnicRole{}), "DeleteResourceFromLocalDB",
		func(_ *physicalresourcerole.VnicRole) error {
			return errors.New(errStr2)
		})
	defer monkey.UnpatchAll()

	convey.Convey("TestSavePhysicalResourceAction_ExecSaveNorthFail", t, func() {
		err := (&SavePhysicalResourceAction{}).Exec(transInfo)
		convey.So(err.Error(), convey.ShouldEqual, errStr)
	})
}

func TestSavePhysicalResourceAction_RollBack(t *testing.T) {
	knitterObj := &knitterobj.KnitterObj{
		CniParam: &cni.CniParam{
			TenantID:    "test-tenant-id",
			ContainerID: "test-container-id",
		},
	}
	knitterInfo := &KnitterInfo{
		KnitterObj: knitterObj,
	}
	repeatIdx := 0
	transInfo := &transdsl.TransInfo{
		RepeatIdx: repeatIdx,
		AppInfo:   knitterInfo,
	}

	southIf := &SouthInterface{
		port: &mgragt.CreatePortInfo{
			PortID: "test-eio-port-id",
		},
		vnicRole: &physicalresourcerole.VnicRole{
			NicType: "direct",
		},
	}
	knitterInfo.southIfs = map[int]*SouthInterface{transInfo.RepeatIdx: southIf}

	monkey.PatchInstanceMethod(reflect.TypeOf(&physicalresourceobj.NouthInterfaceObj{}), "DeleteInterface",
		func(_ *physicalresourceobj.NouthInterfaceObj) error {
			return nil
		})
	monkey.PatchInstanceMethod(reflect.TypeOf(&physicalresourcerole.VnicRole{}), "DeleteResourceFromLocalDB",
		func(_ *physicalresourcerole.VnicRole) error {
			return nil
		})
	defer monkey.UnpatchAll()

	convey.Convey("TestSavePhysicalResourceAction_RollBack", t, func() {
		(&SavePhysicalResourceAction{}).RollBack(transInfo)
	})
}

func TestSavePhysicalResourceAction_RollBack_AllFail(t *testing.T) {
	knitterObj := &knitterobj.KnitterObj{
		CniParam: &cni.CniParam{
			TenantID:    "test-tenant-id",
			ContainerID: "test-container-id",
		},
	}
	knitterInfo := &KnitterInfo{
		KnitterObj: knitterObj,
	}
	repeatIdx := 0
	transInfo := &transdsl.TransInfo{
		RepeatIdx: repeatIdx,
		AppInfo:   knitterInfo,
	}

	southIf := &SouthInterface{
		port: &mgragt.CreatePortInfo{
			PortID: "test-eio-port-id",
		},
		vnicRole: &physicalresourcerole.VnicRole{
			NicType: "direct",
		},
	}
	knitterInfo.southIfs = map[int]*SouthInterface{transInfo.RepeatIdx: southIf}

	errStr := "disk busy"
	monkey.PatchInstanceMethod(reflect.TypeOf(&physicalresourceobj.NouthInterfaceObj{}), "DeleteInterface",
		func(_ *physicalresourceobj.NouthInterfaceObj) error {
			return errors.New(errStr)
		})
	monkey.PatchInstanceMethod(reflect.TypeOf(&physicalresourcerole.VnicRole{}), "DeleteResourceFromLocalDB",
		func(_ *physicalresourcerole.VnicRole) error {
			return errors.New(errStr)
		})
	defer monkey.UnpatchAll()

	convey.Convey("TestSavePhysicalResourceAction_RollBack_AllFail", t, func() {
		(&SavePhysicalResourceAction{}).RollBack(transInfo)
	})
}
