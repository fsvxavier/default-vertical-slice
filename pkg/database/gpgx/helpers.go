package gpgx

import "context"

func TenantIDContext(baseCtx context.Context, tenantID string) context.Context {
	if baseCtx == nil {
		baseCtx = context.TODO()
	}

	//lint:ignore SA1029 - tentant_id is string
	return context.WithValue(baseCtx, "tenant_id", tenantID)
}
