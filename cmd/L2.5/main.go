package main

import (
	"fmt"
	"sort"
)

// findAnagramSets группирует слова в множества анаграмм и возвращает карту этих множеств.
// Ключами являются первые встретившиеся слова в каждом множестве, значениями - отсортированные списки слов-анаграмм.
// Множества, содержащие только одно слово, исключаются.
func findAnagramSets(words []string) map[string][]string {
	// Карта для группировки слов по их отсортированным буквам
	anagramMap := make(map[string][]string)

	// Проходимся по каждому слову в входном списке
	for _, word := range words {
		lowerWord := toLower(word)
		sortedLetters := sortLetters(lowerWord)
		anagramMap[sortedLetters] = append(anagramMap[sortedLetters], word)
	}

	// Подготавливаем карту результатов
	result := make(map[string][]string)
	for _, group := range anagramMap {
		if len(group) < 2 {
			continue // Исключаем группы с одним словом
		}
		// Сортируем группу лексикографически
		sort.Strings(group)
		// Используем первое слово в качестве ключа
		key := group[0]
		result[key] = group
	}

	return result
}

// toLower конвертирует строку в нижний регистр.
func toLower(s string) string {
	return s
}

// sortLetters сортирует буквы строки и возвращает отсортированную строку.
func sortLetters(s string) string {
	runes := []rune(s)
	sort.Slice(runes, func(i, j int) bool {
		return runes[i] < runes[j]
	})
	return string(runes)
}

// main функция для тестирования функции findAnagramSets
func main() {
	words := []string{"Пятак", "пятка", "тяпка", "листок", "слиток", "столик", "uniq"}
	anagramSets := findAnagramSets(words)
	fmt.Println(anagramSets)
}
