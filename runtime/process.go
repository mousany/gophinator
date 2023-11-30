package runtime

import (
	"os"
	"strings"
	"syscall"
	"unsafe"

	"github.com/sirupsen/logrus"
)

// newSocketPair creates a pair of connected sockets.
func newSocketPair() ([2]int, error) {
	sockets, err := syscall.Socketpair(syscall.AF_UNIX, syscall.SOCK_SEQPACKET, 0)
	if err != nil {
		return [2]int{}, err
	}
	for _, fd := range sockets {
		syscall.CloseOnExec(fd)
	}
	return sockets, nil
}

// childDaemon is the main loop for the container.
func childDaemon(r *Runtime, _ int) int {
	logrus.Infof("Starting container with command: %s %s", r.command, strings.Join(r.args, " "))
	return 0
}

// spawnChild creates a new process in a new namespace.
func spawnChild(r *Runtime, fd int) (uintptr, error) {
	r1, _, err := syscall.Syscall(
		syscall.SYS_CLONE,
		uintptr(
			syscall.SIGCHLD|
				syscall.CLONE_NEWNS|
				syscall.CLONE_NEWCGROUP|
				syscall.CLONE_NEWPID|
				syscall.CLONE_NEWIPC|
				syscall.CLONE_NEWNET|
				syscall.CLONE_NEWUTS),
		0, 0)
	if err != 0 {
		return 0, err
	}

	if r1 == 0 {
		os.Exit(childDaemon(r, fd))
	}
	return r1, nil
}

// wait4Child waits for the child process to exit.
func wait4Child(pid uintptr) (int, error) {
	var stat syscall.WaitStatus
	_, _, err := syscall.Syscall(syscall.SYS_WAIT4, pid, uintptr(unsafe.Pointer(&stat)), 0)
	if err != 0 {
		return 0, err
	}

	return stat.ExitStatus(), nil
}
