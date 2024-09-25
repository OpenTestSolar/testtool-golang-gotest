package execute

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/OpenTestSolar/testtool-golang-gotest/pkg/testcase"

	sdkModel "github.com/OpenTestSolar/testtool-sdk-golang/model"
	"github.com/stretchr/testify/assert"
)

var (
	reportRunningCount  = 0
	reportFinishedCount = 0
	reportSucceedCount  = 0
	reportFailedCount   = 0
)

func TestNewExecuteOptions(t *testing.T) {
	o := NewExecuteOptions()
	assert.NotNil(t, o)
}

func TestNewCmdExecute(t *testing.T) {
	cmd := NewCmdExecute()
	assert.NotNil(t, cmd)
}

func TestParseTestcases(t *testing.T) {
	testSelectors := []string{"path?name=test%20name&attr1=value%3D1"}
	testcases, err := parseTestcases(testSelectors)
	assert.NoError(t, err)
	assert.Len(t, testcases, 1)
}

func TestGroupTestCasesByPathAndName(t *testing.T) {
	projPath := "../../testdata"
	testcases := []*testcase.TestCase{
		{
			Path:       "demo",
			Name:       "TestAdd",
			Attributes: map[string]string{},
		},
		{
			Path:       "demo",
			Name:       "TestAdd1",
			Attributes: map[string]string{},
		},
		{
			Path:       "demo",
			Name:       "TestAdd2",
			Attributes: map[string]string{},
		},
		{
			Path:       "demo",
			Name:       "TestSlowFunc",
			Attributes: map[string]string{},
		},
		{
			Path:       "demo",
			Name:       "TestPanic",
			Attributes: map[string]string{},
		},
		{
			Path:       "demo/build",
			Name:       "TestBuildFailed",
			Attributes: map[string]string{},
		},
	}
	result, err := groupTestCasesByPathAndName(projPath, testcases)
	assert.NoError(t, err)
	assert.Equal(t, result, map[string]map[string][]*testcase.TestCase{
		"demo": {
			"": {
				{
					Path:       "demo",
					Name:       "TestAdd",
					Attributes: map[string]string{},
				},
				{
					Path:       "demo",
					Name:       "TestAdd1",
					Attributes: map[string]string{},
				},
				{
					Path:       "demo",
					Name:       "TestAdd2",
					Attributes: map[string]string{},
				},
				{
					Path:       "demo",
					Name:       "TestSlowFunc",
					Attributes: map[string]string{},
				},
				{
					Path:       "demo",
					Name:       "TestPanic",
					Attributes: map[string]string{},
				},
			},
		},
		"demo/build": {
			"": {
				{
					Path:       "demo/build",
					Name:       "TestBuildFailed",
					Attributes: map[string]string{},
				},
			},
		},
	})
}

type MockReporterClient struct{}

func (m *MockReporterClient) ReportLoadResult(loadResult *sdkModel.LoadResult) error {
	return nil
}
func (m *MockReporterClient) ReportCaseResult(caseResult *sdkModel.TestResult) error {
	if caseResult.ResultType == sdkModel.ResultTypeRunning {
		reportRunningCount++
	} else if caseResult.ResultType == sdkModel.ResultTypeFailed || caseResult.ResultType == sdkModel.ResultTypeSucceed {
		if caseResult.ResultType == sdkModel.ResultTypeSucceed {
			reportSucceedCount++
		} else {
			reportFailedCount++
		}
		reportFinishedCount++
	}
	return nil
}
func (m *MockReporterClient) Close() error {
	return nil
}

func TestExecuteTestcases(t *testing.T) {
	projPath, err := filepath.Abs("../../testdata")
	assert.NoError(t, err)
	packages := map[string]map[string][]*testcase.TestCase{
		"demo": {
			"": {
				{
					Path:       "demo",
					Name:       "TestAdd",
					Attributes: map[string]string{},
				},
				{
					Path:       "demo",
					Name:       "TestAdd1",
					Attributes: map[string]string{},
				},
				{
					Path:       "demo",
					Name:       "TestAdd2",
					Attributes: map[string]string{},
				},
				{
					Path:       "demo",
					Name:       "TestSlowFunc",
					Attributes: map[string]string{},
				},
				{
					Path:       "demo",
					Name:       "TestPanic",
					Attributes: map[string]string{},
				},
			},
		},
		"demo/build": {
			"": {
				{
					Path:       "demo/build",
					Name:       "TestBuildFailed",
					Attributes: map[string]string{},
				},
			},
		},
	}
	err = executeTestcases(projPath, packages, &MockReporterClient{})
	assert.NoError(t, err)
}

func Test_discoverExecutableTestcases(t *testing.T) {
	projPath, err := filepath.Abs("../../testdata")
	assert.NoError(t, err)
	os.Chdir(projPath)
	// 验证可以基于指定目录路径找到路径下对应的所有包含测试用例的子目录
	testcases := []*testcase.TestCase{
		{
			Path: "demo",
			Name: "",
		},
	}
	execTestcases, err := discoverExecutableTestcases(testcases)
	assert.NoError(t, err)
	assert.Len(t, execTestcases, 2)
	// 验证如果传入的是文件路径则直接返回
	testcases = []*testcase.TestCase{
		{
			Path: "demo/demo_test.go",
			Name: "",
		},
		{
			Path: "demo/build/build_test.go",
			Name: "",
		},
	}
	execTestcases, err = discoverExecutableTestcases(testcases)
	assert.NoError(t, err)
	assert.Len(t, execTestcases, 2)
	// 验证如果传入的已经是子目录则不会返回额外用例
	testcases = []*testcase.TestCase{
		{
			Path: "demo/build",
			Name: "",
		},
	}
	execTestcases, err = discoverExecutableTestcases(testcases)
	assert.NoError(t, err)
	assert.Len(t, execTestcases, 1)
}
