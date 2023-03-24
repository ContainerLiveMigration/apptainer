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
	path    string
	dirType ImgDirType
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

func (e *Entry) GetImgRealPath() (string, error) {
	if e.dirType == DiskType {
		return "", fmt.Errorf("cannot get image path for disk checkpoint")
	}
	imgDir := filepath.Join(e.path, "img")
	return readRealPath(imgDir)
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

// GetLogFilePaths returns the paths of log files containing
// .err, .out streams, respectively
func (e *Entry) SetRestoreLogFilePaths(name string, uid int) (*os.File, *os.File, error) {
	path := e.getLogDir()
	stderrPath := filepath.Join(path, name+".restore.err")
	stdoutPath := filepath.Join(path, name+".restore.out")

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
	path := e.getLogDir()
	logErrPath := filepath.Join(path, name+".err")
	logOutPath := filepath.Join(path, name+".out")

	return logErrPath, logOutPath, nil
}

// GetLogFilePaths returns the paths of log files containing
// .err, .out streams, respectively
func (e *Entry) GetRestoreLogFilePaths(name string) (string, string, error) {
	path := e.getLogDir()
	logErrPath := filepath.Join(path, name+".restore.err")
	logOutPath := filepath.Join(path, name+".restore.out")

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

// GetLogFilePaths returns the paths of log files containing
// .err, .out streams, respectively
func (e *Entry) GetRestoreLogFile(name string) (*os.File, *os.File, error) {
	logErrPath, logOutPath, _ := e.GetRestoreLogFilePaths(name)
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

func (e *Entry) BindPath() []apptainerConfig.BindPath {
	ret := []apptainerConfig.BindPath{
		{
			Source:      e.path,
			Destination: ContainerStatePath,
			Options: map[string]*apptainerConfig.BindOption{
				"rw": {},
			},
		},
	}
	if e.dirType == MemType {
		p, err := e.GetImgRealPath()
		if err != nil {
			sylog.Fatalf("get image real path failed, %e", err)
		}
		ret = append(ret, apptainerConfig.BindPath{
			Source:      p,
			Destination: CheckpointImagePath,
			Options: map[string]*apptainerConfig.BindOption{
				"rw": {},
			},
		})
	}
	return ret
}

func (e *Entry) Path() string {
	return e.path
}

func (e *Entry) Name() string {
	return filepath.Base(e.path)
}

func (e *Entry) GetConfigPath() string {
	return filepath.Join(e.path, "config.json")
}

func (e *Entry) Type() string {
	switch e.dirType {
	case DiskType:
		return "disk"
	case MemType:
		return "memory"
	default:
		return "unknown"
	}
}

func (e *Entry) Sync() error {
	if e.dirType != MemType {
		return nil
	}
	p, err := e.GetImgRealPath()
	if err != nil {
		return err
	}
	// if p doesn't exist, create it
	if _, err := os.Stat(p); os.IsNotExist(err) {
		if err := os.MkdirAll(p, 0o700); err != nil {
			return err
		}
	}
	return nil
}

type ImgDirType int

const (
	DiskType ImgDirType = iota
	MemType
)

type Manager interface {
	Create(string, ImgDirType) (*Entry, error) // create checkpoint directory for criu state
	Get(string) (*Entry, error)                // ensure directory with criu state exists
	Config(string, ImgDirType) error           // configure checkpoint directory type for criu state
	List() ([]*Entry, error)                   // list checkpoint directories for criu state
	Delete(string) error                       // delete checkpoint directory for criu state
}

type checkpointManager struct{}

func NewManager() Manager {
	return &checkpointManager{}
}

func (checkpointManager) Create(name string, t ImgDirType) (*Entry, error) {
	checkpointDir := filepath.Join(criuDir(), name)
	err := os.MkdirAll(checkpointDir, 0o700)
	if err != nil {
		return nil, err
	}
	err = createImgDir(checkpointDir, t)
	if err != nil {
		return nil, err
	}
	return &Entry{checkpointDir, t}, nil
}

func (checkpointManager) Get(name string) (*Entry, error) {
	if name == "" {
		return nil, fmt.Errorf("checkpoint name must not be empty")
	}

	checkpointDir := filepath.Join(criuDir(), name)
	imgDir := filepath.Join(checkpointDir, "img")
	_, err := os.Stat(checkpointDir)
	if err != nil {
		return nil, err
	}
	dirType, err := checkDirType(imgDir)
	if err != nil {
		return nil, fmt.Errorf("failed to check checkpoint directory type: %s", err)
	}
	return &Entry{checkpointDir, dirType}, nil
}

func (c checkpointManager) Config(name string, t ImgDirType) error {
	entry, err := c.Get(name)
	if err != nil {
		return fmt.Errorf("failed to get checkpoint directory %s: %s", name, err)
	}

	if entry.dirType == t {
		entry.Sync()
		return nil
	}
	checkpointDir := entry.path
	imgDir := filepath.Join(checkpointDir, "img")
	if entry.dirType == MemType {
		deleteRealImgDir(imgDir)
	}
	err = os.RemoveAll(imgDir)
	if err != nil {
		return fmt.Errorf("failed to remove checkpoint image directory %s: %s", imgDir, err)
	}
	return createImgDir(checkpointDir, t)
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
		checkpointDir := filepath.Join(criuDir(), fi.Name())
		imgDir := filepath.Join(checkpointDir, "img")
		dirType, err := checkDirType(imgDir)
		if err != nil {
			continue
		}
		sylog.Debugf("checkpoint dir %s, type %v", checkpointDir, dirType)
		entries = append(entries, &Entry{checkpointDir, dirType})
	}

	return entries, nil
}

func (checkpointManager) Delete(name string) error {
	checkpointDir := filepath.Join(criuDir(), name)
	imgDir := filepath.Join(checkpointDir, "img")
	_, err := os.Stat(checkpointDir)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("checkpoint %q not found", name)
		}
	}
	deleteRealImgDir(imgDir)
	return os.RemoveAll(checkpointDir)
}

func checkDirType(path string) (ImgDirType, error) {
	realPath := filepath.Join(path, RealPath)
	// check whether realPath exits
	f, err := os.Stat(realPath)
	if err != nil || f.Size() == 0 {
		return DiskType, nil
	}
	return MemType, nil
}

func createImgDir(checkpointDir string, dirType ImgDirType) error {
	imgDir := filepath.Join(checkpointDir, "img")
	err := os.Mkdir(imgDir, 0o700)
	if err != nil {
		return fmt.Errorf("failed to create checkpoint image directory %s: %s", imgDir, err)
	}
	if dirType == DiskType {
		return nil
	}
	f, err := os.Create(filepath.Join(imgDir, RealPath))
	if err != nil {
		return fmt.Errorf("failed to create real_path file: %s", err)
	}
	defer f.Close()
	src := filepath.Join(TmpfsPath, checkpointDir)
	_, err = f.WriteString(src)
	if err != nil {
		return fmt.Errorf("failed to write real_path file: %s", err)
	}
	err = os.MkdirAll(src, 0o700)
	if err != nil {
		return fmt.Errorf("failed to create tmpfs directory %s: %s", src, err)
	}
	return nil
}

func readRealPath(imgDir string) (string, error) {
	p := filepath.Join(imgDir, RealPath)
	f, err := os.Open(p)
	if err != nil {
		return "", fmt.Errorf("failed to open real_path file: %s", err)
	}
	defer f.Close()
	b, err := ioutil.ReadAll(f)
	if err != nil {
		return "", fmt.Errorf("failed to read real_path file %s: %s", p, err)
	}
	return string(b), nil
}

func deleteRealImgDir(imgDir string) error {
	realPath, err := readRealPath(imgDir)
	if err != nil {
		return err
	}
	return os.RemoveAll(realPath)
}
