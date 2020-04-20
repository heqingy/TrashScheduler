package svc

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	fsm "github.com/heqingy/TrashScheduler/fsm"
	log "github.com/sirupsen/logrus"
)

type SMSConfig struct {
	AccID  string
	Token  string
	Number string
}

func NewSMS(c SMSConfig) *SMS {
	return &SMS{
		accountSid: c.AccID,
		authToken:  c.Token,
		number:     c.Number,
		urlStr:     "https://api.twilio.com/2010-04-01/Accounts/" + c.AccID + "/Messages.json",
	}
}

type SMS struct {
	accountSid string
	authToken  string
	number     string
	urlStr     string
}

func (s *SMS) Send(phone string, message string) error {
	msgData := url.Values{}
	msgData.Set("To", phone)
	msgData.Set("From", s.number)
	msgData.Set("Body", message)
	msgDataReader := *strings.NewReader(msgData.Encode())
	client := &http.Client{}
	req, _ := http.NewRequest("POST", s.urlStr, &msgDataReader)
	req.SetBasicAuth(s.accountSid, s.authToken)
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		var data map[string]interface{}
		decoder := json.NewDecoder(resp.Body)
		err := decoder.Decode(&data)
		if err == nil {
			log.Printf("SMS sent. SID: %v", data["sid"])
		}
	} else {
		log.Warnf("SMS send status code: %v", resp.StatusCode)
	}
	return nil
}

// Notify send message to users
func (s *SMS) Notify(users []*fsm.User, message string) error {
	errList := []string{}
	for _, u := range users {
		for _, p := range u.Phone {
			if err := s.Send(p, message); err != nil {
				errList = append(errList, err.Error())
			}
		}
	}
	if len(errList) != 0 {
		return fmt.Errorf(strings.Join(errList, "\n"))
	}
	return nil
}
