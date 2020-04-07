package annotations

import (
	"errors"
	"fmt"
)

type InvalidArgumentErr string

func (e InvalidArgumentErr) Error() string {
	return fmt.Sprintf("invalid argument value for path %q", string(e))
}
