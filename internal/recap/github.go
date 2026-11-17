package recap

import (
	"context"
	"fmt"
	"sync"

	"github.com/cli/go-gh/v2/pkg/api"
)

// API matches go-gh's struct-based query method and can be replaced in tests.
type API interface {
	QueryWithContext(ctx context.Context, name string, query any, variables map[string]any) error
}

// GitHub reuses a client configured from gh's existing authentication.
// Lazy initialization lets demo mode run without credentials.
type GitHub struct {
	once    sync.Once
	client  *api.GraphQLClient
	initErr error
}

func (g *GitHub) QueryWithContext(ctx context.Context, name string, query any, variables map[string]any) error {
	g.once.Do(func() { g.client, g.initErr = api.DefaultGraphQLClient() })
	if g.initErr != nil {
		return fmt.Errorf("initialize GitHub client: %w", g.initErr)
	}
	if err := g.client.QueryWithContext(ctx, name, query, variables); err != nil {
		return fmt.Errorf("GitHub request failed: %w", err)
	}
	return nil
}

func Viewer(ctx context.Context, client API) (string, error) {
	var query viewerQuery
	err := client.QueryWithContext(ctx, "RecapViewer", &query, nil)
	return query.Viewer.Login, err
}
