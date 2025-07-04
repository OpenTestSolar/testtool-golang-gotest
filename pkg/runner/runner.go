package runner

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	gotestBuilder "github.com/OpenTestSolar/testtool-golang-gotest/pkg/builder"
	gotestResult "github.com/OpenTestSolar/testtool-golang-gotest/pkg/result"

	gotestTestcase "github.com/OpenTestSolar/testtool-golang-gotest/pkg/testcase"
	gotestUtil "github.com/OpenTestSolar/testtool-golang-gotest/pkg/util"
	"github.com/OpenTestSolar/testtool-sdk-golang/api"
	sdkModel "github.com/OpenTestSolar/testtool-sdk-golang/model"
	"github.com/pkg/errors"
)

func reportFailedResultsAfterErr(err error, testcases []*gotestTestcase.TestCase, reporter api.Reporter) {
	if err == nil {
		return
	}
	for _, testcase := range testcases {
		if reportErr := reporter.ReportCaseResult(&sdkModel.TestResult{
			Test: &sdkModel.TestCase{
				Name:       fmt.Sprintf("%s?%s", testcase.Path, testcase.Name),
				Attributes: testcase.Attributes,
			},
			ResultType: sdkModel.ResultTypeFailed,
			Message:    fmt.Sprintf("Run test failed: %s", err.Error()),
		}); reportErr != nil {
			log.Printf("[PLUGIN]Report err [%s] failed, report err: %s", err.Error(), reportErr.Error())
		}
	}
}

func RunTest(projPath, path, fileName string, testcases []*gotestTestcase.TestCase, reporter api.Reporter) error {
	var err error
	defer func() {
		reportFailedResultsAfterErr(err, testcases, reporter)
	}()
	var cmdline string
	var tcNames []string
	nameFilter := make(map[string]bool)
	for _, testcase := range testcases {
		var name string
		if strings.Contains(testcase.Name, "/") {
			name = strings.SplitN(testcase.Name, "/", 2)[0]
		} else {
			name = testcase.Name
		}
		if _, ok := nameFilter[name]; ok {
			continue
		} else {
			nameFilter[name] = true
		}
		tcNames = append(tcNames, fmt.Sprintf("^%s$", name))
	}
	caseFullRelPath := filepath.Join(path, fileName)
	source, err := strconv.ParseBool(os.Getenv("TESTSOLAR_TTP_EXECUTEFROMSOURCE"))
	if err == nil && source {
		log.Printf("[PLUGIN]Execute test from source")
		cmdline = fmt.Sprintf(`go test -v -json -run "%s" %s`, strings.Join(tcNames, "|"), filepath.Join(projPath, path))
	} else {
		pkgBin := filepath.Join(projPath, path+".test")
		_, err = os.Stat(pkgBin)
		if err != nil {
			log.Printf("[PLUGIN]Can't find package bin file %s during running, try to build it...", pkgBin)
			_, err = gotestBuilder.BuildTestPackage(projPath, path, false)
		}
		if err != nil {
			return errors.Wrapf(err, "build package %s failed", path)
		} else {
			_, minor, err := gotestUtil.ParseGoVersion()
			if err != nil || minor <= 19 {
				cmdline = fmt.Sprintf(`go tool test2json -t -p %s %s -test.v -test.run "%s"`, caseFullRelPath, pkgBin, strings.Join(tcNames, "|"))
			} else {
				cmdline = fmt.Sprintf(`go tool test2json -t -p %s %s -test.v=test2json -test.run "%s"`, caseFullRelPath, pkgBin, strings.Join(tcNames, "|"))
			}
		}
	}
	extra_args := os.Getenv("TESTSOLAR_TTP_EXTRAARGS")
	if extra_args != "" {
		cmdline += " " + extra_args
	}
	cmdline += " 2>&1"
	log.Printf("[PLUGIN]Run cmdline %s", cmdline)
	stdout, _, err := gotestUtil.RunCommand(cmdline, projPath, false, true)
	if err != nil {
		return errors.Wrapf(err, "run cmd %s failed", cmdline)
	}
	return gotestResult.ParseAndReportResult(stdout, caseFullRelPath, reporter)
}
