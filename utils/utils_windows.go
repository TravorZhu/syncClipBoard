package utils

import "os/exec"

func OpenDir(path string) {
	println(exec.Command(`cmd`, `/c`, `explorer`, path).String())

}
