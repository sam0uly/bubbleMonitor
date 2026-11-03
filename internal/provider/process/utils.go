package process

import (
	"sort"
	"sync"

	"samouly.fun/bubbleMonitor/internal/data"
)

var procSlicePool = sync.Pool{
	New: func() any {
		s := make([]data.ProcessInfo, 0, ProcessListCapacity)
		return &s
	},
}

func GetProcSlice() *[]data.ProcessInfo {
	slice, ok := procSlicePool.Get().(*[]data.ProcessInfo)
	if !ok {
		s := make([]data.ProcessInfo, 0, ProcessListCapacity)
		return &s
	}
	return slice
}

func PutProcSlice(s *[]data.ProcessInfo) {
	if s == nil || *s == nil {
		return
	}
	*s = (*s)[:0]
	procSlicePool.Put(s)
}

type Interner struct {
	cache      map[string]*internerEntry
	mu         sync.Mutex // Mutex (not RWMutex): every operation is a write.
	count      uint64
	useCounter uint64 // Incremented on each access for LRU tracking
}

const (
	maxInternerSize    = 5000
	internerEvictBatch = 500 // Evict 500 oldest entries at a time
)

type internerEntry struct {
	value string
	used  uint64 // Last use counter for LRU
}

var globalInterner = &Interner{
	cache: make(map[string]*internerEntry),
}

func Intern(s string) string {
	return globalInterner.Intern(s)
}

func (i *Interner) Intern(s string) string {
	i.mu.Lock()
	if entry, ok := i.cache[s]; ok {
		i.useCounter++
		entry.used = i.useCounter
		v := entry.value
		i.mu.Unlock()
		return v
	}

	i.useCounter++
	i.cache[s] = &internerEntry{value: s, used: i.useCounter}
	i.count++

	over := len(i.cache) > maxInternerSize
	i.mu.Unlock()

	// Eviction runs without holding the lock so a 500-entry sort
	// doesn't stall every other caller. Snapshot the entries first.
	if over {
		i.evictOldest(internerEvictBatch)
	}
	return s
}

// evictOldest removes n oldest entries from the cache.
func (i *Interner) evictOldest(n int) {
	// Step 1: snapshot under lock.
	i.mu.Lock()
	if len(i.cache) <= n {
		i.mu.Unlock()
		return
	}
	type item struct {
		key  string
		used uint64
	}
	items := make([]item, 0, len(i.cache))
	for k, v := range i.cache {
		items = append(items, item{key: k, used: v.used})
	}
	i.mu.Unlock()

	// Step 2: sort outside the lock.
	if len(items) <= n {
		return
	}
	sort.Slice(items, func(a, b int) bool { return items[a].used < items[b].used })

	// Step 3: re-acquire briefly to delete the n oldest.
	i.mu.Lock()
	for j := 0; j < n && j < len(items); j++ {
		delete(i.cache, items[j].key)
	}
	i.mu.Unlock()
}
