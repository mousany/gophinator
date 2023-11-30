package runtime

import (
	"os"
	"strings"
	"syscall"
	"unsafe"

	"github.com/sirupsen/logrus"
)

// socketpair creates a pair of connected sockets.
func socketpair() ([2]int, error) {
	sockets, err := syscall.Socketpair(syscall.AF_UNIX, syscall.SOCK_SEQPACKET, 0)
	if err != nil {
		return [2]int{}, err
	}
	for _, fd := range sockets {
		syscall.CloseOnExec(fd)
	}
	return sockets, nil
}

// subprocess is the main loop for the container.
func subprocess(r *Runtime, fd int) int {
	logrus.Infof("Starting container with command: %s %s", r.command, strings.Join(r.args, " "))
	return 0
}

// spawn creates a new process in a new namespace.
func spawn(r *Runtime, fd int) (uintptr, error) {
	r1, _, err := syscall.Syscall(syscall.SYS_CLONE, uintptr(syscall.SIGCHLD|syscall.CLONE_NEWNS|syscall.CLONE_NEWCGROUP|syscall.CLONE_NEWPID|syscall.CLONE_NEWIPC|syscall.CLONE_NEWNET|syscall.CLONE_NEWUTS), 0, 0)
	if err != 0 {
		return 0, syscall.Errno(err)
	}

	if r1 == 0 {
		os.Exit(subprocess(r, fd))
	}
	return r1, nil
}

// wait waits for the child process to exit.
func wait(pid uintptr) (int, error) {
	var stat syscall.WaitStatus
	_, _, err := syscall.Syscall(syscall.SYS_WAIT4, pid, uintptr(unsafe.Pointer(&stat)), 0)
	if err != 0 {
		return 0, syscall.Errno(err)
	}

	return stat.ExitStatus(), nil
}
