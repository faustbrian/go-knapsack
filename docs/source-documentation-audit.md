# Source documentation audit

## Initial scope and contract plan

Tier A comments and directly contradicted maintained current prose only, from
`cca3bd0b375b2d1d5769b460c77c141515d85781`. The local worker owns this source
batch; the coordinator owns remote-main delivery and CI. The original unrelated
Dependabot checkout has substantial existing edits and is left untouched.
Reproduce inventory with `git ls-files "*.go"`, `git ls-files "*go.mod"` and
`golib source-docs check --json`; entrypoints use
`rg --no-ignore "^func (Test|Fuzz|Benchmark|Example)" --glob "*_test.go"`.

All 105 tracked Go files, 18 implementation files and 87 assurance files are
in scope across four modules: root /v2, objective/gomoney /v2,
objective/money /v3, and integration/references. Thirteen tracked source
locations include twelve production/support/command packages and one
reference-harness test-only package. The initial five structural gaps are a
command package header, three objective interface methods and a request method;
existing extensive comments still require semantic truth/usefulness review.

Outcome: source-grounded packing/geometry/objective/encoding/verification,
provider and adapter contracts with exact arithmetic, input/metadata admission,
lattice normalization, partial plans, deterministic ordering and solver budget
limits explicit. Trace root constructors and solvers, independent verifier,
encoding/visualization/test helpers, Money version-specific adapters and public
reference harness as direct consumers. Named risks are overflow/normalization,
search completeness versus feasible results, caller-owned callback cooperation,
copy versus alias ownership, support-load accounting, malformed input allocation,
and legacy facade compatibility. No algorithm, dependency, fixture, corpus,
benchmark authority or desired expectation repair is scoped.

Smallest checks: pinned Go1.27.2 formatting and every-file AST/executable-token/
literal/directive equivalence; native renders for applicable package/module
variants; configured root/adapter docs recipes including their actual Example
operations; objective source/local links and offline specification structure,
pinned changed-current/added/staged secrets, ordinary complete Tier A review.
Three API gates use apidiff; frozen API baselines remain unchanged for prose.
No generated Go/build variants were found; existing security/deprecation/analyzer
annotations are preserved. Heavy Go/native workflow execution waits for an
explicit serialized coordinator grant with three owned disposable caches.

## Complete tracked coverage

| Module | Tracked package directories |
| --- | --- |
| root `/v2` | `.`, `constraint`, `encoding`, `geometry`, `knapsacktest`, `objective`, `solver`, `verify`, `visualize` |
| `objective/gomoney/v2` | `objective/gomoney` |
| `objective/money/v3` | `objective/money` |
| `integration/references` | test-only root; `cmd/knapsack-compare` command |


| Go file | Module | Initial gaps | Reviewed boundary |
| --- | --- | --- | --- |
| `api_compatibility_test.go` | `.` | 0 | public contracts and private invariant source |
| `architecture_test.go` | `.` | 0 | TestProductionPackageDependencyDirection; TestProductionContainsNoGoroutineLaunches; additional cases and helpers |
| `boxpacker_integration_test.go` | `.` | 0 | TestBoxPackerCommonSubset |
| `boxpacker_runtime_evidence_test.go` | `.` | 0 | TestHistoricalBoxPackerRuntimeEvidenceRetainsIntegrity |
| `boxpacker_runtime_test.go` | `.` | 0 | TestBoxPackerRuntimeComparison; TestRuntimePercentileUsesNearestRank |
| `constraint/admission_bounds_internal_test.go` | `.` | 0 | TestDecisionAdmissionPreservesIndependentInclusiveBounds; TestPlacementViewPreservesInclusivePlacementCount; additional cases and helpers |
| `constraint/constraint.go` | `.` | 0 | public contracts and private invariant source |
| `constraint/constraint_test.go` | `.` | 0 | TestPlacementViewDefensivelyCopiesState; TestPlacementViewDefensivelyCopiesCenterOfGravity; additional cases and helpers |
| `constructor_boundaries_test.go` | `.` | 0 | TestContainerOptionalMassAndReservedOriginBoundaries; TestPlanRequiredLimitsAreIndependentlyNonzero; additional cases and helpers |
| `container.go` | `.` | 0 | public contracts and private invariant source |
| `doc.go` | `.` | 0 | public contracts and private invariant source |
| `encoding/admission_bounds_internal_test.go` | `.` | 0 | TestDecodeRejectsEachNonpositiveLimitBeforeParsing; TestDecodePreservesInclusiveFixtureByteBudget; additional cases and helpers |
| `encoding/compatibility_test.go` | `.` | 0 | TestV1GoldenRequestAndPlanCompatibility; TestUpdateV1GoldenCompatibility |
| `encoding/encoding.go` | `.` | 0 | public contracts and private invariant source |
| `encoding/encoding_test.go` | `.` | 0 | TestDecodeErrorsDoNotDiscloseInput; TestDecodeRedactionPreservesCauseClassification; additional cases and helpers |
| `encoding/fuzz_test.go` | `.` | 0 | FuzzPlanDecode; FuzzRequestDecode |
| `errors.go` | `.` | 0 | public contracts and private invariant source |
| `evidence_test.go` | `.` | 0 | TestEvidenceSourceNormalizationExcludesOnlyParentZipChecksum; TestEvidenceSourceFilesExcludeIgnoredArtifacts; additional cases and helpers |
| `example_test.go` | `.` | 0 | Example_packAll; Example_fixedContainers; additional cases and helpers |
| `geometry/adjacency_bounds_test.go` | `.` | 0 | TestAdjacencyRequiresPositiveFaceAreaOnEveryAxis |
| `geometry/fuzz_test.go` | `.` | 0 | FuzzGeometry; FuzzOrientationEnumeration; additional cases and helpers |
| `geometry/geometry.go` | `.` | 0 | public contracts and private invariant source |
| `geometry/geometry_test.go` | `.` | 0 | TestOrientationsEnumerateAndDeduplicate; TestCuboidUsesHalfOpenGeometry; additional cases and helpers |
| `geometry/oracle_test.go` | `.` | 0 | TestCuboidPredicatesAgreeWithBoundedVoxelOracle; TestCuboidTranslationAndAxisPermutationMetamorphisms; additional cases and helpers |
| `integration/references/adapter_test.go` | `integration/references` | 0 | TestGoComparisonAdapterCommonSubset |
| `integration/references/cmd/knapsack-compare/main.go` | `integration/references` | 1 | public contracts and private invariant source |
| `integration/references/cmd/knapsack-compare/main_test.go` | `integration/references` | 0 | TestRunRejectsInvalidComparisonDelay; TestRunEmitsVerifiedCanonicalPlan; additional cases and helpers |
| `integration/references/differential_test.go` | `integration/references` | 0 | TestCommonIntegralSubset |
| `item.go` | `.` | 0 | public contracts and private invariant source |
| `knapsack_test.go` | `.` | 0 | TestDomainObjectsDefensivelyCopy; TestDomainConstructorsRejectInvalidPhysicalValues; additional cases and helpers |
| `knapsacktest/assertions.go` | `.` | 0 | public contracts and private invariant source |
| `knapsacktest/assertions_test.go` | `.` | 0 | TestRequireCanonicalEqual |
| `native_benchmark_evidence_test.go` | `.` | 0 | TestHistoricalNativeBenchmarkEvidenceRetainsIntegrity; TestHistoricalRSSBenchmarkEvidenceRetainsIntegrity |
| `objective/bounds_test.go` | `.` | 0 | TestComponentsHonorInclusiveCountAndAggregateByteBounds; TestScoresPreserveInclusiveSignedBounds; additional cases and helpers |
| `objective/fuzz_test.go` | `.` | 0 | FuzzObjectiveComparison |
| `objective/gomoney/benchmark_test.go` | `objective/gomoney` | 0 | BenchmarkTotalSixtyFourContainers; BenchmarkComparisonOverhead; additional cases and helpers |
| `objective/gomoney/compatibility_test.go` | `objective/gomoney` | 0 | TestCompatibilityFacadeRetainsIdentityAndDelegatesBehavior |
| `objective/gomoney/example_helpers_test.go` | `objective/gomoney` | 0 | public contracts and private invariant source |
| `objective/gomoney/example_test.go` | `objective/gomoney` | 0 | ExampleNew; ExampleNewWithPolicy |
| `objective/gomoney/fuzz_test.go` | `objective/gomoney` | 0 | FuzzCostTypeIdentifiers; FuzzCostMappings; additional cases and helpers |
| `objective/gomoney/gomoney.go` | `objective/gomoney` | 0 | public contracts and private invariant source |
| `objective/gomoney/gomoney_test.go` | `objective/gomoney` | 0 | TestNewRejectsMoreThanDefaultCostTypes; TestCostObjectiveEntriesRejectNilContext; additional cases and helpers |
| `objective/gomoney/hardening_test.go` | `objective/gomoney` | 0 | TestMain; TestDeterministicRankingAcrossProcessesAndArchitectures; additional cases and helpers |
| `objective/gomoney/policy_test.go` | `objective/gomoney` | 0 | TestCostsRejectNegativeValuesUnlessPolicyAllowsThem; TestEntryConstructionRejectsDuplicatesAndCopiesInput; additional cases and helpers |
| `objective/gomoney/properties_test.go` | `objective/gomoney` | 0 | TestTotalsAndOrderingMatchDirectExactMoneyArithmetic; TestCostsAreSafeForConcurrentSolverUse |
| `objective/money/benchmark_test.go` | `objective/money` | 0 | BenchmarkTotalSixtyFourContainers; BenchmarkComparisonOverhead; additional cases and helpers |
| `objective/money/example_helpers_test.go` | `objective/money` | 0 | public contracts and private invariant source |
| `objective/money/example_test.go` | `objective/money` | 0 | ExampleNew; ExampleNewWithPolicy |
| `objective/money/fuzz_test.go` | `objective/money` | 0 | FuzzCostTypeIdentifiers; FuzzCostMappings; additional cases and helpers |
| `objective/money/gomoney_test.go` | `objective/money` | 0 | TestNewRejectsMoreThanDefaultCostTypes; TestCostObjectiveEntriesRejectNilContext; additional cases and helpers |
| `objective/money/hardening_test.go` | `objective/money` | 0 | TestMain; TestDeterministicRankingAcrossProcessesAndArchitectures; additional cases and helpers |
| `objective/money/moneyobjective.go` | `objective/money` | 0 | public contracts and private invariant source |
| `objective/money/moneyobjective_test.go` | `objective/money` | 0 | TestPublicIdentityAndStableErrors; TestConstructionCopiesInputsAndRetainsContextPrecedence |
| `objective/money/policy_test.go` | `objective/money` | 0 | TestCostsRejectNegativeValuesUnlessPolicyAllowsThem; TestEntryConstructionRejectsDuplicatesAndCopiesInput; additional cases and helpers |
| `objective/money/properties_test.go` | `objective/money` | 0 | TestTotalsAndOrderingMatchDirectExactMoneyArithmetic; TestCostsAreSafeForConcurrentSolverUse |
| `objective/objective.go` | `.` | 3 | public contracts and private invariant source |
| `objective/objective_test.go` | `.` | 0 | TestCallbackPanicDoesNotDiscloseApplicationValue; TestLexicographicPrecedence; additional cases and helpers |
| `plan.go` | `.` | 0 | public contracts and private invariant source |
| `plan_internal_test.go` | `.` | 0 | TestPlanRejectsInvalidStateAndAcceptsEveryDocumentedStatus; TestPlanOwnsAllCollectionState; additional cases and helpers |
| `request.go` | `.` | 1 | public contracts and private invariant source |
| `request_boundaries_test.go` | `.` | 0 | TestPhysicalRequestAcceptsExplicitZeroTare; TestRequestCollectionAndIDLimitsAreInclusive; additional cases and helpers |
| `request_fuzz_test.go` | `.` | 0 | FuzzItemContainerValidation; TestFuzzRegressionContainerBoundsDoNotChangeItemValidity; additional cases and helpers |
| `request_internal_test.go` | `.` | 0 | TestNormalizedRequestCanonicalizesMetadataAndOwnsCopies; TestNewRequestCanonicalizesAndNormalizesOptionalPhysicalLimits; additional cases and helpers |
| `solver/a_exact_constraint_guard_test.go` | `.` | 0 | TestExactSearchHonorsConstraintRejection |
| `solver/aggregate_overflow_test.go` | `.` | 0 | TestSolversRefuseUnrepresentablePlanAggregates; TestExactPackAllPreservesRepresentableAlternatives; additional cases and helpers |
| `solver/aggregate_totals_internal_test.go` | `.` | 0 | TestPlanTotalAdmissionIsExactAndNonMutatingOnRefusal; TestBuildPlanRefusesMalformedInternalAggregates; additional cases and helpers |
| `solver/allocation_admission_internal_test.go` | `.` | 0 | TestExactLatticeScalarAdmission; TestExactCartesianScalarAdmission |
| `solver/benchmark_test.go` | `.` | 0 | BenchmarkHeuristicOrdinaryOrder; BenchmarkExactTinyFixed; additional cases and helpers |
| `solver/boundary_assertions_test.go` | `.` | 0 | TestPlacementOrderingComparesPositiveEndpoints; TestExactPointsHonorsByteBoundaryAndEnumeratesGrid; additional cases and helpers |
| `solver/corpus_test.go` | `.` | 0 | TestDWaveSampleData1 |
| `solver/exact.go` | `.` | 0 | public contracts and private invariant source |
| `solver/exact_grid_memory_contract_test.go` | `.` | 0 | TestExactFixedPointGridMemoryAdmission |
| `solver/exact_helper_state_test.go` | `.` | 0 | TestCheckedProductPreservesExactUnsignedBoundaries; TestExactSearchStopsForEachIndependentTerminalCause |
| `solver/exact_oracle_test.go` | `.` | 0 | TestExactFixedAgreesWithExhaustiveGridOracle; TestExactCenterOfGravityAgreesWithIndependentGridOracle; additional cases and helpers |
| `solver/exact_test.go` | `.` | 0 | TestExactProvesOptimalFixedPacking; TestExactProvesFixedContainersInfeasible; additional cases and helpers |
| `solver/floor_support_boundary_test.go` | `.` | 0 | TestPublicSolversAcceptFloorWithoutItemSupporters |
| `solver/fuzz_test.go` | `.` | 0 | FuzzHeuristic; FuzzFixedContainerPacking; additional cases and helpers |
| `solver/group_transitions_test.go` | `.` | 0 | TestHeuristicSuccessfulGroupContinuesPastIneligibleContainer; TestHeuristicContinuesPastExhaustedType; additional cases and helpers |
| `solver/heuristic.go` | `.` | 0 | public contracts and private invariant source |
| `solver/heuristic_admission_continuation_test.go` | `.` | 0 | TestHeuristicContinuesAfterOverlappingOrientation; TestHeuristicContinuesAfterUnsupportedOrientation; additional cases and helpers |
| `solver/heuristic_overflow_work_test.go` | `.` | 0 | TestHeuristicCountsOrientationsAfterEndpointOverflow |
| `solver/heuristic_rejection_short_circuit_test.go` | `.` | 0 | TestHeuristicRejectedCandidateDoesNotCancelOwner |
| `solver/heuristic_test.go` | `.` | 0 | TestHeuristicPackAllIsDeterministicAndVerified; TestSolversHonorContainerCenterOfGravityBounds; additional cases and helpers |
| `solver/instance_validation_test.go` | `.` | 0 | TestPackFixedClassifiesIndependentInstanceFailures |
| `solver/internal_test.go` | `.` | 0 | TestExactPlacementRejectsEveryFeasibilityBoundary; TestHeuristicPlacementScoringAndCallbackRejection; additional cases and helpers |
| `solver/leak_test.go` | `.` | 0 | TestSolversStopAfterRepeatedCancellation |
| `solver/memory_reservation_boundary_test.go` | `.` | 0 | TestWorkingMemoryReservationAcceptsExactBudget; TestHeuristicPacksAtExactWorkingMemoryReservation |
| `solver/metamorphic_test.go` | `.` | 0 | TestHeuristicIsInvariantToInputPermutation; TestHeuristicPackingScalesWithTheIntegerLattice; additional cases and helpers |
| `solver/orientation_constraint_continuation_test.go` | `.` | 0 | TestFixedPackersContinueAfterRejectedOrientation; TestHeuristicContinuesAfterNonFittingOrientation; additional cases and helpers |
| `solver/physical_boundaries_test.go` | `.` | 0 | TestPlacementOwnersSumSupportAndShareLoadAtBoundaries; TestStackBoundaryIncludesTransitiveLoadAndDepth |
| `solver/placement_boundaries_test.go` | `.` | 0 | TestGravityOriginsRetainInclusiveInteriorIntervals; TestPlacementAcceptsExactGrossCapacity; additional cases and helpers |
| `solver/point_ordering_test.go` | `.` | 0 | TestPointComparisonPreservesLexicographicOrder |
| `solver/rollback_state_test.go` | `.` | 0 | TestGroupRollbackRebuildsRetainedStateAndContinues |
| `solver/search_transitions_test.go` | `.` | 0 | TestHeuristicPreservesCompletePlanAtCandidateAndRepackLimits; TestHeuristicGroupRollbackRestoresFiniteStockForLaterItem; additional cases and helpers |
| `solver/shared_load_test.go` | `.` | 0 | TestPackFixedAcceptsProportionalSharedSupportLoad |
| `solver/variable_budget_transitions_test.go` | `.` | 0 | TestExactVariableSearchConservesItemsAfterObjectiveCancellation; TestExactVariableLowerBoundUsesGrossCapacityBeforeSearch; additional cases and helpers |
| `verify/fuzz_test.go` | `.` | 0 | FuzzSuppliedPlan; FuzzPhysicalConstraints |
| `verify/internal_test.go` | `.` | 0 | TestCheckedAccountingAndCycleDefenses; TestCheckedAccountingInclusiveLimitsAndRejectedState; additional cases and helpers |
| `verify/physical_test.go` | `.` | 0 | TestVerifyAccountsForSuppliedEmptyContainers; TestVerifyRejectsEligibilityGroupingAndUnknownUnpackedItem; additional cases and helpers |
| `verify/verify.go` | `.` | 0 | public contracts and private invariant source |
| `verify/verify_test.go` | `.` | 0 | TestVerifyAcceptsCompleteFeasiblePlan; TestVerifyRejectsOverlapAndAccountingDrift; additional cases and helpers |
| `visualize/svg.go` | `.` | 0 | public contracts and private invariant source |
| `visualize/svg_boundaries_test.go` | `.` | 0 | TestSVGCoordinateAndSceneWidthBoundaries; TestSVGPreservesPlacementGroupsAndOffsets |
| `visualize/svg_test.go` | `.` | 0 | TestSVGConsumesOnlyVerifiedPlansAndEscapesLabels; TestSVGRejectionDoesNotDiscloseVerifierDetails; additional cases and helpers |
| `workflow_test.go` | `.` | 0 | TestWorkflowsPinExternalActionsByCommit; TestSharedToolingContractIsImmutable; additional cases and helpers |

## Investigated contracts and rationale

- Physical construction is not yet request admission. NewItem validates amounts
  and unique orientation identities but sorts the caller's IncompatibleGroups
  backing slice before copying and retains MaxSupportedWeight. Existing copy
  tests cover orientations/attributes, not that retained pointer. Supply owned
  input for sharing; normalized requests do detach load pointers and collections.
  NewContainerType copies optional state; lattice divisibility, volume and
  reserved containment are checked later. Expansion allocates only after count
  admission but malformed instance input can fail after that bounded allocation.
- Normalization converts units exactly, divides rationally and requires integral
  int64 lattice values; no epsilon/rounding is added. Request order is ID-sorted;
  orientation list order is retained. WithLimits is an unchecked policy replacement,
  not re-normalization or count re-admission. Memory uses fixed-size conservative
  estimates, not measured heap/RSS or a guarantee of physical allocation.
- Geometry is half-open with explicit endpoint/volume overflow checks. Positive
  face contact is distinct from intersection; SupportArea false also covers
  unrepresentable area. Zero cuboids are not validated constructor values.
- Plan construction checks status/termination labels and resource estimates,
  clones collections but does not sort, recompute accounting or certify feasible
  geometry/proof. CanonicalString serializes stored order and all metadata;
  equal geometry alone does not imply equal bytes. FieldError renders caller IDs
  and reasons; explicit diagnostics are not a redaction facility. FragileTop
  true forbids supporting another item. Custom score values retain provider
  grammar rather than gaining numeric validation from plan admission.
- Core objectives use checked signed integers but trust supplied statistics.
  Custom whole-plan callbacks can see heuristic partial trials/empty containers.
  Valid executes outside panic conversion and must not panic/block. Output limits
  are applied after callbacks return, not before callback allocation. Borrowed
  providers own synchronization and cancellation; immutable value views do not
  freeze their closure state.
- Heuristic is deterministic extreme-point search with one alternate pass when
  MaxImprovementRounds is positive. Recorded admission probes can exceed a limit
  before rejection. Cancellation is observed at iteration/search and provider
  checkpoints, not by preempting rejected-candidate loops/grid construction.
  Safe interrupted publication can be refused for custom extensions rather than
  replaying them under detached context. Source predicates and default support/
  load behavior are unchanged.
- Exact variable search enumerates all bounded type multisets from a relaxed
  volume/mass lower bound through item-count bins; it does not stop at the first
  feasible count. Shared work is charged across configurations. Exact fixed
  search restores placement/item/accounting slices before sibling recursion.
  Center-of-gravity constraints use full integer grids with scalar/Cartesian
  admission, while ordinary compact points come from existing/reserved faces.
  Tiny oracle tests and hosted 386 admission tests protect their actual named
  boundaries; none is newly claimed global conformance or allocation assurance.
- Internal nonoverlapping bins may sum support rectangles. The untrusted-plan
  verifier uses their union so overlapping supplied faces cannot inflate support.
  Positive-height contact forms descending-Z edges; exact rational shares
  conserve transitive load without duplicating full item weight at each supporter.
  Doubled integer centers preserve half-unit cuboid centers for CoG comparison.
  This remains a cuboid model, not dynamic transport/material-strength simulation.
- Verifier findings and errors are separate. Valid ignores stored Err; PlanContext
  can return cancellation/early-admission errors with no findings and does not
  always copy them into Result.Err. Plan uses background context and discards the
  separate error. Invalid WithConstraints configuration returns Err with no
  violation; SVG and RequireVerified inspect findings only, so custom-option
  callers must preverify/check errors explicitly. Status checks validate metadata
  consistency and a plain work.Solver label, not solver identity or a new proof.
- Schema-v1 encoding rejects unknown fields, exact duplicate decoded names,
  malformed/oversized input, excess nesting/collections and trailing values.
  Typed field lookup retains Go JSON case-insensitive fallback; duplicate aliases
  differing in case are not a separate canonical-name check. Plan decode does not
  certify feasibility. Marshal has no separate output quota; JSON byte form is
  deterministic for the stored order, not a canonicalization of all equivalent input.
- Money /v3 uses Money /v2, one currency and identical fixed Default/Custom context;
  no rescaling/rounding/exchange-rate lookup occurs. Constructor collections are
  owned, costs immutable, default negatives rejected. Total prices every selected
  instance (even empty), accumulating integer coefficients before the final Money
  bound so mixed-sign results are order-independent. Context is checked at entry,
  not inside Total. The deprecated gomoney /v2 facade delegates to the published
  canonical /v2 adapter and keeps Money v1/error identities, rather than converting
  nominal Money types or consuming sibling current /v3 source.
- Reference tests/command isolate their pinned common-integral subset and validate
  supplied outputs; external float models are not authoritative feasibility.
  Command solve timing includes internal solver verification but excludes its
  separate harness verification; no new benchmark/interop claim is introduced.

All 87 assurance files were inspected for entrypoints, partitions, helpers,
assertions and resource ownership, including 322 test functions, 21 fuzz targets,
17 benchmark functions and eight Examples. These counts are inventory, not proof
of test adequacy or newly run assurance. Fuzz/benchmark comments identify bounded
inputs/oracles and timing exclusions; executable bodies and saved seeds stay fixed.
No build/platform variants, generated Go or producers were discovered. Existing
security/analyzer and legacy deprecation guidance remains intact. Dated benchmark,
mutation, corpus, reference pin and compatibility evidence remains immutable.

## Missing versus unresolved

The five initial structural gaps are repaired without adding private-symbol filler.
Existing extensive API prose needed the ownership, admission, ordering, callback,
cancellation and proof-trust clarifications above; absence of checker findings
alone does not establish their semantic accuracy.

Known source limitations remain outside this prose batch:

- NewItem retains an optional load pointer and sorts a caller slice in place.
- A one-item 1x1x1 request in a 3x1x1 fixed bin, with no CoG/reserved regions and
  a custom predicate accepting only candidate X=1, has a lattice-feasible placement
  not enumerated by the initial compact face coordinates. A bounded local
  observation confirmed the exact solver reports infeasible while independent
  verification accepts the explicit X=1 plan. Intended exhaustive
  infeasibility remains the expected policy; this documents a completeness gap,
  not a redefinition that makes the narrower implementation pass.
- Exact PackAll does not enumerate more bins than items. Objectives rewarding
  extra unused containers and arbitrary origin-sensitive objectives/predicates
  need a domain/completeness decision; tiny built-in oracle results cannot settle it.
- Verification error/finding separation and SVG/test-helper finding-only admission
  require explicit caller error checks. No behavior repair or new proof machinery
  is introduced here.
- Objective Valid panics are outside recovery; blocking callbacks and constructor/
  grid/geometry work cannot be interrupted through context alone. Context and
  accounting policy are cooperative admission, not a hard process-memory deadline.

All dependencies, workspace/module catalogs, apidiff baselines, runtime fixtures,
frozen authority pins, dated evidence and desired selected behavior are preserved.
No protocol decision/history register exists in this repository. Coordinator owns
follow-up scope and remote delivery rather than this worker inventing repairs.

## Selected verification outcome

Go 1.27.2 formatting is clean. All 105 files retain identical parsed executable
ASTs and every non-comment token/literal, with no implicit-semicolon difference;
all 24 security/analyzer/deprecation annotations are identical. All 563 comment
groups parse and render through the native Go comment printer. Twelve applicable
package/command `go doc -all` renders passed; the test-only reference root has no
production package to render. No full-prose API snapshot is configured or changed.

Native `golib docs check --module` passed for root, both adapters and reference
integration. Root and both adapter recipes execute their existing Examples only
(four plus two plus two); the reference fallback compiles its two packages with
no Examples selected. Root spelling/links passed with 70 links, no errors and two
configured exclusions; a transient upstream 503 was retried without suppression.
Native `golib specification check` passed with zero applicable decision modules.
Current `golib source-docs check --json` and inventory cover all four modules and
105 files with zero structural findings, without replacing semantic review.

The tiny offline ownership/search observation confirmed both documented caveats;
its explicit feasible plan was independently checked with the same predicate.
No solver, benchmark, active fuzz, race, 386, external comparison, historical
qualification or broader security campaign was run. Existing assurance inputs,
API baselines, manifests, dependencies and pinned history are unchanged.

Final changed-file local destinations are present. Pinned secret scans passed for
full current changed files, added content and the staged diff, with no findings.
Ordinary complete-diff inspection found only useful comments/current prose and
this audit; all runtime signatures, bodies and literals remain unchanged. Known
source limitations above are disclosed rather than repaired or accepted as new
expected behavior. Task-owned Go caches, helpers, copied baseline sources,
rendered docs, native temporary directories and scan output were removed; the
coordinator's reusable tools are retained. Remote delivery/CI remains separate.
