package routers

import (
	"github.com/astaxie/beego"

	"github.com/ZTE/Knitter/knitter-monitor/controllers"
)

func init() {
	beego.Router("/api/v1/pods/:podns/:podname", &controllers.PodController{})
}
