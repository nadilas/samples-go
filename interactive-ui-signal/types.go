package interactive_ui_signal

import "time"

type UpgradeRequest struct {
	Account string `json:"account,omitempty"`
	To      string `json:"to,omitempty"`
	Actor   string `json:"actor,omitempty"`
}

type UpgradeResponse struct {
	ValidFrom time.Time `json:"valid_from"`
}

type DeleteAccountRequest struct {
	Account string `json:"account,omitempty"`
	Actor   string `json:"actor,omitempty"`
}

type Account struct {
	Name          string     `json:"name,omitempty"`
	Plan          string     `json:"plan,omitempty"`
	Created       *time.Time `json:"created,omitempty"`
	Terminated    *time.Time `json:"terminated,omitempty"`
	PlanValidFrom time.Time  `json:"plan_valid_from"`
}
