/*
Copyright The Helm Authors, SUSE.

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
	"fmt"
	"io"

	"github.com/gosuri/uitable"

	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
)

// SharedDependency is the action for building a given chart's shared dependency tree.
//
// It provides the implementation of 'hypper shared-dependency' and its respective subcommands.
type SharedDependency struct {
	*action.Dependency
}

// NewSharedDependency creates a new SharedDependency object with the given configuration.
func NewSharedDependency() *SharedDependency {
	return &SharedDependency{
		action.NewDependency(),
	}
}

// List executes 'helm dependency list'.
func (d *SharedDependency) List(chartpath string, out io.Writer) error {
	c, err := loader.Load(chartpath)
	if err != nil {
		return err
	}

	if val, ok := c.Metadata.Annotations["hypper.cattle.io/shared-dependencies"]; !ok {
		fmt.Fprintf(out, "WARNING: no shared dependencies at %s\n", filepath.Join(chartpath, "charts"))
		return nil
	}

	d.printSharedDependencies(chartpath, out, c)
	fmt.Fprintln(out)
	return nil
}

// dependecyStatus returns a string describing the status of a dependency viz a viz the parent chart.
func (d *SharedDependency) SharedDependencyStatus(chartpath string, dep *chart.Dependency, parent *chart.Chart) string {
	filename := fmt.Sprintf("%s-%s.tgz", dep.Name, "*")

	// If a chart is unpacked, this will check the unpacked chart's `charts/` directory for tarballs.
	// Technically, this is COMPLETELY unnecessary, and should be removed in Helm 4. It is here
	// to preserved backward compatibility. In Helm 2/3, there is a "difference" between
	// the tgz version (which outputs "ok" if it unpacks) and the loaded version (which outouts
	// "unpacked"). Early in Helm 2's history, this would have made a difference. But it no
	// longer does. However, since this code shipped with Helm 3, the output must remain stable
	// until Helm 4.
	switch archives, err := filepath.Glob(filepath.Join(chartpath, "charts", filename)); {
	case err != nil:
		return "bad pattern"
	case len(archives) > 1:
		// See if the second part is a SemVer
		found := []string{}
		for _, arc := range archives {
			// we need to trip the prefix dirs and the extension off.
			filename = strings.TrimSuffix(filepath.Base(arc), ".tgz")
			maybeVersion := strings.TrimPrefix(filename, fmt.Sprintf("%s-", dep.Name))

			if _, err := semver.StrictNewVersion(maybeVersion); err == nil {
				// If the version parsed without an error, it is possibly a valid
				// version.
				found = append(found, arc)
			}
		}

		if l := len(found); l == 1 {
			// If we get here, we do the same thing as in len(archives) == 1.
			if r := statArchiveForStatus(found[0], dep); r != "" {
				return r
			}

			// Fall through and look for directories
		} else if l > 1 {
			return "too many matches"
		}

		// The sanest thing to do here is to fall through and see if we have any directory
		// matches.

	case len(archives) == 1:
		archive := archives[0]
		if r := statArchiveForStatus(archive, dep); r != "" {
			return r
		}

	}
	// End unnecessary code.

	var depChart *chart.Chart
	for _, item := range parent.Dependencies() {
		if item.Name() == dep.Name {
			depChart = item
		}
	}

	if depChart == nil {
		return "missing"
	}

	if depChart.Metadata.Version != dep.Version {
		constraint, err := semver.NewConstraint(dep.Version)
		if err != nil {
			return "invalid version"
		}

		v, err := semver.NewVersion(depChart.Metadata.Version)
		if err != nil {
			return "invalid version"
		}

		if !constraint.Check(v) {
			return "wrong version"
		}
	}

	return "unpacked"
}

}

// printSharedDependencies prints all of the shared dependencies in the yaml file.
func (d *SharedDependency) printSharedDependencies(chartpath string, out io.Writer, c *chart.Chart) {
	table := uitable.New()
	table.MaxColWidth = 80
	table.AddRow("NAME", "VERSION", "REPOSITORY", "STATUS")
	for _, row := range c.Metadata.Dependencies {
		table.AddRow(row.Name, row.Version, row.Repository, d.SharedDependencyStatus(chartpath, row, c))
	}
	fmt.Fprintln(out, table)
}
