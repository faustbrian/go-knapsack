// Package knapsack models exact, deterministic three-dimensional orthogonal
// bin packing with scalar weight capacity. It is not the classical
// value-maximizing knapsack problem. Solvers live in package solver and every
// returned plan can be checked independently with package verify.
//
// Start with NewItem, NewContainerType and NewRequest, or NewNormalizedRequest
// for already scaled input. Normalization requires exact divisibility by positive
// length/mass resolution and int64-representable lattice values; no silent
// rounding or geometric epsilon is applied. Money objectives remain optional
// nested modules, separate from the root Math and Measurement dependencies.
//
// Normalized requests and plans expose copied collections. Physical Item has a
// retained optional load-limit pointer; see NewItem before sharing its input.
// Plan constructors admit bounded representation, not feasibility or proof.
// Verify supplied plans and inspect both verification errors and findings;
// status metadata alone cannot establish optimality or infeasibility.
package knapsack
