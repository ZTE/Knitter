package services

import (
	"encoding/json"
	"errors"
	"github.com/ZTE/Knitter/knitter-monitor/infra"
	"github.com/antonholmquist/jason"
	"github.com/bouk/monkey"
	"github.com/smartystreets/goconvey/convey"
	"reflect"
	"testing"
)

func TestPortService_NewPortsWithEagerAttrAndLazyAttr(t *testing.T) {

	networks := "{\"ports\": " +
		"[{\"attach_to_network\": \"net_api\", \"attributes\": {\"accelerate\": \"false\", \"function\": \"std\", \"nic_name\": \"eth0\", \"nic_type\": \"normal\"}}," +
		" {\"attach_to_network\": \"control\", \"attributes\": {\"accelerate\": \"false\", \"function\": \"control\", \"nic_name\": \"nwrfcontrol\", \"nic_type\": \"normal\"}}, " +
		"{\"attach_to_network\": \"media\", \"attributes\": {\"accelerate\": \"false\", \"function\": \"media\", \"nic_name\": \"nwrfmedia\", \"nic_type\": \"normal\"}}]}"
	nwJSON, _ := jason.NewObjectFromBytes([]byte(networks))

	mportsExpect := []*Port{{}}
	var ps *portService
	monkey.PatchInstanceMethod(reflect.TypeOf(ps), "NewPortsWithEagerAttrFromK8s",
		func(_ *portService, _ *PodForCreatPort, _ *jason.Object) ([]*Port,
			error) {
			return mportsExpect, nil
		})
	defer monkey.UnpatchAll()
	monkey.PatchInstanceMethod(reflect.TypeOf(ps), "CreateBulkPorts", func(_ *portService, pod *PodForCreatPort, ports []*Port) (resp *infra.CreatePortsResp, err error) {
		return &infra.CreatePortsResp{}, nil
	})
	monkey.Patch(fillPortLazyAttr, func(resp *infra.CreatePortsResp, ports []*Port, tenantID string) error {
		return nil
	})
	pod := &PodForCreatPort{}
	convey.Convey("TestPortService_NewPortsWithEagerAttrAndLazyAttr", t, func() {
		mports, err := GetPortService().NewPortsWithEagerAttrAndLazyAttr(pod, nwJSON)
		convey.So(mports, convey.ShouldResemble, mportsExpect)
		convey.So(err, convey.ShouldBeNil)
	})
}

func TestPortService_NewPortsWithEagerAttrAndLazyAttrFail(t *testing.T) {

	networks := "{\"ports\": " +
		"[{\"attach_to_network\": \"net_api\", \"attributes\": {\"accelerate\": \"false\", \"function\": \"std\", \"nic_name\": \"eth0\", \"nic_type\": \"normal\"}}," +
		" {\"attach_to_network\": \"control\", \"attributes\": {\"accelerate\": \"false\", \"function\": \"control\", \"nic_name\": \"nwrfcontrol\", \"nic_type\": \"normal\"}}, " +
		"{\"attach_to_network\": \"media\", \"attributes\": {\"accelerate\": \"false\", \"function\": \"media\", \"nic_name\": \"nwrfmedia\", \"nic_type\": \"normal\"}}]}"
	nwJSON, _ := jason.NewObjectFromBytes([]byte(networks))
	var ps *portService
	monkey.PatchInstanceMethod(reflect.TypeOf(ps), "NewPortsWithEagerAttrFromK8s",
		func(_ *portService, _ *PodForCreatPort, _ *jason.Object) ([]*Port,
			error) {
			return nil, errors.New("new eager attr err")
		})
	defer monkey.UnpatchAll()

	pod := &PodForCreatPort{}
	convey.Convey("TestPortService_NewPortsWithEagerAttrAndLazyAttr", t, func() {
		mports, err := GetPortService().NewPortsWithEagerAttrAndLazyAttr(pod, nwJSON)
		convey.So(mports, convey.ShouldBeNil)
		convey.So(err.Error(), convey.ShouldEqual, "new eager attr err")
	})
}

func TestPortService_NewPortsWithEagerAttrAndLazyAttrFail2(t *testing.T) {

	networks := "{\"ports\": " +
		"[{\"attach_to_network\": \"net_api\", \"attributes\": {\"accelerate\": \"false\", \"function\": \"std\", \"nic_name\": \"eth0\", \"nic_type\": \"normal\"}}," +
		" {\"attach_to_network\": \"control\", \"attributes\": {\"accelerate\": \"false\", \"function\": \"control\", \"nic_name\": \"nwrfcontrol\", \"nic_type\": \"normal\"}}, " +
		"{\"attach_to_network\": \"media\", \"attributes\": {\"accelerate\": \"false\", \"function\": \"media\", \"nic_name\": \"nwrfmedia\", \"nic_type\": \"normal\"}}]}"
	nwJSON, _ := jason.NewObjectFromBytes([]byte(networks))
	mportsExpect := []*Port{{}}
	var ps *portService
	monkey.PatchInstanceMethod(reflect.TypeOf(ps), "NewPortsWithEagerAttrFromK8s",
		func(_ *portService, _ *PodForCreatPort, _ *jason.Object) ([]*Port,
			error) {
			return mportsExpect, nil
		})
	defer monkey.UnpatchAll()
	monkey.PatchInstanceMethod(reflect.TypeOf(ps), "CreateBulkPorts", func(_ *portService, pod *PodForCreatPort, ports []*Port) (resp *infra.CreatePortsResp, err error) {
		return nil, errors.New("CreateBulkPorts err")
	})

	pod := &PodForCreatPort{}
	convey.Convey("TestPortService_NewPortsWithEagerAttrAndLazyAttr", t, func() {
		mports, err := GetPortService().NewPortsWithEagerAttrAndLazyAttr(pod, nwJSON)
		convey.So(err.Error(), convey.ShouldEqual, "CreateBulkPorts err")
		convey.So(mports, convey.ShouldBeNil)
	})
}

func TestPortService_NewPortsWithEagerAttrAndLazyAttrFail3(t *testing.T) {

	networks := "{\"ports\": " +
		"[{\"attach_to_network\": \"net_api\", \"attributes\": {\"accelerate\": \"false\", \"function\": \"std\", \"nic_name\": \"eth0\", \"nic_type\": \"normal\"}}," +
		" {\"attach_to_network\": \"control\", \"attributes\": {\"accelerate\": \"false\", \"function\": \"control\", \"nic_name\": \"nwrfcontrol\", \"nic_type\": \"normal\"}}, " +
		"{\"attach_to_network\": \"media\", \"attributes\": {\"accelerate\": \"false\", \"function\": \"media\", \"nic_name\": \"nwrfmedia\", \"nic_type\": \"normal\"}}]}"
	nwJSON, _ := jason.NewObjectFromBytes([]byte(networks))
	mportsExpect := []*Port{{}}
	var ps *portService
	monkey.PatchInstanceMethod(reflect.TypeOf(ps), "NewPortsWithEagerAttrFromK8s",
		func(_ *portService, _ *PodForCreatPort, _ *jason.Object) ([]*Port,
			error) {
			return mportsExpect, nil
		})
	defer monkey.UnpatchAll()
	monkey.PatchInstanceMethod(reflect.TypeOf(ps), "CreateBulkPorts", func(_ *portService, pod *PodForCreatPort, ports []*Port) (resp *infra.CreatePortsResp, err error) {
		return &infra.CreatePortsResp{}, nil
	})
	monkey.Patch(fillPortLazyAttr, func(resp *infra.CreatePortsResp, ports []*Port, tenantID string) error {
		return errors.New("fill err")
	})
	pod := &PodForCreatPort{}
	convey.Convey("TestPortService_NewPortsWithEagerAttrAndLazyAttr", t, func() {
		mports, err := GetPortService().NewPortsWithEagerAttrAndLazyAttr(pod, nwJSON)
		convey.So(err.Error(), convey.ShouldEqual, "fill err")
		convey.So(mports, convey.ShouldBeNil)
	})
}

func TestPortService_NewPortsWithEagerAttrFromK8s(t *testing.T) {
	networks := "{\"ports\": " +
		"[{\"attach_to_network\": \"net_api\", \"attributes\": {\"accelerate\": \"false\", \"function\": \"std\", \"nic_name\": \"eth0\", \"nic_type\": \"normal\"}}," +
		" {\"attach_to_network\": \"control\", \"attributes\": {\"accelerate\": \"false\", \"function\": \"control\", \"nic_name\": \"nwrfcontrol\", \"nic_type\": \"normal\"}}, " +
		"{\"attach_to_network\": \"media\", \"attributes\": {\"accelerate\": \"false\", \"function\": \"media\", \"nic_name\": \"nwrfmedia\", \"nic_type\": \"normal\"}}]}"
	nwJSON, _ := jason.NewObjectFromBytes([]byte(networks))

	PodName := "test1"
	podNS := "admin"

	podForCreatePort := &PodForCreatPort{PodNs: podNS, PodName: PodName}

	portEagerAttrApi := PortEagerAttr{
		NetworkName:  "net_api",
		NetworkPlane: "std",
		PortName:     "eth0",
		VnicType:     "normal",
		Accelerate:   "false",
		PodName:      PodName,
		PodNs:        podNS,
		FixIP:        "",
		IPGroupName:  "",
		Metadata:     make(map[string]string),
		Combinable:   "false",
		Roles:        []string{"std"},
	}

	portEagerAttrControl := PortEagerAttr{
		NetworkName:  "control",
		NetworkPlane: "control",
		PortName:     "nwrfcontrol",
		VnicType:     "normal",
		Accelerate:   "false",
		PodName:      PodName,
		PodNs:        podNS,
		FixIP:        "",
		IPGroupName:  "",
		Metadata:     make(map[string]string),
		Combinable:   "false",
		Roles:        []string{"control"},
	}

	portEagerAttrMedia := PortEagerAttr{
		NetworkName:  "media",
		NetworkPlane: "media",
		PortName:     "nwrfmedia",
		VnicType:     "normal",
		Accelerate:   "false",
		PodName:      PodName,
		PodNs:        podNS,
		FixIP:        "",
		IPGroupName:  "",
		Metadata:     make(map[string]string),
		Combinable:   "false",
		Roles:        []string{"media"},
	}
	portsForService := []*Port{
		{
			EagerAttr: portEagerAttrApi,
		},
		{
			EagerAttr: portEagerAttrControl,
		},
		{
			EagerAttr: portEagerAttrMedia,
		},
	}
	convey.Convey("TestPortService_NewPortsWithEagerAttrFromK8s", t, func() {
		ports, err := GetPortService().NewPortsWithEagerAttrFromK8s(podForCreatePort, nwJSON)
		convey.So(ports, convey.ShouldResemble, portsForService)
		convey.So(err, convey.ShouldBeNil)
	})

}

func TestDestoryBulkPorts(t *testing.T) {
	var mc *infra.ManagerClient
	monkey.PatchInstanceMethod(reflect.TypeOf(mc), "DeleteNeutronPort",
		func(mc *infra.ManagerClient, tenantID string, portID string) (e error) {
			return errors.New("delete err")
		})
	convey.Convey("TestDestoryBulkPorts", t, func() {
		mc = infra.GetManagerClient()
		teantIDs := []string{"admin", "network"}
		portIDs := []string{"1", "2"}
		destoryBulkPorts(mc, teantIDs, portIDs)
	})
}

func TestPortService_DeleteBulkPorts(t *testing.T) {
	var mc *infra.ManagerClient
	monkey.PatchInstanceMethod(reflect.TypeOf(mc), "DeleteNeutronPort",
		func(mc *infra.ManagerClient, tenantID string, portID string) (e error) {
			return nil
		})
	convey.Convey("TestPortService_DeleteBulkPorts", t, func() {
		portIDs := []string{"1", "2"}
		err := GetPortService().DeleteBulkPorts("admin", portIDs)
		convey.So(err, convey.ShouldBeNil)
	})
}

func TestPortService_DeleteBulkPortsFail(t *testing.T) {
	var mc *infra.ManagerClient
	monkey.PatchInstanceMethod(reflect.TypeOf(mc), "DeleteNeutronPort",
		func(mc *infra.ManagerClient, tenantID string, portID string) (e error) {
			return errors.New("delete err")
		})
	convey.Convey("TestPortService_DeleteBulkPortsFail", t, func() {
		portIDs := []string{"1", "2"}
		err := GetPortService().DeleteBulkPorts("admin", portIDs)
		convey.So(err.Error(), convey.ShouldEqual, "delete err")
	})
}

func TestPortService_CreateBulkPorts(t *testing.T) {
	PodName := "test1"
	podNS := "admin"
	podForCreatePort := &PodForCreatPort{PodNs: podNS, PodName: PodName}

	portEagerAttrApi := PortEagerAttr{
		NetworkName:  "net_api",
		NetworkPlane: "std",
		PortName:     "eth0",
		VnicType:     "normal",
		Accelerate:   "false",
		PodName:      PodName,
		PodNs:        podNS,
		FixIP:        "",
		IPGroupName:  "",
		Metadata:     make(map[string]string),
		Combinable:   "false",
		Roles:        []string{"std"},
	}

	portEagerAttrControl := PortEagerAttr{
		NetworkName:  "control",
		NetworkPlane: "control",
		PortName:     "nwrfcontrol",
		VnicType:     "normal",
		Accelerate:   "false",
		PodName:      PodName,
		PodNs:        podNS,
		FixIP:        "",
		IPGroupName:  "",
		Metadata:     make(map[string]string),
		Combinable:   "false",
		Roles:        []string{"control"},
	}

	portEagerAttrMedia := PortEagerAttr{
		NetworkName:  "media",
		NetworkPlane: "media",
		PortName:     "nwrfmedia",
		VnicType:     "normal",
		Accelerate:   "false",
		PodName:      PodName,
		PodNs:        podNS,
		FixIP:        "",
		IPGroupName:  "",
		Metadata:     make(map[string]string),
		Combinable:   "false",
		Roles:        []string{"media"},
	}
	portsForService := []*Port{
		{
			EagerAttr: portEagerAttrApi,
		},
		{
			EagerAttr: portEagerAttrControl,
		},
		{
			EagerAttr: portEagerAttrMedia,
		},
	}

	resp := &infra.CreatePortsResp{[]infra.CreatePortInfo{
		{
			Name:      "eth0",
			NetworkID: "net_api",
		},
	}}
	respbyte, _ := json.Marshal(resp)

	var mc *infra.ManagerClient
	monkey.PatchInstanceMethod(reflect.TypeOf(mc), "CreateNeutronBulkPorts",
		func(mc *infra.ManagerClient, reqID string, req *infra.ManagerCreateBulkPortsReq, tenantID string) (b []byte, e error) {
			return respbyte, nil
		})
	defer monkey.UnpatchAll()
	monkey.Patch(destoryBulkPorts, func(mc *infra.ManagerClient, tenantIDs, portIDs []string) {

	})
	convey.Convey("TestPortService_CreateBulkPorts", t, func() {
		ps := &portService{}
		_, err := ps.CreateBulkPorts(podForCreatePort, portsForService)
		convey.So(err, convey.ShouldBeNil)
	})

}

func TestPortService_CreateBulkPortsFail(t *testing.T) {
	PodName := "test1"
	podNS := "admin"
	podForCreatePort := &PodForCreatPort{PodNs: podNS, PodName: PodName}

	portEagerAttrApi := PortEagerAttr{
		NetworkName:  "net_api",
		NetworkPlane: "std",
		PortName:     "eth0",
		VnicType:     "normal",
		Accelerate:   "false",
		PodName:      PodName,
		PodNs:        podNS,
		FixIP:        "",
		IPGroupName:  "",
		Metadata:     make(map[string]string),
		Combinable:   "false",
		Roles:        []string{"std"},
	}

	portEagerAttrControl := PortEagerAttr{
		NetworkName:  "control",
		NetworkPlane: "control",
		PortName:     "nwrfcontrol",
		VnicType:     "normal",
		Accelerate:   "false",
		PodName:      PodName,
		PodNs:        podNS,
		FixIP:        "",
		IPGroupName:  "",
		Metadata:     make(map[string]string),
		Combinable:   "false",
		Roles:        []string{"control"},
	}

	portEagerAttrMedia := PortEagerAttr{
		NetworkName:  "media",
		NetworkPlane: "media",
		PortName:     "nwrfmedia",
		VnicType:     "normal",
		Accelerate:   "false",
		PodName:      PodName,
		PodNs:        podNS,
		FixIP:        "",
		IPGroupName:  "",
		Metadata:     make(map[string]string),
		Combinable:   "false",
		Roles:        []string{"media"},
	}
	portsForService := []*Port{
		{
			EagerAttr: portEagerAttrApi,
		},
		{
			EagerAttr: portEagerAttrControl,
		},
		{
			EagerAttr: portEagerAttrMedia,
		},
	}

	var mc *infra.ManagerClient
	monkey.PatchInstanceMethod(reflect.TypeOf(mc), "CreateNeutronBulkPorts",
		func(_ *infra.ManagerClient, reqID string, req *infra.ManagerCreateBulkPortsReq, tenantID string) (b []byte, e error) {
			return nil, errors.New("create err")
		})
	defer monkey.UnpatchAll()
	monkey.Patch(destoryBulkPorts, func(mc *infra.ManagerClient, tenantIDs, portIDs []string) {

	})
	convey.Convey("TestPortService_CreateBulkPorts", t, func() {
		ps := &portService{}
		_, err := ps.CreateBulkPorts(podForCreatePort, portsForService)
		convey.So(err.Error(), convey.ShouldEqual, "create err")
	})

}

func TestCombinePortObjs(t *testing.T) {
	convey.Convey("TestCombinePortObjsUseEth0First\n", t, func() {
		portObjs := []*Port{
			{
				EagerAttr: PortEagerAttr{
					NetworkName:  "media",
					NetworkPlane: "media",
					PortName:     "eth1",
					VnicType:     "nomal",
					Accelerate:   "true",
					PodName:      "test_pod",
					PodNs:        "test_podns",
					IPGroupName:  "test_IPGroupName",
					Combinable:   "true",
					Metadata:     "test_metadata",
				},
			},
			{
				EagerAttr: PortEagerAttr{
					NetworkName:  "media",
					NetworkPlane: "control",
					PortName:     "eth0",
					VnicType:     "nomal",
					Accelerate:   "true",
					PodName:      "test_pod",
					PodNs:        "test_podns",
					IPGroupName:  "test_IPGroupName",
					Combinable:   "true",
					Metadata:     "test_metadata",
				},
			},
		}

		aimPortObjs := []*Port{
			{
				EagerAttr: PortEagerAttr{
					NetworkName:  "media",
					NetworkPlane: "media",
					PortName:     "eth0",
					VnicType:     "nomal",
					Accelerate:   "true",
					PodName:      "test_pod",
					PodNs:        "test_podns",
					IPGroupName:  "test_IPGroupName",
					Combinable:   "true",
					Metadata:     "test_metadata",
					Roles:        []string{"media", "control"},
				},
			},
		}

		portObjsResult, _ := combinePortObjs(portObjs)
		convey.So(portObjsResult, convey.ShouldResemble, aimPortObjs)
	})
	convey.Convey("TestCombinePortObjsVfNotAccelerateSucc\n", t, func() {
		portObjs := []*Port{
			{
				EagerAttr: PortEagerAttr{
					NetworkName:  "media",
					NetworkPlane: "media",
					PortName:     "eth0",
					VnicType:     "direct",
					Accelerate:   "false",
					PodName:      "test_pod",
					PodNs:        "test_podns",
					IPGroupName:  "test_IPGroupName",
					Combinable:   "true",
					Metadata:     "test_metadata",
				},
			},
			{
				EagerAttr: PortEagerAttr{
					NetworkName:  "media",
					NetworkPlane: "control",
					PortName:     "eth1",
					VnicType:     "direct",
					Accelerate:   "false",
					PodName:      "test_pod",
					PodNs:        "test_podns",
					IPGroupName:  "test_IPGroupName",
					Combinable:   "true",
					Metadata:     "test_metadata",
				},
			},
			{
				EagerAttr: PortEagerAttr{
					NetworkName:  "media",
					NetworkPlane: "eio",
					PortName:     "eth2",
					VnicType:     "direct",
					Accelerate:   "false",
					PodName:      "test_pod",
					PodNs:        "test_podns",
					IPGroupName:  "test_IPGroupName",
					Combinable:   "true",
					Metadata:     "test_metadata",
				},
			},
			{
				EagerAttr: PortEagerAttr{
					NetworkName:  "media",
					NetworkPlane: "oam",
					PortName:     "eth3",
					VnicType:     "direct",
					Accelerate:   "false",
					PodName:      "test_pod",
					PodNs:        "test_podns",
					IPGroupName:  "test_IPGroupName",
					Combinable:   "true",
					Metadata:     "test_metadata",
				},
			},
		}

		aimPortObjs := []*Port{
			{
				EagerAttr: PortEagerAttr{
					NetworkName:  "media",
					NetworkPlane: "media",
					PortName:     "eth0",
					VnicType:     "direct",
					Accelerate:   "false",
					PodName:      "test_pod",
					PodNs:        "test_podns",
					IPGroupName:  "test_IPGroupName",
					Combinable:   "true",
					Metadata:     "test_metadata",
					Roles:        []string{"media", "control", "eio", "oam"},
				},
			},
		}

		portObjsResult, _ := combinePortObjs(portObjs)
		convey.So(portObjsResult, convey.ShouldResemble, aimPortObjs)

	})
	convey.Convey("TestCombinePortObjsC0Succ\n", t, func() {
		portObjs := []*Port{
			{
				EagerAttr: PortEagerAttr{
					NetworkName:  "media",
					NetworkPlane: "media",
					PortName:     "eth0",
					VnicType:     "direct",
					Accelerate:   "true",
					PodName:      "test_pod",
					PodNs:        "test_podns",
					IPGroupName:  "test_IPGroupName",
					Combinable:   "true",
					Metadata:     "test_metadata",
				},
			},
			{
				EagerAttr: PortEagerAttr{
					NetworkName:  "media",
					NetworkPlane: "control",
					PortName:     "eth1",
					VnicType:     "direct",
					Accelerate:   "true",
					PodName:      "test_pod",
					PodNs:        "test_podns",
					IPGroupName:  "test_IPGroupName",
					Combinable:   "true",
					Metadata:     "test_metadata",
				},
			},
			{
				EagerAttr: PortEagerAttr{
					NetworkName:  "media",
					NetworkPlane: "oam",
					PortName:     "eth2",
					VnicType:     "direct",
					Accelerate:   "true",
					PodName:      "test_pod",
					PodNs:        "test_podns",
					IPGroupName:  "test_IPGroupName",
					Combinable:   "true",
					Metadata:     "test_metadata",
				},
			},
		}
		aimPortObjs := []*Port{
			{
				EagerAttr: PortEagerAttr{
					NetworkName:  "media",
					NetworkPlane: "media",
					PortName:     "eth0",
					VnicType:     "direct",
					Accelerate:   "true",
					PodName:      "test_pod",
					PodNs:        "test_podns",
					IPGroupName:  "test_IPGroupName",
					Combinable:   "true",
					Metadata:     "test_metadata",
					Roles:        []string{"media", "control", "oam"},
				},
			},
		}

		portObjsResult, _ := combinePortObjs(portObjs)
		convey.So(portObjsResult, convey.ShouldResemble, aimPortObjs)
	})
	convey.Convey("TestCombinePortObjsCombineTrueAndFalseSucc\n", t, func() {
		portObjs := []*Port{
			{
				EagerAttr: PortEagerAttr{
					NetworkName:  "media",
					NetworkPlane: "media",
					PortName:     "eth0",
					VnicType:     "direct",
					Accelerate:   "true",
					PodName:      "test_pod",
					PodNs:        "test_podns",
					IPGroupName:  "test_IPGroupName",
					Combinable:   "true",
					Metadata:     "test_metadata",
				},
			},
			{
				EagerAttr: PortEagerAttr{
					NetworkName:  "media",
					NetworkPlane: "control",
					PortName:     "eth1",
					VnicType:     "direct",
					Accelerate:   "true",
					PodName:      "test_pod",
					PodNs:        "test_podns",
					IPGroupName:  "test_IPGroupName",
					Combinable:   "true",
					Metadata:     "test_metadata",
				},
			},
			{
				EagerAttr: PortEagerAttr{
					NetworkName:  "media",
					NetworkPlane: "oam",
					PortName:     "eth2",
					VnicType:     "direct",
					Accelerate:   "true",
					PodName:      "test_pod",
					PodNs:        "test_podns",
					IPGroupName:  "test_IPGroupName",
					Combinable:   "true",
					Metadata:     "test_metadata",
				},
			},
			{
				EagerAttr: PortEagerAttr{
					NetworkName:  "media",
					NetworkPlane: "oam",
					PortName:     "eth3",
					VnicType:     "direct",
					Accelerate:   "false",
					PodName:      "test_pod",
					PodNs:        "test_podns",
					IPGroupName:  "test_IPGroupName",
					Combinable:   "false",
					Metadata:     "test_metadata",
				},
			},
		}
		aimPortObjs := []*Port{
			{
				EagerAttr: PortEagerAttr{
					NetworkName:  "media",
					NetworkPlane: "media",
					PortName:     "eth0",
					VnicType:     "direct",
					Accelerate:   "true",
					PodName:      "test_pod",
					PodNs:        "test_podns",
					IPGroupName:  "test_IPGroupName",
					Combinable:   "true",
					Metadata:     "test_metadata",
					Roles:        []string{"media", "control", "oam"},
				},
			},
			{
				EagerAttr: PortEagerAttr{
					NetworkName:  "media",
					NetworkPlane: "oam",
					PortName:     "eth3",
					VnicType:     "direct",
					Accelerate:   "false",
					PodName:      "test_pod",
					PodNs:        "test_podns",
					IPGroupName:  "test_IPGroupName",
					Combinable:   "false",
					Metadata:     "test_metadata",
					Roles:        []string{"oam"},
				},
			},
		}

		portObjsResult, _ := combinePortObjs(portObjs)
		convey.So(portObjsResult, convey.ShouldResemble, aimPortObjs)
	})
	convey.Convey("TestCombinePortObjsCombineTrueAndFalseSucc2\n", t, func() {
		portObjs := []*Port{
			{
				EagerAttr: PortEagerAttr{
					NetworkName:  "media",
					NetworkPlane: "media",
					PortName:     "eth0",
					VnicType:     "direct",
					Accelerate:   "true",
					PodName:      "test_pod",
					PodNs:        "test_podns",
					IPGroupName:  "test_IPGroupName",
					Combinable:   "true",
					Metadata:     "test_metadata",
				},
			},
			{
				EagerAttr: PortEagerAttr{
					NetworkName:  "media",
					NetworkPlane: "oam",
					PortName:     "eth1",
					VnicType:     "direct",
					Accelerate:   "false",
					PodName:      "test_pod",
					PodNs:        "test_podns",
					IPGroupName:  "test_IPGroupName",
					Combinable:   "false",
					Metadata:     "test_metadata",
				},
			},
		}
		aimPortObjs := []*Port{
			{
				EagerAttr: PortEagerAttr{
					NetworkName:  "media",
					NetworkPlane: "media",
					PortName:     "eth0",
					VnicType:     "direct",
					Accelerate:   "true",
					PodName:      "test_pod",
					PodNs:        "test_podns",
					IPGroupName:  "test_IPGroupName",
					Combinable:   "true",
					Metadata:     "test_metadata",
					Roles:        []string{"media"},
				},
			},
			{
				EagerAttr: PortEagerAttr{
					NetworkName:  "media",
					NetworkPlane: "oam",
					PortName:     "eth1",
					VnicType:     "direct",
					Accelerate:   "false",
					PodName:      "test_pod",
					PodNs:        "test_podns",
					IPGroupName:  "test_IPGroupName",
					Combinable:   "false",
					Metadata:     "test_metadata",
					Roles:        []string{"oam"},
				},
			},
		}

		portObjsResult, _ := combinePortObjs(portObjs)
		convey.So(portObjsResult, convey.ShouldResemble, aimPortObjs)
	})
	convey.Convey("TestCombinePortObjsWithSameNetPlaneError\n", t, func() {
		portObjs := []*Port{
			{
				EagerAttr: PortEagerAttr{
					NetworkName:  "media",
					NetworkPlane: "media",
					PortName:     "eth0",
					VnicType:     "nomal",
					Accelerate:   "false",
					PodName:      "test_pod",
					PodNs:        "test_podns",
					IPGroupName:  "test_IPGroupName",
					Combinable:   "true",
					Metadata:     "test_metadata",
				},
			},
			{
				EagerAttr: PortEagerAttr{
					NetworkName:  "media",
					NetworkPlane: "media",
					PortName:     "eth1",
					VnicType:     "nomal",
					Accelerate:   "false",
					PodName:      "test_pod",
					PodNs:        "test_podns",
					IPGroupName:  "test_IPGroupName",
					Combinable:   "true",
					Metadata:     "test_metadata",
				},
			},
		}
		portObjsResult, err := combinePortObjs(portObjs)
		convey.So(portObjsResult, convey.ShouldBeNil)
		convey.So(err.Error(), convey.ShouldNotBeNil)
	})
	convey.Convey("TestCombinePortObjsWithEioAndC0ConflictError\n", t, func() {
		portObjs := []*Port{
			{
				EagerAttr: PortEagerAttr{
					NetworkName:  "media",
					NetworkPlane: "media",
					PortName:     "eth0",
					VnicType:     "nomal",
					Accelerate:   "true",
					PodName:      "test_pod",
					PodNs:        "test_podns",
					IPGroupName:  "test_IPGroupName",
					Combinable:   "true",
					Metadata:     "test_metadata",
				},
			},
			{
				EagerAttr: PortEagerAttr{
					NetworkName:  "media",
					NetworkPlane: "eio",
					PortName:     "eth1",
					VnicType:     "nomal",
					Accelerate:   "true",
					PodName:      "test_pod",
					PodNs:        "test_podns",
					IPGroupName:  "test_IPGroupName",
					Combinable:   "true",
					Metadata:     "test_metadata",
				},
			},
		}
		portObjsResult, err := combinePortObjs(portObjs)
		convey.So(portObjsResult, convey.ShouldBeNil)
		convey.So(err.Error(), convey.ShouldNotBeNil)
	})

}

func TestFillPortLazyAttr(t *testing.T) {
	PodName := "testpod"
	podNS := "admin"

	portEagerAttrApi := PortEagerAttr{
		NetworkName:  "net_api",
		NetworkPlane: "net_api",
		PortName:     "eth0",
		VnicType:     "normal",
		Accelerate:   "false",
		PodName:      PodName,
		PodNs:        podNS,
		FixIP:        "192.168.1.0",
		IPGroupName:  "ig0",
		Metadata:     "",
		Combinable:   "false",
		Roles:        []string{"control"},
	}

	portLazyAttrApi := PortLazyAttr{
		ID:         "1111",
		Name:       "eth0",
		TenantID:   podNS,
		MacAddress: "0000",
		GatewayIP:  "192.0.0.0",
		Cidr:       "192.0.0.0/8",
	}

	portsForServiceExpect := []*Port{
		{
			EagerAttr: portEagerAttrApi,
			LazyAttr:  portLazyAttrApi,
		},
	}
	portsForService := []*Port{
		{
			EagerAttr: portEagerAttrApi,
		},
	}

	resp := &infra.CreatePortsResp{[]infra.CreatePortInfo{
		{
			Name:       "eth0",
			NetworkID:  "net_api",
			MacAddress: "0000",
			GatewayIP:  "192.0.0.0",
			Cidr:       "192.0.0.0/8",
			PortID:     "1111",
		},
	}}
	convey.Convey("TestFillPortLazyAttr", t, func() {
		fillPortLazyAttr(resp, portsForService, "admin")
		convey.So(portsForService, convey.ShouldResemble, portsForServiceExpect)
	})

}
