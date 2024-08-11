package execute

import (
	"errors"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"

	gotestRunner "github.com/OpenTestSolar/testtool-golang-gotest/gotest/pkg/runner"
	gotestTestcase "github.com/OpenTestSolar/testtool-golang-gotest/gotest/pkg/testcase"
	gotestUtil "github.com/OpenTestSolar/testtool-golang-gotest/gotest/pkg/util"

	"github.com/OpenTestSolar/testtool-sdk-golang/api"
	sdkClient "github.com/OpenTestSolar/testtool-sdk-golang/client"
	pkgErrors "github.com/pkg/errors"
	"github.com/spf13/cobra"
)

type ExecuteOptions struct {
	executePath string
}

// NewBuildOptions new build options with default value
func NewExecuteOptions() *ExecuteOptions {
	return &ExecuteOptions{}
}

// NewCmdBuild create a build command
func NewCmdExecute() *cobra.Command {
	o := NewExecuteOptions()
	cmd := cobra.Command{
		Use:   "execute",
		Short: "Execute testcases",
		RunE: func(cmd *cobra.Command, args []string) error {
			return o.RunExecute(cmd)
		},
	}
	cmd.Flags().StringVarP(&o.executePath, "path", "p", "", "Path of testcase info")
	cmd.MarkFlagRequired("path")
	return &cmd
}

func groupTestCasesByPathAndName(projPath string, testcases []*gotestTestcase.TestCase) (map[string]map[string][]*gotestTestcase.TestCase, error) {
	packages := map[string]map[string][]*gotestTestcase.TestCase{}
	for _, testcase := range testcases {
		path, name, err := gotestUtil.GetPathAndFileName(projPath, testcase.Path)
		if err != nil {
			log.Printf("[PLUGIN]Get path and file name from %s failed, err: %s", testcase.Path, err.Error())
			return nil, err
		}
		path = strings.TrimSuffix(path, string(os.PathSeparator))
		_, ok := packages[path]
		if !ok {
			packages[path] = map[string][]*gotestTestcase.TestCase{}
		}
		_, ok = packages[path][name]
		if !ok {
			packages[path][name] = []*gotestTestcase.TestCase{}
		}
		packages[path][name] = append(packages[path][name], testcase)
	}
	return packages, nil
}

func executeTestcases(projPath string, packages map[string]map[string][]*gotestTestcase.TestCase, reporter api.Reporter) error {
	for path, filesCases := range packages {
		// test one file each time
		for fileName, cases := range filesCases {
			log.Printf("[PLUGIN]Run test cases under %s", filepath.Join(path, fileName))
			err := gotestRunner.RunTest(projPath, path, fileName, cases, reporter)
			if err != nil {
				return pkgErrors.Wrapf(err, "run test cases failed")
			}
		}
	}
	return nil
}

func parseTestcases(testSelectors []string) ([]*gotestTestcase.TestCase, error) {
	var testcases []*gotestTestcase.TestCase
	for _, selector := range testSelectors {
		testcase, err := gotestTestcase.ParseTestCaseBySelector(selector)
		if err != nil {
			log.Printf("[PLUGIN]parse testcase by selector [%s] failed, err: %s", selector, err.Error())
			continue
		}
		testcases = append(testcases, testcase)
	}
	if len(testcases) == 0 {
		return nil, errors.New("no available testcases")
	}
	return testcases, nil
}

func findTestPackagesByPath(path string) ([]string, error) {
	subDirs := []string{}
	foundSubDirs := map[string]bool{}
	err := filepath.WalkDir(path, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return pkgErrors.Wrapf(err, "walk subdir %s failed", p)
		}
		if !d.IsDir() && strings.HasSuffix(p, "_test.go") {
			subDir := filepath.Dir(p)
			if _, ok := foundSubDirs[subDir]; !ok {
				subDirs = append(subDirs, subDir)
				foundSubDirs[subDir] = true
			}
			return nil
		}
		return nil
	})
	if err != nil {
		return nil, pkgErrors.Wrapf(err, "walk dir %s failed", path)
	}
	return subDirs, nil
}

func discoverExecutableTestcases(testcases []*gotestTestcase.TestCase) ([]*gotestTestcase.TestCase, error) {
	excutableTestcases := []*gotestTestcase.TestCase{}
	for _, testcase := range testcases {
		fd, err := os.Stat(testcase.Path)
		if err != nil {
			return nil, pkgErrors.Wrapf(err, "get file info %s failed", testcase.Path)
		}
		if !fd.IsDir() {
			excutableTestcases = append(excutableTestcases, testcase)
			continue
		}
		packages, err := findTestPackagesByPath(testcase.Path)
		if err != nil {
			return nil, pkgErrors.Wrapf(err, "find packages in %s failed", testcase.Path)
		}
		if len(packages) == 0 {
			return nil, pkgErrors.Wrapf(err, "failed to found available test packages in dir %s", testcase.Path)
		}
		for _, pack := range packages {
			if pack != testcase.Path {
				log.Printf("[PLUGIN]found test package %s in %s", pack, testcase.Path)
			}
			t := &gotestTestcase.TestCase{
				Path:       pack,
				Name:       testcase.Name,
				Attributes: testcase.Attributes,
			}
			excutableTestcases = append(excutableTestcases, t)
		}
	}
	return excutableTestcases, nil
}

func (o *ExecuteOptions) RunExecute(cmd *cobra.Command) error {
	// load case info from yaml file
	config, err := gotestTestcase.UnmarshalCaseInfo(o.executePath)
	if err != nil {
		return pkgErrors.Wrapf(err, "failed to unmarshal case info")
	}
	// parse testcases
	testcases, err := parseTestcases(config.TestSelectors)
	if err != nil {
		return pkgErrors.Wrapf(err, "failed to parse test selectors")
	}
	// 递归查询包含实际可执行用例的目录
	excutableTestcases, err := discoverExecutableTestcases(testcases)
	if err != nil {
		return pkgErrors.Wrapf(err, "failed to discover excutble testcases")
	}
	// get workspace
	projPath := gotestUtil.GetWorkspace(config.ProjectPath)
	_, err = os.Stat(projPath)
	if err != nil {
		return pkgErrors.Wrapf(err, "stat project path %s failed", projPath)
	}
	// get testcases grouped by path and name
	packages, err := groupTestCasesByPathAndName(projPath, excutableTestcases)
	if err != nil {
		return pkgErrors.Wrap(err, "failed to group testcases by path and name")
	}
	// run testcases
	reporter, err := sdkClient.NewReporterClient(config.FileReportPath)
	if err != nil {
		return pkgErrors.Wrap(err, "failed to create reporter")
	}
	err = executeTestcases(projPath, packages, reporter)
	if err != nil {
		return pkgErrors.Wrapf(err, "failed to execute testcases")
	}
	return nil
}
