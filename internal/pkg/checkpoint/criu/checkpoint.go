// Copyright (c) Contributors to the Apptainer project, established as
//   Apptainer a Series of LF Projects LLC.
//   For website terms of use, trademark policy, privacy policy and other
//   project policies see https://lfprojects.org/policies
// This software is licensed under a 3-clause BSD license. Please consult the
// LICENSE.md file distributed with the sources of this project regarding your
// rights to use or distribute this software.

package criu

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"syscall"

	apptainerConfig "github.com/apptainer/apptainer/pkg/runtime/engine/apptainer/config"
	"github.com/apptainer/apptainer/pkg/sylog"
	"github.com/apptainer/apptainer/pkg/util/copy"
)

type Entry struct {
	path string
}

func (e *Entry) GetPid() (string, error) {
	f, err := os.Open(filepath.Join(e.Path(), PidFile))
	if err != nil {
		return "", err
	}
	defer f.Close()

	// scan first line of file for port
	s := bufio.NewScanner(f)
	s.Scan()
	if s.Text() == "" {
		return "", fmt.Errorf("unable to scan port from checkpoint data")
	}

	return s.Text(), nil
}

func (e *Entry) getLogDir() string {
	return filepath.Join(e.path, "log")
}

func (e *Entry) SetLogFile(name string, uid int) (*os.File, *os.File, error) {
	path := e.getLogDir()
	stderrPath := filepath.Join(path, name+".err")
	stdoutPath := filepath.Join(path, name+".out")

	oldumask := syscall.Umask(0)
	defer syscall.Umask(oldumask)

	if err := os.MkdirAll(filepath.Dir(stderrPath), 0o700); err != nil {
		return nil, nil, err
	}
	if err := os.MkdirAll(filepath.Dir(stdoutPath), 0o700); err != nil {
		return nil, nil, err
	}

	stderr, err := os.OpenFile(stderrPath, os.O_RDWR|os.O_CREATE|os.O_APPEND|syscall.O_NOFOLLOW, 0o644)
	if err != nil {
		return nil, nil, err
	}

	stdout, err := os.OpenFile(stdoutPath, os.O_RDWR|os.O_CREATE|os.O_APPEND|syscall.O_NOFOLLOW, 0o644)
	if err != nil {
		return nil, nil, err
	}

	if uid != os.Getuid() || uid == 0 {
		if err := stderr.Chown(uid, os.Getgid()); err != nil {
			return nil, nil, err
		}
		if err := stdout.Chown(uid, os.Getgid()); err != nil {
			return nil, nil, err
		}
	}

	return stdout, stderr, nil
}

func (e *Entry) RollBackLogFile(name string) error {
	path := e.getLogDir()
	stderrPath := filepath.Join(path, name+".err")
	stderrBackPath := stderrPath + BackSuffix
	stdoutPath := filepath.Join(path, name+".out")
	stdoutBackPath := stdoutPath + BackSuffix
	if err := copy.CopyFileContents(stderrPath, stderrBackPath); err != nil {
		sylog.Debugf("copy from %s to %s failed, %e", stderrBackPath, stderrPath, err)
	}

	if err := copy.CopyFileContents(stdoutPath, stdoutBackPath); err != nil {
		sylog.Debugf("copy from %s to %s failed, %e", stdoutBackPath, stdoutPath, err)
	}
	return nil
}

// GetLogFilePaths returns the paths of log files containing
// .err, .out streams, respectively
func (e *Entry) GetLogFilePaths(name string) (string, string, error) {
	path  := e.getLogDir()
	logErrPath := filepath.Join(path, name+".err")
	logOutPath := filepath.Join(path, name+".out")

	return logErrPath, logOutPath, nil
}

// GetLogFilePaths returns the paths of log files containing
// .err, .out streams, respectively
func (e *Entry) GetLogFile(name string) (*os.File, *os.File, error) {
	logErrPath, logOutPath, _ := e.GetLogFilePaths(name)
	// var logOut, logErr *os.Filee
	logOut, err := os.OpenFile(logOutPath, os.O_RDWR, 0)
	if err != nil {
		sylog.Warningf("open logout %s failed, %e", logOutPath, err)
		return nil, nil, err
	}
	logErr, err := os.OpenFile(logOutPath, os.O_RDWR, 0)
	if err != nil {
		sylog.Warningf("open logErr %s failed, %e", logErrPath, err)
		return nil, nil, err
	}
	return logOut, logErr, nil
}

func (e *Entry) BindPath() apptainerConfig.BindPath {
	return apptainerConfig.BindPath{
		Source:      e.path,
		Destination: ContainerStatePath,
		Options: map[string]*apptainerConfig.BindOption{
			"rw": {},
		},
	}
}

func (e *Entry) Path() string {
	return e.path
}

func (e *Entry) Name() string {
	return filepath.Base(e.path)
}

type Manager interface {
	Create(string) (*Entry, error) // create checkpoint directory for criu state
	Get(string) (*Entry, error)    // ensure directory with criu state exists
	List() ([]*Entry, error)       // list checkpoint directories for criu state
	Delete(string) error           // delete checkpoint directory for criu state
}

type checkpointManager struct{}

func NewManager() Manager {
	return &checkpointManager{}
}

func (checkpointManager) Create(name string) (*Entry, error) {
	err := os.MkdirAll(filepath.Join(criuDir(), name), 0o700)
	if err != nil {
		return nil, err
	}

	return &Entry{filepath.Join(criuDir(), name)}, nil
}

func (checkpointManager) Get(name string) (*Entry, error) {
	if name == "" {
		return nil, fmt.Errorf("checkpoint name must not be empty")
	}

	_, err := os.Stat(filepath.Join(criuDir(), name))
	if err != nil {
		return nil, err
	}

	return &Entry{filepath.Join(criuDir(), name)}, nil
}

func (checkpointManager) List() ([]*Entry, error) {
	fis, err := ioutil.ReadDir(criuDir())
	if err != nil {
		return nil, err
	}

	var entries []*Entry
	for _, fi := range fis {
		if !fi.IsDir() {
			continue
		}

		entries = append(entries, &Entry{filepath.Join(criuDir(), fi.Name())})
	}

	return entries, nil
}

func (checkpointManager) Delete(name string) error {
	_, err := os.Stat(filepath.Join(criuDir(), name))
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("checkpoint %q not found", name)
		}
	}

	return os.RemoveAll(filepath.Join(criuDir(), name))
}
