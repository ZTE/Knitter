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

package knittermgrrole

import (
	"encoding/json"
	"fmt"
	"github.com/ZTE/Knitter/knitter-agent/domain/cni"
	"github.com/ZTE/Knitter/knitter-agent/err-obj"
	"github.com/ZTE/Knitter/pkg/inter-cmpt/agt-mgr"
	"github.com/ZTE/Knitter/pkg/klog"
)

type ReportPodRole struct {
}

func (this ReportPodRole) Post(agentPodReq *agtmgr.AgentPodReq) error {
	klog.Infof("ReportPodRole.Post START,agentPodReq is [%v]", agentPodReq)

	agtCtx := cni.GetGlobalContext()
	url := agtCtx.Mc.GetReportPodURL(agentPodReq.TenantId, agentPodReq.PodName)

	requestBody, err := json.Marshal(agentPodReq)
	if err != nil {
		klog.Errorf("ReportPodRole.Post:json.Marshal(agentPodReq) error, error is %v", err)
		return err
	}

	klog.Infof("ReportPodRole.Post: url is [%s],requestBody is [%s]", url, string(requestBody))
	statusCode, rspByte, err := agtCtx.Mc.PostBytes(url, requestBody)
	if err != nil {
		klog.Errorf("ReportPodRole.Post: agtCtx.Mc.PostBytes(%s, %s) error, "+
			"error is [%v]", url, string(requestBody), err)
		return err
	}

	if statusCode < 200 || statusCode >= 300 {
		klog.Errorf("ReportPodRole.Post:(posturl: %v, %v) ok, "+
			"but return status code is %v, response is [%v]",
			url, string(requestBody), string(rspByte), statusCode)
		return fmt.Errorf("ReportPodRole.Post: masterClient.Post return status code:%v error msg:%v",
			statusCode, errobj.GetErrMsg(rspByte))
	}

	klog.Infof("ReportPodRole.Post SUCC, statusCode is [%v], response is [%v]", statusCode, string(rspByte))
	return nil
}

func (this ReportPodRole) Delete(agentPodReq *agtmgr.AgentPodReq) error {
	klog.Infof("ReportPodRole.Delete START,agentPodReq is [%v]", agentPodReq)

	agtCtx := cni.GetGlobalContext()
	url := agtCtx.Mc.GetReportPodURL(agentPodReq.TenantId, agentPodReq.PodName)

	rspByte, statusCode, err := agtCtx.Mc.Delete(url)
	if err != nil {
		klog.Errorf("ReportPodRole.Delete: agtCtx.Mc.Delete(%s) error, "+
			"error is [%v]", url, err)
		return err
	}

	if statusCode < 200 || statusCode >= 300 {
		klog.Errorf("ReportPodRole.Delete:(posturl: %v) ok, "+
			"but return status code is %v, resp is [%v]", url, statusCode, string(rspByte))
		return fmt.Errorf("ReportPodRole.Delete: masterClient.Post return status code:%v error msg:%v",
			statusCode, errobj.GetErrMsg(rspByte))
	}

	klog.Infof("ReportPodRole.Post SUCC, statusCode is [%v], response is [%v]", statusCode, string(rspByte))
	return nil

}
