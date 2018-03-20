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

package manager

import (
	"encoding/json"
	//	"github.com/mitchellh/mapstructure"
	"bytes"
	"errors"
	"fmt"
	"github.com/ZTE/Knitter/knitter-agent/domain/const-value"
	"github.com/ZTE/Knitter/knitter-agent/err-obj"
	"github.com/ZTE/Knitter/pkg/inter-cmpt/agt-mgr"
	"github.com/ZTE/Knitter/pkg/inter-cmpt/mgr-agt"
	"github.com/ZTE/Knitter/pkg/klog"
	"github.com/antonholmquist/jason"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"runtime/debug"
	"strings"
	"time"
)

type IP struct {
	SubnetID string `mapstructure:"subnet_id" json:"subnet_id"`
	Address  string `mapstructure:"ip_address" json:"ip_address,omitempty"`
}

//paasnw etcd store port structure
type Interface struct {
	NetworkID string `json:"networkid"`
	SubnetID  string `json:"subnetid"`
	PortID    string `json:"portid"`
	Name      string `json:"name"`
	IP        string `json:"ip"`
	Mac       string `json:"mac"`
	TenantID  string `json:"tenantid"`
	Ownertype string `json:"ownertype"`
	Ownerid   string `json:"ownerid"`
	Porttype  string `json:"porttype"`
	Businfo   string `json:"businfo"`
	NetPlane  string `json:"netplane"`
}

//cni master return port structure

type ManagerClient struct {
	// Header *http.Header
	URLKnitterManager string
	VMID              string
}

func (self *ManagerClient) GetTenantURL(tenantID string) string {
	return self.URLKnitterManager + "/tenants/" + tenantID
}

func (self *ManagerClient) GetHealthURL() string {
	return self.URLKnitterManager + "/tenants/admin" + "/health"
}

func (self *ManagerClient) GetNetworksURL(tenantID string) string {
	return self.GetTenantURL(tenantID) + "/networks"
}

func (self *ManagerClient) GetNetworkURL(tenantID, netName string) string {
	return self.GetTenantURL(tenantID) + "/network/" + netName
}

func (self *ManagerClient) GetSegmentIDURLByID(tenantID, networkID string) string {
	return self.GetTenantURL(tenantID) + "/vni/" + networkID
}

func (self *ManagerClient) GetSegmentIDURLByName(tenantID, networkName string) string {
	return self.GetTenantURL(tenantID) + "/vlanid/" + networkName
}

func (self *ManagerClient) GetAttachURL(tenantID, portID string) string {
	return self.GetTenantURL(tenantID) + "/port/" + self.VMID + "/" + portID
}

func (self *ManagerClient) GetVnicInterfaceCreateURL(tenantID string) string {
	return self.GetTenantURL(tenantID) + "/interface"
}

func (self *ManagerClient) GetVnicInterfaceDeleteURL(tenantID, vmID, portID string) string {
	return self.GetTenantURL(tenantID) + "/interface/" + vmID + "/" + portID
}

func (self *ManagerClient) GetDetachURL(tenantID, portID string) string {
	return self.GetAttachURL(tenantID, portID)
}

func (self *ManagerClient) GetCreatePortURL(tenantID string) string {
	return self.GetTenantURL(tenantID) + "/port/"
}

func (self *ManagerClient) GetDeletePortURL(tenantID, portID string) string {
	return self.GetCreatePortURL(tenantID) + portID
}

func (self *ManagerClient) GetSyncInGenModURL(tenantID string) string {
	return self.GetTenantURL(tenantID) + "/sync"
}

func (self *ManagerClient) GetDcVnisURL(tenantID, dcID string) string {
	return self.GetTenantURL(tenantID) + "/dcvni/" + dcID
}

func (self *ManagerClient) GetReportPodURL(tenantID string, podName string) string {
	return self.GetTenantURL(tenantID) + "/pods/" + podName
}

type PaaSNwGUID string

func NewGUID(podID string) string {
	return podID
}

func MakeURLReqIDSuffix(reqID string) string {
	return "?" + "req_id=" + reqID
}

const MaxRetryTimesForHTTPReq int = 6

var HTTPPost = func(url string, bodyType string, body io.Reader) (*http.Response, error) {
	var resp *http.Response
	var err error
	for i := 1; i < MaxRetryTimesForHTTPReq; i++ {
		resp, err = post(url, bodyType, body)
		if err == nil {
			return resp, nil
		}
		time.Sleep(time.Second * time.Duration(i))
	}
	return resp, err
}

//func Post(url string, contentType string, body interface{}) ([]byte, error) {
func post(url string, contentType string, body io.Reader) (*http.Response, error) {
	client := &http.Client{Timeout: constvalue.HTTPDefaultTimeoutInSec * time.Second}
	request, err := http.NewRequest("POST", url, body)
	if err != nil {
		klog.Error("Post: http NewRequest error: ", err.Error())
		return nil, fmt.Errorf("%v:http NewRequest error", err)
	}

	request.Header.Set("Content-Type", contentType)
	response, err := client.Do(request)
	if err != nil {
		klog.Error("Post: client.Do error: ", err.Error())
		return nil, errors.New("http client.Do error")
	}

	return response, nil
}

var HTTPGet = func(url string) (*http.Response, error) {
	var resp *http.Response
	var err error
	for i := 1; i < MaxRetryTimesForHTTPReq; i++ {
		resp, err = get(url)
		if err == nil {
			return resp, nil
		}
		time.Sleep(time.Second * time.Duration(i))
	}
	return resp, err
}

func get(url string) (*http.Response, error) {
	client := &http.Client{Timeout: constvalue.HTTPDefaultTimeoutInSec * time.Second}
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		klog.Error("Get: http NewRequest error: ", err.Error())
		return nil, fmt.Errorf("%v:http NewRequest error", err)
	}

	response, err := client.Do(request)
	if err != nil {
		klog.Error("Get: client.Do error: ", err.Error())
		return nil, fmt.Errorf("%v:http client.Do error", err)
	}
	return response, err
}

var Delete = func(url string) (resp *http.Response, err error) {
	client := &http.Client{Timeout: constvalue.HTTPDefaultTimeoutInSec * time.Second}
	req, errreq := http.NewRequest("DELETE", url, nil)
	if errreq != nil {
		klog.Errorf("##Delete2Master http.NewRequest(DELETE,url,nil) error! -%v", errreq)
		return nil, errreq
	}

	resp, errresp := client.Do(req)
	if errresp != nil {
		klog.Errorf("##Delete2Master http.DefaultClient.Do(req) error! -%v", errresp)
		return nil, errresp
	}
	klog.Infof("##temp##:HttpDelete success! errreq and errresp should nil")
	return resp, nil
}

var HTTPDelete = func(url string) (*http.Response, error) {
	var resp *http.Response
	var err error
	for i := 1; i < MaxRetryTimesForHTTPReq; i++ {
		resp, err = Delete(url)
		if err == nil {
			return resp, nil
		}
		time.Sleep(time.Second * time.Duration(i))
	}
	return resp, err
}
var HTTPClose = func(resp *http.Response) error {
	return resp.Body.Close()
}
var HTTPReadAll = func(resp *http.Response) ([]byte, error) {
	return ioutil.ReadAll(resp.Body)
}

func (m *ManagerClient) GetVMIDFromServerConf(ServerInfo []byte) (string, error) {
	serverinfojson, _ := jason.NewObjectFromBytes([]byte(ServerInfo))
	vmid, err := serverinfojson.GetString("host", "vm_id")
	if err != nil {
		klog.Errorf("GetVMIDFromServerConf:serverinfojson.GetString no vmid! -%v", err)
		return "", fmt.Errorf("%v:GetVMIDFromServerConf:serverinfojson.GetString no vmid", err)
	}
	return vmid, nil
}

func (m *ManagerClient) GetManagerURLFromK8SConf(serverconf []byte) (string, error) {
	serverconfjson, jsonerr := jason.NewObjectFromBytes([]byte(serverconf))
	if jsonerr != nil {
		klog.Errorf("GetCniMasterServerUrlFromK8SConf:jason.NewObjectFromBytes serverconf error! -%v", jsonerr)
		return "", fmt.Errorf("%v:GetCniMasterServerUrlFromK8SConf:jason.NewObjectFromBytes serverconf error", jsonerr)
	}
	klog.Infof("NewObjectFromBytes k8s url %v", serverconfjson)
	controlerURL, _ := serverconfjson.GetString("etcd", "url")
	klog.Infof("serverconfjson.GetString k8s url: %v", controlerURL)
	if controlerURL == "" {
		klog.Errorf("GetcontrolerUrlFromConfigure:k8s url is null")
		return "", errors.New("controlerUrl is null")
	}
	//cnimasterurl := strings.Replace(k8surl, ":8080", ":6001", -1)
	strArry := strings.Split(controlerURL, ":")
	managerURL := fmt.Sprintf("%v", strArry[0]) + ":" + fmt.Sprintf("%v", strArry[1]) + ":" + "9527"
	klog.Infof("new str:", managerURL)
	return managerURL, nil
}

func (m *ManagerClient) Init(serverconf []byte) error {
	managerURL, urlerr := m.GetManagerURLFromK8SConf(serverconf)
	if urlerr != nil {
		return fmt.Errorf("%v:K8SClient's init k8surl error", urlerr)
	}
	m.URLKnitterManager = managerURL
	klog.Infof("cni master url is :%v", m.URLKnitterManager)
	vmid, err := m.GetVMIDFromServerConf(serverconf)
	if err != nil {
		klog.Errorf("ManagerClient-Init:GetVMIDFromServerConf error! -%v", err)
		return fmt.Errorf("%v:ManagerClient-Init:GetVMIDFromServerConf error", err)
	}
	m.VMID = vmid
	return nil
}

func (m *ManagerClient) InitClient(cfg *jason.Object) error {
	managerURL, _ := cfg.GetString("manager", "url")
	klog.Infof("cfg.GetString manager url: %v", managerURL)
	if managerURL == "" {
		klog.Errorf("InitClient:manager url is null")
		return errors.New("manager url is null")
	}
	m.URLKnitterManager = managerURL

	vmid, err := cfg.GetString("host", "vm_id")
	if err != nil {
		klog.Errorf("InitClient:cfg.GetString no vmid! -%v", err)
		return fmt.Errorf("%v:InitClient:cfg.GetString no vmid", err)
	}
	m.VMID = vmid
	return nil
}

func (m *ManagerClient) Post(postURL string, postDict map[string]string) (int, []byte, error) {
	postValues := url.Values{}
	for postkey, postvalue := range postDict {
		postValues.Set(postkey, postvalue)
	}
	klog.Infof("postValues=%v", postValues)
	postDataStr := postValues.Encode()
	klog.Infof("postDatastr=%v", postDataStr)
	postDataByte := []byte(postDataStr)
	klog.Infof("postDataByte=%v", postDataByte)
	postBytesReader := bytes.NewReader(postDataByte)
	//bodyType := "application/x-www-form-urlencoded"
	bodyType := "application/json"
	resp, err := HTTPPost(postURL, bodyType, postBytesReader)
	if err != nil {
		klog.Errorf("ManagerClient post error! -%v", err)
		return 444, nil, err
	}
	defer HTTPClose(resp)
	body, _ := HTTPReadAll(resp)
	statuscode := resp.StatusCode
	return statuscode, body, nil
}

func (m *ManagerClient) Get(postURL string) (int, []byte, error) {
	resp, err := HTTPGet(postURL)
	if err != nil {
		klog.Errorf("masterclient post error! -%v", err)
		return 444, nil, err
	}
	defer HTTPClose(resp)
	body, _ := HTTPReadAll(resp)
	statuscode := resp.StatusCode
	return statuscode, body, nil
}

func (m *ManagerClient) PostBytes(postURL string, postData []byte) (int, []byte, error) {
	postBytesReader := bytes.NewReader(postData)

	contentType := "application/json"
	resp, err := HTTPPost(postURL, contentType, postBytesReader)
	klog.Infof("ManagerClient.PostBytes:postURL is [%v]", postURL)
	if err != nil {
		klog.Errorf("masterclient post error! -%v", err)
		return 444, nil, err
	}
	defer HTTPClose(resp)
	body, _ := HTTPReadAll(resp)
	statuscode := resp.StatusCode
	return statuscode, body, nil
}

func (m *ManagerClient) Delete(deleteURL string) (b []byte, statusCode int, e error) {
	resp, err := HTTPDelete(deleteURL)
	if err != nil {
		klog.Errorf("##masterclient delete http.DefaultClient.Do error! -%v", err)
		return nil, resp.StatusCode, err
	}
	defer HTTPClose(resp)
	body, _ := HTTPReadAll(resp)
	return body, resp.StatusCode, nil
}

func (p *Port) Extract(portbyte []byte, netplane, tenantID string) (e error) {
	defer func() {
		if err := recover(); err != nil {
			e = errors.New("panic")
			klog.Info("==cni master Extract pnic recover start!==")
			debug.PrintStack()
			klog.Info("==cni master Extract pnic recover end!==")
		}
	}()
	if portbyte == nil {
		klog.Infof("Get port form cni master error, The reason is the port info is null.")
		return errors.New("get port form cni master error, The reason is the port info is null")
	}
	portjson, _ := jason.NewObjectFromBytes(portbyte)
	name, errname := portjson.GetString("port", "name")
	if errname != nil {
		klog.Infof("Get port form cni master error, The reason is the port's name null")
		return fmt.Errorf("%v:Get port form cni master error, The reason is the port's name null", errname)
	}
	p.Name = name
	klog.Infof("port name is %v", p.Name)

	networkid, errnetid := portjson.GetString("port", "network_id")
	if errnetid != nil {
		klog.Infof("Get port form cni master error, The reason is the port's networkid is null")
		return fmt.Errorf("%v:Get port form cni master error, The reason is the port's networkid is null", errnetid)
	}
	p.NetworkID = networkid
	klog.Infof("port network id is %v ", p.NetworkID)

	p.TenantID = tenantID
	klog.Infof("port tenant id is %v", p.TenantID)
	mac, errmac := portjson.GetString("port", "mac_address")
	if errmac != nil {
		klog.Infof("Get port form cni master error, The reason is the port's mac is null")
		return fmt.Errorf("%v:Get port form cni master error, The reason is the port's mac is null", errmac)
	}
	p.MACAddress = mac
	klog.Infof("port mac is %v", p.MACAddress)
	gatewayip, errgateway := portjson.GetString("port", "gateway_ip")
	if errgateway != nil {
		klog.Infof("Get port form cni master error, The reason is the port's gateway is null")
		return fmt.Errorf("%v:Get port form cni master error, The reason is the port's gateway is null", errgateway)
	}
	p.GatewayIP = gatewayip
	klog.Infof("port gateway is %v ", p.GatewayIP)
	cidr, errcidr := portjson.GetString("port", "cidr")
	if errcidr != nil {
		klog.Infof("Get port form cni master error, The reason is the port's cidr is null")
		return fmt.Errorf("%v:Get port form cni master error, The reason is the port's cidr is null", errcidr)
	}
	p.CIDR = cidr
	klog.Infof("port cidr is %v ", p.CIDR)
	id, errid := portjson.GetString("port", "id")
	if errid != nil {
		klog.Infof("Get port form cni master error, The reason is the port's id null")
		return fmt.Errorf("%v:Get port form cni master error, The reason is the port's id null", errid)
	}
	p.ID = id
	klog.Infof("port id is %v", p.ID)
	ips := []IP{}
	ip := IP{}
	fixip, errfixip := portjson.GetObjectArray("port", "fixed_ips")
	if errfixip != nil {
		klog.Errorf("Get port form cni master error, The reason is the port's fixed_ip is null")
		return fmt.Errorf("%v:Get port form cni master error, The reason is the port's fixed_ip is null", errfixip)
	}
	subnet, errsubnet := fixip[0].GetString("subnet_id")
	if errsubnet != nil {
		klog.Infof("Get port form cni master error, The reason is the port->fixed_ips's subnet_id is null")
		return fmt.Errorf("%v:Get port form cni master error, The reason is the port->fixed_ips's subnet_id is null", errsubnet)
	}
	ip.SubnetID = subnet
	ipaddr, erripaddr := fixip[0].GetString("ip_address")
	if erripaddr != nil {
		klog.Infof("Get port form cni master error, The reason is the port->fixed_ips's ip is null")
		return fmt.Errorf("%v:Get port form cni master error, The reason is the port->fixed_ips's ip is null", erripaddr)
	}
	ip.Address = ipaddr
	ips = append(ips, ip)
	p.FixedIPs = ips
	klog.Infof("port fixedips->ip is %v", p.FixedIPs[0].Address)
	klog.Infof("port fixedips->subnet is %v", p.FixedIPs[0].SubnetID)

	return nil
}

func (m *ManagerClient) CreateNeutronPort(reqID string, req CreatePortReq, tenantID string) (b []byte, e error) {
	defer func() {
		if err := recover(); err != nil {
			b = nil
			e = errors.New("panic")
			klog.Info("==cni master CreateNeutronPort pnic recover start!==")
			debug.PrintStack()
			klog.Info("==cni master CreateNeutronPort pnic recover end!==")
		}
	}()
	reqJSON, err := json.Marshal(&req)
	if err != nil {
		klog.Errorf("CreateNeutronPort: marshall http request body:%v failed, error: -%v", req, err)
		return nil, fmt.Errorf("%v:CreateNeutronPort: marshall http request body error", err)
	}
	postPortURL := m.GetCreatePortURL(tenantID) + MakeURLReqIDSuffix(reqID)
	klog.Infof("CreateNeutronPort: Post url: %s, body: %v", postPortURL, string(reqJSON))
	postStatusCode, rspByte, err := m.PostBytes(postPortURL, reqJSON)
	if err != nil {
		klog.Errorf("CreateNeutronPort: masterClient.Post(posturl: %s, nil) error! -%v", postPortURL, err)
		return nil, fmt.Errorf("%v:CreateNeutronPort: masterClient.Post return error", err)
	} else if postStatusCode != 200 {
		klog.Errorf("CreateNeutronPort: masterClient.Post(posturl: %v, %v) ok, but return status code is %v", postPortURL, string(reqJSON), postStatusCode)
		return nil, fmt.Errorf("createNeutronPort: masterClient.Post return status code:%v error msg:%v",
			postStatusCode, errobj.GetErrMsg(rspByte))
	}
	return rspByte, nil
}

func (m *ManagerClient) CreateNeutronBulkPorts(reqID string, req agtmgr.AgtBulkPortsReq, tenantID string) (b []byte, e error) {
	defer func() {
		if err := recover(); err != nil {
			b = nil
			e = errors.New("panic")
			klog.Info("==cni master CreateNeutronBulkPorts pnic recover start!==")
			debug.PrintStack()
			klog.Info("==cni master CreateNeutronBulkPorts pnic recover end!==")
		}
	}()
	reqJSON, err := json.Marshal(&req)
	if err != nil {
		klog.Errorf("CreateNeutronBulkPorts: marshall http request body:%v failed, error: -%v", req, err)
		return nil, err
	}

	postPortURL := m.GetCreatePortURL(tenantID) + MakeURLReqIDSuffix(reqID)
	klog.Infof("CreateNeutronBulkPorts: Post url: %s, body: %v", postPortURL, string(reqJSON))
	postStatusCode, rspByte, err := m.PostBytes(postPortURL, reqJSON)
	if err != nil {
		klog.Errorf("CreateNeutronBulkPorts: masterClient.Post(posturl: %s, nil) error! -%v", postPortURL, err)
		return nil, err
	} else if postStatusCode != 200 {
		klog.Errorf("CreateNeutronBulkPorts: masterClient.Post(posturl: %v, %v) ok, but return status code is %v", postPortURL, string(reqJSON), postStatusCode)
		return nil, errors.New(errobj.GetErrMsg(rspByte))
	}
	return rspByte, nil
}

var DeleteNeutronPort = func(m *ManagerClient, portID string, tenantID string) error {
	return m.DeleteNeutronPort(portID, tenantID)
}

func (m *ManagerClient) DeleteNeutronPort(portID string, tenantID string) (e error) {
	defer func() {
		if err := recover(); err != nil {
			e = errors.New("panic")
			klog.Info("==cni master DeleteNeutronPort pnic recover start!==")
			debug.PrintStack()
			klog.Info("==cni master DeleteNeutronPort pnic recover end!==")
		}
	}()
	reqID := NewGUID(portID)
	deleteurl := m.GetDeletePortURL(tenantID, portID) + MakeURLReqIDSuffix(reqID)
	klog.Infof("DeleteNeutronPort: delete url is: %s", deleteurl)
	bRsp, statusCode, err := m.Delete(deleteurl)
	if err != nil {
		klog.Errorf("DeleteNeutronPort: masterClient.Delete(deleteurl: %s) error! -%v", deleteurl, err)
		return fmt.Errorf("%v:DeleteNeutronPort: masterClient.Delete error", err)
	}
	if statusCode == 404 {
		klog.Infof("DeleteNeutronPort: masterClient.Delete statusCode is :%d", statusCode)
		return nil
	}
	if statusCode >= 300 || statusCode < 200 {
		klog.Errorf("DeleteNeutronPort: masterClient.Delete(deleteurl: %s) error! statscode: %v  response data: %v", deleteurl, statusCode, string(bRsp))
		return fmt.Errorf("DeleteNeutronPort: masterClient.Delete error! statusCode: %v  response data: %v", statusCode, string(bRsp))
	}

	klog.Infof("DeleteNeutronPort: delete url: %s response data: %v statusCode: %d", deleteurl, string(bRsp), statusCode)
	return nil
}

func (m *ManagerClient) GetNetInfoByNetName(netName, tenantID string) (*mgragt.PaasNetwork, error) {
	postURL := m.GetNetworkURL(tenantID, netName)
	statuscode, netByte, err := m.Get(postURL)
	if err != nil {
		klog.Errorf("GetNetInfoByNetName: GetNetInfoByNetName failed, error! err:-%v, statusNo:%v", err, statuscode)
		return nil, err
	}
	if statuscode != 200 {
		klog.Errorf("GetNetInfoByNetName return statuscode:%v is not 200", statuscode)
		return nil, errors.New("getNetInfoByNetName return statuscode is not 200")
	}
	network := &mgragt.PaasNetwork{}
	err = json.Unmarshal(netByte, network)
	if err != nil {
		klog.Errorf("GetNetInfoByNetName Json unmarshal error! -%v", err)
		return nil, err
	}
	return network, nil
}

func (m *ManagerClient) CheckKnitterManager() error {
	getURL := m.GetHealthURL()
	klog.Infof("CheckKnitterManager: Get url: %s", getURL)
	getStatusCode, healthByte, err := m.Get(getURL)
	if err != nil {
		klog.Errorf("CheckKnitterManager: masterClient.Get(getUrl: %s) error! -%v", getURL, err)
		return fmt.Errorf("%v:CheckKnitterManager: masterClient.Get return error", err)
	}
	if getStatusCode != 200 {
		klog.Errorf("CheckKnitterManager: masterClient.Get(getUrl: %v) ok, but return status code is %v",
			getURL, getStatusCode)
		return errors.New("checkKnitterManager: masterClient.Get return status code is not 200 error")
	}
	healthJSON, err := jason.NewObjectFromBytes(healthByte)
	if err != nil {
		klog.Errorf("CheckKnitterManager: jason.NewObjectFromBytes(healthByte) error, err is %v",
			err)
		return fmt.Errorf("%v:CheckKnitterManager: jason.NewObjectFromBytes(healthByte) error", err)
	}
	healthState, err := healthJSON.GetString("state")
	if err != nil {
		klog.Errorf("CheckKnitterManager: healthJson.GetString(state) error, err is %v",
			err)
		return fmt.Errorf("%v:CheckKnitterManager: healthJson.GetString(state) error", err)
	}
	if healthState != "good" {
		klog.Errorf("CheckKnitterManager: Knitter-manager is not service state!")
		return errors.New("checkKnitterManager: Knitter-manager is not service state")
	}
	klog.Info("CheckKnitterManager Successful!")
	return nil
}

type DcVnisEncap struct {
	Vnis []string `json:"vnis"`
}

func (self *ManagerClient) GetDcVnis(tenantID, dcID string) ([]string, error) {
	url := self.GetDcVnisURL(tenantID, dcID)
	klog.Infof("ManagerClient.GetDcVnis: url: %s", url)
	statCode, rspBody, err := self.Get(url)
	if err != nil {
		klog.Errorf("ManagerClient.GetDcVnis: masterClient.Get(getUrl: %s) error! -%v", url, err)
		return nil, err
	}
	if statCode != 200 {
		klog.Errorf("ManagerClient.GetDcVnis: masterClient.Get(getUrl: %v) ok, but return status code is %v", url, statCode)
		return nil, errobj.ErrInvalidStateCode
	}
	klog.Infof("ManagerClient.GetDcVnis: GET response body is: %s", string(rspBody))

	vnisEncap := DcVnisEncap{}
	err = json.Unmarshal(rspBody, &vnisEncap)
	if err != nil {
		klog.Errorf("ManagerClient.GetDcVnis:, json.Unmarshal(%v) error: %v", rspBody, err)
		return nil, err
	}
	klog.Infof("ManagerClient.GetDcVnis: GET response body Unmarshaled is: %v", vnisEncap)

	return vnisEncap.Vnis, nil
}
