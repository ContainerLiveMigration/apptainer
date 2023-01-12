// Copyright (c) Contributors to the Apptainer project, established as
//   Apptainer a Series of LF Projects LLC.
//   For website terms of use, trademark policy, privacy policy and other
//   project policies see https://lfprojects.org/policies
// Copyright (c) 2018, Sylabs Inc. All rights reserved.
// This software is licensed under a 3-clause BSD license. Please consult the
// LICENSE.md file distributed with the sources of this project regarding your
// rights to use or distribute this software.

package main

import (
	"os"
	// "runtime/pprof"
	"runtime/trace"
	"strings"
	"time"

	"github.com/apptainer/apptainer/cmd/internal/cli"
	"github.com/apptainer/apptainer/internal/pkg/buildcfg"
	useragent "github.com/apptainer/apptainer/pkg/util/user-agent"
	"github.com/apptainer/apptainer/pkg/sylog"
)

var TraceFile* os.File

func main() {
	startTime := time.Now().UnixNano()
	cli.StartTime = startTime
	sylog.Infof("TIMESTAMP: start time %d, accumulate time 0", startTime)

	useragent.InitValue(buildcfg.PACKAGE_NAME, buildcfg.PACKAGE_VERSION)

	// f1, err := os.Create("cli-cpu.prof")
	// defer f1.Close()
	// if err != nil {
	// 	panic(err)
	// }
	// err = pprof.StartCPUProfile(f1)
	// if err != nil {
	// 	panic(err)
	// }
	// defer pprof.StopCPUProfile()

	// f2, err := os.Create("cli-mem.prof")
	// defer f2.Close()
	// if err != nil {
	// 	panic(err)
	// }
	// pprof.WriteHeapProfile(f2)

	// check whether args contain --trace option
	if len(os.Args) >= 2 && strings.HasPrefix(os.Args[1], "--trace") {
		fileName := "cli-" + os.Args[1][8:] + ".out"
		var err error
		TraceFile, err = os.Create(fileName)
		if err = trace.Start(TraceFile); err != nil {
			panic(err)
		}
		defer trace.Stop()
	}
	// In cmd/internal/cli/apptainer.go
	cli.ExecuteApptainer()
}
