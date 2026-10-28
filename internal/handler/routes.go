package handler

import (
	httpapi "github.com/porebric/healthstep-libs/api"
	"github.com/porebric/healthstep-libs/app"

	"github.com/porebric/task-tracker/internal/auth"
	"github.com/porebric/task-tracker/pkg/api"
)

const (
	tagAuth  = "auth"
	tagTeams = "teams"
	tagTasks = "tasks"
)

// Register wires the routes. The bearer token is resolved for every request by the HTTP
// middleware; protected routes additionally require it through auth.Required.
func Register(s *httpapi.Server, lp *app.LoggerProvider, mw *auth.Middleware, h *Handler) {
	s.Use(mw.Authenticate)
	logFn := lp.GetLogger
	protected := httpapi.With(auth.Required)

	httpapi.Post(s, logFn, "/api/v1/register", api.NewRegisterRequest, h.Register,
		httpapi.Summary("Register a user"), httpapi.Tags(tagAuth))
	httpapi.Post(s, logFn, "/api/v1/login", api.NewLoginRequest, h.Login,
		httpapi.Summary("Authenticate and get a JWT"), httpapi.Tags(tagAuth))

	httpapi.Post(s, logFn, "/api/v1/teams", api.NewCreateTeamRequest, h.CreateTeam, protected,
		httpapi.Summary("Create a team"),
		httpapi.Description("The current user becomes the owner of the team."),
		httpapi.Tags(tagTeams))
	httpapi.Get(s, logFn, "/api/v1/teams", api.NewEmptyRequest, h.ListTeams, protected,
		httpapi.Summary("List the teams of the current user"), httpapi.Tags(tagTeams))
	httpapi.Post(s, logFn, "/api/v1/teams/{id}/invite", api.NewInviteRequest, h.Invite, protected,
		httpapi.Summary("Add a user to the team"),
		httpapi.Description("Owners grant the admin or member role, admins only member. "+
			"The user is named by user_id or by email."),
		httpapi.Tags(tagTeams))
	httpapi.Get(s, logFn, "/api/v1/teams/{team_id}/stats", api.NewTeamStatsRequest, h.TeamStats, protected,
		httpapi.Summary("Team report"),
		httpapi.Description("Tasks by status, top 3 assignees by tasks closed in the last 30 days, "+
			"average time to close in seconds and comment counts. Owners and admins only."),
		httpapi.Tags(tagTeams))

	httpapi.Post(s, logFn, "/api/v1/tasks", api.NewCreateTaskRequest, h.CreateTask, protected,
		httpapi.Summary("Create a task"),
		httpapi.Description("Any member of the team can create a task; the assignee must be a member too."),
		httpapi.Tags(tagTasks))
	httpapi.Get(s, logFn, "/api/v1/tasks", api.NewListTasksRequest, h.ListTasks, protected,
		httpapi.Summary("List the tasks of a team"),
		httpapi.Description("Filter by status and assignee_id. Pages are cached for the configured TTL. "+
			"Pass cursor=<next_cursor> instead of offset for keyset pagination."),
		httpapi.Tags(tagTasks))
	httpapi.Get(s, logFn, "/api/v1/tasks/{id}", api.NewGetTaskRequest, h.GetTask, protected,
		httpapi.Summary("Get a task"), httpapi.Tags(tagTasks))
	httpapi.Put(s, logFn, "/api/v1/tasks/{id}", api.NewUpdateTaskRequest, h.UpdateTask, protected,
		httpapi.Summary("Update a task"),
		httpapi.Description("Only the fields present in the body change. Owners and admins edit any task, "+
			"the creator edits their own task, the assignee may only change the status. "+
			"Send version to reject the update when the task changed in between (409)."),
		httpapi.Tags(tagTasks))
	httpapi.Get(s, logFn, "/api/v1/tasks/{id}/history", api.NewTaskHistoryRequest, h.TaskHistory, protected,
		httpapi.Summary("Change history of a task"), httpapi.Tags(tagTasks))
	httpapi.Post(s, logFn, "/api/v1/tasks/{id}/comments", api.NewCreateCommentRequest, h.CreateComment, protected,
		httpapi.Summary("Comment on a task"), httpapi.Tags(tagTasks))
	httpapi.Get(s, logFn, "/api/v1/tasks/{id}/comments", api.NewListCommentsRequest, h.ListComments, protected,
		httpapi.Summary("List the comments of a task"), httpapi.Tags(tagTasks))
}
