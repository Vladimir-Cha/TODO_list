package service

import (
	"net/http"
	"time"

	"github.com/Vladimir-Cha/TODO_list/internal/errors"
	"github.com/Vladimir-Cha/TODO_list/internal/models/api"
	"github.com/Vladimir-Cha/TODO_list/internal/models/storage"
	"github.com/labstack/echo/v4"
)

type Handlers struct {
	service TaskService
}

func NewHandlers(service TaskService) *Handlers {
	return &Handlers{service: service}
}

func (h *Handlers) CreateTask(c echo.Context) error {
	c.Response().Header().Set("Content-Type", "application/json")

	var apiTask api.Task
	if err := c.Bind(&apiTask); err != nil {
		return c.JSON(http.StatusBadRequest, errors.ErrBadRequest.WithMap())
	}

	task := &storage.Task{
		Title:       apiTask.Title,
		Description: apiTask.Comment,
		Date:        apiTask.Date,
		Repeat:      apiTask.Repeat,
	}

	createdTask, err := h.service.CreateTask(c.Request().Context(), task)
	if err != nil {
		return handleServiceError(c, err)
	}

	return c.JSON(http.StatusCreated, map[string]string{"id": createdTask.ID})
}

func (h *Handlers) DeleteTask(c echo.Context) error {
	c.Response().Header().Set("Content-Type", "application/json")

	idStr := c.QueryParam("id")
	if idStr == "" {
		return c.JSON(http.StatusBadRequest, errors.ErrBadRequest.WithDetails("id is required").WithMap())
	}

	if err := h.service.DeleteTask(c.Request().Context(), idStr); err != nil {
		return handleServiceError(c, err)
	}

	return c.JSON(http.StatusOK, map[string]string{})
}

func (h *Handlers) ListTasks(c echo.Context) error {
	c.Response().Header().Set("Content-Type", "application/json")

	tasks, err := h.service.ListTasks(c.Request().Context())
	if err != nil {
		return handleServiceError(c, err)
	}

	// конвеертируем storage.Task в api.Task для JSON ответа
	apiTasks := make([]*api.Task, len(tasks))
	for i, t := range tasks {
		apiTasks[i] = &api.Task{
			ID:        t.ID,
			Title:     t.Title,
			Comment:   t.Description,
			IsDone:    t.IsDone,
			Date:      t.Date,
			Repeat:    t.Repeat,
			CreatedAt: t.CreatedAt,
			UpdatedAt: t.CreatedAt,
		}
	}

	return c.JSON(http.StatusOK, map[string][]*api.Task{"tasks": apiTasks})
}

func (h *Handlers) UpdateTask(c echo.Context) error {
	c.Response().Header().Set("Content-Type", "application/json")

	var apiTask api.Task
	if err := c.Bind(&apiTask); err != nil {
		return c.JSON(http.StatusBadRequest, errors.ErrBadRequest.WithMap())
	}

	idStr := c.QueryParam("id")
	if idStr == "" {
		return c.JSON(http.StatusBadRequest, errors.ErrBadRequest.WithDetails("id is required").WithMap())
	}

	task := &storage.Task{
		Title:       apiTask.Title,
		Description: apiTask.Comment,
		Date:        apiTask.Date,
		Repeat:      apiTask.Repeat,
	}

	if err := h.service.UpdateTask(c.Request().Context(), idStr, task); err != nil {
		return handleServiceError(c, err)
	}

	return c.JSON(http.StatusOK, map[string]string{})
}

func (h *Handlers) GetTaskByID(c echo.Context) error {
	c.Response().Header().Set("Content-Type", "application/json")

	idStr := c.QueryParam("id")
	if idStr == "" {
		return c.JSON(http.StatusBadRequest, errors.ErrBadRequest.WithDetails("id is required").WithMap())
	}

	task, err := h.service.GetTaskByID(c.Request().Context(), idStr)
	if err != nil {
		return handleServiceError(c, err)
	}

	apiTask := &api.Task{
		ID:        task.ID,
		Title:     task.Title,
		Comment:   task.Description,
		IsDone:    task.IsDone,
		Date:      task.Date,
		Repeat:    task.Repeat,
		CreatedAt: task.CreatedAt,
		UpdatedAt: task.CreatedAt,
	}

	return c.JSON(http.StatusOK, apiTask)
}

func (h *Handlers) MarkTaskAsDone(c echo.Context) error {
	c.Response().Header().Set("Content-Type", "application/json")

	idStr := c.QueryParam("id")
	if idStr == "" {
		return c.JSON(http.StatusBadRequest, errors.ErrBadRequest.WithDetails("id is required").WithMap())
	}

	task, err := h.service.MarkTaskAsDone(c.Request().Context(), idStr)
	if err != nil {
		return handleServiceError(c, err)
	}

	apiTask := &api.Task{
		ID:        task.ID,
		Title:     task.Title,
		Comment:   task.Description,
		IsDone:    task.IsDone,
		Date:      task.Date,
		Repeat:    task.Repeat,
		CreatedAt: task.CreatedAt,
		UpdatedAt: task.CreatedAt,
	}

	return c.JSON(http.StatusOK, apiTask)
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

	nextDate, err := h.service.GetNextDate(c.Request().Context(), date, repeat, now)
	if err != nil {
		return c.JSON(http.StatusBadRequest, errors.ErrBadRequest.WithError(err).WithMap())
	}

	return c.String(http.StatusOK, nextDate)
}

func handleServiceError(c echo.Context, err error) error {
	if validationErr, ok := err.(*errors.Error); ok {
		return c.JSON(validationErr.Code, validationErr.WithMap())
	}
	return c.JSON(http.StatusInternalServerError, errors.ErrDatabase.WithError(err).WithMap())
}
