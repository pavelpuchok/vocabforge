package models

type UserID string

func (u UserID) String() string {
	return string(u)
}

type User struct {
	ID UserID
}
