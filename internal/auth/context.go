package auth

import "context"

type userIDKey struct{}

type authErrorKey struct{}

func WithUserID(ctx context.Context, id int64) context.Context {
	return context.WithValue(ctx, userIDKey{}, id)
}

// UserID returns the authenticated user of the request, if any.
func UserID(ctx context.Context) (int64, bool) {
	id, ok := ctx.Value(userIDKey{}).(int64)
	return id, ok
}

func withAuthError(ctx context.Context, err error) context.Context {
	return context.WithValue(ctx, authErrorKey{}, err)
}

func authError(ctx context.Context) error {
	err, _ := ctx.Value(authErrorKey{}).(error)
	return err
}
