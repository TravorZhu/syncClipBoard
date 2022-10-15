package utils

import "os/exec"

func OpenDir(path string) {
	exec.Command(`open`, path)
}
