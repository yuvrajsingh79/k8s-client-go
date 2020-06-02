package controller

import (
	"client-go/k8s-client-go/pkg/apis/dummy/v1alpha1"
	crd "client-go/k8s-client-go/pkg/apis/dummy/v1alpha1"
	"client-go/k8s-client-go/utility"
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/micro/go-micro/util/log"
	"k8s.io/apimachinery/pkg/fields"

	corev1 "k8s.io/api/core/v1"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog"

	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

//Controller struct
type Controller struct {
	indexer  cache.Indexer
	queue    workqueue.RateLimitingInterface
	informer cache.Controller
	client   kubernetes.Interface
}

//NewController creates an instance of a Controller
func NewController(queue workqueue.RateLimitingInterface, indexer cache.Indexer, informer cache.Controller, client kubernetes.Interface) *Controller {
	return &Controller{
		informer: informer,
		indexer:  indexer,
		queue:    queue,
		client:   client,
	}
}

//processNextItem is a execution process of Controller
func (c *Controller) processNextItem() bool {
	// Wait until there is a new item in the working queue
	key, quit := c.queue.Get()
	if quit {
		return false
	}
	// Tell the queue that we are done with processing this key. This unblocks the key for other workers
	// This allows safe parallel processing because two pods with the same key are never processed in
	// parallel.
	defer c.queue.Done(key)

	// Invoke the method containing the business logic
	err := c.syncToStdout(key.(string))
	// Handle the error if something went wrong during the execution of the business logic
	c.handleErr(err, key)
	return true
}

// syncToStdout is the business logic of the controller. In this controller it simply prints
// information about the pod to stdout. In case an error happened, it has to simply return the error.
// The retry logic should not be part of the business logic.
func (c *Controller) syncToStdout(key string) error {
	obj, exists, err := c.indexer.GetByKey(key)
	if err != nil {
		klog.Errorf("Fetching object with key %s from store failed with %v", key, err)
		return err
	}

	if !exists {
		// Below we will warm up our cache with a Pod, so that we will see a delete for one pod
		fmt.Printf("FirstCrd %s does not exist anymore\n", key)
	} else {
		// Note that you also have to check the uid if you have a local controlled resource, which
		// is dependent on the actual instance, to detect that a Pod was recreated with the same name
		fmt.Printf("Sync/Add/Update for FirstCrd %s\n", obj.(*v1alpha1.FirstCrd).GetName())
		createPod(obj, c.client)

	}
	return nil
}

func createPod(instance interface{}, client kubernetes.Interface) {
	pod := newPodForCR(instance.(*crd.FirstCrd))

	_, err := client.CoreV1().Pods(pod.Namespace).Get(context.TODO(), pod.Name, meta_v1.GetOptions{})
	if err != nil && errors.IsNotFound(err) {
		fmt.Println("Creating a new Pod", pod.Name, "in ns ", pod.Namespace)
		_, err := client.CoreV1().Pods(pod.Namespace).Create(context.TODO(), pod, meta_v1.CreateOptions{})
		if err != nil {
			fmt.Println("error in creating pod for cr firstcrd")
		}

	} else if err != nil {
		fmt.Printf(err.Error())
	} else {
		fmt.Println("pod already exists")
	}

	// Pod already exists - don't requeue

}

// newPodForCR returns a busybox pod with the same name/namespace as the cr
func newPodForCR(cr *crd.FirstCrd) *corev1.Pod {
	labels := map[string]string{
		"app": cr.Name,
	}
	return &corev1.Pod{
		ObjectMeta: meta_v1.ObjectMeta{
			Name:      cr.Name + "-pod",
			Namespace: cr.Namespace,
			Labels:    labels,
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:    "busybox",
					Image:   "busybox",
					Command: []string{"/bin/sh", "-c", cr.Spec.Message, "sleep", "3600"},
				},
			},
		},
	}
}

// handleErr checks if an error happened and makes sure we will retry later.
func (c *Controller) handleErr(err error, key interface{}) {
	if err == nil {
		// Forget about the #AddRateLimited history of the key on every successful synchronization.
		// This ensures that future processing of updates for this key is not delayed because of
		// an outdated error history.
		c.queue.Forget(key)
		return
	}

	// This controller retries 5 times if something goes wrong. After that, it stops trying.
	if c.queue.NumRequeues(key) < 5 {
		klog.Infof("Error syncing pod %v: %v", key, err)

		// Re-enqueue the key rate limited. Based on the rate limiter on the
		// queue and the re-enqueue history, the key will be processed later again.
		c.queue.AddRateLimited(key)
		return
	}

	c.queue.Forget(key)
	// Report to an external entity that, even after several retries, we could not successfully process this key
	runtime.HandleError(err)
	klog.Infof("Dropping pod %q out of the queue: %v", key, err)
}

//Run starts the controller
func (c *Controller) Run(threadiness int, stopCh chan struct{}) {
	defer runtime.HandleCrash()

	// Let the workers stop when we are done
	defer c.queue.ShutDown()
	klog.Info("Starting FirstCrd controller")

	go c.informer.Run(stopCh)

	// Wait for all involved caches to be synced, before processing items from the queue is started
	if !cache.WaitForCacheSync(stopCh, c.informer.HasSynced) {
		runtime.HandleError(fmt.Errorf("Timed out waiting for caches to sync"))
		return
	}

	for i := 0; i < threadiness; i++ {
		go wait.Until(c.runWorker, time.Second, stopCh)
	}

	<-stopCh
	klog.Info("Stopping Pod controller")
}

func (c *Controller) runWorker() {
	for c.processNextItem() {
	}
}

//GetPodList fetches all the running pods in the k8s
func GetPodList(w http.ResponseWriter, r *http.Request) {
	clientset, err := utility.GetClientset()
	if err != nil {
		panic(err)
	}

	pods, err := clientset.CoreV1().Pods("").List(context.TODO(), meta_v1.ListOptions{})
	if err != nil {
		panic(err.Error())
	}

	svc, err := clientset.CoreV1().Services("").List(context.TODO(), meta_v1.ListOptions{})
	if err != nil {
		panic(err.Error())
	}

	fmt.Printf("\n\nThere are %d pods in the cluster\n\n", len(pods.Items))
	for _, pod := range pods.Items {
		fmt.Println(utility.PrettyString(pod))
	}

	fmt.Println(strings.Repeat("*", 80))
	fmt.Printf("\nThere are %d services in the cluster\n\n", len(svc.Items))
	for _, svc := range svc.Items {
		fmt.Println(utility.PrettyString(svc))
	}
}

//CreateCR is a handler to create a CRD
func CreateCR(w http.ResponseWriter, r *http.Request) {
	// get the Kubernetes client for connectivity
	client, myresourceClient, kubeClient := utility.GetKubeClient()

	// Create the CRD
	err := CreateCRD(client)
	if err != nil {
		log.Fatalf("Failed to create crd: %v", err)
	}

	// Wait for the CRD to be created before we use it.
	time.Sleep(5 * time.Second)

	FirstCrd := &crd.FirstCrd{
		ObjectMeta: meta_v1.ObjectMeta{
			Name:   "firstcrd123",
			Labels: map[string]string{"mylabel": "crd"},
		},
		Spec: crd.FirstCrdSpec{
			Message: "echo hello",
		},
		Status: crd.FirstCrdStatus{
			Name: "created",
		},
	}
	// Create the SslConfig object we create above in the k8s cluster
	resp, err := myresourceClient.DummyV1alpha1().FirstCrds("default").Create(context.TODO(), FirstCrd, meta_v1.CreateOptions{})
	if err != nil {
		fmt.Printf("error while creating object: %v\n", err)
	} else {
		fmt.Printf("object created: %v , %v\n", resp.GetName(), resp.GetLabels())
	}

	obj, err := myresourceClient.DummyV1alpha1().FirstCrds("default").Get(context.TODO(), FirstCrd.ObjectMeta.Name, meta_v1.GetOptions{})
	if err != nil {
		log.Infof("error while getting the object %v\n", err)
	}
	fmt.Printf("FirstCrd Objects Found: \n%+v\n", obj.GetName())

	queue := workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())

	crdListWatcher := cache.NewListWatchFromClient(myresourceClient.DummyV1alpha1().RESTClient(), "firstcrds", corev1.NamespaceDefault, fields.Everything())

	indexer, informer := cache.NewIndexerInformer(crdListWatcher, &v1alpha1.FirstCrd{}, 0, cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(obj)
			if err == nil {
				queue.Add(key)
			}
		},
		UpdateFunc: func(old interface{}, new interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(new)
			if err == nil {
				queue.Add(key)
			}
		},
		DeleteFunc: func(obj interface{}) {
			// IndexerInformer uses a delta queue, therefore for deletes we have to use this
			// key function.
			key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
			if err == nil {
				queue.Add(key)
			}
		},
	}, cache.Indexers{})

	controller := NewController(queue, indexer, informer, kubeClient)

	// use a channel to synchronize the finalization for a graceful shutdown
	stopCh := make(chan struct{})
	defer close(stopCh)

	// run the controller loop to process items
	go controller.Run(1, stopCh)

	// use a channel to handle OS signals to terminate and gracefully shut
	// down processing
	sigTerm := make(chan os.Signal, 1)
	signal.Notify(sigTerm, syscall.SIGTERM)
	signal.Notify(sigTerm, syscall.SIGINT)
	<-sigTerm
}
