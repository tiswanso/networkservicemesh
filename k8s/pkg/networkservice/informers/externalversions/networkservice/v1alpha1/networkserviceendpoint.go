// Copyright (c) 2019 Cisco and/or its affiliates.
// Copyright (c) 2019 Red Hat Inc. and/or its affiliates.
// Copyright (c) 2019 VMware, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at:
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Code generated by informer-gen. DO NOT EDIT.

package v1alpha1

import (
	"context"
	time "time"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
	watch "k8s.io/apimachinery/pkg/watch"
	cache "k8s.io/client-go/tools/cache"

	networkservicev1alpha1 "cisco-app-networking.github.io/networkservicemesh/k8s/pkg/apis/networkservice/v1alpha1"
	versioned "cisco-app-networking.github.io/networkservicemesh/k8s/pkg/networkservice/clientset/versioned"
	internalinterfaces "cisco-app-networking.github.io/networkservicemesh/k8s/pkg/networkservice/informers/externalversions/internalinterfaces"
	v1alpha1 "cisco-app-networking.github.io/networkservicemesh/k8s/pkg/networkservice/listers/networkservice/v1alpha1"
)

// NetworkServiceEndpointInformer provides access to a shared informer and lister for
// NetworkServiceEndpoints.
type NetworkServiceEndpointInformer interface {
	Informer() cache.SharedIndexInformer
	Lister() v1alpha1.NetworkServiceEndpointLister
}

type networkServiceEndpointInformer struct {
	factory          internalinterfaces.SharedInformerFactory
	tweakListOptions internalinterfaces.TweakListOptionsFunc
	namespace        string
}

// NewNetworkServiceEndpointInformer constructs a new informer for NetworkServiceEndpoint type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewNetworkServiceEndpointInformer(client versioned.Interface, namespace string, resyncPeriod time.Duration, indexers cache.Indexers) cache.SharedIndexInformer {
	return NewFilteredNetworkServiceEndpointInformer(client, namespace, resyncPeriod, indexers, nil)
}

// NewFilteredNetworkServiceEndpointInformer constructs a new informer for NetworkServiceEndpoint type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewFilteredNetworkServiceEndpointInformer(client versioned.Interface, namespace string, resyncPeriod time.Duration, indexers cache.Indexers, tweakListOptions internalinterfaces.TweakListOptionsFunc) cache.SharedIndexInformer {
	return cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options v1.ListOptions) (runtime.Object, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.NetworkserviceV1alpha1().NetworkServiceEndpoints(namespace).List(context.TODO(), options)
			},
			WatchFunc: func(options v1.ListOptions) (watch.Interface, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.NetworkserviceV1alpha1().NetworkServiceEndpoints(namespace).Watch(context.TODO(), options)
			},
		},
		&networkservicev1alpha1.NetworkServiceEndpoint{},
		resyncPeriod,
		indexers,
	)
}

func (f *networkServiceEndpointInformer) defaultInformer(client versioned.Interface, resyncPeriod time.Duration) cache.SharedIndexInformer {
	return NewFilteredNetworkServiceEndpointInformer(client, f.namespace, resyncPeriod, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc}, f.tweakListOptions)
}

func (f *networkServiceEndpointInformer) Informer() cache.SharedIndexInformer {
	return f.factory.InformerFor(&networkservicev1alpha1.NetworkServiceEndpoint{}, f.defaultInformer)
}

func (f *networkServiceEndpointInformer) Lister() v1alpha1.NetworkServiceEndpointLister {
	return v1alpha1.NewNetworkServiceEndpointLister(f.Informer().GetIndexer())
}
