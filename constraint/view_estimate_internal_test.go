package constraint

import (
	"testing"

	"github.com/faustbrian/go-knapsack/v2"
)

func TestViewEstimateAdmitsExactBudget(t *testing.T) {
	estimate := viewEstimate{}
	if !estimate.add(3, 256) || !estimate.add(maxViewBytes-3*256, 1) {
		t.Fatal("exact cumulative budget was refused")
	}
	if estimate.add(1, 1) {
		t.Fatal("one unit beyond the cumulative budget was admitted")
	}
	if !estimate.add(0, 16) {
		t.Fatal("empty collection consumed capacity at the exact budget")
	}
}

func TestViewEstimateRefusalDoesNotConsumeBudget(t *testing.T) {
	estimate := viewEstimate{}
	if !estimate.add(maxViewBytes/2, 1) {
		t.Fatal("first half of budget was refused")
	}
	if estimate.add(maxViewBytes/2+1, 1) {
		t.Fatal("cumulative one-over budget was admitted")
	}
	if !estimate.add(maxViewBytes/2, 1) {
		t.Fatal("refused admission changed the remaining capacity")
	}
	if estimate.add(1, 1) {
		t.Fatal("exactly filled budget retained extra capacity")
	}
}

func TestPlacementEstimateChargesDiagnosticCollection(t *testing.T) {
	estimate := viewEstimate{}
	if !estimate.add(maxViewBytes-64-2, 1) {
		t.Fatal("scalar prefill was refused")
	}
	placement := knapsack.Placement{Diagnostics: []knapsack.Diagnostic{{Code: "c", Message: "m"}}}
	if !placementWithinLimits(&estimate, placement) {
		t.Fatal("one diagnostic at exact remaining collection capacity was refused")
	}
	if estimate.add(1, 1) {
		t.Fatal("diagnostic collection charge was omitted")
	}
}

func TestPlacementEstimateChargesDiagnosticIdentifiers(t *testing.T) {
	estimate := viewEstimate{}
	if !estimate.add(maxViewBytes-64-4, 1) {
		t.Fatal("scalar prefill was refused")
	}
	placement := knapsack.Placement{
		Diagnostics: []knapsack.Diagnostic{{Code: "c", ItemID: "i", ContainerID: "b", Message: "m"}},
	}
	if !placementWithinLimits(&estimate, placement) {
		t.Fatal("diagnostic identifiers at exact remaining capacity were refused")
	}
	if estimate.add(1, 1) {
		t.Fatal("diagnostic identifier charge was omitted")
	}
}
