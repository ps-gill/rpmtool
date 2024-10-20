package cmd

import (
	"fmt"
	"os/exec"

	"github.com/ps-gill/rpmtool/rpm"
	"github.com/spf13/cobra"
)

var (
	toolsCmd = &cobra.Command{
		Use:   "tools",
		Short: "Check if required tools are available",
		RunE:  tools,
	}
	flagToolsExcludeSignature = "exclude-signature"

	checkMark = "✓"
	crossMark = "✗"
)

func init() {
	toolsCmd.Flags().Bool(flagToolsExcludeSignature, false, "exclude signature tools")
	rootCmd.AddCommand(toolsCmd)
}

func tools(cmd *cobra.Command, args []string) error {

	excludeSignatureTools, err := cmd.Flags().GetBool(flagToolsExcludeSignature)
	if err != nil {
		return err
	}

	for _, tool := range rpm.GetTools(!excludeSignatureTools) {
		if _, err := exec.LookPath(tool); err != nil {
			fmt.Printf("  %s %s\n", crossMark, tool)
		} else {
			fmt.Printf("  %s %s\n", checkMark, tool)
		}
	}

	return nil
}
