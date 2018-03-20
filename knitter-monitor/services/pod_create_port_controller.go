package services

import (
	"errors"
	"time"

	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"

	"github.com/ZTE/Knitter/knitter-monitor/const-value"
	"github.com/ZTE/Knitter/knitter-monitor/infra"
	"github.com/ZTE/Knitter/pkg/klog"
)

type createPortForPodController struct {
	clientSet       *kubernetes.Clientset
	podController   cache.Controller
	podStoreIndexer cache.Indexer
	podsCreateQueue workqueue.RateLimitingInterface
	podsDeleteQueue workqueue.RateLimitingInterface
}

func NewCreatePortForPodController() (*createPortForPodController, error) {
	cpc := &createPortForPodController{
		podsCreateQueue: workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "pods"),
		podsDeleteQueue: workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "pods"),
	}
	cpc.clientSet = infra.GetClientset()
	if cpc.clientSet == nil {
		klog.Errorf("newCreatePortForPodController: monitorcommon.GetClientset() kubernetes clientset is nil")
		return nil, errors.New("kubernetes clientset is nil")
	}
	watchlist := cache.NewListWatchFromClient(cpc.clientSet.CoreV1().RESTClient(), "pods", v1.NamespaceAll,
		fields.Everything())

	cpc.podStoreIndexer, cpc.podController = cache.NewIndexerInformer(
		watchlist,
		&v1.Pod{},
		time.Second*0,
		cache.ResourceEventHandlerFuncs{
			AddFunc:    cpc.enqueueCreatePod,
			DeleteFunc: cpc.enqueueDeletePod,
		},
		cache.Indexers{},
	)
	return cpc, nil
}

func (cpc *createPortForPodController) enqueueCreatePod(obj interface{}) {
	klog.Infof("enqueueCreatePod start ")
	klog.Debugf("enqueueCreatePod obj is [%v]", obj)
	key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
	if err != nil {
		klog.Errorf("enqueueCreatePod : DeletionHandlingMetaNamespaceKeyFunc err, error is [%v]", err)
		return
	}
	cpc.podsCreateQueue.Add(key)
	klog.Infof("enqueueCreatePod END:key is [%v]", key)

}

func (cpc *createPortForPodController) enqueueDeletePod(obj interface{}) {
	klog.Infof("enqueueDeletePod start ")
	klog.Debugf("enqueueDeletePod obj is [%v]", obj)

	if _, ok := obj.(cache.DeletedFinalStateUnknown); ok {
		return
	}
	cpc.podsDeleteQueue.Add(obj)
	//key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
	//if err != nil {
	//	klog.Errorf("enqueueCreatePod : DeletionHandlingMetaNamespaceKeyFunc err, error is [%v]", err)
	//	return
	//}
	//cpc.podsDeleteQueue.Add(key)

	klog.Infof("enqueueDeletePod END:obj is [%v]", obj)

}

func (cpc *createPortForPodController) createPodWorker() {
	klog.Info("createPodWorker start ")
	workFunc := func() bool {
		key, quit := cpc.podsCreateQueue.Get()
		if quit {
			return true
		}
		klog.Debugf("cpc.podsCreateQueue.Get() key is [%v]", key)
		defer cpc.podsCreateQueue.Done(key)
		obj, exists, err := cpc.podStoreIndexer.GetByKey(key.(string))
		if !exists {
			klog.Errorf("createPodWorker: Pod has been deleted [%v]", key.(string))
			return false
		}
		if err != nil {
			klog.Errorf("cannot get pod: %v\n", key)
			return false
		}
		k8sPod := obj.(*v1.Pod)
		klog.Debugf("createPodWorker: k8sPod is [%v]", k8sPod)

		pod, err := GetPodService().NewPodFromK8sPod(k8sPod)
		klog.Debugf(" GetPodService().NewPodFromK8sPod(k8sPod) ,Pod is [%v]", pod)

		if err != nil {
			klog.Errorf("createPortForPodController.createPodWorker: GetPodService().NewPodFromK8sPod(k8sPod:[%v]) err,error is [%v]", k8sPod, err)
		}
		if pod == nil {
			pod = &Pod{PodName: k8sPod.Name, ErrorMsg: err.Error(), IsSuccessful: false, PodNs: k8sPod.Name}
		}
		err = GetPodService().Save(pod)
		if err != nil {
			//if saving error , add pod to queue
			klog.Errorf("GetPodService().Save( pod:[%v] ) err ,error is [%v]", pod, err)
			cpc.podsCreateQueue.Add(key)
		}

		klog.Infof("@crate@ pod is [%v]", pod)
		return false
	}
	for {
		if quit := workFunc(); quit {
			klog.Infof("createPodWorker shut down")
			return
		}
	}
}

func (cpc *createPortForPodController) deletePodWorker() {
	klog.Info("deletePodWorker start ")
	workFunc := func() bool {
		obj, quit := cpc.podsDeleteQueue.Get()
		if quit {
			return true
		}
		defer cpc.podsDeleteQueue.Done(obj)

		klog.Debugf("cpc.podsDeleteQueue.Get() key :[%v]", obj)

		pod := obj.(*v1.Pod)
		// todo judge k8s recycling resource
		err := GetPodService().DeletePodAndPorts(pod.Namespace, pod.Name)
		if err != nil {
			klog.Warningf("GetPodService().DeletePodAndPorts() err, error is [%v]")
		}
		return false
	}
	for {
		if quit := workFunc(); quit {
			klog.Infof("deletePodWorker shut down")
			return
		}
	}
}

func (cpc *createPortForPodController) Run(workers int, stopCh <-chan struct{}) {
	defer runtime.HandleCrash()
	klog.Infof("createPortForPodController.Run : Starting serviceLookupController Manager ")
	go cpc.podController.Run(stopCh)
	var i int
	for i = 1; i < constvalue.WaitForCacheSyncTimes; i++ {
		if cache.WaitForCacheSync(stopCh, cpc.podController.HasSynced) {
			klog.Infof("cache.WaitForCacheSync(stopCh, cpc.podController.HasSynced) error")
			break
		}
		time.Sleep(time.Second * 1)
	}
	if i == constvalue.WaitForCacheSyncTimes {
		klog.Errorf("createPortForPodController.Run: cache.WaitForCacheSync(stopCh, cpc.podController.HasSynced:[%v]) error,", cpc.podController.HasSynced())
		return
	}
	for i := 0; i < workers; i++ {
		go wait.Until(cpc.createPodWorker, time.Second, stopCh)
		go wait.Until(cpc.deletePodWorker, time.Second, stopCh)
	}
	klog.Infof("createPortForPodController.Run : Started podWorker")

	<-stopCh
	klog.Infof("Shutting down Service Lookup Controller")
	cpc.podsDeleteQueue.ShutDown()
	cpc.podsCreateQueue.ShutDown()

}
