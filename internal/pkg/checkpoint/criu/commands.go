package criu

func CheckpointArgs(pid string, privileged bool) []string {
	if !privileged {
		return []string{
			"criu",
			"dump",
			"--unprivileged",
			"--tree",
			pid,
			"--images-dir",
			CheckpointImagePath,
			"--work-dir",
			ContainerStatePath,
			"--shell-job",
			"-v4",
			"--log-file",
			"dump.log",
		}
	} else {
		return []string{
			"criu",
			"dump",
			"--tree",
			pid,
			"--images-dir",
			CheckpointImagePath,
			"--work-dir",
			ContainerStatePath,
			"--shell-job",
			"-v4",
			"--log-file",
			"dump.log",
		}
	}
}

func RestoreArgs(privileged bool) []string {
	if !privileged {
		return []string{
			"criu",
			"restore",
			"--unprivileged",
			"--shell-job",
			"-v4",
			"--images-dir",
			CheckpointImagePath,
			"--work-dir",
			ContainerStatePath,
			"--log-file",
			"restore.log",
		}
	} else {
		return []string{
			"criu",
			"restore",
			"--shell-job",
			"-v4",
			"--images-dir",
			CheckpointImagePath,
			"--work-dir",
			ContainerStatePath,
			"--log-file",
			"restore.log",
		}
	}
}

func PageServerArgs(privileged bool) []string {
	if !privileged {
		return []string{
			"criu",
			"page-server",
			"--unprivileged",
			"-v4",
			"--port",
			PageServerPort,
			"--images-dir",
			CheckpointImagePath,
			"--work-dir",
			ContainerStatePath,
			"--log-file",
			"page-server.log",
		}
	} else {
		return []string{
			"criu",
			"page-server",
			"-v4",
			"--port",
			PageServerPort,
			"--images-dir",
			CheckpointImagePath,
			"--work-dir",
			ContainerStatePath,
			"--log-file",
			"page-server.log",
		}
	}
}

func CheckpointWithPageServerArgs(pid string, privileged bool, address string) []string {
	if !privileged {
		return []string{
			"criu",
			"dump",
			"--unprivileged",
			"--tree",
			pid,
			"--images-dir",
			CheckpointImagePath,
			"--work-dir",
			ContainerStatePath,
			"--shell-job",
			"-v4",
			"--log-file",
			"page-server.log",
			"--page-server",
			"--address",
			address,
			"--port",
			PageServerPort,
		}
	} else {
		return []string{
			"criu",
			"dump",
			"--tree",
			pid,
			"--images-dir",
			CheckpointImagePath,
			"--work-dir",
			ContainerStatePath,
			"--shell-job",
			"-v4",
			"--log-file",
			"page-server.log",
			"--page-server",
			"--address",
			address,
			"--port",
		}
	}
}
