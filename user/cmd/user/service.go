package main

import (
	"context"
	"errors"
	"net"
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
	"google.golang.org/grpc"

	"user/api/server/userpublicapi"
	appservice "user/pkg/user/app/service"
	"user/pkg/user/infrastructure/integrationevent"
	inframysql "user/pkg/user/infrastructure/mysql"
	"user/pkg/user/infrastructure/mysql/query"
	"user/pkg/user/infrastructure/transport"
	"user/pkg/user/infrastructure/transport/middlewares"
)

type serviceConfig struct {
	Service  Service  `envconfig:"service"`
	Database Database `envconfig:"database" required:"true"`
}

func service(logger logging.Logger) *cli.Command {
	return &cli.Command{
		Name:   "service",
		Before: migrateImpl(logger),
		Action: func(c *cli.Context) error {
			cnf, err := parseEnvs[serviceConfig]()
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

			libUoW := mysql.NewUnitOfWork(databaseConnectionPool, inframysql.NewRepositoryProvider)
			libLUow := mysql.NewLockableUnitOfWork(libUoW, mysql.NewLocker(databaseConnectionPool))
			uow := inframysql.NewUnitOfWork(libUoW)
			luow := inframysql.NewLockableUnitOfWork(libLUow)
			eventDispatcher := outbox.NewEventDispatcher(appID, integrationevent.TransportName, integrationevent.NewEventSerializer(), libUoW)

			userPublicAPIServer := transport.NewUserInternalAPI(
				query.NewUserQueryService(databaseConnector.TransactionalClient()),
				appservice.NewUserService(uow, luow, eventDispatcher),
			)

			errGroup := errgroup.Group{}
			errGroup.Go(func() error {
				listener, err := net.Listen("tcp", cnf.Service.GRPCAddress)
				if err != nil {
					return err
				}
				grpcServer := grpc.NewServer(grpc.ChainUnaryInterceptor(
					middlewares.NewGRPCLoggingMiddleware(logger),
				))
				userpublicapi.RegisterUserPublicAPIServer(grpcServer, userPublicAPIServer)
				graceCallback(c.Context, logger, cnf.Service.GracePeriod, func(_ context.Context) error {
					grpcServer.GracefulStop()
					return nil
				})
				return grpcServer.Serve(listener)
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
