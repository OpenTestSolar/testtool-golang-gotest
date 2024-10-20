package discover

import (
	"fmt"
	"log"
	"os"
	"strings"

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

func expandTestcaseBySelector(testcase *gotestTestcase.TestCase, testSelectors []*gotestSelector.TestSelector) []*gotestTestcase.TestCase {
	// 根据testselector来扩展加载用例
	// 例如加载的用例为/path/to/target?TestAdd
	// 而testselector中用例包含/path/to/target?TestAdd/1以及/path/to/target?TestAdd/2
	// 则返回/path/to/target?TestAdd/1以及/path/to/target?TestAdd/2
	var expandTestcases []*gotestTestcase.TestCase
	for _, selector := range testSelectors {
		if testcase.Path == selector.Path && strings.HasPrefix(selector.Name, fmt.Sprintf("%s/", testcase.Name)) {
			subTestcase := &gotestTestcase.TestCase{
				Path:       testcase.Path,
				Name:       selector.Name,
				Attributes: testcase.Attributes,
			}
			expandTestcases = append(expandTestcases, subTestcase)
		}
	}
	return expandTestcases
}

func loadTestcases(projPath string, targetSelectors []*gotestSelector.TestSelector) ([]*gotestTestcase.TestCase, []*sdkModel.LoadError) {
	var testcases []*gotestTestcase.TestCase
	var expandTestcases []*gotestTestcase.TestCase
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
	// 考虑到单独下发一个子用例的场景，加载阶段由于都是静态解析用例，因此无法解析到具体的子用例
	// 例如下发的selector格式为/path/to/target?TestAdd/1
	// 则通过LoadTestCase加载到的用例为/path/to/target?TestAdd
	// 这种情况下需要将加载的用例由/path/to/target?TestAdd替换为/path/to/target?TestAdd/1返回
	for _, testcase := range testcases {
		if expand := expandTestcaseBySelector(testcase, targetSelectors); expand != nil {
			expandTestcases = append(expandTestcases, expand...)
		} else {
			expandTestcases = append(expandTestcases, testcase)
		}
	}
	return expandTestcases, loadErrors
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
