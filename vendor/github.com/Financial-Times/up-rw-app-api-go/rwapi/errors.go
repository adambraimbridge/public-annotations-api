package rwapi

import (
	"fmt"
)

type ConstraintOrTransactionError struct {
	Message string
	Details []string
}

func (e ConstraintOrTransactionError) Error() string {
	if len(e.Details) == 0 {
		return e.Message
	}
	msg := fmt.Sprintf("%s\ndetails: {\n", e.Message)
	for _, d := range e.Details {
		msg = fmt.Sprintf("%s\t'%s'\n", msg, d)
	}
	return fmt.Sprintf("%s}\n", msg)
}
