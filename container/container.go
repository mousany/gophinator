package container

import (
	"fmt"
	"strings"
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
	return nil
}

// String returns a string representation of the container.
func (c *Container) String() string {
	return fmt.Sprintf("%s %s", c.command, strings.Join(c.args, " "))
}
