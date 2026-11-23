// Copyright 2026 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package pnpmlock

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/google/osv-scalibr/extractor"
	"github.com/google/osv-scalibr/extractor/filesystem/osv"
	"github.com/google/osv-scalibr/inventory"
	"github.com/google/osv-scalibr/purl"
	"github.com/google/osv-scalibr/testing/extracttest"
)

func TestIsV9DependencyPath(t *testing.T) {
	tests := []struct {
		path string
		want bool
	}{
		// v9 keys
		{"'@rollup/rollup-linux-riscv64-gnu@4.53.5'", true},
		{"@rollup/rollup-linux-riscv64-gnu@4.53.5", true},
		{"react-async-hook@4.0.0", true},
		{"lodash.merge@4.6.2", true},
		{"'@babel/core@7.24.0(supports-color@8.1.1)'", true},
		// v5/v6 path keys
		{"/@rollup/rollup-linux-riscv64-gnu@4.53.5", false},
		{"/left-pad@1.3.0", false},
		{"/foo/1.0.0", false},
		{"/@foo/bar/1.0.0", false},
		{"/foo/1.4.0_@types+babel__core@7.1.14", false},
		// registry, git, tarball and file keys
		{"registry.npmjs.org/foo/1.1.0", false},
		{"registry.npmjs.org/@foo/bar/1.1.0", false},
		{"example.com/foo/1.3.0_bar@2.0.0", false},
		{"github.com/kevva/is-positive", false},
		{"github.com/my-org/my-package/267087851ad5fac92a184749c27cd539e2fc862e", false},
		{"git@bitbucket.org+my-org/my-bitbucket-project/6104ae42cd32c3d724036d3964678f197b2c9cdb", false},
		{"@gitlab.my-org.org/api/v4/projects/1/packages/npm/@my-org/my-package/-/@my-org/my-package-3.2.3.tgz", false},
		{"file:projects/package-a.tgz", false},
		{"invalidpath", false},
	}
	for _, tt := range tests {
		if got := isV9DependencyPath(tt.path); got != tt.want {
			t.Errorf("isV9DependencyPath(%q) = %v, want %v", tt.path, got, tt.want)
		}
	}
}

// A v6 header over v9-style keys (left behind by merging branches written by
// different pnpm versions) must keep every package's scope and must not drop
// unscoped v9 keys as invalid paths.
func TestExtractor_Extract_MixedKeyFormats(t *testing.T) {
	path := "testdata/mixed-key-formats.v6.yaml"
	scanInput := extracttest.GenerateScanInputMock(t, extracttest.ScanInputMockConfig{Path: path})
	defer extracttest.CloseTestScanInput(t, scanInput)

	extr, err := New(nil)
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}
	got, err := extr.Extract(t.Context(), &scanInput)
	if err != nil {
		t.Fatalf("Extract() error: %v", err)
	}

	pkg := func(name, version string) *extractor.Package {
		return &extractor.Package{
			Name:       name,
			Version:    version,
			PURLType:   purl.TypeNPM,
			Locations:  []string{path},
			SourceCode: &extractor.SourceCodeIdentifier{},
			Metadata:   osv.DepGroupMetadata{DepGroupVals: []string{}},
		}
	}
	want := inventory.Inventory{Packages: []*extractor.Package{
		pkg("@typescript-eslint/types", "5.57.1"),
		pkg("left-pad", "1.3.0"),
		pkg("@rollup/rollup-linux-x64-gnu", "4.53.5"),
		pkg("react-async-hook", "4.0.0"),
	}}
	if diff := cmp.Diff(want, got, cmpopts.SortSlices(extracttest.PackageCmpLess)); diff != "" {
		t.Errorf("Extract() diff (-want +got):\n%s", diff)
	}
}
