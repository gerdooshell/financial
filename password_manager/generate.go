package password_manager

import (
	"math/rand"
	"strings"
	"unicode/utf8"
)

type PasswordManager interface {
	// GeneratePassword generates a random strong password of a given length in a range
	GeneratePassword(int) string
}

func NewPasswordManager() PasswordManager {
	return &passwordManager{}
}

type passwordManager struct{}

func (pm *passwordManager) GeneratePassword(length int) string {
	maxLength := 200
	minLength := 4
	if length <= minLength {
		length = minLength
	}
	if length >= maxLength {
		length = maxLength
	}
	dictionaryKeys := "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ_@"
	dictionary := make(map[int]rune)
	for i, r := range dictionaryKeys {
		dictionary[i] = r
	}
	sb := strings.Builder{}
	for i := 0; i < length; i++ {
		index := rand.Intn(utf8.RuneCountInString(dictionaryKeys))
		sb.WriteRune(dictionary[index])
	}
	return sb.String()
}
