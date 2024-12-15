package rpm

/*
#include <rpm/rpmmacro.h>
#include <rpm/rpmsign.h>

void rpmtool_rpmSignSetup(char *sqPath, char *sqSignCmd, char *extraArgs)
{
	rpmPushMacro(NULL, "__gpg", NULL, sqPath, RMIL_GLOBAL);
	rpmPushMacro(NULL, "__gpg_sign_cmd", NULL, sqSignCmd, RMIL_GLOBAL);
	rpmPushMacro(NULL, "_sq_sign_cmd_extra_args", NULL, extraArgs, RMIL_GLOBAL);
}

int rpmtool_rpmPkgSign(char *packagePath, char *keyId)
{
	struct rpmSignArgs signArgs = {.keyid = keyId};
	return rpmPkgSign(packagePath, &signArgs);
}

void rpmtool_rpmSignSetupClear()
{
	rpmPopMacro(NULL, "_sq_sign_cmd_extra_args");
	rpmPopMacro(NULL, "__gpg_sign_cmd");
	rpmPopMacro(NULL, "__gpg");
}
*/
import "C"

import (
	"errors"
	"fmt"
	"os/exec"
	"unsafe"
)

var (
	signatureTools = []string{
		"sq",
	}
	sqSignCmdMacro string = `%{shescape:%{__gpg}} %{__gpg} sign --signer %{_gpg_name} %{?_sq_sign_cmd_extra_args} --signature-file --output %{shescape:%{__signature_filename}} %{shescape:%{__plaintext_filename}}`
	sqSignWithKeyFileCmdMacro string = `%{shescape:%{__gpg}} %{__gpg} sign %{?_sq_sign_cmd_extra_args} --signature-file --output %{shescape:%{__signature_filename}} %{shescape:%{__plaintext_filename}}`
)

type PgpKey struct {
	KeyPath, KeyPassphraseFile, KeyId string
}

func SignPackages(key *PgpKey, rpmPackages ...string) error {
	sqPath, err := exec.LookPath("sq")
	if err != nil {
		return err
	}

	extraArgs := ""
	keyId := key.KeyId
	signCmdMacros := sqSignCmdMacro
	if len(key.KeyPath) != 0 {
		extraArgs = fmt.Sprintf("--signer-file '%s'", key.KeyPath)
		signCmdMacros = sqSignWithKeyFileCmdMacro
		// with key path ignore key id
		keyId = "dummy"
	}

	if len(key.KeyPassphraseFile) != 0 {
		extraArgs = fmt.Sprintf("%s --batch --password-file '%s'", extraArgs, key.KeyPassphraseFile)
	}

	cSqPath := C.CString(sqPath)
	defer C.free(unsafe.Pointer(cSqPath))
	cSqSignCmdMacro := C.CString(signCmdMacros)
	defer C.free(unsafe.Pointer(cSqSignCmdMacro))
	cKeyId := C.CString(keyId)
	defer C.free(unsafe.Pointer(cKeyId))


	cExtraArgs := C.CString(extraArgs)
	defer C.free(unsafe.Pointer(cExtraArgs))

	C.rpmtool_rpmSignSetup(cSqPath, cSqSignCmdMacro, cExtraArgs)

	for _, rpmPackage := range rpmPackages {
		fmt.Printf("Signing %s\n", rpmPackage)
		cRpmPackage := C.CString(rpmPackage)
		defer C.free(unsafe.Pointer(cRpmPackage))
		rc := C.rpmtool_rpmPkgSign(cRpmPackage, cKeyId)
		if rc != 0 {
			err = errors.New(fmt.Sprintf("Signing failed. rpm=%s", rpmPackage))
			break
		}
	}
	C.rpmtool_rpmSignSetupClear()
	return err
}
