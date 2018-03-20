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

package scheduler

import (
	"github.com/ZTE/Knitter/knitter-agent/context"
	"github.com/ZTE/Knitter/knitter-agent/domain/bind"
	"github.com/ZTE/Knitter/knitter-agent/domain/const-value"
	"github.com/ZTE/Knitter/knitter-agent/domain/object/knitter-obj"
	"github.com/ZTE/Knitter/knitter-agent/infra"
	"github.com/ZTE/Knitter/knitter-agent/trans"
	"github.com/ZTE/Knitter/knitter-agent/trans/general-mode"
	"github.com/ZTE/Knitter/pkg/klog"
	"github.com/ZTE/Knitter/pkg/trans-dsl"
	"github.com/antonholmquist/jason"
)

func Init(o *jason.Object) error {
	setRunningMode(o)
	err := modelscom.InitContext4Agent(o)
	if err != nil {
		klog.Error("InitEnv4Manger error, exit agent now!")
		return err
	}
	return nil
}

func setRunningMode(cfg *jason.Object) {
	infra.SetMode(infra.OverlayMode)
}

func Attach(reqBody []byte) (err error) {
	knitterObj, err := knitterobj.CreateKnitterObj(reqBody)
	if err != nil {
		klog.Errorf("Attach : knitterobj.CreateKnitterObj(reqBody: %s) "+
			"error, error is %v", string(reqBody), err)
		return err
	}
	err = generalModeAttachWithDDDTrans(knitterObj, reqBody)
	return err

}

func Detach(reqBody []byte) (err error) {
	knitterObj, err := knitterobj.CreateKnitterObj(reqBody)
	if err != nil {
		klog.Errorf("Detach : knitterobj.CreateKnitterObj(reqBody: %s) "+
			"error, error is %v", string(reqBody), err)
		return err
	}
	if knitterObj.PodProtectionRole.TryAddDetachingTag() != true {
		return nil
	}

	err = generalModeDetachWithDDDTrans(knitterObj, reqBody)
	return err

}

func generalModeAttachWithDDDTrans(knitterObj *knitterobj.KnitterObj, reqBody []byte) (err error) {
	transInfo := &transdsl.TransInfo{AppInfo: &context.KnitterInfo{ReqBody: reqBody, Nics: make([]bind.Dpdknic, 0), IsAttachOrDetachFlag: true}}
	defer func() {
		if p := recover(); p != nil {
			context.RecoverErr(p, &err, "generalModeAttachWithDDDTrans")
		}
		if transInfo.AppInfo.(*context.KnitterInfo).ChanFlag {
			transInfo.AppInfo.(*context.KnitterInfo).Chan <- 1
		}
	}()

	transInfo.AppInfo.(*context.KnitterInfo).KnitterObj = knitterObj

	trans := trans.NewGeneralModeAttachTrans()
	err = trans.Exec(transInfo)
	if err != nil {
		trans.RollBack(transInfo)
	}
	return err
}

func generalModeDetachWithDDDTrans(knitterObj *knitterobj.KnitterObj, reqBody []byte) (err error) {
	transInfo := &transdsl.TransInfo{AppInfo: &context.KnitterInfo{ReqBody: reqBody, Nics: make([]bind.Dpdknic, 0), IsAttachOrDetachFlag: false}}
	defer func() {
		if p := recover(); p != nil {
			context.RecoverErr(p, &err, "generalModeDetachWithDDDTrans")
		}
		knitterInfo := transInfo.AppInfo.(*context.KnitterInfo)
		if knitterInfo.ChanFlag {
			knitterInfo.Chan <- 1
		}
	}()

	transInfo.AppInfo.(*context.KnitterInfo).KnitterObj = knitterObj

	trans := trans.NewGeneralModeDetachTrans()
	err = trans.Exec(transInfo)
	if err != nil {
		trans.RollBack(transInfo)
	}
	return err
}

// only for vm
func IsPciAllocOver(hostType string) bool {
	num := bind.SumOfVirtioNetPci()
	//"0000:00:1f.0" is max pci serial number in vm
	pciID, _ := bind.GetPciID("0000:00:1f.0")
	if num > constvalue.MaxPciDeviceNum && pciID != "" {
		return true
	}
	return false
}

type HealthObj struct {
	Level int64 `json:"health_level"`
}

func (self *HealthObj) GetHealthLevel() int64 {
	//PNWLOG.Info("Health Level return  0!")
	return 0
}
