package util

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetWorkspace(t *testing.T) {
	workspace := GetWorkspace("/data")
	assert.NotEmpty(t, workspace)
	err := os.Setenv("TESTSOLAR_WORKSPACE", "/data")
	assert.NoError(t, err)
	workspace = GetWorkspace("")
	assert.NotEmpty(t, workspace)
}

// 测试GetPathAndFileName函数
func TestGetPathAndFileName(t *testing.T) {
	testdata, err := filepath.Abs("../../testdata/")
	assert.NoError(t, err)
	// 测试空路径
	_, _, err = GetPathAndFileName(testdata, "")
	assert.Equal(t, err.Error(), "path is empty")
	// 测试目录路径
	dir, file, err := GetPathAndFileName(testdata, "demo")
	assert.NoError(t, err)
	assert.NotEmpty(t, dir)
	assert.Empty(t, file)
	// 测试文件路径
	dir, file, err = GetPathAndFileName(filepath.Join(testdata, "demo"), "demo_test.go")
	assert.NoError(t, err)
	assert.NotEmpty(t, dir)
	assert.Equal(t, file, "demo_test.go")
}

func TestElementIsInSlice(t *testing.T) {
	testCases := []struct {
		element  string
		elements []string
		want     bool
	}{
		{"a", []string{"a", "b", "c"}, true},
		{"d", []string{"a", "b", "c"}, false},
		{"", []string{"a", "b", "c"}, false},
		{"a", []string{}, false},
		{"", []string{}, false},
	}
	for _, tc := range testCases {
		got := ElementIsInSlice(tc.element, tc.elements)
		assert.Equal(t, got, tc.want)
	}
}
