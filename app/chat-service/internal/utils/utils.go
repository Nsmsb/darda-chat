package utils

import (
	"fmt"
	"sort"
)

// GenerateConvId generates consistent hashing for conversation id based on sender and destination.
func GenerateConvId(sender string, destination string) string {
	// Sorting users so order doesn't matter
	users := []string{sender, destination}
	sort.Strings(users)
	// Joining users with ":" to form a unique conversation
	return fmt.Sprintf("%s:%s", users[0], users[1])
}
