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

package portrecycle

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ZTE/Knitter/knitter-agent/domain/adapter"
	"github.com/ZTE/Knitter/knitter-agent/domain/cni"
	"github.com/ZTE/Knitter/knitter-agent/err-obj"
	"github.com/ZTE/Knitter/knitter-agent/infra"
	"github.com/ZTE/Knitter/knitter-agent/infra/memtbl"
	"github.com/ZTE/Knitter/knitter-agent/infra/os-encap"
	"github.com/ZTE/Knitter/knitter-agent/infra/refcnt"
	"github.com/ZTE/Knitter/pkg/db-accessor"
	"github.com/ZTE/Knitter/pkg/etcd"
	"github.com/ZTE/Knitter/pkg/iaas-accessor"
	"github.com/ZTE/Knitter/pkg/klog"
	"github.com/antonholmquist/jason"
	"github.com/coreos/etcd/client"
	"strconv"
	"strings"
	"sync"
	"time"
)

type PodList struct {
	PodsList []Pod
}

type Pod struct {
	PodNs   string
	PodName string
}

type RecyclePodValue struct {
	//ContainerId string   `json:"container_id"`
	//NetNs   string       `json:"netns"`
	PodNs   string `json:"podns"`
	PodName string `json:"podname"`
	Count   int    `json:"count"`
}

type RecyclePodTableRole struct {
	PodMap map[string]interface{}
	rwLock sync.RWMutex
	dp     infra.DataPersister
}

type RecyclePodRepo interface {
	Load() error
	memtbl.MemTblOp
	refcnt.RefCounter
}

var recyclePodTableSingleton RecyclePodRepo
var recyclePodTableSingletonLock sync.Mutex

func GetRecyclePodTableTableSingleton() RecyclePodRepo {
	if recyclePodTableSingleton != nil {
		return recyclePodTableSingleton
	}

	recyclePodTableSingletonLock.Lock()
	defer recyclePodTableSingletonLock.Unlock()
	if recyclePodTableSingleton == nil {
		recyclePodTableSingleton = &RecyclePodTableRole{
			make(map[string]interface{}),
			sync.RWMutex{},
			infra.DataPersister{DirName: "recycle", FileName: "recyclePodTable.dat"}}
	}
	return recyclePodTableSingleton
}

func (this *RecyclePodTableRole) Load() error {
	klog.Infof("attempt to load RecyclePodTable.")
	_, err := osencap.OsStat(this.dp.GetFilePath())
	if err != nil {
		klog.Warningf("recycle pod table doesn't exist!, err: %v", err)
		return fmt.Errorf("recycle pod table doesn't exist")
	}
	RecyclePods := make(map[string]RecyclePodValue)
	err = adapter.DataPersisterLoadFromMemFile(&this.dp, &RecyclePods)
	if err != nil {
		klog.Warningf("restore recycle pod table failed!, err: %v", err)
		return errors.New("restore recycle pod table failed")
	}
	for k, v := range RecyclePods {
		this.PodMap[k] = v
	}
	return nil
}

func (this *RecyclePodTableRole) GetAll() (map[string]interface{}, error) {
	klog.Infof("attempt to get all pod info from RecyclePodTable.")
	_, err := osencap.OsStat(this.dp.GetFilePath())
	if err != nil {
		klog.Warningf("recycle pod table doesn't exist!, err: %v", err)
		return nil, fmt.Errorf("recycle pod table doesn't exist")
	}
	RecyclePods := make(map[string]RecyclePodValue)
	err = adapter.DataPersisterLoadFromMemFile(&this.dp, &RecyclePods)
	if err != nil {
		klog.Warningf("restore recycle pod table failed!, err: %v", err)
		return nil, errors.New("restore recycle pod table failed")
	}
	for k, v := range RecyclePods {
		this.PodMap[k] = v
	}
	return this.PodMap, nil
}

func (this *RecyclePodTableRole) Get(key string) (interface{}, error) {
	this.rwLock.RLock()
	defer this.rwLock.RUnlock()
	value, ok := this.PodMap[key]
	if ok {
		return value, nil
	}
	return nil, errobj.ErrRecordNtExist
}

func (this *RecyclePodTableRole) Insert(key string, value interface{}) error {
	actValue, ok := value.(RecyclePodValue)
	if !ok {
		return errobj.ErrArgTypeMismatch
	}
	this.rwLock.Lock()
	defer this.rwLock.Unlock()
	this.PodMap[key] = actValue

	err := adapter.DataPersisterSaveToMemFile(&this.dp, this.PodMap)
	if err != nil {
		klog.Warningf("insert, store to ram failed!")
		delete(this.PodMap, key)
		return fmt.Errorf("%v:store to ram failed", err)
	}
	return nil
}

func (this *RecyclePodTableRole) Delete(key string) error {
	this.rwLock.Lock()
	defer this.rwLock.Unlock()
	delete(this.PodMap, key)

	err := adapter.DataPersisterSaveToMemFile(&this.dp, this.PodMap)
	if err != nil {
		klog.Warningf("delete, store to ram failed!")
		return fmt.Errorf("%v:store to ram failed", err)
	}
	return nil
}

func (this *RecyclePodTableRole) Inc(key, mock string) error {
	this.rwLock.Lock()
	defer this.rwLock.Unlock()
	value, ok := this.PodMap[key]
	if !ok {
		klog.Errorf("incCount, not found value!key:%v ; value:%v", key, value)
		return errobj.ErrRecordNtExist
	}
	actValue := value.(RecyclePodValue)
	klog.Infof("RefCount: key:%v, value.RefCount:%v", key, actValue.Count)
	actValue.Count++
	this.PodMap[key] = actValue
	klog.Infof("RefCount:key:%v, value.RefCount:%v", key, actValue.Count)
	err := adapter.DataPersisterSaveToMemFile(&this.dp, this.PodMap)
	if err != nil {
		klog.Errorf("store to ram failed!")
		return fmt.Errorf("%v:store to ram failed", err)
	}
	return nil
}

func (this *RecyclePodTableRole) Dec(key, refer string) error {
	this.rwLock.Lock()
	defer this.rwLock.Unlock()
	return nil
}

func (this *RecyclePodTableRole) IsEmpty(key string) bool {
	this.rwLock.Lock()
	defer this.rwLock.Unlock()
	return false
}

func CountAndCollectNoUsedPod() []Pod {
	var delPod []Pod
	recyclePodTable, _ := GetRecyclePodTableTableSingleton().GetAll()
	klog.Infof("CountAndCollectNoUsedPod-start:recyclePodTable[%v]", recyclePodTable)
	NoUsePodList, _ := CollectNoUsedPod()
	klog.Infof("CountAndCollectNoUsedPod: NoUsePodList[%v]", NoUsePodList)
	if NoUsePodList == nil {
		for _, recyclePod := range recyclePodTable {
			if (recyclePod.(RecyclePodValue).Count + 1) > MaxCheckTimes {
				delPod = append(delPod, Pod{recyclePod.(RecyclePodValue).PodNs, recyclePod.(RecyclePodValue).PodName})
				klog.Debugf("CountAndCollectNoUsedPod: delPod[%v]", delPod)
				GetRecyclePodTableTableSingleton().Delete(recyclePod.(RecyclePodValue).PodNs + recyclePod.(RecyclePodValue).PodName)
			} else {
				GetRecyclePodTableTableSingleton().Inc(recyclePod.(RecyclePodValue).PodNs+recyclePod.(RecyclePodValue).PodName, "")
			}
		}
	} else {
		for _, noUsePod := range NoUsePodList {
			recyclePod, _ := GetRecyclePodTableTableSingleton().Get(noUsePod.PodNs + noUsePod.PodName)
			if recyclePod != nil {
				if (recyclePod.(RecyclePodValue).Count + 1) > MaxCheckTimes {
					delPod = append(delPod, Pod{recyclePod.(RecyclePodValue).PodNs, recyclePod.(RecyclePodValue).PodName})
					klog.Debugf("CountAndCollectNoUsedPod: delPod[%v]", delPod)
					GetRecyclePodTableTableSingleton().Delete(recyclePod.(RecyclePodValue).PodNs + recyclePod.(RecyclePodValue).PodName)
				} else {
					GetRecyclePodTableTableSingleton().Inc(recyclePod.(RecyclePodValue).PodNs+recyclePod.(RecyclePodValue).PodName, "")
				}
			} else {
				PodValue := RecyclePodValue{
					PodNs:   noUsePod.PodNs,
					PodName: noUsePod.PodName,
					Count:   1,
				}
				klog.Debugf("CountAndCollectNoUsedPod: PodValue[%v]", PodValue)
				GetRecyclePodTableTableSingleton().Insert(noUsePod.PodNs+noUsePod.PodName, PodValue)
			}
		}
	}
	recyclePodTableTemp, _ := GetRecyclePodTableTableSingleton().GetAll()
	klog.Infof("CountAndCollectNoUsedPod-end:recyclePodTable[%v]", recyclePodTableTemp)
	return delPod
}

func isExist(podNs string, podNss []string) bool {
	if podNss == nil || len(podNss) == 0 {
		return false
	}
	for _, ns := range podNss {
		if podNs == "" || podNs == ns {
			return true
		}
	}
	return false
}

func GetPodsFromDb() ([]Pod, error) {
	agtCtx := cni.GetGlobalContext()
	keyOfPodsForNode := dbaccessor.GetKeyOfPodsForNode(agtCtx.ClusterID, agtCtx.HostIP)
	podsFromDb, err := adapter.ReadDirFromDb(keyOfPodsForNode)
	if err != nil {
		return nil, err
	}
	var pods []Pod
	podnss := make([]string, 0)
	var isDbServReliable bool = true
	for _, pod := range podsFromDb {
		ns := strings.TrimPrefix(pod.Key, keyOfPodsForNode+"/")
		if isExist(ns, podnss) {
			continue
		}
		podnss = append(podnss, ns)
		podNss, err := adapter.ReadDirFromDb(keyOfPodsForNode + "/" + ns)
		if err != nil {
			if errobj.IsEqual(err, errobj.ErrDbKeyNotFound) {
				adapter.ClearDirFromDb(pod.Key)
			} else {
				isDbServReliable = false
			}
			continue
		}
		for _, podNs := range podNss {
			name := strings.TrimPrefix(podNs.Key, keyOfPodsForNode+"/"+ns+"/")
			recyclePod := Pod{
				PodNs:   ns,
				PodName: name,
			}
			//			klog.Infof("AnalyzeRecyclePod:recyclePod [%v] infof: %v", key, recyclePod)
			pods = append(pods, recyclePod)
		}
	}
	if len(pods) == 0 {
		if isDbServReliable {
			adapter.ClearDirFromDb(keyOfPodsForNode)
			return nil, errobj.ErrNoPodCurrentNodeDbReliable
		}
		return nil, errobj.ErrNoPodCurrentNodeDbNtReliable
	}
	return pods, nil
}

func AnalyzePods(pods []*jason.Object) []Pod {
	podList := &PodList{}
	if pods == nil {
		return nil
	}
	for _, pod := range pods {
		klog.Debugf("AnalyzePods pod:%v", pod)
		podNs, _ := pod.GetString("metadata", "namespace")
		podName, _ := pod.GetString("metadata", "name")
		recyclePod := Pod{
			PodNs:   podNs,
			PodName: podName,
		}
		//		klog.Infof("AnalyzeRecyclePod:recyclePod [%v] infof: %v", key, recyclePod)
		podList.PodsList = append(podList.PodsList, recyclePod)
	}
	return podList.PodsList
}

func GetPodsFromK8s() ([]Pod, error) {
	agtCtx := cni.GetGlobalContext()
	podList, err := adapter.GetPodsByNodeID(agtCtx.HostIP)
	if err != nil {
		return nil, errobj.ErrK8sGetPodByNodeID
	}
	pods := AnalyzePods(podList)
	return pods, nil
}

func RemoveSliceCopy(slice []Pod, start, end int) []Pod {
	result := make([]Pod, len(slice)-(end-start))
	at := copy(result, slice[:start])
	copy(result[at:], slice[end:])
	return result
}

func CollectNoUsedPod() ([]Pod, error) {
	var podsOfNode []Pod
	var err error
	podsOfNw, errNw := GetPodsFromDb()
	klog.Infof("CollectNoUsedPod: podsOfNw: %v", podsOfNw)
	if errNw != nil {
		return nil, errNw
	}
	podsOfNode, err = GetPodsFromK8s()
	klog.Infof("CollectNoUsedPod: podsOfNode: %v", podsOfNode)
	if err != nil {
		return nil, err
	}
	if podsOfNode == nil {
		return podsOfNw, nil
	}
	for _, pod := range podsOfNode {
		for i := 0; i < len(podsOfNw); i++ {
			if pod.PodNs == podsOfNw[i].PodNs &&
				pod.PodName == podsOfNw[i].PodName {
				podsOfNw = RemoveSliceCopy(podsOfNw, i, i+1)
				//break
			}
		}
	}
	return podsOfNw, nil
}

func GetKeysOfAllPortsInPodDir(ns, name string) ([]*client.Node, error) {
	keyOfInterfaceGroupInPod := dbaccessor.GetKeyOfInterfaceGroupInPod(ns, ns, name)
	portsOfPod, err := adapter.ReadDirFromDb(keyOfInterfaceGroupInPod)
	if err != nil {
		return nil, err
	}
	return portsOfPod, nil
}

func GetPortInfo(ports []*client.Node) ([]iaasaccessor.Interface, error) {
	var portsForDel []iaasaccessor.Interface
	var notFoundNum = 0
	for _, port := range ports {
		portInfo, err := adapter.ReadLeafFromDb(port.Value)
		if err != nil {
			if errobj.IsKeyNotFoundError(err) {
				notFoundNum++
			}
			continue
		}
		portForDel := iaasaccessor.Interface{}
		err = json.Unmarshal([]byte(portInfo), &portForDel)
		if err != nil {
			continue
		}
		portsForDel = append(portsForDel, portForDel)
	}
	if (notFoundNum + len(portsForDel)) == len(ports) {
		return portsForDel, nil
	}
	return portsForDel, errors.New("there were errors, when get port info")
}

func GetPortsOfPod(ns, name string) ([]iaasaccessor.Interface, error) {
	ports, err := GetKeysOfAllPortsInPodDir(ns, name)
	if err != nil {
		return nil, err
	}
	portsInfo, err := GetPortInfo(ports)
	if err != nil {
		return nil, errobj.ErrNwPortInfo
	}
	return portsInfo, nil
}

func RecyclePortSource(port iaasaccessor.Interface) error {
	agtCtx := cni.GetGlobalContext()
	if port.NetPlane != "eio" && port.Accelerate == "true" {
		return nil
	}
	errDelete := adapter.DestroyPort(agtCtx, port)
	if errDelete != nil {
		return errobj.ErrNwRecyclePort
	}
	return nil
}

func ClearPortInDb(ns, name string, port iaasaccessor.Interface) error {
	agtCtx := cni.GetGlobalContext()
	var errLeaf error
	intererfaceID := port.Id + ns + name
	// delete port from etcd
	dirOfInterface := dbaccessor.GetKeyOfInterface(ns, intererfaceID)
	errDir := adapter.ClearDirFromDb(dirOfInterface)
	if errDir != nil {
		return errDir
	}
	keyInterfaceInPod := dbaccessor.GetKeyOfInterfaceInPod(ns, port.Id, ns, name)
	errKeyInterfaceInPod := adapter.ClearLeafFromDb(keyInterfaceInPod)
	if errKeyInterfaceInPod != nil {
		errLeaf = errKeyInterfaceInPod
		klog.Errorf("ClearPortInDb:ClearLeafFromDb(keyInterfaceInPod) error! %v", errKeyInterfaceInPod)
	}
	keyInterfaceInNetwork := dbaccessor.GetKeyOfInterfaceInNetwork(ns, port.NetworkId, intererfaceID)
	errKeyInterfaceInNetwork := adapter.ClearLeafFromDb(keyInterfaceInNetwork)
	if errKeyInterfaceInNetwork != nil {
		errLeaf = errKeyInterfaceInNetwork
		klog.Errorf("ClearPortInDb:ClearLeafFromDb(keyInterfaceInNetwork) error! %v", errKeyInterfaceInNetwork)
	}
	keyPaasInterfaceForNode := dbaccessor.GetKeyOfPaasInterfaceForNode(agtCtx.ClusterID, agtCtx.HostIP, intererfaceID)
	errKeyPaasInterfaceForNode := adapter.ClearLeafFromDb(keyPaasInterfaceForNode)
	if errKeyPaasInterfaceForNode != nil {
		errLeaf = errKeyPaasInterfaceForNode
		klog.Errorf("ClearPortInDb: ClearLeafFromDb(keyPaasInterfaceForNode) error: %v", errKeyPaasInterfaceForNode)
	}
	if port.NetPlane == "eio" {
		keyIaasInterfacesEioForNode := dbaccessor.GetKeyOfIaasEioInterfaceForNode(agtCtx.ClusterID, agtCtx.HostIP, port.Id)
		errKey := adapter.ClearLeafFromDb(keyIaasInterfacesEioForNode)
		if errKey != nil {
			errLeaf = errKey
			klog.Errorf("ClearPortInDb: ClearLeafFromDb(keyIaasInterfacesEioForNode) error: %v", errKey)
		}
	}

	errDeletFixIPPort := adapter.ClearLeafFromRemoteDB(dbaccessor.GetFixIPPortUrl(port.Id))
	if errDeletFixIPPort != nil {
		if !etcd.IsNotFindError(errDeletFixIPPort) {
			klog.Errorf("ClearPortInDb: DeleteLeaf(key: %s) FAILED, error: %v",
				dbaccessor.GetFixIPPortUrl(port.Id), errDeletFixIPPort)
			return errDeletFixIPPort
		}
	}

	return errLeaf
}

func ClearPodInDb(tenantID, ns, name string) error {
	agtCtx := cni.GetGlobalContext()
	keyOfPod := dbaccessor.GetKeyOfPod(tenantID, ns, name)
	//klog.Infof("ClearPodInDb:keyOfPod:%v",keyOfPod)
	errDir := adapter.ClearDirFromDb(keyOfPod)
	if errDir != nil {
		return errDir
	}

	keyOfPodsForNode := dbaccessor.GetKeyOfPodsForNode(agtCtx.ClusterID, agtCtx.HostIP)
	keyOfdelPod := fmt.Sprintf("%v", keyOfPodsForNode) + "/" + fmt.Sprintf("%v", ns) + "/" + fmt.Sprintf("%v", name)
	//klog.Infof("ClearPodInDb:keyOfdelPod:%v",keyOfdelPod)
	errLeaf := adapter.ClearLeafFromDb(keyOfdelPod)
	if errLeaf != nil {
		return errLeaf
	}
	return nil
}

func ClearPortsOfNoUsedPod(ns, name string) error {
	var errClear error
	delPorts, err := GetPortsOfPod(ns, name)
	klog.Infof("ClearPortsOfNoUsedPod:delPorts:%v", delPorts)
	if err != nil && strings.Contains(err.Error(), "key not found") == false {
		return err
	}
	if len(delPorts) == 0 {
		return nil
	}
	for _, port := range delPorts {
		errClear = RecyclePortSource(port)
		if errClear != nil {
			continue
		}
		errClear = ClearPortInDb(ns, name, port)
	}
	return errClear
}

const DefaultRegularCheckInterval = 15 * time.Minute
const MinutesInADay = (24 * 60)
const DefaultMaxCheckTime = 3

var MaxCheckTimes int

func GetRegularCheckInterval() time.Duration {
	timeStr, err := adapter.ReadLeafFromDb(dbaccessor.GetKeyOfRecycleResourceByTimerUrl())
	if err != nil {
		//	klog.Errorf("GetRegularCheckInterval: ReadLeaf error: %s", err)
		return DefaultRegularCheckInterval
	}
	timeInterval, _ := strconv.Atoi(timeStr)
	if timeInterval <= 0 || timeInterval > MinutesInADay {
		return DefaultRegularCheckInterval
	}
	return time.Duration(timeInterval) * time.Minute
}

func GetMaxCheckTimes() int {
	maxCheckTime, err := adapter.ReadLeafFromDb(dbaccessor.GetKeyOfRecycleMaxCheckTimesUrl())
	if err != nil {
		//	klog.Errorf("GetRegularCheckInterval: ReadLeaf error: %s", err)
		return DefaultMaxCheckTime
	}
	times, _ := strconv.Atoi(maxCheckTime)
	if times <= 0 || times > 10 {
		return DefaultMaxCheckTime
	}
	return times
}

func RegularCollectAndClear() {
	klog.Infof("RegularCheck ---------> Start!")
	defer klog.Infof("RegularCheck ---------> Complete!")
	recyclePodList := CountAndCollectNoUsedPod()
	klog.Infof("RegularCollectAndClear:recyclePodList: %v", recyclePodList)
	for _, recyclePod := range recyclePodList {
		klog.Debugf("RegularCollectAndClear: recyclePod[%v]", recyclePod)
		err := ClearPortsOfNoUsedPod(recyclePod.PodNs, recyclePod.PodName)
		if err == nil {
			//	klog.Infof("RegularCollectAndClear:ClearPortsOfNoUsedPod ns[%v]-name[%v] has no error! ClearPodInDb!", recyclePod.PodNs, recyclePod.PodName)
			err := ClearPodInDb(recyclePod.PodNs, recyclePod.PodNs, recyclePod.PodName)
			if err != nil {
				klog.Warningf("RegularCollectAndClear:ClearPodInDb error:%v", err)
			}
		}
	}
}

func CollectAndClear() {
	timeInterval := GetRegularCheckInterval()
	klog.Infof("SetCycleTimer: Regular check interval is [%s];Max check times is [%v] ", timeInterval, MaxCheckTimes)
	cycleTimer := time.NewTicker(timeInterval)
	for {
		select {
		case <-cycleTimer.C:
			RegularCollectAndClear()
			MaxCheckTimes = GetMaxCheckTimes()
			newTimeInterval := GetRegularCheckInterval()
			if newTimeInterval != timeInterval {
				cycleTimer.Stop()
				return
			}
		}
	}
}

func RecycleResourseByTimer() {
	klog.Info("RecycleResourseByTimer:start!!!")
	//node restart
	klog.Info("RecycleResourseByTimer:first recycle for node/agent start--begin!!!")
	RecyclePodTable := GetRecyclePodTableTableSingleton()
	err := RecyclePodTable.Load()
	if err != nil {
		klog.Warningf("load recycle pod table error:[%v]", err)
	}
	MaxCheckTimes = GetMaxCheckTimes()
	RegularCollectAndClear()
	klog.Info("RecycleResourseByTimer:first recycle for node/agent start--end!!!")
	for {
		CollectAndClear()
	}
}
