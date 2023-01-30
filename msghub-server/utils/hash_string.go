package utils

import (
	"golang.org/x/crypto/bcrypt"
	"log"
)

func HashEncrypt(str string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(str), 14)

	if err != nil {
		log.Fatal("Error in encrypting - ", err)
		return string(bytes), err
	}

	return string(bytes), nil
}

func CheckPasswordMatch(formPass, dbHashedPass string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(dbHashedPass), []byte(formPass))
	// if err == nil, then both passwords are same
	return err == nil
}
