package view

import "context"

type Page interface {
	Name() string
	Activate(ctx context.Context)
}
