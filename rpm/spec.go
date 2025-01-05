package rpm

/*
#include <rpm/rpmbuild.h>
#include <rpm/rpmlog.h>

int rpmtool_rpmSpecBuild(char * specFile, int buildSrpm)
{
	// replicate rpmbuild
	int rc = 1;
	rpmSetVerbosity(RPMLOG_INFO);
	rpmts ts = rpmtsCreate();
	rpmtsSetRootDir(ts, rpmcliRootDir);
	rpmtsSetFlags(ts, rpmtsFlags(ts) | RPMTRANS_FLAG_NOPLUGINS);
	struct rpmBuildArguments_s ba;
	if (buildSrpm)
	{
		ba.buildAmount = RPMBUILD_PACKAGESOURCE;
	}
	else
	{
		ba.buildAmount = RPMBUILD_PACKAGESOURCE
			| RPMBUILD_PACKAGEBINARY
			| RPMBUILD_CLEAN
			| RPMBUILD_RMBUILD
			| RPMBUILD_INSTALL
			| RPMBUILD_CHECK
			| RPMBUILD_BUILD
#ifndef RPMTOOL_DISABLE_RPMBUILD_CONF
			| RPMBUILD_CONF
#endif
			| RPMBUILD_BUILDREQUIRES
			| RPMBUILD_DUMPBUILDREQUIRES
			| RPMBUILD_CHECKBUILDREQUIRES
			| RPMBUILD_PREP
#ifndef RPMTOOL_DISABLE_RPMBUILD_MKBUILDDIR
			| RPMBUILD_MKBUILDDIR
#endif
			;
	}
	ba.rootdir = rpmcliRootDir;
	ba.cookie = NULL;

	rpmVSFlags vsflags, ovsflags;
	vsflags = rpmExpandNumeric("%{_vsflags_build}") | rpmcliVSFlags;
	ovsflags = rpmtsSetVSFlags(ts, vsflags);

	rpmSpecFlags spec_flags = 0;
#ifndef RPMTOOL_DISABLE_RPMSPEC_NOFINALIZE
	spec_flags = RPMSPEC_NOFINALIZE;
#endif

	rpmSpec spec = rpmSpecParse(specFile, spec_flags, NULL);
	if (spec == NULL)
	{
		rpmtsFree(ts);
		return rc;
	}

#ifndef RPMTOOL_DISABLE_BUILDROOTDIR
	if (rpmMkdirs(rpmcliRootDir, "%{_topdir}:%{_builddir}:%{_rpmdir}:%{_srcrpmdir}:%{_buildrootdir}"))
#else
	if (rpmMkdirs(rpmcliRootDir, "%{_topdir}:%{_builddir}:%{_rpmdir}:%{_srcrpmdir}"))
#endif
	{
		rpmSpecFree(spec);
		rpmtsFree(ts);
		return rc;
	}

	rc = rpmSpecBuild(ts, spec, &ba);
	rpmSpecFree(spec);
	rpmtsFree(ts);
	return rc;
}

*/
import "C"

import (
	"errors"
	"unsafe"
)

var (
	buildTools = []string{
		"dnf",
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
	buildSrpm := 0
	if srpm {
		buildSrpm = 1
	}

	cSpecPath := C.CString(specPath)
	defer C.free(unsafe.Pointer(cSpecPath))

	rc := C.rpmtool_rpmSpecBuild(cSpecPath, C.int(buildSrpm))

	if rc != 0 {
		return errors.New("build failed")
	}
	return nil
}
