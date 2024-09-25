package result

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"strings"
	"time"

	sdkModel "github.com/OpenTestSolar/testtool-sdk-golang/model"
	"github.com/pkg/errors"
)

type TestEvent struct {
	Time    time.Time
	Action  string
	Package string
	Test    string
	Elapsed float64
	Output  string
}

func (te *TestEvent) TestStart() bool {
	return te.Action == "run"
}

func (te *TestEvent) TestFinished() bool {
	return te.Test != "" && (te.Action == "pass" || te.Action == "fail")
}

func (te *TestEvent) TestOutput() bool {
	return te.Test != "" && te.Action == "output"
}

func ParseCaseResult(caseResult string) sdkModel.ResultType {
	if caseResult == "pass" {
		return sdkModel.ResultTypeSucceed
	} else if caseResult == "skip" {
		return sdkModel.ResultTypeIgnored
	} else if caseResult == "fail" {
		return sdkModel.ResultTypeFailed
	} else if caseResult == "run" {
		return sdkModel.ResultTypeRunning
	} else {
		return sdkModel.ResultTypeUnknown
	}
}

func GenTestResult(name string, caseRunResult string, logs []string, startTime time.Time, endTime time.Time) *sdkModel.TestResult {
	runResult := ParseCaseResult(caseRunResult)
	var logLevel sdkModel.LogLevel
	if runResult == sdkModel.ResultTypeFailed {
		logLevel = sdkModel.LogLevelError
	} else {
		logLevel = sdkModel.LogLevelInfo
	}
	var caseSteps []*sdkModel.TestCaseStep
	var caseRunLogs []*sdkModel.TestCaseLog
	for _, l := range logs {
		caseRunLogs = append(caseRunLogs, &sdkModel.TestCaseLog{
			Time:    startTime,
			Level:   logLevel,
			Content: l,
		})
	}
	if caseRunLogs != nil {
		caseSteps = []*sdkModel.TestCaseStep{
			{
				StartTime: startTime,
				EndTime:   endTime,
				Title:     "TestCase: ",
				Logs:      caseRunLogs,
			},
		}
	}
	test := &sdkModel.TestCase{
		Name: name,
	}
	return &sdkModel.TestResult{
		StartTime:  startTime,
		EndTime:    endTime,
		Test:       test,
		ResultType: runResult,
		Steps:      caseSteps,
	}
}

func ReadLines(stdout io.ReadCloser, output chan string) error {
	defer close(output)
	reader := bufio.NewReader(stdout)
	for {
		line, isPrefix, err := reader.ReadLine()
		if err != nil {
			if err == io.EOF {
				break
			}
			return errors.Wrapf(err, "read line error")
		}
		// 如果isPrefix为true，表示行太长，没有完全读取
		// 需要继续读取直到isPrefix为false，表示行的末尾
		for isPrefix {
			var more []byte
			more, isPrefix, err = reader.ReadLine()
			if err != nil {
				if err == io.EOF {
					output <- string(line)
					return nil
				}
				return errors.Wrapf(err, "read prefix line error")
			}
			line = append(line, more...)
		}
		output <- string(line)
	}
	log.Printf("[PLUGIN]run testcases finished")
	return nil
}

type CurrentRunningCaseInfo struct {
	Log       []string
	StartTime time.Time
}

func ParseTestResult(output chan string, testResults chan *sdkModel.TestResult, filePath string) error {
	defer close(testResults)
	info := make(map[string]*CurrentRunningCaseInfo)
	for line := range output {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		log.Printf("[PLUGIN]%s", line)
		var event TestEvent
		err := json.Unmarshal([]byte(line), &event)
		if err != nil {
			return errors.Wrapf(err, "failed to parse line: %s", line)
		}
		if event.TestStart() {
			info[event.Test] = &CurrentRunningCaseInfo{
				Log:       []string{},
				StartTime: event.Time,
			}
			// 上报当前运行用例
			caseResult := GenTestResult(fmt.Sprintf("%s?%s", filePath, event.Test), event.Action, nil, event.Time, event.Time)
			log.Printf("[PLUGIN]report running case %s", event.Test)
			testResults <- caseResult
		} else if event.TestFinished() {
			var caseResult *sdkModel.TestResult
			name := fmt.Sprintf("%s?%s", filePath, event.Test)
			if _, ok := info[event.Test]; ok {
				caseResult = GenTestResult(name, event.Action, info[event.Test].Log, info[event.Test].StartTime, event.Time)
			} else {
				log.Printf("[PLUGIN]can't find case: %s in current running cases", name)
				caseResult = GenTestResult(name, event.Action, nil, event.Time, event.Time)
			}
			// 上报用例执行结果
			log.Printf("[PLUGIN]report case %s, result: %s", name, event.Action)
			testResults <- caseResult
		} else if event.TestOutput() {
			info[event.Test].Log = append(info[event.Test].Log, event.Output)
		}
	}
	return nil
}
