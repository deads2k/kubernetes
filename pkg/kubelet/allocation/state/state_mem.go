/*
Copyright 2021 The Kubernetes Authors.

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

package state

import (
	"sync"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/klog/v2"
)

type stateMemory struct {
	sync.RWMutex
	podAllocation PodResourceAllocation
}

var _ State = &stateMemory{}

// NewStateMemory creates new State to track resources allocated to pods
func NewStateMemory(alloc PodResourceAllocation) State {
	if alloc == nil {
		alloc = PodResourceAllocation{}
	}
	klog.V(2).InfoS("Initialized new in-memory state store for pod resource allocation tracking")
	return &stateMemory{
		podAllocation: alloc,
	}
}

func (s *stateMemory) GetContainerResourceAllocation(podUID types.UID, containerName string) (v1.ResourceRequirements, bool) {
	s.RLock()
	defer s.RUnlock()

	alloc, ok := s.podAllocation[podUID][containerName]
	return *alloc.DeepCopy(), ok
}

func (s *stateMemory) GetPodResourceAllocation() PodResourceAllocation {
	s.RLock()
	defer s.RUnlock()
	return s.podAllocation.Clone()
}

func (s *stateMemory) SetContainerResourceAllocation(podUID types.UID, containerName string, alloc v1.ResourceRequirements) error {
	s.Lock()
	defer s.Unlock()

	if _, ok := s.podAllocation[podUID]; !ok {
		s.podAllocation[podUID] = make(map[string]v1.ResourceRequirements)
	}

	s.podAllocation[podUID][containerName] = alloc
	klog.V(3).InfoS("Updated container resource allocation", "podUID", podUID, "containerName", containerName, "alloc", alloc)
	return nil
}

func (s *stateMemory) SetPodResourceAllocation(podUID types.UID, alloc map[string]v1.ResourceRequirements) error {
	s.Lock()
	defer s.Unlock()

	s.podAllocation[podUID] = alloc
	klog.V(3).InfoS("Updated pod resource allocation", "podUID", podUID, "allocation", alloc)
	return nil
}

func (s *stateMemory) RemovePod(podUID types.UID) error {
	s.Lock()
	defer s.Unlock()
	delete(s.podAllocation, podUID)
	klog.V(3).InfoS("Deleted pod resource allocation", "podUID", podUID)
	return nil
}

func (s *stateMemory) RemoveOrphanedPods(remainingPods sets.Set[types.UID]) {
	s.Lock()
	defer s.Unlock()

	for podUID := range s.podAllocation {
		if _, ok := remainingPods[types.UID(podUID)]; !ok {
			delete(s.podAllocation, podUID)
		}
	}
}
