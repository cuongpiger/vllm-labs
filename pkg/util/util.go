package util

import "time"

// MyDuration is the encoding.TextUnmarshaler interface for time.Duration
type MyDuration struct {
	time.Duration
}

// Contains searches if a string list contains the given string or not.
func Contains(list []string, strToSearch string) bool {
	for _, item := range list {
		if item == strToSearch {
			return true
		}
	}
	return false
}
