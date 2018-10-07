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

package cachier

import (
	"context"
	"strings"

	"github.com/knative/pkg/apis/duck"
	"github.com/knative/pkg/controller"
	"github.com/knative/pkg/logging"
	"github.com/knative/pkg/logging/logkey"
	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/tools/cache"

	"github.com/mattmoor/cachier/pkg/apis/podspec/v1alpha1"
)

const controllerAgentName = "cachier-controller"

const annotationKey = "cachier.mattmoor.io/decorate"

// Reconciler is the controller implementation for PodSpecable resources
type Reconciler struct {
	lister cache.GenericLister

	// Sugared logger is easier to use but is not as performant as the
	// raw logger. In performance critical paths, call logger.Desugar()
	// and use the returned raw logger instead. In addition to the
	// performance benefits, raw logger also preserves type-safety at
	// the expense of slightly greater verbosity.
	Logger *zap.SugaredLogger
}

// Check that we implement the controller.Reconciler interface.
var _ controller.Reconciler = (*Reconciler)(nil)

// NewController returns a new PodSpecable controller
func NewController(
	logger *zap.SugaredLogger,
	dynamicClient dynamic.Interface,
	psif duck.InformerFactory,
	gvr schema.GroupVersionResource,
) *controller.Impl {

	// Enrich the logs with controller name
	r := &Reconciler{
		Logger: logger.Named(controllerAgentName).
			With(zap.String(logkey.ControllerType, controllerAgentName)),
	}
	impl := controller.NewImpl(r, r.Logger, gvr.String())

	r.Logger.Info("Setting up event handlers")
	cif := &duck.CachedInformerFactory{
		Delegate: &duck.EnqueueInformerFactory{
			Delegate: psif,
			EventHandler: cache.ResourceEventHandlerFuncs{
				AddFunc:    impl.Enqueue,
				UpdateFunc: controller.PassNew(impl.Enqueue),
			},
		},
	}

	_, lister, err := cif.Get(gvr)
	if err != nil {
		r.Logger.Fatalf("Error building informer for %v: %v", gvr, err)
	}
	r.lister = lister

	return impl
}

// Reconcile implements controller.Reconciler
func (c *Reconciler) Reconcile(ctx context.Context, key string) error {
	logger := logging.FromContext(ctx)
	// Convert the namespace/name string into a distinct namespace and name
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		logger.Errorf("invalid resource key: %s", key)
		return nil
	}

	// Get the thing resource with this namespace/name
	untyped, err := c.lister.ByNamespace(namespace).Get(name)
	if errors.IsNotFound(err) {
		logger.Errorf("thing %q in work queue no longer exists", key)
		return nil
	} else if err != nil {
		return err
	}
	thing := untyped.(*v1alpha1.WithPod)

	if v, ok := thing.Annotations[annotationKey]; !ok {
		// Default is to decorate
	} else {
		switch strings.ToLower(v) {
		case "false", "off", "disable", "disabled":
			logger.Errorf("Decoration is disabled for %v", key)
			return nil
		}
	}

	// TODO(mattmoor): Reconcile Image sub-resources.
	logger.Errorf("TODO: Reconcile %T: %v", thing, key)

	return nil
}
