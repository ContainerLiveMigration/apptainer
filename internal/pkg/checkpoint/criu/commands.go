package criu

func CheckpointArgs(pid string) []string {
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
}

func RestoreArgs() []string {
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
}
