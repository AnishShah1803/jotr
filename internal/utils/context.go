package utils

import (
	"context"
)

type configContextKey struct{}

var configKey = &configContextKey{}

// WithConfig stores configuration in the context for retrieval later.
func WithConfig(ctx context.Context, cfg interface{}) context.Context {
	return context.WithValue(ctx, configKey, cfg)
}

// GetConfigFromContext retrieves configuration from the context.
func GetConfigFromContext(ctx context.Context) (interface{}, bool) {
	cfg, ok := ctx.Value(configKey).(interface{})
	return cfg, ok
}

// WithVerboseContext stores verbose flag in the context.
func WithVerboseContext(ctx context.Context, verbose bool) context.Context {
	verboseKey := struct{}{}
	return context.WithValue(ctx, verboseKey, verbose)
}

// GetVerboseFromContext retrieves the verbose flag from the context.
func GetVerboseFromContext(ctx context.Context) bool {
	if verbose, ok := ctx.Value(struct{}{}).(bool); ok {
		return verbose
	}

	return false
}
