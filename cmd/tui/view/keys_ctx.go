// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package view

import (
	"context"
	"errors"
)

// ContextKey represents context key.
type ContextKey string

// A collection of context keys.
const (
	KeyApp ContextKey = "app"
)

func extractApp(ctx context.Context) (*App, error) {
	app, ok := ctx.Value(KeyApp).(*App)
	if !ok {
		return nil, errors.New("no application found in context")
	}

	return app, nil
}
