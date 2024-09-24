package discover

import (
	"path/filepath"
	"testing"

	"github.com/OpenTestSolar/testtool-golang-gotest/pkg/loader"
	"github.com/OpenTestSolar/testtool-golang-gotest/pkg/selector"
	"github.com/OpenTestSolar/testtool-golang-gotest/pkg/testcase"

	sdkModel "github.com/OpenTestSolar/testtool-sdk-golang/model"
	"github.com/agiledragon/gomonkey/v2"
	"github.com/stretchr/testify/assert"
)

func TestNewDiscoverOptions(t *testing.T) {
	o := NewDiscoverOptions()
	assert.NotNil(t, o)
}

func TestNewCmdDiscover(t *testing.T) {
	cmd := NewCmdDiscover()
	assert.NotNil(t, cmd)
}

func TestParseTestSelectors(t *testing.T) {
	testSelectors := []string{"path?name=test%20name&attr1=value%3D1"}
	selectors := parseTestSelectors(testSelectors)
	assert.Len(t, selectors, 1)
}

type MockReporterClient struct{}

func (m *MockReporterClient) ReportLoadResult(loadResult *sdkModel.LoadResult) error {
	return nil
}
func (m *MockReporterClient) ReportCaseResult(caseResult *sdkModel.TestResult) error {
	return nil
}
func (m *MockReporterClient) Close() error {
	return nil
}

func TestReportTestcases(t *testing.T) {
	testcases := []*testcase.TestCase{
		{
			Path:       "",
			Name:       "",
			Attributes: map[string]string{},
		},
	}
	loadErrors := []*sdkModel.LoadError{
		{
			Name:    "",
			Message: "",
		},
	}
	err := reportTestcases(testcases, loadErrors, &MockReporterClient{})
	assert.NoError(t, err)
}

func TestLoadTestcases(t *testing.T) {
	LoadTestCaseMock := gomonkey.ApplyFunc(loader.LoadTestCase, func(projPath string, selectorPath string) ([]*testcase.TestCase, error) {
		return []*testcase.TestCase{
			{
				Path:       "path/to/test",
				Name:       "test01",
				Attributes: map[string]string{},
			},
		}, nil
	})
	defer LoadTestCaseMock.Reset()
	testSelectors := []*selector.TestSelector{
		{
			Value:      "",
			Path:       "path/to/test",
			Name:       "test01",
			Attributes: map[string]string{},
		},
		{
			Value:      "",
			Path:       "path/to/test",
			Name:       "test02",
			Attributes: map[string]string{},
		},
	}
	projPath, err := filepath.Abs("../../testdata")
	assert.NoError(t, err)
	testcases, loadErrors := loadTestcases(projPath, testSelectors)
	assert.Len(t, testcases, 1)
	assert.Len(t, loadErrors, 0)
}
