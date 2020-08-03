package main

import (
	stdflag "flag"
	"os"
	"testing"

	"github.com/rogpeppe/go-internal/testscript"
)

var updateScripts = stdflag.Bool("update-scripts", false, "update testdata/*.txt files with actual command output")

func TestScript(t *testing.T) {
	testscript.Run(t, testscript.Params{
		Dir:           "testdata",
		UpdateScripts: *updateScripts,
	})
}

func TestMain(m *testing.M) {
	os.Exit(testscript.RunMain(m, map[string]func() int{
		"avsc2avdl": main1,
	}))
}
