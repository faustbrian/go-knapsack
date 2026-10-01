package knapsack_test

import (
	"errors"
	"reflect"
	"testing"

	"github.com/faustbrian/go-knapsack/v2"
	"github.com/faustbrian/go-knapsack/v2/geometry"
	"github.com/faustbrian/go-measurement/v2"
)

func boundaryNormalizedSpec() knapsack.NormalizedSpec {
	return knapsack.NormalizedSpec{
		Items:      []knapsack.NormalizedItem{{ID: "aa", Dimensions: geometry.Dimensions{X: 1, Y: 1, Z: 1}, Weight: 1, Orientations: []geometry.Orientation{geometry.OrientationXYZ}}},
		Containers: []knapsack.NormalizedContainer{{ID: "bb", Dimensions: geometry.Dimensions{X: 2, Y: 2, Z: 2}, MaxContentWeight: 2, Stock: knapsack.FiniteStock(1)}},
		Resolution: knapsack.Resolution{Length: length("1", measurement.Centimetre), Mass: mass("1", measurement.Kilogram)}, Limits: knapsack.DefaultLimits(),
	}
}

func TestRequestCollectionAndIDLimitsAreInclusive(t *testing.T) {
	for _, test := range []struct {
		name   string
		change func(*knapsack.NormalizedSpec, *knapsack.ItemSpec, *knapsack.ContainerTypeSpec)
		target error
	}{
		{"items", func(s *knapsack.NormalizedSpec, _ *knapsack.ItemSpec, _ *knapsack.ContainerTypeSpec) {
			next := s.Items[0]
			next.ID = "cc"
			s.Items = append(s.Items, next)
		}, knapsack.ErrInvalidRequest},
		{"containers", func(s *knapsack.NormalizedSpec, _ *knapsack.ItemSpec, _ *knapsack.ContainerTypeSpec) {
			next := s.Containers[0]
			next.ID = "cc"
			s.Containers = append(s.Containers, next)
		}, knapsack.ErrInvalidRequest},
		{"item ID", func(s *knapsack.NormalizedSpec, i *knapsack.ItemSpec, _ *knapsack.ContainerTypeSpec) {
			s.Items[0].ID = "aaa"
			i.ID = "aaa"
		}, knapsack.ErrInvalidItem},
		{"container ID", func(s *knapsack.NormalizedSpec, _ *knapsack.ItemSpec, c *knapsack.ContainerTypeSpec) {
			s.Containers[0].ID = "bbb"
			c.ID = "bbb"
		}, knapsack.ErrInvalidContainer},
	} {
		t.Run(test.name, func(t *testing.T) {
			for _, over := range []bool{false, true} {
				spec := boundaryNormalizedSpec()
				spec.Limits.MaxItems = 1
				spec.Limits.MaxContainerTypes = 1
				spec.Limits.MaxIDBytes = 2
				itemSpec, containerSpec := itemSpec("aa"), boundaryContainerSpec()
				if over {
					test.change(&spec, &itemSpec, &containerSpec)
				}
				normalized, err := knapsack.NewNormalizedRequest(spec)
				if !over {
					if err != nil || normalized.ItemCount() != 1 || normalized.ContainerTypeCount() != 1 || normalized.Items()[0].ID != "aa" || normalized.Containers()[0].ID != "bb" {
						t.Fatalf("exact normalized=%+v error=%v", normalized, err)
					}
				} else if !errors.Is(err, test.target) || !reflect.DeepEqual(normalized, knapsack.NormalizedRequest{}) {
					t.Fatalf("over normalized=%+v error=%v", normalized, err)
				}
				item, err := knapsack.NewItem(itemSpec)
				if err != nil {
					t.Fatal(err)
				}
				container, err := knapsack.NewContainerType(containerSpec)
				if err != nil {
					t.Fatal(err)
				}
				items, containers := []knapsack.Item{item}, []knapsack.ContainerType{container}
				if over && test.name == "items" {
					itemSpec.ID = "cc"
					next, e := knapsack.NewItem(itemSpec)
					if e != nil {
						t.Fatal(e)
					}
					items = append(items, next)
				}
				if over && test.name == "containers" {
					containerSpec.ID = "cc"
					next, e := knapsack.NewContainerType(containerSpec)
					if e != nil {
						t.Fatal(e)
					}
					containers = append(containers, next)
				}
				request, err := knapsack.NewRequest(items, containers, spec.Resolution, spec.Limits)
				if !over {
					if err != nil || request.Normalized().ItemCount() != 1 || request.Normalized().ContainerTypeCount() != 1 || request.Normalized().Items()[0].Weight != 2 {
						t.Fatalf("exact physical=%+v error=%v", request, err)
					}
				} else if !errors.Is(err, test.target) || !reflect.DeepEqual(request, knapsack.Request{}) {
					t.Fatalf("over physical=%+v error=%v", request, err)
				}
			}
		})
	}
}

func TestRequestOrientationLimitIsInclusive(t *testing.T) {
	for _, over := range []bool{false, true} {
		spec := boundaryNormalizedSpec()
		spec.Limits.MaxOrientations = 1
		physicalSpec := itemSpec("aa")
		if over {
			spec.Items[0].Orientations = append(spec.Items[0].Orientations, geometry.OrientationZYX)
			physicalSpec.Orientations = append(physicalSpec.Orientations, geometry.OrientationZYX)
		}
		request, err := knapsack.NewNormalizedRequest(spec)
		if !over {
			if err != nil || len(request.Items()[0].Orientations) != 1 {
				t.Fatalf("exact normalized=%+v error=%v", request, err)
			}
		} else if !errors.Is(err, knapsack.ErrInvalidItem) || !reflect.DeepEqual(request, knapsack.NormalizedRequest{}) {
			t.Fatalf("over normalized=%+v error=%v", request, err)
		}
		item, err := knapsack.NewItem(physicalSpec)
		if err != nil {
			t.Fatal(err)
		}
		container, err := knapsack.NewContainerType(boundaryContainerSpec())
		if err != nil {
			t.Fatal(err)
		}
		physical, err := knapsack.NewRequest([]knapsack.Item{item}, []knapsack.ContainerType{container}, spec.Resolution, spec.Limits)
		if !over {
			if err != nil || len(physical.Normalized().Items()[0].Orientations) != 1 {
				t.Fatalf("exact physical=%+v error=%v", physical, err)
			}
		} else if !errors.Is(err, knapsack.ErrInvalidItem) || !reflect.DeepEqual(physical, knapsack.Request{}) {
			t.Fatalf("over physical=%+v error=%v", physical, err)
		}
	}
}

func TestRequestMemoryBudgetIsInclusiveForNestedMetadata(t *testing.T) {
	physicalSpec := itemSpec("aa")
	physicalSpec.SKU = "s"
	physicalSpec.Group = "g"
	physicalSpec.Attributes = map[string]string{"k": "v"}
	physicalSpec.IncompatibleGroups = []string{"x"}
	item, err := knapsack.NewItem(physicalSpec)
	if err != nil {
		t.Fatal(err)
	}
	containerSpec := boundaryContainerSpec()
	containerSpec.AllowedClasses = []string{"a"}
	containerSpec.Reserved = []knapsack.ReservedRegion{{Dimensions: knapsack.PhysicalDimensions{X: length("1", measurement.Centimetre), Y: length("1", measurement.Centimetre), Z: length("1", measurement.Centimetre)}}}
	container, err := knapsack.NewContainerType(containerSpec)
	if err != nil {
		t.Fatal(err)
	}
	resolution := boundaryNormalizedSpec().Resolution
	baseline, err := knapsack.NewRequest([]knapsack.Item{item}, []knapsack.ContainerType{container}, resolution, knapsack.DefaultLimits())
	if err != nil {
		t.Fatal(err)
	}
	normalized := baseline.Normalized()
	// One item/container, orientation, attribute pair, incompatible group,
	// allowed class and reserved cuboid, plus their short strings total 698.
	if normalized.MemoryBytes() != 698 {
		t.Fatalf("owned estimate=%d, want 698", normalized.MemoryBytes())
	}
	for _, budget := range []uint64{normalized.MemoryBytes(), normalized.MemoryBytes() - 1} {
		limits := knapsack.DefaultLimits()
		limits.MaxMemoryBytes = budget
		request, err := knapsack.NewRequest([]knapsack.Item{item}, []knapsack.ContainerType{container}, resolution, limits)
		if budget == 698 {
			if err != nil || request.Normalized().MemoryBytes() != 698 || request.Normalized().Items()[0].Attributes["k"] != "v" {
				t.Fatalf("exact physical=%+v error=%v", request, err)
			}
		} else if !errors.Is(err, knapsack.ErrMemoryBudgetExhausted) || !reflect.DeepEqual(request, knapsack.Request{}) {
			t.Fatalf("one-less physical=%+v error=%v", request, err)
		}
		n, err := knapsack.NewNormalizedRequest(knapsack.NormalizedSpec{Items: normalized.Items(), Containers: normalized.Containers(), Resolution: resolution, Limits: limits})
		if budget == 698 {
			if err != nil || n.MemoryBytes() != 698 || len(n.Containers()[0].Reserved) != 1 {
				t.Fatalf("exact normalized=%+v error=%v", n, err)
			}
		} else if !errors.Is(err, knapsack.ErrMemoryBudgetExhausted) || !reflect.DeepEqual(n, knapsack.NormalizedRequest{}) {
			t.Fatalf("one-less normalized=%+v error=%v", n, err)
		}
	}
}

func TestNormalizedSemanticBoundsHaveIndependentInclusiveCases(t *testing.T) {
	for _, test := range []struct {
		name   string
		change func(*knapsack.NormalizedSpec)
		target error
	}{
		{"mass dimension", func(s *knapsack.NormalizedSpec) { s.Resolution.Mass = length("1", measurement.Metre) }, knapsack.ErrInvalidRequest},
		{"negative tare", func(s *knapsack.NormalizedSpec) { s.Containers[0].TareWeight = -1 }, knapsack.ErrInvalidContainer},
		{"gross below tare", func(s *knapsack.NormalizedSpec) { s.Containers[0].MaxGrossWeight = 0 }, knapsack.ErrInvalidContainer},
		{"support above limit", func(s *knapsack.NormalizedSpec) { s.Items[0].MinimumSupportPPM = 1_000_001 }, knapsack.ErrInvalidItem},
	} {
		t.Run(test.name, func(t *testing.T) {
			spec := boundaryNormalizedSpec()
			zero := int64(0)
			spec.Items[0].MaxSupportedWeight = &zero
			spec.Items[0].MinimumSupportPPM = 1_000_000
			spec.Containers[0].HasGrossWeight = true
			spec.Containers[0].TareWeight = 1
			spec.Containers[0].MaxGrossWeight = 1
			got, err := knapsack.NewNormalizedRequest(spec)
			if err != nil || got.Containers()[0].MaxGrossWeight != 1 || got.Items()[0].MinimumSupportPPM != 1_000_000 || *got.Items()[0].MaxSupportedWeight != 0 {
				t.Fatalf("inclusive semantic request=%+v error=%v", got, err)
			}
			test.change(&spec)
			got, err = knapsack.NewNormalizedRequest(spec)
			if !errors.Is(err, test.target) || !reflect.DeepEqual(got, knapsack.NormalizedRequest{}) {
				t.Fatalf("invalid request=%+v error=%v", got, err)
			}
		})
	}
}
