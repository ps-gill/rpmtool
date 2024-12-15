package cmd

import (
	"errors"

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
	flagBuildLatestDeps        = "latest-deps"
	flagBuildSkipDeps          = "skip-deps"
	flagBuildSrpm              = "srpm"
	flagBuildKey               = "key"
	flagBuildKeyPassphraseFile = "key-passphrase-file"
	flagBuildKeyId             = "key-id"
)

func init() {
	buildCmd.Flags().Bool(flagBuildLatestDeps, false, "install latest build dependencies")
	buildCmd.Flags().Bool(flagBuildSkipDeps, false, "skip build dependencies installation")
	buildCmd.Flags().Bool(flagBuildSrpm, false, "build srpm instead of rpm")
	buildCmd.Flags().String(flagBuildKey, "", "pgp key")
	buildCmd.Flags().String(flagBuildKeyPassphraseFile, "", "pgp key passphrase")
	buildCmd.Flags().String(flagBuildKeyId, "", "pgp key Id")
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
		return errors.New("no package found")
	}

	return rpm.SignPackages(signatureKey, rpmPackages...)
}

func getSignatureKey(cmd *cobra.Command) *rpm.PgpKey {
	key, err := cmd.Flags().GetString(flagBuildKey)
	if err != nil {
		return nil
	}

	keyPassphraseFile, err := cmd.Flags().GetString(flagBuildKeyPassphraseFile)
	if err != nil {
		return nil
	}

	keyId, err := cmd.Flags().GetString(flagBuildKeyId)
	if err != nil {
		return nil
	}

	if len(key) == 0 && len(keyId) == 0 {
		return nil
	}

	if len(key) != 0 {
		if err = rpm.IsPathFile(key); err != nil {
			return nil
		}
	}

	if len(keyPassphraseFile) != 0 {
		if err = rpm.IsPathFile(keyPassphraseFile); err != nil {
			return nil
		}
	}

	return &rpm.PgpKey{
		KeyPath:           key,
		KeyPassphraseFile: keyPassphraseFile,
		KeyId:             keyId,
	}
}
