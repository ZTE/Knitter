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

package openshift

import (
	"crypto/tls"
	"errors"
	"fmt"
	"github.com/ZTE/Knitter/knitter-agent/domain/manager"
	"github.com/ZTE/Knitter/knitter-agent/infra"
	"github.com/ZTE/Knitter/pkg/klog"
	"github.com/antonholmquist/jason"
	"io/ioutil"
	"net/http"
	"os/exec"
	"time"
)

func ConnectToDocker0(containerID string) error {
	output, _ := exec.Command("connect_pod_to_docker0.sh", containerID).CombinedOutput()
	klog.Infof("%s", string(output))
	return nil
}

type PortAttr struct {
	NetworkName  string
	NetworkPlane string
	PortName     string
	VnicType     string
	Accelerate   string
	NetworkID    string
	PublicNet    bool
	Physical     bool
}

type OsePod struct {
	MangerClient manager.ManagerClient
	PodID        string
	PodType      string
	PodName      string
	PodNs        string
	HostType     string
	VnicType     string
	VMID         string
	PortList     []PortAttr
}

func (self *OsePod) AnalyzeK8sRspd(pod *jason.Object) error {
	podID, podiderr := pod.GetString("metadata", "uid")
	if podiderr != nil {
		klog.Errorf("ose:Get podid from podjson error! podjson.GetString(uuid)-%v", podiderr)
		return fmt.Errorf("%v:ose:Get podid from podjson error", podiderr)
	}
	self.PodID = podID

	nwStr, nwErr := pod.GetString("metadata", "annotations", "networks")
	if nwErr != nil {
		klog.Errorf("ose:podjson.GetString(metadata,annotations,networks) error! -%v", nwErr)
		return nwErr
	}

	klog.Info("ose:nwjson str:", nwStr)
	nwjson, err := jason.NewObjectFromBytes([]byte(nwStr))
	if err != nil {
		klog.Errorf("ose:jason.NewObjectFromBytes(%v) error:%v", nwStr, err)
		return fmt.Errorf("%v:NewObjectFromBytes ERROR", err)
	}
	klog.Info("ose:nwjson obj:", nwjson)

	err = self.AnalyzeV2PodNetTemplate(nwjson)
	if err != nil {
		klog.Info("ose:AnalyzeV2PodNetTemplate error:", err)
		return fmt.Errorf("%v:Analyze-Pod-Net-Template-ERROR", err)
	}

	self.PodType = getPodType(self.PortList)
	klog.Errorf("AnalyzeK8sRspd: podType is: %s", self.PodType)

	return nil
}

func getPodType(portList []PortAttr) string {
	var podType string
	var (
		ctlPlaneExist = false
		medPlaneExist = false
		ctlPlaneAcce  = false
		medPlaneAcce  = false
	)
	for _, port := range portList {
		if port.NetworkPlane == "control" {
			ctlPlaneExist = true
			if port.Accelerate == "true" {
				ctlPlaneAcce = true
			}
		}
		if port.NetworkPlane == "media" {
			medPlaneExist = true
			if port.Accelerate == "true" {
				medPlaneAcce = true
			}
		}
	}

	if ctlPlaneExist && medPlaneExist && ctlPlaneAcce && medPlaneAcce {
		podType = "ct"
	} else if ctlPlaneExist && medPlaneExist && !ctlPlaneAcce && !medPlaneAcce {
		podType = "ct_minus"
	} else {
		podType = "it"
	}

	return podType
}

func (self *OsePod) getPortForPod(portJSON *jason.Object) (*PortAttr, error) {
	networkName, errNw := portJSON.GetString("attach_to_network")
	if errNw != nil {
		klog.Error("Get port attach Network ERROR:", errNw)
		return nil, errNw
	} else if networkName == "" {
		klog.Error("Get port attach Network is blank string")
		return nil, errors.New("get port attach Network is blank string")
	}

	portName, errPort := portJSON.GetString("attributes", "nic_name")
	if errPort != nil || portName == "" {
		portName = "eth_" + networkName
	}
	portFunc, errFunc := portJSON.GetString("attributes", "function")
	if errFunc != nil || portFunc == "" {
		portFunc = "std"
	}
	errPlan := CheckNetworkPlane(portFunc)
	if errPlan != nil {
		portFunc = "std"
	}
	portType, errType := portJSON.GetString("attributes", "nic_type")
	if errType != nil || portType == "" {
		portType = "normal"
	}
	errVnic := CheckVnicType(portType)
	if errVnic != nil {
		portType = "normal"
	}
	isUseDpdk, _ := portJSON.GetString("attributes", "accelerate")
	if isUseDpdk != "true" {
		isUseDpdk = "false"
	} else {
		isUseDpdk = "true"
	}
	if !infra.IsCTNetPlane(portFunc) {
		isUseDpdk = "false"
	}
	var port PortAttr
	port.NetworkName = networkName
	port.NetworkPlane = portFunc
	port.PortName = portName
	port.VnicType = portType
	port.Accelerate = isUseDpdk
	port.Physical = false
	return &port, nil
}

func (self *OsePod) AnalyzeV2PodNetTemplate(nwjson *jason.Object) error {
	portArray, err := nwjson.GetObjectArray("ports")
	if err != nil {
		klog.Errorf("Get Port tArray error! -%v", err)
		return err
	}
	var portListTmp []PortAttr
	for _, obj := range portArray {
		port, err := self.getPortForPod(obj)
		if err != nil {
			return fmt.Errorf("%v:AnalyzeV2PodNetTemplate ERROR", err)
		}
		portListTmp = append(portListTmp, *port)
	}
	self.PortList = portListTmp
	return nil
}

func CheckNetworkPlane(networkPlane string) error {
	switch networkPlane {
	case "control":
		return nil
	case "media":
		return nil
	case "std":
		return nil
	case "eio":
		return nil
	case "oam":
		return nil
	}
	return errors.New("network_plane invalid")
}

func CheckVnicType(vnicType string) error {
	switch vnicType {
	case "normal":
		return nil
	case "direct":
		return nil
	case "physical":
		return nil
	}
	return errors.New("vnic_type invalid")
}

type OseClient struct {
	// Header *http.Header
	ServerURL string
	OseToken  string
}

var GetFunc func(url, token string) (resp *http.Response, err error)
var CloseFunc func(resp *http.Response) (err error)
var ReadAllFunc func(resp *http.Response) (body []byte, err error)

func OseGet(url, token string) (*http.Response, error) {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		klog.Errorf("new request error")
		return nil, fmt.Errorf("%v:new request error", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := client.Do(req)
	if err != nil {
		klog.Errorf("client do error")
		return nil, fmt.Errorf("%v:client do error", err)
	}
	return resp, nil

}

const MaxRetryTimesForHTTPReq int = 6

func Get(url, token string) (*http.Response, error) {
	var resp *http.Response
	var err error
	for i := 1; i < MaxRetryTimesForHTTPReq; i++ {
		resp, err = OseGet(url, token)
		if err != nil {
			time.Sleep(time.Second * time.Duration(i))
			continue
		}
		return resp, nil
	}
	return resp, err
}

func Close(resp *http.Response) error {
	return resp.Body.Close()
}
func ReadAll(resp *http.Response) ([]byte, error) {
	return ioutil.ReadAll(resp.Body)
}

func init() {
	GetFunc = Get
	CloseFunc = Close
	ReadAllFunc = ReadAll
}

func (self *OseClient) Get(url, token string) (int, []byte, error) {
	resp, err := GetFunc(url, token)
	if err != nil {
		klog.Errorf("k8s get podinfo error! -%v", err)
		return 444, nil, err
	}
	defer CloseFunc(resp)
	body, _ := ReadAllFunc(resp)
	return resp.StatusCode, body, nil
}

func (self *OseClient) GetPodInfo(url string) (int, *jason.Object, error) {
	klog.Info("GetPodInfo:req url :", url)
	statusCode, podinfo, err := self.Get(url, self.OseToken)
	if err != nil {
		klog.Errorf("GetPodInfo:Http Get return err:%v", err)
		return statusCode, nil, fmt.Errorf("%v:Get-Podinfo-return-http-error", err)
	}
	podjson, err := jason.NewObjectFromBytes(podinfo)
	if err != nil {
		klog.Errorf("GetPodInfo:NewObjectFromBytes return err:%v", err)
		return statusCode, nil, fmt.Errorf("%v:Get-Podinfo-return-new-object-from-bytes-error", err)
	}
	klog.Infof("GetPodInfo:Get pod info from ose ,the podjson is ", podjson)
	getpodfromvnpm, getpodfromvnpmerr := podjson.GetString("error")
	if getpodfromvnpmerr == nil {
		klog.Error("GetPodInfo:Request from ose success, but the response is error,the response is :", getpodfromvnpm)
		return statusCode, nil, errors.New("get-Podinfo-return-podjson-has-error-value")
	}
	return statusCode, podjson, nil
}

func (self *OseClient) GetPod(podNs, podName string) (int, *jason.Object, error) {
	url := self.ServerURL + "/api/v1/namespaces/" + podNs + "/pods/" + podName
	statusCode, podJason, err := self.GetPodInfo(url)
	if err != nil {
		klog.Error("GetPodFromOse:GetPodInfo error!", err)
	}
	return statusCode, podJason, err
}

func (self *OseClient) GetNodeInfo(nodeIP string) (*jason.Object, error) {
	nodeURL := self.ServerURL + "/api/v1/nodes"
	nodeList, err := self.GetNodeList(nodeURL)
	if err != nil {
		return nil, err
	}
	for _, node := range nodeList {
		addresses, errAddrs := node.GetObjectArray("status", "addresses")
		if errAddrs != nil {
			klog.Errorf("GetNodeInfo:GetNodeInfoAddr err! %v", errAddrs)
			continue
		}
		for _, addr := range addresses {
			typeString, _ := addr.GetString("type")
			addrString, _ := addr.GetString("address")
			if typeString == "InternalIP" && addrString == nodeIP {
				klog.Infof("GetNodeInfo:Get node info from ose ,the nodejson is ", node)
				return node, nil
			}
		}
	}
	klog.Errorf("GetNodeInfo:Not match err nodeIp is %v", nodeIP)
	return nil, errors.New("not match nodeIp")
}

func (self *OseClient) GetNodeList(url string) ([]*jason.Object, error) {
	_, nodeListByte, err := self.Get(url, self.OseToken)
	if err != nil {
		klog.Errorf("GetNodeList:http get return err:%v", err)
		return nil, fmt.Errorf("%v:Get-Nodelist-return-http-get-error", err)
	}
	nodeListjson, err := jason.NewObjectFromBytes(nodeListByte)
	if err != nil {
		klog.Errorf("GetNodeList:NewObjectFromBytes return err:%v", err)
		return nil, fmt.Errorf("%v:Get-Nodelist-return-new object from bytes-error", err)
	}
	nodeList, errList := nodeListjson.GetObjectArray("items")
	if errList != nil {
		klog.Errorf("GetNodeList:GetObjectArray error! %v", errList)
		return nil, errList
	}
	return nodeList, nil
}

func (self *OseClient) GetPodList(url string) ([]*jason.Object, error) {
	klog.Info("GetPodList:req url :", url)
	_, podListByte, _ := self.Get(url, self.OseToken)
	podListjson, _ := jason.NewObjectFromBytes(podListByte)
	checkStatus, errCheck := podListjson.GetString("error")
	if errCheck == nil {
		klog.Error("GetPodList:Request from ose success, but the response is error,the response is :", checkStatus)
		return nil, errors.New("get-Podlist-return-has-error-value")
	}
	podList, errList := podListjson.GetObjectArray("items")
	if errList != nil {
		klog.Errorf("GetPodList:GetObjectArray error! %v", errList)
		return nil, errList
	}
	return podList, nil
}

func (self *OseClient) GetPodsOfNode(nodeID string) ([]*jason.Object, error) {
	url := self.ServerURL + "/api/v1/pods"
	podList, err := self.GetPodList(url)
	if err != nil {
		return nil, err
	}
	podListOfNode := []*jason.Object{}
	for _, pod := range podList {
		hostIP, _ := pod.GetString("status", "hostIP")
		if hostIP != nodeID {
			continue
		}
		podListOfNode = append(podListOfNode, pod)
	}
	return podListOfNode, nil
}
