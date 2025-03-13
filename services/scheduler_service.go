package services

import (
	"time"

	"project/models"
	"project/repository"
)

// SchedulerService 调度服务接口
type SchedulerService interface {
	CreateTask(task models.Task) (uint, error)
	GetAllTasks() ([]models.Task, error)
	GetTaskByID(id uint) (models.Task, error)
	UpdateTask(task models.Task) error
	DeleteTask(id uint) error
}

// schedulerService 调度服务实现
type schedulerService struct {
	repo repository.SchedulerRepository
}

// NewSchedulerService 创建调度服务实例
func NewSchedulerService() SchedulerService {
	return &schedulerService{
		repo: repository.NewSchedulerRepository(),
	}
}

// CreateTask 创建任务
func (s *schedulerService) CreateTask(task models.Task) (uint, error) {
	// 设置任务创建时间
	task.CreatedAt = time.Now()
	task.UpdatedAt = time.Now()
	task.Status = "pending"

	return s.repo.SaveTask(task)
}

// GetAllTasks 获取所有任务
func (s *schedulerService) GetAllTasks() ([]models.Task, error) {
	return s.repo.GetAllTasks()
}

// GetTaskByID 根据ID获取任务
func (s *schedulerService) GetTaskByID(id uint) (models.Task, error) {
	return s.repo.GetTaskByID(id)
}

// UpdateTask 更新任务
func (s *schedulerService) UpdateTask(task models.Task) error {
	task.UpdatedAt = time.Now()
	return s.repo.UpdateTask(task)
}

// DeleteTask 删除任务
func (s *schedulerService) DeleteTask(id uint) error {
	return s.repo.DeleteTask(id)
}
