package tasker

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/jinzhu/gorm"
)

const pollingDelay = time.Second
var ErrAborted = fmt.Errorf("aborted")

type Service struct {
	db *gorm.DB
	strategy func(ctx context.Context, task Task)
}

func NewService(db *gorm.DB) Service {
	return Service{
		db: db,
		strategy: defaultStrategy,
	}
}

func (s *Service) Create(ctx context.Context) (Task, error) {
	task, err := s.createTask(ctx);
	if err != nil {
		return task, err
	}
	go func() {
		log.Println("Start processing ", task)
		if err := s.process(ctx, task); err != nil {
			log.Println(err)
		}
		log.Println("End processing ", task)
	}()
	return task, nil
}

func (s *Service) Read(ctx context.Context, id uuid.UUID) (Task, error) {
	task := Task{ID: id}
	err := s.db.Find(&task).Error
	return task, err
}

func (s *Service) Poll(ctx context.Context, id uuid.UUID) (Task, error) {
	task := Task{ID: id}

	ticker := time.NewTicker(pollingDelay)
	defer ticker.Stop()

	for {
		err := s.db.Find(&task).Error
		if task.finished() {
			return task, err
		}

		select {
		case <-ctx.Done():
			return task, ErrAborted
		case _ = <-ticker.C:
		}
	}
}

func (s *Service) createTask(ctx context.Context) (Task, error) {
	task := Task{
		ID: uuid.New(),
		Status: "created",
	}
	err := s.db.Create(&task).Error
	return task, err
}

func (s *Service) process(ctx context.Context, task Task) error {
	task.setRunning()
	if err := s.db.Save(&task).Error; err != nil {
		return err
	}

	s.strategy(ctx, task)

	task.setFinished()
	return s.db.Save(&task).Error
}

func defaultStrategy(ctx context.Context, task Task) {
	time.Sleep(20 * time.Second) // Imitate processing delay
}
