package knapsack_test

import (
	"errors"
	"reflect"
	"testing"

	"github.com/faustbrian/go-knapsack/v2"
	"github.com/faustbrian/go-knapsack/v2/geometry"
	"github.com/faustbrian/go-measurement/v2"
)

func boundaryContainerSpec() knapsack.ContainerTypeSpec {
	return knapsack.ContainerTypeSpec{
		ID: "bb", InternalDimensions: knapsack.PhysicalDimensions{
			X: length("2", measurement.Metre), Y: length("2", measurement.Metre), Z: length("2", measurement.Metre),
		}, MaxContentWeight: mass("2", measurement.Kilogram), Stock: knapsack.FiniteStock(1),
	}
}

func TestContainerOptionalMassAndReservedOriginBoundaries(t *testing.T) {
	tare, gross := mass("1", measurement.Kilogram), mass("1000", measurement.Gram)
	resolution := knapsack.Resolution{Length: length("1", measurement.Centimetre), Mass: mass("1", measurement.Gram)}
	item, err := knapsack.NewItem(itemSpec("aa"))
	if err != nil {
		t.Fatal(err)
	}
	for _, test := range []struct {
		name                string
		tare, gross         *measurement.Quantity
		wantTare, wantGross int64
	}{
		{"neither", nil, nil, 0, 0},
		{"tare only", &tare, nil, 1000, 0},
		{"gross only", nil, &gross, 0, 1000},
		{"both equal", &tare, &gross, 1000, 1000},
	} {
		t.Run(test.name, func(t *testing.T) {
			spec := boundaryContainerSpec()
			spec.TareWeight, spec.MaxGrossWeight = test.tare, test.gross
			container, err := knapsack.NewContainerType(spec)
			if err != nil {
				t.Fatal(err)
			}
			request, err := knapsack.NewRequest([]knapsack.Item{item}, []knapsack.ContainerType{container}, resolution, knapsack.DefaultLimits())
			if err != nil {
				t.Fatal(err)
			}
			got := request.Normalized().Containers()[0]
			if got.TareWeight != test.wantTare || got.MaxGrossWeight != test.wantGross || got.HasGrossWeight != (test.gross != nil) || got.MaxContentWeight != 2000 {
				t.Fatalf("normalized masses = %+v", got)
			}
		})
	}
	spec := boundaryContainerSpec()
	below := mass("999", measurement.Gram)
	spec.TareWeight, spec.MaxGrossWeight = &tare, &below
	if got, err := knapsack.NewContainerType(spec); !errors.Is(err, knapsack.ErrInvalidContainer) || !reflect.DeepEqual(got, knapsack.ContainerType{}) {
		t.Fatalf("below tare = %+v, %v", got, err)
	}
	for _, origin := range []geometry.Point{{X: -1}, {Y: -1}, {Z: -1}} {
		spec := boundaryContainerSpec()
		spec.Reserved = []knapsack.ReservedRegion{{Origin: origin, Dimensions: spec.InternalDimensions}}
		if got, err := knapsack.NewContainerType(spec); !errors.Is(err, knapsack.ErrInvalidContainer) || !reflect.DeepEqual(got, knapsack.ContainerType{}) {
			t.Fatalf("negative origin %+v = %+v, %v", origin, got, err)
		}
	}
}

func TestPlanRequiredLimitsAreIndependentlyNonzero(t *testing.T) {
	for _, test := range []struct {
		name string
		zero func(*knapsack.PlanLimits)
	}{
		{"containers", func(l *knapsack.PlanLimits) { l.MaxContainers = 0 }},
		{"placements", func(l *knapsack.PlanLimits) { l.MaxPlacements = 0 }},
		{"unpacked", func(l *knapsack.PlanLimits) { l.MaxUnpackedItems = 0 }},
		{"objective", func(l *knapsack.PlanLimits) { l.MaxObjectiveComponents = 0 }},
		{"diagnostics", func(l *knapsack.PlanLimits) { l.MaxDiagnostics = 0 }},
		{"IDs", func(l *knapsack.PlanLimits) { l.MaxIDBytes = 0 }},
		{"metadata", func(l *knapsack.PlanLimits) { l.MaxMetadataBytes = 0 }},
		{"bytes", func(l *knapsack.PlanLimits) { l.MaxBytes = 0 }},
	} {
		t.Run(test.name, func(t *testing.T) {
			limits := knapsack.DefaultPlanLimits()
			test.zero(&limits)
			got, err := knapsack.NewPlanWithLimits(knapsack.PlanSpec{Status: knapsack.StatusFeasible, Termination: knapsack.TerminationCompleted}, limits)
			if !errors.Is(err, knapsack.ErrInvalidOptions) || !reflect.DeepEqual(got, knapsack.Plan{}) {
				t.Fatalf("plan=%+v error=%v", got, err)
			}
		})
	}
}

func TestPlanCollectionLimitsAreInclusive(t *testing.T) {
	for _, test := range []struct {
		name string
		set  func(*knapsack.PlanSpec, int)
	}{
		{"containers", func(s *knapsack.PlanSpec, n int) { s.Containers = make([]knapsack.ContainerInstance, n) }},
		{"placements", func(s *knapsack.PlanSpec, n int) { s.Placements = make([]knapsack.Placement, n) }},
		{"unpacked", func(s *knapsack.PlanSpec, n int) { s.UnpackedItemIDs = make([]string, n) }},
		{"objective", func(s *knapsack.PlanSpec, n int) { s.Objective = make([]knapsack.ScoreComponent, n) }},
		{"top-level diagnostics", func(s *knapsack.PlanSpec, n int) { s.Diagnostics = make([]knapsack.Diagnostic, n) }},
		{"placement diagnostics", func(s *knapsack.PlanSpec, n int) {
			s.Placements = []knapsack.Placement{{Diagnostics: make([]knapsack.Diagnostic, n)}}
		}},
	} {
		t.Run(test.name, func(t *testing.T) {
			limits := knapsack.PlanLimits{MaxContainers: 1, MaxPlacements: 1, MaxUnpackedItems: 1, MaxObjectiveComponents: 1, MaxDiagnostics: 1, MaxIDBytes: 3, MaxMetadataBytes: 3, MaxBytes: 1024}
			for _, n := range []int{1, 2} {
				spec := knapsack.PlanSpec{Status: knapsack.StatusFeasible, Termination: knapsack.TerminationCompleted}
				test.set(&spec, n)
				got, err := knapsack.NewPlanWithLimits(spec, limits)
				if n == 1 {
					if err != nil || !reflect.DeepEqual(got.Spec(), spec) {
						t.Fatalf("exact plan=%+v error=%v", got.Spec(), err)
					}
				} else if !errors.Is(err, knapsack.ErrBudgetExhausted) || !reflect.DeepEqual(got, knapsack.Plan{}) {
					t.Fatalf("over plan=%+v error=%v", got, err)
				}
			}
		})
	}
}

func boundaryPlanSpec() knapsack.PlanSpec {
	return knapsack.PlanSpec{
		Containers:      []knapsack.ContainerInstance{{ID: "c", TypeID: "t"}},
		Placements:      []knapsack.Placement{{ItemID: "i", ContainerID: "c", SupporterIDs: []string{"s"}, Diagnostics: []knapsack.Diagnostic{{Code: "d", ItemID: "i", ContainerID: "c", Message: "m"}}}},
		UnpackedItemIDs: []string{"u"}, Objective: []knapsack.ScoreComponent{{Name: "n", Direction: "min", Unit: "u", Value: "v"}},
		Diagnostics: []knapsack.Diagnostic{{Code: "d", ItemID: "i", ContainerID: "c", Message: "m"}}, Status: knapsack.StatusFeasible, Termination: knapsack.TerminationCompleted,
	}
}

func TestPlanStringAndAggregateLimitsAreInclusive(t *testing.T) {
	for _, test := range []struct {
		name string
		set  func(*knapsack.PlanSpec, string)
	}{
		{"instance ID", func(s *knapsack.PlanSpec, v string) { s.Containers[0].ID = v }},
		{"type ID", func(s *knapsack.PlanSpec, v string) { s.Containers[0].TypeID = v }},
		{"item ID", func(s *knapsack.PlanSpec, v string) { s.Placements[0].ItemID = v }},
		{"placement container ID", func(s *knapsack.PlanSpec, v string) { s.Placements[0].ContainerID = v }},
		{"supporter ID", func(s *knapsack.PlanSpec, v string) { s.Placements[0].SupporterIDs[0] = v }},
		{"unpacked ID", func(s *knapsack.PlanSpec, v string) { s.UnpackedItemIDs[0] = v }},
		{"objective name", func(s *knapsack.PlanSpec, v string) { s.Objective[0].Name = v }},
		{"objective direction", func(s *knapsack.PlanSpec, v string) { s.Objective[0].Direction = v }},
		{"objective unit", func(s *knapsack.PlanSpec, v string) { s.Objective[0].Unit = v }},
		{"objective value", func(s *knapsack.PlanSpec, v string) { s.Objective[0].Value = v }},
		{"placement diagnostic code", func(s *knapsack.PlanSpec, v string) { s.Placements[0].Diagnostics[0].Code = v }},
		{"placement diagnostic item", func(s *knapsack.PlanSpec, v string) { s.Placements[0].Diagnostics[0].ItemID = v }},
		{"placement diagnostic container", func(s *knapsack.PlanSpec, v string) { s.Placements[0].Diagnostics[0].ContainerID = v }},
		{"placement diagnostic message", func(s *knapsack.PlanSpec, v string) { s.Placements[0].Diagnostics[0].Message = v }},
		{"plan diagnostic code", func(s *knapsack.PlanSpec, v string) { s.Diagnostics[0].Code = v }},
		{"plan diagnostic item", func(s *knapsack.PlanSpec, v string) { s.Diagnostics[0].ItemID = v }},
		{"plan diagnostic container", func(s *knapsack.PlanSpec, v string) { s.Diagnostics[0].ContainerID = v }},
		{"plan diagnostic message", func(s *knapsack.PlanSpec, v string) { s.Diagnostics[0].Message = v }},
	} {
		t.Run(test.name, func(t *testing.T) {
			limits := knapsack.DefaultPlanLimits()
			limits.MaxIDBytes = 3
			limits.MaxMetadataBytes = 3
			for _, v := range []string{"abc", "abcd"} {
				spec := boundaryPlanSpec()
				test.set(&spec, v)
				got, err := knapsack.NewPlanWithLimits(spec, limits)
				if v == "abc" {
					if err != nil || !reflect.DeepEqual(got.Spec(), spec) {
						t.Fatalf("exact plan=%+v error=%v", got.Spec(), err)
					}
				} else if !errors.Is(err, knapsack.ErrBudgetExhausted) || !reflect.DeepEqual(got, knapsack.Plan{}) {
					t.Fatalf("over plan=%+v error=%v", got, err)
				}
			}
		})
	}
	limits := knapsack.DefaultPlanLimits()
	limits.MaxDiagnostics = 2
	spec := boundaryPlanSpec()
	if got, err := knapsack.NewPlanWithLimits(spec, limits); err != nil || len(got.Diagnostics()) != 1 || len(got.Placements()[0].Diagnostics) != 1 {
		t.Fatalf("combined exact plan=%+v error=%v", got, err)
	}
	spec.Diagnostics = append(spec.Diagnostics, knapsack.Diagnostic{Code: "d"})
	if got, err := knapsack.NewPlanWithLimits(spec, limits); !errors.Is(err, knapsack.ErrBudgetExhausted) || !reflect.DeepEqual(got, knapsack.Plan{}) {
		t.Fatalf("combined over plan=%+v error=%v", got, err)
	}
	// Literal conservative accounting: container 76, placement with supporter
	// and diagnostic 378, unpacked ID 22, objective 100, plan diagnostic 88.
	for _, budget := range []uint64{664, 663} {
		limits := knapsack.DefaultPlanLimits()
		limits.MaxBytes = budget
		got, err := knapsack.NewPlanWithLimits(boundaryPlanSpec(), limits)
		if budget == 664 {
			if err != nil || len(got.Placements()) != 1 {
				t.Fatalf("exact bytes plan=%+v error=%v", got, err)
			}
		} else if !errors.Is(err, knapsack.ErrBudgetExhausted) || !reflect.DeepEqual(got, knapsack.Plan{}) {
			t.Fatalf("one-less bytes plan=%+v error=%v", got, err)
		}
	}
}
