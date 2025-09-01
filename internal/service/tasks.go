package service

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Vladimir-Cha/TODO_list/internal/errors"
	"github.com/labstack/echo/v4"
)

const dateFormat = "20060102"

type Task struct {
	ID          int64     `json:"id" db:"id"`
	Title       string    `json:"title" db:"title"`
	Description string    `json:"description" db:"description"`
	IsDone      bool      `json:"is_done" db:"is_done"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	Date        string    `json:"date" db:"date"`
	Repeat      string    `json:"repeat" db:"repeat"`
}

type Handlers struct {
	repo TaskRepository
}

// NewHandlers нужна для экспорта структуры repo из Handlers в main
func NewHandlers(repo TaskRepository) *Handlers {
	return &Handlers{repo: repo}
}

type TaskRepository interface {
	CreateTask(ctx context.Context, task *Task) (*Task, error)
	ListTasks(ctx context.Context) ([]*Task, error)
	GetTaskByID(ctx context.Context, id int64) (*Task, error)
	UpdateTask(ctx context.Context, id int64, task *Task) error
	DeleteTask(ctx context.Context, id int64) error
	MarkTaskAsDone(ctx context.Context, id int64) (*Task, error)
	GetNextDate(ctx context.Context, date, repeat string, now time.Time) (string, error)
}

func (h *Handlers) CreateTask(c echo.Context) error {
	c.Response().Header().Set("Content-Type", "application/json")

	var task Task
	if err := c.Bind(&task); err != nil {
		return c.JSON(http.StatusBadRequest, errors.ErrBadRequest.WithMap())
	}

	if task.Title == "" {
		return c.JSON(http.StatusUnprocessableEntity, errors.ErrValidation.WithDetails("title is required").WithMap())
	}

	now := time.Now().Truncate(24 * time.Hour)
	if task.Date == "today" || task.Date == "" {
		task.Date = now.Format(dateFormat)
	} else {
		if _, err := time.Parse(dateFormat, task.Date); err != nil {
			return c.JSON(http.StatusUnprocessableEntity, errors.ErrValidation.WithDetails("invalid date format").WithMap())
		}
		taskDate, _ := time.Parse(dateFormat, task.Date)
		if taskDate.Before(now) && task.Repeat != "" {
			nextDate, err := h.repo.GetNextDate(c.Request().Context(), task.Date, task.Repeat, now)
			if err != nil {
				return c.JSON(http.StatusInternalServerError, errors.ErrDatabase.WithError(err).WithMap())
			}
			task.Date = nextDate
		}
	}

	if task.Repeat != "" && !isValidRepeatRule(task.Repeat) {
		return c.JSON(http.StatusUnprocessableEntity, errors.ErrValidation.WithDetails("invalid repeat rule").WithMap())
	}

	createdTask, err := h.repo.CreateTask(c.Request().Context(), &task)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, errors.ErrDatabase.WithError(err).WithMap())
	}

	return c.JSON(http.StatusCreated, map[string]string{"id": strconv.FormatInt(createdTask.ID, 10)})
}

func (h *Handlers) DeleteTask(c echo.Context) error {
	c.Response().Header().Set("Content-Type", "application/json")

	idStr := c.QueryParam("id")
	if idStr == "" {
		return c.JSON(http.StatusBadRequest, errors.ErrBadRequest.WithDetails("id is required").WithMap())
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, errors.ErrBadRequest.WithDetails("invalid id format").WithMap())
	}

	if err := h.repo.DeleteTask(c.Request().Context(), id); err != nil {
		return c.JSON(http.StatusInternalServerError, errors.ErrDatabase.WithError(err).WithMap())
	}

	return c.JSON(http.StatusOK, map[string]string{})
}

func (h *Handlers) ListTasks(c echo.Context) error {
	c.Response().Header().Set("Content-Type", "application/json")

	tasks, err := h.repo.ListTasks(c.Request().Context())
	if err != nil {
		return c.JSON(http.StatusInternalServerError, errors.ErrDatabase.WithError(err).WithMap())
	}

	if tasks == nil {
		tasks = []*Task{}
	}

	return c.JSON(http.StatusOK, map[string][]*Task{"tasks": tasks})
}

func (h *Handlers) UpdateTask(c echo.Context) error {
	c.Response().Header().Set("Content-Type", "application/json")

	var task Task
	if err := c.Bind(&task); err != nil {
		return c.JSON(http.StatusBadRequest, errors.ErrBadRequest.WithMap())
	}

	idStr := c.QueryParam("id")
	if idStr == "" {
		return c.JSON(http.StatusBadRequest, errors.ErrBadRequest.WithDetails("id is required").WithMap())
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, errors.ErrBadRequest.WithDetails("invalid id format").WithMap())
	}

	if task.Title == "" {
		return c.JSON(http.StatusUnprocessableEntity, errors.ErrValidation.WithDetails("title is required").WithMap())
	}

	if task.Date != "" {
		if _, err := time.Parse(dateFormat, task.Date); err != nil {
			return c.JSON(http.StatusUnprocessableEntity, errors.ErrValidation.WithDetails("invalid date format").WithMap())
		}
	}

	if task.Repeat != "" && !isValidRepeatRule(task.Repeat) {
		return c.JSON(http.StatusUnprocessableEntity, errors.ErrValidation.WithDetails("invalid repeat rule").WithMap())
	}

	if err := h.repo.UpdateTask(c.Request().Context(), id, &task); err != nil {
		if err.Error() == "task not found" {
			return c.JSON(http.StatusNotFound, errors.ErrNotFound.WithMap())
		}
		return c.JSON(http.StatusInternalServerError, errors.ErrDatabase.WithError(err).WithMap())
	}

	return c.JSON(http.StatusOK, map[string]string{})
}

func (h *Handlers) GetTaskByID(c echo.Context) error {
	c.Response().Header().Set("Content-Type", "application/json")

	idStr := c.QueryParam("id")
	if idStr == "" {
		return c.JSON(http.StatusBadRequest, errors.ErrBadRequest.WithDetails("id is required").WithMap())
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, errors.ErrBadRequest.WithDetails("invalid id format").WithMap())
	}

	task, err := h.repo.GetTaskByID(c.Request().Context(), id)
	if err != nil {
		if err.Error() == "task not found" {
			return c.JSON(http.StatusNotFound, errors.ErrNotFound.WithMap())
		}
		return c.JSON(http.StatusInternalServerError, errors.ErrDatabase.WithError(err).WithMap())
	}

	return c.JSON(http.StatusOK, task)
}

func (h *Handlers) MarkTaskAsDone(c echo.Context) error {
	c.Response().Header().Set("Content-Type", "application/json")

	idStr := c.QueryParam("id")
	if idStr == "" {
		return c.JSON(http.StatusBadRequest, errors.ErrBadRequest.WithDetails("id is required").WithMap())
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, errors.ErrBadRequest.WithDetails("invalid id format").WithMap())
	}

	task, err := h.repo.GetTaskByID(c.Request().Context(), id)
	if err != nil {
		if err.Error() == "task not found" {
			return c.JSON(http.StatusNotFound, errors.ErrNotFound.WithMap())
		}
		return c.JSON(http.StatusInternalServerError, errors.ErrDatabase.WithError(err).WithMap())
	}

	// для неповторяющихся задач
	if task.Repeat == "" {
		task, err = h.repo.MarkTaskAsDone(c.Request().Context(), id)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, errors.ErrDatabase.WithError(err).WithMap())
		}
	} else {
		// для повторяющихся задач - обновляем дату и сбрасываем статус выполнения
		now := time.Now()
		nextDate, err := h.repo.GetNextDate(c.Request().Context(), task.Date, task.Repeat, now)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, errors.ErrDatabase.WithError(err).WithMap())
		}

		task.Date = nextDate
		task.IsDone = false

		if err := h.repo.UpdateTask(c.Request().Context(), id, task); err != nil {
			return c.JSON(http.StatusInternalServerError, errors.ErrDatabase.WithError(err).WithMap())
		}

		task, err = h.repo.GetTaskByID(c.Request().Context(), id)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, errors.ErrDatabase.WithError(err).WithMap())
		}
	}

	return c.JSON(http.StatusOK, task)
}

func (h *Handlers) GetNextDate(c echo.Context) error {
	c.Response().Header().Set("Content-Type", "text/plain")

	date := c.QueryParam("date")
	repeat := c.QueryParam("repeat")
	nowStr := c.QueryParam("now")

	now, err := time.Parse(dateFormat, nowStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, errors.ErrBadRequest.WithDetails("invalid 'now' parameter").WithMap())
	}

	nextDate, err := h.repo.GetNextDate(c.Request().Context(), date, repeat, now)
	if err != nil {
		return c.JSON(http.StatusBadRequest, errors.ErrBadRequest.WithError(err).WithMap())
	}

	return c.String(http.StatusOK, nextDate)
}

func isValidRepeatRule(repeat string) bool {
	if repeat == "" {
		return true
	}
	if repeat == "y" {
		return true
	}
	if strings.HasPrefix(repeat, "d ") {
		days := repeat[2:]
		_, err := strconv.Atoi(days)
		return err == nil && len(days) > 0
	}
	if strings.HasPrefix(repeat, "w ") {
		days := strings.Split(repeat[2:], ",")
		for _, day := range days {
			_, err := strconv.Atoi(day)
			if err != nil || len(day) != 1 || day < "1" || day > "7" {
				return false
			}
		}
		return true
	}
	if strings.HasPrefix(repeat, "m ") {
		parts := strings.Split(repeat[2:], " ")
		if len(parts) > 2 {
			return false
		}
		days := strings.Split(parts[0], ",")
		for _, day := range days {
			if day == "-1" || day == "-2" {
				continue
			}
			_, err := strconv.Atoi(day)
			if err != nil || len(day) > 2 || day < "1" || day > "31" {
				return false
			}
		}
		if len(parts) == 2 {
			months := strings.Split(parts[1], ",")
			for _, month := range months {
				_, err := strconv.Atoi(month)
				if err != nil || len(month) > 2 || month < "1" || month > "12" {
					return false
				}
			}
		}
		return true
	}
	return false
}
