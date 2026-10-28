package main

import (
	"os"
	"reflect"
	"testing"
)

func load(t *testing.T, name string) []Finding {
	t.Helper()
	data, err := os.ReadFile("testdata/" + name)
	if err != nil {
		t.Fatal(err)
	}
	f, err := Check(name, data)
	if err != nil {
		t.Fatal(err)
	}
	return f
}

func rules(fs []Finding) map[string]int {
	m := map[string]int{}
	for _, f := range fs {
		m[f.Rule]++
	}
	return m
}

func TestGoodWorkflowIsClean(t *testing.T) {
	if f := load(t, "good-workflow.yml"); len(f) != 0 {
		t.Fatalf("expected no findings, got %v", f)
	}
}

func TestBadWorkflowFindings(t *testing.T) {
	got := rules(load(t, "bad-workflow.yml"))
	want := map[string]int{
		RuleDangerTrigger:  1, // pull_request_target
		RuleNoPermissions:  2, // jobs unpinned + reusable (broad declares its own)
		RuleNoTimeout:      1, // unpinned
		RuleUnpinned:       3, // checkout@v4, docker://alpine:3.20, build.yml@main
		RuleVersionComment: 1, // setup-go SHA without comment
		RulePersistCreds:   1, // checkout@v4 without persist-credentials
		RuleBroadPerms:     1, // write-all
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("findings by rule:\n got  %v\n want %v", got, want)
	}
}

func TestBadActionFindings(t *testing.T) {
	got := load(t, "bad-action.yml")
	if len(got) != 1 || got[0].Rule != RuleUnpinned || got[0].Line != 6 {
		t.Fatalf("expected one GP001 on line 6, got %v", got)
	}
}

func TestNonWorkflowYAMLIgnored(t *testing.T) {
	f, err := Check("values.yaml", []byte("replicaCount: 1\nimage:\n  tag: latest\n"))
	if err != nil || len(f) != 0 {
		t.Fatalf("expected no findings and no error, got %v %v", f, err)
	}
}

func TestOnScalarAndSequence(t *testing.T) {
	for _, src := range []string{
		"on: pull_request_target\npermissions: {}\njobs: {}\n",
		"on: [push, workflow_run]\npermissions: {}\njobs: {}\n",
	} {
		f, err := Check("w.yml", []byte(src))
		if err != nil {
			t.Fatal(err)
		}
		if len(f) != 1 || f[0].Rule != RuleDangerTrigger {
			t.Fatalf("%q: expected one GP006, got %v", src, f)
		}
	}
}

func TestPinnedSHAWithVersionComment(t *testing.T) {
	src := "on: push\npermissions: {}\njobs:\n  a:\n    runs-on: x\n    timeout-minutes: 1\n    steps:\n" +
		"      - uses: actions/cache@55cc8345863c7cc4c66a329aec7e433d2d1c52a9 # v6.1.0\n" +
		"      - uses: actions/cache@55cc8345863c7cc4c66a329aec7e433d2d1c52a9 # latest\n"
	f, err := Check("w.yml", []byte(src))
	if err != nil {
		t.Fatal(err)
	}
	if len(f) != 1 || f[0].Rule != RuleVersionComment || f[0].Line != 9 {
		t.Fatalf("expected one GP002 on line 9, got %v", f)
	}
}

func TestInvalidYAML(t *testing.T) {
	if _, err := Check("broken.yml", []byte("jobs: [\n")); err == nil {
		t.Fatal("expected a parse error")
	}
}
