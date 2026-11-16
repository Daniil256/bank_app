package domain

import "time"

type CheckResult string

const (
	CheckPass CheckResult = "PASS"
	CheckFail CheckResult = "FAIL"
)

type KycCheck struct {
	ID            string
	ApplicationID string
	CheckType     string // "sanctions", "age"
	Result        CheckResult
	Details       string
	CreatedAt     time.Time
}
