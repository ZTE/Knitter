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

package knittermonitorrole

import (
	"encoding/json"
	"errors"
	"github.com/ZTE/Knitter/knitter-agent/domain/cni"
	"github.com/ZTE/Knitter/knitter-agent/domain/const-value"
	"github.com/ZTE/Knitter/knitter-agent/domain/monitor"
	"github.com/ZTE/Knitter/pkg/klog"
	"net/http"
	"time"
)

type GetPodRole struct {
}

func (gpr *GetPodRole) GetPod(podNs, podName string) (*monitor.Pod, error) {
	agtCtx := cni.GetGlobalContext()
	var i int
	var podByte []byte
	var statuscode int
	var err error
	for i = 0; i < constvalue.AgentGetPodRetryTimes; i++ {

		time.Sleep(constvalue.AgentGetPodWaitSecond * time.Second)
		statuscode, podByte, err = agtCtx.MtrC.GetPod(podNs, podName)
		klog.Infof("GetPodRole.GetPod:agtCtx.MtrC.GetPod(podNs, podName),"+
			" statuscode is [%v],podbyte is [%v],", statuscode, string(podByte))
		if err != nil {
			klog.Errorf("agtCtx.MtrC.GetPod(podNs:[%v], podName:[%v]) err, error is [%v]", podNs, podName, err)
			return nil, err
		}
		if statuscode == http.StatusNotFound {
			klog.Debugf("GetPodRole.GetPodretry times [%v]", i)
			continue
		}
		if statuscode != http.StatusOK && statuscode != http.StatusNotFound {
			klog.Errorf("m.Get(postURL) err return statuscode:%v is not 200", statuscode)
			return nil, errors.New("GetPod return statuscode is not 200")
		}
		if statuscode == http.StatusOK {
			klog.Infof(" agtCtx.MtrC.GetPod(podNs, podName) succ")
			break
		}
	}

	if i == constvalue.AgentGetPodRetryTimes {
		err := errors.New("exceed the munber of retries")
		klog.Errorf("get pod retry [%v] err, error is %v ", i, err)
		return nil, err
	}
	pod := &monitor.Pod{}
	err = json.Unmarshal(podByte, pod)
	if err != nil {
		klog.Errorf("GetPod Json unmarshal error! -%v", err)
		return nil, err
	}
	return pod, nil
}
