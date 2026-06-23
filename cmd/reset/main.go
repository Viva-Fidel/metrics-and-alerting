package main

import (
	"fmt"
	"log"
	"os"
	"strings"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	root, err := findModuleRoot()
	if err != nil {
		return err
	}

	pkgs, err := loadPackages(root)
	if err != nil {
		return err
	}

	var firstErr error
	for _, pkg := range pkgs {
		if len(pkg.Errors) > 0 {
			for _, pkgErr := range pkg.Errors {
				log.Printf("skip package %s: %v", pkg.PkgPath, pkgErr)
			}
			continue
		}

		if strings.Contains(pkg.PkgPath, "vendor") {
			continue
		}

		structs := findResetStructs(pkg)
		if len(structs) == 0 {
			continue
		}

		if err := generatePackage(pkg, structs); err != nil {
			log.Printf("generate package %s: %v", pkg.PkgPath, err)
			if firstErr == nil {
				firstErr = err
			}
		}
	}

	return firstErr
}
