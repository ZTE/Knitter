package services

import (
	"encoding/json"

	"github.com/antonholmquist/jason"
	"k8s.io/api/core/v1"

	"github.com/ZTE/Knitter/knitter-monitor/const-value"
	"github.com/ZTE/Knitter/knitter-monitor/daos"
	"github.com/ZTE/Knitter/knitter-monitor/err-obj"
	"github.com/ZTE/Knitter/knitter-monitor/infra"
	"github.com/ZTE/Knitter/pkg/klog"
)

func GetPodService() PodServiceInterface {
	klog.Debugf("")
	return &podService{}
}

type PodServiceInterface interface {
	NewPodFromK8sPod(pod *v1.Pod) (*Pod, error)
	Save(pod *Pod) error
	Get(podNs, podName string) (*Pod, error)
	DeletePodAndPorts(podNs, podName string) error
}

type podService struct {
}

func (ps *podService) NewPodFromK8sPod(k8sPod *v1.Pod) (*Pod, error) {
	pod := &Pod{}
	pod.PodID = string(k8sPod.GetObjectMeta().GetUID())
	pod.PodName = k8sPod.GetObjectMeta().GetName()
	pod.PodNs = k8sPod.GetObjectMeta().GetNamespace()
	pod.TenantId = pod.PodNs
	pod.IsSuccessful = true

	annotations := k8sPod.GetObjectMeta().GetAnnotations()
	networksStr, ok := annotations["networks"]
	klog.Debugf("NewPodFromK8sPod:networksStr is [%v] ", networksStr)
	var networksByte []byte
	networksByte = []byte(networksStr)

	if isNetworkNotConfigExist(networksStr, ok) {
		klog.Warningf("podService.NewPodFromK8sPod: no network message in blurprint")
		var err error
		networksByte, err = GetDefaultNetworkConfig()
		if err != nil {
			klog.Errorf("podService.NewPodFromK8sPod:GetDefaultNetworkConfig() err, error is [%v]", err)
			return nil, err
		}
	}

	nwJSONObj, err := jason.NewObjectFromBytes(networksByte)
	if err != nil {
		klog.Errorf("NewPodFromK8sPod: jason.NewObjectFromBytes([]byte(networksStr: [%v])) err , err is [%v]", networksStr, err)
		return nil, err
	}
	//determine whether it is pod reconstruction
	reconstructionPod, err := ps.Get(pod.PodNs, pod.PodName)
	if err != nil && !errobj.IsKeyNotFoundError(err) {
		klog.Errorf("NewPodFromK8sPod:  GetPodService().Get(pod.PodNs :[%v], pod.PodName:[%v]) err , err is [%v]", pod.PodNs, pod.PodName, err)
		return nil, err
	}
	// created bulk ports successfully and database having data  is pod reconstruction
	if err == nil && pod.IsSuccessful == true {
		klog.Infof("NewPodFromK8sPod: is pod reconstruction ")
		pod.Ports = reconstructionPod.Ports
		return pod, err
	}
	//create port
	pod4CreatePort := pod.transferToPod4CreatePort()
	pod.Ports, err = GetPortService().NewPortsWithEagerAttrAndLazyAttr(pod4CreatePort, nwJSONObj)
	if err != nil {
		klog.Errorf("NewPodFromK8sPod:  GetPortService().NewPortsWithEagerAttrAndLazyAttr(pod4CreatePort: [%v], nwJSONObj [%v]) err , err is [%v]", pod4CreatePort, nwJSONObj, err)
		return nil, err
	}
	klog.Debugf("NewPodFromK8sPod END, pod is [%+v]", pod)
	return pod, nil
}

func isNetworkNotConfigExist(networksStr string, ok bool) bool {
	return networksStr == "" || !ok || networksStr == "\"\""
}

func GetDefaultNetworkConfig() ([]byte, error) {
	bluePrintNetworkMessage := &BluePrintNetworkMessage{}
	err := bluePrintNetworkMessage.NewDelaultNetworkMessage()
	if err != nil {
		klog.Errorf("GetDefaultNetworkConfig:bluePrintNetworkMessage.NewDelaultNetworkMessage() err: %v", err)
	}
	networksByte, err := json.Marshal(bluePrintNetworkMessage)
	if err != nil {
		klog.Errorf("NewPodFromK8sPod:json.Marshal(bluePrintNetworkMessage) err: %v", err)
		return nil, err
	}
	return networksByte, err
}

type BluePrintNetworkMessage struct {
	Ports []BluePrintPort `json:"ports"`
}

func (bpnm *BluePrintNetworkMessage) NewDelaultNetworkMessage() error {
	netName, err := infra.GetManagerClient().GetDefaultNetWork(constvalue.PaaSTenantAdminDefaultUUID)
	if err != nil {
		klog.Errorf("NewDelaultNetworkMessage() err: %v", err)
		return err
	}
	ports := make([]BluePrintPort, 0)
	port := BluePrintPort{
		AttachToNetwork: netName,
		Attributes: BluePrintAttributes{
			Accelerate: constvalue.DefaultIsAccelerate,
			Function:   constvalue.DefaultNetworkPlane,
			NicName:    constvalue.DefaultPortName,
			NicType:    constvalue.DefaultVnicType,
		},
	}
	ports = append(ports, port)
	bpnm.Ports = ports
	return nil
}

type BluePrintPort struct {
	AttachToNetwork string              `json:"attach_to_network"`
	Attributes      BluePrintAttributes `json:"attributes"`
}

type BluePrintAttributes struct {
	Accelerate string `json:"accelerate"`
	Function   string `json:"function"`
	NicName    string `json:"nic_name"`
	NicType    string `json:"nic_type"`
}

func (ps *podService) Save(pod *Pod) error {
	klog.Debugf("podService.Save start pod is [%v]", pod)
	podForDB := pod.transferToPodForDao()
	err := daos.GetPodDao().Save(podForDB)
	if err != nil {
		klog.Errorf("podService.Save:daos.GetPodDao().Save(podForDB:[%v]) err, error is %v", podForDB, err)
		return err
	}
	return nil

}

func (ps *podService) Get(podNs, podName string) (*Pod, error) {
	podForDB, err := daos.GetPodDao().Get(podNs, podName)
	if err != nil {
		klog.Errorf("podService.Get:daos.GetPodDao().Get(podName:[%v]) err, error is [%v]", podName, err)
		return nil, err
	}
	pod := newPodFromPodForDB(podForDB)
	return pod, nil
}

func (ps *podService) DeletePodAndPorts(podNs, podName string) error {
	pod, err := daos.GetPodDao().Get(podNs, podName)
	if err != nil && !errobj.IsKeyNotFoundError(err) {
		klog.Errorf("daos.GetPodDao().Get(podNs:[%v], podName[:%v]) err, error is [%v]", podNs, podName, err)
		return err
	}
	if errobj.IsKeyNotFoundError(err) {
		klog.Warningf("daos.GetPodDao().Get(podNs:[%v], podName[:%v]) not found", podNs, podName)
		return nil
	}
	portIDs := make([]string, 0)
	for _, port := range pod.Ports {
		portIDs = append(portIDs, port.ID)
	}
	err = GetPortService().DeleteBulkPorts(podNs, portIDs)
	if err != nil {
		klog.Errorf("GetPortService().DeleteBulkPorts(podNs:[%v], portIDs:[%v] ) err, error is [%v]", podNs, portIDs, err)
		return err
	}

	err = daos.GetPodDao().Delete(podNs, podName)
	if err != nil && !errobj.IsKeyNotFoundError(err) {
		klog.Errorf("daos.GetPodDao().Delete(podNs:[%v], podName[%v]) err, error is [%v]", podNs, podName, err)
		return err
	}

	return nil

}

type Pod struct {
	TenantId     string
	PodID        string
	PodName      string
	PodNs        string
	PodType      string //todo podtype
	IsSuccessful bool
	ErrorMsg     string
	Ports        []*Port
}

func newPodFromPodForDB(db *daos.PodForDB) *Pod {
	mports := make([]*Port, 0)
	for _, portForDB := range db.Ports {
		mport := newPortFromPortForDB(portForDB)
		mports = append(mports, mport)
	}
	return &Pod{
		TenantId:     db.TenantId,
		PodID:        db.PodID,
		PodName:      db.PodName,
		PodNs:        db.PodNs,
		PodType:      db.PodType,
		IsSuccessful: db.IsSuccessful,
		ErrorMsg:     db.ErrorMsg,
		Ports:        mports,
	}
}

func (p *Pod) transferToPod4CreatePort() *PodForCreatPort {
	return &PodForCreatPort{
		TenantID: p.TenantId,
		PodID:    p.PodID,
		PodName:  p.PodName,
		PodNs:    p.PodNs,
	}

}

func (p *Pod) transferToPodForDao() *daos.PodForDB {
	portsForDB := make([]*daos.PortForDB, 0)
	for _, port := range p.Ports {
		portForDB := port.transferToPortForDB()
		portsForDB = append(portsForDB, portForDB)
	}
	return &daos.PodForDB{
		TenantId:     p.TenantId,
		PodID:        p.PodID,
		PodNs:        p.PodNs,
		PodName:      p.PodName,
		PodType:      p.PodType,
		ErrorMsg:     p.ErrorMsg,
		IsSuccessful: p.IsSuccessful,
		Ports:        portsForDB,
	}

}

type PodForCreatPort struct {
	TenantID string
	PodID    string
	PodName  string
	PodNs    string
}
