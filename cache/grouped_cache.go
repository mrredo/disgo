package cache

import (
	"sync"

	"github.com/disgoorg/snowflake/v2"
)

type GroupedCacheFilterFunc[T any] func(groupID snowflake.ID, entity T) bool

type GroupedCache[T any] interface {
	Get(groupID snowflake.ID, id snowflake.ID) (T, bool)
	Put(groupID snowflake.ID, id snowflake.ID, entity T)
	Remove(groupID snowflake.ID, id snowflake.ID) (T, bool)
	RemoveAll(groupID snowflake.ID)
	RemoveIf(filterFunc GroupedCacheFilterFunc[T])

	All() map[snowflake.ID][]T
	GroupAll(groupID snowflake.ID) []T

	MapAll() map[snowflake.ID]map[snowflake.ID]T
	MapGroupAll(groupID snowflake.ID) map[snowflake.ID]T

	FindFirst(cacheFindFunc GroupedCacheFilterFunc[T]) (T, bool)
	GroupFindFirst(groupID snowflake.ID, cacheFindFunc GroupedCacheFilterFunc[T]) (T, bool)

	FindAll(cacheFindFunc GroupedCacheFilterFunc[T]) []T
	GroupFindAll(groupID snowflake.ID, cacheFindFunc GroupedCacheFilterFunc[T]) []T

	ForEach(func(groupID snowflake.ID, entity T))
	ForEachGroup(groupID snowflake.ID, forEachFunc func(entity T))
}

var _ GroupedCache[any] = (*DefaultGroupedCache[any])(nil)

func NewGroupedCache[T any](flags Flags, neededFlags Flags, policy Policy[T]) GroupedCache[T] {
	return &DefaultGroupedCache[T]{
		flags:       flags,
		neededFlags: neededFlags,
		policy:      policy,
		cache:       make(map[snowflake.ID]map[snowflake.ID]T),
	}
}

type DefaultGroupedCache[T any] struct {
	mu          sync.RWMutex
	flags       Flags
	neededFlags Flags
	policy      Policy[T]
	cache       map[snowflake.ID]map[snowflake.ID]T
}

func (c *DefaultGroupedCache[T]) Get(groupID snowflake.ID, id snowflake.ID) (T, bool) {
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

func (c *DefaultGroupedCache[T]) Put(groupID snowflake.ID, id snowflake.ID, entity T) {
	if c.neededFlags != FlagsNone && c.flags.Missing(c.neededFlags) {
		return
	}
	if c.policy != nil && !c.policy(entity) {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.cache == nil {
		c.cache = make(map[snowflake.ID]map[snowflake.ID]T)
	}

	if groupEntities, ok := c.cache[groupID]; ok {
		groupEntities[id] = entity
	} else {
		groupEntities = make(map[snowflake.ID]T)
		groupEntities[id] = entity
		c.cache[groupID] = groupEntities
	}
}

func (c *DefaultGroupedCache[T]) Remove(groupID snowflake.ID, id snowflake.ID) (entity T, ok bool) {
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

func (c *DefaultGroupedCache[T]) RemoveAll(groupID snowflake.ID) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.cache, groupID)
}

func (c *DefaultGroupedCache[T]) RemoveIf(filterFunc GroupedCacheFilterFunc[T]) {
	c.mu.Lock()
	defer c.mu.Unlock()

	for groupID := range c.cache {
		for id, entity := range c.cache[groupID] {
			if filterFunc(groupID, entity) {
				delete(c.cache[groupID], id)
			}
		}
	}
}

func (c *DefaultGroupedCache[T]) All() map[snowflake.ID][]T {
	c.mu.RLock()
	defer c.mu.RUnlock()

	all := make(map[snowflake.ID][]T)
	for groupID, groupEntities := range c.cache {
		all[groupID] = make([]T, 0, len(groupEntities))
		for _, entity := range groupEntities {
			all[groupID] = append(all[groupID], entity)
		}
	}

	return all
}

func (c *DefaultGroupedCache[T]) GroupAll(groupID snowflake.ID) []T {
	c.mu.RLock()
	defer c.mu.RUnlock()

	groupEntities, ok := c.cache[groupID]
	if !ok {
		return nil
	}
	all := make([]T, 0, len(groupEntities))
	for _, entity := range groupEntities {
		all = append(all, entity)
	}

	return all
}

func (c *DefaultGroupedCache[T]) MapAll() map[snowflake.ID]map[snowflake.ID]T {
	c.mu.RLock()
	defer c.mu.RUnlock()

	all := make(map[snowflake.ID]map[snowflake.ID]T, len(c.cache))
	for groupID, groupEntities := range c.cache {
		all[groupID] = make(map[snowflake.ID]T, len(groupEntities))
		for entityID, entity := range groupEntities {
			all[groupID][entityID] = entity
		}
	}

	return all
}

func (c *DefaultGroupedCache[T]) MapGroupAll(groupID snowflake.ID) map[snowflake.ID]T {
	c.mu.RLock()
	defer c.mu.RUnlock()

	groupEntities, ok := c.cache[groupID]
	if !ok {
		return nil
	}
	all := make(map[snowflake.ID]T, len(groupEntities))
	for entityID, entity := range groupEntities {
		all[entityID] = entity
	}

	return all
}

func (c *DefaultGroupedCache[T]) FindFirst(cacheFindFunc GroupedCacheFilterFunc[T]) (T, bool) {
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

func (c *DefaultGroupedCache[T]) GroupFindFirst(groupID snowflake.ID, cacheFindFunc GroupedCacheFilterFunc[T]) (T, bool) {
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

func (c *DefaultGroupedCache[T]) FindAll(cacheFindFunc GroupedCacheFilterFunc[T]) []T {
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

func (c *DefaultGroupedCache[T]) GroupFindAll(groupID snowflake.ID, cacheFindFunc GroupedCacheFilterFunc[T]) []T {
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

func (c *DefaultGroupedCache[T]) ForEach(forEachFunc func(groupID snowflake.ID, entity T)) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	for groupID, groupEntities := range c.cache {
		for _, entity := range groupEntities {
			forEachFunc(groupID, entity)
		}
	}
}
func (c *DefaultGroupedCache[T]) ForEachGroup(groupID snowflake.ID, forEachFunc func(entity T)) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	for _, entity := range c.cache[groupID] {
		forEachFunc(entity)
	}
}
