package entity

import "time"

type State string

func (s State) Valid() bool {
	switch s {
	case StateAvailable, StateInUse, StateInactive:
		return true
	default:
		return false
	}
}

func (s State) String() string {
	switch s {
	case StateAvailable:
		return "available"
	case StateInUse:
		return "in-use"
	case StateInactive:
		return "inactive"
	default:
		return ""
	}
}

const (
	StateAvailable State = "available"
	StateInUse     State = "in-use"
	StateInactive  State = "inactive"
)

type Device struct {
	ID        string    `bson:"_id"`
	Name      string    `bson:"name"`
	Brand     string    `bson:"brand"`
	State     State     `bson:"state"`
	CreatedAt time.Time `bson:"created_on"`
	Version   int64     `bson:"version"`
}

type DevicePage struct {
	Items      []Device
	NextCursor string
}
