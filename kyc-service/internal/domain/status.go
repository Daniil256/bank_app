package domain

type Status string

const (
	StatusDraft         Status = "DRAFT"
	StatusSubmitted     Status = "SUBMITTED"
	StatusAutoChecks    Status = "AUTO_CHECKS"
	StatusInReview      Status = "IN_REVIEW"
	StatusApproved      Status = "APPROVED"
	StatusRejected      Status = "REJECTED"
	StatusAutoRejected  Status = "AUTO_REJECTED"
	StatusDocsRequested Status = "DOCS_REQUESTED"
)

const RoleClient = "CLIENT"
const RoleAdmin = "ADMIN"
const RoleOfficer = "OFFICER"

func CanTransition(from, to Status) bool {
	transitions := map[Status][]Status{
		StatusInReview: {StatusApproved, StatusRejected, StatusDocsRequested},
	}
	for _, allowed := range transitions[from] {
		if allowed == to {
			return true
		}
	}
	return false
}
