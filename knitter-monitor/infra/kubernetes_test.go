package infra

import (
	"errors"
	"testing"

	"github.com/bouk/monkey"
	"github.com/golang/mock/gomock"
	"github.com/smartystreets/goconvey/convey"
	"k8s.io/client-go/kubernetes"
	rest "k8s.io/client-go/rest"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/ZTE/Knitter/knitter-agent/tests/mock/db-mock"
	"github.com/ZTE/Knitter/knitter-monitor/err-obj"
	"github.com/ZTE/Knitter/pkg/db-accessor"
)

func TestInitKubernetesClientSet(t *testing.T) {

	monkey.Patch(clientcmd.BuildConfigFromFlags, func(_ string, _ string) (*restclient.Config, error) {
		return &restclient.Config{}, nil
	})
	defer monkey.UnpatchAll()
	monkey.Patch(kubernetes.NewForConfig, func(c *rest.Config) (*kubernetes.Clientset, error) {
		return &kubernetes.Clientset{}, nil
	})
	convey.Convey("TestInitKubernetesClientset", t, func() {
		err := InitKubernetesClientset("")
		convey.So(err, convey.ShouldBeNil)
	})
}

func TestInitKubernetesClientSetBuildConfigErr(t *testing.T) {

	monkey.Patch(clientcmd.BuildConfigFromFlags, func(_ string, _ string) (*restclient.Config, error) {
		return &restclient.Config{}, errors.New("build err")
	})
	defer monkey.UnpatchAll()

	convey.Convey("TestInitKubernetesClientSetBuildConfigErr", t, func() {
		err := InitKubernetesClientset("")
		convey.So(err.Error(), convey.ShouldEqual, "build err")
	})
}

func TestInitKubernetesClientSetNewForConfigErr(t *testing.T) {

	monkey.Patch(clientcmd.BuildConfigFromFlags, func(_ string, _ string) (*restclient.Config, error) {
		return &restclient.Config{}, nil
	})
	defer monkey.UnpatchAll()
	monkey.Patch(kubernetes.NewForConfig, func(c *rest.Config) (*kubernetes.Clientset, error) {
		return nil, errors.New("NewForConfig err")
	})
	convey.Convey("TestInitKubernetesClientset", t, func() {
		err := InitKubernetesClientset("")
		convey.So(err.Error(), convey.ShouldEqual, "NewForConfig err")
	})
}

func TestGetClusterUUIDSucc(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()
	dbmock := mockdbaccessor.NewMockDbAccessor(controller)

	monkey.Patch(GetDataBase, func() dbaccessor.DbAccessor {
		return dbmock
	})
	defer monkey.UnpatchAll()
	key := "/knitter/cluster_uuid"
	dbmock.EXPECT().ReadLeaf(key).Return("1111", nil)

	convey.Convey("TestGetClusterUUIDSucc", t, func() {
		id := GetClusterUUID()
		convey.So(id, convey.ShouldEqual, "1111")
	})

}

func TestGetClusterUUIDErr(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()
	dbmock := mockdbaccessor.NewMockDbAccessor(controller)

	monkey.Patch(GetDataBase, func() dbaccessor.DbAccessor {
		return dbmock
	})
	defer monkey.UnpatchAll()
	key := "/knitter/cluster_uuid"
	dbmock.EXPECT().ReadLeaf(key).Return("1111", errors.New("etcd err"))

	convey.Convey("TestGetClusterUUIDErr", t, func() {
		id := GetClusterUUID()
		convey.So(id, convey.ShouldHaveSameTypeAs, "1111")
	})

}

func TestGetClusterUUIDNew(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()
	dbmock := mockdbaccessor.NewMockDbAccessor(controller)

	monkey.Patch(GetDataBase, func() dbaccessor.DbAccessor {
		return dbmock
	})
	defer monkey.UnpatchAll()
	key := "/knitter/cluster_uuid"
	dbmock.EXPECT().ReadLeaf(key).Return("1111", errors.New(errobj.EtcdKeyNotFound))
	dbmock.EXPECT().SaveLeaf(key, gomock.Any()).Return(nil)

	convey.Convey("TestGetClusterUUIDErr", t, func() {
		id := GetClusterUUID()
		convey.So(id, convey.ShouldHaveSameTypeAs, "1111")
	})

}

func TestGetSetClusterId(t *testing.T) {
	monkey.Patch(GetClusterUUID, func() string {
		return "111"
	})
	convey.Convey("TestGetSetClusterId", t, func() {
		SetClusterID()
		id := GetClusterID()
		convey.So(id, convey.ShouldEqual, "111")
	})
}

func TestGetClientset(t *testing.T) {
	clientSetExpect := &kubernetes.Clientset{}
	clientSet = clientSetExpect
	convey.Convey("TestGetClientset", t, func() {
		cs := GetClientset()
		convey.So(cs, convey.ShouldEqual, clientSetExpect)
	})
}
