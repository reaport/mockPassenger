package main

import (
	"bytes"
	"context" // Добавлен импорт для работы с контекстом
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
)

// Структура для тела запроса
type TicketRequest struct {
	PassengerID string `json:"passengerId"`
	FlightID    string `json:"flightId"`
	SeatClass   string `json:"seatClass"`
	MealType    string `json:"mealType"`
	Baggage     string `json:"baggage"`
}

// Структура для ответа
type TicketResponse struct {
	Message          string          `json:"message"`
	AvailableFlights []FlightDetails `json:"availableFlights"`
}

type FlightDetails struct {
	FlightID          string         `json:"flightId"`
	StartRegisterTime time.Time      `json:"registrationStartTime"`
	Direction         string         `json:"direction"`
	DepartureTime     time.Time      `json:"departureTime"`
	AvailableSeats    AvailableSeats `json:"availableSeats"`
}

type AvailableSeats struct {
	Economy  int `json:"economy"`
	Business int `json:"business"`
}

func main() {
	wg := &sync.WaitGroup{}
	// URL API
	url := "https://tickets.reaport.ru/buy"
	// Генерация UUID для PassengerID
	passengerUUID := uuid.New().String()

	// Данные для запроса
	requestBody := TicketRequest{
		PassengerID: passengerUUID,
		FlightID:    "",
		SeatClass:   "economy",
		MealType:    "Vegan",
		Baggage:     "да",
	}

	// Сериализация тела запроса в JSON
	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Println("Ошибка при сериализации JSON:", err)
		return
	}

	// Создание контекста с тайм-аутом в 5 секунд
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel() // Отмена контекста после завершения

	// Создание HTTP-запроса с привязкой к контексту
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Println("Ошибка при создании запроса:", err)
		return
	}

	// Установка заголовка Content-Type
	req.Header.Set("Content-Type", "application/json")

	// Выполнение запроса с использованием клиента
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Ошибка при выполнении запроса:", err)
		return
	}
	defer resp.Body.Close()

	// Проверка статуса ответа
	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Ошибка: код ответа %d\n", resp.StatusCode)
		return
	}

	// Чтение и десериализация ответа
	var responseData TicketResponse
	err = json.NewDecoder(resp.Body).Decode(&responseData)
	if err != nil {
		fmt.Println("Ошибка при декодировании ответа:", err)
		return
	}

	// Вывод результата в читаемом виде
	jsonOutput, _ := json.MarshalIndent(responseData, "", "  ")
	fmt.Println("Ответ от сервера:")
	fmt.Println(string(jsonOutput))
	flights := responseData.AvailableFlights
	for _, flight := range flights {
		passengers := make([]string, 0)
		for i := 0; i < flight.AvailableSeats.Economy/10; i++ {
			// Генерация UUID для PassengerID
			passengerId := uuid.New().String()
			requestBodyOne := TicketRequest{
				PassengerID: passengerId,
				FlightID:    flight.FlightID,
				SeatClass:   "economy",
				MealType:    "Vegetarian",
				Baggage:     "да",
			}
			// Сериализация тела запроса в JSON
			jsonDat, err := json.Marshal(requestBodyOne)
			if err != nil {
				fmt.Println("Ошибка при сериализации JSON:", err)
				return
			}

			// Создание контекста с тайм-аутом в 5 секунд
			ctx, can := context.WithTimeout(context.Background(), 5*time.Second)
			defer can() // Отмена контекста после завершения
			fmt.Println("Пассажир ", passengerId, " покупает билет на рейс", flight.FlightID)
			// Создание HTTP-запроса с привязкой к контексту
			req2, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonDat))
			if err != nil {
				fmt.Println("Ошибка при создании запроса:", err)
				return
			}

			// Установка заголовка Content-Type
			req2.Header.Set("Content-Type", "application/json")

			// Выполнение запроса с использованием клиента
			client := &http.Client{}
			resp, err := client.Do(req2)
			if err != nil {
				fmt.Println("Ошибка при выполнении запроса:", err)
				return
			}
			defer resp.Body.Close()

			// Проверка статуса ответа
			if resp.StatusCode != http.StatusOK {
				fmt.Printf("Ошибка: код ответа %d\n", resp.StatusCode)
				return
			}
			passengers = append(passengers, passengerId)
		}
		// Регистрируемся на рейс
		wg.Add(1)
		go Register(passengers, flight.FlightID, flight.StartRegisterTime, wg)
	}
	wg.Wait()
}

type Passenger struct {
	Uuid          string  `json:"passengerId"`
	BaggageWeight float64 `json:"baggageWeight"`
	MealType      string  `json:"mealType"`
}

func Register(passangers []string, flightId string, departureTime time.Time, wg *sync.WaitGroup) {
	defer wg.Done()
	registationTime := departureTime.Add(-59 * time.Minute).Add(-30 * time.Second)
	timeUntil := time.Until(registationTime)
	timerChan := time.After(timeUntil)
	fmt.Println("✈️ Register wait ", flightId, "ждём ", float64(timeUntil)/float64(time.Minute), "минут")
	// Ждём когда начнётся регистрация
	<-timerChan
	fmt.Println("✈️ Register begin", flightId)
	for _, pass := range passangers {
		fmt.Print("Регистрируется пассажир ", pass, " на рейс ", flightId)
		url := "https://register.reaport.ru/passenger"
		requestBody := Passenger{
			Uuid:          pass,
			BaggageWeight: 0.1,
			MealType:      "Vegetarian",
		}
		// Сериализация тела запроса в JSON
		jsonData, err := json.Marshal(requestBody)
		if err != nil {
			fmt.Println("Ошибка при сериализации JSON:", err)
			return
		}

		// Создание контекста с тайм-аутом в 5 секунд
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel() // Отмена контекста после завершения

		// Создание HTTP-запроса с привязкой к контексту
		req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
		if err != nil {
			fmt.Println("Ошибка при создании запроса:", err)
			return
		}

		// Установка заголовка Content-Type
		req.Header.Set("Content-Type", "application/json")

		// Выполнение запроса с использованием клиента
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			fmt.Println("Ошибка при выполнении запроса:", err)
			return
		}
		defer resp.Body.Close()

		// Проверка статуса ответа
		if resp.StatusCode != http.StatusOK {
			fmt.Printf("Ошибка: код ответа %d\n", resp.StatusCode)
			return
		}
		fmt.Println("  ✅ Succes")
	}
}
