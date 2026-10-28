package auth

import (
	"go.uber.org/fx"

	"github.com/porebric/healthstep-libs/app"
	"github.com/porebric/healthstep-libs/configs"
)

func Module() fx.Option {
	return fx.Provide(newIssuer, NewMiddleware, newPasswordHasher)
}

func newIssuer(c *app.Configs) (*Issuer, error) {
	return NewIssuer(
		configs.Value(c.Ctx, "jwt_secret").String(),
		configs.Value(c.Ctx, "jwt_ttl").Duration(),
	)
}

func newPasswordHasher(c *app.Configs) PasswordHasher {
	return NewPasswordHasher(configs.Value(c.Ctx, "password_cost").Int())
}
