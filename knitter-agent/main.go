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
//
package main

import (
	"flag"

	"github.com/astaxie/beego"

	"github.com/ZTE/Knitter/pkg/version"

	"github.com/ZTE/Knitter/knitter-agent/controllers"
	_ "github.com/ZTE/Knitter/knitter-agent/docs"
	"github.com/ZTE/Knitter/knitter-agent/domain/port-recycle"
	"github.com/ZTE/Knitter/knitter-agent/infra"
	_ "github.com/ZTE/Knitter/knitter-agent/routers"
	"github.com/ZTE/Knitter/pkg/klog"
	"github.com/antonholmquist/jason"
)

func setDevCfg(o *jason.Object) {
	flag, err := o.GetBoolean("recover_flag")
	if err == nil && flag == false {
		klog.Info("setDevCfg: recover_flag is false")
		infra.RecoverFlag = false
	} else {
		klog.Info("setDevCfgdock: recover_flag is true")
		infra.RecoverFlag = true
	}
}

func main() {
	klog.ConfigLog("/root/info/logs/nwnode")
	cfgPath := flag.String("cfg", "", "config file path")
	flag.Parse()

	if version.HasVerFlag() {
		version.PrintVersion()
		return
	}
	cfgObj, err := version.GetConfObject(*cfgPath, "dev")
	if err == nil {
		setDevCfg(cfgObj)
	}

	klog.Infof("RecoverFlag: %v", infra.RecoverFlag)

	confObjBym11, err := version.GetConfObject(*cfgPath, "agent")
	if err != nil {
		klog.Error("ParseInputParams failed, err: ", err.Error())
		return
	}
	err = controllers.Init(confObjBym11)
	if err != nil {
		klog.Errorf("controllers.Init failed, err: %v", err)
		//todo plat event
		return
	}

	if beego.BConfig.RunMode == "dev" {
		beego.BConfig.WebConfig.DirectoryIndex = true
		beego.BConfig.WebConfig.StaticDir["/swagger"] = "swagger"
	}

	go portrecycle.RecycleResourseByTimer()

	beego.Run()
}
