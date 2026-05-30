package config

import (
	"flag"
	"os"
	"testing"
)

func TestInitRegistersNacosEnableFlag(t *testing.T) {
	oldCommandLine := flag.CommandLine
	oldArgs := os.Args
	t.Cleanup(func() {
		flag.CommandLine = oldCommandLine
		os.Args = oldArgs
	})

	flag.CommandLine = flag.NewFlagSet("test", flag.ContinueOnError)
	os.Args = []string{"test"}

	Init()

	if flag.Lookup("nacos-enable") == nil {
		t.Fatal("expected --nacos-enable flag to be registered")
	}
}
