//go:build !windows && !darwin

package util

func OpenDir(path string) {
	exec.Command(`cmd`, `/c`, `explorer`, path)
}
