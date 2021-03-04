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
	// TODO try to find the shared dep, and see if it resolves: installable, not found, installed, etc.
	return "listed"
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
