package fsm

import (
	"fmt"
	"time"
)

// CallForWarrior notify all users about the task
func (m *FSM) CallForWarrior() error {
	if m.LastNotify != nil && m.LastNotify.Add(m.Config.CallInterval).After(time.Now()) {
		return nil
	}
	userList := UserMapToList(m.Users)
	message := "Hi trash agents, it's time to escort trash cans. " +
		"Reply anything (e.g.: I'm in!) to take this mission.\nLeaderboard:\n" +
		stats(userList)
	return m.NotifyUsers(userList, message)
}

// NotifyTaken notify all users that the task is taken
func (m *FSM) NotifyTaken() error {
	if m.LastNotify != nil {
		return nil
	}
	message := fmt.Sprintf("The mission is taken by %v", *m.Taker)
	return m.NotifyUsers(UserMapToList(m.Users), message)

}

// RemindMission remind the user to complete the task
func (m *FSM) RemindMission() error {
	if m.LastNotify != nil && m.LastNotify.Add(m.Config.RemindInterval).After(time.Now()) {
		return nil
	}
	message := "Dear trash agent, is the trash escort mission done?" +
		"Reply the number (1-3) of trash cans you esctored.\n" +
		"I'll remind you every " + fmt.Sprint(m.Config.RemindInterval)
	return m.NotifyUsers([]*User{m.Users[*m.Taker]}, message)
}

// NotifyComplete notify all users that the task is completed
func (m *FSM) NotifyComplete() error {
	if m.LastNotify != nil {
		return nil
	}
	userList := UserMapToList(m.Users)
	message := fmt.Sprintf("The mission is completed by %v.\nLeaderboard:\n%v", *m.Taker, stats(userList))
	return m.NotifyUsers(userList, message)

}
