package osxkeychain

/*
#cgo CFLAGS: -mmacosx-version-min=10.6 -D__MAC_OS_X_VERSION_MAX_ALLOWED=1060
#cgo LDFLAGS: -framework CoreFoundation -framework Security

#include <stdlib.h>
#include <CoreFoundation/CoreFoundation.h>
#include <Security/Security.h>
*/
import "C"

import (
	"fmt"
	"unsafe"
)

type GenericPassword struct {
	ServiceName string
	AccountName string
	Password    string
}

type _OSStatus C.OSStatus

// TODO: Fill this out.
const (
	errDuplicateItem _OSStatus = C.errSecDuplicateItem
)

type keychainError struct {
	errCode C.OSStatus
}

func newKeychainError(errCode C.OSStatus) error {
	if errCode == C.noErr {
		return nil
	}
	return &keychainError{errCode}
}

func (ke *keychainError) getErrCode() _OSStatus {
	return _OSStatus(ke.errCode)
}

func (ke *keychainError) Error() string {
	errorMessageCFString := C.SecCopyErrorMessageString(ke.errCode, nil)
	defer C.CFRelease(C.CFTypeRef(errorMessageCFString))

	errorMessageCString := C.CFStringGetCStringPtr(errorMessageCFString, C.kCFStringEncodingASCII)

	if errorMessageCString != nil {
		return C.GoString(errorMessageCString)
	}

	return fmt.Sprintf("keychainError with unknown error code %d", ke.errCode)
}

func AddGenericPassword(pass *GenericPassword) error {
	cpassword := C.CString(pass.Password)
	defer C.free(unsafe.Pointer(cpassword))
	var itemRef C.SecKeychainItemRef

	errCode := C.SecKeychainAddGenericPassword(
		nil, // default keychain
		C.UInt32(len(pass.ServiceName)),
		C.CString(pass.ServiceName),
		C.UInt32(len(pass.AccountName)),
		C.CString(pass.AccountName),
		C.UInt32(len(pass.Password)),
		unsafe.Pointer(cpassword),
		&itemRef,
	)

	return newKeychainError(errCode)
}

func FindGenericPassword(pass *GenericPassword) (*GenericPassword, error) {
	resp := *pass
	var cpassword unsafe.Pointer
	var cpasslen C.UInt32
	var itemRef C.SecKeychainItemRef

	errCode := C.SecKeychainFindGenericPassword(
		nil, // default keychain
		C.UInt32(len(pass.ServiceName)),
		C.CString(pass.ServiceName),
		C.UInt32(len(pass.AccountName)),
		C.CString(pass.AccountName),
		&cpasslen,
		&cpassword,
		&itemRef,
	)

	if ke := newKeychainError(errCode); ke != nil {
		return nil, ke
	}
	defer C.CFRelease(C.CFTypeRef(itemRef))
	defer C.SecKeychainItemFreeContent(nil, cpassword)

	buf := C.GoStringN((*C.char)(cpassword), C.int(cpasslen))
	resp.Password = string(buf)

	return &resp, nil
}

func FindAndRemoveGenericPassword(pass *GenericPassword) error {
	var itemRef C.SecKeychainItemRef

	errCode := C.SecKeychainFindGenericPassword(
		nil, // default keychain
		C.UInt32(len(pass.ServiceName)),
		C.CString(pass.ServiceName),
		C.UInt32(len(pass.AccountName)),
		C.CString(pass.AccountName),
		nil,
		nil,
		&itemRef,
	)

	if ke := newKeychainError(errCode); ke != nil {
		return ke
	}

	defer C.CFRelease(C.CFTypeRef(itemRef))

	errCode = C.SecKeychainItemDelete(itemRef)
	return newKeychainError(errCode)

	return nil
}
