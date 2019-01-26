/*
Copyright 2017 Adobe. All rights reserved.
This file is licensed to you under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License. You may obtain a copy
of the License at http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software distributed under
the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR REPRESENTATIONS
OF ANY KIND, either express or implied. See the License for the specific language
governing permissions and limitations under the License.
*/

package reloaders

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"

	log "github.com/sirupsen/logrus"
)

func NewCommandReloader(manager string, method string, entry []byte) (Reloader, error) {
	var (
		err      error
		result   CommandReloader
		opts     CommandReloaderOpts
		isString bool
	)

	err = json.Unmarshal(entry, &opts)
	if err != nil {
		return result, err
	}

	switch opts.Command.(type) {
	case string:
		isString = true
	case []interface{}:
		isString = false
	default:
		return result, fmt.Errorf("Command reloader cannot determine the command type [%T].", opts.Command)
	}

	if isString {
		cmd := opts.Command.(string)
		cmdSplit := strings.Split(cmd, " ")
		file, err := os.Lstat(cmdSplit[0])
		if err != nil {
			return result, fmt.Errorf("Cannot use %v as reloader. err=%v", cmd, err.Error())
		}

		if file.IsDir() {
			return result, fmt.Errorf("%v is a directory. It must be an executable file.")
		}
	} else {
		cmds := opts.Command.([]interface{})
		for _, cmd := range cmds {
			cmdSplit := strings.Split(cmd.(string), " ")
			file, err := os.Lstat(cmdSplit[0])
			if err != nil {
				return result, fmt.Errorf("Cannot use %v as reloader. err=%v", cmd, err.Error())
			}

			if file.IsDir() {
				return result, fmt.Errorf("%v is a directory. It must be an executable file.")
			}
		}
	}

	result.Manager = manager
	result.Method = method
	result.Opts = opts

	return result, err
}

type CommandReloader struct {
	Manager string              `json"-"`
	Counter int                 `json:"-"`
	Method  string              `mapstructure:"method" json:"method"`
	Opts    CommandReloaderOpts `json:"opts"`
}

type CommandReloaderOpts struct {
	Command interface{} `json:"command"`
}

type Command struct {
	RetCode int
	StdOut  string
	StdErr  string
}

func (r CommandReloader) Reload() error {
	var (
		err      error
		isString bool
	)

	o := r.GetOpts().(CommandReloaderOpts)
	switch o.Command.(type) {
	case string:
		isString = true
	case []interface{}:
		isString = false
	}

	if isString {
		res := runCommand(o.Command.(string))
		if res.RetCode != 0 {
			msg := fmt.Sprintf("Expected command exit code of 0 got %v", res.RetCode)
			log.Error(msg)
			return NewReloaderError().WithMessage(msg).WithCode(res.RetCode)
		}
		log.Debugf("Command::Reload(): SUCCESS stdout=%#v stderr=%#v", res.StdOut, res.StdErr)
	} else {
		cmds := o.Command.([]interface{})
		for _, cmd := range cmds {
			res := runCommand(cmd.(string))
			if res.RetCode != 0 {
				msg := fmt.Sprintf("Expected command exit code of 0 got %v", res.RetCode)
				log.Error(msg)
				return NewReloaderError().WithMessage(msg).WithCode(res.RetCode)
			}
			log.Debugf("Command::Reload(): SUCCESS stdout=%#v stderr=%#v", res.StdOut, res.StdErr)
		}
	}
	return err
}

func (r CommandReloader) GetMethod() string {
	return r.Method
}

func (r CommandReloader) GetOpts() ReloaderOpts {
	return r.Opts
}

func (r CommandReloader) SetOpts(o ReloaderOpts) bool {
	r.Opts = o.(CommandReloaderOpts)
	return true
}

func (r CommandReloader) SetCounter(c int) Reloader {
	r.Counter = c
	return r
}

func runCommand(command string) Command {
	var (
		cmdStdout bytes.Buffer
		cmdStderr bytes.Buffer
		err       error
		status    syscall.WaitStatus
	)
	cmdSplit := strings.Split(command, " ")
	args := cmdSplit[1:]

	cmd := exec.Command(cmdSplit[0], args...)
	cmd.Stdout = &cmdStdout
	cmd.Stderr = &cmdStderr

	if err := cmd.Start(); err != nil {
		return Command{RetCode: 1, StdErr: fmt.Sprintf("Cannot execute reloader command. err=%v", err.Error())}
	}

	err = cmd.Wait()
	if err != nil {
		return Command{RetCode: 1, StdErr: fmt.Sprintf("Cannot wait for reloader command. err=%v", err.Error())}
	}

	status, ok := cmd.ProcessState.Sys().(syscall.WaitStatus)
	if !ok {
		return Command{RetCode: 1, StdErr: fmt.Sprintf("Cannot convert command exit code.")}
	}

	return Command{RetCode: status.ExitStatus(),
		StdOut: cmdStdout.String(),
		StdErr: cmdStderr.String()}
}
