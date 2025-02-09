package funcs

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
	"sync"
	"time"
)

// Задание 1. Написание базового теста функции
func BasicFunc(a, b int) int {
	return a + b
}

// Задание 2. Тестирование ошибок
func FuncWithError(a, b int) (int, error) {
	if b == 0 {
		return 0, fmt.Errorf("divided by zero")
	}
	return a / b, nil
}

// Задание 3. Тестирование слайсов и мапов
func FuncEvenSlice(slice []int) []int {
	newSlice := make([]int, 0)
	for _, v := range slice {
		if v%2 == 0 {
			newSlice = append(newSlice, v)
		}
	}
	return newSlice
}

func FuncCountLettersMap(s string) map[string]int {
	count := map[string]int{}
	for _, v := range s {
		if string(v) != " " {
			count[string(v)]++
		}
	}
	return count
}

// Задание 4. Использование подпакета testing/quick
func ReverseString(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

// Задание 5. Тестирование HTTP-обработчика
func BasicHTTPResponse(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{"message": "Hello World"}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(response)
}

// Задание 6. Mocking
type OrderService interface {
	GetOrderStatus(orderID int) (string, error)
}

type OrderTracker struct {
	Service OrderService
}

func (o *OrderTracker) TrackOrder(orderID int) (string, error) {
	status, err := o.Service.GetOrderStatus(orderID)
	if err != nil {
		return "", err
	}

	return status, nil
}

// Задание 7. Тестирование времени

type RateLimiter struct {
	lastExecution time.Time
	timeProvider  TimeProvider
}

type TimeProvider interface {
	Now() time.Time
}

func NewRateLimiter(timeProvider TimeProvider) *RateLimiter {
	return &RateLimiter{lastExecution: timeProvider.Now().Add(-time.Minute), timeProvider: timeProvider}
}

func (r *RateLimiter) CanExecute() (bool, error) {
	currentTime := r.timeProvider.Now()
	if currentTime.Sub(r.lastExecution) < time.Minute {
		return false, errors.New("too early, try a little later")
	}
	r.lastExecution = r.timeProvider.Now()
	return true, nil
}

// Задание 8. Тестирование конкурентности
type SafeCounter struct {
	mu    sync.Mutex
	count int
}

func (c *SafeCounter) Inc() {
	c.mu.Lock()
	c.count++
	c.mu.Unlock()
}

func (c *SafeCounter) Value() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.count
}

// Задание 9. Тестирование конкретной ошибки
func Errors(a, b, c int) error {
	if a == b && b == c {
		return nil
	} else if a != b && b == c {
		return errors.New("a")
	} else if a != b && a == c {
		return errors.New("b")
	} else {
		return errors.New("c")
	}
}

// Задание 10. Тестирование HTTP-сервера
type Task struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Done bool   `json:"done"`
}

var Tasks = []Task{
	{ID: 1, Name: "GIT", Done: true},
	{ID: 2, Name: "Computer Networks", Done: true},
	{ID: 3, Name: "Databases fundamentals", Done: true},
	{ID: 4, Name: "Go databases", Done: true},
	{ID: 5, Name: "Go testing", Done: false},
}

func GetTasks(c *gin.Context) {
	c.JSON(http.StatusOK, Tasks)
}

func CreateTask(c *gin.Context) {
	newTask := Task{}
	if err := c.ShouldBindJSON(&newTask); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Bad Request": err.Error()})
		return
	}

	newTask.ID = Tasks[len(Tasks)-1].ID + 1
	Tasks = append(Tasks, newTask)
	c.JSON(http.StatusCreated, newTask)
}

func DeleteTask(c *gin.Context) {
	delID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Bad ID": err.Error()})
		return
	}

	for i, t := range Tasks {
		if t.ID == delID {
			Tasks = append(Tasks[:i], Tasks[i+1:]...)
			c.JSON(http.StatusOK, gin.H{"deleted": delID})
			return
		}
	}

	c.JSON(http.StatusBadRequest, gin.H{"Not found": delID})
}

func DoneTask(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Bad ID": err.Error()})
		return
	}

	for _, t := range Tasks {
		if t.ID == id {
			t.Done = true
			c.JSON(http.StatusOK, gin.H{"done": t.ID})
			return
		}
	}

	c.JSON(http.StatusBadRequest, gin.H{"Not found": id})
}

func TaskServer() *gin.Engine {
	router := gin.Default()
	router.GET("/todo", GetTasks)
	router.PUT("/todo/:id", DoneTask)
	router.POST("/todo", CreateTask)
	router.DELETE("/todo/:id", DeleteTask)
	return router
}
