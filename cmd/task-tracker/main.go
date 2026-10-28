package main

import (
	"go.uber.org/fx"

	httpapi "github.com/porebric/healthstep-libs/api"
	"github.com/porebric/healthstep-libs/app"

	"github.com/porebric/task-tracker/internal/auth"
	"github.com/porebric/task-tracker/internal/cache"
	"github.com/porebric/task-tracker/internal/handler"
	"github.com/porebric/task-tracker/internal/mysql"
	"github.com/porebric/task-tracker/internal/repo"
	"github.com/porebric/task-tracker/internal/service"
)

func main() {
	app.Run(
		mysql.Module(),
		cache.Module(),
		auth.Module(),
		httpapi.Module(),
		fx.Provide(
			repo.New,
			service.NewAuth,
			service.NewTeams,
			service.NewTasks,
			service.NewStats,
			handler.New,
		),
		fx.Invoke(handler.Register),
	)
}
