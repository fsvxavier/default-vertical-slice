package gpgx

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

type MultiTenantConfig struct{}

// beforeAcquireHook should be called before a connection is acquired from the pool.
func (mtc *MultiTenantConfig) beforeAcquireHook(ctx context.Context, conn *pgx.Conn) bool {
	if tid := ctx.Value("tenant_id"); tid != nil && tid != "" {
		if tenantID, ok := tid.(string); ok {
			_, err := conn.Exec(ctx, "select set_config($1,$2,$3);", "app.current_tenant", tenantID, false)
			if err != nil {
				fmt.Printf("Failed to set tenant ID for session: %s\n", err.Error())
			}
		} else {
			fmt.Println("Tenant ID set is not a string")
			return false
		}
	} else {
		return false
	}

	return true
}

// afterReleaseHook should be called after a connection is released, but before it is returned to the pool.
func (mtc *MultiTenantConfig) afterReleaseHook(conn *pgx.Conn) bool {
	// Undo what was done in BeforeAcquire.
	// Set the configuration to empty before this connection is released to the pool.
	_, err := conn.Exec(context.TODO(), "select set_config($1,$2,$3);", "app.current_tenant", "", false)
	if err != nil {
		fmt.Printf("Failed to unset tenant ID for session: %s", err.Error())
		return false
	}

	return true
}
