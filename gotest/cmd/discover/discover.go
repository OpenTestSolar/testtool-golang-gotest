package discover

import (
	"fmt"
	gotestLoader "gotest/pkg/loader"
	gotestSelector "gotest/pkg/selector"
	gotestTestcase "gotest/pkg/testcase"
	gotestUtil "gotest/pkg/util"
	"log"
	"os"

	sdkClient "github.com/OpenTestSolar/testtool-sdk-golang/client"
	sdkModel "github.com/OpenTestSolar/testtool-sdk-golang/model"
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
		Run: func(cmd *cobra.Command, args []string) {
			o.RunDiscover(cmd)
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

func reportTestcases(testcases []*gotestTestcase.TestCase) error {
	reporter, err := sdkClient.NewReporterClient()
	if err != nil {
		log.Printf("[PLUGIN]Failed to create reporter: %v\n", err)
		return err
	}
	var tests []*sdkModel.TestCase
	for _, testcase := range testcases {
		tests = append(tests, &sdkModel.TestCase{
			Name:       testcase.GetSelector(),
			Attributes: testcase.Attributes,
		})
	}
	err = reporter.ReportLoadResult(&sdkModel.LoadResult{
		Tests:      tests,
		LoadErrors: nil,
	})
	if err != nil {
		log.Printf("[PLUGIN]Failed to report load result: %v\n", err)
		return err
	}
	err = reporter.Close()
	if err != nil {
		log.Printf("[PLUGIN]Failed to close report: %v\n", err)
		return err
	}
	return nil
}

func loadTestcases(projPath string, targetSelectors []*gotestSelector.TestSelector) []*gotestTestcase.TestCase {
	var testcases []*gotestTestcase.TestCase
	loadedSelectorPath := make(map[string]struct{})
	for _, testSelector := range targetSelectors {
		// skip the path that has been loaded
		if _, ok := loadedSelectorPath[testSelector.Path]; ok {
			continue
		}
		loadedSelectorPath[testSelector.Path] = struct{}{}

		loadedTestcases, err := gotestLoader.LoadTestCase(projPath, testSelector.Path)
		if err != nil {
			log.Printf("[PLUGIN]Load testcase from path %s failed, err: %v", testSelector.Path, err)
			continue
		}
		testcases = append(testcases, loadedTestcases...)
	}
	return testcases
}

func (o *DiscoverOptions) RunDiscover(cmd *cobra.Command) error {
	// load case info from yaml file
	config, err := gotestTestcase.UnmarshalCaseInfo(o.discoverPath)
	if err != nil {
		return err
	}
	// parse selectors
	targetSelectors := parseTestSelectors(config.TestSelectors)
	log.Printf("[PLUGIN]load testcases from selectors: %s", targetSelectors)
	// get workspace
	projPath := gotestUtil.GetWorkspace(config.ProjectPath)
	_, err = os.Stat(projPath)
	if err != nil {
		return fmt.Errorf("stat project path %s failed, err: %s", projPath, err.Error())
	}
	// load testcases
	testcases := loadTestcases(projPath, targetSelectors)
	// report testcases
	return reportTestcases(testcases)
}
