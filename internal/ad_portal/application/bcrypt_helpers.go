package application

import "golang.org/x/crypto/bcrypt"

func bcryptGenerateDummy() (string, error) {
	h, err := bcrypt.GenerateFromPassword([]byte("__login_dummy__"), bcryptCost)
	if err != nil {
		return "", err
	}
	return string(h), nil
}

func bcryptHashPassword(password string) (string, error) {
	h, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
	if err != nil {
		return "", err
	}
	return string(h), nil
}

func bcryptPasswordMatches(hash, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}
