package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

type Alert struct {
	ID        string    `json:"id"`
	EventID   string    `json:"event_id"`
	RuleID    string    `json:"rule_id"`
	RuleName  string    `json:"rule_name"`
	Severity  string    `json:"severity"`
	Message   string    `json:"message"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

var alertCounter int

func newAlert(event Event, rule Rule) Alert {
	alertCounter++
	return Alert{
		ID:        fmt.Sprintf("alert_%03d", alertCounter),
		EventID:   event.ID,
		RuleID:    rule.ID,
		RuleName:  rule.Name,
		Severity:  rule.Severity,
		Message:   rule.Message,
		Status:    "open",
		CreatedAt: time.Now(),
	}
}

func persistAlerts(alerts []Alert, path string) error {
	data, err := json.MarshalIndent(alerts, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}
