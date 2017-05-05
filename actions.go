package main

type Action struct {
	ActionName   string `json:"action_name"`
	Action       string `json:"action"`
	IsUserAction bool   `json:"is_user_action"`
	Method       string `json:"method"`
	ObjectID     string `json:"object_id"`
	ContentType  string `json:"content_type"`
	State        int    `json:"state"`
	UserAgent    string `json:"user_agent"`
}

func NewAction(method, object string) *Action {
	return &Action{
		ActionName:   "stop",
		Action:       "Server stop",
		IsUserAction: false,
		Method:       method,
		ObjectID:     object,
		ContentType:  "server",
		State:        4,
		UserAgent:    "Docker Stats",
	}
}
