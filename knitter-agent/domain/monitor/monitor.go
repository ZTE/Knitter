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

package monitor

import (
	"errors"
	"fmt"
	"github.com/ZTE/Knitter/knitter-agent/domain/const-value"
	"github.com/ZTE/Knitter/pkg/klog"
	"github.com/antonholmquist/jason"
	"io/ioutil"
	"net/http"
	"time"
)

type MonitorClient struct {
	// Header *http.Header
	URLKnitterMonitor string
	VMID              string
}

func (self *MonitorClient) GetPodURL(podNs, podName string) string {
	return self.URLKnitterMonitor + "/pods" + "/" + podNs + "/" + podName
}

func (m *MonitorClient) GetPod(podNs, PodName string) (int, []byte, error) {
	postURL := m.GetPodURL(podNs, PodName)
	statuscode, netByte, err := m.Get(postURL)
	if err != nil {
		klog.Errorf("GetPod:  m.Get(postURL:[%v]) err, statuscode is [%v],error is[%v]  ", postURL, statuscode, err)
		return 0, nil, err
	}

	return statuscode, netByte, nil
}

func (m *MonitorClient) Get(postURL string) (int, []byte, error) {
	resp, err := HTTPGet(postURL)
	if err != nil {
		klog.Errorf("MonitorClient get error! -%v", err)
		return 444, nil, err
	}
	defer HTTPClose(resp)
	body, _ := HTTPReadAll(resp)
	statuscode := resp.StatusCode
	return statuscode, body, nil
}

//todo refactor : move http method to pkg

var HTTPClose = func(resp *http.Response) error {
	return resp.Body.Close()
}
var HTTPReadAll = func(resp *http.Response) ([]byte, error) {
	return ioutil.ReadAll(resp.Body)
}

const MaxRetryTimesForHTTPReq int = 6

var HTTPGet = func(url string) (*http.Response, error) {
	var resp *http.Response
	var err error
	for i := 1; i < MaxRetryTimesForHTTPReq; i++ {
		resp, err = get(url)
		if err == nil {
			return resp, nil
		}
		time.Sleep(time.Second * time.Duration(i))
	}
	return resp, err
}

func get(url string) (*http.Response, error) {
	client := &http.Client{Timeout: constvalue.HTTPDefaultTimeoutInSec * time.Second}
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		klog.Error("Get: http NewRequest error: ", err.Error())
		return nil, fmt.Errorf("%v:http NewRequest error", err)
	}

	response, err := client.Do(request)
	if err != nil {
		klog.Error("Get: client.Do error: ", err.Error())
		return nil, fmt.Errorf("%v:http client.Do error", err)
	}
	return response, err
}

func (m *MonitorClient) InitClient(cfg *jason.Object) error {
	monitorURL, _ := cfg.GetString("monitor", "url")
	klog.Infof("cfg.GetString monitor url: %v", monitorURL)
	if monitorURL == "" {
		klog.Errorf("InitClient:monitor url is null")
		return errors.New("monitor url is null")
	}
	m.URLKnitterMonitor = monitorURL

	vmid, err := cfg.GetString("host", "vm_id")
	if err != nil {
		klog.Errorf("InitClient:cfg.GetString no vmid! -%v", err)
		return fmt.Errorf("%v:InitClient:cfg.GetString no vmid", err)
	}
	m.VMID = vmid
	return nil
}
