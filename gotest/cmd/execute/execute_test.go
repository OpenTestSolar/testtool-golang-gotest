package execute

import (
	"path/filepath"
	"testing"

	"gotest/pkg/testcase"

	sdkApi "github.com/OpenTestSolar/testtool-sdk-golang/api"
	sdkClient "github.com/OpenTestSolar/testtool-sdk-golang/client"
	sdkModel "github.com/OpenTestSolar/testtool-sdk-golang/model"
	"github.com/agiledragon/gomonkey/v2"
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
	NewReporterClientMock := gomonkey.ApplyFunc(sdkClient.NewReporterClient, func() (sdkApi.Reporter, error) {
		return &MockReporterClient{}, nil
	})
	defer NewReporterClientMock.Reset()
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
	err = executeTestcases(projPath, packages)
	assert.NoError(t, err)
}
