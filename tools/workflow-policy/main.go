package main

import (
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: workflow-policy [path ...]\n\n"+
			"Checks workflows and composite actions under each path (default: .github/workflows actions).\n"+
			"Exits 1 when any finding is reported, 2 on usage or parse errors.\n")
	}
	flag.Parse()
	paths := flag.Args()
	if len(paths) == 0 {
		paths = []string{".github/workflows", "actions"}
	}

	files, err := collect(paths)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	var total int
	for _, f := range files {
		data, err := os.ReadFile(f)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(2)
		}
		findings, err := Check(filepath.ToSlash(f), data)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(2)
		}
		for _, fd := range findings {
			fmt.Println(fd)
		}
		total += len(findings)
	}
	fmt.Fprintf(os.Stderr, "workflow-policy: %d file(s) checked, %d finding(s)\n", len(files), total)
	if total > 0 {
		os.Exit(1)
	}
}

func collect(paths []string) ([]string, error) {
	var out []string
	for _, p := range paths {
		info, err := os.Stat(p)
		if err != nil {
			return nil, err
		}
		if !info.IsDir() {
			out = append(out, p)
			continue
		}
		err = filepath.WalkDir(p, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				if d.Name() == "testdata" {
					return filepath.SkipDir
				}
				return nil
			}
			if ext := strings.ToLower(filepath.Ext(path)); ext == ".yml" || ext == ".yaml" {
				out = append(out, path)
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
	}
	return out, nil
}
