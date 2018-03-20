package main

import (
	"flag"

	"github.com/astaxie/beego"

	"github.com/ZTE/Knitter/knitter-monitor/infra"
	_ "github.com/ZTE/Knitter/knitter-monitor/routers"
	"github.com/ZTE/Knitter/knitter-monitor/services"
	"github.com/ZTE/Knitter/pkg/klog"
	"github.com/ZTE/Knitter/pkg/version"
)

func main() {
	//todo modify the configuration file loading method

	cfgPath := flag.String("cfg", "", "config file path")
	kubeconfig := flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	flag.Parse()

	confObj, err := version.GetConfObject(*cfgPath, "monitor")
	if err != nil {
		klog.Error("ParseInputParams failed, err: ", err.Error())
		return
	}

	err = services.InitConfigurations4Monitor(confObj, *kubeconfig)
	if err != nil {
		klog.Errorf("services.InitConfigurations4Monitor err, error is [%v]", err)
		return
	}

	err = infra.InitKubernetesClientset(*kubeconfig)
	createPortForPodController, err := services.NewCreatePortForPodController()
	if err != nil {
		klog.Errorf("services.NewCreatePortForPodController(*kubeconfig:[%v]) err, error is [%v]", *kubeconfig, err)
		return
	}
	var stopCh <-chan struct{}
	go createPortForPodController.Run(1, stopCh)
	beego.Run()
	klog.Infof("main END")
}
