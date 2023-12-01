package gpgx

import "context"

type ContextualModel interface {
	Context() context.Context
	SetContext(ctx context.Context)
}
