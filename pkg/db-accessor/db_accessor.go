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

package dbaccessor

import (
	"errors"
	"github.com/coreos/etcd/client"
)

type DbAccessor interface {
	SaveLeaf(k, v string) error
	ReadDir(k string) ([]*client.Node, error)
	ReadLeaf(k string) (string, error)
	DeleteLeaf(k string) error
	DeleteDir(url string) error
	WatcherDir(url string) (*client.Response, error)
	Lock(url string) bool
	Unlock(url string) bool
}

func CheckDataBase(Db DbAccessor) error {
	testDIR := "/this-is-a-database-check/"
	key := "etcd"
	value := "DataBase-Config-is-OK"
	if Db == nil {
		return errors.New("DbClient is nil")
	}

	testURL := testDIR + key
	e0 := Db.SaveLeaf(testURL, value)
	if e0 != nil {
		return e0
	}
	v, e1 := Db.ReadLeaf(testURL)
	if e1 != nil {
		return e1
	}
	if v != value {
		return errors.New("DataBase-is-ERROR")
	}
	e2 := Db.DeleteLeaf(testURL)
	if e2 != nil {
		return e2
	}
	e3 := Db.DeleteDir(testDIR)
	if e3 != nil {
		return e3
	}
	return nil
}

/***********************************************************************/

type Agent struct {
	Id     string `json:"id"`
	Ip     string `json:"ip"`
	Status string `json:"status"`
	TTL    int    `json:"ttl"`
}

type Sync struct {
	Interval string   `json:"interval"`
	Client   *Agent   `json:"client"`
	Agents   []*Agent `json:"agents"`
}

type SyncRsp struct {
	Interval string  `json:"interval"`
	Client   Agent   `json:"client"`
	Agents   []Agent `json:"agents"`
}

const (
	AgentStatusReady            string = "READY"
	AgentStatusDown             string = "DOWN"
	DefaultIntervalTimeInSecond int    = 15
)

/*************************************************************************/
func GetCloudTNameSpaceKey(podns, podId string) string {
	var keyCloudT string = "/paas/cloudt"
	url := keyCloudT + "/pods/" + podns + "/" + podId
	return url
}

/*************************************************************************
*
* FUNCTION OF GET DATABASE KEY
*
*************************************************************************/
func GetKeyOfRoot() string {
	return "/paasnet"
}

func GetKeyOfConf() string {
	return GetKeyOfRoot() + "/conf"
}

func GetKeyOfTenants() string {
	return GetKeyOfRoot() + "/tenants"
}

func GetKeyOfPublic() string {
	return GetKeyOfRoot() + "/public"
}

func GetKeyOfDcs() string {
	return GetKeyOfRoot() + "/dcs"
}

func GetKeyOfRuntime() string {
	return GetKeyOfRoot() + "/runtime"
}

func GetKeyOfEmbeddedServer() string {
	return "/embedded_manager" + "/server"
}

func GetKeyOfEmbeddedServerVnis() string {
	return GetKeyOfEmbeddedServer() + "/vnis"
}

func GetKeyOfEmbeddedServerPorts() string {
	return GetKeyOfEmbeddedServer() + "/ports"
}

func GetKeyOfEmbeddedServerPortID(id string) string {
	return GetKeyOfEmbeddedServerPorts() + "/" + id
}

func GetKeyOfEmbeddedServerSubnets() string {
	return GetKeyOfEmbeddedServer() + "/subnets"
}

func GetKeyOfClusterUUID() string {
	return GetKeyOfKnitter() + "/cluster_uuid"
}

func GetKeyOfEmbeddedServerSubnetID(id string) string {
	return GetKeyOfEmbeddedServerSubnets() + "/" + id
}

func GetKeyOfEmbeddedServerNetworks() string {
	return GetKeyOfEmbeddedServer() + "/networks"
}

func GetKeyOfEmbeddedServerNetworkID(id string) string {
	return GetKeyOfEmbeddedServerNetworks() + "/" + id
}

func GetKeyOfOpenstack() string {
	return GetKeyOfConf() + "/openstack"
}

func GetKeyOfVnfm() string {
	return GetKeyOfConf() + "/mano"
}

func GetKeyOfRegularCheck() string {
	return GetKeyOfConf() + "/regular_check_interval"
}

func GetKeyOfOpenShift() string {
	return GetKeyOfConf() + "/openshift"
}

func GetKeyOfKnitterManagerUrl() string {
	return GetKeyOfConf() + "/knitter_manager"
}

func GetKeyOfPaasUUID() string {
	return GetKeyOfConf() + "/paas_uuid"
}

func GetKeyOfTenantSelf(tenant_id string) string {
	return GetKeyOfTenants() + "/" + tenant_id + "/self"
}

func GetKeyOfTenant(tenant_id string) string {
	return GetKeyOfTenants() + "/" + tenant_id
}

func GetKeyOfRouterGroup(tenant_id string) string {
	return GetKeyOfTenants() + "/" + tenant_id + "/routers"
}

func GetKeyOfNetworkGroup(tenant_id string) string {
	return GetKeyOfTenants() + "/" + tenant_id + "/networks"
}

func GetKeyOfInterfaceGroup(tenant_id string) string {
	return GetKeyOfTenants() + "/" + tenant_id + "/interfaces"
}

func GetKeyOfPodNsGroup(tenant_id string) string {
	return GetKeyOfTenants() + "/" + tenant_id + "/pods"
}

func GetKeyOfPodGroup(tenant_id, ns string) string {
	return GetKeyOfPodNsGroup(tenant_id) + "/" + ns
}

func GetKeyOfRouter(tenant_id, router_id string) string {
	return GetKeyOfRouterGroup(tenant_id) + "/" + tenant_id
}

func GetKeyOfNetwork(tenant_id, network_id string) string {
	return GetKeyOfNetworkGroup(tenant_id) + "/" + network_id
}

func GetKeyOfInterface(tenant_id, interfaceID string) string {
	return GetKeyOfInterfaceGroup(tenant_id) + "/" + interfaceID
}

func GetKeyOfPod(tenant_id, pod_ns, pod_name string) string {
	return GetKeyOfPodGroup(tenant_id, pod_ns) + "/" + pod_name
}

func GetKeyOfVmidForPod(tenant_id, pod_ns, pod_name string) string {
	return GetKeyOfPod(tenant_id, pod_ns, pod_name) + "/vmid"
}

func GetKeyOfRouterSelf(tenant_id, router_id string) string {
	return GetKeyOfRouter(tenant_id, router_id) + "/self"
}

func GetKeyOfNetworkSelf(tenant_id, network_id string) string {
	return GetKeyOfNetwork(tenant_id, network_id) + "/self"
}

func GetKeyOfInterfaceSelf(tenant_id, interfaceID string) string {
	return GetKeyOfInterface(tenant_id, interfaceID) + "/self"
}

func GetKeyOfPodSelf(tenant_id, pod_ns, pod_name string) string {
	return GetKeyOfPod(tenant_id, pod_ns, pod_name) + "/self"
}

func GetKeyOfInterfaceGroupInNetwork(tenant_id, network_id string) string {
	return GetKeyOfNetwork(tenant_id, network_id) + "/interfaces"
}

func GetKeyOfInterfaceInNetwork(tenant_id, network_id, port_id string) string {
	return GetKeyOfInterfaceGroupInNetwork(tenant_id, network_id) + "/" + port_id
}

func GetKeyOfInterfaceGroupInRouter(tenant_id, router_id string) string {
	return GetKeyOfRouter(tenant_id, router_id) + "/interfaces"
}

func GetKeyOfInterfaceInRouter(tenant_id, router_id, port_id string) string {
	return GetKeyOfInterfaceGroupInRouter(tenant_id, router_id) + "/" + port_id
}

func GetKeyOfInterfaceGroupInPod(tenant_id, pod_ns, pod_name string) string {
	return GetKeyOfPod(tenant_id, pod_ns, pod_name) + "/interfaces"
}

func GetKeyOfNetworkByInterface(tenant_id, interfaceID string) string {
	return GetKeyOfInterface(tenant_id, interfaceID) + "/network"
}

func GetKeyOfRouterByInterface(tenant_id, interfaceID string) string {
	return GetKeyOfInterface(tenant_id, interfaceID) + "/router"
}

func GetKeyOfPodByInterface(tenant_id, interfaceID string) string {
	return GetKeyOfInterface(tenant_id, interfaceID) + "/pod"
}

func GetKeyOfPublicNetworkGroup() string {
	return GetKeyOfPublic() + "/network"
}

func GetKeyOfPublicNetwork(network_id string) string {
	return GetKeyOfPublicNetworkGroup() + "/" + network_id
}

func GetKeyOfCluster(cluster_ip string) string {
	return GetKeyOfRuntime() + "/clusters/" + cluster_ip
}

func GetKeyOfClusterNodes(cluster_ip string) string {
	return GetKeyOfRuntime() + "/clusters/" + cluster_ip + "/nodes"
}

func GetKeyOfNode(cluster_ip, node_ip string) string {
	return GetKeyOfCluster(cluster_ip) + "/nodes/" + node_ip
}

func GetKeyOfC0(cluster_ip, node_ip string) string {
	return GetKeyOfNode(cluster_ip, node_ip) + "/c0"
}

func GetKeyOfContainerIdForC0(cluster_ip, node_ip string) string {
	return GetKeyOfC0(cluster_ip, node_ip) + "/container_id"
}

func GetKeyOfImageNameForC0(cluster_ip, node_ip string) string {
	return GetKeyOfC0(cluster_ip, node_ip) + "/image_name"
}

func GetKeyOfPodForC0(cluster_ip, node_ip string) string {
	return GetKeyOfC0(cluster_ip, node_ip) + "/pod"
}

func GetKeyOfInterfaceForNode(cluster_ip, node_ip string) string {
	return GetKeyOfNode(cluster_ip, node_ip) + "/interfaces"
}

func GetKeyOfPodsForNode(cluster_ip, node_ip string) string {
	return GetKeyOfNode(cluster_ip, node_ip) + "/pods"
}

func GetKeyOfPodForNode(cluster_ip, node_ip, pod_ns, pod_name string) string {
	return GetKeyOfPodsForNode(cluster_ip, node_ip) + "/" + pod_ns + "/" + pod_name
}

func GetKeyOfExceptionalPortGroup() string {
	return GetKeyOfRuntime() + "/resource/exceptional/ports"
}

func GetKeyOfExceptionalPort(port_id string) string {
	return GetKeyOfExceptionalPortGroup() + "/" + port_id
}

func GetKeyOfTopoSyncData() string {
	return GetKeyOfRuntime() + "/topo/sync"
}

func GetKeyOfOseToken(oseId string) string {
	return GetKeyOfOpenShift() + "/" + oseId
}

func GetKeyOfInterfaceInPod(tenant_id, port_id, pod_ns, pod_name string) string {
	return GetKeyOfInterfaceGroupInPod(tenant_id, pod_ns, pod_name) + "/" + port_id
}

func GetKeyOfGwForNode(cluster_ip, node_ip string) string {
	return GetKeyOfNode(cluster_ip, node_ip) + "/gw/interface"
}

func GetKeyOfIaasInterfacesForNode(cluster_ip, node_ip string) string {
	return GetKeyOfNode(cluster_ip, node_ip) + "/interfaces/iaas"
}

func GetKeyOfIaasC0InterfacesForNode(cluster_ip, node_ip string) string {
	return GetKeyOfIaasInterfacesForNode(cluster_ip, node_ip) + "/c0"
}

func GetKeyOfIaasC0InterfaceForNode(cluster_ip, node_ip, network_id string) string {
	return GetKeyOfIaasC0InterfacesForNode(cluster_ip, node_ip) + "/" + network_id
}

func GetKeyOfIaasBr0InterfacesForNode(cluster_ip, node_ip string) string {
	return GetKeyOfIaasInterfacesForNode(cluster_ip, node_ip) + "/br0"
}

func GetKeyOfIaasBr0InterfaceForNode(cluster_ip, node_ip, network_id string) string {
	return GetKeyOfIaasBr0InterfacesForNode(cluster_ip, node_ip) + "/" + network_id
}

func GetKeyOfIaasEioInterfacesForNode(cluster_ip, node_ip string) string {
	return GetKeyOfIaasInterfacesForNode(cluster_ip, node_ip) + "/eio"
}

func GetKeyOfIaasEioInterfaceForNode(cluster_ip, node_ip, port_id string) string {
	return GetKeyOfIaasEioInterfacesForNode(cluster_ip, node_ip) + "/" + port_id
}

func GetKeyOfPaasInterfacesForNode(cluster_ip, node_ip string) string {
	return GetKeyOfNode(cluster_ip, node_ip) + "/interfaces/paas"
}

func GetKeyOfObr0InterfaceForNode(cluster_ip, node_ip, vethName string) string {
	return GetKeyOfNode(cluster_ip, node_ip) + "/interfaces/obr0/" + vethName
}

func GetKeyOfObr0InterfacesForNode(cluster_ip, node_ip string) string {
	return GetKeyOfNode(cluster_ip, node_ip) + "/interfaces/obr0"
}

func GetKeyOfPaasInterfaceForNode(cluster_ip, node_ip, port_id string) string {
	return GetKeyOfPaasInterfacesForNode(cluster_ip, node_ip) + "/" + port_id
}

func GetKeyOfExceptionalPorts() string {
	return GetKeyOfRuntime() + "/resource/exceptional/ports"
}

func GetKeyOfPodInInterface(tenant_id, interfaceID string) string {
	return GetKeyOfInterface(tenant_id, interfaceID) + "/pod"
}

func GetKeyOfRecycleResourceByTimerUrl() string {
	return "/paasnet/conf/timer/recycle"
}

func GetKeyOfC0PodLabel(cluster_ip, node_ip string) string {
	return GetKeyOfC0(cluster_ip, node_ip) + "/PodLabel"
}

func GetKeyOfInterfacesInDc(dcId string) string {
	return GetKeyOfDcs() + "/" + dcId + "/interfaces"
}

func GetKeyOfInterfaceInDc(dcId, ifId string) string {
	return GetKeyOfInterfacesInDc(dcId) + "/" + ifId
}

func GetVnfmBaseUrl(baseURL string, nfInstanceID string) string {
	return baseURL + "/nfs/" + nfInstanceID
}

func GetVnfmJobUrl(baseURL string) string {
	return baseURL + "/nfs"
}

func GetVnfmNetwrokUrl(baseURL string, nfInstanceID string, netName string) string {
	return GetVnfmBaseUrl(baseURL, nfInstanceID) + "/networks?networkName=" + netName
}

func GetVnfmSubNetwrokUrl(baseURL string, nfInstanceID string, subNetName string) string {
	return GetVnfmBaseUrl(baseURL, nfInstanceID) + "/subnets?subnetName=" + subNetName
}

func GetVnfmCreatePortUrl(baseURL string, nfInstanceID string) string {
	return GetVnfmBaseUrl(baseURL, nfInstanceID) + "/cps"
}

func GetVnfmJobResultUrl(baseURL, jobID, operation string) string {
	return GetVnfmJobUrl(baseURL) + "/job_results/" + jobID + "/" + operation
}

func GetVnfmDeletePortUrl(baseURL string, nfInstanceID string, resourceID string) string {
	return GetVnfmBaseUrl(baseURL, nfInstanceID) + "/cps/" + resourceID
}

func GetVnfmPortInfoUrl(baseURL string, nfInstanceID string, resourceID string) string {
	return GetVnfmBaseUrl(baseURL, nfInstanceID) + "/ports?resourceId=" + resourceID
}

func GetVnfmAttachDetachPortUrl(baseURL string, nfInstanceID string, vmName string) string {
	return GetVnfmBaseUrl(baseURL, nfInstanceID) + "/cp_attach_detach/" + vmName
}

func GetKeyOfFixIP() string {
	return GetKeyOfRoot() + "/fixIP"
}

func GetFixIPPortUrl(portID string) string {
	return GetKeyOfFixIP() + "/" + portID
}

func GetKeyOfDefaultPhysnet() string {
	return "/paasnet/defaultphysnet"
}

func GetKeyOfInitConf() string {
	return "/paasnet/initconf"
}

// keys segment for knitter-manager
func GetKeyOfKnitterManager() string {
	return GetKeyOfRoot() + "/knittermanager"
}

func GetKeyOfKnitterManagerConf() string {
	return GetKeyOfKnitterManager() + "/conf"
}

func GetKeyOfCancelWaitPodsDeletedTimeout() string {
	return GetKeyOfKnitterManagerConf() + "/cancel_wait_pods_timeout"
}

func GetKeyOfRecycleMaxCheckTimesUrl() string {
	return "/paasnet/conf/checktimes/recycle"
}

//keys segment for LogicPod
func GetKeyOfKnitter() string {
	return "/knitter"
}

func GetKeyOfManager() string {
	return GetKeyOfKnitter() + "/manager"
}

func GetKeyOfPodsTenantId(tenantId string) string {
	return GetKeyOfManager() + "/pods/" + tenantId
}

func GetKeyOfPodName(tenantId string, PodName string) string {
	return GetKeyOfPodsTenantId(tenantId) + "/" + PodName
}

func GetKeyOfGateway() string {
	return GetKeyOfManager() + "/gateway"
}

func GetKeyOfGatewayID(gatewayID string) string {
	return GetKeyOfGateway() + "/" + gatewayID
}
func GetKeyOfLogicPort(tenantID, podNs, podName, portID string) string {
	return "/tmppaasnet/tenants/" + tenantID + "/pods/" + podNs + "/" + podName + "/ports/" + portID
}

func GetKeyOfLogicPod(tenantID, podNs, podName string) string {
	return "/tmppaasnet/tenants/" + tenantID + "/pods/" + podNs + "/" + podName
}

func GetKeyOfLogicPortsInPod(tenantID, podNs, podName string) string {
	return "/tmppaasnet/tenants/" + tenantID + "/pods/" + podNs + "/" + podName + "/ports"
}

//func GetKeyOfVnicInterface(tenantID, podNs, podName, containerID, portID string) string {
//	return "/physical_resources/tenants/" + tenantID +
//			"/pods/" + podNs + "/" + podName +
//			"/container/" + containerID +
//			"/vnic/" + portID
//}

//func GetKeyOfVnicList(tenantID, podNs, podName, containerID string) string {
//	return "/physical_resources/tenants/" + tenantID +
//		"/pods/" + podNs + "/" + podName +
//		"/container/" + containerID +
//		"/vnic"
//}

//func GetKeyOfContainerResourceDir(tenantID, podNs, podName, containerID string) string {
//	return "/physical_resources/tenants/" + tenantID +
//		"/pods/" + podNs + "/" + podName +
//		"/container/" + containerID
//}

func GetKeyOfInterfaceSelfForRole(driver, interfacesID string) string {
	return "/phys/role/" + driver + "/" + interfacesID
}

func GetKeyOfNouthInterface(containerID, driver, interfacesID string) string {
	return "/phys/manager/nouth/" + containerID + "/" + driver + "/" + interfacesID
}

func GetKeyOfNouthContainer(containerID string) string {
	return "/phys/manager/nouth/" + containerID
}

func GetKeyOfNouthInterfaceList(containerID, driver string) string {
	return "/phys/manager/nouth/" + containerID + "/" + driver
}

func GetKeyOfSouthInterface(netid, chanType, interfacesID string) string {
	return "/phys/manager/south/" + netid + "/" + chanType + "/" + interfacesID
}

func GetKeyOfSouthInterfacePubAttr(netid string) string {
	return "/phys/manager/south/" + netid + "/ispublic"
}

func GetKeyOfIaasTenantInfo(tenantid string) string {
	return GetKeyOfManager() + "/iaas/" + tenantid
}

func GetKeyOfMonitor() string {
	return GetKeyOfKnitter() + "/monitor"
}

func GetKeyOfMonitorPods() string {
	return GetKeyOfMonitor() + "/pods"
}

func GetKeyOfMonitorPod(podNS, podName string) string {
	return GetKeyOfMonitorPods() + "/" + podNS + "/" + podName

}
