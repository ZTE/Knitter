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

package brtunsubrole

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"time"

	"errors"
	"github.com/ZTE/Knitter/pkg/db-accessor"
	"github.com/ZTE/Knitter/pkg/klog"
	"strconv"
)

type TunMng interface {
	Sync(topo *dbaccessor.Sync) error
}

type SyncClientRole struct {
	FlowMgrRole *FlowMgrRole
	Data        *dbaccessor.Sync
	Host        string
	API         string
	InternalIP  string
}

func SetManager(managerURL, internalIP string) error {
	clientSyncWithManager.Host = managerURL
	clientSyncWithManager.InternalIP = internalIP
	klog.Infof("brtunsubrole: SetManager result: Host: %s, Api: %s, InternalIp: %s",
		clientSyncWithManager.Host, clientSyncWithManager.API, clientSyncWithManager.InternalIP)
	return nil
}

func getManager() string {
	return "http://192.168.1.204:9527"
}

var clientSyncWithManager *SyncClientRole = &SyncClientRole{
	Host: getManager(), API: "/tenants/admin/sync"}

func (this *SyncClientRole) Init() error {
	this.Host = clientSyncWithManager.Host
	this.API = clientSyncWithManager.API
	this.InternalIP = clientSyncWithManager.InternalIP

	this.FlowMgrRole = GetFlowMgrSingleton()

	topo, err := this.Sync()
	if err != nil {
		return err
	}
	this.FlowMgrRole.Sync(topo)

	return nil
}

func (this *SyncClientRole) getSyncURL() string {
	return this.Host + this.API + "/" + this.InternalIP
}

func (this *SyncClientRole) syncToMgr() (*http.Response, error) {
	req, err := http.NewRequest("GET", this.getSyncURL(), nil)
	if err != nil {
		klog.Error("NewRequest error:", err)
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		klog.Error("DefaultClient.Do(req) error:", err)
		return nil, err
	}
	return resp, nil
}

func (self *SyncClientRole) getTopoData(resp *http.Response) (*dbaccessor.Sync, error) {
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		klog.Error("Get Request to Manager return ERROR:[",
			resp.StatusCode, "] vs 200 wanted.")
		return nil, errors.New("respond-code-is-error")
	}

	syncData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		klog.Error("ioutil.ReadAll respond Body error:", err.Error())
		return nil, err
	}

	var globalSync dbaccessor.Sync
	err = json.Unmarshal(syncData, &globalSync)
	if err != nil {
		klog.Errorf("Unmarshal respond Body: [%s] error: %v", string(syncData), err)
		return nil, err
	}

	self.Data = &globalSync
	klog.Info("TOPO:\r\n", string(syncData))
	return self.Data, nil
}

func syncToMgr(cli *SyncClientRole) (*dbaccessor.Sync, error) {
	resp, err := cli.syncToMgr()
	if err != nil {
		klog.Infof("syncToMgr: SyncClientRole.syncToMgr : %v", err)
		return nil, err
	}

	syncData, err := cli.getTopoData(resp)
	if err != nil {
		klog.Infof("syncToMgr: SyncClientRole.getTopoData FAILED, error: %v", err)
		return nil, err
	}
	klog.Infof("syncToMgr: sync with knitter-manager SUCC, syncData: %v", *syncData)
	return syncData, nil
}

func (self *SyncClientRole) Sync() (*dbaccessor.Sync, error) {
	const WaitForManagerReady int = 10
	for {
		klog.Trace("SYNC[", self.getSyncURL(), "] start at:", time.Now().String())
		syncData, err := syncToMgr(self)
		if err != nil {
			klog.Warningf("SyncClientRole.Sync: syncToMgr FAILED, error: %v", err)
			time.Sleep(time.Second * time.Duration(WaitForManagerReady))
			continue
		}
		return syncData, nil
	}
}

func (self *SyncClientRole) getInterVal() int {
	intervalSecond, err := strconv.Atoi(
		self.Data.Interval)
	if err != nil {
		return dbaccessor.DefaultIntervalTimeInSecond
	}
	return intervalSecond
}

func (this *SyncClientRole) HeartBeat() {
	intervalSecond := this.getInterVal()
	klog.Trace("Sync-interval-time-is:", intervalSecond, " seconds")
	cycleTimer := time.NewTicker(time.Duration(intervalSecond) * time.Second)
	for {
		select {
		case <-cycleTimer.C:
			topo, err := this.Sync()
			if err != nil {
				klog.Warningf("SYNC ERROR:%v", err)
				break
			}

			this.FlowMgrRole.Sync(topo)
			if this.getInterVal() != intervalSecond {
				cycleTimer.Stop()
				return
			}
		}
	}
}
