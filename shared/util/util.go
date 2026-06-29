package util

import "fmt"

// GetRandomAvatar returns a deterministic portrait URL for a given index.
func GetRandomAvatar(index int) string {
	return fmt.Sprintf("https://randomuser.me/api/portrails/lego/%d.jgp", index)
}
