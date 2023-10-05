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
	"k8s.io/client-go/features/clientgofeaturegates"
)

// DefaultFeatureGates returns the feature gates exposed by this library.
//
// By default, only the default features gate will be returned.
// The default implementation allows controlling the features
// via environmental variables.
// For example, if you have a feature named "MyFeature,"
// setting an environmental variable "KUBE_FEATURE_MyFeature"
// will allow you to configure the state of that feature.
//
// Please note that the actual set of the features gates
// might be overwritten by calling SetFeatureGates method.
func DefaultFeatureGates() clientgofeaturegates.Reader {
	return featureGates.Load().(clientgofeaturegates.Reader)
}

const (
	// Every feature gate should add method here following this template:
	//
	// // owner: @username
	// // alpha: v1.4
	// MyFeature featuregate.Feature = "MyFeature"
	//
	// Feature gates should be listed in alphabetical, case-sensitive
	// (upper before any lower case character) order. This reduces the risk
	// of code conflicts because changes are more likely to be scattered
	// across the file.

	// owner: @p0lyn0mial
	// alpha: v1.27
	// beta: v1.29
	//
	// Allow the API server to stream individual items instead of chunking
	WatchList clientgofeaturegates.Feature = "WatchListClient"
)

// defaultKubernetesFeatureGates consists of all known Kubernetes-specific feature keys.
//
// To add a new feature, define a key for it above and add it here. The features will be
// available throughout Kubernetes binaries.
var defaultKubernetesFeatureGates = map[clientgofeaturegates.Feature]clientgofeaturegates.FeatureSpec{
	WatchList: {Default: false, PreRelease: clientgofeaturegates.Beta},
}
