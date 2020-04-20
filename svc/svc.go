package svc

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	fsm "github.com/heqingy/TrashScheduler/fsm"
	log "github.com/sirupsen/logrus"
)

type FSMConfig struct {
	fsm.Config
	StatePath string
}

type Config struct {
	Port int
	SMSConfig
	FSMConfig
}

type Service struct {
	m      *fsm.FSM
	config Config
	sms    *SMS
	events chan *fsm.Event
	web    *gin.Engine
}

// HandleEvent handles the event
func (s *Service) HandleEvent(e *fsm.Event) error {
	return s.m.Transit(e)
}

func (s *Service) phoneToUser(phone string) *fsm.User {
	for _, u := range s.m.Users {
		for _, p := range u.Phone {
			if p == phone {
				return u
			}
		}
	}
	return nil
}

func (s *Service) HandleIncomeSMS(phone string, message string) error {
	user := s.phoneToUser(phone)
	if user == nil {
		// user doesn't exist, ignore it
		return nil
	}
	switch s.m.StateName {
	case fsm.Active:
		s.events <- &fsm.Event{
			Type: fsm.Take,
			EventContext: fsm.EventContext{
				Taker: user.Name,
				Time:  time.Now(),
			},
		}
	case fsm.Taken:
		fallthrough
	case fsm.Pending:
		v, err := strconv.Atoi(strings.TrimSpace(message))
		if err != nil {
			s.sms.Send(phone, fmt.Sprintf("invalid count: '%v' is not a number", message))
			return nil
		}
		if v < 1 || v > 3 {
			s.sms.Send(phone, fmt.Sprintf("invalid count %v: count must be 1, 2 or 3", message))
			return nil
		}
		s.events <- &fsm.Event{
			Type: fsm.Complete,
			EventContext: fsm.EventContext{
				Taker:       user.Name,
				PulledCount: v,
				Time:        time.Now(),
			},
		}
	default:
		s.sms.Send(phone, fmt.Sprintf("current state: %v, no action for this message", s.m.StateName))
	}
	return nil
}

// NewService returns new service
func NewService(users map[string]*fsm.User, c Config) *Service {
	r := gin.Default()
	smsSvc := NewSMS(c.SMSConfig)
	c.FSMConfig.NotifySvc = smsSvc
	service := &Service{
		web:    r,
		sms:    smsSvc,
		m:      fsm.NewFSM(users, c.FSMConfig.Config),
		events: make(chan *fsm.Event),
		config: c,
	}

	if stateData, err := ioutil.ReadFile(c.FSMConfig.StatePath); err != nil {
		log.Warnf("failed to load state file: %v", err)
	} else if err := service.m.Load(stateData); err != nil {
		log.Warnf("failed to load state: %v", err)
	} else {
		log.Printf("successfully load state: %v", string(stateData))
	}

	r.POST("/sms", func(c *gin.Context) {
		var body incomeSMSBody
		if err := c.ShouldBind(&body); err != nil {
			c.AbortWithStatus(http.StatusInternalServerError)
			log.Error(err)
		}
		if err := service.HandleIncomeSMS(body.From, body.Body); err != nil {
			log.Error(err)
		}
		c.Abort()
	})
	return service
}

// Run spawns the service
func (s *Service) Run() {
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp: true,
	})

	// IncomeSMS Event Generator
	go s.web.Run()
	// Time Event Generator
	go func() {
		for {
			s.events <- &fsm.Event{
				Type: fsm.Time,
				EventContext: fsm.EventContext{
					Time: time.Now(),
				},
			}
			time.Sleep(time.Second)
		}
	}()

	// Main Event Loop
	for {
		e := <-s.events
		curState := s.m.StateName
		err := s.HandleEvent(e)
		if err != nil {
			log.Error(err)
		}
		if curState != s.m.StateName {
			log.Infof("State change: %v -> %v, Trigger: %#v", curState, s.m.State.StateName, e)
			data, err := s.m.Dump()
			if err != nil {
				log.Errorf("failed to dump state: %v", err)
			} else if err := ioutil.WriteFile(s.config.StatePath, data, 0600); err != nil {
				log.Errorf("failed to save state: %v", err)
			} else {
				log.Infof("State saved.")
			}
		}

	}
}
