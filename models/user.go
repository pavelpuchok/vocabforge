package models

import "fmt"

type UserID string

func (u UserID) String() string {
	return string(u)
}

const objectIdHexLen = 24

func UserIDFromText(s string) (UserID, error) {
	if len(s) != objectIdHexLen {
		return "", fmt.Errorf("models.UserIDFromText invalid user ID string %s", s)
	}

	return UserID(s), nil
}

type User struct {
	ID UserID
}
