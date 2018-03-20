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

package test

import (
	"bytes"
	"encoding/json"
	"github.com/ZTE/Knitter/pkg/klog"
	"github.com/antonholmquist/jason"
	"github.com/astaxie/beego"
	"github.com/coreos/etcd/client"
	"github.com/golang/mock/gomock"
	"net/http"
	"net/http/httptest"
	"strings"
)

// TestGet is a sample to run an endpoint test
func getRouterUUID(str string) string {
	klog.Info("Router-Json:", str)
	objPort, err := jason.NewObjectFromBytes([]byte(str))
	if err != nil {
		klog.Error("Unmarshal Router Error:", err)
		return ""
	}
	id, err := objPort.GetString("router", "router_id")
	if err != nil {
		return ""
	}
	return id
}

func getNetworkUUID(str string) string {
	klog.Info("Network-Json:", str)
	objPort, err := jason.NewObjectFromBytes([]byte(str))
	if err != nil {
		klog.Error("Unmarshal Network Error:", err)
		return ""
	}
	id, err := objPort.GetString("network", "network_id")
	if err != nil {
		return ""
	}
	return id
}

func getPortUUID(str string) string {
	klog.Info("Port-Json:", str)
	objPort, err := jason.NewObjectFromBytes([]byte(str))
	if err != nil {
		klog.Error("Unmarshal Port Error:", err)
		return ""
	}
	id, err := objPort.GetString("port", "port_id")
	if err != nil {
		return ""
	}
	return id
}

func deleteMultiPorts(str string) *httptest.ResponseRecorder {
	klog.Info("Ports-Json:", str)
	var rtnRsp httptest.ResponseRecorder
	obj, _ := jason.NewObjectFromBytes([]byte(str))
	objPortList, _ := obj.GetObjectArray("ports")
	for _, objPort := range objPortList {
		id, _ := objPort.GetString("port_id")
		klog.Info("Delete Port:" + id)
		rsp := DeletePort(id)
		if rsp.Code != 200 {
			return rsp
		}
		rtnRsp = *rsp
	}

	return &rtnRsp
}

func ConfigOpenStack(cfg string) *httptest.ResponseRecorder {
	klog.Info("ConfigOpenStack ---->\n", cfg)
	r, _ := http.NewRequest("POST", "/nw/v1/tenants/paas-admin/configration", strings.NewReader(cfg))
	w := httptest.NewRecorder()
	beego.BeeApp.Handlers.ServeHTTP(w, r)

	klog.Info("ConfigOpenStack", "Code[%d]\n%s", w.Code, w.Body.String())
	return w
}

func CreateNetwork(cfg string) *httptest.ResponseRecorder {
	/*	netConfig := `{
		"network":{
			"name":"m11-create-delete",
			"gateway":"192.168.199.1",
			"cidr":"192.168.199.0/24"
			}
		}`*/
	klog.Info("CreateNetwork ---->\n", cfg)
	r, _ := http.NewRequest("POST", "/nw/v1/tenants/normal-tenant-id/networks", strings.NewReader(cfg))
	w := httptest.NewRecorder()
	beego.BeeApp.Handlers.ServeHTTP(w, r)

	klog.Info("CreateNetwork", "Code[%d]\n%s", w.Code, w.Body.String())
	return w
}

func CreateIPGroup(cfg string) *httptest.ResponseRecorder {
	klog.Info("CreateIPGroup ---->\n", cfg)
	r, _ := http.NewRequest("POST", "/nw/v1/tenants/paas-admin/ipgroups", strings.NewReader(cfg))
	w := httptest.NewRecorder()
	beego.BeeApp.Handlers.ServeHTTP(w, r)

	klog.Info("CreateIPGroup", "Code[%d]\n%s", w.Code, w.Body.String())
	return w
}

func CreateNetworkAdmin(cfg string) *httptest.ResponseRecorder {
	klog.Info("CreateNetwork ---->\n", cfg)
	r, _ := http.NewRequest("POST", "/nw/v1/tenants/admin/networks", strings.NewReader(cfg))
	w := httptest.NewRecorder()
	beego.BeeApp.Handlers.ServeHTTP(w, r)

	klog.Info("CreateNetwork", "Code[%d]\n%s", w.Code, w.Body.String())
	return w
}

func UpdateTenant(cfg string) *httptest.ResponseRecorder {
	klog.Info("UpdateTenant ---->\n", cfg)
	r, _ := http.NewRequest("PUT", "/nw/v1/tenants/paas-admin/quota?value="+cfg, nil)
	w := httptest.NewRecorder()
	beego.BeeApp.Handlers.ServeHTTP(w, r)

	klog.Info("UpdateTenant", "Code[%d]\n%s", w.Code, w.Body.String())
	return w
}

func GetTenant(id string) *httptest.ResponseRecorder {
	klog.Info("GetTenant ---->\n", id)
	r, _ := http.NewRequest("GET", "/nw/v1/tenants/"+id, nil)
	w := httptest.NewRecorder()
	beego.BeeApp.Handlers.ServeHTTP(w, r)

	klog.Info("GetTenant", "Code[%d]\n%s", w.Code, w.Body.String())
	return w
}

func GetTenants() *httptest.ResponseRecorder {
	klog.Info("GetTenants ---->\n")
	r, _ := http.NewRequest("GET", "/nw/v1/tenants", nil)
	w := httptest.NewRecorder()
	beego.BeeApp.Handlers.ServeHTTP(w, r)

	klog.Info("GetTenants", "Code[%d]\n%s", w.Code, w.Body.String())
	return w
}

func DeleteTenant() *httptest.ResponseRecorder {
	klog.Info("DeleteTenant ---->\n")
	r, _ := http.NewRequest("DELETE", "/nw/v1/tenants/paas-admin", nil)
	w := httptest.NewRecorder()
	beego.BeeApp.Handlers.ServeHTTP(w, r)

	klog.Info("DeleteTenant", "Code[%d]\n%s", w.Code, w.Body.String())
	return w
}

func DeleteNetwork(id string) *httptest.ResponseRecorder {
	klog.Info("DelNetwork ---->\n", id)
	r, _ := http.NewRequest("DELETE", "/nw/v1/tenants/paas-admin/networks/"+id, nil)
	w := httptest.NewRecorder()
	beego.BeeApp.Handlers.ServeHTTP(w, r)

	klog.Info("DelNetwork", "Code[%d]\n%s", w.Code, w.Body.String())
	return w
}

func GetAllNetwork() *httptest.ResponseRecorder {
	klog.Info("GetAllNetwork ---->\n")
	r, _ := http.NewRequest("GET", "/nw/v1/tenants/paas-admin/networks/", nil)
	w := httptest.NewRecorder()
	beego.BeeApp.Handlers.ServeHTTP(w, r)

	klog.Info("GetAllNetwork", "Code[%d]\n%s", w.Code, w.Body.String())
	return w
}

func GetAllIPGroup() *httptest.ResponseRecorder {
	klog.Info("GetAllIPGroup ---->\n")
	r, _ := http.NewRequest("GET", "/nw/v1/tenants/paas-admin/ipgroups", nil)
	w := httptest.NewRecorder()
	beego.BeeApp.Handlers.ServeHTTP(w, r)

	klog.Info("GetAllIPGroup", "Code[%d]\n%s", w.Code, w.Body.String())
	return w
}

func GetAllPublicNetwork() *httptest.ResponseRecorder {
	klog.Info("GetAllNetwork ---->\n")
	r, _ := http.NewRequest("GET", "/nw/v1/tenants/paas-admin/networks?public=true", nil)
	w := httptest.NewRecorder()
	beego.BeeApp.Handlers.ServeHTTP(w, r)

	klog.Info("GetAllNetwork", "Code[%d]\n%s", w.Code, w.Body.String())
	return w
}

func GetAllUserNetwork() *httptest.ResponseRecorder {
	klog.Info("GetAllNetwork ---->\n")
	r, _ := http.NewRequest("GET", "/nw/v1/tenants/paas-admin/networks?all=true", nil)
	w := httptest.NewRecorder()
	beego.BeeApp.Handlers.ServeHTTP(w, r)

	klog.Info("GetAllNetwork", "Code[%d]\n%s", w.Code, w.Body.String())
	return w
}

func GetAllNetworkByName(name string) *httptest.ResponseRecorder {
	klog.Info("GetAllNetwork ---->\n")
	r, _ := http.NewRequest("GET", "/nw/v1/tenants/paas-admin/networks/?name= "+name, nil)
	w := httptest.NewRecorder()
	beego.BeeApp.Handlers.ServeHTTP(w, r)

	klog.Info("GetAllNetwork", "Code[%d]\n%s", w.Code, w.Body.String())
	return w
}

func GetNetwork(id string) *httptest.ResponseRecorder {
	klog.Info("GetNetwork ---->\n", id)
	r, _ := http.NewRequest("GET", "/nw/v1/tenants/paas-admin/networks/"+id, nil)
	w := httptest.NewRecorder()
	beego.BeeApp.Handlers.ServeHTTP(w, r)

	klog.Info("GetNetwork", "Code[%d]\n%s", w.Code, w.Body.String())
	return w
}

func GetIPGroup(id string) *httptest.ResponseRecorder {
	klog.Info("GetNetwork ---->\n", id)
	r, _ := http.NewRequest("GET", "/nw/v1/tenants/paas-admin/ipgroups/"+id, nil)
	w := httptest.NewRecorder()
	beego.BeeApp.Handlers.ServeHTTP(w, r)

	klog.Info("GetNetwork", "Code[%d]\n%s", w.Code, w.Body.String())
	return w
}

func CreateRouter(cfg string) *httptest.ResponseRecorder {
	/*	netConfig := `{
		"router":{
			"name":"auto-test-create-router-should-be-delete"
			}
		}`*/
	klog.Info("CreateRouter ---->\n", cfg)
	r, _ := http.NewRequest("POST", "/nw/v1/tenants/paas-admin/routers", strings.NewReader(cfg))
	w := httptest.NewRecorder()
	beego.BeeApp.Handlers.ServeHTTP(w, r)

	klog.Info("CreateRouter", "Code[%d]\n%s", w.Code, w.Body.String())
	return w
}

func UpdateRouter(id, cfg string) *httptest.ResponseRecorder {
	/*	netConfig := `{
		"router":{
			"name":"auto-test-create-router-should-be-delete"
			}
		}`*/
	klog.Info("UpdateRouter ---->\n", id, "----with----", cfg)
	r, _ := http.NewRequest("PUT", "/nw/v1/tenants/paas-admin/routers/"+id, strings.NewReader(cfg))
	w := httptest.NewRecorder()
	beego.BeeApp.Handlers.ServeHTTP(w, r)

	klog.Info("UpdateRouter", "Code[%d]\n%s", w.Code, w.Body.String())
	return w
}

func UpdateIPGroup(id, cfg string) *httptest.ResponseRecorder {
	klog.Info("UpdateIPGroup ---->\n", id, "----with----", cfg)
	r, _ := http.NewRequest("PUT", "/nw/v1/tenants/paas-admin/ipgroups/"+id, strings.NewReader(cfg))
	w := httptest.NewRecorder()
	beego.BeeApp.Handlers.ServeHTTP(w, r)

	klog.Info("UpdateIPGroup", "Code[%d]\n%s", w.Code, w.Body.String())
	return w
}

func DeleteRouter(id string) *httptest.ResponseRecorder {
	klog.Info("DeleteRouter ---->\n", id)
	r, _ := http.NewRequest("DELETE", "/nw/v1/tenants/paas-admin/routers/"+id, nil)
	w := httptest.NewRecorder()
	beego.BeeApp.Handlers.ServeHTTP(w, r)

	klog.Info("DeleteRouter", "Code[%d]\n%s", w.Code, w.Body.String())
	return w
}

func GetAllRouter() *httptest.ResponseRecorder {
	klog.Info("GetAllRouter ---->\n")
	r, _ := http.NewRequest("GET", "/nw/v1/tenants/paas-admin/routers/", nil)
	w := httptest.NewRecorder()
	beego.BeeApp.Handlers.ServeHTTP(w, r)

	klog.Info("GetAllRouter", "Code[%d]\n%s", w.Code, w.Body.String())
	return w
}

func GetRouter(id string) *httptest.ResponseRecorder {
	klog.Info("GetRouter ---->\n", id)
	r, _ := http.NewRequest("GET", "/nw/v1/tenants/paas-admin/routers/"+id, nil)
	w := httptest.NewRecorder()
	beego.BeeApp.Handlers.ServeHTTP(w, r)

	klog.Info("GetRouter", "Code[%d]\n%s", w.Code, w.Body.String())
	return w
}

func AttachRouter(routerID, networkID string) *httptest.ResponseRecorder {
	klog.Info("AttachRouter ---->\n", networkID, "----to------", routerID)
	nwJSON := string(`{"network":{"network_id":"` + networkID + `"}}`)
	r, _ := http.NewRequest("PUT", "/nw/v1/tenants/paas-admin/routers/"+
		routerID+"/attach", strings.NewReader(nwJSON))
	w := httptest.NewRecorder()
	beego.BeeApp.Handlers.ServeHTTP(w, r)

	klog.Info("AttachRouter", "Code[%d]\n%s", w.Code, w.Body.String())
	return w
}

func AttachRouter403(routerID, networkID string) *httptest.ResponseRecorder {
	klog.Info("AttachRouter ---->\n", networkID, "----to------", routerID)
	nwJSON := string(`{"network":{"network_id":"` + networkID + `"}`)
	r, _ := http.NewRequest("PUT", "/nw/v1/tenants/paas-admin/routers/"+
		routerID+"/attach", strings.NewReader(nwJSON))
	w := httptest.NewRecorder()
	beego.BeeApp.Handlers.ServeHTTP(w, r)

	klog.Info("AttachRouter", "Code[%d]\n%s", w.Code, w.Body.String())
	return w
}

func DetachRouter(routerID, networkID string) *httptest.ResponseRecorder {
	klog.Info("DetachRouter ---->\n", networkID, "----from------", routerID)
	nwJSON := string(`{"network":{"network_id":"` + networkID + `"}}`)
	r, _ := http.NewRequest("PUT", "/nw/v1/tenants/paas-admin/routers/"+
		routerID+"/detach", strings.NewReader(nwJSON))
	w := httptest.NewRecorder()
	beego.BeeApp.Handlers.ServeHTTP(w, r)

	klog.Info("DetachRouter", "Code[%d]\n%s", w.Code, w.Body.String())
	return w
}

func DetachRouter403(routerID, networkID string) *httptest.ResponseRecorder {
	klog.Info("DetachRouter ---->\n", networkID, "----from------", routerID)
	nwJSON := string(`{"network":{"network_id":"` + networkID + `"}`)
	r, _ := http.NewRequest("PUT", "/nw/v1/tenants/paas-admin/routers/"+
		routerID+"/detach", strings.NewReader(nwJSON))
	w := httptest.NewRecorder()
	beego.BeeApp.Handlers.ServeHTTP(w, r)

	klog.Info("DetachRouter", "Code[%d]\n%s", w.Code, w.Body.String())
	return w
}

func CreatePort(cfg string) *httptest.ResponseRecorder {
	klog.Info("CreatePort ---->\n", cfg)
	r, _ := http.NewRequest("POST", "/nw/v1/tenants/paas-admin/ports", strings.NewReader(cfg))
	w := httptest.NewRecorder()
	beego.BeeApp.Handlers.ServeHTTP(w, r)

	klog.Info("CreatePort", "Code[%d]\n%s", w.Code, w.Body.String())
	return w
}
func DeletePort(id string) *httptest.ResponseRecorder {
	klog.Info("DeletePort ---->\n", id)
	r, _ := http.NewRequest("DELETE", "/nw/v1/tenants/paas-admin/ports/"+id, nil)
	w := httptest.NewRecorder()
	beego.BeeApp.Handlers.ServeHTTP(w, r)

	klog.Info("DeletePort", "Code[%d]\n%s", w.Code, w.Body.String())
	return w
}

func DeleteIPGroup(id string) *httptest.ResponseRecorder {
	klog.Info("DeleteIPGroup ---->\n", id)
	r, _ := http.NewRequest("DELETE", "/nw/v1/tenants/paas-admin/networks/net0/ipgroups/"+id, nil)
	w := httptest.NewRecorder()
	beego.BeeApp.Handlers.ServeHTTP(w, r)

	klog.Info("DeleteIPGroup", "Code[%d]\n%s", w.Code, w.Body.String())
	return w
}

func GetAllPort() *httptest.ResponseRecorder {
	klog.Info("GetAllPort ---->\n")
	r, _ := http.NewRequest("GET", "/nw/v1/tenants/paas-admin/ports/", nil)
	w := httptest.NewRecorder()
	beego.BeeApp.Handlers.ServeHTTP(w, r)

	klog.Info("GetAllPort", "Code[%d]\n%s", w.Code, w.Body.String())
	return w
}

func GetPort(id string) *httptest.ResponseRecorder {
	klog.Info("GetPort ---->\n", id)
	r, _ := http.NewRequest("GET", "/nw/v1/tenants/paas-admin/ports/"+id, nil)
	w := httptest.NewRecorder()
	beego.BeeApp.Handlers.ServeHTTP(w, r)

	klog.Info("GetPort", "Code[%d]\n%s", w.Code, w.Body.String())
	return w
}

func GetAllPod() *httptest.ResponseRecorder {
	klog.Info("GetAllPod ---->\n")
	r, _ := http.NewRequest("GET", "/nw/v1/tenants/paas-admin/pods/", nil)
	w := httptest.NewRecorder()
	beego.BeeApp.Handlers.ServeHTTP(w, r)

	klog.Info("GetAllPod", "Code[%d]\n%s", w.Code, w.Body.String())
	return w
}

func GetAllPodByNetID() *httptest.ResponseRecorder {
	klog.Info("GetAllPod ---->\n")
	r, _ := http.NewRequest("GET", "/nw/v1/tenants/paas-admin/pods?network_id=networkid1", nil)
	w := httptest.NewRecorder()
	beego.BeeApp.Handlers.ServeHTTP(w, r)

	klog.Info("GetAllPod", "Code[%d]\n%s", w.Code, w.Body.String())
	return w
}

func GetPod(name string) *httptest.ResponseRecorder {
	klog.Info("GetPod ---->\n", name)
	r, _ := http.NewRequest("GET", "/nw/v1/tenants/paas-admin/pods/"+name, nil)
	w := httptest.NewRecorder()
	beego.BeeApp.Handlers.ServeHTTP(w, r)

	klog.Info("GetPod", "Code[%d]\n%s", w.Code, w.Body.String())
	return w
}

func APICreatePort(cfg string) *httptest.ResponseRecorder {
	klog.Info("CreatePort ---->\n", cfg)
	r, _ := http.NewRequest("POST",
		"/api/v1/tenants/paas-admin/port?req_id=xxxx-xxxx",
		strings.NewReader(cfg))
	w := httptest.NewRecorder()
	beego.BeeApp.Handlers.ServeHTTP(w, r)

	klog.Info("CreatePort", "Code[%d]\n%s", w.Code, w.Body.String())
	return w
}
func APIDeletePort(id string) *httptest.ResponseRecorder {
	klog.Info("DeletePort ---->\n", id)
	r, _ := http.NewRequest("DELETE",
		"/api/v1/tenants/paas-admin/port/"+id+"?req_id=xxxx-xxxx",
		nil)
	w := httptest.NewRecorder()
	beego.BeeApp.Handlers.ServeHTTP(w, r)

	klog.Info("DeletePort", "Code[%d]\n%s", w.Code, w.Body.String())
	return w
}

func APIAttach(id, vm string) *httptest.ResponseRecorder {
	klog.Info("DeletePort ---->\n", id)
	r, _ := http.NewRequest("POST",
		"/api/v1/tenants/paas-admin/port/"+
			vm+"/"+id+"?req_id=xxxx-xxxx",
		nil)
	w := httptest.NewRecorder()
	beego.BeeApp.Handlers.ServeHTTP(w, r)

	klog.Info("DeletePort", "Code[%d]\n%s", w.Code, w.Body.String())
	return w
}

func APIDetach(id, vm string) *httptest.ResponseRecorder {
	klog.Info("ApiDetach ---->\n", id, "----", vm)
	r, _ := http.NewRequest("DELETE",
		"/api/v1/tenants/paas-admin/port/"+
			vm+"/"+id+"?req_id=xxxx-xxxx",
		nil)
	w := httptest.NewRecorder()
	beego.BeeApp.Handlers.ServeHTTP(w, r)

	klog.Info("ApiDetach", "Code[%d]\n%s", w.Code, w.Body.String())
	return w
}

func APIGetNetExt(networkNames []string, needProvider bool) *httptest.ResponseRecorder {
	klog.Info("ApiGetNetExt ---->\n", networkNames)
	type NetworkAttr struct {
		NetworkNames []string `json:"network_names"`
	}
	networkAttr := &NetworkAttr{NetworkNames: networkNames}
	networkAttrByte, err := json.Marshal(networkAttr)
	if err != nil {
		klog.Info("json.Marshal(networkAttr) error!")
	}
	var url = ""
	if needProvider {
		url = "/api/v1/tenants/paas-admin/networks?provider=true"
	} else {
		url = "/api/v1/tenants/paas-admin/networks?provider=false"
	}
	r, _ := http.NewRequest("POST",
		url, strings.NewReader(string(networkAttrByte)))

	w := httptest.NewRecorder()
	beego.BeeApp.Handlers.ServeHTTP(w, r)

	klog.Info("ApiGetNetExt", "Code[%d]\n%s", w.Code, w.Body.String())
	return w
}

func apiGetDcVnis(dcID string) *httptest.ResponseRecorder {
	klog.Info("apiGetDcVnis ---->\n", dcID)
	r, _ := http.NewRequest("GET", "/api/v1/tenants/admin/dcvni/"+dcID, nil)
	w := httptest.NewRecorder()
	beego.BeeApp.Handlers.ServeHTTP(w, r)

	klog.Info("ApiGetNetExt", "Code[%d]\n%s", w.Code, w.Body.String())
	return w
}

func GetType() *httptest.ResponseRecorder {
	klog.Info("GetType ---->\n")
	r, _ := http.NewRequest("GET", "/nw/v1/tenants/paas-admin/networkmanagertype", nil)
	w := httptest.NewRecorder()
	beego.BeeApp.Handlers.ServeHTTP(w, r)

	klog.Info("GetType", "Code[%d]\n%s", w.Code, w.Body.String())
	return w
}

func GetHealth() *httptest.ResponseRecorder {
	klog.Info("GetHealth ---->\n")
	r, _ := http.NewRequest("GET", "/nw/v1/tenants/admin/health", nil)
	w := httptest.NewRecorder()
	beego.BeeApp.Handlers.ServeHTTP(w, r)

	klog.Info("GetHealth", "Code[%d]\n%s", w.Code, w.Body.String())
	return w
}

func GetVni(networkID string) *httptest.ResponseRecorder {
	klog.Info("GetVni ---->\n", networkID)
	r, _ := http.NewRequest("GET", "/api/v1/tenants/paas-admin/vni/"+networkID, nil)
	w := httptest.NewRecorder()
	beego.BeeApp.Handlers.ServeHTTP(w, r)

	klog.Info("GetVni", "Code[%d]\n%s", w.Code, w.Body.String())
	return w
}

func APIGetNetAttrs(networkName string) *httptest.ResponseRecorder {
	klog.Info("ApiGetNetAttrs ---->\n", networkName)
	type NetworkAttr struct {
		NetworkNames []string `json:"network_names"`
	}
	url := "/api/v1/tenants/paas-admin/network/" + networkName
	r, _ := http.NewRequest("GET", url, nil)

	w := httptest.NewRecorder()
	beego.BeeApp.Handlers.ServeHTTP(w, r)

	klog.Info("ApiGetNetAttrs", "Code[%d]\n%s", w.Code, w.Body.String())
	return w
}

func APIDbrestore(body []byte) *httptest.ResponseRecorder {
	klog.Info("ApiDbrestore ---->\n")
	url := "/nw/v1/tenants/admin/dbrestore"
	r, _ := http.NewRequest("POST", url, bytes.NewReader(body))

	w := httptest.NewRecorder()
	beego.BeeApp.Handlers.ServeHTTP(w, r)

	klog.Info("ApiDbrestore", "Code[%d]\n%s", w.Code, w.Body.String())
	return w
}

func APIRestore(body []byte) *httptest.ResponseRecorder {
	klog.Info("ApiRestore ---->\n")
	url := "/nw/v1/tenants/admin/restore"
	r, _ := http.NewRequest("POST", url, bytes.NewReader(body))

	w := httptest.NewRecorder()
	beego.BeeApp.Handlers.ServeHTTP(w, r)

	klog.Info("ApiRestore", "Code[%d]\n%s", w.Code, w.Body.String())
	return w
}

func MockPaasAdminCheck(mockDB *MockDbAccessor) {
	tenant := string(`{
        		"tenant_uuid": "paas-admin",
        		"tenant_name": "paas-admin",
        		"net_quota": 2,
        		"net_number": 1
    	}`)

	var tenantList1 []*client.Node
	tenantNode := client.Node{Key: "/paasnet/tenants/paas-admin", Value: "tenant-info"}
	tenantList1 = append(tenantList1, &tenantNode)
	gomock.InOrder(
		mockDB.EXPECT().ReadDir(gomock.Any()).Return(tenantList1, nil),
		mockDB.EXPECT().ReadLeaf(gomock.Any()).Return(tenant, nil),
	)
}

func MockTenantCheck(mockDB *MockDbAccessor, tenantID string) {
	tenant := string(`{
        		"tenant_uuid": "paas-admin",
        		"tenant_name": "paas-admin",
        		"net_quota": 2,
        		"net_number": 1
    	}`)

	tenant = strings.Replace(tenant, "paas-admin", tenantID, -1)

	var tenantList1 []*client.Node
	tenantNode := client.Node{Key: "/paasnet/tenants/" + tenantID, Value: "tenant-info"}
	tenantList1 = append(tenantList1, &tenantNode)
	gomock.InOrder(
		mockDB.EXPECT().ReadDir(gomock.Any()).Return(tenantList1, nil),
		mockDB.EXPECT().ReadLeaf(gomock.Any()).Return(tenant, nil),
	)
}

func APIUptdateDefaultPhysnet(body []byte) *httptest.ResponseRecorder {
	klog.Info("UptdateDefaultPhysnet ---->\n")
	url := "/nw/v1/conf/default_physnet"
	r, _ := http.NewRequest("POST", url, bytes.NewReader(body))
	w := httptest.NewRecorder()
	beego.BeeApp.Handlers.ServeHTTP(w, r)

	klog.Info("UptdateDefaultPhysnet", "Code[%d]\n%s", w.Code, w.Body.String())
	return w
}

func APIGetDefaultPhysnet() *httptest.ResponseRecorder {
	klog.Info("GetDefaultPhysnet ---->\n")
	url := "/nw/v1/conf/default_physnet"
	r, _ := http.NewRequest("GET", url, nil)
	w := httptest.NewRecorder()
	beego.BeeApp.Handlers.ServeHTTP(w, r)

	klog.Info("GetDefaultPhysnet", "Code[%d]\n%s", w.Code, w.Body.String())
	return w
}

func APIPostInitConfiguration(cfg string) *httptest.ResponseRecorder {
	klog.Info("APIPostInitConfiguration ---->\n", cfg)
	r, _ := http.NewRequest("POST", "/nw/v1/configuration", strings.NewReader(cfg))
	w := httptest.NewRecorder()
	beego.BeeApp.Handlers.ServeHTTP(w, r)

	klog.Info("APIPostInitConfiguration", "Code[%d]\n%s", w.Code, w.Body.String())
	return w
}
