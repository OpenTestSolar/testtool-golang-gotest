package runner

import (
	"gotest/pkg/builder"
	gotestTestcase "gotest/pkg/testcase"
	"path/filepath"
	"testing"

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
	targetRunningTests := 8
	targetFinishedTests := 8
	targetSucceedTests := 6
	targetFailedTests := 2
	NewReporterClientMock := gomonkey.ApplyFunc(sdkClient.NewReporterClient, func() (sdkApi.Reporter, error) {
		return &MockReporterClient{}, nil
	})
	defer NewReporterClientMock.Reset()
	absPath, err := filepath.Abs("../../testdata/")
	assert.NoError(t, err)
	err = builder.Build(absPath)
	assert.NoError(t, err)
	// pkgBin := "../../testdata/demo.test"
	// _, err = os.Stat(pkgBin)
	// assert.NoError(t, err)
	// defer os.Remove("../../testdata/demo.test")
	binPath := filepath.Join(absPath, "demo.test")
	err = RunTest(absPath, binPath, []*gotestTestcase.TestCase{
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
			Name: "SlowFunc",
		},
		{
			Path: "demo/demo_test.go",
			Name: "TestPanic",
		},
	})
	assert.NoError(t, err)
	assert.Equal(t, reportRunningCount, targetRunningTests)
	assert.Equal(t, reportFinishedCount, targetFinishedTests)
	assert.Equal(t, reportSucceedCount, targetSucceedTests)
	assert.Equal(t, reportFailedCount, targetFailedTests)
}
