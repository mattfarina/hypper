/*
Copyright The Helm Authors, SUSE

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package action

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chartutil"

	"github.com/rancher-sandbox/hypper/internal/test"
)

func TestList(t *testing.T) {
	for _, tcase := range []struct {
		chart  string
		golden string
	}{
		{
			chart:  "testdata/charts/chart-with-compressed-dependencies",
			golden: "output/list-compressed-deps.txt",
		},
		{
			chart:  "testdata/charts/chart-with-compressed-dependencies-2.1.8.tgz",
			golden: "output/list-compressed-deps-tgz.txt",
		},
		{
			chart:  "testdata/charts/chart-with-uncompressed-dependencies",
			golden: "output/list-uncompressed-deps.txt",
		},
		{
			chart:  "testdata/charts/chart-with-uncompressed-dependencies-2.1.8.tgz",
			golden: "output/list-uncompressed-deps-tgz.txt",
		},
		{
			chart:  "testdata/charts/chart-missing-deps",
			golden: "output/list-missing-deps.txt",
		},
	} {
		buf := bytes.Buffer{}
		if err := NewSharedDependency().List(tcase.chart, &buf); err != nil {
			t.Fatal(err)
		}
		test.AssertGoldenBytes(t, buf.Bytes(), tcase.golden)
	}
}

// TestsharedDependencyStatus_Dashes is a regression test to make sure that dashes in
// chart names do not cause resolution problems.
func TestSharedDependencyStatus_Dashes(t *testing.T) {
	// Make a temp dir
	dir, err := ioutil.TempDir("", "helmtest-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	chartpath := filepath.Join(dir, "charts")
	if err := os.MkdirAll(chartpath, 0700); err != nil {
		t.Fatal(err)
	}

	// Add some fake charts
	first := buildChart(withName("first-chart"))
	_, err = chartutil.Save(first, chartpath)
	if err != nil {
		t.Fatal(err)
	}

	second := buildChart(withName("first-chart-second-chart"))
	_, err = chartutil.Save(second, chartpath)
	if err != nil {
		t.Fatal(err)
	}

	dep := &chart.Dependency{
		Name:    "first-chart",
		Version: "0.1.0",
	}

	// Now try to get the deps
	stat := NewSharedDependency().SharedDependencyStatus(dir, dep, first)
	if stat != "ok" {
		t.Errorf("Unexpected status: %q", stat)
	}
}