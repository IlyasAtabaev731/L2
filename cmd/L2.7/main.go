package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
)

func main() {
	// Command-line flags
	fieldsStr := flag.String("f", "", "поля для выбора")
	delimiter := flag.String("d", "\t", "символ разделителя")
	onlySeparated := flag.Bool("s", false, "только строки, содержащие разделитель")

	flag.Parse()

	if *fieldsStr == "" {
		fmt.Fprintln(os.Stderr, "опция -f обязательна")
		os.Exit(1)
	}

	// Разбор индексов полей
	fields := parseFields(*fieldsStr)

	// Чтение входных строк
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Text()
		if *onlySeparated && !strings.Contains(line, *delimiter) {
			continue
		}
		parts := strings.Split(line, *delimiter)
		var output []string
		for _, field := range fields {
			if field-1 < len(parts) {
				output = append(output, parts[field-1])
			}
		}
		fmt.Println(strings.Join(output, ""))
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "чтение из стандартного ввода:", err)
		os.Exit(1)
	}
}

// parseFields разбирает аргумент полей в срез целых чисел
func parseFields(fieldsStr string) []int {
	var fields []int
	parts := strings.Split(fieldsStr, ",")
	for _, part := range parts {
		if strings.Contains(part, "-") {
			rangeParts := strings.Split(part, "-")
			if len(rangeParts) != 2 {
				continue // неверный диапазон, пропускаем
			}
			start, _ := strconv.Atoi(rangeParts[0])
			end, _ := strconv.Atoi(rangeParts[1])
			for i := start; i <= end; i++ {
				fields = append(fields, i)
			}
		} else {
			field, _ := strconv.Atoi(part)
			fields = append(fields, field)
		}
	}
	return fields
}
