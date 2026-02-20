package user

type TgUser struct {
	ID         int64  `json:"id"`
	Username   string `json:"username"`
	FirstName  string `json:"first_name"`
	LastName   string `json:"last_name"`
	AccessHash int64  `json:"access_hash"`
}

func NewTgUser(id int64, username, firstName, lastName string, accessHash int64) *TgUser {
	return &TgUser{
		ID:         id,
		Username:   username,
		FirstName:  firstName,
		LastName:   lastName,
		AccessHash: accessHash,
	}
}
