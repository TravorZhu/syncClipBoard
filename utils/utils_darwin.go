package utils

import "os/exec"

func OpenDir(path string) {
	println(exec.Command(`open`, path).String())
}
