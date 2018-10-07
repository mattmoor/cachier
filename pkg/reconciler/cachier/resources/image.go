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

package resources

import (
	"fmt"

	caching "github.com/knative/caching/pkg/apis/caching/v1alpha1"
	"github.com/knative/pkg/kmeta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/mattmoor/cachier/pkg/apis/podspec/v1alpha1"
)

func MakeImages(ps *v1alpha1.WithPod) map[string]caching.Image {
	images := make(map[string]caching.Image)
	// Build the deduplicated set of Image resources.
	podspec := ps.Spec.Template.Spec
	for idx, c := range podspec.Containers {
		if _, ok := images[c.Image]; ok {
			continue
		}
		images[c.Image] = caching.Image{
			ObjectMeta: metav1.ObjectMeta{
				GenerateName:    fmt.Sprintf("%s-%02d-", ps.Name, idx),
				Namespace:       ps.Namespace,
				Labels:          kmeta.MakeGenerationLabels(ps),
				OwnerReferences: []metav1.OwnerReference{*kmeta.NewControllerRef(ps)},
			},
			Spec: caching.ImageSpec{
				Image:              c.Image,
				ServiceAccountName: podspec.ServiceAccountName,
				ImagePullSecrets:   podspec.ImagePullSecrets,
			},
		}
	}
	return images
}
