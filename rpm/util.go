package rpm

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"os/user"
	"path"
	"path/filepath"
	"strings"
)

func CheckTools(signature bool) error {
	tools := buildTools
	if signature {
		tools = append(tools, signatureTools...)
	}
	toolsNotFound := make([]string, 0)
	for _, tool := range tools {
		if _, err := exec.LookPath(tool); err != nil {
			toolsNotFound = append(toolsNotFound, tool)
		}
	}

	if len(toolsNotFound) != 0 {
		return errors.New(fmt.Sprintf("Required tools not found. [%s]", strings.Join(toolsNotFound, ",")))
	}

	return nil
}

func IsPathFile(path string) error {
	pathStat, err := os.Stat(path)
	if err != nil {
		return err
	}
	if pathStat.IsDir() {
		return errors.New(fmt.Sprintf("path is a directory. path=%s", path))
	}
	return nil
}

func GetTools(signature bool) []string {
	tools := buildTools
	if signature {
		tools = append(tools, signatureTools...)
	}
	return tools
}

func setupRpmTree() error {
	for _, dir := range []string{rpmTree.BuildDir, rpmTree.RpmDir, rpmTree.SourceDir, rpmTree.SpecDir, rpmTree.SrpmDir} {
		stat, err := os.Stat(dir)
		if err != nil && !os.IsNotExist(err) {
			return err
		}
		if stat == nil {
			return os.MkdirAll(dir, 0750)
		}
		if !stat.IsDir() {
			return errors.New(fmt.Sprintf("path already exists and is not a directory. %s", dir))
		}
	}
	return nil
}

func emptyDirectory(directoryPath string) error {
	d, err := os.Open(directoryPath)
	if err != nil {
		return err
	}
	defer d.Close()

	contents, err := d.Readdir(0)
	if err != nil && !errors.Is(err, io.EOF) {
		return err
	}

	for _, content := range contents {
		contentPath := path.Join(directoryPath, content.Name())
		err = os.RemoveAll(contentPath)
		if err != nil {
			return err
		}
	}

	return nil
}

func EmptyRpmDirectory() error {
	return emptyDirectory(rpmTree.RpmDir)
}

func EmptySrpmDirectory() error {
	return emptyDirectory(rpmTree.SrpmDir)
}

func getPackages(directoryPath string) ([]string, error) {
	packages := make([]string, 0)

	err := filepath.Walk(directoryPath, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if path == directoryPath || info.IsDir() {
			return nil
		}

		if strings.HasSuffix(path, ".rpm") {
			packages = append(packages, path)
		}

		return nil
	})

	return packages, err
}

func GetRpmPackages() ([]string, error) {
	return getPackages(rpmTree.RpmDir)
}

func GetSrpmPackages() ([]string, error) {
	return getPackages(rpmTree.SrpmDir)
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
			return errors.New("can't download, destination path is a directory")
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
	if err := setupRpmTree(); err != nil {
		return errors.Join(errors.New("unable to setup rpm tree"), err)
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
			destinationPath := path.Join(rpmTree.SourceDir, source.Path)
			sourcePath := path.Join(specDirectory, source.Path)
			fmt.Printf("Copying %s to %s\n", sourcePath, destinationPath)
			if err = copyFile(destinationPath, sourcePath); err != nil {
				errors.Join(errors.New(fmt.Sprintf("Unable to copy source. destination=%s source=%s", destinationPath, sourcePath)), err)
			}
		} else if sourceUrl.Scheme == "http" || sourceUrl.Scheme == "https" {
			destinationPath := path.Join(rpmTree.SourceDir, source.FileName)
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

	dnf5, err := isDnf5()
	if err != nil {
		return err
	}

	cmd := "dnf"
	args := []string{"builddep", "--assumeyes", "--spec", specPath}

	if dnf5 {
		args = []string{"builddep", "--assumeyes", specPath}
	}

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

func isDnf5() (bool, error) {
	dnfExec, err := exec.LookPath("dnf")
	if err != nil {
		return false, err
	}

	dnfExecInfo, err := os.Lstat(dnfExec)
	if err != nil {
		return false, err
	}

	if dnfExecInfo.Mode()&os.ModeSymlink == 0 {
		return false, nil
	}

	dnfDest, err := os.Readlink(dnfExec)
	if err != nil {
		return false, err
	}

	return strings.HasSuffix(dnfDest, "dnf5"), nil
}
