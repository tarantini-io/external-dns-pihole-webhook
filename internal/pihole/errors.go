package pihole

import "errors"

var ErrNoPiholeServer = errors.New("no pihole server found in the environment or flags")
