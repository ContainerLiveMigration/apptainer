// Copyright (c) 2018, Sylabs Inc. All rights reserved.
// This software is licensed under a 3-clause BSD license. Please consult the
// LICENSE file distributed with the sources of this project regarding your
// rights to use or distribute this software.

package rpc

import "github.com/singularityware/singularity/src/pkg/util/loop"

type MkdirArgs struct {
	Path string
}

type LoopArgs struct {
	Image string
	Mode  int
	Info  loop.LoopInfo64
}

type MountArgs struct {
	Source     string
	Target     string
	Filesystem string
	Mountflags uintptr
	Data       string
}

type ChrootArgs struct {
	Root string
}
