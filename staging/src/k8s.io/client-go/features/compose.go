/*
Copyright 2023 The Kubernetes Authors.
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

package features

import (
	"sync/atomic"

	"k8s.io/client-go/features/clientgofeaturegates"
)

func init() {
	envVarGates := clientgofeaturegates.NewEnvVarFeatureGate(defaultKubernetesFeatureGates)

	wrappedFeatureGate := &featureGateWrapper{envVarGates}
	featureGates.Store(wrappedFeatureGate)
}

type FeatureRegistry interface {
	Add(map[clientgofeaturegates.Feature]clientgofeaturegates.FeatureSpec) error
}

// AddFeaturesToExistingFeatureGates adds the default feature gates to the provided set.
// Usually this function is combined with SetFeatureGates to take control of the
// features exposed by this library.
func AddFeaturesToExistingFeatureGates(featureRegistry FeatureRegistry) error {
	return featureRegistry.Add(defaultKubernetesFeatureGates)
}

// SetFeatureGates overwrites the default implementation of the feature gate
// used by this library.
//
// Useful for binaries that would like to have full control of the features
// exposed by this library. For example to allow consumers of a binary
// to interact with the features via a command line flag.
func SetFeatureGates(newFeatureGates clientgofeaturegates.Reader) {
	wrappedFeatureGate := &featureGateWrapper{newFeatureGates}
	featureGates.Store(wrappedFeatureGate)
}

// featureGateWrapper a thin wrapper to satisfy featureGates variable (atomic.Value)
type featureGateWrapper struct {
	clientgofeaturegates.Reader
}

var (
	// featureGates is a shared global FeatureGate.
	//
	// Top-level commands/options setup that needs to modify this feature gate
	// should use AddFeaturesToExistingFeatureGates followed by SetFeatureGates.
	featureGates = &atomic.Value{}
)
