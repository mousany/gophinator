package runtime

import (
	"fmt"
	"strings"
	"syscall"

	"github.com/msaf1980/go-uname"
	"github.com/sirupsen/logrus"
)

type Runtime struct {
	command string
	args    []string
	uid     uint
	volume  string
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
	if major < 4.8 {
		return nil, ErrUnsupportedVersion
	}

	return &Runtime{
		command: command,
		args:    args,
		uid:     uid,
		volume:  volume,
	}, nil
}

// Run executes the container's command with the given arguments.
func (r *Runtime) Run() error {
	sockets, err := socketpair()
	if err != nil {
		return err
	}

	defer func() {
		for _, fd := range sockets {
			syscall.Close(fd)
		}
	}()

	pid, err := spawn(r, sockets[1])
	if err != nil {
		return err
	}
	logrus.Debugf("Spawned container with PID %d", pid)

	stat, err := wait(pid)
	if err != nil {
		return err
	}
	logrus.Infof("Container exited with status %d", stat)

	return nil
}

// String returns a string representation of the container.
func (c *Runtime) String() string {
	return fmt.Sprintf("%s %s", c.command, strings.Join(c.args, " "))
}
