package cmd

import (
	"os"
	"os/exec"

	"github.com/ps-gill/rpmtool/rpm"
	"github.com/spf13/cobra"
)

var (
	buildCmd = &cobra.Command{
		Use:  "build [spec]",
		Short: "Build package from a .spec file",
		Args: cobra.ExactArgs(1),
		RunE: build,
	}
	flagLatestDeps = "latest-deps"
	flagSkipDeps   = "skip-deps"
	flagSrpm       = "srpm"
)

func init() {
	buildCmd.Flags().Bool(flagLatestDeps, false, "install latest build dependencies")
	buildCmd.Flags().Bool(flagSkipDeps, false, "skip build dependencies installation")
	buildCmd.Flags().Bool(flagSrpm, false, "build srpm instead of rpm")
	rootCmd.AddCommand(buildCmd)
}

func build(cmd *cobra.Command, args []string) error {
	specPath := args[0]

	srpm, err := cmd.Flags().GetBool(flagSrpm)
	if err != nil {
		return err
	}

	skipDeps, err := cmd.Flags().GetBool(flagSkipDeps)
	if err != nil {
		return err
	}

	if err := rpm.DownloadSources(specPath); err != nil {
		return err
	}

	if !srpm && !skipDeps {
		latestDeps, err := cmd.Flags().GetBool(flagLatestDeps)
		if err != nil {
			return err
		}

		if err := rpm.InstallBuildDependencies(specPath, latestDeps); err != nil {
			return err
		}
	}

	if err := runRpmBuild(specPath, srpm); err != nil {
		return err
	}

	return nil
}

func runRpmBuild(specPath string, srpm bool) error {
	buildType := "-bb"
	if srpm {
		buildType = "-bs"
	}

	runRpmCmd := exec.Command("rpmbuild", buildType, specPath)
	runRpmCmd.Stdin = os.Stdin
	runRpmCmd.Stdout = os.Stdout
	runRpmCmd.Stderr = os.Stderr
	return runRpmCmd.Run()
}
