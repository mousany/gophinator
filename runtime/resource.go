package runtime

import (
	"fmt"
	"os"
	"syscall"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

const (
	filesysPrefix  = "/tmp/gophinator."
	filesysOldRoot = "/oldroot."
)

// mountFilesys mounts the filesystem.
func mountFilesys(rootUUID string, volume string) error {
	root := filesysPrefix + rootUUID
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

// cleanupFilesys cleans up the filesystem.
func cleanupFilesys(rootUUID string) {
	os.RemoveAll(filesysPrefix + rootUUID)
}

const (
	setupNamespaceFail    = 0x0
	setupNamespaceSuccess = 0x1
)

// setupNamespace sets up the namespaces.
func setupNamespace(fd int) error {
	err := syscall.Unshare(syscall.CLONE_NEWUSER)
	if err != nil {
		logrus.Debugf("Unsharing user namespace is not supported: %s", err)
		err = syscall.Sendto(fd, []byte{setupNamespaceFail}, 0, nil)
		if err != nil {
			return err
		}
	} else {
		logrus.Debugf("Unsharing user namespace successfully")
		err = syscall.Sendto(fd, []byte{setupNamespaceSuccess}, 0, nil)
		if err != nil {
			return err
		}
	}

	recv := make([]byte, 1)
	_, _, err = syscall.Recvfrom(fd, recv, 0)
	if err != nil {
		return err
	}
	logrus.Debug("Mapping UID/GID from parent successfully")

	return nil
}

const (
	mapNamespaceOffset = 10000
	mapNamespaceLength = 2000
)

// mapNamespace maps the namespaces.
func mapNamespace(pid uintptr) error {
	proc := fmt.Sprintf("/proc/%d", pid)
	for _, file := range []string{"/uid_map", "/gid_map"} {
		fd, err := syscall.Creat(proc+file, 0755)
		if err != nil {
			return err
		}

		mapEntry := fmt.Sprintf("%d %d %d", 0, mapNamespaceOffset, mapNamespaceLength)
		_, err = syscall.Write(fd, []byte(mapEntry))
		if err != nil {
			return err
		}

		err = syscall.Close(fd)
		if err != nil {
			return err
		}
	}

	logrus.Debugf("Mapping UID/GID successfully")

	return nil
}

// switchNamespace switches the namespaces.
func switchNamespace(uid int) error {
	err := syscall.Setgroups([]int{uid})
	if err != nil {
		return err
	}
	err = syscall.Setregid(uid, uid)
	if err != nil {
		return err
	}
	err = syscall.Setreuid(uid, uid)
	if err != nil {
		return err
	}

	logrus.Debugf("Switching UID/GID to %d successfully", uid)
	return nil
}
