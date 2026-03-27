package consolemgr

import (
	"os"
	"os/exec"
	"syscall"

	"golang.org/x/sys/windows"
)

var ConsoleProcess *exec.Cmd

func OpenConsole() error {
	if ConsoleProcess != nil {
		return nil
	}

	cmd := exec.Command("cmd.exe")
	cmd.SysProcAttr = &syscall.SysProcAttr{CreationFlags: windows.CREATE_NEW_CONSOLE}
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Start()
	if err != nil {
		return err
	}

	ConsoleProcess = cmd
	return nil
}

func AttachStdIO() error {

	return nil
}

func CloseConsole() error {
	if ConsoleProcess != nil && ConsoleProcess.Process != nil {
		err := ConsoleProcess.Process.Kill()
		ConsoleProcess = nil
		return err
	}
	return nil
}

func ConsoleActive() bool {
	return ConsoleProcess != nil
}
