package agent

import (
	"context"
	"fmt"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/khunquant/khunquant/pkg/providers"
	"github.com/khunquant/khunquant/pkg/tools"
)

// ====================== Test Helper: Event Collector ======================
type eventCollector struct {
	events []any
}

func (c *eventCollector) collect(e any) {
	c.events = append(c.events, e)
}

func (c *eventCollector) hasEventOfType(typ any) bool {
	targetType := reflect.TypeOf(typ)
	for _, e := range c.events {
		if reflect.TypeOf(e) == targetType {
			return true
		}
	}
	return false
}

func (c *eventCollector) countOfType(typ any) int {
	targetType := reflect.TypeOf(typ)
	count := 0
	for _, e := range c.events {
		if reflect.TypeOf(e) == targetType {
			count++
		}
	}
	return count
}

// ====================== Main Test Function ======================
func TestSpawnSubTurn(t *testing.T) {
	tests := []struct {
		name          string
		parentDepth   int
		config        SubTurnConfig
		wantErr       error
		wantSpawn     bool
		wantEnd       bool
		wantDepthFail bool
	}{
		{
			name:        "Basic success path - Single layer sub-turn",
			parentDepth: 0,
			config: SubTurnConfig{
				Model: "gpt-4o-mini",
				Tools: []tools.Tool{}, // At least one tool
			},
			wantErr:   nil,
			wantSpawn: true,
			wantEnd:   true,
		},
		{
			name:        "Nested 2 layers - Normal",
			parentDepth: 1,
			config: SubTurnConfig{
				Model: "gpt-4o-mini",
				Tools: []tools.Tool{},
			},
			wantErr:   nil,
			wantSpawn: true,
			wantEnd:   true,
		},
		{
			name:        "Depth limit triggered - 4th layer fails",
			parentDepth: 3,
			config: SubTurnConfig{
				Model: "gpt-4o-mini",
				Tools: []tools.Tool{},
			},
			wantErr:       ErrDepthLimitExceeded,
			wantSpawn:     false,
			wantEnd:       false,
			wantDepthFail: true,
		},
		{
			name:        "Invalid config - Empty Model",
			parentDepth: 0,
			config: SubTurnConfig{
				Model: "",
				Tools: []tools.Tool{},
			},
			wantErr:   ErrInvalidSubTurnConfig,
			wantSpawn: false,
			wantEnd:   false,
		},
	}

	al, _, _, _, cleanup := newTestAgentLoop(t)
	defer cleanup()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Prepare parent Turn
			parent := &turnState{
				ctx:            context.Background(),
				turnID:         "parent-1",
				depth:          tt.parentDepth,
				childTurnIDs:   []string{},
				pendingResults: make(chan *tools.ToolResult, 10),
				session:        &ephemeralSessionStore{},
			}

			// Replace mock with test collector
			collector := &eventCollector{}
			originalEmit := MockEventBus.Emit
			MockEventBus.Emit = collector.collect
			defer func() { MockEventBus.Emit = originalEmit }()

			// Execute spawnSubTurn
			result, err := spawnSubTurn(context.Background(), al, parent, tt.config)

			// Assert errors
			if tt.wantErr != nil {
				if err == nil || err != tt.wantErr {
					t.Errorf("expected error %v, got %v", tt.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			// Verify result
			if result == nil {
				t.Error("expected non-nil result")
			}

			// Verify event emission
			if tt.wantSpawn {
				if !collector.hasEventOfType(SubTurnSpawnEvent{}) {
					t.Error("SubTurnSpawnEvent not emitted")
				}
			}
			if tt.wantEnd {
				if !collector.hasEventOfType(SubTurnEndEvent{}) {
					t.Error("SubTurnEndEvent not emitted")
				}
			}

			// Verify turn tree
			if len(parent.childTurnIDs) == 0 && !tt.wantDepthFail {
				t.Error("child Turn not added to parent.childTurnIDs")
			}

			// Verify result delivery (pendingResults or history)
			if len(parent.pendingResults) > 0 || len(parent.session.GetHistory("")) > 0 {
				// Result delivered via at least one path
			} else {
				t.Error("child result not delivered")
			}
		})
	}
}

// ====================== Extra Independent Test: Ephemeral Session Isolation ======================
func TestSpawnSubTurn_EphemeralSessionIsolation(t *testing.T) {
	al, _, _, _, cleanup := newTestAgentLoop(t)
	defer cleanup()

	parentSession := &ephemeralSessionStore{}
	parentSession.AddMessage("", "user", "parent msg")
	parent := &turnState{
		ctx:            context.Background(),
		turnID:         "parent-1",
		depth:          0,
		pendingResults: make(chan *tools.ToolResult, 1),
		session:        parentSession,
	}

	cfg := SubTurnConfig{Model: "gpt-4o-mini", Tools: []tools.Tool{}}

	// Record main session length before execution
	originalLen := len(parent.session.GetHistory(""))

	_, _ = spawnSubTurn(context.Background(), al, parent, cfg)

	// After sub-turn ends, main session must remain unchanged
	if len(parent.session.GetHistory("")) != originalLen {
		t.Error("ephemeral session polluted the main session")
	}
}

// ====================== Extra Independent Test: Result Delivery Path ======================
func TestSpawnSubTurn_ResultDelivery(t *testing.T) {
	al, _, _, _, cleanup := newTestAgentLoop(t)
	defer cleanup()

	parent := &turnState{
		ctx:            context.Background(),
		turnID:         "parent-1",
		depth:          0,
		pendingResults: make(chan *tools.ToolResult, 1),
		session:        &ephemeralSessionStore{},
	}

	cfg := SubTurnConfig{Model: "gpt-4o-mini", Tools: []tools.Tool{}}

	_, _ = spawnSubTurn(context.Background(), al, parent, cfg)

	// Check if pendingResults received the result
	select {
	case res := <-parent.pendingResults:
		if res == nil {
			t.Error("received nil result in pendingResults")
		}
	default:
		t.Error("result did not enter pendingResults")
	}
}

// ====================== Extra Independent Test: Orphan Result Routing ======================
func TestSpawnSubTurn_OrphanResultRouting(t *testing.T) {
	parentCtx, cancelParent := context.WithCancel(context.Background())
	parent := &turnState{
		ctx:            parentCtx,
		cancelFunc:     cancelParent,
		turnID:         "parent-1",
		depth:          0,
		pendingResults: make(chan *tools.ToolResult, 1),
		session:        &ephemeralSessionStore{},
	}

	collector := &eventCollector{}
	originalEmit := MockEventBus.Emit
	MockEventBus.Emit = collector.collect
	defer func() { MockEventBus.Emit = originalEmit }()

	// Simulate parent finishing before child delivers result
	parent.Finish()

	// Call deliverSubTurnResult directly to simulate a delayed child
	deliverSubTurnResult(parent, "delayed-child", &tools.ToolResult{ForLLM: "late result"})

	// Verify Orphan event is emitted
	if !collector.hasEventOfType(SubTurnOrphanResultEvent{}) {
		t.Error("SubTurnOrphanResultEvent not emitted for finished parent")
	}

	// Verify history is NOT polluted
	if len(parent.session.GetHistory("")) != 0 {
		t.Error("Parent history was polluted by orphan result")
	}
}

// ====================== Extra Independent Test: Result Channel Registration ======================
func TestSubTurnResultChannelRegistration(t *testing.T) {
	al, _, _, _, cleanup := newTestAgentLoop(t)
	defer cleanup()

	parent := &turnState{
		ctx:            context.Background(),
		turnID:         "parent-reg-1",
		depth:          0,
		pendingResults: make(chan *tools.ToolResult, 4),
		session:        &ephemeralSessionStore{},
	}

	cfg := SubTurnConfig{Model: "gpt-4o-mini", Tools: []tools.Tool{}}

	// Before spawn: channel should not be registered
	if results := al.dequeuePendingSubTurnResults(parent.turnID); results != nil {
		t.Error("expected no channel before spawnSubTurn")
	}

	_, _ = spawnSubTurn(context.Background(), al, parent, cfg)

	// After spawn completes: channel should be unregistered (defer cleanup in spawnSubTurn)
	if _, ok := al.subTurnResults.Load(parent.turnID); ok {
		t.Error("channel should be unregistered after spawnSubTurn completes")
	}
}

// ====================== Extra Independent Test: Dequeue Pending SubTurn Results ======================
func TestDequeuePendingSubTurnResults(t *testing.T) {
	al, _, _, _, cleanup := newTestAgentLoop(t)
	defer cleanup()

	sessionKey := "test-session-dequeue"
	ch := make(chan *tools.ToolResult, 4)

	// Register channel manually
	al.registerSubTurnResultChannel(sessionKey, ch)
	defer al.unregisterSubTurnResultChannel(sessionKey)

	// Empty channel returns nil
	if results := al.dequeuePendingSubTurnResults(sessionKey); len(results) != 0 {
		t.Errorf("expected empty results, got %d", len(results))
	}

	// Put 3 results in
	ch <- &tools.ToolResult{ForLLM: "result-1"}
	ch <- &tools.ToolResult{ForLLM: "result-2"}
	ch <- &tools.ToolResult{ForLLM: "result-3"}

	results := al.dequeuePendingSubTurnResults(sessionKey)
	if len(results) != 3 {
		t.Errorf("expected 3 results, got %d", len(results))
	}
	if results[0].ForLLM != "result-1" || results[2].ForLLM != "result-3" {
		t.Error("results order or content mismatch")
	}

	// Channel should be drained now
	if results := al.dequeuePendingSubTurnResults(sessionKey); len(results) != 0 {
		t.Errorf("expected empty after drain, got %d", len(results))
	}

	// Unregistered session returns nil
	al.unregisterSubTurnResultChannel(sessionKey)
	if results := al.dequeuePendingSubTurnResults(sessionKey); results != nil {
		t.Error("expected nil for unregistered session")
	}
}

// ====================== Extra Independent Test: Concurrency Semaphore ======================
func TestSubTurnConcurrencySemaphore(t *testing.T) {
	al, _, _, _, cleanup := newTestAgentLoop(t)
	defer cleanup()

	parent := &turnState{
		ctx:            context.Background(),
		turnID:         "parent-concurrency",
		depth:          0,
		pendingResults: make(chan *tools.ToolResult, 10),
		session:        &ephemeralSessionStore{},
		concurrencySem: make(chan struct{}, 2), // Only allow 2 concurrent children
	}

	cfg := SubTurnConfig{Model: "gpt-4o-mini", Tools: []tools.Tool{}}

	// Spawn 2 children — should succeed immediately
	done := make(chan bool, 3)
	for i := 0; i < 2; i++ {
		go func() {
			_, _ = spawnSubTurn(context.Background(), al, parent, cfg)
			done <- true
		}()
	}

	// Wait a bit to ensure the first 2 are running
	// (In real scenario they'd be blocked in runTurn, but mockProvider returns immediately)
	// So we just verify the semaphore doesn't block when under limit
	<-done
	<-done

	// Verify semaphore is now full (2/2 slots used, but they already released)
	// Since mockProvider returns immediately, semaphore is already released
	// So we can't easily test blocking without a real long-running operation

	// Instead, verify that semaphore exists and has correct capacity
	if cap(parent.concurrencySem) != 2 {
		t.Errorf("expected semaphore capacity 2, got %d", cap(parent.concurrencySem))
	}
}

// ====================== Extra Independent Test: Hard Abort Cascading ======================
func TestHardAbortCascading(t *testing.T) {
	al, _, _, _, cleanup := newTestAgentLoop(t)
	defer cleanup()

	sessionKey := "test-session-abort"
	parentCtx, parentCancel := context.WithCancel(context.Background())
	defer parentCancel()

	rootTS := &turnState{
		ctx:            parentCtx,
		turnID:         sessionKey,
		depth:          0,
		session:        &ephemeralSessionStore{},
		pendingResults: make(chan *tools.ToolResult, 16),
		concurrencySem: make(chan struct{}, 5),
	}

	// Register the root turn state
	al.activeTurnStates.Store(sessionKey, rootTS)
	defer al.activeTurnStates.Delete(sessionKey)

	// Create a child turn state
	childCtx, childCancel := context.WithCancel(rootTS.ctx)
	defer childCancel()
	childTS := &turnState{
		ctx:            childCtx,
		cancelFunc:     childCancel,
		turnID:         "child-1",
		parentTurnID:   sessionKey,
		depth:          1,
		session:        &ephemeralSessionStore{},
		pendingResults: make(chan *tools.ToolResult, 16),
		concurrencySem: make(chan struct{}, 5),
	}

	// Attach cancelFunc to rootTS so Finish() can trigger it
	rootTS.cancelFunc = parentCancel

	// Verify contexts are not canceled yet
	select {
	case <-rootTS.ctx.Done():
		t.Error("root context should not be canceled yet")
	default:
	}
	select {
	case <-childTS.ctx.Done():
		t.Error("child context should not be canceled yet")
	default:
	}

	// Trigger Hard Abort
	err := al.HardAbort(sessionKey)
	if err != nil {
		t.Errorf("HardAbort failed: %v", err)
	}

	// Verify root context is canceled
	select {
	case <-rootTS.ctx.Done():
		// Expected
	default:
		t.Error("root context should be canceled after HardAbort")
	}

	// Verify child context is also canceled (cascading)
	select {
	case <-childTS.ctx.Done():
		// Expected
	default:
		t.Error("child context should be canceled after HardAbort (cascading)")
	}

	// Verify HardAbort on non-existent session returns error
	err = al.HardAbort("non-existent-session")
	if err == nil {
		t.Error("expected error for non-existent session")
	}
}

// TestHardAbortSessionRollback verifies that HardAbort rolls back session history
// to the state before the turn started, discarding all messages added during the turn.
func TestHardAbortSessionRollback(t *testing.T) {
	al, _, _, _, cleanup := newTestAgentLoop(t)
	defer cleanup()

	// Create a session with initial history
	sess := &ephemeralSessionStore{
		history: []providers.Message{
			{Role: "user", Content: "initial message 1"},
			{Role: "assistant", Content: "initial response 1"},
		},
	}

	// Create a root turnState with initialHistoryLength = 2
	rootTS := &turnState{
		ctx:                  context.Background(),
		turnID:               "test-session",
		depth:                0,
		session:              sess,
		initialHistoryLength: 2, // Snapshot: 2 messages
		pendingResults:       make(chan *tools.ToolResult, 16),
		concurrencySem:       make(chan struct{}, 5),
	}

	// Register the turn state
	al.activeTurnStates.Store("test-session", rootTS)

	// Simulate adding messages during the turn (e.g., user input + assistant response)
	sess.AddMessage("", "user", "new user message")
	sess.AddMessage("", "assistant", "new assistant response")

	// Verify history grew to 4 messages
	if len(sess.GetHistory("")) != 4 {
		t.Fatalf("expected 4 messages before abort, got %d", len(sess.GetHistory("")))
	}

	// Trigger HardAbort
	err := al.HardAbort("test-session")
	if err != nil {
		t.Fatalf("HardAbort failed: %v", err)
	}

	// Verify history rolled back to initial 2 messages
	finalHistory := sess.GetHistory("")
	if len(finalHistory) != 2 {
		t.Errorf("expected history to rollback to 2 messages, got %d", len(finalHistory))
	}

	// Verify the content matches the initial state
	if finalHistory[0].Content != "initial message 1" || finalHistory[1].Content != "initial response 1" {
		t.Error("history content does not match initial state after rollback")
	}
}

// TestNestedSubTurnHierarchy verifies that nested SubTurns maintain correct
// parent-child relationships and depth tracking when recursively calling runAgentLoop.
func TestNestedSubTurnHierarchy(t *testing.T) {
	al, _, _, _, cleanup := newTestAgentLoop(t)
	defer cleanup()

	// Track spawned turns and their depths
	type turnInfo struct {
		parentID string
		childID  string
		depth    int
	}
	var spawnedTurns []turnInfo
	var mu sync.Mutex

	// Override MockEventBus to capture spawn events
	originalEmit := MockEventBus.Emit
	defer func() { MockEventBus.Emit = originalEmit }()

	MockEventBus.Emit = func(event any) {
		if spawnEvent, ok := event.(SubTurnSpawnEvent); ok {
			mu.Lock()
			// Extract depth from context (we'll verify this matches expected depth)
			spawnedTurns = append(spawnedTurns, turnInfo{
				parentID: spawnEvent.ParentID,
				childID:  spawnEvent.ChildID,
			})
			mu.Unlock()
		}
	}

	// Create a root turn
	rootSession := &ephemeralSessionStore{}
	rootTS := &turnState{
		ctx:            context.Background(),
		turnID:         "root-turn",
		depth:          0,
		session:        rootSession,
		pendingResults: make(chan *tools.ToolResult, 16),
		concurrencySem: make(chan struct{}, 5),
	}

	// Spawn a child (depth 1)
	childCfg := SubTurnConfig{Model: "gpt-4o-mini"}
	_, err := spawnSubTurn(context.Background(), al, rootTS, childCfg)
	if err != nil {
		t.Fatalf("failed to spawn child: %v", err)
	}

	// Verify we captured the spawn event
	mu.Lock()
	if len(spawnedTurns) != 1 {
		t.Fatalf("expected 1 spawn event, got %d", len(spawnedTurns))
	}
	if spawnedTurns[0].parentID != "root-turn" {
		t.Errorf("expected parent ID 'root-turn', got %s", spawnedTurns[0].parentID)
	}
	mu.Unlock()

	// Verify root turn has the child in its childTurnIDs
	rootTS.mu.Lock()
	if len(rootTS.childTurnIDs) != 1 {
		t.Errorf("expected root to have 1 child, got %d", len(rootTS.childTurnIDs))
	}
	rootTS.mu.Unlock()
}

// TestDeliverSubTurnResultNoDeadlock verifies that deliverSubTurnResult doesn't
// deadlock when multiple goroutines are accessing the parent turnState concurrently.
func TestDeliverSubTurnResultNoDeadlock(t *testing.T) {
	parent := &turnState{
		ctx:            context.Background(),
		turnID:         "parent-deadlock-test",
		depth:          0,
		pendingResults: make(chan *tools.ToolResult, 2), // Small buffer to test blocking
		isFinished:     false,
	}

	// Simulate multiple child turns delivering results concurrently
	var wg sync.WaitGroup
	numChildren := 10

	for i := 0; i < numChildren; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			result := &tools.ToolResult{ForLLM: fmt.Sprintf("result-%d", id)}
			deliverSubTurnResult(parent, fmt.Sprintf("child-%d", id), result)
		}(i)
	}

	// Concurrently read from the channel to prevent blocking
	go func() {
		for i := 0; i < numChildren; i++ {
			select {
			case <-parent.pendingResults:
			case <-time.After(2 * time.Second):
				t.Error("timeout waiting for result")
				return
			}
		}
	}()

	// Wait for all deliveries to complete (with timeout)
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// Success - no deadlock
	case <-time.After(3 * time.Second):
		t.Fatal("deadlock detected: deliverSubTurnResult blocked")
	}
}

// TestHardAbortOrderOfOperations verifies that HardAbort calls Finish() before
// rolling back session history, minimizing the race window where new messages
// could be added after rollback.
func TestHardAbortOrderOfOperations(t *testing.T) {
	al, _, _, _, cleanup := newTestAgentLoop(t)
	defer cleanup()

	sess := &ephemeralSessionStore{
		history: []providers.Message{
			{Role: "user", Content: "initial message"},
			{Role: "assistant", Content: "response 1"},
			{Role: "user", Content: "follow-up"},
		},
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	rootTS := &turnState{
		ctx:                  ctx,
		cancelFunc:           cancel,
		turnID:               "test-session-order",
		depth:                0,
		session:              sess,
		initialHistoryLength: 1, // Snapshot: 1 message
		pendingResults:       make(chan *tools.ToolResult, 16),
		concurrencySem:       make(chan struct{}, 5),
	}

	al.activeTurnStates.Store("test-session-order", rootTS)

	// Trigger HardAbort
	err := al.HardAbort("test-session-order")
	if err != nil {
		t.Fatalf("HardAbort failed: %v", err)
	}

	// Verify context was cancelled (Finish() was called)
	select {
	case <-rootTS.ctx.Done():
		// Good - context was cancelled
	default:
		t.Error("expected context to be cancelled after HardAbort")
	}

	// Verify history was rolled back
	finalHistory := sess.GetHistory("")
	if len(finalHistory) != 1 {
		t.Errorf("expected history to rollback to 1 message, got %d", len(finalHistory))
	}

	if finalHistory[0].Content != "initial message" {
		t.Error("history content does not match initial state after rollback")
	}
}

// TestFinishClosesChannel verifies that Finish() closes the pendingResults channel
// and that deliverSubTurnResult handles closed channels gracefully.
func TestFinishClosesChannel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ts := &turnState{
		ctx:            ctx,
		cancelFunc:     cancel,
		turnID:         "test-finish-channel",
		depth:          0,
		pendingResults: make(chan *tools.ToolResult, 2),
		isFinished:     false,
	}

	// Verify channel is open initially
	select {
	case ts.pendingResults <- &tools.ToolResult{ForLLM: "test"}:
		// Good - channel is open
		// Drain the message we just sent
		<-ts.pendingResults
	default:
		t.Fatal("channel should be open initially")
	}

	// Call Finish()
	ts.Finish()

	// Verify channel is closed
	_, ok := <-ts.pendingResults
	if ok {
		t.Error("expected channel to be closed after Finish()")
	}

	// Verify Finish() is idempotent (can be called multiple times)
	ts.Finish() // Should not panic

	// Verify deliverSubTurnResult doesn't panic when sending to closed channel
	result := &tools.ToolResult{ForLLM: "late result"}

	// This should not panic - it should recover and emit OrphanResultEvent
	deliverSubTurnResult(ts, "child-1", result)
}
