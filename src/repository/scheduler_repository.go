package repository

import (
	"errors"

	"project/src/models"
)

// SchedulerRepository 调度仓库接口
type SchedulerRepository interface {
	SaveTask(task models.Task) (uint, error)
	GetAllTasks() ([]models.Task, error)
	GetTaskByID(id uint) (models.Task, error)
	UpdateTask(task models.Task) error
	DeleteTask(id uint) error
}

// schedulerRepository 调度仓库实现
type schedulerRepository struct {
	// 这里可以添加数据库连接等
	tasks []models.Task // 临时存储，实际应使用数据库
}

// NewSchedulerRepository 创建调度仓库实例
func NewSchedulerRepository() SchedulerRepository {
	return &schedulerRepository{
		tasks: make([]models.Task, 0),
	}
}

// SaveTask 保存任务
func (r *schedulerRepository) SaveTask(task models.Task) (uint, error) {
	// 模拟ID自增
	task.ID = uint(len(r.tasks) + 1)
	r.tasks = append(r.tasks, task)
	return task.ID, nil
}

// GetAllTasks 获取所有任务
func (r *schedulerRepository) GetAllTasks() ([]models.Task, error) {
	return r.tasks, nil
}

// GetTaskByID 根据ID获取任务
func (r *schedulerRepository) GetTaskByID(id uint) (models.Task, error) {
	for _, task := range r.tasks {
		if task.ID == id {
			return task, nil
		}
	}

	return models.Task{}, errors.New("任务不存在")
}

// UpdateTask 更新任务
func (r *schedulerRepository) UpdateTask(task models.Task) error {
	for i, t := range r.tasks {
		if t.ID == task.ID {
			r.tasks[i] = task
			return nil
		}
	}

	return errors.New("任务不存在")
}

// DeleteTask 删除任务
func (r *schedulerRepository) DeleteTask(id uint) error {
	for i, task := range r.tasks {
		if task.ID == id {
			// 从切片中删除元素
			r.tasks = append(r.tasks[:i], r.tasks[i+1:]...)
			return nil
		}
	}

	return errors.New("任务不存在")
}
