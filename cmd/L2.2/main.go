package main

import (
	"fmt"
	"github.com/beevik/ntp"
	"log"
)

func main() {
	// Получаем текущее точное время через NTP
	time, err := ntp.Time("pool.ntp.org")
	if err != nil {
		// Логируем ошибку в STDERR и завершаем выполнение программы с кодом 1
		log.Fatalf("Ошибка при получении времени через NTP: %v", err)
	}

	// Печатаем текущее точное время
	fmt.Println("Текущее точное время:", time.Format("2006-01-02 15:04:05 MST"))
}
