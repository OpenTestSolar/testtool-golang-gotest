package discover

import (
	"log"
	"os"

	gotestLoader "github.com/OpenTestSolar/testtool-golang-gotest/pkg/loader"
	gotestSelector "github.com/OpenTestSolar/testtool-golang-gotest/pkg/selector"
	gotestTestcase "github.com/OpenTestSolar/testtool-golang-gotest/pkg/testcase"
	gotestUtil "github.com/OpenTestSolar/testtool-golang-gotest/pkg/util"

	"github.com/OpenTestSolar/testtool-sdk-golang/api"
	sdkClient "github.com/OpenTestSolar/testtool-sdk-golang/client"
	sdkModel "github.com/OpenTestSolar/testtool-sdk-golang/model"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

type DiscoverOptions struct {
	discoverPath string
}

// NewBuildOptions new build options with default value
func NewDiscoverOptions() *DiscoverOptions {
	return &DiscoverOptions{}
}

// NewCmdBuild create a build command
func NewCmdDiscover() *cobra.Command {
	o := NewDiscoverOptions()
	cmd := cobra.Command{
		Use:   "discover",
		Short: "Discover testcases",
		RunE: func(cmd *cobra.Command, args []string) error {
			return o.RunDiscover(cmd)
		},
	}
	cmd.Flags().StringVarP(&o.discoverPath, "path", "p", "", "Path of testcase info")
	cmd.MarkFlagRequired("path")
	return &cmd
}

func parseTestSelectors(testSelector []string) []*gotestSelector.TestSelector {
	if len(testSelector) == 0 {
		testSelector = []string{"."}
	}
	var targetSelectors []*gotestSelector.TestSelector
	for _, selector := range testSelector {
		testSelector, err := gotestSelector.NewTestSelector(selector)
		if err != nil {
			log.Printf("[PLUGIN]Ignore invalid test selector: %s", selector)
			continue
		}
		if !testSelector.IsExclude() {
			targetSelectors = append(targetSelectors, testSelector)
		}
	}
	return targetSelectors
}

func reportTestcases(testcases []*gotestTestcase.TestCase, loadErrors []*sdkModel.LoadError, reporter api.Reporter) error {
	var tests []*sdkModel.TestCase
	for _, testcase := range testcases {
		tests = append(tests, &sdkModel.TestCase{
			Name:       testcase.GetSelector(),
			Attributes: testcase.Attributes,
		})
	}
	err := reporter.ReportLoadResult(&sdkModel.LoadResult{
		Tests:      tests,
		LoadErrors: loadErrors,
	})
	if err != nil {
		return errors.Wrap(err, "failed to report load result")
	}
	return nil
}

func loadTestcases(projPath string, targetSelectors []*gotestSelector.TestSelector) ([]*gotestTestcase.TestCase, []*sdkModel.LoadError) {
	var testcases []*gotestTestcase.TestCase
	var loadErrors []*sdkModel.LoadError
	loadedSelectorPath := make(map[string]struct{})
	for _, testSelector := range targetSelectors {
		// skip the path that has been loaded
		if _, ok := loadedSelectorPath[testSelector.Path]; ok {
			continue
		}
		loadedSelectorPath[testSelector.Path] = struct{}{}
		loadedTestcases, err := gotestLoader.LoadTestCase(projPath, testSelector.Path)
		if err != nil {
			loadErrors = append(loadErrors, &sdkModel.LoadError{
				Name:    testSelector.Path,
				Message: err.Error(),
			})
			continue
		}
		testcases = append(testcases, loadedTestcases...)
	}
	return testcases, loadErrors
}

func (o *DiscoverOptions) RunDiscover(cmd *cobra.Command) error {
	config, err := gotestTestcase.UnmarshalCaseInfo(o.discoverPath)
	if err != nil {
		return errors.Wrapf(err, "failed to unmarshal case info")
	}
	targetSelectors := parseTestSelectors(config.TestSelectors)
	log.Printf("[PLUGIN]load testcases from selectors: %s", targetSelectors)
	projPath := gotestUtil.GetWorkspace(config.ProjectPath)
	_, err = os.Stat(projPath)
	if err != nil {
		return errors.Wrapf(err, "stat project path %s failed", projPath)
	}
	testcases, loadErrors := loadTestcases(projPath, targetSelectors)
	reporter, err := sdkClient.NewReporterClient(config.FileReportPath)
	if err != nil {
		return errors.Wrapf(err, "failed to create reporter")
	}
	err = reportTestcases(testcases, loadErrors, reporter)
	if err != nil {
		return errors.Wrapf(err, "failed to report testcases")
	}
	return nil
}
