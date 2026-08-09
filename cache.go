package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"
)

type CacheManager struct {
	mu       sync.RWMutex
	data     map[string]CacheEntry
	filePath string
}

type CacheEntry struct {
	Value     interface{} `json:"`value"`
	ExpiresAt time.Time   `json:"`expires_at"`
}

func NewCacheManager(filePath string) *CacheManager {
	cm := &CacheManager{
		data:     make(map[string]CacheEntry),
		filePath: filePath,
	}
	cm.load()
	return cm
}

func (cm *CacheManager) Get(key string) (interface{}, bool) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	entry, exists := cm.data[key]
	if !exists {
		return nil, false
	}

	if time.Now().After(entry.ExpiresAt) {
		delete(cm.data, key)
		return nil, false
	}

	return entry.Value, true
}

func (cm *CacheManager) Set(key string, value interface{}, ttl time.Duration) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	cm.data[key] = CacheEntry{
		Value:     value,
		ExpiresAt: time.Now().Add(ttl),
	}

	cm.save()
}

func (cm *CacheManager) Delete(key string) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	delete(cm.data, key)
	cm.save()
}

func (cm *CacheManager) Clear() {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	cm.data = make(map[string]CacheEntry)
	cm.save()
}

func (cm *CacheManager) GetOrSet(key string, fn func() interface{}, ttl time.Duration) interface{} {
	if val, exists := cm.Get(key); exists {
		return val
	}

	val := fn()
	cm.Set(key, val, ttl)
	return val
}

func (cm *CacheManager) Keys() []string {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	var keys []string
	for k, v := range cm.data {
		if time.Now().Before(v.ExpiresAt) {
			keys = append(keys, k)
		}
	}
	return keys
}

func (cm *CacheManager) Size() int {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return len(cm.data)
}

func (cm *CacheManager) Cleanup() int {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	count := 0
	now := time.Now()
	for k, v := range cm.data {
		if now.After(v.ExpiresAt) {
			delete(cm.data, k)
			count++
		}
	}

	cm.save()
	return count
}

func (cm *CacheManager) load() {
	data, err := os.ReadFile(cm.filePath)
	if err != nil {
		return
	}
	json.Unmarshal(data, &cm.data)
}

func (cm *CacheManager) save() {
	data, _ := json.Marshal(cm.data)
	os.WriteFile(cm.filePath, data, 0644)
}

func (cm *CacheManager) Stats() string {
	stats := fmt.Sprintf("Cache Stats\n")
	stats += fmt.Sprintf("Entries: %d\n", cm.Size())
	stats += fmt.Sprintf("Keys: %v\n", cm.Keys())
	return stats
}