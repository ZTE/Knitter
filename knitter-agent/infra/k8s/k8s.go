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

package k8s

import (
	"errors"
	"fmt"
	"github.com/ZTE/Knitter/pkg/klog"
	"github.com/antonholmquist/jason"
	"io/ioutil"
	"net/http"
	"time"
)

var GetFunc func(url string) (resp *http.Response, err error)
var CloseFunc func(resp *http.Response) (err error)
var ReadAllFunc func(resp *http.Response) (body []byte, err error)

func Get(url string) (*http.Response, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		klog.Errorf("new request error")
		return nil, fmt.Errorf("%v:new request error", err)
	}
	return client.Do(req)
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

type K8sClient struct {
	// Header *http.Header
	ServerURL string
}

func (self *K8sClient) GetPod(podNs, podName string) (int, *jason.Object, error) {
	url := self.ServerURL + "/api/v1/namespaces/" + podNs + "/pods/" + podName
	statusCode, podJason, err := self.GetPodInfo(url)
	if err != nil {
		klog.Errorf("GetPodFromOse:GetPodInfo error! %v", err)
		return statusCode, nil, err
	}
	return statusCode, podJason, err
}

func (self *K8sClient) GetPodInfo(url string) (int, *jason.Object, error) {
	klog.Info("GetPodInfo:req url :", url)
	statusCode, podinfo, err := self.Get(url)
	if err != nil || podinfo == nil {
		return statusCode, nil, err
	}
	klog.Info("GetPodInfo:", string(podinfo))
	podjson, err := jason.NewObjectFromBytes(podinfo)
	if err != nil {
		return statusCode, nil, err
	}
	klog.Infof("GetPodInfo:Get pod info from k8s ,the podjson is ", podjson)
	return statusCode, podjson, nil
}

func (self *K8sClient) GetK8sRspPodInfo(url string) (int, []byte, error) {
	resp, errReq := GetFunc(url)
	if errReq != nil {
		klog.Errorf("k8s-get-podinfo-rsp-error:%v", errReq)
		return 444, nil, errReq
	}
	defer CloseFunc(resp)
	RspBody, errBody := ReadAllFunc(resp)
	if errBody != nil {
		klog.Errorf("k8s-get-podinfo-rsp-body-error:%v", errBody)
		klog.Warning("k8s get podinfo error! -%v", errBody)
		return resp.StatusCode, nil, errBody
	}
	return resp.StatusCode, RspBody, nil
}

func (self *K8sClient) Get(url string) (int, []byte, error) {
	var errReq error
	var statusCode int = 444
	const MaxRetryTimes int = 5
	for i := 1; i < MaxRetryTimes; i++ {
		klog.Info("GetPodInfo--[", i, "]-->URL:", url)
		var RspBody []byte
		statusCode, RspBody, errReq = self.GetK8sRspPodInfo(url)
		if errReq != nil {
			time.Sleep(time.Second * time.Duration(i))
			continue
		}
		return statusCode, RspBody, nil
	}
	return statusCode, nil, errReq
}

func (self *K8sClient) GetPodList(url string) ([]*jason.Object, error) {
	klog.Info("GetPodList:req url :", url)
	_, podListByte, err := self.Get(url)
	if err != nil {
		klog.Errorf("GetPodList:Http Get return err:%v", err)
		return nil, err
	}
	podListjson, err := jason.NewObjectFromBytes(podListByte)
	if err != nil {
		klog.Error("GetPodList:Get pod list from ose error! code:", err)
		return nil, fmt.Errorf("%v:Get-Podlist-return-code-error", err)
	}
	checkStatus, errCheck := podListjson.GetString("error")
	if errCheck == nil {
		klog.Error("GetPodList:Request from ose success, but the response is error,the response is :", checkStatus)
		return nil, errors.New("get-podlist-return-has-error-value")
	}
	podList, errList := podListjson.GetObjectArray("items")
	if errList != nil {
		klog.Errorf("GetPodList:GetObjectArray error! %v", errList)
		return nil, errList
	}
	return podList, nil
}

func (self *K8sClient) GetNodeList(url string) ([]*jason.Object, error) {
	resp, errReq := GetFunc(url)
	if errReq != nil {
		klog.Errorf("k8s-get-nodelistinfo-rsp-error:%v", errReq)
		return nil, errReq
	}
	defer CloseFunc(resp)
	if resp.StatusCode != 200 {
		klog.Errorf("k8s-get-nodelistinfo-rsp-code-error:%v", resp.StatusCode)
		return nil, errors.New("k8s-rsp-code-error")
	}
	RspBody, errBody := ReadAllFunc(resp)
	if errBody != nil {
		klog.Errorf("k8s-get-nodelist-rsp-body-error:%v", errBody)
		klog.Warning("k8s get nodelist error! -%v", errBody)
		return nil, errBody
	}
	nodeListjson, err := jason.NewObjectFromBytes(RspBody)
	if err != nil {
		klog.Errorf("GetNodeList:Get node list from ose error! code:%v", err)
		return nil, fmt.Errorf("%v:Get-Nodelist-return-code-error", err)
	}
	nodeList, errList := nodeListjson.GetObjectArray("items")
	if errList != nil {
		klog.Errorf("GetNodeList:GetObjectArray error! %v", errList)
		return nil, errList
	}
	return nodeList, nil
}

func (self *K8sClient) GetPodsOfNode(nodeID string) ([]*jason.Object, error) {
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

func (self *K8sClient) GetNodeInfo(nodeIP string) (*jason.Object, error) {
	url := self.ServerURL + "/api/v1/nodes"
	nodeList, err := self.GetNodeList(url)
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
	return nil, errors.New("not match nodeIP")
}
