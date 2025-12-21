package utils

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

func HashPassword(password string) (string, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	return string(hashed), nil
}

func CheckPassword(password, hashPassword string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashPassword), []byte(password))

	return err == nil
}

func ValidateRegisterRequest(firstName, lastName, password, gender, biography, city string) error {
	if firstName == "" || len(firstName) < 2 {
		return fmt.Errorf("Укажити фамилию")
	}
	if lastName == "" || len(lastName) < 2 {
		return fmt.Errorf("Укажите имя")
	}
	if password == "" || len(password) < 6 {
		return fmt.Errorf("Пароль менее 6 символов")
	}
	if gender == "" {
		return fmt.Errorf("Укажите пол")
	}
	if biography == "" {
		return fmt.Errorf("Напишите о себе")
	}
	if city == "" {
		return fmt.Errorf("Укажите город")
	}

	return nil
}

func ValidatePostRequest(title, content string) error {
	if title == "" || len(title) < 2 {
		return fmt.Errorf("Отсутствует заголовок поста")
	}
	if content == "" || len(content) < 2 {
		return fmt.Errorf("Отсутствует описание поста")
	}

	return nil
}
