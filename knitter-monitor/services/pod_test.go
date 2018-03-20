package services

import (
	"encoding/json"
	"errors"
	"github.com/ZTE/Knitter/knitter-monitor/const-value"
	"github.com/ZTE/Knitter/knitter-monitor/daos"
	"github.com/ZTE/Knitter/knitter-monitor/infra"
	"github.com/ZTE/Knitter/knitter-monitor/tests/mocks/daos"
	"github.com/antonholmquist/jason"
	"github.com/bouk/monkey"
	"github.com/golang/mock/gomock"
	"github.com/smartystreets/goconvey/convey"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"reflect"
	"testing"
)

func TestPodService_NewPodFromK8sPodSucc(t *testing.T) {
	podNS, PodName, k8sPod := createK8sPod()

	portsForService, podForServiceExpect := createPod(PodName, podNS)

	var service *portService
	monkey.PatchInstanceMethod(reflect.TypeOf(service),
		"NewPortsWithEagerAttrAndLazyAttr",
		func(_ *portService, pod *PodForCreatPort, nwJSON *jason.Object) ([]*Port, error) {
			return portsForService, nil
		})
	var ps *podService
	monkey.PatchInstanceMethod(reflect.TypeOf(ps), "Get", func(_ *podService, podNs, podName string) (*Pod, error) {
		return nil, errors.New("Key not found")
	})
	defer monkey.UnpatchAll()
	convey.Convey("TestNewPodFromK8sPodSucc", t, func() {
		ps := &podService{}
		pod, err := ps.NewPodFromK8sPod(k8sPod)
		convey.So(pod, convey.ShouldResemble, podForServiceExpect)
		convey.So(err, convey.ShouldBeNil)
	})

}

func createPod(PodName string, podNS string) ([]*Port, *Pod) {
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
		ID:         "net_api",
		Name:       "eth0",
		TenantID:   podNS,
		MacAddress: "",
	}

	portEagerAttrControl := PortEagerAttr{
		NetworkName:  "control",
		NetworkPlane: "control",
		PortName:     "nwrfcontrol",
		VnicType:     "normal",
		Accelerate:   "false",
		PodName:      PodName,
		PodNs:        podNS,
		FixIP:        "192.168.1.1",
		IPGroupName:  "ig1",
		Metadata:     "",
		Combinable:   "false",
		Roles:        []string{"control"},
	}
	portLazyAttrControl := PortLazyAttr{
		ID:         "11111-control",
		Name:       "nwrfcontrol",
		TenantID:   podNS,
		MacAddress: "",
	}

	portEagerAttrMedia := PortEagerAttr{
		NetworkName:  "media",
		NetworkPlane: "media",
		PortName:     "eth2",
		VnicType:     "normal",
		Accelerate:   "false",
		PodName:      PodName,
		PodNs:        podNS,
		FixIP:        "192.168.1.2",
		IPGroupName:  "ig2",
		Metadata:     "",
		Combinable:   "false",
		Roles:        []string{"media"},
	}
	portLazyAttrMedia := PortLazyAttr{
		ID:         "11111-media",
		Name:       "eth1",
		TenantID:   podNS,
		MacAddress: "",
	}
	portsForService := []*Port{
		{
			EagerAttr: portEagerAttrApi,
			LazyAttr:  portLazyAttrApi,
		},
		{
			EagerAttr: portEagerAttrControl,
			LazyAttr:  portLazyAttrControl,
		},
		{
			EagerAttr: portEagerAttrMedia,
			LazyAttr:  portLazyAttrMedia,
		},
	}
	podForServiceExpect := &Pod{
		TenantId:     podNS,
		PodID:        "testpod-1111",
		PodName:      PodName,
		PodNs:        podNS,
		PodType:      "",
		IsSuccessful: true,
		ErrorMsg:     "",
		Ports:        portsForService,
	}
	return portsForService, podForServiceExpect
}

func createK8sPod() (string, string, *v1.Pod) {
	podNS := "admin"
	PodName := "pod1"
	networks := "{\"ports\": " +
		"[{\"attach_to_network\": \"net_api\", \"attributes\": {\"accelerate\": \"false\", \"function\": \"std\", \"nic_name\": \"eth0\", \"nic_type\": \"normal\"}}," +
		" {\"attach_to_network\": \"control\", \"attributes\": {\"accelerate\": \"false\", \"function\": \"control\", \"nic_name\": \"nwrfcontrol\", \"nic_type\": \"normal\"}}, " +
		"{\"attach_to_network\": \"media\", \"attributes\": {\"accelerate\": \"false\", \"function\": \"media\", \"nic_name\": \"nwrfmedia\", \"nic_type\": \"normal\"}}]}"
	annotations := map[string]string{
		"networks": networks,
	}
	objectMeta := metav1.ObjectMeta{
		Name:        PodName,
		UID:         "testpod-1111",
		Namespace:   podNS,
		Annotations: annotations,
	}
	k8sPod := &v1.Pod{
		ObjectMeta: objectMeta,
	}
	return podNS, PodName, k8sPod
}

func TestPodService_NewPodFromK8sPodUseDefaultNetwork(t *testing.T) {
	podNs, podName, k8sPod := createK8sPodNoNetwokStr()
	portsForService, podForServiceExpect := createPodDefaultNetwork(podName, podNs)

	var service *portService
	monkey.PatchInstanceMethod(reflect.TypeOf(service),
		"NewPortsWithEagerAttrAndLazyAttr",
		func(_ *portService, pod *PodForCreatPort, nwJSON *jason.Object) ([]*Port, error) {
			return portsForService, nil
		})
	defer monkey.UnpatchAll()

	var ps *podService
	monkey.PatchInstanceMethod(reflect.TypeOf(ps), "Get", func(_ *podService, podNs, podName string) (*Pod, error) {
		return nil, errors.New("Key not found")
	})

	var mc *infra.ManagerClient
	monkey.PatchInstanceMethod(reflect.TypeOf(mc), "GetDefaultNetWork",
		func(_ *infra.ManagerClient, _ string) (string, error) {
			return "net_api", nil
		})

	convey.Convey("TestNewPodFromK8sPodSucc", t, func() {
		ps := &podService{}
		pod, err := ps.NewPodFromK8sPod(k8sPod)
		convey.So(pod, convey.ShouldResemble, podForServiceExpect)
		convey.So(err, convey.ShouldBeNil)
	})

}

func createK8sPodNoNetwokStr() (string, string, *v1.Pod) {
	podNS := "admin"
	PodName := "pod1"
	networks := ""
	annotations := map[string]string{
		"networks": networks,
	}
	objectMeta := metav1.ObjectMeta{
		Name:        PodName,
		UID:         "testpod-1111",
		Namespace:   podNS,
		Annotations: annotations,
	}
	k8sPod := &v1.Pod{
		ObjectMeta: objectMeta,
	}
	return podNS, PodName, k8sPod
}

func createPodDefaultNetwork(PodName string, podNS string) ([]*Port, *Pod) {
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
		ID:         "net_api",
		Name:       "eth0",
		TenantID:   podNS,
		MacAddress: "",
	}

	portsForService := []*Port{
		{
			EagerAttr: portEagerAttrApi,
			LazyAttr:  portLazyAttrApi,
		},
	}
	podForServiceExpect := &Pod{
		TenantId:     podNS,
		PodID:        "testpod-1111",
		PodName:      PodName,
		PodNs:        podNS,
		PodType:      "",
		IsSuccessful: true,
		ErrorMsg:     "",
		Ports:        portsForService,
	}
	return portsForService, podForServiceExpect
}

func TestIsNetworkNotConfigExist(t *testing.T) {

	convey.Convey("TestIsNetworkNotConfigExist", t, func() {
		nwStr := ""
		ok := true
		expect := isNetworkNotConfigExist(nwStr, ok)
		convey.So(expect, convey.ShouldBeTrue)

		nwStr = "\"\""
		expect = isNetworkNotConfigExist(nwStr, ok)
		convey.So(expect, convey.ShouldBeTrue)

		ok = false
		nwStr = ""
		expect = isNetworkNotConfigExist(nwStr, ok)
		convey.So(expect, convey.ShouldBeTrue)

		ok = true
		nwStr = "nwstr"
		expect = isNetworkNotConfigExist(nwStr, ok)
		convey.So(expect, convey.ShouldBeFalse)

	})
}

func TestGetDefaultNetworkConfigSucc(t *testing.T) {
	defaultNetName := "net_api"
	networkByteExpect := createDefaultNwbyte(defaultNetName)

	var mc *infra.ManagerClient
	monkey.PatchInstanceMethod(reflect.TypeOf(mc), "GetDefaultNetWork",
		func(_ *infra.ManagerClient, _ string) (string, error) {
			return defaultNetName, nil
		})
	defer monkey.UnpatchAll()

	convey.Convey("TestGetDefaultNetworkConfigSucc", t, func() {
		networkByte, err := GetDefaultNetworkConfig()
		convey.So(networkByte, convey.ShouldResemble, networkByteExpect)
		convey.So(err, convey.ShouldBeNil)
	})

}

func createDefaultNwbyte(defaultNetName string) []byte {
	bpnmExpect := BluePrintNetworkMessage{}
	ports := make([]BluePrintPort, 0)
	port := BluePrintPort{
		AttachToNetwork: defaultNetName,
		Attributes: BluePrintAttributes{
			Accelerate: constvalue.DefaultIsAccelerate,
			Function:   constvalue.DefaultNetworkPlane,
			NicName:    constvalue.DefaultPortName,
			NicType:    constvalue.DefaultVnicType,
		},
	}
	ports = append(ports, port)
	bpnmExpect.Ports = ports
	networkByteExpect, _ := json.Marshal(&bpnmExpect)
	return networkByteExpect
}

func TestPodService_SaveSucc(t *testing.T) {
	podNs := "admin"
	podName := "test1"

	pod := &Pod{
		TenantId: podNs,
		PodName:  podName,
		PodNs:    podNs,
	}

	mockController := gomock.NewController(t)
	defer mockController.Finish()
	mockDao := mockdaos.NewMockPodDaoInterface(mockController)
	mockDao.EXPECT().Save(gomock.Any()).Return(nil)

	monkey.Patch(daos.GetPodDao, func() daos.PodDaoInterface {
		return mockDao
	})
	defer monkey.UnpatchAll()

	convey.Convey("TestPodService_GetSucc", t, func() {
		err := GetPodService().Save(pod)
		convey.So(err, convey.ShouldBeNil)
	})
}
func TestPodService_SaveFail(t *testing.T) {
	podNs := "admin"
	podName := "test1"
	pod := &Pod{
		TenantId: podNs,
		PodName:  podName,
		PodNs:    podNs,
	}

	mockController := gomock.NewController(t)
	defer mockController.Finish()
	mockDao := mockdaos.NewMockPodDaoInterface(mockController)
	mockDao.EXPECT().Save(gomock.Any()).Return(errors.New("save err"))

	monkey.Patch(daos.GetPodDao, func() daos.PodDaoInterface {
		return mockDao
	})
	defer monkey.UnpatchAll()

	convey.Convey("TestPodService_SaveFail", t, func() {
		err := GetPodService().Save(pod)
		convey.So(err.Error(), convey.ShouldEqual, "save err")
	})
}

func TestPodService_Get(t *testing.T) {
	podNs := "admin"
	podName := "test1"

	portsForDb := []*daos.PortForDB{
		{
			NetworkName:  "net_api",
			NetworkPlane: "net_api",
			PortName:     "eth0",
			VnicType:     "normal",
			Accelerate:   "false",
			PodName:      podName,
			PodNs:        podNs,
			FixIP:        "192.168.1.0",
			IPGroupName:  "ig0",
			Metadata:     "",
			Combinable:   "false",
			Roles:        []string{"control"},

			ID:         "net_api",
			LazyName:   "eth0",
			TenantID:   podNs,
			MacAddress: "",
		},
	}

	podForDb := &daos.PodForDB{
		TenantId: podNs,
		PodNs:    podNs,
		PodName:  podName,
		Ports:    portsForDb,
	}

	portEagerAttrApi := PortEagerAttr{
		NetworkName:  "net_api",
		NetworkPlane: "net_api",
		PortName:     "eth0",
		VnicType:     "normal",
		Accelerate:   "false",
		PodName:      podName,
		PodNs:        podNs,
		FixIP:        "192.168.1.0",
		IPGroupName:  "ig0",
		Metadata:     "",
		Combinable:   "false",
		Roles:        []string{"control"},
	}

	portLazyAttrApi := PortLazyAttr{
		ID:         "net_api",
		Name:       "eth0",
		TenantID:   podNs,
		MacAddress: "",
	}

	portsForService := []*Port{
		{
			EagerAttr: portEagerAttrApi,
			LazyAttr:  portLazyAttrApi,
		},
	}

	podExpect := &Pod{
		TenantId: podNs,
		PodName:  podName,
		PodNs:    podNs,
		Ports:    portsForService,
	}

	mockController := gomock.NewController(t)
	defer mockController.Finish()
	mockDao := mockdaos.NewMockPodDaoInterface(mockController)
	mockDao.EXPECT().Get(podNs, podName).Return(podForDb, nil)

	monkey.Patch(daos.GetPodDao, func() daos.PodDaoInterface {
		return mockDao
	})
	defer monkey.UnpatchAll()

	convey.Convey("TestPodService_GetSucc", t, func() {
		pod, err := GetPodService().Get(podNs, podName)
		convey.So(pod, convey.ShouldResemble, podExpect)
		convey.So(err, convey.ShouldBeNil)
	})
}

func TestPodService_GetFail(t *testing.T) {
	podNs := "admin"
	podName := "test1"

	mockController := gomock.NewController(t)
	defer mockController.Finish()
	mockDao := mockdaos.NewMockPodDaoInterface(mockController)
	mockDao.EXPECT().Get(podNs, podName).Return(nil, errors.New("get err"))

	monkey.Patch(daos.GetPodDao, func() daos.PodDaoInterface {
		return mockDao
	})
	defer monkey.UnpatchAll()

	convey.Convey("TestPodService_GetFail", t, func() {
		pod, err := GetPodService().Get(podNs, podName)
		convey.So(pod, convey.ShouldBeNil)
		convey.So(err.Error(), convey.ShouldEqual, "get err")
	})
}

func TestPodService_DeleteSucc(t *testing.T) {
	podNs := "admin"
	podName := "test1"
	podForDb := &daos.PodForDB{
		TenantId: podNs,
		PodNs:    podNs,
		PodName:  podName,
	}

	mockController := gomock.NewController(t)
	defer mockController.Finish()
	mockDao := mockdaos.NewMockPodDaoInterface(mockController)
	mockDao.EXPECT().Get(podNs, podName).Return(podForDb, nil)
	mockDao.EXPECT().Delete(podNs, podName).Return(nil)
	monkey.Patch(daos.GetPodDao, func() daos.PodDaoInterface {
		return mockDao
	})
	defer monkey.UnpatchAll()

	var service *portService
	monkey.PatchInstanceMethod(reflect.TypeOf(service),
		"DeleteBulkPorts",
		func(_ *portService, tenantID string, portIDs []string) error {
			return nil
		})

	convey.Convey("TestPodService_GetSucc", t, func() {
		err := GetPodService().DeletePodAndPorts(podNs, podName)
		convey.So(err, convey.ShouldBeNil)
	})
}

func TestPodService_DeleteGetFail(t *testing.T) {
	podNs := "admin"
	podName := "test1"

	mockController := gomock.NewController(t)
	defer mockController.Finish()
	mockDao := mockdaos.NewMockPodDaoInterface(mockController)
	mockDao.EXPECT().Get(podNs, podName).Return(nil, errors.New("delete err"))
	monkey.Patch(daos.GetPodDao, func() daos.PodDaoInterface {
		return mockDao
	})
	defer monkey.UnpatchAll()

	convey.Convey("TestPodService_DeleteGetFail", t, func() {
		err := GetPodService().DeletePodAndPorts(podNs, podName)
		convey.So(err.Error(), convey.ShouldEqual, "delete err")
	})
}

func TestPodService_DeleteGetNotFoundSucc(t *testing.T) {
	podNs := "admin"
	podName := "test1"

	mockController := gomock.NewController(t)
	defer mockController.Finish()
	mockDao := mockdaos.NewMockPodDaoInterface(mockController)
	mockDao.EXPECT().Get(podNs, podName).Return(nil, errors.New("Key not found"))
	monkey.Patch(daos.GetPodDao, func() daos.PodDaoInterface {
		return mockDao
	})
	defer monkey.UnpatchAll()

	convey.Convey("TestPodService_DeleteGetNotFoundSucc", t, func() {
		err := GetPodService().DeletePodAndPorts(podNs, podName)
		convey.So(err, convey.ShouldBeNil)
	})
}

func TestPodService_DeletebulkPortsFail(t *testing.T) {
	podNs := "admin"
	podName := "test1"
	podForDb := &daos.PodForDB{
		TenantId: podNs,
		PodNs:    podNs,
		PodName:  podName,
	}

	mockController := gomock.NewController(t)
	defer mockController.Finish()
	mockDao := mockdaos.NewMockPodDaoInterface(mockController)
	mockDao.EXPECT().Get(podNs, podName).Return(podForDb, nil)
	monkey.Patch(daos.GetPodDao, func() daos.PodDaoInterface {
		return mockDao
	})
	defer monkey.UnpatchAll()

	var service *portService
	monkey.PatchInstanceMethod(reflect.TypeOf(service),
		"DeleteBulkPorts",
		func(_ *portService, tenantID string, portIDs []string) error {
			return errors.New("delete bulk ports err")
		})

	convey.Convey("TestPodService_DeletebulkPortsFail", t, func() {
		err := GetPodService().DeletePodAndPorts(podNs, podName)
		convey.So(err.Error(), convey.ShouldEqual, "delete bulk ports err")
	})
}

func TestPodService_DeleteFail(t *testing.T) {
	podNs := "admin"
	podName := "test1"

	portsForDb := []*daos.PortForDB{
		{
			NetworkName:  "net_api",
			NetworkPlane: "net_api",
			PortName:     "eth0",
			VnicType:     "normal",
			Accelerate:   "false",
			PodName:      podName,
			PodNs:        podNs,
			FixIP:        "192.168.1.0",
			IPGroupName:  "ig0",
			Metadata:     "",
			Combinable:   "false",
			Roles:        []string{"control"},
			ID:           "net_api",
			TenantID:     podNs,
			MacAddress:   "",
		}}
	podForDb := &daos.PodForDB{
		TenantId: podNs,
		PodNs:    podNs,
		PodName:  podName,
		Ports:    portsForDb,
	}

	mockController := gomock.NewController(t)
	defer mockController.Finish()
	mockDao := mockdaos.NewMockPodDaoInterface(mockController)
	mockDao.EXPECT().Get(podNs, podName).Return(podForDb, nil)
	mockDao.EXPECT().Delete(podNs, podName).Return(errors.New("dao delete err"))
	monkey.Patch(daos.GetPodDao, func() daos.PodDaoInterface {
		return mockDao
	})
	defer monkey.UnpatchAll()

	var service *portService
	monkey.PatchInstanceMethod(reflect.TypeOf(service),
		"DeleteBulkPorts",
		func(_ *portService, tenantID string, portIDs []string) error {
			return nil
		})

	convey.Convey("TestPodService_DeleteFail", t, func() {
		err := GetPodService().DeletePodAndPorts(podNs, podName)
		convey.So(err.Error(), convey.ShouldEqual, "dao delete err")
	})
}

func TestTransferToPodForDaoSucc(t *testing.T) {

	podNs := "admin"
	podName := "test1"
	portsForDb := []*daos.PortForDB{
		{
			NetworkName:  "net_api",
			NetworkPlane: "net_api",
			PortName:     "eth0",
			VnicType:     "normal",
			Accelerate:   "false",
			PodName:      podName,
			PodNs:        podNs,
			FixIP:        "192.168.1.0",
			IPGroupName:  "ig0",
			Metadata:     "",
			Combinable:   "false",
			Roles:        []string{"control"},

			ID:         "net_api",
			LazyName:   "eth0",
			TenantID:   podNs,
			MacAddress: "",
		},
	}
	podForDb := &daos.PodForDB{
		TenantId:     podNs,
		PodID:        "testpod-1111",
		PodNs:        podNs,
		PodName:      podName,
		PodType:      "",
		ErrorMsg:     "",
		IsSuccessful: true,
		Ports:        portsForDb,
	}

	portEagerAttrApi := PortEagerAttr{
		NetworkName:  "net_api",
		NetworkPlane: "net_api",
		PortName:     "eth0",
		VnicType:     "normal",
		Accelerate:   "false",
		PodName:      podName,
		PodNs:        podNs,
		FixIP:        "192.168.1.0",
		IPGroupName:  "ig0",
		Metadata:     "",
		Combinable:   "false",
		Roles:        []string{"control"},
	}

	portLazyAttrApi := PortLazyAttr{
		ID:         "net_api",
		Name:       "eth0",
		TenantID:   podNs,
		MacAddress: "",
	}

	portsForService := []*Port{
		{
			EagerAttr: portEagerAttrApi,
			LazyAttr:  portLazyAttrApi,
		},
	}
	pod := &Pod{
		TenantId:     podNs,
		PodID:        "testpod-1111",
		PodName:      podName,
		PodNs:        podNs,
		PodType:      "",
		IsSuccessful: true,
		ErrorMsg:     "",
		Ports:        portsForService,
	}
	convey.Convey("TeestTransferToPodForDaoSucc", t, func() {
		podForDbExpect := pod.transferToPodForDao()
		convey.So(podForDb, convey.ShouldResemble, podForDbExpect)
	})
}
