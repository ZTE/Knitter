package apps

import (
	"errors"
	"github.com/ZTE/Knitter/knitter-monitor/err-obj"
	"github.com/ZTE/Knitter/knitter-monitor/services"
	"github.com/ZTE/Knitter/knitter-monitor/tests/mocks/services"
	"github.com/bouk/monkey"
	"github.com/golang/mock/gomock"
	"github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestPodAppGet(t *testing.T) {
	podNs := "admin"
	podName := "pod9527"
	portEagerAttr := services.PortEagerAttr{
		NetworkName:  "control",
		NetworkPlane: "control",
		PortName:     "eth1",
		VnicType:     "normal",
		Accelerate:   "true",
		PodName:      podName,
		PodNs:        podNs,
		FixIP:        "",
		IPGroupName:  "ig1",
		Metadata:     "",
		Combinable:   "true",
		Roles:        []string{"media", "control"},
	}

	portLazyAttr := services.PortLazyAttr{
		ID:         "11111",
		Name:       "",
		TenantID:   podNs,
		MacAddress: "",
	}

	portsForService := []*services.Port{
		{
			EagerAttr: portEagerAttr,
			LazyAttr:  portLazyAttr,
		},
	}

	podForService := &services.Pod{
		TenantId:     podNs,
		PodID:        "pod1",
		PodName:      podName,
		PodNs:        podNs,
		PodType:      "test",
		IsSuccessful: true,
		ErrorMsg:     "",
		Ports:        portsForService,
	}

	portsForAgentExpect := []*PortForAgent{
		{
			EagerAttr: PortEagerAttrForAgent{
				NetworkName:  "control",
				NetworkPlane: "control",
				PortName:     "eth1",
				VnicType:     "normal",
				Accelerate:   "true",
				PodName:      podName,
				PodNs:        podNs,
				FixIP:        "",
				IPGroupName:  "ig1",
				Metadata:     "",
				Combinable:   "true",
				Roles:        []string{"media", "control"},
			},
			LazyAttr: PortLazyAttrForAgent{
				ID:         "11111",
				Name:       "",
				TenantID:   podNs,
				MacAddress: "",
			},
		},
	}
	podForAgentExcept := &PodForAgent{
		TenantId:     podNs,
		PodID:        "pod1",
		PodName:      podName,
		PodNs:        podNs,
		PodType:      "test",
		IsSuccessful: true,
		ErrorMsg:     "",
		Ports:        portsForAgentExpect,
	}

	convey.Convey("TestPodAppGetSuccess", t, func() {

		mockController := gomock.NewController(t)
		defer mockController.Finish()
		mockPodService := mockservices.NewMockPodServiceInterface(mockController)
		mockPodService.EXPECT().Get(podNs, podName).Return(podForService, nil)

		guard := monkey.Patch(services.GetPodService, func() services.PodServiceInterface {
			return mockPodService
		})
		defer monkey.Unpatch(guard)

		podForAgent, err := GetPodApp().Get(podNs, podName)
		convey.So(podForAgent, convey.ShouldResemble, podForAgentExcept)
		convey.So(err, convey.ShouldBeNil)
	})

}

func TestPodAppGetPodNameIsNilFail(t *testing.T) {
	podNs := ""
	podName := ""
	podForAgent, err := GetPodApp().Get(podNs, podName)
	convey.Convey("TestPodAppGetPodNameIsNilFail", t, func() {
		convey.So(podForAgent, convey.ShouldBeNil)
		convey.So(err, convey.ShouldResemble, errobj.ErrPodNSOrPodNameIsNil)
	})
}

func TestPodAppGetServiceGetFail(t *testing.T) {
	podNs := "admin"
	podName := "pod9527"

	convey.Convey("TestPodAppGetServiceGetFail", t, func() {

		mockController := gomock.NewController(t)
		defer mockController.Finish()
		mockPodService := mockservices.NewMockPodServiceInterface(mockController)
		mockPodService.EXPECT().Get(podNs, podName).Return(nil, errors.New("service get error"))

		guard := monkey.Patch(services.GetPodService, func() services.PodServiceInterface {
			return mockPodService
		})
		defer monkey.Unpatch(guard)

		podForAgent, err := GetPodApp().Get(podNs, podName)
		convey.So(podForAgent, convey.ShouldBeNil)
		convey.So(err.Error(), convey.ShouldEqual, "service get error")
	})

}
