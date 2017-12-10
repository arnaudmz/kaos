/*
Copyright 2017 The Kubernetes Authors.

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

package main

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/golang/glog"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"

	kaosv1 "github.com/arnaudmz/kaos/pkg/apis/kaos/v1"
	clientset "github.com/arnaudmz/kaos/pkg/client/clientset/versioned"
	kaosscheme "github.com/arnaudmz/kaos/pkg/client/clientset/versioned/scheme"
	informers "github.com/arnaudmz/kaos/pkg/client/informers/externalversions"
	listers "github.com/arnaudmz/kaos/pkg/client/listers/kaos/v1"
	"github.com/robfig/cron"
)

const (
	// controllerAgentName is the name in event sources
	controllerAgentName = "kaos-controller"

	// SuccessSynced is used as part of the Event 'reason' when a Database is synced
	SuccessSynced = "Synced"

	// KaosCreated is used to create events when a pod has been deleted
	KaosCreated = "Kaos"

	// CronError is used to store events regarding invalid Cron Syntax
	CronError = "Cron Error"

	// PodSelectingError is used to store events regarding Pod selecting issues
	PodSelectingError = "Pod Selecting Error"

	// PodListingEmpty is used to store events regarding Pod listing that return no pods
	PodListingEmpty = "Pod List Empty"

	// PodListingError is used to store events regarding Pod listing issues
	PodListingError = "Pod Listing Error"

	// PodDeletingError is used o store events regarding Pod deleting issue
	PodDeletingError = "Pod Deleting Error"

	// MessageResourceSynced is the message used for an Event fired when a Database
	// is synced successfully
	MessageResourceSynced = "Kaos Rule synced successfully and cron installed"
)

// CronRuleItem hold data to get back to the cron configuration
type CronRuleItem struct {
	cron            *cron.Cron
	resourceVersion string
}

// Controller is the controller implementation for Database resources
type Controller struct {
	// kubeclientset is a standard kubernetes clientset
	kubeclientset kubernetes.Interface
	// sampleclientset is a clientset for our own API group
	kaosruleclientset clientset.Interface

	kaosrulesLister listers.KaosRuleLister
	kaosrulesSynced cache.InformerSynced

	// workqueue is a rate limited work queue. This is used to queue work to be
	// processed instead of performing it as soon as a change happens. This
	// means we can ensure we only process a fixed amount of resources at a
	// time, and makes it easy to ensure we are never processing the same item
	// simultaneously in two different workers.
	workqueue workqueue.RateLimitingInterface
	// recorder is an event recorder for recording Event resources to the
	// Kubernetes API.
	recorder record.EventRecorder

	// cronPerRule
	cronPerRule map[string]*CronRuleItem

	// random number source
	rand *rand.Rand
}

// NewController returns a new sample controller
func NewController(
	kubeclientset kubernetes.Interface,
	kaosruleclientset clientset.Interface,
	kaosruleInformerFactory informers.SharedInformerFactory) *Controller {

	// obtain references to shared index informers for the Deployment and Database
	// types.
	kaosruleInformer := kaosruleInformerFactory.Kaos().V1().KaosRules()

	// Create event broadcaster
	// Add sample-controller types to the default Kubernetes Scheme so Events can be
	// logged for sample-controller types.
	kaosscheme.AddToScheme(scheme.Scheme)
	glog.V(4).Info("Creating event broadcaster")
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartLogging(glog.Infof)
	eventBroadcaster.StartRecordingToSink(&typedcorev1.EventSinkImpl{Interface: kubeclientset.CoreV1().Events("")})
	recorder := eventBroadcaster.NewRecorder(scheme.Scheme, corev1.EventSource{Component: controllerAgentName})

	controller := &Controller{
		kubeclientset:     kubeclientset,
		kaosruleclientset: kaosruleclientset,
		kaosrulesLister:   kaosruleInformer.Lister(),
		kaosrulesSynced:   kaosruleInformer.Informer().HasSynced,
		workqueue:         workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "DBs"),
		recorder:          recorder,
		cronPerRule:       make(map[string]*CronRuleItem),
		rand:              rand.New(rand.NewSource(time.Now().UnixNano())),
	}

	glog.Info("Setting up event handlers")
	// Set up an event handler for when KaosRule resources change
	kaosruleInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: controller.enqueueKR,
		UpdateFunc: func(old, new interface{}) {
			controller.enqueueKR(new)
		},
		DeleteFunc: controller.enqueueKR,
	})
	return controller
}

// Run will set up the event handlers for types we are interested in, as well
// as syncing informer caches and starting workers. It will block until stopCh
// is closed, at which point it will shutdown the workqueue and wait for
// workers to finish processing their current work items.
func (c *Controller) Run(threadiness int, stopCh <-chan struct{}) error {
	defer runtime.HandleCrash()
	defer c.workqueue.ShutDown()

	// Start the informer factories to begin populating the informer caches
	glog.Info("Starting KaosRule controller")

	// Wait for the caches to be synced before starting workers
	glog.Info("Waiting for informer caches to sync")
	if ok := cache.WaitForCacheSync(stopCh, c.kaosrulesSynced); !ok {
		return fmt.Errorf("failed to wait for caches to sync")
	}

	glog.Info("Starting workers")
	// Launch two workers to process Database resources
	for i := 0; i < threadiness; i++ {
		go wait.Until(c.runWorker, time.Second, stopCh)
	}

	glog.Info("Started workers")
	<-stopCh
	glog.Info("Shutting down workers")

	return nil
}

// runWorker is a long-running function that will continually call the
// processNextWorkItem function in order to read and process a message on the
// workqueue.
func (c *Controller) runWorker() {
	for c.processNextWorkItem() {
	}
}

// processNextWorkItem will read a single work item off the workqueue and
// attempt to process it, by calling the syncHandler.
func (c *Controller) processNextWorkItem() bool {
	obj, shutdown := c.workqueue.Get()

	if shutdown {
		return false
	}

	// We wrap this block in a func so we can defer c.workqueue.Done.
	err := func(obj interface{}) error {
		// We call Done here so the workqueue knows we have finished
		// processing this item. We also must remember to call Forget if we
		// do not want this work item being re-queued. For example, we do
		// not call Forget if a transient error occurs, instead the item is
		// put back on the workqueue and attempted again after a back-off
		// period.
		defer c.workqueue.Done(obj)
		var key string
		var ok bool
		// We expect strings to come off the workqueue. These are of the
		// form namespace/name. We do this as the delayed nature of the
		// workqueue means the items in the informer cache may actually be
		// more up to date that when the item was initially put onto the
		// workqueue.
		if key, ok = obj.(string); !ok {
			// As the item in the workqueue is actually invalid, we call
			// Forget here else we'd go into a loop of attempting to
			// process a work item that is invalid.
			c.workqueue.Forget(obj)
			runtime.HandleError(fmt.Errorf("expected string in workqueue but got %#v", obj))
			return nil
		}
		// Run the syncHandler, passing it the namespace/name string of the
		// Database resource to be synced.
		if err := c.syncHandler(key); err != nil {
			return fmt.Errorf("error syncing '%s': %s", key, err.Error())
		}
		// Finally, if no error occurs we Forget this item so it does not
		// get queued again until another change happens.
		c.workqueue.Forget(obj)
		glog.Infof("Successfully synced '%s'", key)
		return nil
	}(obj)

	if err != nil {
		runtime.HandleError(err)
		return true
	}

	return true
}

// syncHandler compares the actual state with the desired, and attempts to
// converge the two.
func (c *Controller) syncHandler(key string) error {
	// Convert the namespace/name string into a distinct namespace and name
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		runtime.HandleError(fmt.Errorf("invalid resource key: %s", key))
		return nil
	}

	// Get the KaosRule resource with this namespace/name
	kr, err := c.kaosrulesLister.KaosRules(namespace).Get(name)
	if err != nil {
		// The KaosRule resource may no longer exist, in which case we stop
		// processing.
		if errors.IsNotFound(err) {
			glog.V(2).Info(fmt.Sprintf("KaosRule %s no longer exists", key))
			cronItem, ok := c.cronPerRule[key]
			if ok {
				glog.V(4).Info(fmt.Sprintf("Removing existing Cron for %s", key))
				cronItem.cron.Stop()
				delete(c.cronPerRule, key)
				glog.V(4).Info(c.cronPerRule)
			}
			return nil
		}

		return err
	}

	cronItem, ok := c.cronPerRule[key]
	if ok {
		if kr.ResourceVersion == cronItem.resourceVersion {
			return nil
		}

		glog.V(4).Info(fmt.Sprintf("%s Removing existing Cron to update it", kr.String()))
		cronItem.cron.Stop()
		delete(c.cronPerRule, key)
		glog.V(4).Info(c.cronPerRule)
	}
	myCron := cron.New()
	err = myCron.AddFunc(kr.Spec.Cron, func() { c.applyKR(namespace, name, kr) })
	if err != nil {
		c.recorder.Event(kr, corev1.EventTypeWarning, CronError, fmt.Sprintf("Error parsing Cron %s: %v", kr.Spec.Cron, err))
		return err
	}
	myCron.Start()
	c.cronPerRule[key] = &CronRuleItem{
		resourceVersion: kr.ResourceVersion,
		cron:            myCron,
	}
	glog.V(4).Info(c.cronPerRule)
	c.recorder.Event(kr, corev1.EventTypeNormal, SuccessSynced, MessageResourceSynced)
	return nil
}

// enqueueKR takes a KaosRule resource and converts it into a namespace/name
// string which is then put onto the work queue. This method should *not* be
// passed resources of any type other than Db.
func (c *Controller) enqueueKR(obj interface{}) {
	var key string
	var err error
	if key, err = cache.MetaNamespaceKeyFunc(obj); err != nil {
		runtime.HandleError(err)
		return
	}
	c.workqueue.AddRateLimited(key)
}

// applyKR is woken up when it needs to apply a KaosRule
// and eventually delete a pod (in a match is found)
func (c *Controller) applyKR(namespace string, name string, kr *kaosv1.KaosRule) {

	sel, err := metav1.LabelSelectorAsSelector(kr.Spec.PodSelector)
	if err != nil {
		glog.Fatalf("%s Error parsing PodSelector: %v", kr.String(), err)
		c.recorder.Event(kr, corev1.EventTypeWarning, PodSelectingError, fmt.Sprintf("Error selecting pods: %v", err))
		return
	}

	glog.V(4).Info(fmt.Sprintf("%s Apply filtered rule (filter=%s)", kr.String(), sel.String()))
	list, err := c.kubeclientset.CoreV1().Pods(namespace).List(metav1.ListOptions{LabelSelector: sel.String()})
	if err != nil {
		glog.Fatalf("%s Error listing pods: %v", kr.String(), err)
		c.recorder.Event(kr, corev1.EventTypeWarning, PodListingError, fmt.Sprintf("Error listing pods: %v", err))
		return
	}

	glog.V(4).Info(fmt.Sprintf("%s Apply rule got %d items from list %s", kr.String(), len(list.Items), list))
	if len(list.Items) == 0 {
		glog.V(2).Info(fmt.Sprintf("%s No pod matching (%s), leaving", kr.String(), sel.String()))
		c.recorder.Event(kr, corev1.EventTypeWarning, PodListingEmpty, fmt.Sprintf("No pods matching %s", sel.String()))
		return
	}

	victimIndex := c.rand.Intn(len(list.Items))
	victimName := list.Items[victimIndex].Name

	glog.V(2).Info(fmt.Sprintf("%s About to delete pod index %d from list matching %s", kr.String(), victimIndex, sel.String()))
	err = c.kubeclientset.CoreV1().Pods(namespace).Delete(victimName, nil)
	if err != nil && !errors.IsConflict(err) && !errors.IsNotFound(err) {
		glog.Fatalf("%s Error deleting pod %s in: %v", kr.String(), victimName, err)
		c.recorder.Event(kr, corev1.EventTypeWarning, PodDeletingError, fmt.Sprintf("Error deleting pod %s: %v", victimName, err))
	} else {
		c.recorder.Event(kr, corev1.EventTypeNormal, KaosCreated, fmt.Sprintf("Pod %s has been deleted", victimName))
	}
}
