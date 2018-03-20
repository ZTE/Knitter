package context

import (
	"github.com/ZTE/Knitter/knitter-agent/domain/object/knitter-obj"
	"github.com/ZTE/Knitter/pkg/trans-dsl"
	"github.com/containernetworking/cni/pkg/skel"
	"github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestIsNetNsExistFalse(t *testing.T) {
	action := &IsNetNsExist{}

	//cniParam := &cni.CniParam{PodNs: "nw001"}
	//knitterObj := &knitterobj.KnitterObj{CniParam: cniParam}
	//
	//var portObjs []*portobj.PortObj
	//portObjs = append(portObjs,
	//    &portobj.PortObj{EagerAttr: portobj.PortEagerAttr{NetworkName: "lan"}},
	//    &portobj.PortObj{EagerAttr: portobj.PortEagerAttr{NetworkName: "net_api"}},
	//    &portobj.PortObj{EagerAttr: portobj.PortEagerAttr{NetworkName: "control"}})
	//podObj := &podobj.PodObj{PortObjs: portObjs}
	//knitterInfo := &KnitterInfo{KnitterObj: knitterObj, podObj: podObj}
	//
	//cfgMock := gomock.NewController(t)
	//defer cfgMock.Finish()
	//mockDB := mockdbaccessor.NewMockDbAccessor(cfgMock)
	//var ports []*client.Node
	//node := &client.Node{Key:"etcd_key"}
	//ports = append(ports, node)
	//
	//mockDB.EXPECT().ReadDir(gomock.Any()).Return(ports,errobj.ErrAny)
	//
	//stubs := gostub.StubFunc(&cni.GetGlobalContext, &cni.AgentContext{DB:mockDB})
	//defer stubs.Reset()
	transInfo := &transdsl.TransInfo{
		AppInfo: &KnitterInfo{
			KnitterObj: &knitterobj.KnitterObj{
				Args: &skel.CmdArgs{
					Netns: "net",
				},
			},
		},
	}

	convey.Convey("Test--IsNetNsExist--false\n", t, func() {
		ok := action.Ok(transInfo)
		convey.So(ok, convey.ShouldEqual, false)
	})
}

//func TestIsNetNsExist_false(t *testing.T){
//    action := &IsNewPod{}
//
//    cniParam := &cni.CniParam{PodNs: "nw001"}
//    knitterObj := &knitterobj.KnitterObj{CniParam: cniParam}
//
//    var portObjs []*portobj.PortObj
//    portObjs = append(portObjs,
//        &portobj.PortObj{EagerAttr: portobj.PortEagerAttr{NetworkName: "lan"}},
//        &portobj.PortObj{EagerAttr: portobj.PortEagerAttr{NetworkName: "net_api"}},
//        &portobj.PortObj{EagerAttr: portobj.PortEagerAttr{NetworkName: "control"}})
//    podObj := &podobj.PodObj{PortObjs: portObjs}
//    knitterInfo := &KnitterInfo{KnitterObj: knitterObj, podObj: podObj}
//    transInfo := &transdsl.TransInfo{AppInfo: knitterInfo}
//
//    cfgMock := gomock.NewController(t)
//    defer cfgMock.Finish()
//    mockDB := mockdbaccessor.NewMockDbAccessor(cfgMock)
//    var ports []*client.Node
//    node := &client.Node{Key:"etcd_key"}
//    ports = append(ports, node)
//
//    mockDB.EXPECT().ReadDir(gomock.Any()).Return(ports,nil)
//
//    stubs := gostub.StubFunc(&cni.GetGlobalContext, &cni.AgentContext{DB:mockDB})
//    defer stubs.Reset()
//
//    action.RollBack(transInfo)
//    convey.Convey("TestIsNetNsExist for return false\n", t, func() {
//        ok := action.Ok(transInfo)
//        convey.So(ok, convey.ShouldEqual, false)
//    })
//}
