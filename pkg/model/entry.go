package model

import (
	sdkModel "github.com/OpenTestSolar/testtool-sdk-golang/model"
)

type EntryParam struct {
	taskId         string
	projectPath    string
	context        map[string]string
	testSelectors  []string
	collectors     []string
	fileReportPath string
}

func NewEntryParam(entry sdkModel.EntryParam) *EntryParam {
	return &EntryParam{
		taskId:         entry.TaskId,
		projectPath:    entry.ProjectPath,
		context:        entry.Context,
		testSelectors:  entry.TestSelectors,
		collectors:     entry.Collectors,
		fileReportPath: entry.FileReportPath,
	}
}

func (ep *EntryParam) GetProjectPath() string {
	return ep.projectPath
}

func (ep *EntryParam) GetTestSelectors() []string {
	return ep.testSelectors
}

func (ep *EntryParam) GetFileReportPath() string {
	return ep.fileReportPath
}

func (ep *EntryParam) RetrieveRawCmdline() string {
	if cmdline, ok := ep.context["raw_cmdline"]; ok {
		return cmdline
	}
	return ""
}
