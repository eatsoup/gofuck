package rules

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCdCorrection(t *testing.T) {
	withTmpDir(t)
	// Create some dirs
	mkDir(t, "docs")
	mkDir(t, filepath.Join("docs", "assets"))
	
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	opts := []string{"cd docs", "cd: ducs: No such file or directory"}
	
	// Test matching
	assertMatch(t, "cd_correction", cmd("cd ducs", "cd: ducs: No such file or directory"), true)
	assertMatch(t, "cd_correction", cmd("cd docs/assels", "cd: docs/assels: No such file or directory"), true)
	assertMatch(t, "cd_correction", cmd("cd docs", ""), false)
	assertMatch(t, "cd_correction", cmd("", ""), false)

	// In Go 'gofuck' implementation we return absolute path upon successful resolution, 
	// or fallback to 'mkdir -p' upon complete failure to resolve.
	
	// Case: successful spellcheck => return absolute path wrapped in quotes
	expectedDocs := `cd "` + filepath.Join(cwd, "docs") + `"`
	assertNewCommand(t, "cd_correction", cmd("cd ducs", opts[1]), expectedDocs)
	
	expectedAssets := `cd "` + filepath.Join(cwd, "docs", "assets") + `"`
	assertNewCommand(t, "cd_correction", cmd("cd docs/assels", opts[1]), expectedAssets)

	// Case: complete resolution failure => mkdir -p
	assertNewCommand(t, "cd_correction", cmd("cd missing_totally", opts[1]), "mkdir -p missing_totally && cd missing_totally")
}
