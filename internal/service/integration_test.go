package service_test

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	gomysql "github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/porebric/task-tracker/internal/domain"
	"github.com/porebric/task-tracker/internal/mysql"
	"github.com/porebric/task-tracker/internal/repo"
	"github.com/porebric/task-tracker/internal/service"
)

const day = 24 * time.Hour

// Integration tests need a MySQL reachable through TEST_MYSQL_DSN and are skipped otherwise.
// The database named in the DSN is created on demand and emptied before every test.
func newIntegrationStore(t *testing.T) *repo.Store {
	t.Helper()
	dsn := os.Getenv("TEST_MYSQL_DSN")
	if dsn == "" {
		t.Skip("TEST_MYSQL_DSN is not set")
	}
	ctx := context.Background()
	require.NoError(t, ensureDatabase(ctx, dsn))

	db, err := mysql.New(mysql.Config{DSN: dsn})
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	require.NoError(t, db.Migrate(ctx, "../../migrations"))
	for _, table := range []string{"task_comments", "task_history", "tasks", "team_members", "teams", "users"} {
		require.NoError(t, db.Gorm().WithContext(ctx).Exec("DELETE FROM "+table).Error)
	}
	return repo.New(db)
}

func ensureDatabase(ctx context.Context, dsn string) error {
	cfg, err := gomysql.ParseDSN(dsn)
	if err != nil {
		return err
	}
	name := cfg.DBName
	cfg.DBName = ""
	conn, err := sql.Open("mysql", cfg.FormatDSN())
	if err != nil {
		return err
	}
	defer func() { _ = conn.Close() }()
	_, err = conn.ExecContext(ctx, "CREATE DATABASE IF NOT EXISTS `"+name+"` CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci")
	return err
}

type fixture struct {
	t     *testing.T
	ctx   context.Context
	store *repo.Store
	now   time.Time
}

func newFixture(t *testing.T) *fixture {
	t.Helper()
	return &fixture{
		t:     t,
		ctx:   context.Background(),
		store: newIntegrationStore(t),
		now:   time.Now().UTC().Truncate(time.Second),
	}
}

func (f *fixture) user(name string) *domain.User {
	u := &domain.User{Email: name + "@example.com", PasswordHash: "hash", Name: name, CreatedAt: f.now}
	require.NoError(f.t, f.store.CreateUser(f.ctx, u))
	return u
}

func (f *fixture) team(owner *domain.User, name string) *domain.Team {
	team := &domain.Team{Name: name, CreatedBy: owner.ID, CreatedAt: f.now}
	require.NoError(f.t, f.store.CreateTeam(f.ctx, team))
	f.member(team, owner, domain.RoleOwner)
	return team
}

func (f *fixture) member(team *domain.Team, u *domain.User, role domain.Role) {
	require.NoError(f.t, f.store.AddMember(f.ctx, &domain.TeamMember{
		TeamID:    team.ID,
		UserID:    u.ID,
		Role:      role,
		CreatedAt: f.now,
	}))
}

type taskSpec struct {
	status     domain.Status
	assignee   *domain.User
	createdAgo time.Duration
	closedAgo  time.Duration
}

func (f *fixture) task(team *domain.Team, creator *domain.User, title string, spec taskSpec) *domain.Task {
	created := f.now.Add(-spec.createdAgo)
	task := &domain.Task{
		TeamID:    team.ID,
		Title:     title,
		Status:    spec.status,
		CreatedBy: creator.ID,
		CreatedAt: created,
		UpdatedAt: created,
		Version:   1,
	}
	if spec.assignee != nil {
		task.AssigneeID = &spec.assignee.ID
	}
	if spec.status == domain.StatusDone {
		closed := f.now.Add(-spec.closedAgo)
		task.ClosedAt = &closed
	}
	require.NoError(f.t, f.store.CreateTask(f.ctx, task))
	return task
}

func (f *fixture) comment(task *domain.Task, u *domain.User, text string) {
	require.NoError(f.t, f.store.AddComment(f.ctx, &domain.TaskComment{
		TaskID:    task.ID,
		UserID:    u.ID,
		Content:   text,
		CreatedAt: f.now,
	}))
}

func TestIntegration_TeamStats(t *testing.T) {
	f := newFixture(t)
	owner, admin, ann, bob, outsider := f.user("owner"), f.user("admin"), f.user("ann"), f.user("bob"), f.user("outsider")

	team := f.team(owner, "core")
	f.member(team, admin, domain.RoleAdmin)
	f.member(team, ann, domain.RoleMember)
	f.member(team, bob, domain.RoleMember)
	other := f.team(outsider, "other")

	todo := f.task(team, owner, "todo", taskSpec{status: domain.StatusTodo})
	f.task(team, owner, "in progress", taskSpec{status: domain.StatusInProgress, assignee: ann})
	closed := f.task(team, owner, "closed by ann", taskSpec{status: domain.StatusDone, assignee: ann, createdAgo: 5 * day, closedAgo: 3 * day})
	f.task(team, owner, "closed by ann again", taskSpec{status: domain.StatusDone, assignee: ann, createdAgo: 10 * day, closedAgo: 6 * day})
	f.task(team, owner, "closed by bob", taskSpec{status: domain.StatusDone, assignee: bob, createdAgo: 2 * day, closedAgo: day})
	f.task(team, owner, "closed by admin long ago", taskSpec{status: domain.StatusDone, assignee: admin, createdAgo: 60 * day, closedAgo: 40 * day})
	f.task(team, owner, "closed without assignee", taskSpec{status: domain.StatusDone, createdAgo: 3 * day, closedAgo: day})

	f.comment(todo, owner, "first")
	f.comment(todo, ann, "second")
	f.comment(closed, bob, "third")

	foreign := f.task(other, outsider, "foreign", taskSpec{status: domain.StatusDone, assignee: outsider, createdAgo: day})
	f.comment(foreign, outsider, "foreign comment")

	stats := service.NewStats(f.store)

	_, err := stats.Team(f.ctx, outsider.ID, team.ID)
	require.ErrorIs(t, err, domain.ErrNotTeamMember)
	_, err = stats.Team(f.ctx, ann.ID, team.ID)
	require.ErrorIs(t, err, domain.ErrRoleNotAllowed)
	_, err = stats.Team(f.ctx, owner.ID, team.ID+1000)
	require.ErrorIs(t, err, domain.ErrTeamNotFound)

	for _, viewer := range []*domain.User{owner, admin} {
		report, err := stats.Team(f.ctx, viewer.ID, team.ID)
		require.NoError(t, err)

		assert.Equal(t, map[domain.Status]int64{
			domain.StatusTodo:       1,
			domain.StatusInProgress: 1,
			domain.StatusDone:       5,
		}, report.TasksByStatus)
		assert.Equal(t, []domain.Closer{
			{UserID: ann.ID, Name: "ann", ClosedTasks: 2},
			{UserID: bob.ID, Name: "bob", ClosedTasks: 1},
		}, report.TopClosers)
		require.NotNil(t, report.AvgCloseSeconds)
		assert.InDelta(t, (29 * day / 5).Seconds(), *report.AvgCloseSeconds, 1)
		assert.EqualValues(t, 3, report.CommentsTotal)
		require.Len(t, report.CommentsByTask, 7)
		assert.Equal(t, domain.TaskCommentsCount{TaskID: todo.ID, Title: "todo", Comments: 2}, report.CommentsByTask[0])
		assert.Equal(t, domain.TaskCommentsCount{TaskID: closed.ID, Title: "closed by ann", Comments: 1}, report.CommentsByTask[2])
	}

	empty, err := stats.Team(f.ctx, outsider.ID, other.ID)
	require.NoError(t, err)
	assert.Equal(t, map[domain.Status]int64{domain.StatusTodo: 0, domain.StatusInProgress: 0, domain.StatusDone: 1}, empty.TasksByStatus)
	assert.EqualValues(t, 1, empty.CommentsTotal)
}

func TestIntegration_ConcurrentTaskUpdates(t *testing.T) {
	f := newFixture(t)
	owner := f.user("owner")
	team := f.team(owner, "core")
	tasks := service.NewTasks(f.store, nil)

	task, err := tasks.Create(f.ctx, owner.ID, service.CreateTaskInput{TeamID: team.ID, Title: "v0"})
	require.NoError(t, err)

	const writers = 8
	errs := make(chan error, writers)
	var wg sync.WaitGroup
	for i := range writers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			title := fmt.Sprintf("v%d", i+1)
			_, err := tasks.Update(f.ctx, owner.ID, task.ID, service.UpdateTaskInput{Title: &title})
			errs <- err
		}()
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		require.NoError(t, err)
	}

	final, err := f.store.TaskByID(f.ctx, task.ID)
	require.NoError(t, err)
	assert.EqualValues(t, writers+1, final.Version)

	history, err := f.store.HistoryByTask(f.ctx, task.ID)
	require.NoError(t, err)
	require.Len(t, history, writers+1)

	previous := "v0"
	for _, entry := range history[1:] {
		change := entry.Changes[service.FieldTitle]
		assert.Equal(t, previous, change.Old)
		previous, _ = change.New.(string)
	}
	assert.Equal(t, final.Title, previous)

	stale, title := int64(1), "stale"
	_, err = tasks.Update(f.ctx, owner.ID, task.ID, service.UpdateTaskInput{Title: &title, Version: &stale})
	require.ErrorIs(t, err, domain.ErrVersionConflict)
}

func TestIntegration_TaskAccessRules(t *testing.T) {
	f := newFixture(t)
	owner, ann, bob, outsider := f.user("owner"), f.user("ann"), f.user("bob"), f.user("outsider")
	team := f.team(owner, "core")
	f.member(team, ann, domain.RoleMember)
	f.member(team, bob, domain.RoleMember)
	f.team(outsider, "other")

	tasks := service.NewTasks(f.store, nil)
	task, err := tasks.Create(f.ctx, ann.ID, service.CreateTaskInput{TeamID: team.ID, Title: "task", AssigneeID: &bob.ID})
	require.NoError(t, err)

	_, err = tasks.Get(f.ctx, outsider.ID, task.ID)
	require.ErrorIs(t, err, domain.ErrNotTeamMember)
	_, err = tasks.List(f.ctx, outsider.ID, repo.TaskFilter{TeamID: team.ID, Limit: 10})
	require.ErrorIs(t, err, domain.ErrNotTeamMember)
	_, err = tasks.Create(f.ctx, outsider.ID, service.CreateTaskInput{TeamID: team.ID, Title: "x"})
	require.ErrorIs(t, err, domain.ErrNotTeamMember)
	_, err = tasks.Create(f.ctx, ann.ID, service.CreateTaskInput{TeamID: team.ID, Title: "x", AssigneeID: &outsider.ID})
	require.ErrorIs(t, err, domain.ErrAssigneeNotMember)

	done := domain.StatusDone
	updated, err := tasks.Update(f.ctx, bob.ID, task.ID, service.UpdateTaskInput{Status: &done})
	require.NoError(t, err)
	assert.NotNil(t, updated.ClosedAt)

	var forbidden *domain.ForbiddenFieldsError
	title := "renamed"
	_, err = tasks.Update(f.ctx, bob.ID, task.ID, service.UpdateTaskInput{Title: &title, AssigneeID: &ann.ID})
	require.ErrorAs(t, err, &forbidden)
	assert.Equal(t, []string{service.FieldTitle, service.FieldAssignee}, forbidden.Fields)

	_, err = tasks.Update(f.ctx, ann.ID, task.ID, service.UpdateTaskInput{Title: &title})
	require.NoError(t, err)
	none := int64(0)
	updated, err = tasks.Update(f.ctx, owner.ID, task.ID, service.UpdateTaskInput{AssigneeID: &none})
	require.NoError(t, err)
	assert.Nil(t, updated.AssigneeID)

	other := "other"
	_, err = tasks.Update(f.ctx, bob.ID, task.ID, service.UpdateTaskInput{Title: &other})
	require.ErrorAs(t, err, &forbidden)

	page, err := tasks.List(f.ctx, ann.ID, repo.TaskFilter{TeamID: team.ID, Status: domain.StatusDone, Limit: 10})
	require.NoError(t, err)
	assert.EqualValues(t, 1, page.Total)
	page, err = tasks.List(f.ctx, ann.ID, repo.TaskFilter{TeamID: team.ID, AssigneeID: bob.ID, Limit: 10})
	require.NoError(t, err)
	assert.Empty(t, page.Tasks)

	history, err := tasks.History(f.ctx, ann.ID, task.ID)
	require.NoError(t, err)
	assert.Len(t, history, 4)
}

func TestIntegration_TeamInvites(t *testing.T) {
	f := newFixture(t)
	owner, admin, member, newbie := f.user("owner"), f.user("admin"), f.user("member"), f.user("newbie")
	teams := service.NewTeams(f.store)

	team, err := teams.Create(f.ctx, owner.ID, "core")
	require.NoError(t, err)
	role, err := f.store.MemberRole(f.ctx, team.ID, owner.ID)
	require.NoError(t, err)
	assert.Equal(t, domain.RoleOwner, role)

	_, err = teams.Invite(f.ctx, owner.ID, team.ID, service.InviteInput{UserID: admin.ID, Role: domain.RoleAdmin})
	require.NoError(t, err)
	_, err = teams.Invite(f.ctx, admin.ID, team.ID, service.InviteInput{Email: member.Email})
	require.NoError(t, err)
	_, err = teams.Invite(f.ctx, admin.ID, team.ID, service.InviteInput{UserID: newbie.ID, Role: domain.RoleAdmin})
	require.ErrorIs(t, err, domain.ErrRoleNotAllowed)
	_, err = teams.Invite(f.ctx, member.ID, team.ID, service.InviteInput{UserID: newbie.ID})
	require.ErrorIs(t, err, domain.ErrRoleNotAllowed)
	_, err = teams.Invite(f.ctx, owner.ID, team.ID, service.InviteInput{UserID: newbie.ID, Role: domain.RoleOwner})
	require.ErrorIs(t, err, domain.ErrRoleNotAllowed)
	_, err = teams.Invite(f.ctx, owner.ID, team.ID, service.InviteInput{UserID: member.ID})
	require.ErrorIs(t, err, domain.ErrAlreadyMember)
	_, err = teams.Invite(f.ctx, newbie.ID, team.ID, service.InviteInput{UserID: newbie.ID})
	require.ErrorIs(t, err, domain.ErrNotTeamMember)
	_, err = teams.Invite(f.ctx, owner.ID, team.ID+1000, service.InviteInput{UserID: newbie.ID})
	require.ErrorIs(t, err, domain.ErrTeamNotFound)
	_, err = teams.Invite(f.ctx, owner.ID, team.ID, service.InviteInput{Email: "nobody@example.com"})
	require.ErrorIs(t, err, domain.ErrUserNotFound)

	visible, err := teams.ListForUser(f.ctx, newbie.ID)
	require.NoError(t, err)
	assert.Empty(t, visible)
	visible, err = teams.ListForUser(f.ctx, member.ID)
	require.NoError(t, err)
	require.Len(t, visible, 1)
	assert.Equal(t, team.ID, visible[0].ID)
	assert.Equal(t, domain.RoleMember, visible[0].Role)
}
