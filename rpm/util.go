package rpm

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"os/user"
	"path"
	"strings"
)

func setupRpmTree() error {
	setupTreeCmd := exec.Command("rpmdev-setuptree")
	setupTreeCmd.Stdin = os.Stdin
	setupTreeCmd.Stdout = os.Stdout
	setupTreeCmd.Stderr = os.Stderr
	return setupTreeCmd.Run()
}

func getSourceDirectory() (string, error) {

	if err := setupRpmTree(); err != nil {
		return "", errors.Join(errors.New("Unable to setup rpm tree"), err)
	}

	rpmOutput, err := exec.Command("rpm", "--eval", "%{_sourcedir}").Output()
	if err != nil {
		return "", err
	}
	sourceDir := strings.TrimSpace(string(rpmOutput))
	info, err := os.Stat(sourceDir)
	if err != nil {
		return "", err
	}

	if !info.IsDir() {
		return "", errors.New(fmt.Sprintf("%s is not a directory", sourceDir))
	}

	return sourceDir, nil
}

func copyFile(destinationPath, sourcePath string) error {
	source, err := os.Open(sourcePath)
	if err != nil {
		return err
	}
	defer source.Close()

	destination, err := os.OpenFile(destinationPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer destination.Close()

	_, err = io.Copy(destination, source)
	return err
}

func downloadUrl(destinationPath string, sourceUrl *url.URL) error {
	info, err := os.Stat(destinationPath)
	if err == nil {
		if info.IsDir() {
			return errors.New("Can't download, destination path is a directory")
		}
		fmt.Printf("Download skipped. File already exists. path=%s\n", destinationPath)
		return nil
	}

	resp, err := http.Get(sourceUrl.String())
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.New(fmt.Sprintf("Failed to download. status=%s", resp.Status))
	}

	destination, err := os.OpenFile(destinationPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer destination.Close()

	_, err = io.Copy(destination, resp.Body)
	if err != nil {
		return err
	}

	return nil
}

func DownloadSources(specPath string) error {
	sourceDir, err := getSourceDirectory()
	if err != nil {
		return err
	}

	spec, err := ParseSpec(specPath)
	if err != nil {
		return err
	}
	defer spec.Close()
	sources, err := spec.Sources()
	if err != nil {
		return err
	}

	specDirectory := path.Dir(specPath)

	for _, source := range sources {
		sourceUrl, err := url.Parse(source.Path)
		if err != nil {
			return err
		}

		if sourceUrl.Scheme == "" {
			destinationPath := path.Join(sourceDir, source.Path)
			sourcePath := path.Join(specDirectory, source.Path)
			fmt.Printf("Copying %s to %s\n", sourcePath, destinationPath)
			if err = copyFile(destinationPath, sourcePath); err != nil {
				errors.Join(errors.New(fmt.Sprintf("Unable to copy source. destination=%s source=%s", destinationPath, sourcePath)), err)
			}
		} else if sourceUrl.Scheme == "http" || sourceUrl.Scheme == "https" {
			destinationPath := path.Join(sourceDir, source.FileName)
			fmt.Printf("Downloading %s to %s\n", sourceUrl.String(), destinationPath)
			if err = downloadUrl(destinationPath, sourceUrl); err != nil {
				return errors.Join(errors.New(fmt.Sprintf("Unable to copy source. destination=%s source=%s", destinationPath, sourceUrl.String())), err)
			}
		} else {
			return errors.New(fmt.Sprintf("Unexpected source path. scheme=%s path=%s", sourceUrl.Scheme, source.Path))
		}
	}

	return nil
}

func InstallBuildDependencies(specPath string, latest bool) error {
	currentUser, err := user.Current()
	if err != nil {
		return err
	}

	cmd := "dnf"
	args := []string{"builddep", "--assumeyes", "--spec", specPath}
	if currentUser.Uid != "0" {
		cmd = "sudo"
		args = append([]string{"dnf"}, args...)
	}

	if latest {
		args = append(args, "--refresh")
	}

	setupTreeCmd := exec.Command(cmd, args...)
	setupTreeCmd.Stdin = os.Stdin
	setupTreeCmd.Stdout = os.Stdout
	setupTreeCmd.Stderr = os.Stderr
	return setupTreeCmd.Run()
}
