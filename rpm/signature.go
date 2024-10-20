package rpm

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path"
)

var (
	signatureTools []string = []string{
		"rpmsign",
		"gpg-agent",
		"gpgconf",
		"gpg",
	}
)

type GpgKey struct {
	Key, KeyPassphrase, KeyId string
}

type GpgHome struct {
	Path string
}

func (gh *GpgHome) Close() {
	environ := os.Environ()
	environ = append(environ, fmt.Sprintf("GNUPGHOME=%s", gh.Path))
	shutdownGpgAgentCmd := exec.Command("gpgconf", "--kill", "--gpg-agent")
	shutdownGpgAgentCmd.Env = environ
	err := shutdownGpgAgentCmd.Run()
	if err != nil {
		fmt.Printf("%s\n", err.Error())
	}

	err = os.RemoveAll(gh.Path)
	if err != nil {
		fmt.Printf("%s\n", err.Error())
	}
}

func configureGpgAgent(gpgHome string) error {
	gpgAgentConfPath := path.Join(gpgHome, "gpg-agent.conf")
	gpgAgentConf, err := os.OpenFile(gpgAgentConfPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0700)
	if err != nil {
		return err
	}
	defer gpgAgentConf.Close()
	_, err = gpgAgentConf.WriteString("allow-loopback-pinentry\n")
	return err
}

func configureGpgHome(gpgHome string, signatureKey *GpgKey) error {
	environ := os.Environ()
	environ = append(environ, fmt.Sprintf("GNUPGHOME=%s", gpgHome))

	if err := configureGpgAgent(gpgHome); err != nil {
		return err
	}

	gpgImportCmd := exec.Command("gpg", "--batch", "--yes", "--import", signatureKey.Key)
	gpgImportCmd.Env = environ
	err := gpgImportCmd.Run()
	return err
}

func SetupGpgKey(signatureKey *GpgKey) (*GpgHome, error) {
	gpgHome, err := os.MkdirTemp("", "rpmtool.gpg.")
	if err != nil {
		return nil, err
	}

	err = configureGpgHome(gpgHome, signatureKey)
	if err != nil {
		return nil, err
	}

	return &GpgHome{
		Path: gpgHome,
	}, nil
}

func SignPackages(gpgHome *GpgHome, signatureKey *GpgKey, rpmPackages ...string) error {
	for _, rpmPackage := range rpmPackages {
		fmt.Printf("Signing %s\n", rpmPackage)
		args := []string{
			"--define", fmt.Sprintf("_gpg_name %s", signatureKey.KeyId),
			"--define", fmt.Sprintf("_gpg_path %s", gpgHome.Path),
			"--define", fmt.Sprintf("_gpg_sign_cmd_extra_args --batch --yes --pinentry-mode loopback --passphrase '%s'", signatureKey.KeyPassphrase),
			"--addsign", rpmPackage,
		}

		rpmSignCmd := exec.Command("rpmsign", args...)
		err := rpmSignCmd.Run()
		if err != nil {
			return errors.Join(errors.New(fmt.Sprintf("Failed to sign package. path=%s", rpmPackage)), err)
		}
	}
	return nil
}
