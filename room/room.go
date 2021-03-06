package room

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/Balhazraell/logger"
	"github.com/streadway/amqp"
)

// Room - основная структура данных этого пакета.
var Room room

type chunc struct {
	ID          int      `json:"id"`
	State       int      `json:"state"`
	Сoordinates [][2]int `json:"coordinates"`
}

type room struct {
	ID      int
	Map     map[int]*chunc
	clients []int
	channel *amqp.Channel

	// Переменные логики.
	GameState     int // Делаем крестики нолики, по этому 2 состояния - ходит один потом другой.
	lastMovedUser int // Пользователь последний сделавший ход.

	// Каналы
	shutdownLoop chan bool
	updateMap    chan bool

	//--- RabbitMQ
	APIandCallbackMetods map[string]func(string)
	connectRMQ           *amqp.Connection
	channelRMQ           *amqp.Channel
}

// StartNewRoom - метод запуска новой комнаты.
// На вход подается id комнаты котурую надо создать.
func StartNewRoom(id int) {
	// TODO: Временное решение, для запуска как отдельные приложения.
	rand.Seed(time.Now().UnixNano())
	id = rand.Intn(100)
	Room = room{
		ID:                   id,
		Map:                  make(map[int]*chunc),
		shutdownLoop:         make(chan bool),
		updateMap:            make(chan bool),
		APIandCallbackMetods: fillMetods(),
	}

	createMap()
	logger.InfoPrintf("Комната с именем %v начала свою работу", fmt.Sprintf("room_%v", id))

	StartRabbitMQ(fmt.Sprintf("room_%v", id))

	go Room.loop()

	CreateMessage(id, "RoomConnect")
}

// Stop - Останавлием работу комнаты
func (r *room) Stop() {
	// ...какая-нибудь логика завершения работы.
	r.shutdownLoop <- true
}

func (r *room) loop() {
	defer func() {
		r.connectRMQ.Close()
		r.channelRMQ.Close()
		logger.InfoPrintf("Комната с id=%v закончила работу.", r.ID)
	}()

	logger.InfoPrintf("Комната с id=%v начала работу.", r.ID)

	for {
		// Обновление логики происходит тут.

		select {
		case <-r.shutdownLoop:
			return

		// Даже не знаю на сколько целесообразно делать это в отдельном потоке.
		// Мсль была в том, что update карт должен произоти не моментально после изменений
		// но хз на сколько это грамотоное решение.
		case <-r.updateMap:
			updateClientsMap(Room.clients)
		}
	}
}

func fillMetods() map[string]func(string) {
	result := APIMetods
	for key, value := range CallbackMetods {
		result[key] = value
	}

	return result
}
