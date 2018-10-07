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
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	caching "github.com/knative/caching/pkg/apis/caching/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/mattmoor/cachier/pkg/apis/podspec/v1alpha1"
)

func TestMakeImages(t *testing.T) {
	boolTrue := true

	tests := []struct {
		name string
		ps   *v1alpha1.WithPod
		want map[string]caching.Image
	}{{
		name: "no containers",
		ps: &v1alpha1.WithPod{
			ObjectMeta: metav1.ObjectMeta{
				Name:       "foo",
				Namespace:  "bar",
				UID:        "asdfdffsd",
				Generation: 98345,
			},
			Spec: v1alpha1.WithPodSpec{
				Template: v1alpha1.PodSpecable{
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{},
					},
				},
			},
		},
		want: map[string]caching.Image{},
	}, {
		name: "single container",
		ps: &v1alpha1.WithPod{
			ObjectMeta: metav1.ObjectMeta{
				Name:       "foo",
				Namespace:  "bar",
				UID:        "deadbeef",
				Generation: 37837,
			},
			Spec: v1alpha1.WithPodSpec{
				Template: v1alpha1.PodSpecable{
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{{
							Image: "busybox",
						}},
					},
				},
			},
		},
		want: map[string]caching.Image{
			"busybox": {
				ObjectMeta: metav1.ObjectMeta{
					GenerateName: "foo",
					Namespace:    "bar",
					Labels: map[string]string{
						"controller": "deadbeef",
						"generation": "37837",
					},
					OwnerReferences: []metav1.OwnerReference{{
						Name:               "foo",
						UID:                "deadbeef",
						Controller:         &boolTrue,
						BlockOwnerDeletion: &boolTrue,
					}},
				},
				Spec: caching.ImageSpec{
					Image: "busybox",
				},
			}},
	}, {
		name: "multiple containers",
		ps: &v1alpha1.WithPod{
			ObjectMeta: metav1.ObjectMeta{
				Name:       "foo",
				Namespace:  "bar",
				UID:        "deadbeef",
				Generation: 37837,
			},
			Spec: v1alpha1.WithPodSpec{
				Template: v1alpha1.PodSpecable{
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{{
							Image: "busybox",
						}, {
							Image: "hello-world",
						}, {
							Image: "k8s.gcr.io/pause:latest",
						}},
					},
				},
			},
		},
		want: map[string]caching.Image{
			"busybox": {
				ObjectMeta: metav1.ObjectMeta{
					GenerateName: "foo",
					Namespace:    "bar",
					Labels: map[string]string{
						"controller": "deadbeef",
						"generation": "37837",
					},
					OwnerReferences: []metav1.OwnerReference{{
						Name:               "foo",
						UID:                "deadbeef",
						Controller:         &boolTrue,
						BlockOwnerDeletion: &boolTrue,
					}},
				},
				Spec: caching.ImageSpec{
					Image: "busybox",
				},
			},
			"hello-world": {
				ObjectMeta: metav1.ObjectMeta{
					GenerateName: "foo",
					Namespace:    "bar",
					Labels: map[string]string{
						"controller": "deadbeef",
						"generation": "37837",
					},
					OwnerReferences: []metav1.OwnerReference{{
						Name:               "foo",
						UID:                "deadbeef",
						Controller:         &boolTrue,
						BlockOwnerDeletion: &boolTrue,
					}},
				},
				Spec: caching.ImageSpec{
					Image: "hello-world",
				},
			},
			"k8s.gcr.io/pause:latest": {
				ObjectMeta: metav1.ObjectMeta{
					GenerateName: "foo",
					Namespace:    "bar",
					Labels: map[string]string{
						"controller": "deadbeef",
						"generation": "37837",
					},
					OwnerReferences: []metav1.OwnerReference{{
						Name:               "foo",
						UID:                "deadbeef",
						Controller:         &boolTrue,
						BlockOwnerDeletion: &boolTrue,
					}},
				},
				Spec: caching.ImageSpec{
					Image: "k8s.gcr.io/pause:latest",
				},
			},
		},
	}, {
		name: "with service account",
		ps: &v1alpha1.WithPod{
			ObjectMeta: metav1.ObjectMeta{
				Name:       "foo",
				Namespace:  "bar",
				UID:        "deadbeef",
				Generation: 37837,
			},
			Spec: v1alpha1.WithPodSpec{
				Template: v1alpha1.PodSpecable{
					Spec: corev1.PodSpec{
						ServiceAccountName: "T-1000",
						Containers: []corev1.Container{{
							Image: "busybox",
						}},
					},
				},
			},
		},
		want: map[string]caching.Image{
			"busybox": {
				ObjectMeta: metav1.ObjectMeta{
					GenerateName: "foo",
					Namespace:    "bar",
					Labels: map[string]string{
						"controller": "deadbeef",
						"generation": "37837",
					},
					OwnerReferences: []metav1.OwnerReference{{
						Name:               "foo",
						UID:                "deadbeef",
						Controller:         &boolTrue,
						BlockOwnerDeletion: &boolTrue,
					}},
				},
				Spec: caching.ImageSpec{
					ServiceAccountName: "T-1000",
					Image:              "busybox",
				},
			},
		},
	}, {
		name: "with image pull secrets",
		ps: &v1alpha1.WithPod{
			ObjectMeta: metav1.ObjectMeta{
				Name:       "foo",
				Namespace:  "bar",
				UID:        "deadbeef",
				Generation: 37837,
			},
			Spec: v1alpha1.WithPodSpec{
				Template: v1alpha1.PodSpecable{
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{{
							Image: "busybox",
						}},
						ImagePullSecrets: []corev1.LocalObjectReference{{
							Name: "secret1",
						}, {
							Name: "secret2",
						}, {
							Name: "secret3",
						}},
					},
				},
			},
		},
		want: map[string]caching.Image{
			"busybox": {
				ObjectMeta: metav1.ObjectMeta{
					GenerateName: "foo",
					Namespace:    "bar",
					Labels: map[string]string{
						"controller": "deadbeef",
						"generation": "37837",
					},
					OwnerReferences: []metav1.OwnerReference{{
						Name:               "foo",
						UID:                "deadbeef",
						Controller:         &boolTrue,
						BlockOwnerDeletion: &boolTrue,
					}},
				},
				Spec: caching.ImageSpec{
					Image: "busybox",
					ImagePullSecrets: []corev1.LocalObjectReference{{
						Name: "secret1",
					}, {
						Name: "secret2",
					}, {
						Name: "secret3",
					}},
				},
			},
		},
	}, {
		name: "single unique container",
		ps: &v1alpha1.WithPod{
			ObjectMeta: metav1.ObjectMeta{
				Name:       "foo",
				Namespace:  "bar",
				UID:        "deadbeef",
				Generation: 37837,
			},
			Spec: v1alpha1.WithPodSpec{
				Template: v1alpha1.PodSpecable{
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{{
							Image: "busybox",
						}, {
							Image: "busybox",
						}, {
							Image: "busybox",
						}},
					},
				},
			},
		},
		want: map[string]caching.Image{
			"busybox": {
				ObjectMeta: metav1.ObjectMeta{
					GenerateName: "foo",
					Namespace:    "bar",
					Labels: map[string]string{
						"controller": "deadbeef",
						"generation": "37837",
					},
					OwnerReferences: []metav1.OwnerReference{{
						Name:               "foo",
						UID:                "deadbeef",
						Controller:         &boolTrue,
						BlockOwnerDeletion: &boolTrue,
					}},
				},
				Spec: caching.ImageSpec{
					Image: "busybox",
				},
			},
		},
	}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := MakeImages(test.ps)
			if diff := cmp.Diff(test.want, got, cmpopts.IgnoreUnexported(resource.Quantity{})); diff != "" {
				t.Errorf("MakeImages (-want, +got) = %v", diff)
			}
		})
	}
}
