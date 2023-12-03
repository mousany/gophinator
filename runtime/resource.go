package runtime

import (
	"fmt"
	"os"
	"syscall"

	"github.com/containerd/cgroups/v3/cgroup1"
	"github.com/google/uuid"
	"github.com/opencontainers/runtime-spec/specs-go"
	seccomp "github.com/seccomp/libseccomp-golang"
	"github.com/sirupsen/logrus"
)

const (
	filesysPrefix       = "/tmp/gophinator."
	filesysOldRoot      = "/oldroot."
	filesysMountFail    = 0x0
	filesysMountSuccess = 0x1
)

// mountFilesys mounts the filesystem.
func mountFilesys(rt *Runtime, fd int) error {
	root := filesysPrefix + rt.uuid
	err := os.MkdirAll(root, 0755)
	if err != nil {
		return err
	}
	logrus.Debugf("Creating root directory %s", root)

	err = syscall.Mount(rt.root, root, "", uintptr(syscall.MS_BIND|syscall.MS_PRIVATE), "")
	if err != nil {
		return err
	}
	logrus.Debugf("Mounting root %s to %s", rt.root, root)

	for _, volume := range rt.volumes {
		source := volume.Source
		target := root + "/" + volume.Target
		err = os.MkdirAll(target, 0755)
		if err != nil {
			return err
		}
		err = syscall.Mount(source, target, "", uintptr(syscall.MS_BIND|syscall.MS_PRIVATE), "")
		if err != nil {
			return err
		}
		logrus.Debugf("Mounting volume %s to %s", source, target)
	}

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

	err = syscall.Sendto(fd, []byte{filesysMountSuccess}, 0, nil)
	if err != nil {
		return err
	}
	logrus.Infof("Mount %s => %s => /", rt.root, root)
	for _, volume := range rt.volumes {
		logrus.Infof("Mount %s => %s => %s", volume.Source, root+"/"+volume.Target, volume.Target)
	}

	return nil
}

// cleanupFilesys cleans up the filesystem.
func cleanupFilesys(rootUUID string) {
	os.RemoveAll(filesysPrefix + rootUUID)
}

const (
	namespaceSetupFail    = 0x0
	namespaceSetupSuccess = 0x1
)

// setupNamespace sets up the namespaces.
func setupNamespace(fd int) error {
	err := syscall.Unshare(syscall.CLONE_NEWUSER)
	if err != nil {
		logrus.Debugf("Unsharing user namespace is not supported: %s", err)
		err = syscall.Sendto(fd, []byte{namespaceSetupFail}, 0, nil)
		if err != nil {
			return err
		}
	} else {
		logrus.Debugf("Unsharing user namespace successfully")
		err = syscall.Sendto(fd, []byte{namespaceSetupSuccess}, 0, nil)
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
	namespaceMapOffset = 10000
	namespaceMapLength = 2000
)

// mapNamespace maps the namespaces.
func mapNamespace(pid uintptr) error {
	proc := fmt.Sprintf("/proc/%d", pid)
	for _, file := range []string{"/uid_map", "/gid_map"} {
		fd, err := syscall.Creat(proc+file, 0755)
		if err != nil {
			return err
		}

		mapEntry := fmt.Sprintf("%d %d %d", 0, namespaceMapOffset, namespaceMapLength)
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

// setupSyscall sets up the seccomp syscall.
func setupSyscall() error {
	filter, err := seccomp.NewFilter(seccomp.ActAllow)
	if err != nil {
		return err
	}

	refusedSyscalls := &[]seccomp.ScmpSyscall{
		seccomp.ScmpSyscall(syscall.SYS_KEYCTL),
		seccomp.ScmpSyscall(syscall.SYS_ADD_KEY),
		seccomp.ScmpSyscall(syscall.SYS_REQUEST_KEY),
		seccomp.ScmpSyscall(syscall.SYS_MBIND),
		seccomp.ScmpSyscall(syscall.SYS_MIGRATE_PAGES),
		seccomp.ScmpSyscall(syscall.SYS_MOVE_PAGES),
		seccomp.ScmpSyscall(syscall.SYS_SET_MEMPOLICY),
		// seccomp.ScmpSyscall(syscall.SYS_USERFAULTFD),
		seccomp.ScmpSyscall(syscall.SYS_PERF_EVENT_OPEN),
	}
	for _, sc := range *refusedSyscalls {
		err = filter.AddRule(sc, seccomp.ActErrno.SetReturnCode(int16(syscall.EPERM)))
		if err != nil {
			return err
		}
	}

	refusedCondSyscalls := &[]struct {
		seccomp.ScmpSyscall
		uint
		uint64
	}{
		{seccomp.ScmpSyscall(syscall.SYS_CHMOD), 1, syscall.S_ISUID},
		{seccomp.ScmpSyscall(syscall.SYS_CHMOD), 1, syscall.S_ISGID},
		{seccomp.ScmpSyscall(syscall.SYS_FCHMOD), 1, syscall.S_ISUID},
		{seccomp.ScmpSyscall(syscall.SYS_FCHMOD), 1, syscall.S_ISGID},
		{seccomp.ScmpSyscall(syscall.SYS_FCHMODAT), 2, syscall.S_ISUID},
		{seccomp.ScmpSyscall(syscall.SYS_FCHMODAT), 2, syscall.S_ISGID},
		{seccomp.ScmpSyscall(syscall.SYS_UNSHARE), 0, syscall.CLONE_NEWUSER},
		{seccomp.ScmpSyscall(syscall.SYS_CLONE), 0, syscall.CLONE_NEWUSER},
		{seccomp.ScmpSyscall(syscall.SYS_IOCTL), 1, syscall.TIOCSTI},
	}
	for _, sc := range *refusedCondSyscalls {
		cond, err := seccomp.MakeCondition(
			sc.uint,
			seccomp.CompareMaskedEqual,
			sc.uint64,
			sc.uint64,
		)
		if err != nil {
			return err
		}
		err = filter.AddRuleConditional(
			sc.ScmpSyscall,
			seccomp.ActErrno.SetReturnCode(int16(syscall.EPERM)),
			[]seccomp.ScmpCondition{cond},
		)
		if err != nil {
			return err
		}
	}

	err = filter.Load()
	return err
}

// setupCgroup sets up the cgroup.
func setupCgroup(hostname string, pid uintptr) (cgroup1.Cgroup, error) {
	var (
		cgroupRLimit       uint64 = 64
		cgroupCPUShare     uint64 = 256
		cgroupMemoryLimit  int64  = 1024 * 1024 * 1024
		cgroupMemoryKernel int64  = 1024 * 1024 * 1024
		cgroupPidLimit     int64  = 64
		cgroupBlkIOWeight  uint16 = 50
	)

	err := syscall.Setrlimit(syscall.RLIMIT_NOFILE, &syscall.Rlimit{
		Cur: cgroupRLimit,
		Max: cgroupRLimit,
	})
	if err != nil {
		return nil, err
	}

	control, err := cgroup1.New(cgroup1.StaticPath(hostname), &specs.LinuxResources{
		CPU: &specs.LinuxCPU{
			Shares: &cgroupCPUShare,
		},
		Memory: &specs.LinuxMemory{
			Limit:  &cgroupMemoryLimit,
			Kernel: &cgroupMemoryKernel,
		},
		Pids: &specs.LinuxPids{
			Limit: cgroupPidLimit,
		},
		BlockIO: &specs.LinuxBlockIO{
			Weight: &cgroupBlkIOWeight,
		},
	})
	if err != nil {
		return nil, err
	}

	err = control.Add(cgroup1.Process{
		Pid: int(pid),
	})
	if err != nil {
		control.Delete()
		return nil, err
	}

	return control, nil
}

// cleanupCgroup cleans up the cgroup.
func cleanupCgroup(control cgroup1.Cgroup) {
	control.Delete()
}
