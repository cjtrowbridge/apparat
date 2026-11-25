// Verifies historical Ollama token rates for one executor and exact model identity.
// Includes cached input, unequal durations, missing telemetry and replicated jobs.
package persistence

import (
	"context"
	"github.com/cjtrowbridge/apparat/internal/domain"
	"testing"
)

func TestModelSpeed(t *testing.T) {
	ctx := context.Background()
	db, err := Open(ctx, t.TempDir()+"/worker.db")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	err = db.Migrate(ctx)
	if err != nil {
		t.Fatal(err)
	}
	model := domain.ModelRef{Provider: "ollama", Name: "a", Digest: "digest"}
	record := func(id, executor string, ref domain.ModelRef, metrics map[string]float64) {
		t.Helper()
		job := domain.InferenceJob{Packet: domain.ChatRequest{Job: id, Model: ref}}
		claim := domain.InferenceClaim{Job: id, Executor: executor, Digest: domain.InferenceDigest(job)}
		err := db.AdmitInference(ctx, job)
		if err == nil {
			err = db.SealInference(ctx, claim)
		}
		if err == nil {
			err = db.FinishInference(ctx, claim, domain.ChatOutput{Metrics: metrics})
		}
		if err != nil {
			t.Fatal(err)
		}
	}
	m := map[string]float64{"prompt_eval_count": 120, "prompt_eval_cached_count": 20, "prompt_eval_duration": 1e9, "eval_count": 12, "eval_duration": 2e9}
	record("a", "worker", model, m)
	record("b", "worker", model, map[string]float64{"eval_count": 6, "eval_duration": 1e9, "prompt_eval_count": 20, "prompt_eval_duration": 3e9})
	record("replica", "other", model, m)
	other := model
	other.Digest = "other"
	record("other-model", "worker", other, m)
	record("missing", "worker", model, nil)
	record("a", "worker", model, m)
	usage, err := db.DeviceUsage(ctx, "worker")
	if err != nil || usage.Prompts != 4 || usage.Tokens != 30 {
		t.Fatal("executor usage after replay", usage, err)
	}
	v, err := db.ModelSpeed(ctx, "worker", model)
	if err != nil || v.Prefill == nil || *v.Prefill != 30 || v.Eval == nil || *v.Eval != 6 {
		t.Fatalf("wrong measured rates: %+v %v", v, err)
	}
	v, err = db.ModelSpeed(ctx, "unused", model)
	if err != nil || v.Prefill != nil || v.Eval != nil {
		t.Fatal("expected missing measurements", v, err)
	}
}
