package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	"github.com/kelseyhightower/envconfig"
	grpc_zerolog "github.com/philip-bui/grpc-zerolog"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"golang.org/x/sync/errgroup"
	"golang.org/x/xerrors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/peer"

	pb "github.com/kzmake/greeter/api/greeter/v1"
	"github.com/kzmake/greeter/handler"
)

type Env struct {
	Address string `default:"0.0.0.0:50051"`
	MTLS    bool   `default:"true"`
}

const (
	prefix  = "SERVICE"
	crtFile = "certs/server.greeter.crt"
	keyFile = "certs/server.greeter.key"
	caFile  = "certs/ca.crt"
)

var (
	env   Env
	creds credentials.TransportCredentials
)

func init() {
	if err := envconfig.Process(prefix, &env); err != nil {
		panic(err)
	}

	log.Logger = zerolog.New(zerolog.ConsoleWriter{Out: os.Stdout}).With().Timestamp().Logger()

	log.Debug().Msgf("%+v", env)

	if env.MTLS {
		var err error
		creds, err = loadCreds()
		if err != nil {
			log.Panic().Msgf("%+v", err)
		}
	}
}

func loadCreds() (credentials.TransportCredentials, error) {
	cert, err := tls.LoadX509KeyPair(crtFile, keyFile)
	if err != nil {
		return nil, xerrors.Errorf("failed to load %s or %s: %w", crtFile, keyFile, err)
	}

	ca, err := ioutil.ReadFile(caFile)
	if err != nil {
		return nil, xerrors.Errorf("failed to load %s: %w", caFile, err)
	}

	cp := x509.NewCertPool()
	if !cp.AppendCertsFromPEM(ca) {
		return nil, xerrors.Errorf("failed to append certificates")
	}

	return credentials.NewTLS(&tls.Config{
		ClientAuth:   tls.RequireAndVerifyClientCert,
		Certificates: []tls.Certificate{cert},
		ClientCAs:    cp,
	}), nil
}

func newServer() *grpc.Server {
	s := grpc.NewServer(
		grpc.Creds(creds),
		grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
			func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
				if p, ok := peer.FromContext(ctx); ok {
					if mtls, ok := p.AuthInfo.(credentials.TLSInfo); ok {
						for _, item := range mtls.State.PeerCertificates {
							log.Info().Msgf("request by %s", item.Subject)
						}
					}
				}

				return handler(ctx, req)
			},
			grpc_zerolog.NewUnaryServerInterceptorWithLogger(&log.Logger),
			grpc_recovery.UnaryServerInterceptor(),
		)),
	)

	pb.RegisterGreeterServer(s, handler.NewGreeter())

	return s
}

func run() error {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	g, ctx := errgroup.WithContext(ctx)

	grpcS := newServer()
	g.Go(func() error {
		lis, err := net.Listen("tcp", env.Address)
		if err != nil {
			return xerrors.Errorf("failed to listen: %w", err)
		}

		return grpcS.Serve(lis)
	})

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(quit)

	select {
	case <-quit:
		break
	case <-ctx.Done():
		break
	}

	cancel()

	log.Info().Msg("Shutting down server...")

	_, timeout := context.WithTimeout(context.Background(), 5*time.Second)
	defer timeout()

	grpcS.GracefulStop()

	if err := g.Wait(); err != nil {
		return xerrors.Errorf("failed to shutdown: %w", err)
	}

	log.Info().Msgf("Server exiting")

	return nil
}

func main() {
	if err := run(); err != nil {
		log.Fatal().Msgf("Failed to run server: %v", err)
	}
}
