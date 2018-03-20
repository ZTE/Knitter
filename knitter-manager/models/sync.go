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

package models

import (
	"encoding/json"
	"math/rand"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/ZTE/Knitter/knitter-manager/public"
	"github.com/ZTE/Knitter/pkg/db-accessor"
	"github.com/ZTE/Knitter/pkg/klog"
	"sync"
)

type SyncMgt struct {
	Active bool
	Data   dbaccessor.Sync
	lock   sync.Mutex
}

const DefaultTTL int = 10

var syncManager SyncMgt = SyncMgt{Active: false}

func GetSyncMgt() *SyncMgt {
	if syncManager.Active == false {
		syncManager.read()
	}
	return &syncManager
}

func (self *SyncMgt) SetInterval(interval string) error {
	_, err := strconv.Atoi(interval)
	if err != nil {
		self.Data.Interval = strconv.Itoa(dbaccessor.DefaultIntervalTimeInSecond)
	} else {
		self.Data.Interval = interval
	}
	self.save()
	return nil
}

func (self *SyncMgt) getDefaultTTL() int {
	return DefaultTTL * self.getNumbOfReadyAgents()
}

func (self *SyncMgt) getNumbOfReadyAgents() int {
	var readyStatusAgentNumb int = 0
	for _, agent := range self.Data.Agents {
		if agent.Status == dbaccessor.AgentStatusReady {
			readyStatusAgentNumb++
		}
	}
	return readyStatusAgentNumb
}

func (self *SyncMgt) save() error {
	ttlValue := self.getDefaultTTL()
	for _, agent := range self.Data.Agents {
		agent.TTL = ttlValue
	}
	data, _ := json.Marshal(self.Data)
	err := common.GetDataBase().SaveLeaf(dbaccessor.GetKeyOfTopoSyncData(), string(data))
	if err == nil {
		self.Active = false
	}
	return err
}

func (self *SyncMgt) read() error {
	syncData, err := common.GetDataBase().ReadLeaf(dbaccessor.GetKeyOfTopoSyncData())
	if err != nil {
		klog.Error("Read-sync-data--error:", err.Error())
		return err
	}
	var sync dbaccessor.Sync
	json.Unmarshal([]byte(syncData), &sync)
	self.Data = sync
	self.Active = true
	return nil
}

func (self *SyncMgt) isExist(ip string) bool {
	agent := self.getAgent(ip)
	if agent == nil {
		return false
	}
	return true
}

func (self *SyncMgt) isReady(ip string) bool {
	agent := self.getAgent(ip)
	if agent == nil {
		return false
	}
	if agent.Status == dbaccessor.AgentStatusReady {
		return true
	}
	return false
}

func (self *SyncMgt) getAgent(ip string) *dbaccessor.Agent {
	for _, agent := range self.Data.Agents {
		if agent.Ip == ip {
			return agent
		}
	}
	return nil
}

func (self *SyncMgt) ip2int(ip net.IP) int64 {
	bits := strings.Split(ip.String(), ".")
	b0, _ := strconv.Atoi(bits[0])
	b1, _ := strconv.Atoi(bits[1])
	b2, _ := strconv.Atoi(bits[2])
	b3, _ := strconv.Atoi(bits[3])

	var sum int64
	sum += int64(b0) << 24
	sum += int64(b1) << 16
	sum += int64(b2) << 8
	sum += int64(b3)
	return sum
}

func (self *SyncMgt) getAgentID(ip string) string {
	ipAddr := net.ParseIP(ip)
	rand.Seed(self.ip2int(ipAddr))
	return strconv.Itoa(rand.Int())
}

func (self *SyncMgt) FlushTTL(reqIP string) error {
	if self.Active == false {
		self.read()
	}

	for _, agent := range self.Data.Agents {
		if agent.Status == dbaccessor.AgentStatusDown {
			continue
		}
		if agent.Ip == reqIP {
			agent.TTL = self.getDefaultTTL()
		} else {
			agent.TTL--
		}
		if agent.TTL <= 0 {
			agent.Status = dbaccessor.AgentStatusDown
		}
	}
	return nil
}

func (self *SyncMgt) Sync(reqIP string) *dbaccessor.SyncRsp {
	self.lock.Lock()
	defer self.lock.Unlock()
	klog.Trace("SYNC:[", reqIP, "] request at:", time.Now().String())
	if self.isExist(reqIP) == false {
		agent := dbaccessor.Agent{Id: self.getAgentID(reqIP),
			Ip: reqIP, Status: dbaccessor.AgentStatusReady}
		self.Data.Agents = append(self.Data.Agents, &agent)
		self.save()
	}

	if self.isReady(reqIP) == false {
		agent := self.getAgent(reqIP)
		agent.Status = dbaccessor.AgentStatusReady
		self.save()
	}

	self.FlushTTL(reqIP)
	self.Data.Client = self.getAgent(reqIP)

	rspData := dbaccessor.SyncRsp{Interval: self.Data.Interval, Client: *self.Data.Client}
	for _, agent := range self.Data.Agents {
		rspData.Agents = append(rspData.Agents, *agent)
	}

	return &rspData
}
