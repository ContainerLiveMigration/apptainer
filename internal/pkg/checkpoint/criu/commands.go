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
			ContainerStatePath,
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
			ContainerStatePath,
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
			ContainerStatePath,
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
			ContainerStatePath,
			"--work-dir",
			ContainerStatePath,
			"--log-file",
			"restore.log",
		}
	}
}
