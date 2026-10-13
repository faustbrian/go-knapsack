//go:build ignore

// Package main enforces the source-bound Knapsack v2.0.2 coverage exception.
// It is excluded from production inventory; use scripts/check-coverage.sh.
package main

import (
	"bufio"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"slices"
	"strconv"
	"strings"
)

// approval is the single reviewed exception, not a coverage threshold.
type approval struct {
	Module    string            `json:"module"`
	Version   string            `json:"version"`
	Toolchain string            `json:"toolchain"`
	Base      string            `json:"base"`
	TestDiff  string            `json:"test_diff"`
	Reason    string            `json:"reason"`
	Expires   string            `json:"expires"`
	Inputs    map[string]string `json:"inputs"`
	Blocks    map[string]int    `json:"blocks"`
}

// packagePolicy retains the authoritative root coverage selection.
type packagePolicy struct {
	Module     string `json:"module_directory"`
	Directory  string `json:"directory"`
	Import     string `json:"import_path"`
	Production bool   `json:"production"`
	Required   bool   `json:"coverage_required"`
}

// material selects source, admitted collaborators and existing verification policy.
func material(name string) bool {
	if strings.HasSuffix(name, ".go") || filepath.Base(name) == "go.mod" || filepath.Base(name) == "go.sum" || filepath.Base(name) == "go.work" || filepath.Base(name) == "go.work.sum" || filepath.Base(name) == "composer.json" || filepath.Base(name) == "composer.lock" || filepath.Base(name) == "package.json" || filepath.Base(name) == "package-lock.json" {
		return true
	}
	switch name {
	case ".go-version", ".golib.yaml", ".golangci.yml", "modules.json", "packages.json":
		return true
	}
	return name == "vendor/modules.txt" || strings.HasPrefix(name, "scripts/") || strings.HasPrefix(name, ".github/workflows/") || name == "verification/package.mk" || strings.HasSuffix(name, ".tsv") && strings.HasPrefix(name, "verification/")
}

// inputs includes new/removal detection, not only hashes of previously known files.
func inputs(root string) (map[string]string, error) {
	owner, err := os.OpenRoot(root)
	if err != nil {
		return nil, err
	}
	defer owner.Close()
	result := map[string]string{}
	err = filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		name, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		name = filepath.ToSlash(name)
		if entry.IsDir() {
			// Composer installs the pinned PHP oracle, not Go dependency source.
			// A Go vendor tree remains material through its modules.txt and Go files.
			if entry.Name() == "vendor" {
				if _, err := owner.Stat(filepath.Join(name, "modules.txt")); errors.Is(err, os.ErrNotExist) {
					return filepath.SkipDir
				} else if err != nil {
					return err
				}
			}
			if name != "." && (strings.HasPrefix(entry.Name(), ".") || entry.Name() == "node_modules") {
				return filepath.SkipDir
			}
			return nil
		}
		// Workflow policy lives under an otherwise hidden directory.
		if !material(name) {
			return nil
		}
		if !entry.Type().IsRegular() {
			return fmt.Errorf("non-regular bound input %s", name)
		}
		data, err := owner.ReadFile(name)
		if err != nil {
			return err
		}
		result[name] = fmt.Sprintf("%x", sha256.Sum256(data))
		return nil
	})
	if err != nil {
		return nil, err
	}
	// Walk hidden workflow policy separately; never follow symlinks.
	if info, err := os.Lstat(filepath.Join(root, ".github")); err == nil && !info.IsDir() {
		return nil, errors.New("non-directory workflow parent")
	}
	workflows := filepath.Join(root, ".github", "workflows")
	if info, err := os.Lstat(workflows); err == nil {
		if !info.IsDir() {
			return nil, errors.New("workflow policy is not a directory")
		}
		err = filepath.WalkDir(workflows, func(path string, entry fs.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			if entry.IsDir() {
				return nil
			}
			if !entry.Type().IsRegular() {
				return errors.New("non-regular workflow policy")
			}
			name, err := filepath.Rel(root, path)
			if err != nil {
				return err
			}
			data, err := owner.ReadFile(name)
			if err != nil {
				return err
			}
			result[filepath.ToSlash(name)] = fmt.Sprintf("%x", sha256.Sum256(data))
			return nil
		})
		if err != nil {
			return nil, err
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}
	return result, nil
}

// binding expires acceptance on any material source or policy change.
func binding(root string, record approval) ([]packagePolicy, error) {
	if record.Module != "github.com/faustbrian/go-knapsack/v2" || record.Version != "2.0.2" || record.Toolchain != runtime.Version() || record.Base != "68ce072943edb33748727e9f2a6f314a26656926" || record.TestDiff != "7fc1332a2ad666edf4951db2506d98519495840bbbf31bcd5638bf40d685cac1" || len(record.Blocks) != 7 || record.Reason == "" || record.Expires == "" {
		return nil, errors.New("unsupported release exception")
	}
	actual, err := inputs(root)
	if err != nil {
		return nil, err
	}
	if len(actual) != len(record.Inputs) {
		return nil, errors.New("exception expired: material input inventory changed")
	}
	for name, digest := range actual {
		if record.Inputs[name] != digest {
			return nil, fmt.Errorf("exception expired: material input changed: %s", name)
		}
	}
	var modules struct {
		Modules []struct {
			Directory, Version string
			Path               string `json:"module_path"`
			Packages           []packagePolicy
			Gates              map[string]bool
		}
	}
	var catalog struct{ Packages []packagePolicy }
	for _, entry := range []struct {
		Name  string
		Value any
	}{{"modules.json", &modules}, {"packages.json", &catalog}} {
		data, err := os.ReadFile(filepath.Join(root, entry.Name))
		if err != nil {
			return nil, err
		}
		if err = json.Unmarshal(data, entry.Value); err != nil {
			return nil, err
		}
	}
	var selected []packagePolicy
	for _, module := range modules.Modules {
		if module.Directory == "." {
			if module.Path != record.Module || module.Version != record.Version || !module.Gates["coverage"] || !module.Gates["tests"] {
				return nil, errors.New("release or mandatory gate identity changed")
			}
			for _, policy := range module.Packages {
				if policy.Required {
					if !policy.Production || policy.Module != "." {
						return nil, errors.New("invalid production coverage catalog")
					}
					selected = append(selected, policy)
				}
			}
		}
	}
	if len(selected) == 0 {
		return nil, errors.New("missing root production coverage catalog")
	}
	var other []packagePolicy
	for _, policy := range catalog.Packages {
		if policy.Module == "." && policy.Required {
			other = append(other, policy)
		}
	}
	sortPolicy := func(a, b packagePolicy) int { return strings.Compare(a.Import, b.Import) }
	slices.SortFunc(selected, sortPolicy)
	slices.SortFunc(other, sortPolicy)
	if !slices.Equal(selected, other) {
		return nil, errors.New("production catalogs disagree")
	}
	for index, policy := range selected {
		if policy.Import == "" || index > 0 && policy.Import == selected[index-1].Import {
			return nil, errors.New("duplicate or missing production package")
		}
	}
	return selected, nil
}

// profileBlock merges repeated package-test instrumentation without changing counts.
type profileBlock struct {
	Statements int
	Covered    bool
}

// verify rejects every gap other than the exact reviewed unreachable blocks.
func verify(profile io.Reader, selected []packagePolicy, record approval) (string, error) {
	expected := map[string]string{}
	for _, policy := range selected {
		expected[policy.Import] = policy.Directory
	}
	blocks := map[string]profileBlock{}
	limited := &io.LimitedReader{R: profile, N: (32 << 20) + 1}
	scanner := bufio.NewScanner(limited)
	scanner.Buffer(make([]byte, 4096), 64<<10)
	if !scanner.Scan() || scanner.Text() != "mode: atomic" {
		return "", errors.New("missing or unsupported atomic coverage profile")
	}
	coordinates := regexp.MustCompile(`^[1-9][0-9]*\.[1-9][0-9]*,[1-9][0-9]*\.[1-9][0-9]*$`)
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) != 3 {
			return "", errors.New("malformed coverage block")
		}
		colon := strings.LastIndexByte(fields[0], ':')
		if colon < 1 || !coordinates.MatchString(fields[0][colon+1:]) {
			return "", errors.New("malformed source coordinates")
		}
		parts := strings.FieldsFunc(fields[0][colon+1:], func(r rune) bool { return r == '.' || r == ',' })
		values := make([]int, 0, 4)
		for _, part := range parts {
			value, err := strconv.Atoi(part)
			if err != nil {
				return "", errors.New("source coordinate overflow")
			}
			values = append(values, value)
		}
		if values[2] < values[0] || values[2] == values[0] && values[3] < values[1] {
			return "", errors.New("reversed source coordinates")
		}
		file := fields[0][:colon]
		directory := filepath.ToSlash(filepath.Dir(file))
		sourceDirectory, knownPackage := expected[directory]
		_, knownFile := record.Inputs[filepath.ToSlash(filepath.Join(sourceDirectory, filepath.Base(file)))]
		if !knownPackage || !knownFile || !strings.HasSuffix(file, ".go") || strings.HasSuffix(file, "_test.go") || strings.ContainsAny(file, `\`) {
			return "", errors.New("profile contains unknown production source")
		}
		statements, err := strconv.Atoi(fields[1])
		if err != nil || statements < 0 {
			return "", errors.New("invalid statement count")
		}
		count, err := strconv.ParseUint(fields[2], 10, 64)
		if err != nil {
			return "", errors.New("invalid execution count")
		}
		old, present := blocks[fields[0]]
		if present && old.Statements != statements {
			return "", errors.New("inconsistent duplicate block")
		}
		blocks[fields[0]] = profileBlock{statements, old.Covered || count > 0}
	}
	if err := scanner.Err(); err != nil {
		return "", err
	}
	if limited.N <= 1 {
		return "", errors.New("coverage profile exceeds byte bound")
	}
	seen := map[string]bool{}
	covered, total, excepted := 0, 0, 0
	for identity, block := range blocks {
		if block.Statements > int(^uint(0)>>1)-total {
			return "", errors.New("statement count overflow")
		}
		total += block.Statements
		if block.Statements > 0 {
			seen[filepath.ToSlash(filepath.Dir(strings.Split(identity, ":")[0]))] = true
		}
		if approved, ok := record.Blocks[identity]; ok {
			if block.Statements != approved || block.Covered {
				return "", errors.New("approved unreachable block changed count or was executed")
			}
			excepted += approved
		} else if block.Statements > 0 && !block.Covered {
			return "", fmt.Errorf("unapproved uncovered block: %s", identity)
		}
		if block.Covered {
			covered += block.Statements
		}
	}
	for identity := range record.Blocks {
		if _, ok := blocks[identity]; !ok {
			return "", errors.New("approved block absent from profile")
		}
	}
	for name := range expected {
		if !seen[name] {
			return "", fmt.Errorf("missing executable coverage: %s", name)
		}
	}
	if excepted != 8 {
		return "", errors.New("exception is not the approved eight statements")
	}
	return fmt.Sprintf("%d/%d statements executed; exactly 8 reviewed unreachable statements excepted for root v2.0.2; every other production statement covered (not 100%% execution coverage)\n", covered, total), nil
}

// run keeps profile paths caller-owned and reports only bounded source metadata.
func run(arguments []string, output io.Writer) error {
	flags := flag.NewFlagSet("coverage-exception", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	profile := flags.String("profile", "", "owned coverage profile")
	targets := flags.Bool("targets", false, "canonical production package set")
	if err := flags.Parse(arguments); err != nil {
		return err
	}
	if flags.NArg() != 0 || *targets == (*profile != "") {
		return errors.New("select --targets or --profile")
	}
	file, err := os.Open("verification/coverage-exception.json")
	if err != nil {
		return err
	}
	defer file.Close()
	decoder := json.NewDecoder(io.LimitReader(file, 1<<20))
	decoder.DisallowUnknownFields()
	var record approval
	if err := decoder.Decode(&record); err != nil {
		return err
	}
	var trailing any
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		return errors.New("trailing approval data")
	}
	selected, err := binding(".", record)
	if err != nil {
		return err
	}
	if *targets {
		names := make([]string, 0, len(selected))
		for _, policy := range selected {
			names = append(names, policy.Import)
		}
		_, err = fmt.Fprintln(output, strings.Join(names, ","))
		return err
	}
	opened, err := os.Open(*profile)
	if err != nil {
		return err
	}
	defer opened.Close()
	report, err := verify(opened, selected, record)
	if err != nil {
		return err
	}
	_, err = io.WriteString(output, report)
	return err
}

// main is a release-specific checker, not a public module command.
func main() {
	if err := run(os.Args[1:], os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, "coverage exception refused:", err)
		os.Exit(1)
	}
}
