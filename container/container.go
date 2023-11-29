package container

import (
	"fmt"
	"strings"

	"github.com/msaf1980/go-uname"
	"github.com/sirupsen/logrus"
)

type Container struct {
	command string
	args    []string
	uid     uint
	volume  string
}

// New creates a new container with the given command and arguments.
func New(command string, args []string, uid uint, volume string) *Container {
	return &Container{
		command: command,
		args:    args,
		uid:     uid,
		volume:  volume,
	}
}

// Run executes the container's command with the given arguments.
func (c *Container) Run() error {
	u, err := uname.New()
	if err != nil {
		return err
	}

	logrus.Debugf("Detected OS: %s %s %s %s", u.Sysname(), u.Nodename(), u.KernelRelease(), u.Machine())
	if u.Machine() != "x86_64" {
		return ErrUnsupportedArch
	}
	if u.Sysname() != "Linux" {
		return ErrUnsupportedOS
	}

	var major, minor float32
	_, err = fmt.Sscanf(u.KernelRelease(), "%f.%f", &major, &minor)
	if err != nil {
		return err
	}
	if major < 4.8 {
		return ErrUnsupportedVersion
	}

	return nil
}

// String returns a string representation of the container.
func (c *Container) String() string {
	return fmt.Sprintf("%s %s", c.command, strings.Join(c.args, " "))
}
