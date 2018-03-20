package context

import (
	"github.com/ZTE/Knitter/knitter-agent/domain/cni"
	"github.com/ZTE/Knitter/knitter-agent/domain/object/knitter-obj"
	"github.com/ZTE/Knitter/knitter-agent/domain/object/pod-obj"
	"github.com/ZTE/Knitter/knitter-agent/domain/object/port-obj"
	"github.com/ZTE/Knitter/pkg/iaas-accessor"
	"github.com/ZTE/Knitter/pkg/trans-dsl"
	"github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestIsVxLan(t *testing.T) {
	action := &IsVxLan{}

	cniParam := &cni.CniParam{PodNs: "nw001"}
	knitterObj := &knitterobj.KnitterObj{CniParam: cniParam}

	var portObjs []*portobj.PortObj
	portObjs = append(portObjs,
		&portobj.PortObj{LazyAttr: portobj.PortLazyAttr{NetAttr: portobj.NetworkAttrs{Provider: iaasaccessor.NetworkExtenAttrs{NetworkType: "vxlan"}}}},
	)
	podObj := &podobj.PodObj{PortObjs: portObjs}
	knitterInfo := &KnitterInfo{KnitterObj: knitterObj, podObj: podObj}
	transInfo := &transdsl.TransInfo{AppInfo: knitterInfo}

	convey.Convey("TestIsVxLan for succ!\n", t, func() {
		ok := action.Ok(transInfo)
		convey.So(ok, convey.ShouldEqual, true)
	})
}

func TestIsVxLan_false(t *testing.T) {
	action := &IsVxLan{}

	cniParam := &cni.CniParam{PodNs: "nw001"}
	knitterObj := &knitterobj.KnitterObj{CniParam: cniParam}

	var portObjs []*portobj.PortObj
	portObjs = append(portObjs,
		&portobj.PortObj{LazyAttr: portobj.PortLazyAttr{NetAttr: portobj.NetworkAttrs{Provider: iaasaccessor.NetworkExtenAttrs{NetworkType: ""}}}},
	)
	podObj := &podobj.PodObj{PortObjs: portObjs}
	knitterInfo := &KnitterInfo{KnitterObj: knitterObj, podObj: podObj}
	transInfo := &transdsl.TransInfo{AppInfo: knitterInfo}

	convey.Convey("TestIsVxLan for fail!\n", t, func() {
		ok := action.Ok(transInfo)
		convey.So(ok, convey.ShouldEqual, false)
	})
}
