// Copyright 2018 The huayulei_2003@hotmail.com Authors
// This file is part of the airfk library.
//
// The airfk library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The airfk library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the airfk library. If not, see <http://www.gnu.org/licenses/>.
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
	result, err := cmd.CombinedOutput()
	if err != nil {
		return nil, err
	}
	return result, nil
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
