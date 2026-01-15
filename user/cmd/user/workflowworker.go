package main

import (
	"errors"
	"net/http"
	"reflect"
	"unsafe"

	"gitea.xscloud.ru/xscloud/golib/pkg/application/logging"
	libio "gitea.xscloud.ru/xscloud/golib/pkg/common/io"
	"gitea.xscloud.ru/xscloud/golib/pkg/infrastructure/mysql"
	"gitea.xscloud.ru/xscloud/golib/pkg/infrastructure/outbox"
	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx"
	"github.com/urfave/cli/v2"
	"golang.org/x/sync/errgroup"

	appservice "user/pkg/user/app/service"
	"user/pkg/user/infrastructure/integrationevent"
	inframysql "user/pkg/user/infrastructure/mysql"
	"user/pkg/user/infrastructure/temporal"
	"user/pkg/user/infrastructure/temporal/worker"
)

type workflowWorkerConfig struct {
	Service  Service  `envconfig:"service"`
	Database Database `envconfig:"database" required:"true"`
	Temporal Temporal `envconfig:"temporal" required:"true"`
}

func workflowWorker(logger logging.Logger) *cli.Command {
	return &cli.Command{
		Name:   "workflow-worker",
		Before: migrateImpl(logger),
		Action: func(c *cli.Context) error {
			cnf, err := parseEnvs[workflowWorkerConfig]()
			if err != nil {
				return err
			}

			closer := libio.NewMultiCloser()
			defer func() {
				err = errors.Join(err, closer.Close())
			}()

			databaseConnector, err := newDatabaseConnector(cnf.Database)
			if err != nil {
				return err
			}
			closer.AddCloser(databaseConnector)

			val := reflect.ValueOf(databaseConnector).Elem()
			field := val.FieldByName("db")
			//nolint:gosec
			realField := reflect.NewAt(field.Type(), unsafe.Pointer(field.UnsafeAddr())).Elem()
			sqlxDB := realField.Interface().(*sqlx.DB)
			sqlDB := sqlxDB.DB

			databaseConnectionPool := mysql.NewConnectionPool(databaseConnector.TransactionalClient())

			temporalClient, err := temporal.NewClient(logger, cnf.Temporal.Host)
			if err != nil {
				return err
			}
			closer.AddCloser(libio.CloserFunc(func() error {
				temporalClient.Close()
				return nil
			}))

			libUoW := mysql.NewUnitOfWork(databaseConnectionPool, inframysql.NewRepositoryProvider)
			libLUow := mysql.NewLockableUnitOfWork(libUoW, mysql.NewLocker(databaseConnectionPool))
			uow := inframysql.NewUnitOfWork(libUoW)
			luow := inframysql.NewLockableUnitOfWork(libLUow)
			eventDispatcher := outbox.NewEventDispatcher(appID, integrationevent.TransportName, integrationevent.NewEventSerializer(), libUoW)

			errGroup := errgroup.Group{}
			errGroup.Go(func() error {
				w := worker.NewWorker(temporalClient, appservice.NewUserService(uow, luow, eventDispatcher))
				logger.Info("Worker created, starting...")
				err := w.Run(worker.InterruptChannel())
				if err != nil {
					logger.Info("Worker creating err = ", err)
				}
				return err
			})

			errGroup.Go(func() error {
				router := mux.NewRouter()
				registerHealthcheck(router)
				registerMetrics(router, sqlDB)
				// nolint:gosec
				server := http.Server{
					Addr:    cnf.Service.HTTPAddress,
					Handler: router,
				}
				graceCallback(c.Context, logger, cnf.Service.GracePeriod, server.Shutdown)
				return server.ListenAndServe()
			})

			return errGroup.Wait()
		},
	}
}
