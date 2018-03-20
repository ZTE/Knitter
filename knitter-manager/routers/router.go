// @APIVersion 1.0.0
// @Title beego Test API
// @Description beego has a very cool tools to autogenerate documents for your API
// @Contact astaxie@gmail.com
// @TermsOfServiceUrl http://beego.me/
// @License Apache 2.0
// @LicenseUrl http://www.apache.org/licenses/LICENSE-2.0.html
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

package routers

import (
	"github.com/ZTE/Knitter/knitter-manager/controllers"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/context"

	"github.com/ZTE/Knitter/knitter-manager/models"
	"github.com/ZTE/Knitter/pkg/klog"
)

func init() {
	beego.Router("/nw/v1/tenants/:user/networks", &controllers.NetworkController{}, "post:Post")
	beego.Router("/nw/v1/tenants/:user/networks/:network_id", &controllers.NetworkController{}, "get:Get")
	beego.Router("/nw/v1/tenants/:user/networks", &controllers.NetworkController{}, "get:GetAll")
	beego.Router("/nw/v1/tenants/:user/networks/:network_id", &controllers.NetworkController{}, "delete:Delete")

	beego.Router("/nw/v1/tenants/:user/routers", &controllers.RouterController{}, "post:Post")
	beego.Router("/nw/v1/tenants/:user/routers/:router_id", &controllers.RouterController{}, "put:Update")
	beego.Router("/nw/v1/tenants/:user/routers/:router_id", &controllers.RouterController{}, "get:Get")
	beego.Router("/nw/v1/tenants/:user/routers", &controllers.RouterController{}, "get:GetAll")
	beego.Router("/nw/v1/tenants/:user/routers/:router_id", &controllers.RouterController{}, "delete:Delete")
	beego.Router("/nw/v1/tenants/:user/routers/:router_id/attach", &controllers.RouterController{}, "put:Attach")
	beego.Router("/nw/v1/tenants/:user/routers/:router_id/detach", &controllers.RouterController{}, "put:Detach")

	beego.Router("/nw/v1/tenants/:user/pods/:pod_name", &controllers.PodController{}, "get:Get")
	beego.Router("/nw/v1/tenants/:user/pods", &controllers.PodController{}, "get:GetAll")

	beego.Router("/nw/v1/tenants/:user/configurations", &controllers.CfgController{}, "post:Post")
	beego.Router("/nw/v1/tenants/:user/configration", &controllers.CfgController{}, "post:Post")

	beego.Router("/nw/v1/tenants/:user/networkmanagertype", &controllers.NetworkManagerController{}, "get:Get")

	beego.Router("/nw/v1/tenants/:user/ipgroups", &controllers.IPGroupController{}, "post:Post")
	beego.Router("/nw/v1/tenants/:user/ipgroups", &controllers.IPGroupController{}, "get:GetAll")
	beego.Router("/nw/v1/tenants/:user/ipgroups/:group", &controllers.IPGroupController{}, "get:Get")
	beego.Router("/nw/v1/tenants/:user/ipgroups/:group", &controllers.IPGroupController{}, "put:Put")
	beego.Router("/nw/v1/tenants/:user/ipgroups/:group", &controllers.IPGroupController{}, "delete:Delete")

	beego.Router("/nw/v1/tenants/admin/health", &controllers.HealthController{}, "get:Get")

	beego.Router("/nw/v1/tenants/:user", &controllers.TenantController{}, "get:Get")
	beego.Router("/nw/v1/tenants/:user", &controllers.TenantController{}, "delete:Delete")
	beego.Router("/nw/v1/tenants/:user", &controllers.TenantController{}, "post:Post")
	beego.Router("/nw/v1/tenants/:user/quota", &controllers.TenantController{}, "put:Update")
	beego.Router("/nw/v1/tenants", &controllers.TenantController{}, "get:GetAll")
	beego.Router("/nw/v1/tenants", &controllers.TenantController{}, "post:PostExclusive")

	beego.Router("/nw/v1/conf/default_physnet", &controllers.PhysnetController{}, "post:Update")
	beego.Router("/nw/v1/conf/default_physnet", &controllers.PhysnetController{}, "get:Get")

	beego.Router("/nw/v1/configuration", &controllers.InitCfgController{}, "post:Post")

	beego.Router("/api/v1/tenants/:user/port", &controllers.CniMasterPortController{}, "post:Post")
	beego.Router("/api/v1/tenants/:user/port/:port_id", &controllers.CniMasterPortController{}, "delete:Delete")
	beego.Router("/api/v1/tenants/:user/port/:vm_id/:port_id", &controllers.CniMasterPortController{}, "post:Attach")
	beego.Router("/api/v1/tenants/:user/port/:vm_id/:port_id", &controllers.CniMasterPortController{}, "delete:Detach")

	beego.Router("/api/v1/tenants/:user/interface", &controllers.PhyPortController{}, "post:Post")
	beego.Router("/api/v1/tenants/:user/interface/:vm_id/:port_id", &controllers.PhyPortController{}, "delete:Delete")

	beego.Router("/api/v1/tenants/:user/vni/:network_id", &controllers.VniController{}, "get:Get")

	beego.Router("/api/v1/tenants/:user/network", &controllers.PaasNetController{}, "post:Post")
	beego.Router("/api/v1/tenants/:user/network/:network_name", &controllers.PaasNetController{}, "get:Get")

	beego.Router("/api/v1/tenants/:user/networks", &controllers.PaasNetController{}, "post:Post")
	beego.Router("/api/v1/tenants/:user/networks/:network_name", &controllers.PaasNetController{}, "get:Get")

	beego.Router("/api/v1/tenants/:user/sync/:internal_ip", &controllers.SyncController{}, "get:Get")

	beego.Router("/api/v1/tenants/admin/health", &controllers.HealthController{}, "get:Get")

	beego.Router("/api/v1/loglevel/:log_level", &controllers.LogController{}, "put:Put")

	beego.InsertFilter("/nw/v1/tenants/:user/*", beego.BeforeExec, models.BeforeExecTenantCheck, false)
	beego.InsertFilter("/*", beego.BeforeStatic, func(ctx *context.Context) {
		klog.Infof("receive http request: [%v]", ctx.Request.URL.RequestURI())
	}, false)
}
