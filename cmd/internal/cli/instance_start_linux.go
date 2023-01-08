// Copyright (c) Contributors to the Apptainer project, established as
//   Apptainer a Series of LF Projects LLC.
//   For website terms of use, trademark policy, privacy policy and other
//   project policies see https://lfprojects.org/policies
// Copyright (c) 2018-2020, Sylabs Inc. All rights reserved.
// This software is licensed under a 3-clause BSD license. Please consult the
// LICENSE.md file distributed with the sources of this project regarding your
// rights to use or distribute this software.

package cli

import (
	"strings"
	"syscall"

	"github.com/apptainer/apptainer/docs"
	"github.com/apptainer/apptainer/internal/app/apptainer"
	"github.com/apptainer/apptainer/pkg/cmdline"
	"github.com/apptainer/apptainer/pkg/sylog"
	"github.com/prometheus/procfs"
	"github.com/spf13/cobra"
)

func init() {
	addCmdInit(func(cmdManager *cmdline.CommandManager) {
		cmdManager.RegisterFlagForCmd(&instanceStartPidFileFlag, instanceStartCmd)
		cmdManager.RegisterFlagForCmd(&actionDMTCPLaunchFlag, instanceStartCmd)
		cmdManager.RegisterFlagForCmd(&actionDMTCPRestartFlag, instanceStartCmd)
		cmdManager.RegisterFlagForCmd(&actionCRIULaunchFlag, instanceStartCmd)
		cmdManager.RegisterFlagForCmd(&actionCRIURestartFlag, instanceStartCmd)
	})
}

// --pid-file
var instanceStartPidFile string

var instanceStartPidFileFlag = cmdline.Flag{
	ID:           "instanceStartPidFileFlag",
	Value:        &instanceStartPidFile,
	DefaultValue: "",
	Name:         "pid-file",
	Usage:        "write instance PID to the file with the given name",
	EnvKeys:      []string{"PID_FILE"},
}

// apptainer instance start
var instanceStartCmd = &cobra.Command{
	Args:                  cobra.MinimumNArgs(2),
	PreRun:                actionPreRun,
	DisableFlagsInUseLine: true,
	Run: func(cmd *cobra.Command, args []string) {
		image := args[0]
		name := args[1]

		a := append([]string{"/.singularity.d/actions/start"}, args[2:]...)
		setVM(cmd)
		if VM {
			execVM(cmd, image, a)
			return
		}
		
		// close some open fds to avoid criu dump error
		if (CRIULaunch != "") {
			procSelf, err := procfs.Self()
			if err != nil {
				sylog.Fatalf("can't open procfs, %e", err)
			}
			fds, err := procSelf.FileDescriptors()
			if err != nil {
				sylog.Fatalf("fail to get open files, %e", err)
			}
			targets, _ := procSelf.FileDescriptorTargets()
			sylog.Debugf("%d == %d, open fds are %v, targets: %v", len(fds), len(targets), fds, targets)
			for i, fd := range fds {
				sylog.Debugf("fd is %v, target is %v", fd, targets[i])
			}
			for i, fd := range fds {
				if fd > 2 && len(targets[i]) > 0 && targets[i][0] == '/' {
					if strings.Contains(targets[i], "cli-") {
						continue
					}
					if syscall.Close(int(fd)) != nil {
						sylog.Warningf("close fd %v %v failed, %e", fd, targets[i], err)
					}
				}
			}
		}
		execStarter(cmd, image, a, name)

		if instanceStartPidFile != "" {
			err := apptainer.WriteInstancePidFile(name, instanceStartPidFile)
			if err != nil {
				sylog.Warningf("Failed to write pid file: %v", err)
			}
		}
	},

	Use:     docs.InstanceStartUse,
	Short:   docs.InstanceStartShort,
	Long:    docs.InstanceStartLong,
	Example: docs.InstanceStartExample,
}
