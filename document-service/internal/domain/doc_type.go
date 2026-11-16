package domain

type DocType string

const (
	DocTypePassport       DocType = "PASSPORT"
	DocTypeSelfie         DocType = "SELFIE"
	DocTypeProofOfAddress DocType = "PROOF_OF_ADDRESS"
)

func (d DocType) IsValid() bool {
	switch d {
	case DocTypePassport, DocTypeSelfie, DocTypeProofOfAddress:
		return true
	}
	return false
}
