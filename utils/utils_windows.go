package utils

import "os/exec"

func OpenDir(path string) {
	exec.Command(`cmd`, `/c`, `explorer`, path)
}
