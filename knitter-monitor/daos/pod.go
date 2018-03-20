package daos

import (
	"encoding/json"

	"github.com/ZTE/Knitter/knitter-monitor/err-obj"
	"github.com/ZTE/Knitter/knitter-monitor/infra"
	"github.com/ZTE/Knitter/pkg/db-accessor"
	"github.com/ZTE/Knitter/pkg/klog"
)

func GetPodDao() PodDaoInterface {
	//do not delete, for monkey ut
	klog.Debugf("")
	return &podDao{}
}

type PodDaoInterface interface {
	Get(podNs, podName string) (*PodForDB, error)
	Save(podForDB *PodForDB) error
	Delete(podNs, podName string) error
}

type podDao struct {
}

func (pd *podDao) Get(podNs, podName string) (*PodForDB, error) {
	klog.Debugf("podDao.get start,pod name is [%v] ", podName)
	key := dbaccessor.GetKeyOfMonitorPod(podNs, podName)

	value, err := infra.GetDataBase().ReadLeaf(key)
	if err != nil {
		klog.Errorf("monitorcommon.GetDataBase().ReadLeaf(key:[%v]) err, error is [%v]", key, err)
		return nil, err
	}
	podForDB := &PodForDB{}
	err = json.Unmarshal([]byte(value), podForDB)
	if err != nil {
		klog.Errorf("json.Unmarshal([]byte(value:[%v]), podForDB: [%v]) err, error is [%v]", value, podForDB, err)
		return nil, err
	}
	klog.Debugf("podDao.get SUCC, podForDB is [%v]", podForDB)
	return podForDB, nil

}

func (pd *podDao) Save(podForDB *PodForDB) error {
	klog.Debugf("podDao.Save start : pod forDB is [%v]", podForDB)
	key := dbaccessor.GetKeyOfMonitorPod(podForDB.PodNs, podForDB.PodName)
	podByte, err := json.Marshal(podForDB)
	if err != nil {
		klog.Errorf("podDao.Save: json.Marshal(podForDB :[%v]) err, error is [%v]", podForDB, err)
		return err
	}
	err = infra.GetDataBase().SaveLeaf(key, string(podByte))
	if err != nil {
		klog.Errorf("podDao.Save: monitorcommon.GetDataBase().SaveLeaf(key, string(podByte)) err, error is [%v]", err)
		return err
	}
	klog.Debugf("podDao.Save END : pod forDB is [%v]", podForDB)

	return nil
}

func (pd *podDao) Delete(podNs, podName string) error {
	klog.Debugf("podDao delete start,pod name is [%v] ", podName)
	key := dbaccessor.GetKeyOfMonitorPod(podNs, podName)

	err := infra.GetDataBase().DeleteLeaf(key)
	if err != nil && !errobj.IsKeyNotFoundError(err) {
		klog.Errorf("monitorcommon.GetDataBase().DeleteLeaf(key)")
	}
	klog.Debugf("podDao.delete:podDao delete SUCC")
	return nil

}

type PodForDB struct {
	TenantId     string       `json:"tenant_id"`
	PodID        string       `json:"pod_id"`
	PodName      string       `json:"pod_name"`
	PodNs        string       `json:"pod_ns"`
	PodType      string       `json:"pod_type"`
	IsSuccessful bool         `json:"is_successful"`
	ErrorMsg     string       `json:"error_msg"`
	Ports        []*PortForDB `json:"ports"`
}
