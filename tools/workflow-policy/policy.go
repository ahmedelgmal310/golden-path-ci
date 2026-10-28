// Package main implements workflow-policy, a small linter that enforces the
// golden-path supply-chain rules on GitHub Actions workflows and composite
// actions. actionlint checks syntax and expressions and zizmor checks for
// known-dangerous patterns; this tool checks the house rules that neither of
// them enforces as hard failures.
package main

import (
	"fmt"
	"regexp"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

// Rule identifiers. They are stable so findings can be referenced in docs.
const (
	RuleUnpinned       = "GP001" // remote uses: not pinned to a full commit SHA
	RuleVersionComment = "GP002" // pinned SHA without a "# vX.Y.Z" comment
	RuleNoPermissions  = "GP003" // workflow or job does not declare permissions
	RuleBroadPerms     = "GP004" // write-all / read-all shorthand
	RuleNoTimeout      = "GP005" // job without timeout-minutes
	RuleDangerTrigger  = "GP006" // pull_request_target / workflow_run
	RulePersistCreds   = "GP007" // actions/checkout without persist-credentials: false
)

// Finding is a single policy violation.
type Finding struct {
	File    string
	Line    int
	Rule    string
	Message string
}

func (f Finding) String() string {
	return fmt.Sprintf("%s:%d: %s %s", f.File, f.Line, f.Rule, f.Message)
}

var (
	shaRef     = regexp.MustCompile(`^[^@\s]+@[0-9a-f]{40}$`)
	dockerRef  = regexp.MustCompile(`^docker://[^@\s]+@sha256:[0-9a-f]{64}$`)
	versionTag = regexp.MustCompile(`^#\s*v?\d+(\.\d+)*`)
)

// Check parses one YAML document and returns its findings. Files that are
// neither a workflow (has "jobs") nor an action (has "runs") are ignored.
func Check(file string, data []byte) ([]Finding, error) {
	var doc yaml.Node
	if err := yaml.Unmarshal(data, &doc); err != nil {
		return nil, fmt.Errorf("%s: %w", file, err)
	}
	if len(doc.Content) == 0 {
		return nil, nil
	}
	root := doc.Content[0]
	if root.Kind != yaml.MappingNode {
		return nil, nil
	}
	c := &checker{file: file}
	switch {
	case get(root, "jobs") != nil:
		c.workflow(root)
	case get(root, "runs") != nil:
		c.action(root)
	}
	sort.SliceStable(c.findings, func(i, j int) bool { return c.findings[i].Line < c.findings[j].Line })
	return c.findings, nil
}

type checker struct {
	file     string
	findings []Finding
}

func (c *checker) add(n *yaml.Node, rule, format string, args ...any) {
	c.findings = append(c.findings, Finding{File: c.file, Line: n.Line, Rule: rule, Message: fmt.Sprintf(format, args...)})
}

func (c *checker) workflow(root *yaml.Node) {
	if on := get(root, "on"); on != nil {
		for _, trig := range triggers(on) {
			if trig.Value == "pull_request_target" || trig.Value == "workflow_run" {
				c.add(trig, RuleDangerTrigger, "trigger %q runs untrusted code with a privileged token; use pull_request", trig.Value)
			}
		}
	}

	topPerms := get(root, "permissions")
	if topPerms != nil {
		c.permissions(topPerms)
	}

	jobs := get(root, "jobs")
	if jobs == nil || jobs.Kind != yaml.MappingNode {
		return
	}
	for i := 0; i+1 < len(jobs.Content); i += 2 {
		name, job := jobs.Content[i], jobs.Content[i+1]
		if job.Kind != yaml.MappingNode {
			continue
		}
		jobPerms := get(job, "permissions")
		if jobPerms != nil {
			c.permissions(jobPerms)
		}
		if topPerms == nil && jobPerms == nil {
			c.add(name, RuleNoPermissions, "job %q has no permissions block and the workflow sets no default", name.Value)
		}
		if uses := get(job, "uses"); uses != nil {
			// A job that calls a reusable workflow: timeout is set inside it.
			c.uses(uses)
			continue
		}
		if get(job, "timeout-minutes") == nil {
			c.add(name, RuleNoTimeout, "job %q has no timeout-minutes (the default is 360)", name.Value)
		}
		c.steps(get(job, "steps"))
	}
}

func (c *checker) action(root *yaml.Node) {
	runs := get(root, "runs")
	if runs == nil || runs.Kind != yaml.MappingNode {
		return
	}
	if using := get(runs, "using"); using == nil || using.Value != "composite" {
		return
	}
	c.steps(get(runs, "steps"))
}

func (c *checker) steps(steps *yaml.Node) {
	if steps == nil || steps.Kind != yaml.SequenceNode {
		return
	}
	for _, step := range steps.Content {
		if step.Kind != yaml.MappingNode {
			continue
		}
		uses := get(step, "uses")
		if uses == nil {
			continue
		}
		c.uses(uses)
		if strings.HasPrefix(uses.Value, "actions/checkout@") {
			with := get(step, "with")
			pc := get(with, "persist-credentials")
			if pc == nil || pc.Value != "false" {
				c.add(uses, RulePersistCreds, "actions/checkout must set persist-credentials: false so the token is not left in .git/config")
			}
		}
	}
}

func (c *checker) uses(n *yaml.Node) {
	v := n.Value
	switch {
	case strings.HasPrefix(v, "$/"):
		return // self-repository syntax: resolves to the running commit, a pin by construction
	case strings.HasPrefix(v, "./"):
		return // local action or workflow: versioned with this repository
	case strings.HasPrefix(v, "docker://"):
		if !dockerRef.MatchString(v) {
			c.add(n, RuleUnpinned, "%q must be pinned by digest (docker://image@sha256:...)", v)
		}
		return
	case !shaRef.MatchString(v):
		c.add(n, RuleUnpinned, "%q must be pinned to a full 40-character commit SHA", v)
		return
	}
	if !versionTag.MatchString(strings.TrimSpace(n.LineComment)) {
		c.add(n, RuleVersionComment, "%q is pinned but has no version comment (e.g. \"# v1.2.3\") for reviewers and Dependabot", v)
	}
}

func (c *checker) permissions(n *yaml.Node) {
	if n.Kind == yaml.ScalarNode && (n.Value == "write-all" || n.Value == "read-all") {
		c.add(n, RuleBroadPerms, "permissions: %s grants every scope; list the scopes the job needs", n.Value)
	}
}

// triggers returns the event-name nodes of an "on:" value in any of its forms.
func triggers(on *yaml.Node) []*yaml.Node {
	switch on.Kind {
	case yaml.ScalarNode:
		return []*yaml.Node{on}
	case yaml.SequenceNode:
		return on.Content
	case yaml.MappingNode:
		var out []*yaml.Node
		for i := 0; i < len(on.Content); i += 2 {
			out = append(out, on.Content[i])
		}
		return out
	}
	return nil
}

// get returns the value node for key in a mapping node, or nil.
func get(m *yaml.Node, key string) *yaml.Node {
	if m == nil || m.Kind != yaml.MappingNode {
		return nil
	}
	for i := 0; i+1 < len(m.Content); i += 2 {
		if m.Content[i].Value == key {
			return m.Content[i+1]
		}
	}
	return nil
}
