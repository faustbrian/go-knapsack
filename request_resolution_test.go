package knapsack_test

import (
	"errors"
	"reflect"
	"testing"

	k "github.com/faustbrian/go-knapsack/v2"
	"github.com/faustbrian/go-knapsack/v2/geometry"
	"github.com/faustbrian/go-math/decimal"
	m "github.com/faustbrian/go-measurement/v2"
)

func resolutionContractQuantity(value string, unit m.Unit) m.Quantity {
	return m.MustNew(decimal.MustParse(value), unit)
}
func resolutionContractSpec() k.NormalizedSpec {
	return k.NormalizedSpec{
		Items:      []k.NormalizedItem{{ID: "item", Dimensions: geometry.Dimensions{X: 1, Y: 1, Z: 1}, Weight: 1, Orientations: []geometry.Orientation{geometry.OrientationXYZ}}},
		Containers: []k.NormalizedContainer{{ID: "box", Dimensions: geometry.Dimensions{X: 2, Y: 2, Z: 2}, MaxContentWeight: 2, Stock: k.UnlimitedStock()}},
		Resolution: k.Resolution{Length: resolutionContractQuantity("1", m.Metre), Mass: resolutionContractQuantity("1", m.Kilogram)}, Limits: k.DefaultLimits()}
}
func TestNormalizedResolutionContract(t *testing.T) {
	cases := []struct {
		name         string
		length, mass m.Quantity
		valid        bool
	}{
		{"valid", resolutionContractQuantity("1", m.Metre), resolutionContractQuantity("1", m.Kilogram), true},
		{"length unknown", m.Quantity{}, resolutionContractQuantity("1", m.Kilogram), false},
		{"mass unknown", resolutionContractQuantity("1", m.Metre), m.Quantity{}, false},
		{"both unknown", m.Quantity{}, m.Quantity{}, false},
		{"length wrong", resolutionContractQuantity("1", m.Kilogram), resolutionContractQuantity("1", m.Kilogram), false},
		{"mass wrong", resolutionContractQuantity("1", m.Metre), resolutionContractQuantity("1", m.Metre), false},
		{"length dimensionless", resolutionContractQuantity("1", m.One), resolutionContractQuantity("1", m.Kilogram), false},
		{"mass dimensionless", resolutionContractQuantity("1", m.Metre), resolutionContractQuantity("1", m.One), false},
		{"length zero", resolutionContractQuantity("0", m.Metre), resolutionContractQuantity("1", m.Kilogram), false},
		{"mass zero", resolutionContractQuantity("1", m.Metre), resolutionContractQuantity("0", m.Kilogram), false},
		{"length negative", resolutionContractQuantity("-1", m.Metre), resolutionContractQuantity("1", m.Kilogram), false},
		{"mass negative", resolutionContractQuantity("1", m.Metre), resolutionContractQuantity("-1", m.Kilogram), false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			s := resolutionContractSpec()
			s.Resolution = k.Resolution{Length: c.length, Mass: c.mass}
			got, err := k.NewNormalizedRequest(s)
			if c.valid {
				if err != nil {
					t.Fatal(err)
				}
				if !reflect.DeepEqual(got.Resolution(), s.Resolution) || len(got.Items()) != 1 || got.Items()[0].ID != "item" || len(got.Containers()) != 1 || got.Containers()[0].ID != "box" {
					t.Fatal("valid input not preserved")
				}
				return
			}
			if !errors.Is(err, k.ErrInvalidRequest) || !reflect.DeepEqual(got, k.NormalizedRequest{}) {
				t.Fatalf("got %#v error %v, want zero request and ErrInvalidRequest", got, err)
			}
		})
	}
}
