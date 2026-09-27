package knapsack_test

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"testing"
)

const boxPackerRuntimeRaw = "docs/benchmarks/raw/2026-09-13-boxpacker-runtime.json"
const boxPackerRuntimeEvidenceDate = "2026-09-13"

func TestHistoricalBoxPackerRuntimeEvidenceRetainsIntegrity(t *testing.T) {
	t.Parallel()

	data, err := os.ReadFile(boxPackerRuntimeRaw)
	if err != nil {
		t.Fatal(err)
	}
	var evidence runtimeComparisonEvidence
	if err := json.Unmarshal(data, &evidence); err != nil {
		t.Fatal(err)
	}
	if value, err := hex.DecodeString(evidence.InputSHA256); err != nil || len(value) != sha256.Size {
		t.Fatal("historical BoxPacker runtime evidence has an invalid recorded input digest")
	}
	if evidence.SchemaVersion != "v1" || evidence.Date != boxPackerRuntimeEvidenceDate ||
		evidence.Environment == "" || evidence.Processor == "" ||
		evidence.SamplesPerAdapter < 10 || evidence.WarmupsPerAdapter < 1 ||
		evidence.CommonSemantics == "" || evidence.AllocationMeasurement == "" ||
		evidence.Verification == "" || len(evidence.Samples) != 2*evidence.SamplesPerAdapter ||
		len(evidence.Summaries) != 2 || len(evidence.Cancellation) != 2 {
		t.Fatalf("BoxPacker runtime evidence is incomplete: %+v", evidence)
	}
	for _, sample := range evidence.Samples {
		if sample.Implementation == "" || sample.RuntimeVersion == "" || sample.Sample < 1 ||
			sample.WallNanoseconds <= 0 || sample.SolveNanoseconds <= 0 ||
			sample.PeakRSSBytes <= 0 || sample.ContainerCount != 1 || sample.PackedItems != 2 {
			t.Fatalf("invalid BoxPacker runtime sample: %+v", sample)
		}
	}
	for _, cancellation := range evidence.Cancellation {
		if cancellation.Implementation == "" || cancellation.DeadlineNanoseconds <= 0 ||
			cancellation.ReturnNanoseconds <= 0 || !cancellation.ProcessStopped {
			t.Fatalf("invalid BoxPacker cancellation evidence: %+v", cancellation)
		}
	}
}
