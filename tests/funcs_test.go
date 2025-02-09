package tests

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go-testing/funcs"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"testing/quick"
	"time"
)

// Задание 1. Написание базового теста функции
func TestBasicFunc(t *testing.T) {
	a := 2
	b := 6
	assert.Equal(t, a+b, funcs.BasicFunc(a, b))
}

// Задание 2. Тестирование ошибок
func TestFuncWithError(t *testing.T) {
	result, err := funcs.FuncWithError(10, 2)
	assert.NoError(t, err, "Error should be nil")
	assert.Equal(t, 5, result)

	_, err = funcs.FuncWithError(10, 0)
	assert.Error(t, err, "Error should not be nil")
	assert.EqualError(t, err, "divided by zero")
}

// Задание 3. Тестирование слайсов и мапов
func TestFuncEvenSlice(t *testing.T) {
	input := []int{1, 2, 3, 4, 5, 6}
	expected := []int{2, 4, 6}
	result := funcs.FuncEvenSlice(input)
	assert.Equal(t, expected, result)
}

func TestFuncCountLettersMap(t *testing.T) {
	input := "hello world"
	expected := map[string]int{"h": 1, "l": 3, "e": 1, "o": 2, "w": 1, "r": 1, "d": 1}
	result := funcs.FuncCountLettersMap(input)
	assert.Equal(t, expected, result)
}

// Задание 4. Использование подпакета testing/quick
func TestReverseStringQuick(t *testing.T) {
	f := func(s string) bool {
		return funcs.ReverseString(funcs.ReverseString(s)) == s
	}

	if err := quick.Check(f, nil); err != nil {
		t.Error(err)
	}
}

// Задание 5. Тестирование HTTP-обработчика
func TestBasicHTTPResponse(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	funcs.BasicHTTPResponse(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	assert.Equal(t, resp.StatusCode, 200)

	expected := `{"message":"Hello World"}` + "\n"
	assert.Equal(t, expected, w.Body.String())
}

// Задание 6. Mocking
type MockOrderService struct {
	mock.Mock
}

func (m *MockOrderService) GetOrderStatus(orderID int) (string, error) {
	args := m.Called(orderID)
	return args.String(0), args.Error(1)
}

func TestTrackOrder(t *testing.T) {
	mockService := new(MockOrderService)

	mockService.On("GetOrderStatus", 2301).Return("Shipped", nil)
	mockService.On("GetOrderStatus", 2302).Return("Delivered", nil)
	mockService.On("GetOrderStatus", 2303).Return("", errors.New("order not found"))

	orderTracker := &funcs.OrderTracker{Service: mockService}

	status, err := orderTracker.TrackOrder(2301)
	assert.Equal(t, status, "Shipped")
	assert.NoError(t, err)

	status, err = orderTracker.TrackOrder(2302)
	assert.Equal(t, status, "Delivered")
	assert.NoError(t, err)

	status, err = orderTracker.TrackOrder(2303)
	assert.Error(t, err)
	assert.Equal(t, "", status)

	mockService.AssertExpectations(t)
}

// Задание 7. Тестирование времени
type MockTimeProvider struct {
	mock.Mock
}

func (m *MockTimeProvider) Now() time.Time {
	args := m.Called()
	return args.Get(0).(time.Time)
}

func TestRateLimiter(t *testing.T) {
	mockTimeProvider := new(MockTimeProvider)

	currentTime := time.Now()
	// Один вызов для создания NewRateLimiter и два для работы внутри функции
	mockTimeProvider.On("Now").Return(currentTime).Times(3)

	limiter := funcs.NewRateLimiter(mockTimeProvider)

	canExecute, err := limiter.CanExecute()
	assert.True(t, canExecute)
	assert.NoError(t, err)

	currentTime = time.Now().Add(30 * time.Second)
	// Внутри функции только один вызов Now(), поэтому Once()
	mockTimeProvider.On("Now").Return(currentTime).Once()
	canExecute, err = limiter.CanExecute()
	assert.False(t, canExecute)
	assert.Error(t, err)

	currentTime = currentTime.Add(31 * time.Second)
	// Тут будет два вызова Now() внутри функции, поэтому Times(2)
	mockTimeProvider.On("Now").Return(currentTime).Times(2)
	canExecute, err = limiter.CanExecute()
	assert.True(t, canExecute)
	assert.NoError(t, err)

	mockTimeProvider.AssertExpectations(t)
}

// Задание 8. Тестирование конкурентности
func TestSafeCounter(t *testing.T) {
	var counter funcs.SafeCounter
	var wg sync.WaitGroup

	const goroutines = 50
	const increments = 20

	wg.Add(goroutines)
	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < increments; j++ {
				counter.Inc()
			}
		}()
	}

	wg.Wait()

	expected := increments * goroutines
	assert.Equal(t, expected, counter.Value())
}

// Задание 9. Тестирование конкретной ошибки
func TestErrors(t *testing.T) {
	a, b, c := 1, 1, 1
	err := funcs.Errors(a, b, c)
	assert.NoError(t, err)

	a = 2
	err = funcs.Errors(a, b, c)
	assert.Error(t, err)
	assert.EqualError(t, err, "a")

	c = 2
	err = funcs.Errors(a, b, c)
	assert.Error(t, err)
	assert.EqualError(t, err, "b")

	a, b = 3, 3
	err = funcs.Errors(a, b, c)
	assert.Error(t, err)
	assert.EqualError(t, err, "c")
}

// Задание 10. Тестирование HTTP-сервера
func TestGetTasks(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := funcs.TaskServer()

	req, err := http.NewRequest("GET", "/todo", nil)
	assert.NoError(t, err)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	resp := w.Result()

	assert.Equal(t, 200, resp.StatusCode)

	var tasks []funcs.Task
	err = json.Unmarshal(w.Body.Bytes(), &tasks)

	assert.NoError(t, err)
	assert.Equal(t, 5, len(tasks))
}

func TestCreateTask(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := funcs.TaskServer()

	// С правильным запросом
	newTask := funcs.Task{Name: "DevOps fundamentals", Done: false}
	body, err := json.Marshal(newTask)
	assert.NoError(t, err)

	req, err := http.NewRequest("POST", "/todo", bytes.NewReader(body))
	assert.NoError(t, err)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	resp := w.Result()

	assert.Equal(t, 201, resp.StatusCode)

	var task funcs.Task
	err = json.Unmarshal(w.Body.Bytes(), &task)
	assert.NoError(t, err)

	assert.Equal(t, newTask.Name, task.Name)
	assert.Equal(t, newTask.Done, task.Done)

	// С пустым запросом
	req, err = http.NewRequest("POST", "/todo", nil)
	assert.NoError(t, err)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	resp = w.Result()

	assert.Equal(t, 400, resp.StatusCode)
}

func TestDeleteTask(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := funcs.TaskServer()

	// Правильный запрос
	req, err := http.NewRequest("DELETE", "/todo/4", nil)
	assert.NoError(t, err)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	resp := w.Result()

	assert.Equal(t, 200, resp.StatusCode)
	expected := `{"deleted": 4}` + "\n"
	assert.JSONEq(t, expected, w.Body.String())
	assert.Equal(t, 4, len(funcs.Tasks))
	// Если запускать тесты по-отдельности, то пройдет. Если запускать весь файл сразу, то не пройдет из-за теста CreateTask

	// Неправильный запрос (несуществующий id)
	req, err = http.NewRequest("DELETE", "/todo/8", nil)
	assert.NoError(t, err)

	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	resp = w.Result()

	assert.Equal(t, 400, resp.StatusCode)
	expected = `{"Not found": 8}` + "\n"
	assert.JSONEq(t, expected, w.Body.String())
	assert.Equal(t, 4, len(funcs.Tasks))
}

// DoneTask то же самое, что и DeleteTask по тестам.
