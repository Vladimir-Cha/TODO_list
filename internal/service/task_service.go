package service

import (
	"context"
	"time"

	"github.com/Vladimir-Cha/TODO_list/internal/errors"
	"github.com/Vladimir-Cha/TODO_list/internal/models/storage"
	"github.com/Vladimir-Cha/TODO_list/internal/repository"
)

const dateFormat = "20060102"

type TaskService interface {
	CreateTask(ctx context.Context, task *storage.Task) (*storage.Task, error)
	ListTasks(ctx context.Context) ([]*storage.Task, error)
	GetTaskByID(ctx context.Context, id string) (*storage.Task, error)
	UpdateTask(ctx context.Context, id string, task *storage.Task) error
	DeleteTask(ctx context.Context, id string) error
	MarkTaskAsDone(ctx context.Context, id string) (*storage.Task, error)
	GetNextDate(ctx context.Context, date, repeat string, now time.Time) (string, error)
}

type taskService struct {
	repo repository.TaskRepository
}

func NewTaskService(repo repository.TaskRepository) TaskService {
	return &taskService{repo: repo}
}

func (s *taskService) CreateTask(ctx context.Context, task *storage.Task) (*storage.Task, error) {
	if task.Title == "" {
		return nil, errors.ErrValidation.WithDetails("title is required")
	}

	now := time.Now().Truncate(24 * time.Hour)
	if task.Date == "today" || task.Date == "" {
		task.Date = now.Format(dateFormat)
	} else {
		taskDate, err := time.Parse(dateFormat, task.Date)
		if err != nil {
			return nil, errors.ErrValidation.WithDetails("invalid date format")
		}
		if taskDate.Before(now) && task.Repeat != "" {
			nextDate, err := s.GetNextDate(ctx, task.Date, task.Repeat, now)
			if err != nil {
				return nil, err
			}
			task.Date = nextDate
		}
	}

	if task.Repeat != "" && !isValidRepeatRule(task.Repeat) {
		return nil, errors.ErrValidation.WithDetails("invalid repeat rule")
	}

	return s.repo.CreateTask(ctx, task)
}

func (s *taskService) ListTasks(ctx context.Context) ([]*storage.Task, error) {
	return s.repo.ListTasks(ctx)
}

func (s *taskService) GetTaskByID(ctx context.Context, id string) (*storage.Task, error) {
	return s.repo.GetTaskByID(ctx, id)
}

func (s *taskService) UpdateTask(ctx context.Context, id string, task *storage.Task) error {
	return s.repo.UpdateTask(ctx, id, task)
}

func (s *taskService) DeleteTask(ctx context.Context, id string) error {
	return s.repo.DeleteTask(ctx, id)
}

func (s *taskService) MarkTaskAsDone(ctx context.Context, id string) (*storage.Task, error) {
	task, err := s.GetTaskByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if task.Repeat == "" {
		return s.repo.MarkTaskAsDone(ctx, id)
	}

	// для повторяющихся тасок обновляем дату и сбрасываем IsDone
	now := time.Now()
	nextDate, err := s.GetNextDate(ctx, task.Date, task.Repeat, now)
	if err != nil {
		return nil, err
	}

	task.Date = nextDate
	task.IsDone = false

	if err := s.UpdateTask(ctx, id, task); err != nil {
		return nil, err
	}

	return s.GetTaskByID(ctx, id)
}

func (s *taskService) GetNextDate(ctx context.Context, date, repeat string, now time.Time) (string, error) {
	return s.repo.GetNextDate(ctx, date, repeat, now)
}
