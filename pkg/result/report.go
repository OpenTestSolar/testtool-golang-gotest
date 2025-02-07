package result

import (
	"context"
	"io"
	"log"

	"github.com/OpenTestSolar/testtool-sdk-golang/api"
	sdkModel "github.com/OpenTestSolar/testtool-sdk-golang/model"
	"github.com/pkg/errors"
	"github.com/sourcegraph/conc/pool"
)

func ReportTestResults(testResults chan *sdkModel.TestResult, reporter api.Reporter) error {
	for result := range testResults {
		log.Printf("[PLUGIN]Reporting test result: %s, result: %d", result.Test.Name, result.ResultType)
		err := reporter.ReportCaseResult(result)
		if err != nil {
			return errors.Wrapf(err, "Report test result failed")
		}
	}
	log.Printf("[PLUGIN]Report test results finished")
	return nil
}

func ParseAndReportResult(stdout io.ReadCloser, filePath string, reporter api.Reporter) error {
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
			return ReadLines(stdout, output)
		},
	)
	p.Go(
		func(ctx context.Context) error {
			return ParseTestResult(output, testResults, filePath)
		},
	)
	p.Go(
		func(ctx context.Context) error {
			return ReportTestResults(testResults, reporter)
		},
	)
	return p.Wait()
}
