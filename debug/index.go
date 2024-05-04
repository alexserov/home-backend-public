package debug

import "context"

func MakeLocalDevelopmentContext() context.Context {
	todo := context.TODO()
	local := context.WithValue(todo, "isLocalDevelopment", true)
	return local
}

func IsLocalDevelopmentContext(ctx context.Context) bool {
	return ctx.Value("isLocalDevelopment") == true
}
