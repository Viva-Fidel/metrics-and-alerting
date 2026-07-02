package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"golang.org/x/tools/go/packages"
)

func TestGenerateReset(t *testing.T) {
	root := filepath.Join("testdata", "project")
	cfg := &packages.Config{
		Mode: packages.NeedName | packages.NeedFiles | packages.NeedSyntax |
			packages.NeedTypes | packages.NeedTypesInfo | packages.NeedModule,
		Dir: root,
	}

	pkgs, err := packages.Load(cfg, "./...")
	if err != nil {
		t.Fatalf("load packages: %v", err)
	}

	var examplePkg *packages.Package
	for _, pkg := range pkgs {
		if pkg.Name == "example" {
			examplePkg = pkg
			break
		}
	}
	if examplePkg == nil {
		t.Fatal("example package not found")
	}

	structs := findResetStructs(examplePkg)
	if len(structs) != 1 || structs[0].name != "ResetableStruct" {
		t.Fatalf("unexpected structs: %#v", structs)
	}

	outPath := filepath.Join(root, "example", "reset.gen.go")
	_ = os.Remove(outPath)

	if err := generatePackage(examplePkg, structs); err != nil {
		t.Fatalf("generate package: %v", err)
	}
	t.Cleanup(func() { _ = os.Remove(outPath) })

	content, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("read generated file: %v", err)
	}

	generated := string(content)
	for _, want := range []string{
		"func (s *ResetableStruct) Reset()",
		"s.i = 0",
		`s.str = ""`,
		`*s.strP = ""`,
		"s.s = s.s[:0]",
		"clear(s.m)",
		"resetter.Reset()",
	} {
		if !strings.Contains(generated, want) {
			t.Fatalf("generated code does not contain %q:\n%s", want, generated)
		}
	}
}
