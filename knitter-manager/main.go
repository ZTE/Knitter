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

package main

import (
	"flag"
	"github.com/ZTE/Knitter/knitter-manager/controllers"
	"github.com/ZTE/Knitter/knitter-manager/models"
	_ "github.com/ZTE/Knitter/knitter-manager/routers"
	"github.com/ZTE/Knitter/pkg/klog"
	"github.com/ZTE/Knitter/pkg/version"
	"github.com/astaxie/beego"
)

func main() {
	//version.ConfigLog("/root/info/logs/nwmaster/", "knitter-manager")
	cfgPath := flag.String("cfg", "", "server_info config file path")
	flag.Parse()

	if version.HasVerFlag() {
		version.PrintVersion()
		return
	}

	klog.ConfigLog("/root/info/logs/nwmaster/")

	confObj, err := version.GetConfObject(*cfgPath, "manager")
	if err != nil {
		klog.Errorf("ParseInputParams failed, err: %v", err)
		return
	}
	//todo err event
	err = models.InitEnv4Manger(confObj)

	if err != nil {
		klog.Error("InitEnv4Manger error, exit manager now!")
		return
	}
	klog.Warning("PaaS-NetWork-Manager run Model ---->", beego.BConfig.RunMode)
	if beego.BConfig.RunMode == "dev" {
		beego.BConfig.WebConfig.DirectoryIndex = true
		beego.BConfig.WebConfig.StaticDir["/swagger"] = "swagger"
	}
	err = controllers.CreateDefaultNetwork()
	if err != nil {
		klog.Errorf("controllers.CreateDefaulNetwork() err, error is [%v]", err)
		return
	}
	beego.Run()
}
