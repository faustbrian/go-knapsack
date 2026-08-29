package knapsack_test

import (
	"bufio"
	"os"
	"regexp"
	"strings"
	"testing"
)

var actionCommit = regexp.MustCompile(`^[0-9a-f]{40}$`)

func TestWorkflowsPinExternalActionsByCommit(t *testing.T) {
	t.Parallel()
	if actions := pinnedWorkflowActions(t); len(actions) == 0 {
		t.Fatal("knapsack workflows contain no external actions")
	}
}

func TestSharedToolingContractIsImmutable(t *testing.T) {
	t.Parallel()
	configuration, err := os.ReadFile(".golib.yaml")
	if err != nil {
		t.Fatal(err)
	}
	workflow, err := os.ReadFile(".github/workflows/ci.yml")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(workflow), "continue-on-error") ||
		!strings.Contains(string(configuration), "tool_version: v1.0.7") ||
		!strings.Contains(string(workflow), "go-library-tools/.github/workflows/library-ci.yml@") {
		t.Fatal("root workflow must use the immutable shared module contract")
	}
}

func TestReleaseDryRunRunsStrictRootContract(t *testing.T) {
	t.Parallel()
	workflow, err := os.ReadFile(".github/workflows/ci.yml")
	if err != nil {
		t.Fatal(err)
	}
	for _, required := range []string{
		"workflow_dispatch:",
		"release_dry_run:",
		"go-library-tools/.github/workflows/library-ci.yml@",
		"tooling_sha:",
		"name: Required",
	} {
		if !strings.Contains(string(workflow), required) {
			t.Fatalf("release-event contract omits %q", required)
		}
	}
}

func pinnedWorkflowActions(t *testing.T) map[string]string {
	t.Helper()
	result := map[string]string{}
	for _, path := range []string{".github/workflows/ci.yml"} {
		file, err := os.Open(path)
		if err != nil {
			t.Fatal(err)
		}
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			value, step := strings.CutPrefix(line, "- uses: ")
			if !step {
				value, step = strings.CutPrefix(line, "uses: ")
			}
			if !step || strings.HasPrefix(value, "./") {
				continue
			}
			reference, version, documented := strings.Cut(value, " # ")
			action, commit, pinned := strings.Cut(reference, "@")
			if action == "" || !pinned || !actionCommit.MatchString(commit) ||
				!documented || version == "" {
				t.Errorf("%s contains an unpinned action: %s", path, value)
				continue
			}
			pin := commit + " # " + version
			if previous, exists := result[action]; exists && previous != pin {
				t.Errorf("%s uses conflicting pins for %s", path, action)
				continue
			}
			result[action] = pin
		}
		if closeErr := file.Close(); closeErr != nil {
			t.Fatal(closeErr)
		}
		if err := scanner.Err(); err != nil {
			t.Fatal(err)
		}
	}

	return result
}
