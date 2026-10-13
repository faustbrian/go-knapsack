//go:build ignore

package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"testing"
)

// fixtureApproval represents parser input, not executable solver proof.
func fixtureApproval() approval {
	blocks := map[string]int{}
	for index := 1; index <= 7; index++ {
		blocks[fmt.Sprintf("github.com/faustbrian/go-knapsack/v2/solver/exact.go:%d.1,%d.2", index, index)] = 1
	}
	blocks["github.com/faustbrian/go-knapsack/v2/solver/exact.go:1.1,1.2"] = 2
	return approval{Module: "github.com/faustbrian/go-knapsack/v2", Version: "2.0.2", Toolchain: runtime.Version(), Base: "68ce072943edb33748727e9f2a6f314a26656926", TestDiff: "7fc1332a2ad666edf4951db2506d98519495840bbbf31bcd5638bf40d685cac1", Reason: "reviewed unreachable guards", Expires: "affected input change", Inputs: map[string]string{"solver/exact.go": "fixture"}, Blocks: blocks}
}

// fixtureProfile is deliberately small so malformed evidence tests stay bounded.
func fixtureProfile(record approval) string {
	var result strings.Builder
	result.WriteString("mode: atomic\n")
	keys := make([]string, 0, len(record.Blocks))
	for key := range record.Blocks {
		keys = append(keys, key)
	}
	slices.Sort(keys)
	for _, key := range keys {
		fmt.Fprintf(&result, "%s %d 0\n", key, record.Blocks[key])
	}
	result.WriteString("github.com/faustbrian/go-knapsack/v2/solver/exact.go:10.1,10.2 1 1\n")
	return result.String()
}

func TestStrictExceptionRejectsEveryOtherGapAndMalformedEvidence(t *testing.T) {
	record := fixtureApproval()
	valid := fixtureProfile(record)
	selected := make([]packagePolicy, 0, 2)
	selected = append(selected, packagePolicy{Import: record.Module + "/solver", Directory: "solver"})
	report, err := verify(strings.NewReader(valid), selected, record)
	if err != nil || !strings.Contains(report, "exactly 8 reviewed") || !strings.Contains(report, "not 100%") {
		t.Fatalf("valid selective acceptance: %q %v", report, err)
	}
	// Executing these blocks with this frozen admitted input cohort would
	// disprove the reviewed premise; require review instead of inflating coverage.
	cases := map[string]string{
		"fully executed contradicts frozen premise": strings.ReplaceAll(valid, " 0\n", " 1\n"),
		"empty": "", "wrong mode": strings.Replace(valid, "atomic", "set", 1),
		"missing approved block":       strings.Replace(valid, "github.com/faustbrian/go-knapsack/v2/solver/exact.go:1.1,1.2 2 0\n", "", 1),
		"wrong approved count":         strings.Replace(valid, "1.1,1.2 2 0", "1.1,1.2 1 0", 1),
		"executed unreachable block":   strings.Replace(valid, "1.1,1.2 2 0", "1.1,1.2 2 1", 1),
		"extra gap":                    valid + "github.com/faustbrian/go-knapsack/v2/solver/exact.go:11.1,11.2 1 0\n",
		"unknown source":               valid + "github.com/faustbrian/go-knapsack/v2/solver/unknown.go:1.1,1.2 1 1\n",
		"unknown package":              valid + "unknown/package.go:1.1,1.2 1 1\n",
		"duplicate inconsistent count": valid + "github.com/faustbrian/go-knapsack/v2/solver/exact.go:10.1,10.2 2 1\n",
		"negative count":               valid + "github.com/faustbrian/go-knapsack/v2/solver/exact.go:11.1,11.2 1 -1\n",
		"reversed coordinates":         valid + "github.com/faustbrian/go-knapsack/v2/solver/exact.go:12.3,11.1 1 1\n",
		"bad coordinates":              valid + "github.com/faustbrian/go-knapsack/v2/solver/exact.go:unknown 1 1\n",
		"truncated record":             valid + "github.com/faustbrian/go-knapsack/v2/solver/exact.go:11.1,11.2 1\n",
	}
	for name, profile := range cases {
		t.Run(name, func(t *testing.T) {
			if _, err := verify(strings.NewReader(profile), selected, record); err == nil {
				t.Fatal("invalid coverage accepted")
			}
		})
	}
	selected = append(selected, packagePolicy{Import: record.Module + "/verify", Directory: "verify"})
	if _, err := verify(strings.NewReader(valid), selected, record); err == nil {
		t.Fatal("missing production package accepted")
	}
	// Repeated instrumentation merges actual coverage, not a new denominator.
	if _, err := verify(strings.NewReader(valid+"github.com/faustbrian/go-knapsack/v2/solver/exact.go:10.1,10.2 1 0\n"), selected[:1], record); err != nil {
		t.Fatalf("consistent duplicated module-test record refused: %v", err)
	}
}

func TestMaterialBindingExpiresOnSourceAdmissionDependencyAndPolicyChanges(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "solver"), 0o700); err != nil {
		t.Fatal(err)
	}
	write := func(name, contents string) {
		t.Helper()
		if err := os.MkdirAll(filepath.Dir(filepath.Join(root, name)), 0o700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(root, name), []byte(contents), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	policy := packagePolicy{Module: ".", Directory: "solver", Import: "github.com/faustbrian/go-knapsack/v2/solver", Production: true, Required: true}
	packages := struct{ Packages []packagePolicy }{[]packagePolicy{policy}}
	data, err := json.Marshal(packages)
	if err != nil {
		t.Fatal(err)
	}
	write("packages.json", string(data))
	module := fmt.Sprintf(`{"modules":[{"directory":".","module_path":"github.com/faustbrian/go-knapsack/v2","version":"2.0.2","gates":{"coverage":true,"tests":true},"packages":%s}]}`, strings.TrimSuffix(strings.TrimPrefix(string(data), `{"Packages":`), "}"))
	write("modules.json", module)
	write("solver/exact.go", "package solver\n")
	write("request.go", "package knapsack\n")
	write("plan.go", "package knapsack\n")
	write("objective/money/moneyobjective.go", "package moneyobjective\n")
	write("package-lock.json", "pinned documentation dependency\n")
	write("integration/boxpacker/composer.lock", "pinned reference dependency\n")
	write("go.mod", "module github.com/faustbrian/go-knapsack/v2\n")
	write(".golib.yaml", "pinned policy\n")
	record := fixtureApproval()
	record.Inputs, err = inputs(root)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := binding(root, record); err != nil {
		t.Fatalf("unchanged binding refused: %v", err)
	}
	for _, name := range []string{"solver/exact.go", "request.go", "plan.go", "objective/money/moneyobjective.go", "go.mod", "package-lock.json", "integration/boxpacker/composer.lock", ".golib.yaml"} {
		t.Run(name, func(t *testing.T) {
			original, err := os.ReadFile(filepath.Join(root, name))
			if err != nil {
				t.Fatal(err)
			}
			write(name, string(original)+"changed\n")
			if _, err := binding(root, record); err == nil {
				t.Fatal("changed material input accepted")
			}
			write(name, string(original))
		})
	}
	write("new.go", "package knapsack\n")
	if _, err := binding(root, record); err == nil {
		t.Fatal("new material source accepted")
	}
	if err := os.Remove(filepath.Join(root, "new.go")); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(filepath.Join(root, "request.go")); err != nil {
		t.Fatal(err)
	}
	if _, err := binding(root, record); err == nil {
		t.Fatal("removed admission source accepted")
	}
}
