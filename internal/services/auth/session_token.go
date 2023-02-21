package auth

import (
	uuid "github.com/nu7hatch/gouuid"
)

func GenerateToken() (string, error) {
	u4, err := uuid.NewV4()
	if err != nil {
		return "", err
	}

	return u4.String(), nil
}
