package tasks

import (
	"context"
	"fmt"
	"time"

	"api-go/internal/common"

	"github.com/patrickmn/go-cache"
	gtasks "google.golang.org/api/tasks/v1"
	"google.golang.org/api/option"
)

type Service struct {
	cache *cache.Cache
}

func NewService() *Service {
	return &Service{
		cache: cache.New(30*time.Second, 60*time.Second),
	}
}

func (s *Service) getTasksService(ctx context.Context, opts ...option.ClientOption) (*gtasks.Service, error) {
	return gtasks.NewService(ctx, opts...)
}

func (s *Service) ListTaskLists(ctx context.Context, maxResults int64, pageToken string, opts ...option.ClientOption) (*common.PaginatedResponse[*gtasks.TaskList], error) {
	cacheKey := fmt.Sprintf("tasks:lists:%d:%s", maxResults, pageToken)
	if cached, found := s.cache.Get(cacheKey); found {
		return cached.(*common.PaginatedResponse[*gtasks.TaskList]), nil
	}

	svc, err := s.getTasksService(ctx, opts...)
	if err != nil {
		return nil, err
	}

	call := svc.Tasklists.List().MaxResults(maxResults)
	if pageToken != "" {
		call = call.PageToken(pageToken)
	}

	result, err := call.Do()
	if err != nil {
		return nil, err
	}

	resp := common.NewPaginatedResponse(result.Items, result.NextPageToken)
	s.cache.Set(cacheKey, &resp, cache.DefaultExpiration)
	return &resp, nil
}

func (s *Service) CreateTaskList(ctx context.Context, req *CreateTaskListRequest, opts ...option.ClientOption) (*gtasks.TaskList, error) {
	svc, err := s.getTasksService(ctx, opts...)
	if err != nil {
		return nil, err
	}

	taskList := &gtasks.TaskList{Title: req.Title}
	created, err := svc.Tasklists.Insert(taskList).Do()
	if err != nil {
		return nil, err
	}

	s.invalidateCache()
	return created, nil
}

func (s *Service) ListTasks(ctx context.Context, taskListID string, maxResults int64, pageToken string, opts ...option.ClientOption) (*common.PaginatedResponse[*gtasks.Task], error) {
	if taskListID == "" {
		taskListID = "@default"
	}

	cacheKey := fmt.Sprintf("tasks:%s:%d:%s", taskListID, maxResults, pageToken)
	if cached, found := s.cache.Get(cacheKey); found {
		return cached.(*common.PaginatedResponse[*gtasks.Task]), nil
	}

	svc, err := s.getTasksService(ctx, opts...)
	if err != nil {
		return nil, err
	}

	call := svc.Tasks.List(taskListID).MaxResults(maxResults)
	if pageToken != "" {
		call = call.PageToken(pageToken)
	}

	result, err := call.Do()
	if err != nil {
		return nil, err
	}

	resp := common.NewPaginatedResponse(result.Items, result.NextPageToken)
	s.cache.Set(cacheKey, &resp, cache.DefaultExpiration)
	return &resp, nil
}

func (s *Service) CreateTask(ctx context.Context, taskListID string, req *CreateTaskRequest, opts ...option.ClientOption) (*gtasks.Task, error) {
	if taskListID == "" {
		taskListID = "@default"
	}

	svc, err := s.getTasksService(ctx, opts...)
	if err != nil {
		return nil, err
	}

	task := &gtasks.Task{
		Title: req.Title,
		Notes: req.Notes,
		Due:   req.Due,
	}
	if req.Status != "" {
		task.Status = string(req.Status)
	}

	created, err := svc.Tasks.Insert(taskListID, task).Do()
	if err != nil {
		return nil, err
	}

	s.invalidateCache()
	return created, nil
}

func (s *Service) GetTask(ctx context.Context, taskListID, taskID string, opts ...option.ClientOption) (*gtasks.Task, error) {
	svc, err := s.getTasksService(ctx, opts...)
	if err != nil {
		return nil, err
	}

	task, err := svc.Tasks.Get(taskListID, taskID).Do()
	if err != nil {
		return nil, err
	}
	return task, nil
}

func (s *Service) UpdateTask(ctx context.Context, taskListID, taskID string, req *UpdateTaskRequest, opts ...option.ClientOption) (*gtasks.Task, error) {
	svc, err := s.getTasksService(ctx, opts...)
	if err != nil {
		return nil, err
	}

	existing, err := svc.Tasks.Get(taskListID, taskID).Do()
	if err != nil {
		return nil, err
	}

	if req.Title != "" {
		existing.Title = req.Title
	}
	if req.Notes != nil {
		existing.Notes = *req.Notes
	}
	if req.Due != nil {
		existing.Due = *req.Due
	}
	if req.Status != "" {
		existing.Status = string(req.Status)
	}
	if req.Completed != nil {
		if *req.Completed {
			existing.Status = string(TaskStatusCompleted)
		} else {
			existing.Status = string(TaskStatusNeedsAction)
			existing.Completed = nil
		}
	}

	updated, err := svc.Tasks.Update(taskListID, taskID, existing).Do()
	if err != nil {
		return nil, err
	}

	s.invalidateCache()
	return updated, nil
}

func (s *Service) DeleteTask(ctx context.Context, taskListID, taskID string, opts ...option.ClientOption) error {
	svc, err := s.getTasksService(ctx, opts...)
	if err != nil {
		return err
	}

	if err := svc.Tasks.Delete(taskListID, taskID).Do(); err != nil {
		return err
	}

	s.invalidateCache()
	return nil
}

func (s *Service) CompleteTask(ctx context.Context, taskListID, taskID string, opts ...option.ClientOption) (*gtasks.Task, error) {
	svc, err := s.getTasksService(ctx, opts...)
	if err != nil {
		return nil, err
	}

	existing, err := svc.Tasks.Get(taskListID, taskID).Do()
	if err != nil {
		return nil, err
	}

	existing.Status = string(TaskStatusCompleted)
	updated, err := svc.Tasks.Update(taskListID, taskID, existing).Do()
	if err != nil {
		return nil, err
	}

	s.invalidateCache()
	return updated, nil
}

func (s *Service) UncompleteTask(ctx context.Context, taskListID, taskID string, opts ...option.ClientOption) (*gtasks.Task, error) {
	svc, err := s.getTasksService(ctx, opts...)
	if err != nil {
		return nil, err
	}

	existing, err := svc.Tasks.Get(taskListID, taskID).Do()
	if err != nil {
		return nil, err
	}

	existing.Status = string(TaskStatusNeedsAction)
	existing.Completed = nil
	updated, err := svc.Tasks.Update(taskListID, taskID, existing).Do()
	if err != nil {
		return nil, err
	}

	s.invalidateCache()
	return updated, nil
}

func (s *Service) invalidateCache() {
	s.cache.Flush()
}
