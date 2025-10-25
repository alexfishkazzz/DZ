package main

import (
	"fmt"
	"io"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func main() {
	url := "http://srv.msk01.gigacorp.local/_stats"
	errorCount := 0

	for {
		resp, err := http.Get(url)
		if err != nil {
			errorCount++
			handleError(&errorCount)
			time.Sleep(1 * time.Second)
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			errorCount++
			handleError(&errorCount)
			time.Sleep(1 * time.Second)
			continue
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			errorCount++
			handleError(&errorCount)
			time.Sleep(1 * time.Second)
			continue
		}

		data := strings.Split(strings.TrimSpace(string(body)), ",")
		if len(data) != 7 {
			errorCount++
			handleError(&errorCount)
			time.Sleep(1 * time.Second)
			continue
		}

		values := make([]float64, 7)
		for i, s := range data {
			val, err := strconv.ParseFloat(s, 64)
			if err != nil {
				errorCount++
				handleError(&errorCount)
				time.Sleep(1 * time.Second)
				break // Выходим из for, чтобы не продолжать с неверными данными
			}
			values[i] = val
		}
		if errorCount > 0 { // Если была ошибка в парсинге, уже обработано
			continue
		}

		// Сброс счётчика ошибок при успешном запросе
		errorCount = 0

		// Проверки
		loadAvg := values[0]
		if loadAvg > 30 {
			fmt.Printf("Load Average is too high: %.0f\n", loadAvg)
		}

		totalMem := values[1]
		usedMem := values[2]
		memPercent := (usedMem / totalMem) * 100
		if memPercent > 80 {
			fmt.Printf("Memory usage too high: %.0f%%\n", math.Round(memPercent))
		}

		totalDisk := values[3]
		usedDisk := values[4]
		diskPercent := usedDisk / totalDisk
		if diskPercent > 0.9 {
			freeMb := (totalDisk - usedDisk) / (1024 * 1024)
			fmt.Printf("Free disk space is too low: %.0f Mb left\n", math.Round(freeMb))
		}

		totalBandwidth := values[5]
		usedBandwidth := values[6]
		bandPercent := usedBandwidth / totalBandwidth
		if bandPercent > 0.9 {
			freeMbit := ((totalBandwidth - usedBandwidth) * 8) / 1000000
			fmt.Printf("Network bandwidth usage high: %.0f Mbit/s available\n", math.Round(freeMbit))
		}

		time.Sleep(1 * time.Second)
	}
}

func handleError(count *int) {
	if *count >= 3 {
		fmt.Println("Unable to fetch server statistic")
	}
}