package mango

import (
	"testing"
)

func TestComputed(t *testing.T) {
	a := NewReactive(2)
	b := NewComputed(func() int {
		return a.Get() * 3
	})

	if b.Get() != 6 {
		t.Errorf("expected 6, got %v", b.Get())
	}

	a.Set(4)
	if b.Get() != 12 {
		t.Errorf("expected 12, got %v", b.Get())
	}
}

func TestWatchEffect(t *testing.T) {
	a := NewReactive(1)
	b := NewReactive(2)

	var results []int

	stop := WatchEffect(func() {
		sum := a.Get() + b.Get()
		results = append(results, sum)
	})

	defer stop()

	a.Set(3)
	b.Set(5)

	expected := []int{3, 5, 8}
	if len(results) != len(expected) {
		t.Fatalf("expected %d results, got %d", len(expected), len(results))
	}
	for i := range expected {
		if results[i] != expected[i] {
			t.Errorf("at index %d: expected %d, got %d", i, expected[i], results[i])
		}
	}
}

func TestWatchEffectDerived(t *testing.T) {
	a := NewReactive(1)

	b := NewComputed(func() int {
		return a.Get() + 10
	})

	var results []int

	stop := WatchEffect(func() {
		results = append(results, b.Get())
	})

	defer stop()

	a.Set(3)

	expected := []int{11, 13}
	if len(results) != len(expected) {
		t.Fatalf("expected %d results, got %d", len(expected), len(results))
	}
	for i := range expected {
		if results[i] != expected[i] {
			t.Errorf("at index %d: expected %d, got %d", i, expected[i], results[i])
		}
	}
}

func TestComputedWithWatchEffect(t *testing.T) {
	a := NewReactive(2)
	b := NewReactive(3)

	sum := NewComputed(func() int {
		return a.Get() + b.Get()
	})

	var results []int

	stop := WatchEffect(func() {
		val := sum.Get()
		results = append(results, val)
	})

	defer stop()

	if len(results) != 1 || results[0] != 5 {
		t.Errorf("expected initial sum 5, got %v", results)
	}

	a.Set(4)
	b.Set(1)

	expected := []int{5, 7, 5}
	if len(results) != len(expected) {
		t.Fatalf("expected %d results, got %d", len(expected), len(results))
	}
	for i := range expected {
		if results[i] != expected[i] {
			t.Errorf("at index %d: expected %d, got %d", i, expected[i], results[i])
		}
	}
}

func TestStopWatchEffect(t *testing.T) {
	a := NewReactive(10)

	var called int
	stop := WatchEffect(func() {
		called++
	})

	a.Set(20)
	if called != 1 {
		t.Errorf("expected 1 call before stop, got %d", called)
	}

	stop()

	a.Set(30)
	if called != 1 {
		t.Errorf("effect should not be called after stop, got %d", called)
	}
}

func TestComputedIsLazy(t *testing.T) {
	a := NewReactive(1)
	var computedCalls int

	comp := NewComputed(func() int {
		computedCalls++
		return a.Get() * 2
	})

	if computedCalls != 0 {
		t.Errorf("computed should be lazy and not called yet")
	}

	val := comp.Get()
	if val != 2 || computedCalls != 1 {
		t.Errorf("computed should compute on first Get, got %v, calls %d", val, computedCalls)
	}
}
