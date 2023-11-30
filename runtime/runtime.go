package runtime

import (
	"fmt"
	"strings"
	"syscall"

	"github.com/msaf1980/go-uname"
	"github.com/sirupsen/logrus"
)

const minimalKernelVersion = 4.8

type Runtime struct {
	command  string
	args     []string
	uid      uint
	volume   string
	hostname string
}

// New creates a new container with the given command and arguments.
func New(command string, args []string, uid uint, volume string) (*Runtime, error) {
	u, err := uname.New()
	if err != nil {
		return nil, err
	}
	logrus.Debugf("Detected OS: %s %s %s %s", u.Sysname(), u.Nodename(), u.KernelRelease(), u.Machine())
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

	return &Runtime{
		command:  command,
		args:     args,
		uid:      uid,
		volume:   volume,
		hostname: hostname,
	}, nil
}

// Run executes the container's command with the given arguments.
func (r *Runtime) Run() error {
	sockets, err := newSocketPair()
	if err != nil {
		return err
	}

	defer func() {
		for _, fd := range sockets {
			syscall.Close(fd)
		}
	}()

	pid, err := spawnChild(r, sockets[1])
	if err != nil {
		return err
	}
	logrus.Debugf("Spawned container with PID %d", pid)

	stat, err := wait4Child(pid)
	if err != nil {
		return err
	}
	logrus.Infof("Container exited with status %d", stat)

	return nil
}

// String returns a string representation of the container.
func (r *Runtime) String() string {
	return fmt.Sprintf("%s %s", r.command, strings.Join(r.args, " "))
}
