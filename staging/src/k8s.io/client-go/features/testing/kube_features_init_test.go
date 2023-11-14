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

package testing

import (
	"testing"

	"k8s.io/client-go/features/clientgofeaturegates"

	"github.com/stretchr/testify/require"

	"k8s.io/client-go/features"
)

func TestDriveInitDefaultFeatureGates(t *testing.T) {
	defaultFeatureGates := features.DefaultFeatureGates()
	require.Panics(t, func() { defaultFeatureGates.Enabled("FakeFeatureGate") })

	featureGate := &fakeReader{}
	require.True(t, featureGate.Enabled("FakeFeatureGate"))

	features.SetFeatureGates(featureGate)
	defaultFeatureGates = features.DefaultFeatureGates()
	require.True(t, defaultFeatureGates.Enabled("FakeFeatureGate"))
}

type fakeReader struct{}

func (f *fakeReader) Enabled(clientgofeaturegates.Feature) bool {
	return true
}
