package rpm

/*
#include <rpm/rpmbuild.h>
*/
import "C"

import (
	"errors"
	"os"
	"os/exec"
	"unsafe"
)

var (
	buildTools []string = []string{
		"dnf",
		"rpm",
		"rpmbuild",
	}
)

type Spec struct {
	rpmSpec C.rpmSpec
}

type SpecSource struct {
	FileName, Path string
}

func (rs *Spec) Close() {
	if rs.rpmSpec == nil {
		return
	}
	C.rpmSpecFree(rs.rpmSpec)
	rs.rpmSpec = nil
}

func (rs *Spec) Sources() ([]SpecSource, error) {
	sources := make([]SpecSource, 0)

	srcIterator := C.rpmSpecSrcIterInit(rs.rpmSpec)
	if srcIterator == nil {
		return sources, errors.New("unable to iterate sources")
	}
	defer C.rpmSpecSrcIterFree(srcIterator)

	specPackage := C.rpmSpecSrcIterNext(srcIterator)
	for specPackage != nil {
		sourceFileName := C.GoString(C.rpmSpecSrcFilename(specPackage, 0))
		sourcePath := C.GoString(C.rpmSpecSrcFilename(specPackage, 1))
		sources = append(sources, SpecSource{
			FileName: sourceFileName,
			Path:     sourcePath,
		})
		specPackage = C.rpmSpecSrcIterNext(srcIterator)
	}

	return sources, nil
}

func ParseSpec(path string) (*Spec, error) {
	if !rpmInitialized {
		return nil, errors.New("rpmlib failed to initialize")
	}
	cPath := C.CString(path)
	defer C.free(unsafe.Pointer(cPath))
	rpmSpec := C.rpmSpecParse(cPath, C.RPMSPEC_ANYARCH|C.RPMSPEC_FORCE, nil)
	if rpmSpec == nil {
		return nil, errors.New("failed to parse spec file")
	}

	return &Spec{
		rpmSpec: rpmSpec,
	}, nil
}

func Build(specPath string, srpm bool) error {
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
