package cmd

import (
	"errors"
	"os"

	"github.com/ps-gill/rpmtool/rpm"
	"github.com/spf13/cobra"
)

var (
	buildCmd = &cobra.Command{
		Use:   "build [spec]",
		Short: "Build package from a .spec file",
		Args:  cobra.ExactArgs(1),
		RunE:  build,
	}
	flagBuildLatestDeps       = "latest-deps"
	flagBuildSkipDeps         = "skip-deps"
	flagBuildSrpm             = "srpm"
	flagBuildGpgKey           = "gpg-key"
	flagBuildGpgKeyPassphrase = "gpg-key-passphrase"
	flagBuildGpgKeyId         = "gpg-key-id"
)

func init() {
	buildCmd.Flags().Bool(flagBuildLatestDeps, false, "install latest build dependencies")
	buildCmd.Flags().Bool(flagBuildSkipDeps, false, "skip build dependencies installation")
	buildCmd.Flags().Bool(flagBuildSrpm, false, "build srpm instead of rpm")
	buildCmd.Flags().String(flagBuildGpgKey, "", "gpg key")
	buildCmd.Flags().String(flagBuildGpgKeyPassphrase, "", "gpg key passphrase")
	buildCmd.Flags().String(flagBuildGpgKeyId, "", "gpg key Id")
	rootCmd.AddCommand(buildCmd)
}

func build(cmd *cobra.Command, args []string) error {
	specPath := args[0]

	srpm, err := cmd.Flags().GetBool(flagBuildSrpm)
	if err != nil {
		return err
	}

	skipDeps, err := cmd.Flags().GetBool(flagBuildSkipDeps)
	if err != nil {
		return err
	}

	signatureKey := getSignatureKey(cmd)
	if err := rpm.CheckTools(signatureKey != nil); err != nil {
		return err
	}

	if err := rpm.DownloadSources(specPath); err != nil {
		return err
	}

	if !srpm && !skipDeps {
		latestDeps, err := cmd.Flags().GetBool(flagBuildLatestDeps)
		if err != nil {
			return err
		}

		if err := rpm.InstallBuildDependencies(specPath, latestDeps); err != nil {
			return err
		}
	}

	if srpm {
		if err := rpm.EmptySrpmDirectory(); err != nil {
			return err
		}
	} else {
		if err := rpm.EmptyRpmDirectory(); err != nil {
			return err
		}
	}

	if err := rpm.Build(specPath, srpm); err != nil {
		return err
	}

	if signatureKey == nil {
		return nil
	}

	gpgHome, err := rpm.SetupGpgKey(signatureKey)
	if err != nil {
		return err
	}
	defer gpgHome.Close()

	var rpmPackages []string
	if srpm {
		if rpmPackages, err = rpm.GetSrpmPackages(); err != nil {
			return err
		}
	} else {
		if rpmPackages, err = rpm.GetRpmPackages(); err != nil {
			return err
		}
	}

	if rpmPackages == nil || len(rpmPackages) == 0 {
		return errors.New("No package found")
	}

	return rpm.SignPackages(gpgHome, signatureKey, rpmPackages...)
}

func getSignatureKey(cmd *cobra.Command) *rpm.GpgKey {
	gpgKey, err := cmd.Flags().GetString(flagBuildGpgKey)
	if err != nil {
		return nil
	}

	gpgKeyPassphrase, err := cmd.Flags().GetString(flagBuildGpgKeyPassphrase)
	if err != nil {
		return nil
	}

	gpgKeyId, err := cmd.Flags().GetString(flagBuildGpgKeyId)
	if err != nil {
		return nil
	}

	if gpgKey == "" || gpgKeyPassphrase == "" || gpgKeyId == "" {
		return nil
	}
	info, err := os.Stat(gpgKey)
	if err != nil {
		return nil
	}
	if info.IsDir() {
		return nil
	}
	return &rpm.GpgKey{
		Key:           gpgKey,
		KeyPassphrase: gpgKeyPassphrase,
		KeyId:         gpgKeyId,
	}
}
