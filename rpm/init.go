package rpm

/*
#cgo LDFLAGS: -lrpmbuild -lrpmio -lrpmsign -lrpm
#include <rpm/rpmbuild.h>
*/
import "C"
import "log"

var (
	rpmInitialized bool
	rpmTree        *Tree
)

func init() {
	r := int(C.rpmReadConfigFiles(nil, nil))
	if r != 0 {
		return
	}
	rpmInitialized = true

	var err error
	rpmTree, err = GetTree()
	if err != nil {
		CloseRpmLib()
		log.Fatalf("init failed. %s\n", err.Error())
	}
}

func CloseRpmLib() {
	if !rpmInitialized {
		return
	}
	C.rpmFreeRpmrc()
	rpmInitialized = false
}
