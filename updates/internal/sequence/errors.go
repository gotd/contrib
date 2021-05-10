package sequence

import (
	"fmt"

	"golang.org/x/xerrors"
)

// ErrGap says that there is a gap in the sequence
// that must be recovered manually.
var ErrGap = xerrors.Errorf("gap")

// ResultError allows apply-functions to pass errors
// through the box without affecting it.
type ResultError struct {
	Err error
}

// Unwrap implements error unwrap interface.
func (e *ResultError) Unwrap() error { return e.Err }

func (e *ResultError) Error() string {
	return fmt.Sprintf("error: %s", e.Err)
}
