package main

import "time"

type Action struct {
	ActionName   string    `json:"action_name"`
	Action       string    `json:"action"`
	IsUserAction bool      `json:"is_user_action"`
	Method       string    `json:"method"`
	ObjectID     string    `json:"object_id"`
	ContentType  string    `json:"content_type"`
	State        int       `json:"state"`
	UserAgent    string    `json:"user_agent"`
	StartDate    time.Time `json:"start_date"`
}

func NewAction(method, object string, startDate time.Time) *Action {
	return &Action{
		ActionName:   "stop",
		Action:       "Server stop",
		IsUserAction: false,
		Method:       method,
		ObjectID:     object,
		ContentType:  "server",
		State:        4,
		UserAgent:    "Docker Stats",
		StartDate:    startDate,
	}
}
