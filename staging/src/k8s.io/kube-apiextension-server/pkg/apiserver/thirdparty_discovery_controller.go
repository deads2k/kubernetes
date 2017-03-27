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

package apiserver

import (
	"fmt"
	"time"

	"github.com/golang/glog"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime/schema"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/apiserver/pkg/endpoints/discovery"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"

	"k8s.io/kube-apiextension-server/pkg/apis/apiextensions"
	informers "k8s.io/kube-apiextension-server/pkg/client/informers/internalversion/apiextensions/internalversion"
	listers "k8s.io/kube-apiextension-server/pkg/client/listers/apiextensions/internalversion"
)

type TPRDiscoveryController struct {
	versionHandler *tprVersionDiscoveryHandler
	groupHandler   *tprGroupDiscoveryHandler

	thirdPartyLister   listers.ThirdPartyLister
	thirdPartiesSynced cache.InformerSynced

	// To allow injection for testing.
	syncFn func(version schema.GroupVersion) error

	queue workqueue.RateLimitingInterface
}

func NewTPRDiscoveryController(tprInformer informers.ThirdPartyInformer, versionHandler *tprVersionDiscoveryHandler, groupHandler *tprGroupDiscoveryHandler) *TPRDiscoveryController {
	c := &TPRDiscoveryController{
		versionHandler:     versionHandler,
		groupHandler:       groupHandler,
		thirdPartyLister:   tprInformer.Lister(),
		thirdPartiesSynced: tprInformer.Informer().HasSynced,

		queue: workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "TPRDiscoveryController"),
	}

	tprInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    c.addThirdParty,
		UpdateFunc: c.updateThirdParty,
		DeleteFunc: c.deleteThirdParty,
	})

	c.syncFn = c.sync

	return c
}

func (c *TPRDiscoveryController) sync(version schema.GroupVersion) error {

	foundVersion := false
	foundGroup := false

	apiVersionsForDiscovery := []metav1.GroupVersionForDiscovery{}
	apiResourcesForDiscovery := []metav1.APIResource{}

	thirdParties, err := c.thirdPartyLister.List(labels.Everything())
	if err != nil {
		return err
	}
	for _, thirdParty := range thirdParties {
		if thirdParty.Spec.Group != version.Group {
			continue
		}
		foundGroup = true
		apiVersionsForDiscovery = append(apiVersionsForDiscovery, metav1.GroupVersionForDiscovery{
			GroupVersion: thirdParty.Spec.Group + "/" + thirdParty.Spec.Version,
			Version:      thirdParty.Spec.Version,
		})

		if thirdParty.Spec.Version != version.Version {
			continue
		}
		foundVersion = true

		apiResourcesForDiscovery = append(apiResourcesForDiscovery, metav1.APIResource{
			Name:         thirdParty.Spec.Name,
			SingularName: thirdParty.Spec.Singular,
			Namespaced:   !thirdParty.Spec.ClusterScoped,
			Kind:         thirdParty.Spec.Kind,
			Verbs:        metav1.Verbs([]string{"delete", "deletecollection", "get", "list", "patch", "create", "update", "watch"}),
			ShortNames:   thirdParty.Spec.ShortNames,
		})
	}

	if !foundGroup {
		c.groupHandler.unsetDiscovery(version.Group)
		c.versionHandler.unsetDiscovery(version)
		return nil
	}

	apiGroup := metav1.APIGroup{
		Name:             version.Group,
		Versions:         apiVersionsForDiscovery,
		PreferredVersion: apiVersionsForDiscovery[0],
	}
	c.groupHandler.setDiscovery(version.Group, discovery.NewAPIGroupDiscoveryHandler(Codecs, apiGroup))

	if !foundVersion {
		c.versionHandler.unsetDiscovery(version)
		return nil
	}
	c.versionHandler.setDiscovery(version, discovery.NewAPIVersionDiscoveryHandlerForResources(Codecs, version, apiResourcesForDiscovery))

	return nil
}

func (c *TPRDiscoveryController) Run(stopCh <-chan struct{}) {
	defer utilruntime.HandleCrash()
	defer c.queue.ShutDown()
	defer glog.Infof("Shutting down TPRDiscoveryController")

	glog.Infof("Starting TPRDiscoveryController")

	if !cache.WaitForCacheSync(stopCh, c.thirdPartiesSynced) {
		utilruntime.HandleError(fmt.Errorf("timed out waiting for caches to sync"))
		return
	}

	// only start one worker thread since its a slow moving API
	go wait.Until(c.runWorker, time.Second, stopCh)

	<-stopCh
}

func (c *TPRDiscoveryController) runWorker() {
	for c.processNextWorkItem() {
	}
}

// processNextWorkItem deals with one key off the queue.  It returns false when it's time to quit.
func (c *TPRDiscoveryController) processNextWorkItem() bool {
	key, quit := c.queue.Get()
	if quit {
		return false
	}
	defer c.queue.Done(key)

	err := c.syncFn(key.(schema.GroupVersion))
	if err == nil {
		c.queue.Forget(key)
		return true
	}

	utilruntime.HandleError(fmt.Errorf("%v failed with : %v", key, err))
	c.queue.AddRateLimited(key)

	return true
}

func (c *TPRDiscoveryController) enqueue(obj *apiextensions.ThirdParty) {
	c.queue.Add(schema.GroupVersion{Group: obj.Spec.Group, Version: obj.Spec.Version})
}

func (c *TPRDiscoveryController) addThirdParty(obj interface{}) {
	castObj := obj.(*apiextensions.ThirdParty)
	glog.V(4).Infof("Adding %s", castObj.Name)
	c.enqueue(castObj)
}

func (c *TPRDiscoveryController) updateThirdParty(obj, _ interface{}) {
	castObj := obj.(*apiextensions.ThirdParty)
	glog.V(4).Infof("Updating %s", castObj.Name)
	c.enqueue(castObj)
}

func (c *TPRDiscoveryController) deleteThirdParty(obj interface{}) {
	castObj, ok := obj.(*apiextensions.ThirdParty)
	if !ok {
		tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
		if !ok {
			glog.Errorf("Couldn't get object from tombstone %#v", obj)
			return
		}
		castObj, ok = tombstone.Obj.(*apiextensions.ThirdParty)
		if !ok {
			glog.Errorf("Tombstone contained object that is not expected %#v", obj)
			return
		}
	}
	glog.V(4).Infof("Deleting %q", castObj.Name)
	c.enqueue(castObj)
}
