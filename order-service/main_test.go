
package main

import (
	"testing"
	"time"
)

func TestCircuitBreaker_InitiallyAllowsRequests(t *testing.T) {
	breaker := &circuitBreaker{}

	if !breaker.allow() {
		t.Fatal("expected circuit breaker to allow requests initially")
	}
}

func TestCircuitBreaker_OpensAfterThreeFailures(t *testing.T) {
	breaker := &circuitBreaker{}

	breaker.recordFailure()
	breaker.recordFailure()

	if !breaker.allow() {
		t.Fatal("expected requests to be allowed before three failures")
	}

	breaker.recordFailure()

	if breaker.allow() {
		t.Fatal("expected circuit breaker to block requests after three failures")
	}
}

func TestCircuitBreaker_RecordSuccessResetsFailures(t *testing.T) {
	breaker := &circuitBreaker{}

	breaker.recordFailure()
	breaker.recordFailure()
	breaker.recordSuccess()

	if !breaker.allow() {
		t.Fatal("expected requests to be allowed after success")
	}

	if breaker.failures != 0 {
		t.Errorf("expected failure count 0, got %d", breaker.failures)
	}
}

func TestCircuitBreaker_RecoversAfterOpenDuration(t *testing.T) {
	breaker := &circuitBreaker{}

	breaker.recordFailure()
	breaker.recordFailure()
	breaker.recordFailure()

	// Simulate the circuit having been open for longer than five seconds.
	breaker.mu.Lock()
	breaker.openedAt = time.Now().Add(-openDuration - time.Second)
	breaker.mu.Unlock()

	if !breaker.allow() {
		t.Fatal("expected circuit breaker to allow a trial request after recovery time")
	}

	if breaker.open {
		t.Error("expected circuit breaker to leave the open state")
	}
}
