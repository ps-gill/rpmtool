package rpm

/*
#cgo LDFLAGS: -lrpmbuild -lrpm
#include <rpm/rpmbuild.h>
*/
import "C"

var (
	rpmInitialized bool
)

func init() {
	r := int(C.rpmReadConfigFiles(nil, nil))
	if r != 0 {
		return
	}
	rpmInitialized = true
}

func CloseRpmLib() {
	if !rpmInitialized {
		return
	}
	C.rpmFreeRpmrc()
	rpmInitialized = false
}
