package build

import (
	"testing"

	"github.com/OpenTestSolar/testtool-golang-gotest/gotest/pkg/builder"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/stretchr/testify/assert"
)

func TestNewBuildOptions(t *testing.T) {
	o := NewBuildOptions()
	if o == nil {
		t.Error("NewBuildOptions() should not return nil")
	}
}

func TestNewCmdBuild(t *testing.T) {
	cmd := NewCmdBuild()
	if cmd == nil {
		t.Error("NewCmdBuild() should not return nil")
	}
}

func TestRunBuild(t *testing.T) {
	BuildMock := gomonkey.ApplyFunc(builder.Build, func(projPath string) error {
		return nil
	})
	defer BuildMock.Reset()
	o := NewBuildOptions()
	assert.NotNil(t, o)
	cmd := NewCmdBuild()
	assert.NotNil(t, cmd)
	err := o.RunBuild(cmd, nil)
	assert.NoError(t, err)
}
