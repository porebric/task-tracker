package handler

import (
	"context"

	"github.com/porebric/task-tracker/internal/domain"
	"github.com/porebric/task-tracker/internal/repo"
	"github.com/porebric/task-tracker/internal/service"
	"github.com/porebric/task-tracker/pkg/api"
)

func (h *Handler) CreateTask(ctx context.Context, in api.CreateTaskRequest) (api.CreateTaskResponse, error) {
	return withActor(ctx, func(actor int64) (api.CreateTaskResponse, error) {
		task, err := h.tasks.Create(ctx, actor, service.CreateTaskInput{
			TeamID:      in.TeamID,
			Title:       in.Title,
			Description: in.Description,
			Status:      domain.Status(in.Status),
			AssigneeID:  in.AssigneeID,
		})
		if err != nil {
			return api.CreateTaskResponse{}, err
		}
		return api.CreateTaskResponse{Task: toTask(task)}, nil
	})
}

func (h *Handler) ListTasks(ctx context.Context, in api.ListTasksRequest) (api.ListTasksResponse, error) {
	return withActor(ctx, func(actor int64) (api.ListTasksResponse, error) {
		page, err := h.tasks.List(ctx, actor, repo.TaskFilter{
			TeamID:     in.TeamID,
			Status:     domain.Status(in.Status),
			AssigneeID: in.AssigneeID,
			Limit:      in.Limit,
			Offset:     in.Offset,
			Cursor:     in.Cursor,
		})
		if err != nil {
			return api.ListTasksResponse{}, err
		}
		return api.ListTasksResponse{
			Tasks:      toTasks(page.Tasks),
			Total:      page.Total,
			Limit:      in.Limit,
			Offset:     in.Offset,
			NextCursor: page.NextCursor,
		}, nil
	})
}

func (h *Handler) GetTask(ctx context.Context, in api.GetTaskRequest) (api.GetTaskResponse, error) {
	return withActor(ctx, func(actor int64) (api.GetTaskResponse, error) {
		task, err := h.tasks.Get(ctx, actor, in.ID)
		if err != nil {
			return api.GetTaskResponse{}, err
		}
		return api.GetTaskResponse{Task: toTask(task)}, nil
	})
}

func (h *Handler) UpdateTask(ctx context.Context, in api.UpdateTaskRequest) (api.UpdateTaskResponse, error) {
	return withActor(ctx, func(actor int64) (api.UpdateTaskResponse, error) {
		input := service.UpdateTaskInput{
			Title:       in.Title,
			Description: in.Description,
			AssigneeID:  in.AssigneeID,
			Version:     in.Version,
		}
		if in.Status != nil {
			status := domain.Status(*in.Status)
			input.Status = &status
		}
		task, err := h.tasks.Update(ctx, actor, in.ID, input)
		if err != nil {
			return api.UpdateTaskResponse{}, err
		}
		return api.UpdateTaskResponse{Task: toTask(task)}, nil
	})
}

func (h *Handler) TaskHistory(ctx context.Context, in api.TaskHistoryRequest) (api.TaskHistoryResponse, error) {
	return withActor(ctx, func(actor int64) (api.TaskHistoryResponse, error) {
		entries, err := h.tasks.History(ctx, actor, in.ID)
		if err != nil {
			return api.TaskHistoryResponse{}, err
		}
		return api.TaskHistoryResponse{History: toHistory(entries)}, nil
	})
}

func (h *Handler) CreateComment(ctx context.Context, in api.CreateCommentRequest) (api.CreateCommentResponse, error) {
	return withActor(ctx, func(actor int64) (api.CreateCommentResponse, error) {
		comment, err := h.tasks.AddComment(ctx, actor, in.ID, in.Content)
		if err != nil {
			return api.CreateCommentResponse{}, err
		}
		return api.CreateCommentResponse{Comment: toComment(comment)}, nil
	})
}

func (h *Handler) ListComments(ctx context.Context, in api.ListCommentsRequest) (api.ListCommentsResponse, error) {
	return withActor(ctx, func(actor int64) (api.ListCommentsResponse, error) {
		comments, err := h.tasks.Comments(ctx, actor, in.ID)
		if err != nil {
			return api.ListCommentsResponse{}, err
		}
		return api.ListCommentsResponse{Comments: toComments(comments)}, nil
	})
}
