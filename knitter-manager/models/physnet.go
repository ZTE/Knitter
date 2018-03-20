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

package models

import (
	"errors"
	"github.com/ZTE/Knitter/knitter-manager/const-value"
	"github.com/ZTE/Knitter/knitter-manager/iaas"
	"github.com/ZTE/Knitter/pkg/iaas-accessor"
	"github.com/ZTE/Knitter/pkg/klog"
	"github.com/ZTE/Knitter/pkg/noauth_openstack"
	"math/rand"
	"net/http"
	"strconv"
	"time"
)

type Physnet struct {
	DefaultPhysnet string `json:"default_physnet"`
}

var UpdateDefaultPhysnet = func(phy string) error {
	value := iaas.GetIaaS(constvalue.DefaultIaasTenantID)
	actValue, ok := value.(*noauth_openstack.NoauthOpenStack)
	if !ok {
		klog.Errorf("UpdateDefaultPhysnet Err: GetIaaS Err")
		return BuildErrWithCode(http.StatusUnauthorized, errors.New("getIaaS error"))
	}
	actValue.NeutronConf.ProviderConf.PhyscialNetwork = phy
	errChechk := CheckPhysnet(actValue)
	if errChechk != nil {
		klog.Errorf("UpdateDefaultPhysnet Err: CheckPhysnet[%v] Err[%v]", phy, errChechk)
		return BuildErrWithCode(http.StatusBadRequest, errors.New("check physnet failed"))
	}
	iaas.SetIaaS(constvalue.DefaultIaasTenantID, actValue)
	iaas.SaveDefaultPhysnet(actValue.NeutronConf.ProviderConf.PhyscialNetwork)
	return nil
}

var CheckCreateNetInDefaultPhysnet = func(noAuthOpn iaasaccessor.IaaS) error {
	randNum := rand.Intn(100)
	testNetName := "test" + strconv.Itoa(randNum) + "physnet"
	net, err := noAuthOpn.CreateNetwork(testNetName)
	if err != nil {
		klog.Errorf("CheckPhysnet Err: CreateNetwork Err: %v", err)
		return err
	}
	errDel := noAuthOpn.DeleteNetwork(net.Id)
	if errDel != nil {
		klog.Errorf("CheckPhysnet Err: DeleteNetwork Err: %v", errDel)
		return errDel
	}
	return nil
}

var CheckPhysnet = func(opn iaasaccessor.IaaS) error {
	var errPhysnet error
	for i := 0; i < 3; i++ {
		errPhysnet = CheckCreateNetInDefaultPhysnet(opn)
		if errPhysnet != nil {
			klog.Errorf("InitNoauthOpenStack Err: CheckPhysnet Err: %v", errPhysnet)
			time.Sleep(5 * time.Second)
			continue
		}
		return nil
	}
	return errPhysnet
}
