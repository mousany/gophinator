package runtime

import (
	"os"
	"strings"
	"syscall"
	"unsafe"

	"github.com/google/uuid"
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

// closeSocketPair closes a pair of connected sockets.
func closeSocketPair(sockets [2]int) {
	for _, fd := range sockets {
		syscall.Close(fd)
	}
}

// mountFilesys mounts the filesystem.
func mountFilesys(volume string) error {
	const (
		filesysPrefix  = "/tmp/gophinator."
		filesysOldRoot = "/oldroot."
	)

	root := filesysPrefix + uuid.NewString()
	err := os.MkdirAll(root, 0755)
	if err != nil {
		return err
	}
	logrus.Debugf("Creating root directory %s", root)

	err = syscall.Mount(volume, root, "", uintptr(syscall.MS_BIND|syscall.MS_PRIVATE), "")
	if err != nil {
		return err
	}
	logrus.Debugf("Mounting volume %s to %s", volume, root)

	uid := uuid.NewString()
	oldRoot := root + filesysOldRoot + uid
	err = os.MkdirAll(oldRoot, 0755)
	if err != nil {
		return err
	}
	logrus.Debugf("Creating old root directory %s", oldRoot)

	err = syscall.PivotRoot(root, oldRoot)
	if err != nil {
		return err
	}
	logrus.Debugf("Pivoting root to %s", root)

	err = os.Chdir("/")
	if err != nil {
		return err
	}
	err = syscall.Unmount(filesysOldRoot+uid, syscall.MNT_DETACH)
	if err != nil {
		return err
	}
	err = os.RemoveAll(filesysOldRoot + uid)
	if err != nil {
		return err
	}
	logrus.Debugf("Unmounting old root")

	logrus.Infof("Mount %s => %s => /", volume, root)

	return nil
}

// unmountFilesys unmounts the filesystem.
func unmountFilesys(volume string) {

}

// childDaemon is the main loop for the container.
func childDaemon(r *Runtime, _ int) int {
	logrus.Infof("Starting container with command: %s %s", r.command, strings.Join(r.args, " "))
	err := syscall.Sethostname([]byte(r.hostname))
	if err != nil {
		logrus.Errorf("Fail to set hostname: %s", err)
		return -1
	}
	err = mountFilesys(r.volume)
	if err != nil {
		logrus.Errorf("Fail to mount filesystem: %s", err)
		return -1
	}

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
