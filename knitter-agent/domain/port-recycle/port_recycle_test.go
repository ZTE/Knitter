/*
Copyright 2018 ZTE Corporation. All rights reserved.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package portrecycle

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ZTE/Knitter/knitter-agent/domain/adapter"
	"github.com/ZTE/Knitter/knitter-agent/domain/cni"
	"github.com/ZTE/Knitter/knitter-agent/err-obj"
	"github.com/ZTE/Knitter/knitter-agent/infra/os-encap"
	"github.com/ZTE/Knitter/pkg/iaas-accessor"
	"github.com/antonholmquist/jason"
	"github.com/coreos/etcd/client"
	"github.com/golang/gostub"
	. "github.com/golang/gostub"
	"github.com/smartystreets/goconvey/convey"
	. "github.com/smartystreets/goconvey/convey"
	"strings"
	"testing"
	"time"
)

//GetPodsFromDb
func TestGetPodsFromDbForSucc(t *testing.T) {
	etcdPods := []*client.Node{
		{Key: "/paasnet/runtime/clusters/cluster1/nodes/node1/pods/ns1"},
		{Key: "/paasnet/runtime/clusters/cluster1/nodes/node1/pods/ns2"},
		{Key: "/paasnet/runtime/clusters/cluster1/nodes/node1/pods/ns3"},
	}
	etcdPodsNs1 := []*client.Node{
		{Key: "/paasnet/runtime/clusters/cluster1/nodes/node1/pods/ns1/name1"},
	}
	etcdPodsNs2 := []*client.Node{
		{Key: "/paasnet/runtime/clusters/cluster1/nodes/node1/pods/ns2/name2"},
	}
	etcdPodsNs3 := []*client.Node{
		{Key: "/paasnet/runtime/clusters/cluster1/nodes/node1/pods/ns3/name3"},
	}
	outputs := []Output{
		{StubVals: Values{etcdPods, nil}},
		{StubVals: Values{etcdPodsNs1, nil}},
		{StubVals: Values{etcdPodsNs2, nil}},
		{StubVals: Values{etcdPodsNs3, nil}},
	}

	stubs := StubFuncSeq(&adapter.ReadDirFromDb, outputs)
	defer stubs.Reset()
	stubs.StubFunc(&cni.GetGlobalContext,
		&cni.AgentContext{ClusterID: "cluster1", HostIP: "node1"})

	convey.Convey("TestGetPodsFromDbForSucc\n", t, func() {
		pods, err := GetPodsFromDb()
		fmt.Printf("pods:%v\n", pods)
		convey.So(err, convey.ShouldEqual, nil)
		convey.So(len(pods), convey.ShouldEqual, 3)

		convey.So(pods[0].PodNs, convey.ShouldEqual, "ns1")
		convey.So(pods[0].PodName, convey.ShouldEqual, "name1")

		convey.So(pods[1].PodNs, convey.ShouldEqual, "ns2")
		convey.So(pods[1].PodName, convey.ShouldEqual, "name2")

		convey.So(pods[2].PodNs, convey.ShouldEqual, "ns3")
		convey.So(pods[2].PodName, convey.ShouldEqual, "name3")
	})
}

func TestGetPodsFromDbForErrInBranch1(t *testing.T) {
	etcdPods := []*client.Node{
		{Key: "/paasnet/runtime/clusters/cluster1/nodes/node1/pods/ns1"},
		{Key: "/paasnet/runtime/clusters/cluster1/nodes/node1/pods/ns2"},
		{Key: "/paasnet/runtime/clusters/cluster1/nodes/node1/pods/ns3"},
	}

	stubs := StubFunc(&adapter.ReadDirFromDb, etcdPods, errobj.ErrDbKeyNotFound)
	defer stubs.Reset()
	stubs.StubFunc(&cni.GetGlobalContext,
		&cni.AgentContext{ClusterID: "cluster1", HostIP: "node1"})

	convey.Convey("TestGetPodsFromDbForErrInBranch1\n", t, func() {
		pods, err := GetPodsFromDb()
		fmt.Printf("pods:%v\n", pods)
		fmt.Printf("err:%v\n", err)
		convey.So(errobj.IsEqual(err, errobj.ErrDbKeyNotFound),
			convey.ShouldBeTrue)
		convey.So(len(pods), convey.ShouldEqual, 0)
	})
}

func TestGetPodsFromDbForErrInBranch2(t *testing.T) {
	etcdPods := []*client.Node{
		{Key: "/paasnet/runtime/clusters/cluster1/nodes/node1/pods/ns1"},
		{Key: "/paasnet/runtime/clusters/cluster1/nodes/node1/pods/ns2"},
		{Key: "/paasnet/runtime/clusters/cluster1/nodes/node1/pods/ns3"},
	}
	etcdPodsNs1 := []*client.Node{
		{Key: "/paasnet/runtime/clusters/cluster1/nodes/node1/pods/ns1/name1"},
	}
	etcdPodsNs2 := []*client.Node{
		{Key: "/paasnet/runtime/clusters/cluster1/nodes/node1/pods/ns2/name2"},
	}
	etcdPodsNs3 := []*client.Node{
		{Key: "/paasnet/runtime/clusters/cluster1/nodes/node1/pods/ns3/name3"},
	}
	outputs := []Output{
		{StubVals: Values{etcdPods, nil}},
		{StubVals: Values{etcdPodsNs1, errobj.ErrDbKeyNotFound}},
		{StubVals: Values{etcdPodsNs2, errobj.ErrDbKeyNotFound}},
		{StubVals: Values{etcdPodsNs3, errobj.ErrDbKeyNotFound}},
	}

	stubs := StubFuncSeq(&adapter.ReadDirFromDb, outputs)
	defer stubs.Reset()
	stubs.StubFunc(&cni.GetGlobalContext,
		&cni.AgentContext{ClusterID: "cluster1", HostIP: "node1"})

	stubs.StubFunc(&adapter.ClearDirFromDb, nil)
	convey.Convey("TestGetPodsFromDbForErrInBranch2\n", t, func() {
		pods, err := GetPodsFromDb()
		fmt.Printf("pods:%v\n", pods)
		convey.So(errobj.IsEqual(errobj.ErrNoPodCurrentNodeDbReliable, err),
			convey.ShouldBeTrue)
	})
}

func TestGetPodsFromDbForErrInBranch3(t *testing.T) {
	etcdPods := []*client.Node{
		{Key: "/paasnet/runtime/clusters/cluster1/nodes/node1/pods/ns1"},
		{Key: "/paasnet/runtime/clusters/cluster1/nodes/node1/pods/ns2"},
		{Key: "/paasnet/runtime/clusters/cluster1/nodes/node1/pods/ns3"},
	}
	etcdPodsNs1 := []*client.Node{
		{Key: "/paasnet/runtime/clusters/cluster1/nodes/node1/pods/ns1/name1"},
	}
	etcdPodsNs2 := []*client.Node{
		{Key: "/paasnet/runtime/clusters/cluster1/nodes/node1/pods/ns2/name2"},
	}
	etcdPodsNs3 := []*client.Node{
		{Key: "/paasnet/runtime/clusters/cluster1/nodes/node1/pods/ns3/name3"},
	}
	outputs := []Output{
		{StubVals: Values{etcdPods, nil}},
		{StubVals: Values{etcdPodsNs1, errobj.ErrDbKeyNotFound}},
		{StubVals: Values{etcdPodsNs2, errobj.ErrDbKeyNotFound}},
		{StubVals: Values{etcdPodsNs3, errobj.ErrDbConnRefused}},
	}

	stubs := StubFuncSeq(&adapter.ReadDirFromDb, outputs)
	defer stubs.Reset()
	stubs.StubFunc(&cni.GetGlobalContext,
		&cni.AgentContext{ClusterID: "cluster1", HostIP: "node1"})

	stubs.StubFunc(&adapter.ClearDirFromDb, nil)
	convey.Convey("TestGetPodsFromDbForErrInBranch3\n", t, func() {
		pods, err := GetPodsFromDb()
		fmt.Printf("pods:%v\n", pods)
		convey.So(errobj.IsEqual(errobj.ErrNoPodCurrentNodeDbNtReliable, err),
			convey.ShouldBeTrue)
	})
}

//AnalyzePods
type RspPodMetadataOfK8s struct {
	Metadata RspPodOfK8s `json:"metadata"`
}
type RspPodOfK8s struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
}

func TestAnalyzePodsSucc(t *testing.T) {
	podList := []*jason.Object{}
	pod1 := RspPodMetadataOfK8s{Metadata: RspPodOfK8s{Name: "name1", Namespace: "ns1"}}
	pod2 := RspPodMetadataOfK8s{Metadata: RspPodOfK8s{Name: "name2", Namespace: "ns2"}}

	rsp1Byte, _ := json.Marshal(pod1)
	pod1Json, _ := jason.NewObjectFromBytes(rsp1Byte)
	rsp2Byte, _ := json.Marshal(pod2)
	pod2Json, _ := jason.NewObjectFromBytes(rsp2Byte)
	podList = append(podList, pod1Json)
	podList = append(podList, pod2Json)

	convey.Convey("TestAnalyzePodsSucc\n", t, func() {
		pods := AnalyzePods(podList)
		fmt.Printf("AnalyzePods:%v\n", pods)
		convey.So(pods, convey.ShouldNotBeNil)
		convey.So(pods[0].PodNs, convey.ShouldEqual, "ns1")
		convey.So(pods[0].PodName, convey.ShouldEqual, "name1")
		convey.So(pods[1].PodNs, convey.ShouldEqual, "ns2")
		convey.So(pods[1].PodName, convey.ShouldEqual, "name2")
	})
}

func TestAnalyzePodsErrInBranch1(t *testing.T) {
	var podList []*jason.Object

	convey.Convey("TestAnalyzePodsErrInBranch1\n", t, func() {
		pods := AnalyzePods(podList)
		fmt.Printf("pods:%v\n", pods)
		convey.So(pods, convey.ShouldBeNil)
	})
}

func TestAnalyzePodsErrInBranch2(t *testing.T) {
	podList := []*jason.Object{}
	pod1 := RspPodMetadataOfK8s{Metadata: RspPodOfK8s{Namespace: "ns1"}}
	pod2 := RspPodMetadataOfK8s{Metadata: RspPodOfK8s{Name: "name2"}}

	rsp1Byte, _ := json.Marshal(pod1)
	fmt.Printf("rsp1Byte:%v\n", string(rsp1Byte))
	pod1Json, _ := jason.NewObjectFromBytes(rsp1Byte)
	fmt.Printf("pod1Json:%v\n", pod1Json)
	rsp2Byte, _ := json.Marshal(pod2)
	fmt.Printf("rsp2Byte:%v\n", string(rsp2Byte))
	pod2Json, _ := jason.NewObjectFromBytes(rsp2Byte)
	fmt.Printf("pod2Json:%v\n", pod2Json)
	podList = append(podList, pod1Json)
	podList = append(podList, pod2Json)

	convey.Convey("TestAnalyzePodsErrInBranch2\n", t, func() {
		pods := AnalyzePods(podList)
		fmt.Printf("pods:%v\n", pods)
		convey.So(pods, convey.ShouldNotBeNil)
		convey.So(pods[0].PodNs, convey.ShouldEqual, "ns1")
		convey.So(pods[0].PodName, convey.ShouldEqual, "")
		convey.So(pods[1].PodNs, convey.ShouldEqual, "")
		convey.So(pods[1].PodName, convey.ShouldEqual, "name2")
	})
}

//GetPodsFromK8s
func TestGetPodsFromK8sForSucc(t *testing.T) {
	podList := []*jason.Object{}
	pod1 := RspPodMetadataOfK8s{Metadata: RspPodOfK8s{Name: "name1", Namespace: "ns1"}}
	pod2 := RspPodMetadataOfK8s{Metadata: RspPodOfK8s{Name: "name2", Namespace: "ns2"}}

	rsp1Byte, _ := json.Marshal(pod1)
	pod1Json, _ := jason.NewObjectFromBytes(rsp1Byte)
	rsp2Byte, _ := json.Marshal(pod2)
	pod2Json, _ := jason.NewObjectFromBytes(rsp2Byte)
	podList = append(podList, pod1Json)
	podList = append(podList, pod2Json)

	stubs := StubFunc(&adapter.GetPodsByNodeID, podList, nil)
	defer stubs.Reset()
	stubs.StubFunc(&cni.GetGlobalContext,
		&cni.AgentContext{ClusterID: "cluster1", HostIP: "node1"})

	convey.Convey("TestGetPodsFromK8sForSucc\n", t, func() {
		pods, err := GetPodsFromK8s()
		fmt.Printf("pods:%v\n", pods)
		convey.So(err, convey.ShouldEqual, nil)
		convey.So(len(pods), convey.ShouldEqual, 2)

		convey.So(pods[0].PodNs, convey.ShouldEqual, "ns1")
		convey.So(pods[0].PodName, convey.ShouldEqual, "name1")

		convey.So(pods[1].PodNs, convey.ShouldEqual, "ns2")
		convey.So(pods[1].PodName, convey.ShouldEqual, "name2")
	})
}

func TestGetPodsFromK8sForErrInBranch1(t *testing.T) {
	podList := []*jason.Object{}
	pod1 := RspPodMetadataOfK8s{Metadata: RspPodOfK8s{Name: "name1", Namespace: "ns1"}}
	pod2 := RspPodMetadataOfK8s{Metadata: RspPodOfK8s{Name: "name2", Namespace: "ns2"}}

	rsp1Byte, _ := json.Marshal(pod1)
	pod1Json, _ := jason.NewObjectFromBytes(rsp1Byte)
	rsp2Byte, _ := json.Marshal(pod2)
	pod2Json, _ := jason.NewObjectFromBytes(rsp2Byte)
	podList = append(podList, pod1Json)
	podList = append(podList, pod2Json)

	stubs := StubFunc(&adapter.GetPodsByNodeID, podList, errobj.ErrK8sGetPodByNodeID)
	defer stubs.Reset()
	stubs.StubFunc(&cni.GetGlobalContext,
		&cni.AgentContext{ClusterID: "cluster1", HostIP: "node1"})

	convey.Convey("TestGetPodsFromK8sForErrInBranch1\n", t, func() {
		pods, err := GetPodsFromK8s()
		fmt.Printf("pods:%v\n", pods)
		convey.So(errobj.IsEqual(err, errobj.ErrK8sGetPodByNodeID), convey.ShouldBeTrue)
	})
}

//GetPodsFromOse

//CollectNoUsedPod
func TestCollectNoUsedPodSuccInBranchK8s(t *testing.T) {
	etcdPods := []*client.Node{
		{Key: "/paasnet/runtime/clusters/cluster1/nodes/node1/pods/ns1"},
		{Key: "/paasnet/runtime/clusters/cluster1/nodes/node1/pods/ns2"},
		{Key: "/paasnet/runtime/clusters/cluster1/nodes/node1/pods/ns3"},
	}
	etcdPodsNs1 := []*client.Node{
		{Key: "/paasnet/runtime/clusters/cluster1/nodes/node1/pods/ns1/name1"},
	}
	etcdPodsNs2 := []*client.Node{
		{Key: "/paasnet/runtime/clusters/cluster1/nodes/node1/pods/ns2/name2"},
	}
	etcdPodsNs3 := []*client.Node{
		{Key: "/paasnet/runtime/clusters/cluster1/nodes/node1/pods/ns3/name3"},
	}
	outputs := []Output{
		{StubVals: Values{etcdPods, nil}},
		{StubVals: Values{etcdPodsNs1, nil}},
		{StubVals: Values{etcdPodsNs2, nil}},
		{StubVals: Values{etcdPodsNs3, nil}},
	}

	stubs := StubFuncSeq(&adapter.ReadDirFromDb, outputs)
	defer stubs.Reset()
	stubs.StubFunc(&cni.GetGlobalContext,
		&cni.AgentContext{ClusterID: "cluster1", HostIP: "node1", ClusterType: "k8s"})

	podList := []*jason.Object{}
	pod2 := RspPodMetadataOfK8s{Metadata: RspPodOfK8s{Name: "name2", Namespace: "ns2"}}
	pod3 := RspPodMetadataOfK8s{Metadata: RspPodOfK8s{Name: "name3", Namespace: "ns3"}}
	rsp2Byte, _ := json.Marshal(pod2)
	pod2Json, _ := jason.NewObjectFromBytes(rsp2Byte)
	rsp3Byte, _ := json.Marshal(pod3)
	pod3Json, _ := jason.NewObjectFromBytes(rsp3Byte)
	podList = append(podList, pod2Json)
	podList = append(podList, pod3Json)
	stubs.StubFunc(&adapter.GetPodsByNodeID, podList, nil)

	convey.Convey("TestCollectNoUsedPodSuccInBranchK8s\n", t, func() {
		noUsedPods, err := CollectNoUsedPod()
		fmt.Printf("pods:%v\n", noUsedPods)
		convey.So(err, convey.ShouldEqual, nil)
		convey.So(len(noUsedPods), convey.ShouldEqual, 1)

		convey.So(noUsedPods[0].PodNs, convey.ShouldEqual, "ns1")
		convey.So(noUsedPods[0].PodName, convey.ShouldEqual, "name1")
	})
}

func TestCollectNoUsedPodErrInBranch1(t *testing.T) {
	etcdPods := []*client.Node{
		{Key: "/paasnet/runtime/clusters/cluster1/nodes/node1/pods/ns1"},
		{Key: "/paasnet/runtime/clusters/cluster1/nodes/node1/pods/ns2"},
		{Key: "/paasnet/runtime/clusters/cluster1/nodes/node1/pods/ns3"},
	}

	stubs := StubFunc(&adapter.ReadDirFromDb, etcdPods, errobj.ErrDbKeyNotFound)
	defer stubs.Reset()
	stubs.StubFunc(&cni.GetGlobalContext,
		&cni.AgentContext{ClusterID: "cluster1", HostIP: "node1"})

	convey.Convey("TestCollectNoUsedPodErrInBranch1\n", t, func() {
		noUsedPods, err := CollectNoUsedPod()
		fmt.Printf("pods:%v\n", noUsedPods)
		convey.So(errobj.IsEqual(err, errobj.ErrDbKeyNotFound), convey.ShouldBeTrue)
		convey.So(noUsedPods, convey.ShouldBeNil)
	})

}

func TestCollectNoUsedPodErrInBranch2OfK8s(t *testing.T) {
	etcdPods := []*client.Node{
		{Key: "/paasnet/runtime/clusters/cluster1/nodes/node1/pods/ns1"},
		{Key: "/paasnet/runtime/clusters/cluster1/nodes/node1/pods/ns2"},
		{Key: "/paasnet/runtime/clusters/cluster1/nodes/node1/pods/ns3"},
	}
	etcdPodsNs1 := []*client.Node{
		{Key: "/paasnet/runtime/clusters/cluster1/nodes/node1/pods/ns1/name1"},
	}
	etcdPodsNs2 := []*client.Node{
		{Key: "/paasnet/runtime/clusters/cluster1/nodes/node1/pods/ns2/name2"},
	}
	etcdPodsNs3 := []*client.Node{
		{Key: "/paasnet/runtime/clusters/cluster1/nodes/node1/pods/ns3/name3"},
	}
	outputs := []Output{
		{StubVals: Values{etcdPods, nil}},
		{StubVals: Values{etcdPodsNs1, nil}},
		{StubVals: Values{etcdPodsNs2, nil}},
		{StubVals: Values{etcdPodsNs3, nil}},
	}

	stubs := StubFuncSeq(&adapter.ReadDirFromDb, outputs)
	defer stubs.Reset()
	stubs.StubFunc(&cni.GetGlobalContext,
		&cni.AgentContext{ClusterID: "cluster1", HostIP: "node1", ClusterType: "k8s"})

	podList := []*jason.Object{}
	pod2 := RspPodMetadataOfK8s{Metadata: RspPodOfK8s{Name: "name2", Namespace: "ns2"}}
	pod3 := RspPodMetadataOfK8s{Metadata: RspPodOfK8s{Name: "name3", Namespace: "ns3"}}
	rsp2Byte, _ := json.Marshal(pod2)
	pod2Json, _ := jason.NewObjectFromBytes(rsp2Byte)
	rsp3Byte, _ := json.Marshal(pod3)
	pod3Json, _ := jason.NewObjectFromBytes(rsp3Byte)
	podList = append(podList, pod2Json)
	podList = append(podList, pod3Json)
	stubs.StubFunc(&adapter.GetPodsByNodeID, podList, errobj.ErrK8sGetPodByNodeID)

	convey.Convey("TestCollectNoUsedPodErrInBranch2OfK8s\n", t, func() {
		noUsedPods, err := CollectNoUsedPod()
		fmt.Printf("pods:%v\n", noUsedPods)
		convey.So(errobj.IsEqual(err, errobj.ErrK8sGetPodByNodeID), convey.ShouldBeTrue)
		convey.So(noUsedPods, convey.ShouldBeNil)
	})

}

//GetKeysOfAllPortsInPodDir
func TestGetKeysOfAllPortsInPodDirSucc(t *testing.T) {
	rspInterfaceLeafOfPod := []*client.Node{
		{
			Key:   "/paasnet/tenants/ns1/pods/ns1/name1/interfaces/if1",
			Value: "/paasnet/tenant/ns1/interfaces/portId1ns1name1/self"},
		{
			Key:   "/paasnet/tenants/ns1/pods/ns1/name1/interfaces/if2",
			Value: "/paasnet/tenant/ns1/interfaces/portId2ns1name1/self"},
		{
			Key:   "/paasnet/tenants/ns1/pods/ns1/name1/interfaces/if3",
			Value: "/paasnet/tenant/ns1/interfaces/portId3ns1name1/self"},
	}

	outputs := []Output{
		{StubVals: Values{rspInterfaceLeafOfPod, nil}},
	}

	stubs := StubFuncSeq(&adapter.ReadDirFromDb, outputs)
	defer stubs.Reset()
	stubs.StubFunc(&cni.GetGlobalContext,
		&cni.AgentContext{ClusterID: "cluster1", HostIP: "node1"})

	convey.Convey("TestGetKeysOfAllPortsInPodDirSucc\n", t, func() {
		ports, err := GetKeysOfAllPortsInPodDir("ns1", "name1")
		fmt.Printf("ports:%v\n", ports)
		convey.So(err, convey.ShouldBeNil)
		convey.So(len(ports), convey.ShouldEqual, 3)

		convey.So(ports[0].Value, convey.ShouldEqual, "/paasnet/tenant/ns1/interfaces/portId1ns1name1/self")
		convey.So(ports[1].Value, convey.ShouldEqual, "/paasnet/tenant/ns1/interfaces/portId2ns1name1/self")
		convey.So(ports[2].Value, convey.ShouldEqual, "/paasnet/tenant/ns1/interfaces/portId3ns1name1/self")
	})
}

func TestGetKeysOfAllPortsInPodDirErr(t *testing.T) {
	rspInterfaceLeafOfPod := []*client.Node{
		{
			Key:   "/paasnet/tenants/ns1/pods/ns1/name1/interfaces/if1",
			Value: "/paasnet/tenant/ns1/interfaces/portId1ns1name1/self"},
		{
			Key:   "/paasnet/tenants/ns1/pods/ns1/name1/interfaces/if2",
			Value: "/paasnet/tenant/ns1/interfaces/portId2ns1name1/self"},
		{
			Key:   "/paasnet/tenants/ns1/pods/ns1/name1/interfaces/if3",
			Value: "/paasnet/tenant/ns1/interfaces/portId3ns1name1/self"},
	}

	outputs := []Output{
		{StubVals: Values{rspInterfaceLeafOfPod, errobj.ErrDbKeyNotFound}},
	}

	stubs := StubFuncSeq(&adapter.ReadDirFromDb, outputs)
	defer stubs.Reset()
	stubs.StubFunc(&cni.GetGlobalContext,
		&cni.AgentContext{ClusterID: "cluster1", HostIP: "node1"})

	convey.Convey("TestGetKeysOfAllPortsInPodDirErr\n", t, func() {
		ports, err := GetKeysOfAllPortsInPodDir("ns1", "name1")
		fmt.Printf("ports:%v\n", ports)
		convey.So(errobj.IsEqual(err, errobj.ErrDbKeyNotFound), convey.ShouldBeTrue)
		convey.So(ports, convey.ShouldBeNil)
	})
}

//GetPortInfo
func TestGetPortInfoSucc(t *testing.T) {
	rspInterfaceLeafOfPod := []*client.Node{
		{
			Key:   "/paasnet/tenants/ns1/pods/ns1/name1/interfaces/if1",
			Value: "/paasnet/tenant/ns1/interfaces/portId1ns1name1/self"},
		{
			Key:   "/paasnet/tenants/ns1/pods/ns1/name1/interfaces/if2",
			Value: "/paasnet/tenant/ns1/interfaces/portId2ns1name1/self"},
		{
			Key:   "/paasnet/tenants/ns1/pods/ns1/name1/interfaces/if3",
			Value: "/paasnet/tenant/ns1/interfaces/portId3ns1name1/self"},
	}

	outputs := []Output{
		{StubVals: Values{rspInterfaceLeafOfPod, nil}},
	}

	stubs := StubFuncSeq(&adapter.ReadDirFromDb, outputs)
	defer stubs.Reset()
	stubs.StubFunc(&cni.GetGlobalContext,
		&cni.AgentContext{ClusterID: "cluster1", HostIP: "node1"})

	rspInterfaceInfo1 := "{\"name\":\"eth0\",\"status\":\"\",\"port_id\":\"portId1\",\"ip\":\"192.168.183.237\"," +
		"\"mac_address\":\"fa:16:3e:3e:ed:69\",\"network_id\":\"4cfd6854-ade7-4e1b-b68d-163b5b4002d5\"," +
		"\"subnet_id\":\"723c34d7-fb8b-4fcd-a3d5-0d06be3f4ecc\",\"device_id\":\"bc1b77a5-1f31-11e7-a72b-fa163e05987b\"," +
		"\"owner_type\":\"pod\",\"port_type\":\"nodpdk\",\"bus_info\":\"\",\"net_plane_type\":\"std\"," +
		"\"net_plane_name\":\"net_api\",\"tenant_id\":\"admin\",\"nic_type\":\"normal\",\"pod_name\":\"name1\"," +
		"\"pod_ns\":\"ns1\",\"accelerate\":\"false\"}"
	rspInterfaceInfo2 := "{\"name\":\"eth0\",\"status\":\"\",\"port_id\":\"portId2\",\"ip\":\"192.168.183.238\"," +
		"\"mac_address\":\"fa:16:3e:3e:ed:69\",\"network_id\":\"4cfd6854-ade7-4e1b-b68d-163b5b4002d5\"," +
		"\"subnet_id\":\"723c34d7-fb8b-4fcd-a3d5-0d06be3f4ecc\",\"device_id\":\"bc1b77a5-1f31-11e7-a72b-fa163e05987b\"," +
		"\"owner_type\":\"pod\",\"port_type\":\"nodpdk\",\"bus_info\":\"\",\"net_plane_type\":\"std\"," +
		"\"net_plane_name\":\"net_api\",\"tenant_id\":\"admin\",\"nic_type\":\"normal\",\"pod_name\":\"name1\"," +
		"\"pod_ns\":\"ns1\",\"accelerate\":\"false\"}"
	rspInterfaceInfo3 := "{\"name\":\"eth0\",\"status\":\"\",\"port_id\":\"portId3\",\"ip\":\"192.168.183.239\"," +
		"\"mac_address\":\"fa:16:3e:3e:ed:69\",\"network_id\":\"4cfd6854-ade7-4e1b-b68d-163b5b4002d5\"," +
		"\"subnet_id\":\"723c34d7-fb8b-4fcd-a3d5-0d06be3f4ecc\",\"device_id\":\"bc1b77a5-1f31-11e7-a72b-fa163e05987b\"," +
		"\"owner_type\":\"pod\",\"port_type\":\"nodpdk\",\"bus_info\":\"\",\"net_plane_type\":\"std\"," +
		"\"net_plane_name\":\"net_api\",\"tenant_id\":\"admin\",\"nic_type\":\"normal\",\"pod_name\":\"name1\"," +
		"\"pod_ns\":\"ns1\",\"accelerate\":\"false\"}"
	outputs1 := []Output{
		{StubVals: Values{rspInterfaceInfo1, nil}},
		{StubVals: Values{rspInterfaceInfo2, nil}},
		{StubVals: Values{rspInterfaceInfo3, nil}},
	}

	stubs.StubFuncSeq(&adapter.ReadLeafFromDb, outputs1)
	ports, _ := GetKeysOfAllPortsInPodDir("ns1", "name1")

	convey.Convey("TestGetPortInfoSucc\n", t, func() {
		portsInfo, _ := GetPortInfo(ports)
		fmt.Printf("ports:%v\n", portsInfo)
		convey.So(len(portsInfo), convey.ShouldEqual, 3)

		convey.So(portsInfo[0].PodNs, convey.ShouldEqual, "ns1")
		convey.So(portsInfo[0].PodName, convey.ShouldEqual, "name1")

		convey.So(portsInfo[1].PodNs, convey.ShouldEqual, "ns1")
		convey.So(portsInfo[1].PodName, convey.ShouldEqual, "name1")

		convey.So(portsInfo[2].PodNs, convey.ShouldEqual, "ns1")
		convey.So(portsInfo[2].PodName, convey.ShouldEqual, "name1")
	})
}

func TestGetPortInfoErr(t *testing.T) {
	rspInterfaceLeafOfPod := []*client.Node{
		{
			Key:   "/paasnet/tenants/ns1/pods/ns1/name1/interfaces/if1",
			Value: "/paasnet/tenant/ns1/interfaces/portId1ns1name1/self"},
		{
			Key:   "/paasnet/tenants/ns1/pods/ns1/name1/interfaces/if2",
			Value: "/paasnet/tenant/ns1/interfaces/portId2ns1name1/self"},
		{
			Key:   "/paasnet/tenants/ns1/pods/ns1/name1/interfaces/if3",
			Value: "/paasnet/tenant/ns1/interfaces/portId3ns1name1/self"},
	}

	outputs := []Output{
		{StubVals: Values{rspInterfaceLeafOfPod, nil}},
	}

	stubs := StubFuncSeq(&adapter.ReadDirFromDb, outputs)
	defer stubs.Reset()
	stubs.StubFunc(&cni.GetGlobalContext,
		&cni.AgentContext{ClusterID: "cluster1", HostIP: "node1"})

	rspInterfaceInfo1 := "{\"name\":\"eth0\",\"status\":\"\",\"port_id\":\"portId1\",\"ip\":\"192.168.183.237\"," +
		"\"mac_address\":\"fa:16:3e:3e:ed:69\",\"network_id\":\"4cfd6854-ade7-4e1b-b68d-163b5b4002d5\"," +
		"\"subnet_id\":\"723c34d7-fb8b-4fcd-a3d5-0d06be3f4ecc\",\"device_id\":\"bc1b77a5-1f31-11e7-a72b-fa163e05987b\"," +
		"\"owner_type\":\"pod\",\"port_type\":\"nodpdk\",\"bus_info\":\"\",\"net_plane_type\":\"std\"," +
		"\"net_plane_name\":\"net_api\",\"tenant_id\":\"admin\",\"nic_type\":\"normal\",\"pod_name\":\"name1\"," +
		"\"pod_ns\":\"ns1\",\"accelerate\":\"false\"}"
	rspInterfaceInfo2 := "{\"name\":\"eth0\",\"status\":\"\",\"port_id\":\"portId2\",\"ip\":\"192.168.183.238\"," +
		"\"mac_address\":\"fa:16:3e:3e:ed:69\",\"network_id\":\"4cfd6854-ade7-4e1b-b68d-163b5b4002d5\"," +
		"\"subnet_id\":\"723c34d7-fb8b-4fcd-a3d5-0d06be3f4ecc\",\"device_id\":\"bc1b77a5-1f31-11e7-a72b-fa163e05987b\"," +
		"\"owner_type\":\"pod\",\"port_type\":\"nodpdk\",\"bus_info\":\"\",\"net_plane_type\":\"std\"," +
		"\"net_plane_name\":\"net_api\",\"tenant_id\":\"admin\",\"nic_type\":\"normal\",\"pod_name\":\"name1\"," +
		"\"pod_ns\":\"ns1\",\"accelerate\":\"false\"}"
	rspInterfaceInfo3 := "{\"name\":\"eth0\",\"status\":\"\",\"port_id\":\"portId3\",\"ip\":\"192.168.183.239\"," +
		"\"mac_address\":\"fa:16:3e:3e:ed:69\",\"network_id\":\"4cfd6854-ade7-4e1b-b68d-163b5b4002d5\"," +
		"\"subnet_id\":\"723c34d7-fb8b-4fcd-a3d5-0d06be3f4ecc\",\"device_id\":\"bc1b77a5-1f31-11e7-a72b-fa163e05987b\"," +
		"\"owner_type\":\"pod\",\"port_type\":\"nodpdk\",\"bus_info\":\"\",\"net_plane_type\":\"std\"," +
		"\"net_plane_name\":\"net_api\",\"tenant_id\":\"admin\",\"nic_type\":\"normal\",\"pod_name\":\"name1\"," +
		"\"pod_ns\":\"ns1\",\"accelerate\":\"false\"}"
	outputs1 := []Output{
		{StubVals: Values{rspInterfaceInfo1, errobj.ErrDbKeyNotFound}},
		{StubVals: Values{rspInterfaceInfo2, errobj.ErrDbKeyNotFound}},
		{StubVals: Values{rspInterfaceInfo3, errobj.ErrDbKeyNotFound}},
	}

	stubs.StubFuncSeq(&adapter.ReadLeafFromDb, outputs1)
	ports, _ := GetKeysOfAllPortsInPodDir("ns1", "name1")

	convey.Convey("TestGetPortInfoErr\n", t, func() {
		portsInfo, _ := GetPortInfo(ports)
		fmt.Printf("ports:%v\n", portsInfo)
		convey.So(len(portsInfo), convey.ShouldEqual, 0)
	})
}

//GetPortsOfPod
func TestGetPortsOfPodSucc(t *testing.T) {
	rspInterfaceLeafOfPod := []*client.Node{
		{
			Key:   "/paasnet/tenants/ns1/pods/ns1/name1/interfaces/if1",
			Value: "/paasnet/tenant/ns1/interfaces/portId1ns1name1/self"},
		{
			Key:   "/paasnet/tenants/ns1/pods/ns1/name1/interfaces/if2",
			Value: "/paasnet/tenant/ns1/interfaces/portId2ns1name1/self"},
		{
			Key:   "/paasnet/tenants/ns1/pods/ns1/name1/interfaces/if3",
			Value: "/paasnet/tenant/ns1/interfaces/portId3ns1name1/self"},
	}

	outputs := []Output{
		{StubVals: Values{rspInterfaceLeafOfPod, nil}},
	}

	stubs := StubFuncSeq(&adapter.ReadDirFromDb, outputs)
	defer stubs.Reset()
	stubs.StubFunc(&cni.GetGlobalContext,
		&cni.AgentContext{ClusterID: "cluster1", HostIP: "node1"})

	rspInterfaceInfo1 := "{\"name\":\"eth0\",\"status\":\"\",\"port_id\":\"portId1\",\"ip\":\"192.168.183.237\"," +
		"\"mac_address\":\"fa:16:3e:3e:ed:69\",\"network_id\":\"4cfd6854-ade7-4e1b-b68d-163b5b4002d5\"," +
		"\"subnet_id\":\"723c34d7-fb8b-4fcd-a3d5-0d06be3f4ecc\",\"device_id\":\"bc1b77a5-1f31-11e7-a72b-fa163e05987b\"," +
		"\"owner_type\":\"pod\",\"port_type\":\"nodpdk\",\"bus_info\":\"\",\"net_plane_type\":\"std\"," +
		"\"net_plane_name\":\"net_api\",\"tenant_id\":\"admin\",\"nic_type\":\"normal\",\"pod_name\":\"name1\"," +
		"\"pod_ns\":\"ns1\",\"accelerate\":\"false\"}"
	rspInterfaceInfo2 := "{\"name\":\"eth0\",\"status\":\"\",\"port_id\":\"portId2\",\"ip\":\"192.168.183.238\"," +
		"\"mac_address\":\"fa:16:3e:3e:ed:69\",\"network_id\":\"4cfd6854-ade7-4e1b-b68d-163b5b4002d5\"," +
		"\"subnet_id\":\"723c34d7-fb8b-4fcd-a3d5-0d06be3f4ecc\",\"device_id\":\"bc1b77a5-1f31-11e7-a72b-fa163e05987b\"," +
		"\"owner_type\":\"pod\",\"port_type\":\"nodpdk\",\"bus_info\":\"\",\"net_plane_type\":\"std\"," +
		"\"net_plane_name\":\"net_api\",\"tenant_id\":\"admin\",\"nic_type\":\"normal\",\"pod_name\":\"name1\"," +
		"\"pod_ns\":\"ns1\",\"accelerate\":\"false\"}"
	rspInterfaceInfo3 := "{\"name\":\"eth0\",\"status\":\"\",\"port_id\":\"portId3\",\"ip\":\"192.168.183.239\"," +
		"\"mac_address\":\"fa:16:3e:3e:ed:69\",\"network_id\":\"4cfd6854-ade7-4e1b-b68d-163b5b4002d5\"," +
		"\"subnet_id\":\"723c34d7-fb8b-4fcd-a3d5-0d06be3f4ecc\",\"device_id\":\"bc1b77a5-1f31-11e7-a72b-fa163e05987b\"," +
		"\"owner_type\":\"pod\",\"port_type\":\"nodpdk\",\"bus_info\":\"\",\"net_plane_type\":\"std\"," +
		"\"net_plane_name\":\"net_api\",\"tenant_id\":\"admin\",\"nic_type\":\"normal\",\"pod_name\":\"name1\"," +
		"\"pod_ns\":\"ns1\",\"accelerate\":\"false\"}"
	outputs1 := []Output{
		{StubVals: Values{rspInterfaceInfo1, nil}},
		{StubVals: Values{rspInterfaceInfo2, nil}},
		{StubVals: Values{rspInterfaceInfo3, nil}},
	}

	stubs.StubFuncSeq(&adapter.ReadLeafFromDb, outputs1)

	convey.Convey("TestGetPortsOfPodSucc\n", t, func() {
		interfaces, err := GetPortsOfPod("ns1", "name1")
		fmt.Printf("ports:%v\n", interfaces)
		convey.So(err, convey.ShouldBeNil)
		convey.So(len(interfaces), convey.ShouldEqual, 3)

		convey.So(interfaces[0].PodNs, convey.ShouldEqual, "ns1")
		convey.So(interfaces[0].PodName, convey.ShouldEqual, "name1")
		convey.So(interfaces[0].Id, convey.ShouldEqual, "portId1")

		convey.So(interfaces[1].PodNs, convey.ShouldEqual, "ns1")
		convey.So(interfaces[1].PodName, convey.ShouldEqual, "name1")
		convey.So(interfaces[1].Id, convey.ShouldEqual, "portId2")

		convey.So(interfaces[2].PodNs, convey.ShouldEqual, "ns1")
		convey.So(interfaces[2].PodName, convey.ShouldEqual, "name1")
		convey.So(interfaces[2].Id, convey.ShouldEqual, "portId3")
	})
}

func TestGetPortsOfPodErrInBranch1(t *testing.T) {
	rspInterfaceLeafOfPod := []*client.Node{
		{
			Key:   "/paasnet/tenants/ns1/pods/ns1/name1/interfaces/if1",
			Value: "/paasnet/tenant/ns1/interfaces/portId1ns1name1/self"},
		{
			Key:   "/paasnet/tenants/ns1/pods/ns1/name1/interfaces/if2",
			Value: "/paasnet/tenant/ns1/interfaces/portId2ns1name1/self"},
		{
			Key:   "/paasnet/tenants/ns1/pods/ns1/name1/interfaces/if3",
			Value: "/paasnet/tenant/ns1/interfaces/portId3ns1name1/self"},
	}

	outputs := []Output{
		{StubVals: Values{rspInterfaceLeafOfPod, errobj.ErrDbKeyNotFound}},
	}

	stubs := StubFuncSeq(&adapter.ReadDirFromDb, outputs)
	defer stubs.Reset()
	stubs.StubFunc(&cni.GetGlobalContext,
		&cni.AgentContext{ClusterID: "cluster1", HostIP: "node1"})

	convey.Convey("TestGetPortOfPodErrInBranch1\n", t, func() {
		interfaces, err := GetPortsOfPod("ns1", "name1")
		fmt.Printf("interfaces:%v\n", interfaces)
		convey.So(errobj.IsEqual(err, errobj.ErrDbKeyNotFound), convey.ShouldBeTrue)
		convey.So(len(interfaces), convey.ShouldEqual, 0)
	})
}

func TestGetPortsOfPodErrInBranch2(t *testing.T) {
	rspInterfaceLeafOfPod := []*client.Node{
		{
			Key:   "/paasnet/tenants/ns1/pods/ns1/name1/interfaces/if1",
			Value: "/paasnet/tenant/ns1/interfaces/portId1ns1name1/self"},
		{
			Key:   "/paasnet/tenants/ns1/pods/ns1/name1/interfaces/if2",
			Value: "/paasnet/tenant/ns1/interfaces/portId2ns1name1/self"},
		{
			Key:   "/paasnet/tenants/ns1/pods/ns1/name1/interfaces/if3",
			Value: "/paasnet/tenant/ns1/interfaces/portId3ns1name1/self"},
	}

	outputs := []Output{
		{StubVals: Values{rspInterfaceLeafOfPod, nil}},
	}

	stubs := StubFuncSeq(&adapter.ReadDirFromDb, outputs)
	defer stubs.Reset()
	stubs.StubFunc(&cni.GetGlobalContext,
		&cni.AgentContext{ClusterID: "cluster1", HostIP: "node1"})

	rspInterfaceInfo1 := "{\"name\":\"eth0\",\"status\":\"\",\"port_id\":\"portId1\",\"ip\":\"192.168.183.237\"," +
		"\"mac_address\":\"fa:16:3e:3e:ed:69\",\"network_id\":\"4cfd6854-ade7-4e1b-b68d-163b5b4002d5\"," +
		"\"subnet_id\":\"723c34d7-fb8b-4fcd-a3d5-0d06be3f4ecc\",\"device_id\":\"bc1b77a5-1f31-11e7-a72b-fa163e05987b\"," +
		"\"owner_type\":\"pod\",\"port_type\":\"nodpdk\",\"bus_info\":\"\",\"net_plane_type\":\"std\"," +
		"\"net_plane_name\":\"net_api\",\"tenant_id\":\"admin\",\"nic_type\":\"normal\",\"pod_name\":\"name1\"," +
		"\"pod_ns\":\"ns1\",\"accelerate\":\"false\"}"
	rspInterfaceInfo2 := "{\"name\":\"eth0\",\"status\":\"\",\"port_id\":\"portId2\",\"ip\":\"192.168.183.238\"," +
		"\"mac_address\":\"fa:16:3e:3e:ed:69\",\"network_id\":\"4cfd6854-ade7-4e1b-b68d-163b5b4002d5\"," +
		"\"subnet_id\":\"723c34d7-fb8b-4fcd-a3d5-0d06be3f4ecc\",\"device_id\":\"bc1b77a5-1f31-11e7-a72b-fa163e05987b\"," +
		"\"owner_type\":\"pod\",\"port_type\":\"nodpdk\",\"bus_info\":\"\",\"net_plane_type\":\"std\"," +
		"\"net_plane_name\":\"net_api\",\"tenant_id\":\"admin\",\"nic_type\":\"normal\",\"pod_name\":\"name1\"," +
		"\"pod_ns\":\"ns1\",\"accelerate\":\"false\"}"
	rspInterfaceInfo3 := "{\"name\":\"eth0\",\"status\":\"\",\"port_id\":\"portId3\",\"ip\":\"192.168.183.239\"," +
		"\"mac_address\":\"fa:16:3e:3e:ed:69\",\"network_id\":\"4cfd6854-ade7-4e1b-b68d-163b5b4002d5\"," +
		"\"subnet_id\":\"723c34d7-fb8b-4fcd-a3d5-0d06be3f4ecc\",\"device_id\":\"bc1b77a5-1f31-11e7-a72b-fa163e05987b\"," +
		"\"owner_type\":\"pod\",\"port_type\":\"nodpdk\",\"bus_info\":\"\",\"net_plane_type\":\"std\"," +
		"\"net_plane_name\":\"net_api\",\"tenant_id\":\"admin\",\"nic_type\":\"normal\",\"pod_name\":\"name1\"," +
		"\"pod_ns\":\"ns1\",\"accelerate\":\"false\"}"
	outputs1 := []Output{
		{StubVals: Values{rspInterfaceInfo1, errobj.ErrDbConnRefused}},
		{StubVals: Values{rspInterfaceInfo2, errobj.ErrDbConnRefused}},
		{StubVals: Values{rspInterfaceInfo3, errobj.ErrDbConnRefused}},
	}

	stubs.StubFuncSeq(&adapter.ReadLeafFromDb, outputs1)

	convey.Convey("TestGetPortOfPodErrInBranch2\n", t, func() {
		interfaces, err := GetPortsOfPod("ns1", "name1")
		fmt.Printf("interfaces:%v\n", interfaces)
		convey.So(errobj.IsEqual(err, errobj.ErrNwPortInfo), convey.ShouldBeTrue)
		convey.So(len(interfaces), convey.ShouldEqual, 0)
	})
}

//RecyclePortSource

func TestRecyclePortSourceErrInBranch2(t *testing.T) {
	delPort := iaasaccessor.Interface{
		Accelerate: "false",
		NetPlane:   "media",
	}
	stubs := StubFunc(&adapter.DestroyPort, errors.New("destroy port error"))
	defer stubs.Reset()

	convey.Convey("TestRecyclePortSourceErrInBranch2\n", t, func() {
		err := RecyclePortSource(delPort)
		convey.So(errobj.IsEqual(err, errobj.ErrNwRecyclePort), convey.ShouldBeTrue)
	})
}

//ClearPortInDb
func TestClearPortInDbSucc(t *testing.T) {
	stubs := StubFunc(&adapter.ClearDirFromDb, nil)
	defer stubs.Reset()
	stubs.StubFunc(&adapter.ClearLeafFromDb, nil)
	stubs.StubFunc(&adapter.ClearLeafFromRemoteDB, nil)
	port := iaasaccessor.Interface{
		Id:        "protId1",
		NetworkId: "networkId1",
	}
	convey.Convey("TestClearPortInDbSucc\n", t, func() {
		err := ClearPortInDb("ns1", "name1", port)
		convey.So(err, convey.ShouldBeNil)
	})
}

func TestClearPortInDbErrInBranch1(t *testing.T) {
	stubs := StubFunc(&adapter.ClearDirFromDb, errobj.ErrDbConnRefused)
	defer stubs.Reset()

	port := iaasaccessor.Interface{
		Id:        "protId1",
		NetworkId: "networkId1",
	}
	convey.Convey("TestClearPortInDbErrInBranch1\n", t, func() {
		err := ClearPortInDb("ns1", "name1", port)
		convey.So(errobj.IsEqual(err, errobj.ErrDbConnRefused), convey.ShouldBeTrue)
	})
}

func TestClearPortInDbErrInBranch2(t *testing.T) {
	stubs := StubFunc(&adapter.ClearDirFromDb, nil)
	defer stubs.Reset()
	outputs := []Output{
		{StubVals: Values{errobj.ErrDbConnRefused}},
		{StubVals: Values{nil}},
		{StubVals: Values{nil}},
		{StubVals: Values{nil}},
		{StubVals: Values{nil}},
	}
	stubs.StubFuncSeq(&adapter.ClearLeafFromDb, outputs)
	stubs.StubFuncSeq(&adapter.ClearLeafFromRemoteDB, outputs)

	port := iaasaccessor.Interface{
		Id:        "protId1",
		NetworkId: "networkId1",
	}

	convey.Convey("TestClearPortInDbErrInBranch2\n", t, func() {
		err := ClearPortInDb("ns1", "name1", port)
		convey.So(errobj.IsEqual(err, errobj.ErrDbConnRefused), convey.ShouldBeTrue)
	})
}

func TestClearPortInDbErrInBranch3(t *testing.T) {
	stubs := StubFunc(&adapter.ClearDirFromDb, nil)
	defer stubs.Reset()
	outputs := []Output{
		{StubVals: Values{nil}},
		{StubVals: Values{nil}},
		{StubVals: Values{errobj.ErrDbConnRefused}},
		{StubVals: Values{nil}},
		{StubVals: Values{nil}},
	}
	stubs.StubFuncSeq(&adapter.ClearLeafFromDb, outputs)
	stubs.StubFuncSeq(&adapter.ClearLeafFromRemoteDB, outputs)

	port := iaasaccessor.Interface{
		Id:        "protId1",
		NetworkId: "networkId1",
	}

	convey.Convey("TestClearPortInDbErrInBranch3\n", t, func() {
		err := ClearPortInDb("ns1", "name1", port)
		convey.So(errobj.IsEqual(err, errobj.ErrDbConnRefused), convey.ShouldBeTrue)
	})
}

func TestClearPortInDbErrInBranch4(t *testing.T) {
	stubs := StubFunc(&adapter.ClearDirFromDb, nil)
	defer stubs.Reset()
	outputs := []Output{
		{StubVals: Values{nil}},
		{StubVals: Values{nil}},
		{StubVals: Values{nil}},
		{StubVals: Values{errobj.ErrDbConnRefused}},
		{StubVals: Values{nil}},
	}
	stubs.StubFuncSeq(&adapter.ClearLeafFromDb, outputs)
	stubs.StubFuncSeq(&adapter.ClearLeafFromRemoteDB, outputs)

	port := iaasaccessor.Interface{
		Id:        "protId1",
		NetworkId: "networkId1",
		NetPlane:  "eio",
	}

	convey.Convey("TestClearPortInDbErrInBranch4\n", t, func() {
		err := ClearPortInDb("ns1", "name1", port)
		convey.So(errobj.IsEqual(err, errobj.ErrDbConnRefused), convey.ShouldBeTrue)
	})
}

//ClearPodInDb
func TestClearPodInDbSucc(t *testing.T) {
	stubs := StubFunc(&adapter.ClearDirFromDb, nil)
	defer stubs.Reset()
	stubs.StubFunc(&adapter.ClearLeafFromDb, nil)
	convey.Convey("TestClearPodInDbSucc\n", t, func() {
		err := ClearPodInDb("ns1", "ns1", "name1")
		convey.So(err, convey.ShouldBeNil)
	})
}

func TestClearPodInDbErrInBranch1(t *testing.T) {
	stubs := StubFunc(&adapter.ClearDirFromDb, errobj.ErrDbConnRefused)
	defer stubs.Reset()

	convey.Convey("TestClearPodInDbErrInBranch1\n", t, func() {
		err := ClearPodInDb("ns1", "ns1", "name1")
		convey.So(errobj.IsEqual(err, errobj.ErrDbConnRefused), convey.ShouldBeTrue)
	})
}

func TestClearPodInDbErrInBranch2(t *testing.T) {
	stubs := StubFunc(&adapter.ClearDirFromDb, nil)
	defer stubs.Reset()
	stubs.StubFunc(&adapter.ClearLeafFromDb, errobj.ErrDbConnRefused)
	convey.Convey("TestClearPodInDbErrInBranch2\n", t, func() {
		err := ClearPodInDb("ns1", "ns1", "name1")
		convey.So(errobj.IsEqual(err, errobj.ErrDbConnRefused), convey.ShouldBeTrue)
	})
}

//ClearPortsOfNoUsedPod
func TestClearPortsOfNoUsedPodSucc(t *testing.T) {
	rspInterfaceLeafOfPod := []*client.Node{
		{
			Key:   "/paasnet/tenants/ns1/pods/ns1/name1/interfaces/if1",
			Value: "/paasnet/tenant/ns1/interfaces/portId1ns1name1/self"},
		{
			Key:   "/paasnet/tenants/ns1/pods/ns1/name1/interfaces/if2",
			Value: "/paasnet/tenant/ns1/interfaces/portId2ns1name1/self"},
		{
			Key:   "/paasnet/tenants/ns1/pods/ns1/name1/interfaces/if3",
			Value: "/paasnet/tenant/ns1/interfaces/portId3ns1name1/self"},
	}

	outputs := []Output{
		{StubVals: Values{rspInterfaceLeafOfPod, nil}},
	}

	stubs := StubFuncSeq(&adapter.ReadDirFromDb, outputs)
	defer stubs.Reset()
	stubs.StubFunc(&cni.GetGlobalContext,
		&cni.AgentContext{ClusterID: "cluster1", HostIP: "node1"})

	rspInterfaceInfo1 := "{\"name\":\"eth0\",\"status\":\"\",\"port_id\":\"portId1\",\"ip\":\"192.168.183.237\"," +
		"\"mac_address\":\"fa:16:3e:3e:ed:69\",\"network_id\":\"4cfd6854-ade7-4e1b-b68d-163b5b4002d5\"," +
		"\"subnet_id\":\"723c34d7-fb8b-4fcd-a3d5-0d06be3f4ecc\",\"device_id\":\"bc1b77a5-1f31-11e7-a72b-fa163e05987b\"," +
		"\"owner_type\":\"pod\",\"port_type\":\"nodpdk\",\"bus_info\":\"\",\"net_plane_type\":\"std\"," +
		"\"net_plane_name\":\"net_api\",\"tenant_id\":\"admin\",\"nic_type\":\"normal\",\"pod_name\":\"name1\"," +
		"\"pod_ns\":\"ns1\",\"accelerate\":\"false\"}"
	rspInterfaceInfo2 := "{\"name\":\"eth0\",\"status\":\"\",\"port_id\":\"portId2\",\"ip\":\"192.168.183.238\"," +
		"\"mac_address\":\"fa:16:3e:3e:ed:69\",\"network_id\":\"4cfd6854-ade7-4e1b-b68d-163b5b4002d5\"," +
		"\"subnet_id\":\"723c34d7-fb8b-4fcd-a3d5-0d06be3f4ecc\",\"device_id\":\"bc1b77a5-1f31-11e7-a72b-fa163e05987b\"," +
		"\"owner_type\":\"pod\",\"port_type\":\"nodpdk\",\"bus_info\":\"\",\"net_plane_type\":\"std\"," +
		"\"net_plane_name\":\"net_api\",\"tenant_id\":\"admin\",\"nic_type\":\"normal\",\"pod_name\":\"name1\"," +
		"\"pod_ns\":\"ns1\",\"accelerate\":\"false\"}"
	rspInterfaceInfo3 := "{\"name\":\"eth0\",\"status\":\"\",\"port_id\":\"portId3\",\"ip\":\"192.168.183.239\"," +
		"\"mac_address\":\"fa:16:3e:3e:ed:69\",\"network_id\":\"4cfd6854-ade7-4e1b-b68d-163b5b4002d5\"," +
		"\"subnet_id\":\"723c34d7-fb8b-4fcd-a3d5-0d06be3f4ecc\",\"device_id\":\"bc1b77a5-1f31-11e7-a72b-fa163e05987b\"," +
		"\"owner_type\":\"pod\",\"port_type\":\"nodpdk\",\"bus_info\":\"\",\"net_plane_type\":\"std\"," +
		"\"net_plane_name\":\"net_api\",\"tenant_id\":\"admin\",\"nic_type\":\"normal\",\"pod_name\":\"name1\"," +
		"\"pod_ns\":\"ns1\",\"accelerate\":\"false\"}"
	outputs1 := []Output{
		{StubVals: Values{rspInterfaceInfo1, nil}},
		{StubVals: Values{rspInterfaceInfo2, nil}},
		{StubVals: Values{rspInterfaceInfo3, nil}},
	}

	stubs.StubFuncSeq(&adapter.ReadLeafFromDb, outputs1)
	stubs.StubFunc(&adapter.DestroyPort, nil)
	stubs.StubFunc(&adapter.ClearDirFromDb, nil)
	stubs.StubFunc(&adapter.ClearLeafFromDb, nil)
	stubs.StubFunc(&adapter.ClearLeafFromRemoteDB, nil)

	convey.Convey("TestClearPortsOfNoUsedPodSucc\n", t, func() {
		err := ClearPortsOfNoUsedPod("ns1", "name1")
		convey.So(err, convey.ShouldBeNil)
	})
}

func TestClearPortsOfNoUsedPodErrInBranch1(t *testing.T) {
	rspInterfaceLeafOfPod := []*client.Node{
		{
			Key:   "/paasnet/tenants/ns1/pods/ns1/name1/interfaces/if1",
			Value: "/paasnet/tenant/ns1/interfaces/portId1ns1name1/self"},
		{
			Key:   "/paasnet/tenants/ns1/pods/ns1/name1/interfaces/if2",
			Value: "/paasnet/tenant/ns1/interfaces/portId2ns1name1/self"},
		{
			Key:   "/paasnet/tenants/ns1/pods/ns1/name1/interfaces/if3",
			Value: "/paasnet/tenant/ns1/interfaces/portId3ns1name1/self"},
	}

	outputs := []Output{
		{StubVals: Values{rspInterfaceLeafOfPod, nil}},
	}

	stubs := StubFuncSeq(&adapter.ReadDirFromDb, outputs)
	defer stubs.Reset()
	stubs.StubFunc(&cni.GetGlobalContext,
		&cni.AgentContext{ClusterID: "cluster1", HostIP: "node1"})

	rspInterfaceInfo1 := "{\"name\":\"eth0\",\"status\":\"\",\"port_id\":\"portId1\",\"ip\":\"192.168.183.237\"," +
		"\"mac_address\":\"fa:16:3e:3e:ed:69\",\"network_id\":\"4cfd6854-ade7-4e1b-b68d-163b5b4002d5\"," +
		"\"subnet_id\":\"723c34d7-fb8b-4fcd-a3d5-0d06be3f4ecc\",\"device_id\":\"bc1b77a5-1f31-11e7-a72b-fa163e05987b\"," +
		"\"owner_type\":\"pod\",\"port_type\":\"nodpdk\",\"bus_info\":\"\",\"net_plane_type\":\"std\"," +
		"\"net_plane_name\":\"net_api\",\"tenant_id\":\"admin\",\"nic_type\":\"normal\",\"pod_name\":\"name1\"," +
		"\"pod_ns\":\"ns1\",\"accelerate\":\"false\"}"
	rspInterfaceInfo2 := "{\"name\":\"eth0\",\"status\":\"\",\"port_id\":\"portId2\",\"ip\":\"192.168.183.238\"," +
		"\"mac_address\":\"fa:16:3e:3e:ed:69\",\"network_id\":\"4cfd6854-ade7-4e1b-b68d-163b5b4002d5\"," +
		"\"subnet_id\":\"723c34d7-fb8b-4fcd-a3d5-0d06be3f4ecc\",\"device_id\":\"bc1b77a5-1f31-11e7-a72b-fa163e05987b\"," +
		"\"owner_type\":\"pod\",\"port_type\":\"nodpdk\",\"bus_info\":\"\",\"net_plane_type\":\"std\"," +
		"\"net_plane_name\":\"net_api\",\"tenant_id\":\"admin\",\"nic_type\":\"normal\",\"pod_name\":\"name1\"," +
		"\"pod_ns\":\"ns1\",\"accelerate\":\"false\"}"
	rspInterfaceInfo3 := "{\"name\":\"eth0\",\"status\":\"\",\"port_id\":\"portId3\",\"ip\":\"192.168.183.239\"," +
		"\"mac_address\":\"fa:16:3e:3e:ed:69\",\"network_id\":\"4cfd6854-ade7-4e1b-b68d-163b5b4002d5\"," +
		"\"subnet_id\":\"723c34d7-fb8b-4fcd-a3d5-0d06be3f4ecc\",\"device_id\":\"bc1b77a5-1f31-11e7-a72b-fa163e05987b\"," +
		"\"owner_type\":\"pod\",\"port_type\":\"nodpdk\",\"bus_info\":\"\",\"net_plane_type\":\"std\"," +
		"\"net_plane_name\":\"net_api\",\"tenant_id\":\"admin\",\"nic_type\":\"normal\",\"pod_name\":\"name1\"," +
		"\"pod_ns\":\"ns1\",\"accelerate\":\"false\"}"
	outputs1 := []Output{
		{StubVals: Values{rspInterfaceInfo1, errobj.ErrDbConnRefused}},
		{StubVals: Values{rspInterfaceInfo2, errobj.ErrDbConnRefused}},
		{StubVals: Values{rspInterfaceInfo3, errobj.ErrDbConnRefused}},
	}

	stubs.StubFuncSeq(&adapter.ReadLeafFromDb, outputs1)

	convey.Convey("TestClearPortsOfNoUsedPodErrInBranch1\n", t, func() {
		err := ClearPortsOfNoUsedPod("ns1", "name1")
		fmt.Printf("error:%v", err)
		convey.So(errobj.IsEqual(err, errobj.ErrNwPortInfo), convey.ShouldBeTrue)
	})
}

func TestClearPortsOfNoUsedPodErrInBranch2(t *testing.T) {
	rspInterfaceLeafOfPod := []*client.Node{
		{
			Key:   "/paasnet/tenants/ns1/pods/ns1/name1/interfaces/if1",
			Value: "/paasnet/tenant/ns1/interfaces/portId1ns1name1/self"},
		{
			Key:   "/paasnet/tenants/ns1/pods/ns1/name1/interfaces/if2",
			Value: "/paasnet/tenant/ns1/interfaces/portId2ns1name1/self"},
		{
			Key:   "/paasnet/tenants/ns1/pods/ns1/name1/interfaces/if3",
			Value: "/paasnet/tenant/ns1/interfaces/portId3ns1name1/self"},
	}

	outputs := []Output{
		{StubVals: Values{rspInterfaceLeafOfPod, nil}},
	}

	stubs := StubFuncSeq(&adapter.ReadDirFromDb, outputs)
	defer stubs.Reset()
	stubs.StubFunc(&cni.GetGlobalContext,
		&cni.AgentContext{ClusterID: "cluster1", HostIP: "node1"})

	rspInterfaceInfo1 := "{\"name\":\"eth0\",\"status\":\"\",\"port_id\":\"portId1\",\"ip\":\"192.168.183.237\"," +
		"\"mac_address\":\"fa:16:3e:3e:ed:69\",\"network_id\":\"4cfd6854-ade7-4e1b-b68d-163b5b4002d5\"," +
		"\"subnet_id\":\"723c34d7-fb8b-4fcd-a3d5-0d06be3f4ecc\",\"device_id\":\"bc1b77a5-1f31-11e7-a72b-fa163e05987b\"," +
		"\"owner_type\":\"pod\",\"port_type\":\"nodpdk\",\"bus_info\":\"\",\"net_plane_type\":\"std\"," +
		"\"net_plane_name\":\"net_api\",\"tenant_id\":\"admin\",\"nic_type\":\"normal\",\"pod_name\":\"name1\"," +
		"\"pod_ns\":\"ns1\",\"accelerate\":\"false\"}"
	rspInterfaceInfo2 := "{\"name\":\"eth0\",\"status\":\"\",\"port_id\":\"portId2\",\"ip\":\"192.168.183.238\"," +
		"\"mac_address\":\"fa:16:3e:3e:ed:69\",\"network_id\":\"4cfd6854-ade7-4e1b-b68d-163b5b4002d5\"," +
		"\"subnet_id\":\"723c34d7-fb8b-4fcd-a3d5-0d06be3f4ecc\",\"device_id\":\"bc1b77a5-1f31-11e7-a72b-fa163e05987b\"," +
		"\"owner_type\":\"pod\",\"port_type\":\"nodpdk\",\"bus_info\":\"\",\"net_plane_type\":\"std\"," +
		"\"net_plane_name\":\"net_api\",\"tenant_id\":\"admin\",\"nic_type\":\"normal\",\"pod_name\":\"name1\"," +
		"\"pod_ns\":\"ns1\",\"accelerate\":\"false\"}"
	rspInterfaceInfo3 := "{\"name\":\"eth0\",\"status\":\"\",\"port_id\":\"portId3\",\"ip\":\"192.168.183.239\"," +
		"\"mac_address\":\"fa:16:3e:3e:ed:69\",\"network_id\":\"4cfd6854-ade7-4e1b-b68d-163b5b4002d5\"," +
		"\"subnet_id\":\"723c34d7-fb8b-4fcd-a3d5-0d06be3f4ecc\",\"device_id\":\"bc1b77a5-1f31-11e7-a72b-fa163e05987b\"," +
		"\"owner_type\":\"pod\",\"port_type\":\"nodpdk\",\"bus_info\":\"\",\"net_plane_type\":\"std\"," +
		"\"net_plane_name\":\"net_api\",\"tenant_id\":\"admin\",\"nic_type\":\"normal\",\"pod_name\":\"name1\"," +
		"\"pod_ns\":\"ns1\",\"accelerate\":\"false\"}"
	outputs1 := []Output{
		{StubVals: Values{rspInterfaceInfo1, nil}},
		{StubVals: Values{rspInterfaceInfo2, nil}},
		{StubVals: Values{rspInterfaceInfo3, nil}},
	}

	stubs.StubFuncSeq(&adapter.ReadLeafFromDb, outputs1)
	stubs.StubFunc(&adapter.DestroyPort, errors.New("error delete port"))

	convey.Convey("TestClearPortsOfNoUsedPodErrInBranch1\n", t, func() {
		err := ClearPortsOfNoUsedPod("ns1", "name1")
		convey.So(errobj.IsEqual(err, errobj.ErrNwRecyclePort), convey.ShouldBeTrue)
	})
}

func TestClearPortsOfNoUsedPodErrInBranch3(t *testing.T) {
	rspInterfaceLeafOfPod := []*client.Node{
		{
			Key:   "/paasnet/tenants/ns1/pods/ns1/name1/interfaces/if1",
			Value: "/paasnet/tenant/ns1/interfaces/portId1ns1name1/self"},
		{
			Key:   "/paasnet/tenants/ns1/pods/ns1/name1/interfaces/if2",
			Value: "/paasnet/tenant/ns1/interfaces/portId2ns1name1/self"},
		{
			Key:   "/paasnet/tenants/ns1/pods/ns1/name1/interfaces/if3",
			Value: "/paasnet/tenant/ns1/interfaces/portId3ns1name1/self"},
	}

	outputs := []Output{
		{StubVals: Values{rspInterfaceLeafOfPod, nil}},
	}

	stubs := StubFuncSeq(&adapter.ReadDirFromDb, outputs)
	defer stubs.Reset()
	stubs.StubFunc(&cni.GetGlobalContext,
		&cni.AgentContext{ClusterID: "cluster1", HostIP: "node1"})

	rspInterfaceInfo1 := "{\"name\":\"eth0\",\"status\":\"\",\"port_id\":\"portId1\",\"ip\":\"192.168.183.237\"," +
		"\"mac_address\":\"fa:16:3e:3e:ed:69\",\"network_id\":\"4cfd6854-ade7-4e1b-b68d-163b5b4002d5\"," +
		"\"subnet_id\":\"723c34d7-fb8b-4fcd-a3d5-0d06be3f4ecc\",\"device_id\":\"bc1b77a5-1f31-11e7-a72b-fa163e05987b\"," +
		"\"owner_type\":\"pod\",\"port_type\":\"nodpdk\",\"bus_info\":\"\",\"net_plane_type\":\"std\"," +
		"\"net_plane_name\":\"net_api\",\"tenant_id\":\"admin\",\"nic_type\":\"normal\",\"pod_name\":\"name1\"," +
		"\"pod_ns\":\"ns1\",\"accelerate\":\"false\"}"
	rspInterfaceInfo2 := "{\"name\":\"eth0\",\"status\":\"\",\"port_id\":\"portId2\",\"ip\":\"192.168.183.238\"," +
		"\"mac_address\":\"fa:16:3e:3e:ed:69\",\"network_id\":\"4cfd6854-ade7-4e1b-b68d-163b5b4002d5\"," +
		"\"subnet_id\":\"723c34d7-fb8b-4fcd-a3d5-0d06be3f4ecc\",\"device_id\":\"bc1b77a5-1f31-11e7-a72b-fa163e05987b\"," +
		"\"owner_type\":\"pod\",\"port_type\":\"nodpdk\",\"bus_info\":\"\",\"net_plane_type\":\"std\"," +
		"\"net_plane_name\":\"net_api\",\"tenant_id\":\"admin\",\"nic_type\":\"normal\",\"pod_name\":\"name1\"," +
		"\"pod_ns\":\"ns1\",\"accelerate\":\"false\"}"
	rspInterfaceInfo3 := "{\"name\":\"eth0\",\"status\":\"\",\"port_id\":\"portId3\",\"ip\":\"192.168.183.239\"," +
		"\"mac_address\":\"fa:16:3e:3e:ed:69\",\"network_id\":\"4cfd6854-ade7-4e1b-b68d-163b5b4002d5\"," +
		"\"subnet_id\":\"723c34d7-fb8b-4fcd-a3d5-0d06be3f4ecc\",\"device_id\":\"bc1b77a5-1f31-11e7-a72b-fa163e05987b\"," +
		"\"owner_type\":\"pod\",\"port_type\":\"nodpdk\",\"bus_info\":\"\",\"net_plane_type\":\"std\"," +
		"\"net_plane_name\":\"net_api\",\"tenant_id\":\"admin\",\"nic_type\":\"normal\",\"pod_name\":\"name1\"," +
		"\"pod_ns\":\"ns1\",\"accelerate\":\"false\"}"
	outputs1 := []Output{
		{StubVals: Values{rspInterfaceInfo1, nil}},
		{StubVals: Values{rspInterfaceInfo2, nil}},
		{StubVals: Values{rspInterfaceInfo3, nil}},
	}

	stubs.StubFuncSeq(&adapter.ReadLeafFromDb, outputs1)
	stubs.StubFunc(&adapter.DestroyPort, nil)
	stubs.StubFunc(&adapter.ClearDirFromDb, nil)
	stubs.StubFunc(&adapter.ClearLeafFromDb, errobj.ErrDbConnRefused)
	stubs.StubFunc(&adapter.ClearLeafFromRemoteDB, nil)

	convey.Convey("TestClearPortsOfNoUsedPodErrInBranch1\n", t, func() {
		err := ClearPortsOfNoUsedPod("ns1", "name1")
		convey.So(errobj.IsEqual(err, errobj.ErrDbConnRefused), convey.ShouldBeTrue)
	})
}

//GetRegularCheckInterval
func TestGetRegularCheckIntervalSucc(t *testing.T) {
	stubs := StubFunc(&adapter.ReadLeafFromDb, "30", nil)
	defer stubs.Reset()

	convey.Convey("TestGetRegularCheckIntervalSucc\n", t, func() {
		timeStr := GetRegularCheckInterval()
		convey.So(timeStr, convey.ShouldEqual, 30*time.Minute)
	})
}

func TestGetRegularCheckIntervalErrInBranch1(t *testing.T) {
	stubs := StubFunc(&adapter.ReadLeafFromDb, "", errors.New("error"))
	defer stubs.Reset()

	convey.Convey("TestGetRegularCheckIntervalErrInBranch1\n", t, func() {
		timeStr := GetRegularCheckInterval()
		convey.So(timeStr, convey.ShouldEqual, 15*time.Minute)
	})
}

func TestGetRegularCheckIntervalErrInBranch2(t *testing.T) {
	stubs := StubFunc(&adapter.ReadLeafFromDb, "-1", nil)
	defer stubs.Reset()

	convey.Convey("TestGetRegularCheckIntervalErrInBranch2\n", t, func() {
		timeStr := GetRegularCheckInterval()
		convey.So(timeStr, convey.ShouldEqual, 15*time.Minute)
	})
}

func TestGetRegularCheckIntervalErrInBranch3(t *testing.T) {
	stubs := StubFunc(&adapter.ReadLeafFromDb, "1441", nil)
	defer stubs.Reset()

	convey.Convey("TestGetRegularCheckIntervalErrInBranch3\n", t, func() {
		timeStr := GetRegularCheckInterval()
		convey.So(timeStr, convey.ShouldEqual, 15*time.Minute)
	})
}

//RegularCollectAndClear
func TestRegularCollectAndClear(t *testing.T) {
	etcdPods := []*client.Node{
		{Key: "/paasnet/runtime/clusters/cluster1/nodes/node1/pods/ns1"},
		{Key: "/paasnet/runtime/clusters/cluster1/nodes/node1/pods/ns2"},
		{Key: "/paasnet/runtime/clusters/cluster1/nodes/node1/pods/ns3"},
	}
	etcdPodsNs1 := []*client.Node{
		{Key: "/paasnet/runtime/clusters/cluster1/nodes/node1/pods/ns1/name1"},
	}
	etcdPodsNs2 := []*client.Node{
		{Key: "/paasnet/runtime/clusters/cluster1/nodes/node1/pods/ns2/name2"},
	}
	etcdPodsNs3 := []*client.Node{
		{Key: "/paasnet/runtime/clusters/cluster1/nodes/node1/pods/ns3/name3"},
	}
	outputs := []Output{
		{StubVals: Values{etcdPods, nil}},
		{StubVals: Values{etcdPodsNs1, nil}},
		{StubVals: Values{etcdPodsNs2, nil}},
		{StubVals: Values{etcdPodsNs3, nil}},
	}

	stubs := StubFuncSeq(&adapter.ReadDirFromDb, outputs)
	defer stubs.Reset()
	stubs.StubFunc(&cni.GetGlobalContext,
		&cni.AgentContext{ClusterID: "cluster1", HostIP: "node1", ClusterType: "k8s"})

	podList := []*jason.Object{}
	pod2 := RspPodMetadataOfK8s{Metadata: RspPodOfK8s{Name: "name2", Namespace: "ns2"}}
	pod3 := RspPodMetadataOfK8s{Metadata: RspPodOfK8s{Name: "name3", Namespace: "ns3"}}
	rsp2Byte, _ := json.Marshal(pod2)
	pod2Json, _ := jason.NewObjectFromBytes(rsp2Byte)
	rsp3Byte, _ := json.Marshal(pod3)
	pod3Json, _ := jason.NewObjectFromBytes(rsp3Byte)
	podList = append(podList, pod2Json)
	podList = append(podList, pod3Json)
	stubs.StubFunc(&adapter.GetPodsByNodeID, podList, nil)

	rspInterfaceLeafOfPod := []*client.Node{
		{
			Key:   "/paasnet/tenants/ns1/pods/ns1/name1/interfaces/if1",
			Value: "/paasnet/tenant/ns1/interfaces/portId1ns1name1/self"},
		{
			Key:   "/paasnet/tenants/ns1/pods/ns1/name1/interfaces/if2",
			Value: "/paasnet/tenant/ns1/interfaces/portId2ns1name1/self"},
		{
			Key:   "/paasnet/tenants/ns1/pods/ns1/name1/interfaces/if3",
			Value: "/paasnet/tenant/ns1/interfaces/portId3ns1name1/self"},
	}

	stubs.StubFunc(&adapter.ReadDirFromDb, rspInterfaceLeafOfPod, nil)
	stubs.StubFunc(&cni.GetGlobalContext,
		&cni.AgentContext{ClusterID: "cluster1", HostIP: "node1"})

	rspInterfaceInfo1 := "{\"name\":\"eth0\",\"status\":\"\",\"port_id\":\"portId1\",\"ip\":\"192.168.183.237\"," +
		"\"mac_address\":\"fa:16:3e:3e:ed:69\",\"network_id\":\"4cfd6854-ade7-4e1b-b68d-163b5b4002d5\"," +
		"\"subnet_id\":\"723c34d7-fb8b-4fcd-a3d5-0d06be3f4ecc\",\"device_id\":\"bc1b77a5-1f31-11e7-a72b-fa163e05987b\"," +
		"\"owner_type\":\"pod\",\"port_type\":\"nodpdk\",\"bus_info\":\"\",\"net_plane_type\":\"std\"," +
		"\"net_plane_name\":\"net_api\",\"tenant_id\":\"admin\",\"nic_type\":\"normal\",\"pod_name\":\"name1\"," +
		"\"pod_ns\":\"ns1\",\"accelerate\":\"false\"}"
	rspInterfaceInfo2 := "{\"name\":\"eth0\",\"status\":\"\",\"port_id\":\"portId2\",\"ip\":\"192.168.183.238\"," +
		"\"mac_address\":\"fa:16:3e:3e:ed:69\",\"network_id\":\"4cfd6854-ade7-4e1b-b68d-163b5b4002d5\"," +
		"\"subnet_id\":\"723c34d7-fb8b-4fcd-a3d5-0d06be3f4ecc\",\"device_id\":\"bc1b77a5-1f31-11e7-a72b-fa163e05987b\"," +
		"\"owner_type\":\"pod\",\"port_type\":\"nodpdk\",\"bus_info\":\"\",\"net_plane_type\":\"std\"," +
		"\"net_plane_name\":\"net_api\",\"tenant_id\":\"admin\",\"nic_type\":\"normal\",\"pod_name\":\"name1\"," +
		"\"pod_ns\":\"ns1\",\"accelerate\":\"false\"}"
	rspInterfaceInfo3 := "{\"name\":\"eth0\",\"status\":\"\",\"port_id\":\"portId3\",\"ip\":\"192.168.183.239\"," +
		"\"mac_address\":\"fa:16:3e:3e:ed:69\",\"network_id\":\"4cfd6854-ade7-4e1b-b68d-163b5b4002d5\"," +
		"\"subnet_id\":\"723c34d7-fb8b-4fcd-a3d5-0d06be3f4ecc\",\"device_id\":\"bc1b77a5-1f31-11e7-a72b-fa163e05987b\"," +
		"\"owner_type\":\"pod\",\"port_type\":\"nodpdk\",\"bus_info\":\"\",\"net_plane_type\":\"std\"," +
		"\"net_plane_name\":\"net_api\",\"tenant_id\":\"admin\",\"nic_type\":\"normal\",\"pod_name\":\"name1\"," +
		"\"pod_ns\":\"ns1\",\"accelerate\":\"false\"}"
	outputs2 := []Output{
		{StubVals: Values{rspInterfaceInfo1, nil}},
		{StubVals: Values{rspInterfaceInfo2, nil}},
		{StubVals: Values{rspInterfaceInfo3, nil}},
	}

	stubs.StubFuncSeq(&adapter.ReadLeafFromDb, outputs2)
	stubs.StubFunc(&adapter.DestroyPort, nil)
	stubs.StubFunc(&adapter.ClearDirFromDb, nil)
	stubs.StubFunc(&adapter.ClearLeafFromDb, nil)

	convey.Convey("TestRegularCollectAndClear\n", t, func() {
		RegularCollectAndClear()
	})
}

//CollectAndClear
//func TestCollectAndClear(t *testing.T) {
//	etcdPods := []*client.Node{
//		&client.Node{Key: "/paasnet/runtime/clusters/cluster1/nodes/node1/pods/ns1"},
//		&client.Node{Key: "/paasnet/runtime/clusters/cluster1/nodes/node1/pods/ns2"},
//		&client.Node{Key: "/paasnet/runtime/clusters/cluster1/nodes/node1/pods/ns3"},
//	}
//	etcdPodsNs1 := []*client.Node{
//		&client.Node{Key: "/paasnet/runtime/clusters/cluster1/nodes/node1/pods/ns1/name1"},
//	}
//	etcdPodsNs2 := []*client.Node{
//		&client.Node{Key: "/paasnet/runtime/clusters/cluster1/nodes/node1/pods/ns2/name2"},
//	}
//	etcdPodsNs3 := []*client.Node{
//		&client.Node{Key: "/paasnet/runtime/clusters/cluster1/nodes/node1/pods/ns3/name3"},
//	}
//	outputs := []Output{
//		[]interface{}{etcdPods, nil},
//		[]interface{}{etcdPodsNs1, nil},
//		[]interface{}{etcdPodsNs2, nil},
//		[]interface{}{etcdPodsNs3, nil},
//	}
//
//	stubs := StubFuncSeq(&adapter.ReadDirFromDb, 0, outputs)
//	defer stubs.Reset()
//	stubs.StubFunc(&cni.GetGlobalContext,
//		&cni.AgentContext{ClusterId: "cluster1", HostIp: "node1", ClusterType:"k8s"})
//
//	podList := []*jason.Object{}
//	pod2 := RspPodMetadataOfK8s{Metadata: RspPodOfK8s{Name:"name2", Namespace:"ns2"}}
//	pod3 := RspPodMetadataOfK8s{Metadata: RspPodOfK8s{Name:"name3", Namespace:"ns3"}}
//	rsp2Byte, _  := json.Marshal(pod2)
//	pod2Json, _ := jason.NewObjectFromBytes(rsp2Byte)
//	rsp3Byte, _  := json.Marshal(pod3)
//	pod3Json, _ := jason.NewObjectFromBytes(rsp3Byte)
//	podList = append(podList, pod2Json)
//	podList = append(podList, pod3Json)
//	stubs.StubFunc(&adapter.GetPodsByNodeId, podList, nil)
//
//
//	rspInterfaceLeafOfPod := []*client.Node{
//		&client.Node{
//			Key: "/paasnet/tenants/ns1/pods/ns1/name1/interfaces/if1",
//			Value: "/paasnet/tenant/ns1/interfaces/portId1ns1name1/self"},
//		&client.Node{
//			Key: "/paasnet/tenants/ns1/pods/ns1/name1/interfaces/if2",
//			Value: "/paasnet/tenant/ns1/interfaces/portId2ns1name1/self"},
//		&client.Node{
//			Key: "/paasnet/tenants/ns1/pods/ns1/name1/interfaces/if3",
//			Value: "/paasnet/tenant/ns1/interfaces/portId3ns1name1/self"},
//	}
//
//	outputs1 := []Output{
//		[]interface{}{rspInterfaceLeafOfPod, nil},
//	}
//
//	stubs.StubFuncSeq(&adapter.ReadDirFromDb, 0, outputs1)
//	stubs.StubFunc(&cni.GetGlobalContext,
//		&cni.AgentContext{ClusterId: "cluster1", HostIp: "node1"})
//
//	rspInterfaceInfo1 := "{\"name\":\"eth0\",\"status\":\"\",\"port_id\":\"portId1\",\"ip\":\"192.168.183.237\","+
//	"\"mac_address\":\"fa:16:3e:3e:ed:69\",\"network_id\":\"4cfd6854-ade7-4e1b-b68d-163b5b4002d5\","+
//	"\"subnet_id\":\"723c34d7-fb8b-4fcd-a3d5-0d06be3f4ecc\",\"device_id\":\"bc1b77a5-1f31-11e7-a72b-fa163e05987b\","+
//	"\"owner_type\":\"pod\",\"port_type\":\"nodpdk\",\"bus_info\":\"\",\"net_plane_type\":\"std\","+
//	"\"net_plane_name\":\"net_api\",\"tenant_id\":\"admin\",\"nic_type\":\"normal\",\"pod_name\":\"name1\","+
//	"\"pod_ns\":\"ns1\",\"accelerate\":\"false\"}"
//	rspInterfaceInfo2 := "{\"name\":\"eth0\",\"status\":\"\",\"port_id\":\"portId2\",\"ip\":\"192.168.183.238\","+
//	"\"mac_address\":\"fa:16:3e:3e:ed:69\",\"network_id\":\"4cfd6854-ade7-4e1b-b68d-163b5b4002d5\","+
//	"\"subnet_id\":\"723c34d7-fb8b-4fcd-a3d5-0d06be3f4ecc\",\"device_id\":\"bc1b77a5-1f31-11e7-a72b-fa163e05987b\","+
//	"\"owner_type\":\"pod\",\"port_type\":\"nodpdk\",\"bus_info\":\"\",\"net_plane_type\":\"std\","+
//	"\"net_plane_name\":\"net_api\",\"tenant_id\":\"admin\",\"nic_type\":\"normal\",\"pod_name\":\"name1\","+
//	"\"pod_ns\":\"ns1\",\"accelerate\":\"false\"}"
//	rspInterfaceInfo3 := "{\"name\":\"eth0\",\"status\":\"\",\"port_id\":\"portId3\",\"ip\":\"192.168.183.239\","+
//	"\"mac_address\":\"fa:16:3e:3e:ed:69\",\"network_id\":\"4cfd6854-ade7-4e1b-b68d-163b5b4002d5\","+
//	"\"subnet_id\":\"723c34d7-fb8b-4fcd-a3d5-0d06be3f4ecc\",\"device_id\":\"bc1b77a5-1f31-11e7-a72b-fa163e05987b\","+
//	"\"owner_type\":\"pod\",\"port_type\":\"nodpdk\",\"bus_info\":\"\",\"net_plane_type\":\"std\","+
//	"\"net_plane_name\":\"net_api\",\"tenant_id\":\"admin\",\"nic_type\":\"normal\",\"pod_name\":\"name1\","+
//	"\"pod_ns\":\"ns1\",\"accelerate\":\"false\"}"
//	outputs2 := []Output{
//		[]interface{}{rspInterfaceInfo1, nil},
//		[]interface{}{rspInterfaceInfo2, nil},
//		[]interface{}{rspInterfaceInfo3, nil},
//	}
//
//	stubs.StubFuncSeq(&adapter.ReadLeafFromDb, 0, outputs2)
//	stubs.StubFunc(&adapter.DestroyPort, nil)
//	stubs.StubFunc(&adapter.ClearDirFromDb, nil)
//	stubs.StubFunc(&adapter.ClearLeafFromDb, nil)
//	stubs.StubFunc(&adapter.ReadLeafFromDb, "1", nil)
//
//	convey.Convey("TestCollectAndClear\n", t, func() {
//		CollectAndClear()
//	})
//}

func TestRecyclePodTableRoleLoad(t *testing.T) {
	tnTle := GetRecyclePodTableTableSingleton()
	Convey("TestRecyclePodTableRoleLoad\n", t, func() {
		Convey("test SUCC", func() {
			stubs := gostub.StubFunc(&osencap.OsStat, nil, nil)
			stubs.StubFunc(&adapter.DataPersisterLoadFromMemFile, nil)
			err := tnTle.Load()
			stubs.Reset()
			So(err, ShouldBeNil)
		})
		Convey("test FAIL- os.Stat error", func() {
			errStr := "recycle pod table doesn't exist"
			stubs := gostub.StubFunc(&osencap.OsStat, nil, errors.New(errStr))
			err := tnTle.Load()
			stubs.Reset()
			So(strings.Contains(err.Error(), errStr), ShouldBeTrue)
		})
		Convey("test FAIL- save to memory file error", func() {
			errStr := "restore recycle pod table failed"
			stubs := gostub.StubFunc(&osencap.OsStat, nil, nil)
			stubs.StubFunc(&adapter.DataPersisterLoadFromMemFile, errors.New(errStr))
			err := tnTle.Load()
			stubs.Reset()
			So(strings.Contains(err.Error(), errStr), ShouldBeTrue)
		})
	})
}

func TestRecyclePodTableRoleGetAll(t *testing.T) {
	tnTle := GetRecyclePodTableTableSingleton()
	Convey("TestRecyclePodTableRoleGetAll\n", t, func() {
		Convey("GetAll() test-empty map\n", func() {
			stubs := gostub.StubFunc(&osencap.OsStat, nil, nil)
			stubs.StubFunc(&adapter.DataPersisterLoadFromMemFile, nil)

			tnMaps, err := tnTle.GetAll()
			So(err, ShouldBeNil)
			So(len(tnMaps), ShouldBeZeroValue)
			stubs.Reset()
		})
		Convey("GetAll() test\n", func() {
			stubs := gostub.StubFunc(&osencap.OsStat, nil, nil)
			stubs.StubFunc(&adapter.DataPersisterSaveToMemFile, nil)

			tnValue := RecyclePodValue{
				PodNs:   "ns",
				PodName: "name",
				Count:   1,
			}
			err := tnTle.Insert("nsname", tnValue)
			So(err, ShouldBeNil)

			stubs.StubFunc(&adapter.DataPersisterLoadFromMemFile, nil)
			tnMaps, err := tnTle.GetAll()
			So(err, ShouldBeNil)
			So(len(tnMaps), ShouldEqual, 1)
			So(tnMaps["nsname"], ShouldResemble, tnValue)
			err = tnTle.Delete("nsname")
			So(err, ShouldBeNil)
			stubs.Reset()
		})
		Convey("GetAll() faile: table not exist\n", func() {
			errStr := "recycle pod table doesn't exist"
			stub := gostub.StubFunc(&osencap.OsStat, nil, errors.New(errStr))

			tnMap, err := tnTle.GetAll()
			So(strings.Contains(err.Error(), errStr), ShouldBeTrue)
			So(tnMap, ShouldBeNil)
			stub.Reset()
		})
		Convey("GetAll() faile:restore pod table error\n", func() {
			errStr := "restore recycle pod table failed"
			stub := gostub.StubFunc(&osencap.OsStat, nil, nil)
			stub.StubFunc(&adapter.DataPersisterSaveToMemFile, errors.New(errStr))
			tnMap, err := tnTle.GetAll()
			So(strings.Contains(err.Error(), errStr), ShouldBeTrue)
			So(tnMap, ShouldBeNil)
			stub.Reset()
		})
	})
}

func TestRecyclePodTableRoleGet(t *testing.T) {
	tnTle := GetRecyclePodTableTableSingleton()
	Convey("TestRecyclePodTableRoleGet\n", t, func() {
		Convey("Get() test\n", func() {
			value, err := tnTle.Get("nsname")
			So(err, ShouldEqual, errobj.ErrRecordNtExist)
			So(value, ShouldBeNil)

			outputs := []gostub.Output{{StubVals: gostub.Values{nil}, Times: 2}}
			stub := gostub.StubFuncSeq(&adapter.DataPersisterSaveToMemFile, outputs)
			tnValue := RecyclePodValue{
				PodNs:   "ns",
				PodName: "name",
				Count:   1,
			}
			err = tnTle.Insert("nsname", tnValue)
			So(err, ShouldBeNil)

			value, err = tnTle.Get("nsname")
			So(err, ShouldBeNil)
			So(value.(RecyclePodValue), ShouldResemble, tnValue)

			err = tnTle.Delete("nsname")
			So(err, ShouldBeNil)
			stub.Reset()
		})
		Convey("Get() test failed\n", func() {
			errStr := "store to ram failed"
			stub := gostub.StubFunc(&adapter.DataPersisterSaveToMemFile, errors.New(errStr))
			tnValue2 := RecyclePodValue{
				PodNs:   "ns1",
				PodName: "name1",
				Count:   1,
			}
			err := tnTle.Insert("ns1name1", tnValue2)
			So(strings.Contains(err.Error(), errStr), ShouldBeTrue)
			stub.Reset()
		})
	})
}

func TestRecyclePodTableRoleInc(t *testing.T) {
	tnTle := GetRecyclePodTableTableSingleton()
	Convey("TestRecyclePodTableRoleInc\n", t, func() {
		Convey("SUCC test\n", func() {
			stubs := gostub.StubFunc(&osencap.OsStat, nil, nil)
			stubs.StubFunc(&adapter.DataPersisterSaveToMemFile, nil)

			tnValue := RecyclePodValue{
				PodNs:   "ns",
				PodName: "name",
				Count:   1,
			}
			err := tnTle.Insert("nsname", tnValue)
			So(err, ShouldBeNil)

			err1 := tnTle.Inc("nsname", "")
			So(err1, ShouldBeNil)

			err2 := tnTle.Dec("", "")
			So(err2, ShouldBeNil)

			bl := tnTle.IsEmpty("")
			So(bl, ShouldBeFalse)
			stubs.Reset()
		})
		Convey("FAILED Inc test\n", func() {
			stubs := gostub.StubFunc(&osencap.OsStat, nil, nil)
			err := tnTle.Inc("nsname1", "")
			So(err, ShouldEqual, errobj.ErrRecordNtExist)

			//errStr := "store to ram failed"
			//stubs.StubFunc(&adapter.DataPersisterLoadFromMemFile, errors.New(errStr))
			//err = tnTle.Inc("nsname", "")
			//So(strings.Contains(err.Error(), errStr), ShouldBeTrue)
			stubs.Reset()
		})
	})
}

func TestIsPodNsExist(t *testing.T) {
	podnss := []string{"ns1", "ns2", "ns3"}
	Convey("TestIsPodNsExist\n", t, func() {
		Convey("SUCC test\n", func() {
			bl := isExist("ns1", podnss)
			So(bl, ShouldBeTrue)

			ble2 := isExist("", podnss)
			So(ble2, ShouldBeTrue)
		})
		Convey("FAILED test\n", func() {
			ble1 := isExist("ns4", podnss)
			So(ble1, ShouldBeFalse)

			ble3 := isExist("ns1", nil)
			So(ble3, ShouldBeFalse)
		})

	})
}
