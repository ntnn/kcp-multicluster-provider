/*
Copyright 2026 The kcp Authors.

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

package cache

import (
	"context"
	"sync"

	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Lister can list objects across all registered shard caches.
type Lister interface {
	List(ctx context.Context, list client.ObjectList, opts ...client.ListOption) error
}

// AggregateCache provides an aggregated view across multiple per-shard WildcardCaches.
type AggregateCache struct {
	lock   sync.RWMutex
	caches map[string]WildcardCache
}

// NewAggregateCache returns an initialized AggregateCache.
func NewAggregateCache() *AggregateCache {
	return &AggregateCache{
		caches: make(map[string]WildcardCache),
	}
}

// AddCache registers a WildcardCache under the given id.
func (a *AggregateCache) AddCache(id string, c WildcardCache) {
	a.lock.Lock()
	defer a.lock.Unlock()
	a.caches[id] = c
}

// RemoveCache unregisters the WildcardCache with the given id.
func (a *AggregateCache) RemoveCache(id string) {
	a.lock.Lock()
	defer a.lock.Unlock()
	delete(a.caches, id)
}

// List aggregates results from all registered caches.
func (a *AggregateCache) List(ctx context.Context, list client.ObjectList, opts ...client.ListOption) error {
	a.lock.RLock()
	defer a.lock.RUnlock()

	var allItems []runtime.Object

	for _, c := range a.caches {
		if err := c.List(ctx, list, opts...); err != nil {
			continue
		}

		items, err := meta.ExtractList(list)
		if err != nil {
			continue
		}
		allItems = append(allItems, items...)
	}

	return meta.SetList(list, allItems)
}
