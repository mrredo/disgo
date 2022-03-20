package core

import (
	"sync"

	"github.com/DisgoOrg/snowflake"
)

type GroupedCacheFindFunc[T any] func(groupID snowflake.Snowflake, entity T) bool

type GroupedCache[T any] interface {
	Get(groupID snowflake.Snowflake, id snowflake.Snowflake) (T, bool)
	Put(groupID snowflake.Snowflake, id snowflake.Snowflake, entity T)
	Remove(groupID snowflake.Snowflake, id snowflake.Snowflake) (T, bool)
	RemoveAll(groupID snowflake.Snowflake)

	All() map[snowflake.Snowflake][]T
	GroupAll(groupID snowflake.Snowflake) []T

	FindFirst(cacheFindFunc GroupedCacheFindFunc[T]) (T, bool)
	GroupFindFirst(groupID snowflake.Snowflake, cacheFindFunc GroupedCacheFindFunc[T]) (T, bool)

	FindAll(cacheFindFunc GroupedCacheFindFunc[T]) []T
	GroupFindAll(groupID snowflake.Snowflake, cacheFindFunc GroupedCacheFindFunc[T]) []T

	ForEach(func(groupID snowflake.Snowflake, entity T))
	ForEachGroup(groupID snowflake.Snowflake, forEachFunc func(entity T))
}

var _ GroupedCache[any] = (*DefaultGroupedCache[any])(nil)

func NewGroupedCache[T any](flags CacheFlags, neededFlags CacheFlags, policy CachePolicy[T]) GroupedCache[T] {
	return &DefaultGroupedCache[T]{
		flags:       flags,
		neededFlags: neededFlags,
		policy:      policy,
		cache:       make(map[snowflake.Snowflake]map[snowflake.Snowflake]T),
	}
}

type DefaultGroupedCache[T any] struct {
	mu          sync.RWMutex
	flags       CacheFlags
	neededFlags CacheFlags
	policy      CachePolicy[T]
	cache       map[snowflake.Snowflake]map[snowflake.Snowflake]T
}

func (c *DefaultGroupedCache[T]) Get(groupID snowflake.Snowflake, id snowflake.Snowflake) (T, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if groupEntities, ok := c.cache[groupID]; ok {
		if entity, ok := groupEntities[id]; ok {
			return entity, true
		}
	}

	var entity T
	return entity, false
}

func (c *DefaultGroupedCache[T]) Put(groupID snowflake.Snowflake, id snowflake.Snowflake, entity T) {
	if c.neededFlags != CacheFlagsNone && c.flags.Missing(c.neededFlags) {
		return
	}
	if c.policy != nil && !c.policy(entity) {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.cache == nil {
		c.cache = make(map[snowflake.Snowflake]map[snowflake.Snowflake]T)
	}

	if groupEntities, ok := c.cache[groupID]; ok {
		groupEntities[id] = entity
	} else {
		groupEntities = make(map[snowflake.Snowflake]T)
		groupEntities[id] = entity
		c.cache[groupID] = groupEntities
	}
}

func (c *DefaultGroupedCache[T]) Remove(groupID snowflake.Snowflake, id snowflake.Snowflake) (entity T, ok bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if groupEntities, ok := c.cache[groupID]; ok {
		if entity, ok := groupEntities[id]; ok {
			delete(groupEntities, id)
			return entity, ok
		}
	}
	ok = false
	return
}

func (c *DefaultGroupedCache[T]) RemoveAll(groupID snowflake.Snowflake) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.cache, groupID)
}

func (c *DefaultGroupedCache[T]) Cache() map[snowflake.Snowflake]map[snowflake.Snowflake]T {
	return c.cache
}

func (c *DefaultGroupedCache[T]) All() map[snowflake.Snowflake][]T {
	c.mu.RLock()
	defer c.mu.RUnlock()

	all := make(map[snowflake.Snowflake][]T)
	for groupID, groupEntities := range c.cache {
		all[groupID] = make([]T, 0, len(groupEntities))
		for _, entity := range groupEntities {
			all[groupID] = append(all[groupID], entity)
		}
	}

	return all
}

func (c *DefaultGroupedCache[T]) GroupCache(groupID snowflake.Snowflake) map[snowflake.Snowflake]T {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if groupEntities, ok := c.cache[groupID]; ok {
		return groupEntities
	}

	return nil
}

func (c *DefaultGroupedCache[T]) GroupAll(groupID snowflake.Snowflake) []T {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if groupEntities, ok := c.cache[groupID]; ok {
		all := make([]T, 0, len(groupEntities))
		for _, entity := range groupEntities {
			all = append(all, entity)
		}

		return all
	}

	return nil
}

func (c *DefaultGroupedCache[T]) FindFirst(cacheFindFunc GroupedCacheFindFunc[T]) (T, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	for groupID, groupEntities := range c.cache {
		for _, entity := range groupEntities {
			if cacheFindFunc(groupID, entity) {
				return entity, true
			}
		}
	}

	var entity T
	return entity, false
}

func (c *DefaultGroupedCache[T]) GroupFindFirst(groupID snowflake.Snowflake, cacheFindFunc GroupedCacheFindFunc[T]) (T, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	for _, entity := range c.cache[groupID] {
		if cacheFindFunc(groupID, entity) {
			return entity, true
		}
	}

	var entity T
	return entity, false
}

func (c *DefaultGroupedCache[T]) FindAll(cacheFindFunc GroupedCacheFindFunc[T]) []T {
	c.mu.RLock()
	defer c.mu.RUnlock()

	all := make([]T, 0)
	for groupID, groupEntities := range c.cache {
		for _, entity := range groupEntities {
			if cacheFindFunc(groupID, entity) {
				all = append(all, entity)
			}
		}
	}
	return all
}

func (c *DefaultGroupedCache[T]) GroupFindAll(groupID snowflake.Snowflake, cacheFindFunc GroupedCacheFindFunc[T]) []T {
	c.mu.RLock()
	defer c.mu.RUnlock()

	all := make([]T, 0)
	for _, entity := range c.cache[groupID] {
		if cacheFindFunc(groupID, entity) {
			all = append(all, entity)
		}
	}
	return all
}

func (c *DefaultGroupedCache[T]) ForEach(forEachFunc func(groupID snowflake.Snowflake, entity T)) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	for groupID, groupEntities := range c.cache {
		for _, entity := range groupEntities {
			forEachFunc(groupID, entity)
		}
	}
}
func (c *DefaultGroupedCache[T]) ForEachGroup(groupID snowflake.Snowflake, forEachFunc func(entity T)) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	for _, entity := range c.cache[groupID] {
		forEachFunc(entity)
	}
}
