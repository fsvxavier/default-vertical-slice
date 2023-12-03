package ports

import "github.com/fsvxavier/default-vertical-slice/internal/app/healthcheck/core/domains"

type IHealthCheckService interface {
	GetHealthcheck() (healthStatus *domains.HealthCheck, err error)
}
