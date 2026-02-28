package runtime

import "syscall"

const (
	ioctlGetTermios = syscall.TIOCGETA
	ioctlSetTermios = syscall.TIOCSETA
)
