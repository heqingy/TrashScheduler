package fsm

import (
	"time"
)

// Notifier is the service to notify user
type Notifier interface {
	Notify([]*User, string) error
}

// NotifyUsers send message to all users
func (m *FSM) NotifyUsers(users []*User, message string) error {
	m.LastNotify = timeAddr(time.Now())
	return m.NotifySvc.Notify(users, message)
}
