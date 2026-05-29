//go:build windows

package main

import (
	"fmt"

	"golang.org/x/sys/windows"
)

func runAsAdmin(exePath string, args string) error {
	verbPtr, _ := windows.UTF16PtrFromString("runas")

	var targetExe string
	var targetArgs string

	switch args {
	case "reinstall":
		targetExe = "cmd.exe"
		targetArgs = fmt.Sprintf("/s /c \" %q stop && %q uninstall && %q install && %q start\"", exePath, exePath, exePath, exePath)

	default:
		targetExe = exePath
		targetArgs = args
	}

	exePtr, err := windows.UTF16PtrFromString(targetExe)
	if err != nil {
		return err
	}
	argsPtr, err := windows.UTF16PtrFromString(targetArgs)
	if err != nil {
		return err
	}
	cwdPtr, _ := windows.UTF16PtrFromString("")

	err = windows.ShellExecute(0, verbPtr, exePtr, argsPtr, cwdPtr, windows.SW_HIDE)
	if err != nil {
		return fmt.Errorf("ShellExecute error: %w", err)
	}

	return nil
}

func isAdmin() bool {
	var sid *windows.SID

	err := windows.AllocateAndInitializeSid(
		&windows.SECURITY_NT_AUTHORITY,
		2,
		windows.SECURITY_BUILTIN_DOMAIN_RID,
		windows.DOMAIN_ALIAS_RID_ADMINS,
		0, 0, 0, 0, 0, 0,
		&sid,
	)
	if err != nil {
		return false
	}
	defer windows.FreeSid(sid)

	token := windows.Token(0)
	member, err := token.IsMember(sid)
	if err != nil {
		return false
	}

	return member
}
