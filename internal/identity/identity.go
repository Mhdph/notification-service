package identity

import "context"

type Identity struct {
	UserID string
	AppID  string
}

type contextKey struct{}

func WithContext(
	ctx context.Context,
	identity Identity,
) context.Context {
	return context.WithValue(
		ctx,
		contextKey{},
		identity,
	)
}

func FromContext(
	ctx context.Context,
) (Identity, bool) {
	identity, ok :=
		ctx.Value(
			contextKey{},
		).(Identity)

	return identity, ok
}
