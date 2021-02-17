package main

import (
	"bytes"
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/Masterminds/log-go"
	"github.com/mattfarina/hypper/internal/test"
	"helm.sh/helm/v3/pkg/chartutil"
	kubefake "helm.sh/helm/v3/pkg/kube/fake"
	"helm.sh/helm/v3/pkg/release"
	"helm.sh/helm/v3/pkg/storage"
	"helm.sh/helm/v3/pkg/storage/driver"
	"helm.sh/helm/v3/pkg/time"

	logcli "github.com/Masterminds/log-go/impl/cli"
	"github.com/mattfarina/hypper/pkg/cli"
	"github.com/mattn/go-shellwords"
	"github.com/spf13/cobra"
	helmAction "helm.sh/helm/v3/pkg/action"
)

func testTimestamper() time.Time { return time.Unix(242085845, 0).UTC() }

func init() {
	helmAction.Timestamper = testTimestamper
}

func runTestCmd(t *testing.T, tests []cmdTestCase) {
	t.Helper()
	for _, tt := range tests {
		for i := 0; i <= tt.repeat; i++ {
			t.Run(tt.name, func(t *testing.T) {
				defer resetEnv()()

				storage := storageFixture()
				for _, rel := range tt.rels {
					if err := storage.Create(rel); err != nil {
						t.Fatal(err)
					}
				}
				t.Logf("running cmd (attempt %d): %s", i+1, tt.cmd)
				_, out, err := executeActionCommandC(storage, tt.cmd)
				if (err != nil) != tt.wantError {
					t.Errorf("expected error, got '%v'", err)
				}
				if tt.golden != "" {
					test.AssertGoldenString(t, out, tt.golden)
				}
			})
		}
	}
}

func runTestActionCmd(t *testing.T, tests []cmdTestCase) {
	t.Helper()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer resetEnv()()

			store := storageFixture()
			for _, rel := range tt.rels {
				if err := store.Create(rel); err != nil {
					t.Fatal(err)
				}
			}
			_, out, err := executeActionCommandC(store, tt.cmd)
			if (err != nil) != tt.wantError {
				t.Errorf("expected error, got '%v'", err)
			}
			if tt.golden != "" {
				test.AssertGoldenString(t, out, tt.golden)
			}
		})
	}
}

func storageFixture() *storage.Storage {
	return storage.Init(driver.NewMemory())
}

func executeActionCommandC(store *storage.Storage, cmd string) (*cobra.Command, string, error) {
	return executeActionCommandStdinC(store, nil, cmd)
}

func executeActionCommandStdinC(store *storage.Storage, in *os.File, cmd string) (*cobra.Command, string, error) {
	args, err := shellwords.Parse(cmd)
	if err != nil {
		return nil, "", err
	}

	actionConfig := &helmAction.Configuration{
		Releases:     store,
		KubeClient:   &kubefake.PrintingKubeClient{Out: ioutil.Discard},
		Capabilities: chartutil.DefaultCapabilities,
		Log:          func(format string, v ...interface{}) {},
	}

	// create our own Logger that satisfies impl/cli.Logger, but with a buffer for tests
	buf := new(bytes.Buffer)
	logger := logcli.NewStandard()
	logger.InfoOut = buf
	logger.ErrorOut = buf
	log.Current = logger

	root, err := newRootCmd(actionConfig, log.Current, args)
	if err != nil {
		return nil, "", err
	}

	root.SetOut(logger.InfoOut)
	root.SetErr(logger.ErrorOut)
	root.SetArgs(args)

	oldStdin := os.Stdin
	if in != nil {
		root.SetIn(in)
		os.Stdin = in
	}

	if mem, ok := store.Driver.(*driver.Memory); ok {
		mem.SetNamespace(settings.Namespace())
	}
	c, err := root.ExecuteC()
	result := buf.String()
	os.Stdin = oldStdin

	return c, result, err
}

// cmdTestCase describes a test case that works with releases.
type cmdTestCase struct {
	name      string
	cmd       string
	golden    string
	wantError bool
	// Rels are the available releases at the start of the test.
	rels []*release.Release
	// Number of repeats (in case a feature was previously flaky and the test checks
	// it's now stably producing identical results). 0 means test is run exactly once.
	// TODO(itxaka): Disabled for now, we are not using it but we may want to keep it 1:1 with helm?
	repeat int
}

func resetEnv() func() {
	origEnv := os.Environ()
	return func() {
		os.Clearenv()
		for _, pair := range origEnv {
			kv := strings.SplitN(pair, "=", 2)
			os.Setenv(kv[0], kv[1])
		}
		settings = cli.New()
	}
}

func executeCommandStdinC(cmd string) (*cobra.Command, string, error) {

	args, err := shellwords.Parse(cmd)
	actionConfig := new(helmAction.Configuration)

	if err != nil {
		return nil, "", err
	}

	// create our own Logger that satisfies impl/cli.Logger, but with a buffer for tests
	buf := new(bytes.Buffer)
	logger := logcli.NewStandard()
	logger.InfoOut = buf
	logger.ErrorOut = buf
	log.Current = logger

	root, err := newRootCmd(actionConfig, log.Current, args)
	if err != nil {
		return nil, "", err
	}

	root.SetOut(logger.InfoOut)
	root.SetErr(logger.ErrorOut)
	root.SetArgs(args)

	oldStdin := os.Stdin

	c, err := root.ExecuteC()
	result := buf.String()
	os.Stdin = oldStdin

	return c, result, err
}
