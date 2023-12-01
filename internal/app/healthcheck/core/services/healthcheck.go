package services

import (
	"context"
	"errors"
	"os"
	"strings"

	"github.com/gomodule/redigo/redis"
	"github.com/jackc/pgx/v5/pgxpool"

	rep "github.com/fsvxavier/default-vertical-slice/internal/adapters/repositories"
	"github.com/fsvxavier/default-vertical-slice/internal/core/commons/constants"
	"github.com/fsvxavier/default-vertical-slice/internal/core/domains"
	"github.com/fsvxavier/default-vertical-slice/internal/core/ports"
	"github.com/fsvxavier/default-vertical-slice/pkg/httpclient/nethttp"
)

type healthcheckService struct {
	Db     *pgxpool.Conn
	Redigo ports.IRedigoRepository
}

func NewHealthCheckService(db *pgxpool.Conn, rdbConn redis.Conn) ports.IHealthCheckService {
	return &healthcheckService{
		Db:     db,
		Redigo: rep.NewRedigoRepository(rdbConn),
	}
}

func (hlc *healthcheckService) GetHealthcheck() (healthStatus *domains.HealthCheck, err error) {
	healthStatus = &domains.HealthCheck{
		AppStatus:          constants.OK,
		DbStatus:           constants.OK,
		RdbStatus:          constants.OK,
		DrachmaStatus:      constants.OK,
		MedjatStatus:       constants.OK,
		ExchangeRateStatus: constants.OK,
	}

	err = hlc.Db.Ping(context.TODO())
	if err != nil {
		healthStatus.DbStatus = constants.ERROR
		healthStatus.DbMsg = err.Error()
	}

	err = hlc.Redigo.Ping(context.TODO())
	if err != nil {
		healthStatus.RdbStatus = constants.ERROR
		healthStatus.RdbMsg = err.Error()
	}

	healthStatus, err = hlc.checkExternalApps(healthStatus)

	// Else return notes
	return healthStatus, err
}

func (hlc *healthcheckService) checkExternalApps(actualHealthStatus *domains.HealthCheck) (healthStatus *domains.HealthCheck, err error) {
	actualStatus := []string{
		"MEDJAT_HEADER",
		"DRACHMA_HEADER",
		"EXCHANGE_RATE_HEADER",
	}

	var sliceErrors []error

	for i := range actualStatus {
		client := nethttp.NewRequester(nethttp.New())
		client.SetTimeOutRequest(30)
		client.SetBaseURL("https://vpce-078c9ba44be2cbce6-e7abdu82.execute-api.us-east-2.vpce.amazonaws.com/Live")
		client.Headers = map[string]string{
			"Content-Type": "application/json",
			"Accept":       "application/json",
		}

		switch strings.ToLower(actualStatus[i]) {
		case strings.ToLower("MEDJAT_HEADER"):
			client.Headers["x-apigw-api-id"] = os.Getenv("MEDJAT_HEADER")

			_, err = client.Get(context.TODO(), "/medjat/health")
			if err != nil {
				sliceErrors = append(sliceErrors, err)
				actualHealthStatus.MedjatStatus = "NOK"
				actualHealthStatus.MedjatMsg = err.Error()
			}

		case strings.ToLower("DRACHMA_HEADER"):
			client.Headers["x-apigw-api-id"] = os.Getenv("DRACHMA_HEADER")

			_, err = client.Get(context.TODO(), "/health")
			if err != nil {
				sliceErrors = append(sliceErrors, err)
				actualHealthStatus.DrachmaStatus = "NOK"
				actualHealthStatus.DrachmaMsg = err.Error()
			}

		case strings.ToLower("EXCHANGE_RATE_HEADER"):
			client.Headers["x-apigw-api-id"] = os.Getenv("EXCHANGE_RATE_HEADER")

			_, err = client.Get(context.TODO(), "/health")
			if err != nil {
				sliceErrors = append(sliceErrors, err)
				actualHealthStatus.ExchangeRateStatus = "NOK"
				actualHealthStatus.ExchangeRateMsg = err.Error()
			}
		}
	}

	if len(sliceErrors) > 0 {
		err = errors.New("Error in healthcheck External Apps")
	}

	return actualHealthStatus, err
}
