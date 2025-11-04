package collector

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestPeriodicTrigger_Start tests that the trigger function is called immediately on start
func TestPeriodicTrigger_StartImmediateExecution(t *testing.T) {
	callCount := atomic.Int32{}
	triggerFn := func(ctx context.Context) error {
		callCount.Add(1)
		return nil
	}

	pt := NewPeriodicTrigger("test-trigger", triggerFn, 1*time.Second)

	// Start the trigger (it should execute immediately)
	go pt.Start()

	// Give it a moment to execute the initial call
	time.Sleep(100 * time.Millisecond)

	// Stop the trigger
	pt.Stop()

	// Should have been called at least once (the immediate execution)
	assert.GreaterOrEqual(t, callCount.Load(), int32(1))
}

// TestPeriodicTrigger_PeriodicExecution tests that the trigger function is called periodically
func TestPeriodicTrigger_PeriodicExecution(t *testing.T) {
	callCount := atomic.Int32{}
	var mu sync.Mutex
	callTimes := []time.Time{}

	triggerFn := func(ctx context.Context) error {
		mu.Lock()
		defer mu.Unlock()
		callCount.Add(1)
		callTimes = append(callTimes, time.Now())
		return nil
	}

	interval := 100 * time.Millisecond
	pt := NewPeriodicTrigger("test-trigger", triggerFn, interval)

	// Start the trigger
	go pt.Start()

	// Let it run for enough time to get multiple calls
	time.Sleep(350 * time.Millisecond)

	// Stop the trigger
	pt.Stop()

	// Should have been called multiple times (initial + periodic)
	count := callCount.Load()
	assert.GreaterOrEqual(t, count, int32(3), "Expected at least 3 calls (initial + 2 periodic)")

	// Verify the timing between calls is approximately the interval
	mu.Lock()
	defer mu.Unlock()
	if len(callTimes) >= 2 {
		for i := 1; i < len(callTimes); i++ {
			diff := callTimes[i].Sub(callTimes[i-1])
			// Allow some tolerance (±50ms)
			assert.InDelta(t, interval.Milliseconds(), diff.Milliseconds(), 50.0,
				"Time between calls should be approximately %v, got %v", interval, diff)
		}
	}
}

// TestPeriodicTrigger_TriggerFunctionError tests that errors from trigger function don't stop the periodic execution
func TestPeriodicTrigger_TriggerFunctionError(t *testing.T) {
	callCount := atomic.Int32{}

	triggerFn := func(ctx context.Context) error {
		count := callCount.Add(1)
		if count == 2 {
			// Return error on second call
			return errors.New("test error")
		}
		return nil
	}

	interval := 100 * time.Millisecond
	pt := NewPeriodicTrigger("test-trigger", triggerFn, interval)

	// Start the trigger
	go pt.Start()

	// Let it run for enough time to get multiple calls
	time.Sleep(350 * time.Millisecond)

	// Stop the trigger
	pt.Stop()

	// Should have been called multiple times despite the error
	assert.GreaterOrEqual(t, callCount.Load(), int32(3), "Expected at least 3 calls despite error")
}

// TestPeriodicTrigger_StartOnce tests that Start can only be called once
func TestPeriodicTrigger_StartOnce(t *testing.T) {
	callCount := atomic.Int32{}

	triggerFn := func(ctx context.Context) error {
		callCount.Add(1)
		time.Sleep(50 * time.Millisecond) // Make the function take some time
		return nil
	}

	interval := 200 * time.Millisecond
	pt := NewPeriodicTrigger("test-trigger", triggerFn, interval)

	// Start the trigger multiple times
	go pt.Start()
	go pt.Start()
	go pt.Start()

	// Let it run for a bit
	time.Sleep(250 * time.Millisecond)

	// Stop the trigger
	pt.Stop()

	// Even though Start was called 3 times, it should only actually start once
	// So we should have the initial call + maybe 1 periodic call
	count := callCount.Load()
	assert.LessOrEqual(t, count, int32(3), "Start should only execute once despite multiple calls")
}

// TestPeriodicTrigger_StopOnce tests that Stop can be called multiple times safely
func TestPeriodicTrigger_StopOnce(t *testing.T) {
	triggerFn := func(ctx context.Context) error {
		return nil
	}

	pt := NewPeriodicTrigger("test-trigger", triggerFn, 100*time.Millisecond)

	// Start the trigger
	go pt.Start()

	// Give it a moment to start
	time.Sleep(50 * time.Millisecond)

	// Stop multiple times - should not panic
	pt.Stop()
	pt.Stop()
	pt.Stop()

	// Test passes if no panic occurs
}

// TestPeriodicTrigger_ContextPassed tests that a valid context is passed to the trigger function
func TestPeriodicTrigger_ContextPassed(t *testing.T) {
	var capturedCtx context.Context
	var mu sync.Mutex

	triggerFn := func(ctx context.Context) error {
		mu.Lock()
		capturedCtx = ctx
		mu.Unlock()
		return nil
	}

	pt := NewPeriodicTrigger("test-trigger", triggerFn, 100*time.Millisecond)

	// Start the trigger
	go pt.Start()

	// Give it a moment to execute
	time.Sleep(50 * time.Millisecond)

	mu.Lock()
	ctx := capturedCtx
	mu.Unlock()

	require.NotNil(t, ctx, "Context should have been captured")

	// Context should be valid (not cancelled) during execution
	select {
	case <-ctx.Done():
		t.Fatal("Context should not be cancelled during execution")
	default:
		// Good, context is valid
	}

	// Stop the trigger
	pt.Stop()
}

// TestPeriodicTrigger_Reset tests that reset properly reinitializes the trigger
func TestPeriodicTrigger_Reset(t *testing.T) {
	triggerFn := func(ctx context.Context) error {
		return nil
	}

	pt := NewPeriodicTrigger("test-trigger", triggerFn, 100*time.Millisecond)

	// Verify initial state - all fields should be non-nil
	assert.NotNil(t, pt.stopCh, "stopCh should not be nil after creation")
	assert.NotNil(t, pt.startOnce, "startOnce should not be nil after creation")
	assert.NotNil(t, pt.stopOnce, "stopOnce should not be nil after creation")
	assert.NotNil(t, pt.triggerCtx, "triggerCtx should not be nil after creation")
	assert.NotNil(t, pt.triggerCtxCancelFunc, "triggerCtxCancelFunc should not be nil after creation")

	// Call reset
	pt.reset()

	// Verify that after reset, all fields are still non-nil (re-initialized)
	assert.NotNil(t, pt.stopCh, "stopCh should not be nil after reset")
	assert.NotNil(t, pt.startOnce, "startOnce should not be nil after reset")
	assert.NotNil(t, pt.stopOnce, "stopOnce should not be nil after reset")
	assert.NotNil(t, pt.triggerCtx, "triggerCtx should not be nil after reset")
	assert.NotNil(t, pt.triggerCtxCancelFunc, "triggerCtxCancelFunc should not be nil after reset")
}

// TestPeriodicTrigger_StartStopStartAgain tests that a trigger can be restarted after stopping
func TestPeriodicTrigger_StartStopStartAgain(t *testing.T) {
	callCount := atomic.Int32{}

	triggerFn := func(ctx context.Context) error {
		callCount.Add(1)
		return nil
	}

	interval := 100 * time.Millisecond
	pt := NewPeriodicTrigger("test-trigger", triggerFn, interval)

	// First start
	go pt.Start()
	time.Sleep(150 * time.Millisecond)
	pt.Stop()

	firstCount := callCount.Load()
	assert.GreaterOrEqual(t, firstCount, int32(1))

	// Wait a bit to ensure no more calls happen after stop
	time.Sleep(150 * time.Millisecond)
	countAfterStop := callCount.Load()
	assert.Equal(t, firstCount, countAfterStop, "No new calls should happen after stop")

	// Second start
	go pt.Start()
	time.Sleep(150 * time.Millisecond)
	pt.Stop()

	finalCount := callCount.Load()
	assert.Greater(t, finalCount, firstCount, "Should have more calls after restarting")
}

// TestPeriodicTrigger_ShortInterval tests with a very short interval
func TestPeriodicTrigger_ShortInterval(t *testing.T) {
	callCount := atomic.Int32{}

	triggerFn := func(ctx context.Context) error {
		callCount.Add(1)
		return nil
	}

	interval := 10 * time.Millisecond
	pt := NewPeriodicTrigger("test-trigger", triggerFn, interval)

	go pt.Start()
	time.Sleep(100 * time.Millisecond)
	pt.Stop()

	// With 10ms interval over 100ms, we should get at least 5 calls
	assert.GreaterOrEqual(t, callCount.Load(), int32(5))
}

// TestPeriodicTrigger_LongRunningTriggerFunction tests behavior when trigger function takes longer than interval
func TestPeriodicTrigger_LongRunningTriggerFunction(t *testing.T) {
	callCount := atomic.Int32{}

	triggerFn := func(ctx context.Context) error {
		callCount.Add(1)
		time.Sleep(150 * time.Millisecond) // Longer than interval
		return nil
	}

	interval := 50 * time.Millisecond
	pt := NewPeriodicTrigger("test-trigger", triggerFn, interval)

	go pt.Start()
	time.Sleep(400 * time.Millisecond)
	pt.Stop()

	// Should still be called multiple times, but calls won't overlap
	count := callCount.Load()
	assert.GreaterOrEqual(t, count, int32(2))
}
