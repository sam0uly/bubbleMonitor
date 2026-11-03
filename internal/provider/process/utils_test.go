package process

import (
	"fmt"
	"sync"
	"testing"

	"samouly.fun/bubbleMonitor/internal/data"
)

func TestGetProcSliceTypeSafety(t *testing.T) {
	slice := GetProcSlice()
	if slice == nil || *slice == nil {
		t.Fatal("GetProcSlice returned nil or nil slice")
	}

	if cap(*slice) < 500 {
		t.Errorf("Expected capacity >= 500, got %d", cap(*slice))
	}
}

func TestInternerBasic(t *testing.T) {
	interner := &Interner{
		cache: make(map[string]*internerEntry),
	}

	s1 := "test_process"
	s2 := "test_process"

	interned1 := interner.Intern(s1)
	interned2 := interner.Intern(s2)

	if interned1 != interned2 {
		t.Error("Same strings should return same interned value")
	}

	if &interned1 != &interned2 {
		t.Log("Warning: Interned strings don't share same address")
	}
}

func TestInternerPruning(t *testing.T) {
	interner := &Interner{
		cache: make(map[string]*internerEntry),
	}

	for i := 0; i < maxInternerSize+2000; i++ {
		interner.Intern(fmt.Sprintf("process-%d", i))
	}

	if len(interner.cache) > maxInternerSize*2 {
		t.Errorf("Interner not pruning: %d entries exceeds reasonable limit", len(interner.cache))
	}
}

func TestInternerConcurrency(t *testing.T) {
	interner := &Interner{
		cache: make(map[string]*internerEntry),
	}

	var wg sync.WaitGroup
	numGoroutines := 10
	iterations := 1000

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				str := fmt.Sprintf("process-%d", j%100)
				_ = interner.Intern(str)
			}
		}(i)
	}

	wg.Wait()

	if len(interner.cache) == 0 {
		t.Error("Interner cache should not be empty after concurrent operations")
	}
}

func TestSlicePoolReuseability(t *testing.T) {
	slice1 := GetProcSlice()
	*slice1 = append(*slice1, data.ProcessInfo{Pid: 123, Name: "test"})

	PutProcSlice(slice1)

	slice2 := GetProcSlice()
	if len(*slice2) != 0 {
		t.Errorf("Expected reset slice to have length 0, got %d", len(*slice2))
	}
}

func BenchmarkIntern(b *testing.B) {
	interner := &Interner{
		cache: make(map[string]*internerEntry),
	}

	strings := make([]string, 100)
	for i := 0; i < 100; i++ {
		strings[i] = fmt.Sprintf("process-%d", i)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = interner.Intern(strings[i%100])
	}
}
