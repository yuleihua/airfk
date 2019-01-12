package cmd

import (
	"context"
	"os/exec"
	"time"
)

func ExecCmd(cmd string, retry int) ([]byte, error) {
	var result []byte
	var err error
	for i := 0; i < retry; i++ {
		result, err = execCmdline(cmd)
		if err == nil {
			break
		}
	}
	return result, err
}

func ExecCmdWithTimeout(cmd string, retry, timeout int) ([]byte, error) {
	var result []byte
	var err error
	for i := 0; i < retry; i++ {
		result, err = execCmdlineWithTimeout(cmd, timeout)
		if err == nil {
			break
		}
	}
	return result, err
}

func ExecCmdFile(cmd string, retry int) ([]byte, error) {
	var result []byte
	var err error
	for i := 0; i < retry; i++ {
		result, err = execCmdfile(cmd)
		if err == nil {
			break
		}
	}
	return result, err
}

func ExecCmdFileWithTimeout(cmd string, retry, timeout int) ([]byte, error) {
	var result []byte
	var err error
	for i := 0; i < retry; i++ {
		result, err = execCmdfileWithTimeout(cmd, timeout)
		if err == nil {
			break
		}
	}
	return result, err
}

func execCmdline(c string) ([]byte, error) {
	cmd := exec.Command("/bin/sh", "-c", c)
	return cmd.CombinedOutput()
}

func execCmdlineWithTimeout(c string, timeout int) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "/bin/sh", "-c", c)
	return cmd.CombinedOutput()
}

func execCmdfile(cmdFile string) ([]byte, error) {
	cmd := exec.Command("/bin/sh", cmdFile)
	return cmd.CombinedOutput()
}

func execCmdfileWithTimeout(cmdFile string, timeout int) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "/bin/sh", cmdFile)
	return cmd.CombinedOutput()
}
