package main

import (
	"github.com/OpenTestSolar/testtool-golang-gotest/gotest/cmd/build"
	"github.com/OpenTestSolar/testtool-golang-gotest/gotest/cmd/discover"
	"github.com/OpenTestSolar/testtool-golang-gotest/gotest/cmd/execute"

	"github.com/spf13/cobra"
)

func main() {
	rootCmd := cobra.Command{
		Use: "solar-gotest",
	}
	rootCmd.AddCommand(discover.NewCmdDiscover())
	rootCmd.AddCommand(execute.NewCmdExecute())
	rootCmd.AddCommand(build.NewCmdBuild())
	rootCmd.Execute()
}
