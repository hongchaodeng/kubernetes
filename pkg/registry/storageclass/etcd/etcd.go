/*
Copyright 2015 The Kubernetes Authors.

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

package etcd

import (
	"github.com/golang/glog"
	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/apis/extensions"
	"k8s.io/kubernetes/pkg/registry/generic"
	"k8s.io/kubernetes/pkg/registry/generic/registry"
	"k8s.io/kubernetes/pkg/registry/storageclass"
	"k8s.io/kubernetes/pkg/runtime"
	"k8s.io/kubernetes/pkg/storage"
)

type REST struct {
	*registry.Store
}

// NewREST returns a RESTStorage object that will work against persistent volumes.
func NewREST(opts generic.RESTOptions) *REST {
	prefix := "/storageclasses"

	newListFunc := func() runtime.Object { return &extensions.StorageClassList{} }
	s := registry.StorageWithCacher(
		opts.StorageConfig,
		100,
		&extensions.StorageClass{},
		prefix,
		storageclass.Strategy,
		newListFunc,
		storage.NoTriggerPublisher,
	)

	store := &registry.Store{
		NewFunc:     func() runtime.Object { return &extensions.StorageClass{} },
		NewListFunc: newListFunc,
		KeyRootFunc: func(ctx api.Context) string {
			return prefix
		},
		KeyFunc: func(ctx api.Context, name string) (string, error) {
			return registry.NoNamespaceKeyFunc(ctx, prefix, name)
		},
		ObjectNameFunc: func(obj runtime.Object) (string, error) {
			return obj.(*extensions.StorageClass).Name, nil
		},
		PredicateFunc:           storageclass.MatchStorageClasses,
		QualifiedResource:       api.Resource("storageclasses"),
		DeleteCollectionWorkers: opts.DeleteCollectionWorkers,

		CreateStrategy:      storageclass.Strategy,
		UpdateStrategy:      storageclass.Strategy,
		DeleteStrategy:      storageclass.Strategy,
		ReturnDeletedObject: true,

		Storage: s,
		FVGetFunc: func(field string, obj runtime.Object) (string, bool) {
			o, ok := obj.(*extensions.StorageClass)
			if !ok {
				glog.Warningf("Unexpected type: %T", obj)
				return "", false
			}
			return registry.GetFVCommon(field, o.Labels, storageclass.StorageClassToSelectableFields(o))
		},
	}

	return &REST{store}
}
