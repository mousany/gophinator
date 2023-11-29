package runtime

import "errors"

var (
	ErrUnsupportedArch    = errors.New("unsupported architecture")
	ErrUnsupportedOS      = errors.New("unsupported operating system")
	ErrUnsupportedVersion = errors.New("unsupported kernel version")
)
