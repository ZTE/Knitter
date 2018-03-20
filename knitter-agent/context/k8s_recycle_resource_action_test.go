package context

import (
	"github.com/ZTE/Knitter/knitter-agent/domain/cni"
	"github.com/ZTE/Knitter/knitter-agent/domain/object/knitter-obj"
	"github.com/ZTE/Knitter/pkg/trans-dsl"
	"github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestK8SRecycleResourceAction(t *testing.T) {
	action := K8SRecycleResourceAction{}
	knitterInfo := &KnitterInfo{
		KnitterObj: &knitterobj.KnitterObj{
			CniParam: &cni.CniParam{
				ContainerID: "c1",
			},
		},
	}
	transInfo := &transdsl.TransInfo{AppInfo: knitterInfo}

	convey.Convey("Test_K8SRecycleResourceAction\n", t, func() {
		err := action.Exec(transInfo)
		convey.So(err, convey.ShouldEqual, nil)
	})
}
