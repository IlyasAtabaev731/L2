package main

import (
	"errors"
	"strconv"
	"strings"
	"unicode"
)

// UnpackString выполняет примитивную распаковку входной строки.
func UnpackString(input string) (string, error) {
	if input == "" {
		return "", nil
	}

	var result strings.Builder
	var escape, prevChar bool
	var prevRune rune

	for _, r := range input {
		switch {
		case escape:
			// Обработка экранированных символов
			result.WriteRune(r)
			prevRune = r
			prevChar = true
			escape = false

		case r == '\\':
			escape = true

		case unicode.IsDigit(r):
			if prevChar {
				// Рассматриваем цифру как количество повторений
				count, _ := strconv.Atoi(string(r))
				result.WriteString(strings.Repeat(string(prevRune), count-1))
				prevChar = false
			} else {
				return "", errors.New("invalid string format")
			}

		default:
			result.WriteRune(r)
			prevRune = r
			prevChar = true
		}
	}

	if escape {
		return "", errors.New("invalid string format")
	}

	return result.String(), nil
}

// Юнит-тесты
func main() {
	tests := []struct {
		input    string
		expected string
		hasError bool
	}{
		{"a4bc2d5e", "aaaabccddddde", false},
		{"abcd", "abcd", false},
		{"45", "", true},
		{"", "", false},
		{"qwe\\4\\5", "qwe45", false},
		{"qwe\\45", "qwe44444", false},
		{"qwe\\\\5", "qwe\\\\\\\\\\", false},
		{"a\\4b3", "a4bbb", false},
		{"a4\\", "", true},
	}

	for _, test := range tests {
		result, err := UnpackString(test.input)
		if (err != nil) != test.hasError || result != test.expected {
			status := "FAILED"
			if err != nil {
				status += ": " + err.Error()
			}
			println(status, "| Input:", test.input, "| Expected:", test.expected, "| Got:", result)
		} else {
			println("PASSED | Input:", test.input)
		}
	}
}
