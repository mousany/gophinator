package runtime

import (
	"fmt"
	"strings"
	"syscall"

	"github.com/google/uuid"
	"github.com/msaf1980/go-uname"
	"github.com/sirupsen/logrus"
)

const minimalKernelVersion = 4.8

type Runtime struct {
	command  string
	args     []string
	uid      int
	volume   string
	hostname string
	uuid     string
}

// New creates a new container with the given command and arguments.
func New(command string, args []string, uid int, volume string) (*Runtime, error) {
	u, err := uname.New()
	if err != nil {
		return nil, err
	}
	logrus.Debugf("Detecting OS: %s %s %s %s", u.Sysname(), u.Nodename(), u.KernelRelease(), u.Machine())
	if u.Machine() != "x86_64" {
		return nil, ErrUnsupportedArch
	}
	if u.Sysname() != "Linux" {
		return nil, ErrUnsupportedOS
	}
	var major, minor float32
	_, err = fmt.Sscanf(u.KernelRelease(), "%f.%f", &major, &minor)
	if err != nil {
		return nil, err
	}
	if major < minimalKernelVersion {
		return nil, ErrUnsupportedVersion
	}

	hostname, err := newHostname()
	if err != nil {
		return nil, err
	}
	logrus.Debugf("Using hostname: %s", hostname)

	uuid := uuid.NewString()

	return &Runtime{
		command:  command,
		args:     args,
		uid:      uid,
		volume:   volume,
		hostname: hostname,
		uuid:     uuid,
	}, nil
}

// Run executes the container's command with the given arguments.
func (r *Runtime) Run() (int, error) {
	sockets, err := newSocketPair()
	if err != nil {
		return 0, err
	}
	logrus.Debugf("Creating socket pair: %d %d", sockets[0], sockets[1])
	defer cleanupSocketPair(sockets)

	pid, err := spawnChild(r, sockets[1])
	if err != nil {
		return 0, err
	}
	logrus.Debugf("Spawning container with PID %d", pid)
	recv := make([]byte, 1)
	_, _, err = syscall.Recvfrom(sockets[0], recv, 0)
	if err != nil {
		return 0, err
	}
	if recv[0] == filesysMountFail {
		logrus.Debugf("Mounting filesystem in child failed")
	} else {
		logrus.Debugf("Mounting filesystem in child successfully")
		defer cleanupFilesys(r.uuid)
	}

	// control, err := setupCgroup(r.hostname, pid)
	// if err != nil {
	// 	return 0, err
	// }
	// logrus.Debugf("Setting up cgroup: %s", control)
	// defer cleanupCgroup(control)

	_, _, err = syscall.Recvfrom(sockets[0], recv, 0)
	if err != nil {
		return 0, err
	}
	if recv[0] == namespaceSetupFail {
		logrus.Debugf("Unsharing user namespace from child failed")
	} else {
		logrus.Debugf("Unsharing user namespace from child successfully")
		err = mapNamespace(pid)
		if err != nil {
			return 0, err
		}
	}
	err = syscall.Sendto(sockets[0], []byte{0x0}, 0, nil)
	if err != nil {
		return 0, err
	}

	stat, err := waitChild(pid)
	if err != nil {
		return 0, err
	}

	return stat, nil
}

// String returns a string representation of the container.
func (r *Runtime) String() string {
	return fmt.Sprintf("%s %s", r.command, strings.Join(r.args, " "))
}
