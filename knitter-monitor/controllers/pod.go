package controllers

import (
	"github.com/astaxie/beego"

	"github.com/ZTE/Knitter/knitter-monitor/apps"
	"github.com/ZTE/Knitter/knitter-monitor/err-obj"
	"github.com/ZTE/Knitter/pkg/klog"
)

type PodController struct {
	beego.Controller
}

// Title Get
// Description find pod by pod_name
// Param	pod_name	path 	string	true		"the pod_name you want to get"
// Param	pod_ns	path 	string	true		"the pod_ns you want to get"
// Success 200 {object}
// Failure 404 :
// router /api/v1/pods/:podns/:podname [get]

func (pc *PodController) Get() {
	klog.Infof("RECV agent get pod start  ")
	podns := pc.Ctx.Input.Param(":podns")
	podName := pc.Ctx.Input.Param(":podname")

	podForAgent, err := apps.GetPodApp().Get(podns, podName)
	if err != nil {
		klog.Errorf("apps.GetPodApp().Get(podns :[%v],podName:[%v]) err, error is [%v]", podns, podName, err)
		if errobj.IsKeyNotFoundError(err) {
			errobj.NotfoundErr404(&pc.Controller, err)
			return
		}
		errobj.Err500(&pc.Controller, err)
		return
	}
	pc.Data["json"] = podForAgent
	klog.Infof("Ageng get pod end, pod for agent is [%+v]", podForAgent)

	pc.ServeJSON()
}
