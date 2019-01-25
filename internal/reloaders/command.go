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
		err    error
		result CommandReloader
		opts   CommandReloaderOpts
	)

	err = json.Unmarshal(entry, &opts)
	if err != nil {
		return result, err
	}

	file, err := os.Lstat(opts.Command)
	if err != nil {
		return result, fmt.Errorf("Cannot use %v as reloader. err=%v", opts.Command, err.Error())
	}

	if file.IsDir() {
		return result, fmt.Errorf("%v is a directory. It must be an executable file.")
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
	Command  string `json:"command"`
}

func (r CommandReloader) Reload() error {
	var (
		cmdStdout bytes.Buffer
		cmdStderr bytes.Buffer
		err       error
		status    syscall.WaitStatus
	)
	o := r.GetOpts().(CommandReloaderOpts)
	cmdSplit := strings.Split(o.Command, " ")
	args := cmdSplit[1:]

	cmd := exec.Command(cmdSplit[0], args...)
	cmd.Stdout = &cmdStdout
	cmd.Stderr = &cmdStderr

	if err := cmd.Start(); err != nil {
		log.Errorf("Cannot execute reloader command. err=%v", err.Error())
		return NewReloaderError().WithMessage(err.Error()).WithCode(1)
	}

	err = cmd.Wait()
	if err != nil {
		log.Errorf("Cannot wait for reloader command. err=%v", err.Error())
		return NewReloaderError().WithMessage(err.Error()).WithCode(1)
	}

	status, ok := cmd.ProcessState.Sys().(syscall.WaitStatus)
	if !ok {
		log.Errorf("Cannot convert command exit code.")
		return NewReloaderError().WithMessage(err.Error()).WithCode(1)
	}

	// the script should exit with the proper exit code
	if status.ExitStatus() != 0 {
		msg := fmt.Sprintf("Expected command exit code of 0 got %v", status.ExitStatus())
		log.Error(msg)
		return NewReloaderError().WithMessage(msg).WithCode(status.ExitStatus())
	}

	log.Debugf("Command::Reload(): stdout=%#v stderr=%#v", cmdStdout.String(), cmdStderr.String())
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
