package rpm

/*
#include <rpm/rpmmacro.h>

char *rpmtool_rpmExpand(char *arg)
{
	return rpmExpand(arg, NULL);
}
*/
import "C"

import (
	"errors"
	"fmt"
	"unsafe"
)

var (
	macroBuildDir  = "%{_builddir}"
	macroRpmDir    = "%{_rpmdir}"
	macroSourceDir = "%{_sourcedir}"
	macroSpecDir   = "%{_specdir}"
	macroSrpmDir   = "%{_srcrpmdir}"
)

func ExpandMacro(macro string) (string, error) {
	cMacro := C.CString(macro)
	defer C.free(unsafe.Pointer(cMacro))
	cValue := C.rpmtool_rpmExpand(cMacro)
	if cValue == nil {
		return "", errors.New(fmt.Sprintf("unable to expand macros %s", macro))
	}
	defer C.free(unsafe.Pointer(cValue))
	value := C.GoString(cValue)
	return value, nil
}
