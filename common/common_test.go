package common

import (
	"testing"
)

func TestHashPassword(t *testing.T) {
	pwd := "MySuperStrongPassword"

	hashedPwd, _ := HashPassword(pwd)

	if len(hashedPwd) < 1 {
		t.Error("Password hashing failed")
	}
}
