package core

import "fmt"

const (
	Follower = iota + 1
	Candidate
	Leader
)

type Role int

func (r Role) String() string {
	switch r {
	case Follower:
		return "Follower"
	case Candidate:
		return "Candidate"
	case Leader:
		return "Leader"
	}
	return fmt.Sprintf("Role(%d)", int(r))
}
