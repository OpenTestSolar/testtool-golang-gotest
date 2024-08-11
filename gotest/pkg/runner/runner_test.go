package runner

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/OpenTestSolar/testtool-golang-gotest/gotest/pkg/builder"
	gotestTestcase "github.com/OpenTestSolar/testtool-golang-gotest/gotest/pkg/testcase"

	sdkModel "github.com/OpenTestSolar/testtool-sdk-golang/model"
	"github.com/stretchr/testify/assert"
)

var (
	reportRunningCount  = 0
	reportFinishedCount = 0
	reportSucceedCount  = 0
	reportFailedCount   = 0
	runningTests        = []*sdkModel.TestResult{}
	failedTests         = []*sdkModel.TestResult{}
	succeedTests        = []*sdkModel.TestResult{}
	finishedTests       = []*sdkModel.TestResult{}
)

type MockReporterClient struct{}

func (m *MockReporterClient) ReportLoadResult(loadResult *sdkModel.LoadResult) error {
	return nil
}
func (m *MockReporterClient) ReportCaseResult(caseResult *sdkModel.TestResult) error {
	if caseResult.ResultType == sdkModel.ResultTypeRunning {
		runningTests = append(runningTests, caseResult)
		reportRunningCount++
	} else if caseResult.ResultType == sdkModel.ResultTypeFailed || caseResult.ResultType == sdkModel.ResultTypeSucceed {
		if caseResult.ResultType == sdkModel.ResultTypeSucceed {
			succeedTests = append(succeedTests, caseResult)
			reportSucceedCount++
		} else {
			failedTests = append(failedTests, caseResult)
			reportFailedCount++
		}
		finishedTests = append(finishedTests, caseResult)
		reportFinishedCount++
	}
	return nil
}
func (m *MockReporterClient) Close() error {
	return nil
}

func TestRunTest(t *testing.T) {
	targetRunningTests := 7
	targetFinishedTests := 7
	targetSucceedTests := 5
	targetFailedTests := 2
	absPath, err := filepath.Abs("../../testdata/")
	assert.NoError(t, err)
	err = builder.Build(absPath)
	assert.NoError(t, err)
	for _, s := range []string{"0", "1"} {
		err = os.Setenv("TESTSOLAR_TTP_EXECUTEFROMSOURCE", s)
		assert.NoError(t, err)
		reportRunningCount = 0
		reportFinishedCount = 0
		reportSucceedCount = 0
		reportFailedCount = 0
		err = RunTest(absPath, "demo", "demo_test.go", []*gotestTestcase.TestCase{
			{
				Path: "demo/demo_test.go",
				Name: "TestAdd",
			},
			{
				Path: "demo/demo_test.go",
				Name: "TestAdd1",
			},
			{
				Path: "demo/demo_test.go",
				Name: "TestAdd2",
			},
			{
				Path: "demo/demo_test.go",
				Name: "TestPanic",
			},
		}, &MockReporterClient{})
		assert.NoError(t, err)
		assert.Equal(t, reportRunningCount, targetRunningTests)
		assert.Equal(t, reportFinishedCount, targetFinishedTests)
		assert.Equal(t, reportSucceedCount, targetSucceedTests)
		assert.Equal(t, reportFailedCount, targetFailedTests)
	}
}
