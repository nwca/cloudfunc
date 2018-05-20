package main

type Context struct {
	EventID   string   `json:"eventId"`
	EventType string   `json:"eventType"`
	Timestamp string   `json:"timestamp"`
	Resource  Resource `json:"resource"`
}

type Resource struct {
	Service string `json:"service"`
	Name    string `json:"name"`
	Type    string `json:"type"`
}
