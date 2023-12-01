package runtime

import (
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
