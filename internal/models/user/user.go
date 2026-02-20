package user

import "time"

type User struct {
	Id               int64     `json:"id"`
	TotalLinks       int       `json:"total_links"`
	IsDeleted        bool      `json:"is_deleted"`
	IsPremium        bool      `json:"is_premium"`
	IsVerified       bool      `json:"is_verified"`
	Credit           int       `json:"refer_count"`
	LastCreditUpdate time.Time `json:"last_credit_update"`
	IsBanned         bool      `json:"is_banned"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

type TgUser struct {
	Id         int64  `json:"id"`
	Username   string `json:"username"`
	FirstName  string `json:"first_name"`
	LastName   string `json:"last_name"`
	AccessHash int64  `json:"access_hash"`
}

func InitTgUser(id int64, username, firstName, lastName string, accessHash int64) *TgUser {
	return &TgUser{
		Id:         id,
		Username:   username,
		FirstName:  firstName,
		LastName:   lastName,
		AccessHash: accessHash,
	}
}
