package loader

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadTestCase(t *testing.T) {
	// test static loading testcase in directory
	absPath, err := filepath.Abs("../../testdata/")
	assert.NoError(t, err)
	testcases, err := LoadTestCase(absPath, "demo")
	assert.NoError(t, err)
	assert.NotEqual(t, len(testcases), 0)
	// test static loading testcase in file
	testcases, err = LoadTestCase(absPath, "demo/demo_test.go")
	assert.NoError(t, err)
	assert.NotEqual(t, len(testcases), 0)
}
