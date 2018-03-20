package accessor

import (
	//	"github.com/rackspace/gophercloud"
	//	"github.com/ZTE/Knitter/pkg/klog"
	"testing"
	//	"github.com/coreos/etcd/Godeps/_workspace/src/github.com/golang/glog"
	//	. "github.com/smartystreets/goconvey/convey"
	//	"github.com/rackspace/gophercloud/openstack/networking/v2/ports"
	//	"runtime"
	//	"path/filepath"
	"fmt"
	"io/ioutil"
	"log"
	//	"os"
	//	"github.com/coreos/etcd/Godeps/_workspace/src/golang.org/x/net/context"
	"errors"
	"github.com/coreos/etcd/client"
	"github.com/smartystreets/goconvey/convey"
	//	"path"
	"github.com/antonholmquist/jason"
	//	"net/http"
	//	"github.com/rackspace/gophercloud/openstack/networking/v2/ports"
	. "github.com/ZTE/Knitter/knitter-agent/domain/cni"
	"os"
)

type TestEtcd struct {
	Etcd
}

func getServerInfo() ([]byte, error) {
	//currentPath, _ := os.Getwd()
	pwdPath := os.Getenv("PWD")
	serverconffilepath := pwdPath + "/" + "server_info.json"
	serverinfo, err := ioutil.ReadFile(serverconffilepath)
	if err != nil {
		log.Fatal(err)
		return nil, fmt.Errorf("getServerInfo:get server info error")
	}
	return serverinfo, nil
}

func getServerInfoErr() ([]byte, error) {
	//currentPath, _ := os.Getwd()
	//serverconffilepath :=  "/home/yzh/paasnw/cni-master/gopath/src/accessor/server_info_err.json"
	pwdPath := os.Getenv("PWD")
	serverconffilepath := pwdPath + "/" + "server_info_err.json"
	serverInfo, err := ioutil.ReadFile(serverconffilepath)
	if err != nil {
		log.Fatal(err)
		return nil, fmt.Errorf("getServerInfo:get server info error")
	}
	return serverInfo, nil
}

func TestGetEtcdServerUrlFromConfigureSuccess(t *testing.T) {
	var etcd Etcd
	serverinfo, _ := getServerInfo()

	etcdurl, _ := etcd.GetEtcdServerURLFromConfigure(serverinfo)

	convey.Convey("Subject:TestGetEtcdServerUrlFromConfigureSuccess\n", t, func() {
		convey.Convey("Result Data Should be http://127.0.0.1:2379", func() {
			convey.So(etcdurl, convey.ShouldEqual, "http://127.0.0.1:2379")
		})
	})
}

func TestGetEtcdServerUrlFromConfigureFailed(t *testing.T) {
	var etcd Etcd
	serverinfo, _ := getServerInfoErr()

	_, err := etcd.GetEtcdServerURLFromConfigure(serverinfo)

	convey.Convey("Subject:TestGetEtcdServerUrlFromConfigureFailed\n", t, func() {
		convey.Convey("Result Data Should be failed", func() {
			convey.So(err, convey.ShouldNotBeNil)
		})
	})
}

//func TestGetK8SPodUrlSuccess(t *testing.T){
//	cniparam := CniParam{
//		ContainerId:"d4cf28298c8a9414786b24cb0e84331dc08cf3d43d99b76868cd0f34d9df506c",
//		IfName:"eth0",
//		Netns:"/proc/1351/ns/net",
//		Path:"/opt/cni/bin:/opt/paasnwk8scni/bin",
//		Args:"K8S_POD_NAMESPACE=default;K8S_POD_NAME=rc-test-seqandanno-kxk59;K8S_POD_INFRA_CONTAINER_ID=d4cf28298c8a9414786b24cb0e84331dc08cf3d43d99b76868cd0f34d9df506c"}
//	var e Etcd
//
//	podurl, _ := e.CreateK8SPodUrl(cniparam.Args)
//
//	convey.Convey("Subject:TestGetK8SPodUrl\n", t, func(){
//		convey.Convey("Result Data Should be /registry/pods/default/rc-test-seqandanno-kxk59", func(){
//			convey.So(podurl, convey.ShouldEqual, "/registry/pods/default/rc-test-seqandanno-kxk59")
//		})
//	})
//}

//func TestGetK8SPodUrlFailed(t *testing.T){
//	cniparam := CniParam{
//		ContainerId:"d4cf28298c8a9414786b24cb0e84331dc08cf3d43d99b76868cd0f34d9df506c",
//		IfName:"eth0",
//		Netns:"/proc/1351/ns/net",
//		Path:"/opt/cni/bin:/opt/paasnwk8scni/bin",
//		Args:"K8S_POD_NAMESPACE=;K8S_POD_NAME=rc-test-seqandanno-kxk59;K8S_POD_INFRA_CONTAINER_ID=d4cf28298c8a9414786b24cb0e84331dc08cf3d43d99b76868cd0f34d9df506c"}
//	var e Etcd
//
//	_, err := e.CreateK8SPodUrl(cniparam.Args)
//
//	convey.Convey("Subject:TestGetK8SPodUrl\n", t, func(){
//		convey.Convey("Result Data Should be err", func(){
//			convey.So(err, convey.ShouldNotBeNil)
//		})
//	})
//}

func TestGetPodID(t *testing.T) {
	podjson, _ := jason.NewObjectFromBytes([]byte(valuestr))
	var etcd Etcd

	podid, _ := etcd.GetPodID(podjson)

	convey.Convey("Subject:TestGetPodID\n", t, func() {
		convey.Convey("Result Data Should be 077bcb03-673a-11d9-a185-fa163e22af82", func() {
			convey.So(podid, convey.ShouldEqual, "077bcb03-673a-11d9-a185-fa163e22af82")
		})
	})

}

func TestGetPodIDErr(t *testing.T) {
	podjson, _ := jason.NewObjectFromBytes([]byte(valuestrerr))
	var etcd Etcd

	_, err := etcd.GetPodID(podjson)

	convey.Convey("Subject:TestGetPodID\n", t, func() {
		convey.Convey("Result Data Should be null", func() {
			convey.So(err, convey.ShouldNotBeNil)
		})
	})
}

func TestGetPodName(t *testing.T) {
	podjson, _ := jason.NewObjectFromBytes([]byte(valuestr))
	var etcd Etcd

	podname, _ := etcd.GetPodName(podjson)

	convey.Convey("Subject:TestGetPodName\n", t, func() {
		convey.Convey("Result Data Should be rc-test-seqandanno-kxk59", func() {
			convey.So(podname, convey.ShouldEqual, "rc-test-seqandanno-kxk59")
		})
	})
}

func TestGetPodNameErr(t *testing.T) {
	podjson, _ := jason.NewObjectFromBytes([]byte(valuestrerr))
	var etcd Etcd

	_, err := etcd.GetPodName(podjson)

	convey.Convey("Subject:TestGetPodNameErr\n", t, func() {
		convey.Convey("Result Data Should be null", func() {
			convey.So(err, convey.ShouldNotBeNil)
		})
	})
}

func TestGetPodNS(t *testing.T) {
	podjson, _ := jason.NewObjectFromBytes([]byte(valuestr))
	var etcd Etcd

	podns, _ := etcd.GetPodNS(podjson)

	convey.Convey("Subject:TestGetPodNS\n", t, func() {
		convey.Convey("Result Data Should be default", func() {
			convey.So(podns, convey.ShouldEqual, "default")
		})
	})
}

func TestGetPodNSErr(t *testing.T) {
	podjson, _ := jason.NewObjectFromBytes([]byte(valuestrerr))
	var etcd Etcd

	_, err := etcd.GetPodNS(podjson)

	convey.Convey("Subject:TestGetPodNSErr\n", t, func() {
		convey.Convey("Result Data Should be null", func() {
			convey.So(err, convey.ShouldNotBeNil)
		})
	})
}

func (e *TestEtcd) set(key string, value string) error {
	if key != "" {
		return nil
	}
	return errors.New("testEtcd set() error")
}

//
//func TestStorePod2Etcd(t *testing.T){
//	podjson, _ := jason.NewObjectFromBytes([]byte(valuestr))
//	port := &ports.Port{
//		ID : "4f6725e2-cc48-422e-a675-b350ae767c7d",
//		NetworkID : "174b7b5d-ce8d-450d-8740-493f967b78c4",
//		Name : "control",
//		AdminStateUp : true,
//		Status : "DOWN",
//		MACAddress : "fa:16:3e:b7:89:c4",
//		FixedIPs : []ports.IP{
//			ports.IP{
//			    SubnetID:"137d6796-2c64-401c-b8f4-c51aa9bd5e63",
//			    IPAddress:"192.168.109.16"}},
//		TenantID : "a4d775c2b52147dd819d21e769d70a95"}
//	var etcd TestEtcd
//
//	err := etcd.storePod2Etcd(port, podjson)
//
//	Convey("Subject:TestStorePod2Etcd\n", t, func(){
//		Convey("Result Data Should be nil", func(){
//			So(err, ShouldBeNil)
//		})
//	})
//}
//
//func TestStorePod2EtcdErrPodInfo(t *testing.T){
//	podjson, _ := jason.NewObjectFromBytes([]byte(valuestrerr))
//	port := &ports.Port{
//		ID : "4f6725e2-cc48-422e-a675-b350ae767c7d",
//		NetworkID : "174b7b5d-ce8d-450d-8740-493f967b78c4",
//		Name : "control",
//		AdminStateUp : true,
//		Status : "DOWN",
//		MACAddress : "fa:16:3e:b7:89:c4",
//		FixedIPs : []ports.IP{
//			ports.IP{
//				SubnetID:"137d6796-2c64-401c-b8f4-c51aa9bd5e63",
//				IPAddress:"192.168.109.16"}},
//		TenantID : "a4d775c2b52147dd819d21e769d70a95"}
//	var etcd Etcd
//
//	err := etcd.storePod2Etcd(port, podjson)
//
//	Convey("Subject:TestStorePod2Etcd\n", t, func(){
//		Convey("Result Data Should be nil", func(){
//			So(err, ShouldNotBeNil)
//		})
//	})
//}
//
//func TestStorePod2EtcdErrPortInfo(t *testing.T){
//	podjson, _ := jason.NewObjectFromBytes([]byte(valuestrerr))
//	port := &ports.Port{
//		ID : "4f6725e2-cc48-422e-a675-b350ae767c7d",
//		NetworkID : "174b7b5d-ce8d-450d-8740-493f967b78c4",
//		Name : "control",
//		AdminStateUp : true,
//		Status : "DOWN",
//		MACAddress : "fa:16:3e:b7:89:c4",
//		FixedIPs : []ports.IP{
//			ports.IP{
//				SubnetID:"137d6796-2c64-401c-b8f4-c51aa9bd5e63",
//				IPAddress:"192.168.109.16"}},
//		TenantID : "erra4d775c2b52147dd819d21e769d70a95"}
//	var etcd Etcd
//
//	err := etcd.storePod2Etcd(port, podjson)
//
//	Convey("Subject:TestStorePod2Etcd\n", t, func(){
//		Convey("Result Data Should be nil", func(){
//			So(err, ShouldNotBeNil)
//		})
//	})
//}

//func TestPreparePortInfo(t *testing.T){
//	podjson, _ := jason.NewObjectFromBytes([]byte(valuestr))
//	port := &ports.Port{
//		ID : "4f6725e2-cc48-422e-a675-b350ae767c7d",
//		NetworkID : "174b7b5d-ce8d-450d-8740-493f967b78c4",
//		Name : "control",
//		AdminStateUp : true,
//		Status : "DOWN",
//		MACAddress : "fa:16:3e:b7:89:c4",
//		FixedIPs : []ports.IP{
//			ports.IP{
//				SubnetID:"137d6796-2c64-401c-b8f4-c51aa9bd5e63",
//				IPAddress:"192.168.109.16"}},
//		TenantID : "erra4d775c2b52147dd819d21e769d70a95"}
//	var etcd Etcd
//
//	_, err := etcd.PreparePortInfo(port, podjson)
//
//	convey.Convey("Subject:TestPreparePortInfo\n", t, func(){
//		convey.Convey("Result Data Should be not nil", func(){
//			convey.So(err, convey.ShouldBeNil)
//		})
//	})
//}

//func TestPreparePortInfoErr(t *testing.T){
//	podjson, _ := jason.NewObjectFromBytes([]byte(valuestrerr))
//	port := &ports.Port{
//		ID : "4f6725e2-cc48-422e-a675-b350ae767c7d",
//		NetworkID : "174b7b5d-ce8d-450d-8740-493f967b78c4",
//		Name : "control",
//		AdminStateUp : true,
//		Status : "DOWN",
//		MACAddress : "fa:16:3e:b7:89:c4",
//		FixedIPs : []ports.IP{
//			ports.IP{
//				SubnetID:"137d6796-2c64-401c-b8f4-c51aa9bd5e63",
//				IPAddress:"192.168.109.16"}},
//		TenantID : "erra4d775c2b52147dd819d21e769d70a95"}
//	var etcd Etcd
//
//	_, err := etcd.PreparePortInfo(port, podjson)
//
//	convey.Convey("Subject:TestPreparePortInfoErr\n", t, func(){
//		convey.Convey("Result Data Should be nil", func(){
//			convey.So(err, convey.ShouldNotBeNil)
//		})
//	})
//}

//func TestStorePort2Etcd(t *testing.T){
//	podjson, _ := jason.NewObjectFromBytes([]byte(valuestr))
//	port := &ports.Port{
//		ID : "4f6725e2-cc48-422e-a675-b350ae767c7d",
//		NetworkID : "174b7b5d-ce8d-450d-8740-493f967b78c4",
//		Name : "control",
//		AdminStateUp : true,
//		Status : "DOWN",
//		MACAddress : "fa:16:3e:b7:89:c4",
//		FixedIPs : []ports.IP{
//			ports.IP{
//				SubnetID:"137d6796-2c64-401c-b8f4-c51aa9bd5e63",
//				IPAddress:"192.168.109.16"}},
//		TenantID : "erra4d775c2b52147dd819d21e769d70a95"}
//	var etcd Etcd
//
//	err := etcd.storePort2Etcd(port, podjson)
//
//	Convey("Subject:TestStorePort2Etcd\n", t, func(){
//		Convey("Result Data Should be nil", func(){
//			So(err, ShouldBeNil)
//		})
//	})
//}

func (t *TestEtcd) createClient(cfg client.Config) (client.Client, error) {
	url := cfg.Endpoints[0]
	if url != "" {
		return nil, nil
	}
	return nil, errors.New("createClient:create etcdclient error")
}

func (t *TestEtcd) createAPI() (client.KeysAPI, error) {

	return nil, nil
}

func TestInit(t *testing.T) {
	cniparam := CniParam{
		ContainerID: "d4cf28298c8a9414786b24cb0e84331dc08cf3d43d99b76868cd0f34d9df506c",
		IfName:      "eth0",
		Netns:       "/proc/1351/ns/net",
		Path:        "/opt/cni/bin:/opt/paasnwk8scni/bin",
		Args:        "K8S_POD_NAMESPACE=default;K8S_POD_NAME=rc-test-seqandanno-kxk59;K8S_POD_INFRA_CONTAINER_ID=d4cf28298c8a9414786b24cb0e84331dc08cf3d43d99b76868cd0f34d9df506c"}
	serverinfo, _ := getServerInfo()
	var etcd TestEtcd

	err := etcd.Init(cniparam.Args, serverinfo)

	convey.Convey("Subject:TestInit\n", t, func() {
		convey.Convey("Result Data Should be nil", func() {
			convey.So(err, convey.ShouldBeNil)
		})
	})
}

func TestInitErrCniparam(t *testing.T) {
	cniparam := CniParam{
		ContainerID: "d4cf28298c8a9414786b24cb0e84331dc08cf3d43d99b76868cd0f34d9df506c",
		IfName:      "eth0",
		Netns:       "/proc/1351/ns/net",
		Path:        "/opt/cni/bin:/opt/paasnwk8scni/bin",
		Args:        "K8S_POD_NAMESPACE=;K8S_POD_NAME=rc-test-seqandanno-kxk59;K8S_POD_INFRA_CONTAINER_ID=d4cf28298c8a9414786b24cb0e84331dc08cf3d43d99b76868cd0f34d9df506c"}
	serverinfo, _ := getServerInfoErr()
	var etcd TestEtcd

	err := etcd.Init(cniparam.Args, serverinfo)

	convey.Convey("Subject:TestInitErrCniparam\n", t, func() {
		convey.Convey("Result Data Should be nil", func() {
			convey.So(err, convey.ShouldNotBeNil)
		})
	})
}

func TestInitErrServerinfo(t *testing.T) {
	cniparam := CniParam{
		ContainerID: "d4cf28298c8a9414786b24cb0e84331dc08cf3d43d99b76868cd0f34d9df506c",
		IfName:      "eth0",
		Netns:       "/proc/1351/ns/net",
		Path:        "/opt/cni/bin:/opt/paasnwk8scni/bin",
		Args:        "K8S_POD_NAMESPACE=default;K8S_POD_NAME=rc-test-seqandanno-kxk59;K8S_POD_INFRA_CONTAINER_ID=d4cf28298c8a9414786b24cb0e84331dc08cf3d43d99b76868cd0f34d9df506c"}
	serverinfo, _ := getServerInfoErr()
	var etcd TestEtcd

	err := etcd.Init(cniparam.Args, serverinfo)

	convey.Convey("Subject:TestInitErrServerinfo\n", t, func() {
		convey.Convey("Result Data Should be nil", func() {
			convey.So(err, convey.ShouldNotBeNil)
		})
	})
}

var valuestr = "{\"kind\":\"Pod\",\"apiVersion\":\"v1\",\"metadata\":{\"name\":\"rc-test-seqandanno-kxk59\",\"generateName\":\"rc-test-seqandanno-\",\"namespace\":\"default\",\"selfLink\":\"/api/v1/namespaces/default/pods/rc-test-seqandanno-kxk59\",\"uid\":\"077bcb03-673a-11d9-a185-fa163e22af82\",\"resourceVersion\":\"127157\",\"creationTimestamp\":\"2005-01-15T21:11:31Z\",\"labels\":{\"name\":\"80-nginx\"},\"annotations\":{\"control\":\"174b7b5d-ce8d-450d-8740-493f967b78c4\",\"kubernetes.io/created-by\":\"{\\\"kind\\\":\\\"SerializedReference\\\",\\\"apiVersion\\\":\\\"v1\\\",\\\"reference\\\":{\\\"kind\\\":\\\"ReplicationController\\\",\\\"namespace\\\":\\\"default\\\",\\\"name\\\":\\\"rc-test-seqandanno\\\",\\\"uid\\\":\\\"07741492-673a-11d9-a185-fa163e22af82\\\",\\\"apiVersion\\\":\\\"v1\\\",\\\"resourceVersion\\\":\\\"127127\\\"}}\\n\",\"manager\":\"f454745a-0e1d-4841-8ee8-4cf49bd102f6\",\"media\":\"75878d12-a1a4-4ee7-a83e-d0796e05683a\",\"network-plane\":\"control,media,manager\"}},\"spec\":{\"containers\":[{\"name\":\"nginx-con\",\"image\":\"nginx\",\"resources\":{},\"terminationMessagePath\":\"/dev/termination-log\",\"imagePullPolicy\":\"IfNotPresent\"}],\"restartPolicy\":\"Always\",\"terminationGracePeriodSeconds\":30,\"dnsPolicy\":\"ClusterFirst\",\"nodeName\":\"192.168.77.5\"},\"status\":{\"phase\":\"Running\",\"conditions\":[{\"type\":\"Ready\",\"status\":\"True\",\"lastProbeTime\":null,\"lastTransitionTime\":null}],\"hostIP\":\"192.168.77.5\",\"podIP\":\"192.168.106.33\",\"startTime\":\"2005-01-15T21:08:59Z\",\"containerStatuses\":[{\"name\":\"nginx-con\",\"state\":{\"running\":{\"startedAt\":\"2005-01-15T21:09:21Z\"}},\"lastState\":{},\"ready\":true,\"restartCount\":0,\"image\":\"nginx\",\"imageID\":\"docker://6e36f46089ed3c0326d2f56d6282af5eab6000caaa04e44f327c37f13d13c933\",\"containerID\":\"docker://2f0a8fe69fb4562b18ef12558f464f4a02e9e8a6a288454c1654478dde2dff66\"}]}}\n"
var valuestrerr = "{\"kind\":\"Pod\",\"apiVersion\":\"v1\",\"metadata\":{\"generateName\":\"rc-test-seqandanno-\",\"selfLink\":\"/api/v1/namespaces/default/pods/rc-test-seqandanno-kxk59\",\"resourceVersion\":\"127157\",\"creationTimestamp\":\"2005-01-15T21:11:31Z\",\"labels\":{\"name\":\"80-nginx\"},\"annotations\":{\"control\":\"174b7b5d-ce8d-450d-8740-493f967b78c4\",\"kubernetes.io/created-by\":\"{\\\"kind\\\":\\\"SerializedReference\\\",\\\"apiVersion\\\":\\\"v1\\\",\\\"reference\\\":{\\\"kind\\\":\\\"ReplicationController\\\",\\\"namespace\\\":\\\"default\\\",\\\"name\\\":\\\"rc-test-seqandanno\\\",\\\"apiVersion\\\":\\\"v1\\\",\\\"resourceVersion\\\":\\\"127127\\\"}}\\n\",\"manager\":\"f454745a-0e1d-4841-8ee8-4cf49bd102f6\",\"media\":\"75878d12-a1a4-4ee7-a83e-d0796e05683a\",\"network-plane\":\"control,media,manager\"}},\"spec\":{\"containers\":[{\"name\":\"nginx-con\",\"image\":\"nginx\",\"resources\":{},\"terminationMessagePath\":\"/dev/termination-log\",\"imagePullPolicy\":\"IfNotPresent\"}],\"restartPolicy\":\"Always\",\"terminationGracePeriodSeconds\":30,\"dnsPolicy\":\"ClusterFirst\",\"nodeName\":\"192.168.77.5\"},\"status\":{\"phase\":\"Running\",\"conditions\":[{\"type\":\"Ready\",\"status\":\"True\",\"lastProbeTime\":null,\"lastTransitionTime\":null}],\"hostIP\":\"192.168.77.5\",\"podIP\":\"192.168.106.33\",\"startTime\":\"2005-01-15T21:08:59Z\",\"containerStatuses\":[{\"name\":\"nginx-con\",\"state\":{\"running\":{\"startedAt\":\"2005-01-15T21:09:21Z\"}},\"lastState\":{},\"ready\":true,\"restartCount\":0,\"image\":\"nginx\",\"imageID\":\"docker://6e36f46089ed3c0326d2f56d6282af5eab6000caaa04e44f327c37f13d13c933\",\"containerID\":\"docker://2f0a8fe69fb4562b18ef12558f464f4a02e9e8a6a288454c1654478dde2dff66\"}]}}\n"
