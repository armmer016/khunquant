package cron

import (
	"testing"
	"time"

	"github.com/adhocore/gronx"
)

// newTestSvc returns a minimal CronService suitable for pure-logic tests.
// It has no storePath and no onJob — only the in-memory store is initialised.
func newTestSvc(jobs ...CronJob) *CronService {
	return &CronService{
		store: &CronStore{Jobs: jobs},
		gronx: gronx.New(),
	}
}

// --- computeNextRun ---

func TestComputeNextRun_At_FutureReturnsAtMS(t *testing.T) {
	svc := newTestSvc()
	future := int64(9_999_999_999_000)
	sched := &CronSchedule{Kind: "at", AtMS: &future}
	got := svc.computeNextRun(sched, 0)
	if got == nil || *got != future {
		t.Errorf("want %d, got %v", future, got)
	}
}

func TestComputeNextRun_At_PastReturnsNil(t *testing.T) {
	svc := newTestSvc()
	past := int64(1000)
	sched := &CronSchedule{Kind: "at", AtMS: &past}
	got := svc.computeNextRun(sched, 2000)
	if got != nil {
		t.Errorf("expected nil for past AtMS, got %d", *got)
	}
}

func TestComputeNextRun_At_NilAtMSReturnsNil(t *testing.T) {
	svc := newTestSvc()
	sched := &CronSchedule{Kind: "at", AtMS: nil}
	if got := svc.computeNextRun(sched, 0); got != nil {
		t.Errorf("expected nil, got %d", *got)
	}
}

func TestComputeNextRun_Every_ReturnsNowPlusDuration(t *testing.T) {
	svc := newTestSvc()
	everyMS := int64(60_000) // 1 minute
	sched := &CronSchedule{Kind: "every", EveryMS: &everyMS}
	nowMS := int64(1_000_000)
	got := svc.computeNextRun(sched, nowMS)
	if got == nil {
		t.Fatal("expected non-nil next run")
	}
	if *got != nowMS+everyMS {
		t.Errorf("want %d, got %d", nowMS+everyMS, *got)
	}
}

func TestComputeNextRun_Every_NilEveryMSReturnsNil(t *testing.T) {
	svc := newTestSvc()
	sched := &CronSchedule{Kind: "every", EveryMS: nil}
	if got := svc.computeNextRun(sched, 0); got != nil {
		t.Errorf("expected nil for nil EveryMS, got %d", *got)
	}
}

func TestComputeNextRun_Every_ZeroEveryMSReturnsNil(t *testing.T) {
	svc := newTestSvc()
	zero := int64(0)
	sched := &CronSchedule{Kind: "every", EveryMS: &zero}
	if got := svc.computeNextRun(sched, 0); got != nil {
		t.Errorf("expected nil for zero EveryMS, got %d", *got)
	}
}

func TestComputeNextRun_Cron_ValidExprReturnsNextTick(t *testing.T) {
	svc := newTestSvc()
	// Run at midnight every day.
	sched := &CronSchedule{Kind: "cron", Expr: "0 0 * * *"}
	nowMS := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC).UnixMilli()
	got := svc.computeNextRun(sched, nowMS)
	if got == nil {
		t.Fatal("expected non-nil next run for valid cron expr")
	}
	next := time.UnixMilli(*got)
	// Next should be after now.
	if !next.After(time.UnixMilli(nowMS)) {
		t.Errorf("expected next run after now, got %v", next)
	}
}

func TestComputeNextRun_Cron_EmptyExprReturnsNil(t *testing.T) {
	svc := newTestSvc()
	sched := &CronSchedule{Kind: "cron", Expr: ""}
	if got := svc.computeNextRun(sched, 0); got != nil {
		t.Errorf("expected nil for empty cron expr, got %d", *got)
	}
}

func TestComputeNextRun_UnknownKindReturnsNil(t *testing.T) {
	svc := newTestSvc()
	sched := &CronSchedule{Kind: "unknown"}
	if got := svc.computeNextRun(sched, 0); got != nil {
		t.Errorf("expected nil for unknown schedule kind, got %d", *got)
	}
}

// --- getNextWakeMS ---

func TestGetNextWakeMS_EmptyJobsReturnsNil(t *testing.T) {
	svc := newTestSvc()
	if got := svc.getNextWakeMS(); got != nil {
		t.Errorf("expected nil for empty jobs, got %d", *got)
	}
}

func TestGetNextWakeMS_SingleEnabledJob(t *testing.T) {
	ms := int64(5000)
	svc := newTestSvc(CronJob{
		Enabled: true,
		State:   CronJobState{NextRunAtMS: &ms},
	})
	got := svc.getNextWakeMS()
	if got == nil || *got != ms {
		t.Errorf("want %d, got %v", ms, got)
	}
}

func TestGetNextWakeMS_ReturnsMinOfEnabled(t *testing.T) {
	a, b := int64(3000), int64(1000)
	svc := newTestSvc(
		CronJob{Enabled: true, State: CronJobState{NextRunAtMS: &a}},
		CronJob{Enabled: true, State: CronJobState{NextRunAtMS: &b}},
	)
	got := svc.getNextWakeMS()
	if got == nil || *got != b {
		t.Errorf("want %d (min), got %v", b, got)
	}
}

func TestGetNextWakeMS_DisabledJobsIgnored(t *testing.T) {
	ms := int64(9000)
	svc := newTestSvc(
		CronJob{Enabled: false, State: CronJobState{NextRunAtMS: &ms}},
	)
	if got := svc.getNextWakeMS(); got != nil {
		t.Errorf("expected nil for all-disabled jobs, got %d", *got)
	}
}

func TestGetNextWakeMS_NilNextRunIgnored(t *testing.T) {
	valid := int64(7000)
	svc := newTestSvc(
		CronJob{Enabled: true, State: CronJobState{NextRunAtMS: nil}},
		CronJob{Enabled: true, State: CronJobState{NextRunAtMS: &valid}},
	)
	got := svc.getNextWakeMS()
	if got == nil || *got != valid {
		t.Errorf("want %d, got %v", valid, got)
	}
}

// --- generateID ---

func TestGenerateID_NonEmpty(t *testing.T) {
	id := generateID()
	if id == "" {
		t.Error("generateID returned empty string")
	}
}

func TestGenerateID_Unique(t *testing.T) {
	ids := make(map[string]bool)
	for i := 0; i < 100; i++ {
		id := generateID()
		if ids[id] {
			t.Errorf("duplicate ID generated: %q", id)
		}
		ids[id] = true
	}
}

// --- ListJobs ---

func TestListJobs_IncludeDisabled_ReturnsAll(t *testing.T) {
	svc := newTestSvc(
		CronJob{ID: "1", Enabled: true},
		CronJob{ID: "2", Enabled: false},
	)
	jobs := svc.ListJobs(true)
	if len(jobs) != 2 {
		t.Errorf("want 2, got %d", len(jobs))
	}
}

func TestListJobs_ExcludeDisabled_ReturnsOnlyEnabled(t *testing.T) {
	svc := newTestSvc(
		CronJob{ID: "1", Enabled: true},
		CronJob{ID: "2", Enabled: false},
		CronJob{ID: "3", Enabled: true},
	)
	jobs := svc.ListJobs(false)
	if len(jobs) != 2 {
		t.Errorf("want 2 enabled jobs, got %d", len(jobs))
	}
	for _, j := range jobs {
		if !j.Enabled {
			t.Errorf("unexpected disabled job %q in result", j.ID)
		}
	}
}

func TestListJobs_Empty(t *testing.T) {
	svc := newTestSvc()
	if jobs := svc.ListJobs(true); len(jobs) != 0 {
		t.Errorf("expected 0 jobs, got %d", len(jobs))
	}
}

// --- Status ---

func TestStatus_ReturnsJobCount(t *testing.T) {
	svc := newTestSvc(
		CronJob{ID: "1", Enabled: true},
		CronJob{ID: "2", Enabled: false},
	)
	status := svc.Status()
	if status["jobs"] != 2 {
		t.Errorf("want jobs=2, got %v", status["jobs"])
	}
}

func TestStatus_RunningFalseWhenNotStarted(t *testing.T) {
	svc := newTestSvc()
	status := svc.Status()
	if status["enabled"] != false {
		t.Errorf("expected enabled=false for unstarted service, got %v", status["enabled"])
	}
}
