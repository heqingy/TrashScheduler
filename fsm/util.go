package fsm

import (
	"fmt"
	"sort"
	"time"
)

func stats(users []*User) string {
	var result string
	sort.Slice(
		users,
		func(i, j int) bool { return users[i].CanPulled > users[j].CanPulled },
	)
	for _, u := range users {
		result += fmt.Sprintf("%v: %v\n", u.Name, u.CanPulled)
	}
	return result
}

func timeAddr(t time.Time) *time.Time {
	return &t
}

// UserMapToList converts the user map to list
func UserMapToList(users map[string]*User) []*User {
	result := []*User{}
	for _, v := range users {
		result = append(result, v)
	}
	return result
}

// UserListToMap converts the user list to map
func UserListToMap(users []*User) map[string]*User {
	result := map[string]*User{}
	for _, u := range users {
		result[u.Name] = u
	}
	return result
}
