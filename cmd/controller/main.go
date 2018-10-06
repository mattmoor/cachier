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
	"context"
	"flag"
	"time"

	"github.com/knative/pkg/apis/duck"
	duckv1alpha1 "github.com/knative/pkg/apis/duck/v1alpha1"
	"github.com/knative/pkg/controller"
	"github.com/knative/pkg/logging"
	"github.com/knative/pkg/signals"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/mattmoor/cachier/pkg/reconciler/cachier"
)

const (
	threadsPerController = 2
)

var (
	masterURL  = flag.String("kubeconfig", "", "Path to a kubeconfig. Only required if out-of-cluster.")
	kubeconfig = flag.String("master", "", "The address of the Kubernetes API server. Overrides any value in kubeconfig. Only required if out-of-cluster.")
)

func main() {
	flag.Parse()

	// set up signals so we handle the first shutdown signal gracefully
	stopCh := signals.SetupSignalHandler()

	logger := logging.FromContext(context.TODO()).Named("controller")

	cfg, err := clientcmd.BuildConfigFromFlags(*masterURL, *kubeconfig)
	if err != nil {
		logger.Fatalf("Error building kubeconfig: %s", err.Error())
	}

	dynamicClient, err := dynamic.NewForConfig(cfg)
	if err != nil {
		logger.Fatalf("Error building build clientset: %v", err)
	}

	tif := &duck.TypedInformerFactory{
		Client:       dynamicClient,
		Type:         &duckv1alpha1.Target{},
		ResyncPeriod: 10 * time.Hour,
		StopChannel:  stopCh,
	}

	// Add new controllers here.
	controllers := []*controller.Impl{
		cachier.NewController(
			logger,
			dynamicClient,
			tif,
			schema.GroupVersionResource{
				Group:    "serving.knative.dev",
				Version:  "v1alpha1",
				Resource: "routes",
			},
		),
		cachier.NewController(
			logger,
			dynamicClient,
			tif,
			schema.GroupVersionResource{
				Group:    "serving.knative.dev",
				Version:  "v1alpha1",
				Resource: "services",
			},
		),
	}

	// Start all of the controllers.
	for _, ctrlr := range controllers {
		go func(ctrlr *controller.Impl) {
			// We don't expect this to return until stop is called,
			// but if it does, propagate it back.
			if err := ctrlr.Run(threadsPerController, stopCh); err != nil {
				logger.Fatalf("Error running controller: %s", err.Error())
			}
		}(ctrlr)
	}

	<-stopCh
}
