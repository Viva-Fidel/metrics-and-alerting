package config

import (
	"flag"
	"os"
)

func newFlagSet() *flag.FlagSet {
	return flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
}
