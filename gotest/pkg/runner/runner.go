package runner

import (
	"context"
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
	"github.com/sourcegraph/conc/pool"
)

func RunTest(projPath, path, fileName string, testcases []*gotestTestcase.TestCase, reporter api.Reporter) error {
	var cmdline string
	var tcNames []string
	for _, testcase := range testcases {
		tcNames = append(tcNames, testcase.Name)
	}
	caseFullRelPath := filepath.Join(path, fileName)
	if source, err := strconv.ParseBool(os.Getenv("TESTSOLAR_TTP_EXECUTEFROMSOURCE")); err == nil && source {
		log.Printf("[PLUGIN]Execute test from source")
		cmdline = fmt.Sprintf(`go test -v -json -run "%s$" %s`, strings.Join(tcNames, "|"), filepath.Join(projPath, path))
	} else {
		pkgBin := filepath.Join(projPath, path+".test")
		_, err := os.Stat(pkgBin)
		if err != nil {
			log.Printf("[PLUGIN]Can't find package bin file %s during running, try to build it...", pkgBin)
			_, err := gotestBuilder.BuildTestPackage(projPath, path, false)
			if err != nil {
				return errors.Wrapf(err, "Build package %s during running failed", path)
			}
		}
		_, minor, err := gotestUtil.ParseGoVersion()
		if err != nil || minor <= 19 {
			cmdline = fmt.Sprintf(`go tool test2json -t -p %s %s -test.v -test.run "%s$"`, caseFullRelPath, pkgBin, strings.Join(tcNames, "|"))
		} else {
			cmdline = fmt.Sprintf(`go tool test2json -t -p %s %s -test.v=test2json -test.run "%s$"`, caseFullRelPath, pkgBin, strings.Join(tcNames, "|"))
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
	testResults := make(chan *sdkModel.TestResult)
	output := make(chan string)
	// 并发启动协程，如果协程中有返回错误则报错
	// 1. 读取子进程标准输出流
	// 2. 解析子进程标准输出流
	// 3. 上报用例执行结果
	p := pool.New().
		WithContext(context.Background()).
		WithCancelOnError()
	p.Go(
		func(ctx context.Context) error {
			return gotestResult.ReadLines(stdout, output)
		},
	)
	p.Go(
		func(ctx context.Context) error {
			return gotestResult.ParseTestResult(output, testResults, caseFullRelPath)
		},
	)
	p.Go(
		func(ctx context.Context) error {
			return gotestResult.ReportTestResults(testResults, reporter)
		},
	)
	return p.Wait()
}
