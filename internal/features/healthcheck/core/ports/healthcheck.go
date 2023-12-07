package ports

import "github.com/fsvxavier/default-vertical-slice/internal/features/healthcheck/core/domains"

type IHealthCheckService interface {
	GetHealthcheck() (healthStatus *domains.HealthCheck, err error)
}
