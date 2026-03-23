package entity

type PendingUser struct {
	Email    string
	PassHash []byte
	Role string
	OrganizationName string
	OrganizationURL string
	Code     string
}
