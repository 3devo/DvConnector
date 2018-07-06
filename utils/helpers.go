package utils

import "github.com/google/uuid"

// IsValidUUID is a function that returns true if the input is a valid uuid.
// And false if the input isn't a valid uuid
// https://stackoverflow.com/a/46315070
func IsValidUUID(u string) bool {
	_, err := uuid.Parse(u)
	return err == nil
}
