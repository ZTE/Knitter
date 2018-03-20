package services

import (
	"errors"
	"reflect"
	"testing"

	"github.com/bouk/monkey"
	"github.com/golang/mock/gomock"
	"github.com/smartystreets/goconvey/convey"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/util/workqueue"

	"github.com/ZTE/Knitter/knitter-monitor/infra"
	"github.com/ZTE/Knitter/knitter-monitor/tests/mocks/kubernetes"
)

func TestNewCreatePortForPodController(t *testing.T) {
	monkey.Patch(infra.GetClientset, func() *kubernetes.Clientset {
		return &kubernetes.Clientset{}
	})
	defer monkey.UnpatchAll()

	convey.Convey("TestNewCreatePortForPodController", t, func() {
		_, err := NewCreatePortForPodController()
		convey.So(err, convey.ShouldBeNil)
	})
}

func TestNewCreatePortForPodControllerKubernetesNilFail(t *testing.T) {
	convey.Convey("TestNewCreatePortForPodControllerKubernetesNilFail", t, func() {
		_, err := NewCreatePortForPodController()
		convey.So(err.Error(), convey.ShouldEqual, "kubernetes clientset is nil")
	})
}

func TestEnqueueCreatePod(t *testing.T) {
	controller := &createPortForPodController{
		podsCreateQueue: workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "pods"),
		podsDeleteQueue: workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "pods"),
	}
	objectMeta := metav1.ObjectMeta{
		Name:      "pod1",
		UID:       "testpod-1111",
		Namespace: "admin",
	}
	k8sPod := &v1.Pod{
		ObjectMeta: objectMeta,
	}

	convey.Convey("TestEnqueueCreatePod", t, func() {
		controller.enqueueCreatePod(k8sPod)
		key, _ := controller.podsCreateQueue.Get()
		expect := "admin/pod1"

		convey.So(key.(string), convey.ShouldEqual, expect)
	})

}

func TestEnqueueDeletePod(t *testing.T) {
	controller := &createPortForPodController{
		podsCreateQueue: workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "pods"),
		podsDeleteQueue: workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "pods"),
	}
	objectMeta := metav1.ObjectMeta{
		Name:      "pod1",
		UID:       "testpod-1111",
		Namespace: "admin",
	}
	k8sPod := &v1.Pod{
		ObjectMeta: objectMeta,
	}

	convey.Convey("TestEnqueueCreatePod", t, func() {
		controller.enqueueDeletePod(k8sPod)
		obj, _ := controller.podsDeleteQueue.Get()

		convey.So(obj.(*v1.Pod), convey.ShouldResemble, k8sPod)
	})

}

func TestCreatePodWorker(t *testing.T) {

	controller := &createPortForPodController{
		podsCreateQueue: workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "pods"),
		podsDeleteQueue: workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "pods"),
	}
	objectMeta := metav1.ObjectMeta{
		Name:      "pod1",
		UID:       "testpod-1111",
		Namespace: "admin",
	}
	k8sPod := &v1.Pod{
		ObjectMeta: objectMeta,
	}
	mockControlelr := gomock.NewController(t)
	defer mockControlelr.Finish()
	mockIndexer := mockcache.NewMockIndexer(mockControlelr)
	mockIndexer.EXPECT().GetByKey("admin/pod1").Return(k8sPod, true, nil)
	controller.podStoreIndexer = mockIndexer
	controller.podsCreateQueue.Add("admin/pod1")

	var ps *podService
	monkey.PatchInstanceMethod(reflect.TypeOf(ps), "NewPodFromK8sPod",
		func(ps *podService, k8sPod *v1.Pod) (*Pod, error) {
			return &Pod{PodName: "pod1"}, nil
		})
	defer monkey.UnpatchAll()
	monkey.PatchInstanceMethod(reflect.TypeOf(ps), "Save",
		func(ps *podService, pod *Pod) error {
			return nil
		})
	convey.Convey("TestCreatePodWorker", t, func() {
		controller.podsCreateQueue.ShutDown()
		controller.createPodWorker()

	})
}

func TestDeletePodWorker(t *testing.T) {

	controller := &createPortForPodController{
		podsCreateQueue: workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "pods"),
		podsDeleteQueue: workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "pods"),
	}
	objectMeta := metav1.ObjectMeta{
		Name:      "pod1",
		UID:       "testpod-1111",
		Namespace: "admin",
	}
	k8sPod := &v1.Pod{
		ObjectMeta: objectMeta,
	}

	controller.podsCreateQueue.Add(k8sPod)

	var ps *podService
	monkey.PatchInstanceMethod(reflect.TypeOf(ps), "DeletePodAndPorts",
		func(ps *podService, podNs, podName string) error {
			return nil
		})
	defer monkey.UnpatchAll()

	convey.Convey("TestDeletePodWorker", t, func() {
		controller.podsDeleteQueue.ShutDown()
		controller.deletePodWorker()

	})
}

func TestDeletePodWorkerFail(t *testing.T) {

	controller := &createPortForPodController{
		podsCreateQueue: workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "pods"),
		podsDeleteQueue: workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "pods"),
	}
	objectMeta := metav1.ObjectMeta{
		Name:      "pod1",
		UID:       "testpod-1111",
		Namespace: "admin",
	}
	k8sPod := &v1.Pod{
		ObjectMeta: objectMeta,
	}

	controller.podsCreateQueue.Add(k8sPod)

	var ps *podService
	monkey.PatchInstanceMethod(reflect.TypeOf(ps), "DeletePodAndPorts",
		func(ps *podService, podNs, podName string) error {
			return errors.New("delete err")
		})
	defer monkey.UnpatchAll()

	convey.Convey("TestDeletePodWorker", t, func() {
		controller.podsDeleteQueue.ShutDown()
		controller.deletePodWorker()

	})
}
