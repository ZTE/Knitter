package context

import (
	"github.com/ZTE/Knitter/knitter-agent/domain/cni"
	"github.com/ZTE/Knitter/knitter-agent/domain/object/knitter-obj"
	"github.com/ZTE/Knitter/knitter-agent/domain/object/physical-resource-obj"
	"github.com/ZTE/Knitter/pkg/trans-dsl"
	"github.com/bouk/monkey"
	"github.com/smartystreets/goconvey/convey"
	"reflect"
	"testing"
)

func TestCleanPhysicalResourceRecordAction(t *testing.T) {
	action := CleanPhysicalResourceRecordAction{}
	knitterInfo := &KnitterInfo{
		KnitterObj: &knitterobj.KnitterObj{
			CniParam: &cni.CniParam{
				ContainerID: "c1",
			},
		},
	}
	transInfo := &transdsl.TransInfo{AppInfo: knitterInfo}
	var drivers = []string{"veth"}
	var nouth *physicalresourceobj.NouthInterfaceObj
	guard := monkey.PatchInstanceMethod(reflect.TypeOf(nouth), "ReadDriversFromContainer",
		func(_ *physicalresourceobj.NouthInterfaceObj) ([]string, error) {
			return drivers, nil
		})
	defer guard.Unpatch()

	var inters = []string{"12345"}
	guard1 := monkey.PatchInstanceMethod(reflect.TypeOf(nouth), "ReadInterfacesFromDriver",
		func(_ *physicalresourceobj.NouthInterfaceObj) ([]string, error) {
			return inters, nil
		})
	defer guard1.Unpatch()

	guard2 := monkey.PatchInstanceMethod(reflect.TypeOf(nouth), "DeleteInterface",
		func(_ *physicalresourceobj.NouthInterfaceObj) error {
			return nil
		})
	defer guard2.Unpatch()

	guard4 := monkey.PatchInstanceMethod(reflect.TypeOf(nouth), "CleanDriver",
		func(_ *physicalresourceobj.NouthInterfaceObj) error {
			return nil
		})
	defer guard4.Unpatch()

	guard5 := monkey.PatchInstanceMethod(reflect.TypeOf(nouth), "CleanContainer",
		func(_ *physicalresourceobj.NouthInterfaceObj) error {
			return nil
		})
	defer guard5.Unpatch()

	convey.Convey("Test_CleanPhysicalResourceRecordAction\n", t, func() {
		err := action.Exec(transInfo)
		convey.So(err, convey.ShouldEqual, nil)
	})
}
