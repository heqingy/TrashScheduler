package fsm

import (
	"fmt"
	"time"

	"gopkg.in/yaml.v3"
)

// StateType is the type of state
type StateType string

// Values for the StateType
const (
	Inactive StateType = "inactive"
	Active   StateType = "active"
	Taken    StateType = "taken"
	Pending  StateType = "pending"
	Done     StateType = "done"
)

// EventType is event
type EventType string

// Values for the EventType
const (
	Time     EventType = "time"
	Complete EventType = "complete"
	Take     EventType = "take"
)

// SideEffectFunc is the side effect of activating a state
type SideEffectFunc func() error

// User stores user information
type User struct {
	Name      string
	Phone     []string
	CanPulled int
}

// State is the state
type State struct {
	Users       map[string]*User
	Taker       *string
	StateName   StateType
	LastTask    *time.Time
	LastNotify  *time.Time
	LastTransit *time.Time
}

// EventContext is the context for the event
type EventContext struct {
	Taker       string
	PulledCount int
	Time        time.Time
}

// Event is a struct
type Event struct {
	Type EventType
	EventContext
}

// Config is the config for FSM
type Config struct {
	CallInterval   time.Duration
	RemindInterval time.Duration
	NotifySvc      Notifier
}

// NewFSM generates a new FSM
func NewFSM(users map[string]*User, c Config) *FSM {
	return &FSM{
		Config: &c,
		State: &State{
			Users:     users,
			StateName: Inactive,
		},
	}
}

// FSM is the finite state machine
type FSM struct {
	*State
	*Config
}

// transitToNew reset some state
// after trransit
func (m *FSM) transitToNew(s StateType) {
	m.StateName = s
	m.LastNotify = nil
	m.LastTransit = &[]time.Time{time.Now()}[0]
}

func (m *FSM) Load(data []byte) error {
	err := yaml.Unmarshal(data, &m.State)
	if err != nil {
		return err
	}
	return nil
}

func (m *FSM) Dump() ([]byte, error) {
	return yaml.Marshal(m.State)
}

// Transit transits the state given current state and event
func (m *FSM) Transit(e *Event) error {
	var err error
	t := e.Time
	switch m.StateName {
	case Inactive:
		switch e.Type {
		case Time:
			if ((t.Weekday() == time.Monday) ||
				(t.Weekday() == time.Sunday)) &&
				(t.Hour() > 17 && t.Hour() <= 23) &&
				// never completed or last completed mission is not on the same day
				(m.LastTask == nil ||
					!(m.LastTask.Year() == t.Year() &&
						m.LastTask.Month() == t.Month() &&
						m.LastTask.Day() == t.Day())) {
				m.transitToNew(Active)
			}
		default:
			return fmt.Errorf("invalid request type %v for state %v", e.Type, m.StateName)
		}
	case Active:
		switch e.Type {
		case Time:
			err = m.CallForWarrior()
		case Take:
			m.Taker = &e.Taker
			m.transitToNew(Taken)
			err = m.NotifyTaken()
		default:
			return fmt.Errorf("invalid request type %v for state %v", e.Type, m.StateName)
		}
	case Taken:
		switch e.Type {
		case Time:
			m.transitToNew(Pending)
		default:
			return fmt.Errorf("invalid request type %v for state %v", e.Type, m.StateName)
		}
	case Pending:
		switch e.Type {
		case Time:
			err = m.RemindMission()
		case Complete:
			m.Users[e.Taker].CanPulled += e.PulledCount
			m.transitToNew(Done)
		}
	case Done:
		switch e.Type {
		case Time:
			err = m.NotifyComplete()
			m.Taker = nil
			m.LastTask = &e.Time
			m.transitToNew(Inactive)
		}
	default:
		panic("invalid state")
	}
	return err
}
