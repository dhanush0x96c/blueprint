package prompt

import (
	"fmt"
	"strconv"
	"strings"
)

// ValidateNonEmptyString validates that a string is not empty
func ValidateNonEmptyString(s string) error {
	if strings.TrimSpace(s) == "" {
		return fmt.Errorf("cannot be empty")
	}
	return nil
}

// ValidateInteger validates that a string is a valid integer
func ValidateInteger(s string) error {
	if _, err := strconv.Atoi(s); err != nil {
		return fmt.Errorf("must be a valid integer")
	}
	return nil
}
