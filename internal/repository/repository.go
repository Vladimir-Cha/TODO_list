package repository

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/Vladimir-Cha/TODO_list/internal/errors"
	"github.com/Vladimir-Cha/TODO_list/internal/models/storage"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const dateFormat = "20060102"

type TaskRepository interface {
	CreateTask(ctx context.Context, task *storage.Task) (*storage.Task, error)
	ListTasks(ctx context.Context) ([]*storage.Task, error)
	GetTaskByID(ctx context.Context, id string) (*storage.Task, error)
	UpdateTask(ctx context.Context, id string, task *storage.Task) error
	DeleteTask(ctx context.Context, id string) error
	MarkTaskAsDone(ctx context.Context, id string) (*storage.Task, error)
	GetNextDate(ctx context.Context, date, repeat string, now time.Time) (string, error)
}

type taskRepository struct {
	db *pgxpool.Pool
}

func NewTaskRepository(db *pgxpool.Pool) *taskRepository {
	return &taskRepository{db: db}
}

func (r *taskRepository) CreateTask(ctx context.Context, task *storage.Task) (*storage.Task, error) {
	query := `INSERT INTO tasks (title, description, is_done, created_at, date, repeat) 
              VALUES ($1, $2, $3, $4, $5, $6) 
              RETURNING id, created_at`

	var id int64
	err := r.db.QueryRow(ctx, query,
		task.Title,
		task.Description,
		task.IsDone,
		time.Now(),
		task.Date,
		task.Repeat,
	).Scan(&id, &task.CreatedAt)

	if err != nil {
		return nil, errors.ErrDatabase.WithError(err)
	}
	task.ID = strconv.FormatInt(id, 10) // Convert DB int64 to string
	return task, nil
}

func (r *taskRepository) ListTasks(ctx context.Context) ([]*storage.Task, error) {
	query := `SELECT id, title, description, is_done, created_at, date, repeat FROM tasks ORDER BY created_at`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, errors.ErrDatabase.WithError(err)
	}
	defer rows.Close()

	var tasks []*storage.Task
	for rows.Next() {
		var task storage.Task
		var id int64
		err := rows.Scan(
			&id,
			&task.Title,
			&task.Description,
			&task.IsDone,
			&task.CreatedAt,
			&task.Date,
			&task.Repeat,
		)
		if err != nil {
			return nil, errors.ErrDatabase.WithError(err)
		}
		task.ID = strconv.FormatInt(id, 10)
		tasks = append(tasks, &task)
	}

	if err := rows.Err(); err != nil {
		return nil, errors.ErrDatabase.WithError(err)
	}

	return tasks, nil
}

func (r *taskRepository) GetTaskByID(ctx context.Context, id string) (*storage.Task, error) {
	idInt, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return nil, errors.ErrBadRequest.WithDetails("invalid id format")
	}

	var task storage.Task
	query := `SELECT id, title, description, is_done, created_at, date, repeat FROM tasks WHERE id = $1`

	var dbID int64
	err = r.db.QueryRow(ctx, query, idInt).Scan(
		&dbID,
		&task.Title,
		&task.Description,
		&task.IsDone,
		&task.CreatedAt,
		&task.Date,
		&task.Repeat,
	)

	if err == pgx.ErrNoRows {
		return nil, errors.ErrNotFound
	}
	if err != nil {
		return nil, errors.ErrDatabase.WithError(err)
	}
	task.ID = strconv.FormatInt(dbID, 10)
	return &task, nil
}

func (r *taskRepository) UpdateTask(ctx context.Context, id string, task *storage.Task) error {
	idInt, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return errors.ErrBadRequest.WithDetails("invalid id format")
	}

	query := `UPDATE tasks SET title = $1, description = $2, is_done = $3, date = $4, repeat = $5 
              WHERE id = $6`

	result, err := r.db.Exec(ctx, query,
		task.Title,
		task.Description,
		task.IsDone,
		task.Date,
		task.Repeat,
		idInt,
	)

	if err != nil {
		return errors.ErrDatabase.WithError(err)
	}

	if result.RowsAffected() == 0 {
		return errors.ErrNotFound
	}
	return nil
}

func (r *taskRepository) DeleteTask(ctx context.Context, id string) error {
	idInt, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return errors.ErrBadRequest.WithDetails("invalid id format")
	}

	query := `DELETE FROM tasks WHERE id = $1`

	result, err := r.db.Exec(ctx, query, idInt)
	if err != nil {
		return errors.ErrDatabase.WithError(err)
	}

	if result.RowsAffected() == 0 {
		return errors.ErrNotFound
	}
	return nil
}

func (r *taskRepository) MarkTaskAsDone(ctx context.Context, id string) (*storage.Task, error) {
	idInt, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return nil, errors.ErrBadRequest.WithDetails("invalid id format")
	}

	var task storage.Task
	query := `UPDATE tasks SET is_done = true WHERE id = $1 
              RETURNING id, title, description, is_done, created_at, date, repeat`

	var dbID int64
	err = r.db.QueryRow(ctx, query, idInt).Scan(
		&dbID,
		&task.Title,
		&task.Description,
		&task.IsDone,
		&task.CreatedAt,
		&task.Date,
		&task.Repeat,
	)

	if err == pgx.ErrNoRows {
		return nil, errors.ErrNotFound
	}
	if err != nil {
		return nil, errors.ErrDatabase.WithError(err)
	}
	task.ID = strconv.FormatInt(dbID, 10)
	return &task, nil
}

func (r *taskRepository) GetNextDate(ctx context.Context, date, repeat string, now time.Time) (string, error) {
	taskDate, err := time.Parse(dateFormat, date)
	if err != nil {
		return "", errors.ErrValidation.WithDetails("invalid date format")
	}

	switch {
	case strings.HasPrefix(repeat, "d "):
		days, err := strconv.Atoi(repeat[2:])
		if err != nil {
			return "", errors.ErrValidation.WithDetails("invalid days value")
		}
		if days < 1 || days > 400 {
			return "", errors.ErrValidation.WithDetails("days must be between 1 and 400")
		}
		nextDate := taskDate
		for {
			nextDate = nextDate.AddDate(0, 0, days)
			if nextDate.After(now) {
				break
			}
		}
		return nextDate.Format(dateFormat), nil

	case repeat == "y":
		nextDate := taskDate
		for {
			nextDate = nextDate.AddDate(1, 0, 0)
			if nextDate.Month() == time.February && nextDate.Day() == 29 && !isLeapYear(nextDate.Year()) {
				nextDate = nextDate.AddDate(0, 0, 1) // Move to March 1st
			}
			if nextDate.After(now) {
				break
			}
		}
		return nextDate.Format(dateFormat), nil

	case strings.HasPrefix(repeat, "w "):
		daysOfWeek := strings.Split(repeat[2:], ",")
		return calculateNextWeekDate(taskDate, daysOfWeek, now)

	case strings.HasPrefix(repeat, "m "):
		parts := strings.Split(repeat[2:], " ")
		daysOfMonth := strings.Split(parts[0], ",")
		var months []string
		if len(parts) > 1 {
			months = strings.Split(parts[1], ",")
		}
		return calculateNextMonthDate(taskDate, daysOfMonth, months, now)

	default:
		return "", errors.ErrValidation.WithDetails("invalid repeat rule")
	}
}

func isLeapYear(year int) bool {
	return (year%4 == 0 && year%100 != 0) || year%400 == 0
}

func calculateNextWeekDate(taskDate time.Time, daysOfWeek []string, now time.Time) (string, error) {
	for {
		taskDate = taskDate.AddDate(0, 0, 1)
		currentWeekday := int(taskDate.Weekday())
		if currentWeekday == 0 {
			currentWeekday = 7
		}

		for _, dayStr := range daysOfWeek {
			day, err := strconv.Atoi(dayStr)
			if err != nil || day < 1 || day > 7 {
				return "", errors.ErrValidation.WithDetails("invalid day of week")
			}

			if currentWeekday == day && taskDate.After(now) {
				return taskDate.Format(dateFormat), nil
			}
		}
	}
}

func calculateNextMonthDate(taskDate time.Time, daysOfMonth []string, months []string, now time.Time) (string, error) {
	for _, dayStr := range daysOfMonth {
		if dayStr != "-1" && dayStr != "-2" {
			day, err := strconv.Atoi(dayStr)
			if err != nil || day < 1 || day > 31 {
				return "", errors.ErrValidation.WithDetails("invalid day of month")
			}
		}
	}

	for {
		taskDate = taskDate.AddDate(0, 0, 1)
		day := taskDate.Day()
		month := int(taskDate.Month())

		dayMatch := false
		for _, dayStr := range daysOfMonth {
			if dayStr == "-1" {
				if day == lastDayOfMonth(taskDate) {
					dayMatch = true
					break
				}
			} else if dayStr == "-2" {
				if day == lastDayOfMonth(taskDate)-1 {
					dayMatch = true
					break
				}
			} else {
				dayInt, err := strconv.Atoi(dayStr)
				if err == nil && dayInt == day {
					dayMatch = true
					break
				}
			}
		}

		monthMatch := true
		if len(months) > 0 {
			monthMatch = false
			for _, monthStr := range months {
				monthInt, err := strconv.Atoi(monthStr)
				if err == nil && monthInt == month {
					monthMatch = true
					break
				}
			}
		}

		if dayMatch && monthMatch && (taskDate.After(now) || taskDate.Equal(now)) {
			return taskDate.Format(dateFormat), nil
		}
	}
}

func lastDayOfMonth(date time.Time) int {
	return time.Date(date.Year(), date.Month()+1, 0, 0, 0, 0, 0, time.UTC).Day()
}
