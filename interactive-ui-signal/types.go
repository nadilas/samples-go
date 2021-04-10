package interactive_ui_signal

import "time"

type UpgradeRequest struct {
	To    string
	Actor string
}

type UpgradeResponse struct {
	ValidFrom time.Time
}

type DeleteAccountRequest struct {
	Actor string
}

type Account struct {
	Name          string
	Plan          string
	Created       *time.Time
	Terminated    *time.Time
	PlanValidFrom time.Time
}
