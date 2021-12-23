package base

import (
	"encoding/json"
	"fmt"
	"net/http/pprof"
	rtp "runtime/pprof"
	"strings"

	"github.com/fasthttp/router"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttp/fasthttpadaptor"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/status"

	libs "github.com/dysnix/predictkube-libs/external/configs"
	"github.com/dysnix/predictkube-libs/external/grpc/client"
	libsSrv "github.com/dysnix/predictkube-libs/external/grpc/server"
	health "google.golang.org/grpc/health/grpc_health_v1"
)

type fastHttpLogger struct {
	zap.SugaredLogger
}

func (l *fastHttpLogger) Printf(format string, args ...interface{}) {
	l.SugaredLogger.Debugf(format, args...)
}

var (
	_ libsSrv.Server = (*FastHttpServer)(nil)
)

type FastHttpServer struct {
	conf   libs.SingleGetter
	logger *fastHttpLogger
	server *fasthttp.Server

	grpcClient health.HealthClient
	grpcConn   *grpc.ClientConn
	createConn bool
}

func NewFastHttpServer(options ...libs.Option) (out *FastHttpServer, err error) {
	defer func() {
		if out != nil && err == nil {
			out.init()
		}
	}()

	out = &FastHttpServer{}

	for _, op := range options {
		err := op(out)
		if err != nil {
			return nil, err
		}
	}

	if out.conf.GetGrpc() != nil {
		if out.grpcConn == nil {
			// grpc client for ping with some service
			var grpcClientOpt []grpc.DialOption
			if out.conf.GetClient() != nil {
				grpcClientOpt, err = client.SetGrpcClientOptions(out.conf.GetGrpc(), out.conf.GetBase(), client.InjectClientMetadataInterceptor(*out.conf.GetClient()))
			} else {
				grpcClientOpt, err = client.SetGrpcClientOptions(out.conf.GetGrpc(), out.conf.GetBase())
			}

			if err != nil {
				return nil, err
			}

			out.grpcConn, err = grpc.Dial(fmt.Sprintf("%s:%d", out.conf.GetGrpc().Conn.Host, out.conf.GetGrpc().Conn.Port), grpcClientOpt...)
			if err != nil {
				return nil, err
			}
		}

		out.grpcClient = health.NewHealthClient(out.grpcConn)
		out.createConn = true
	}

	return out, nil
}

func (s *FastHttpServer) SetConfigs(configs libs.SingleGetter) {
	s.conf = configs
}

func (s *FastHttpServer) SetLogger(lg *zap.SugaredLogger) {
	s.logger = &fastHttpLogger{
		SugaredLogger: *lg,
	}
}

func (s *FastHttpServer) SetGrpcClientConn(conn *grpc.ClientConn) {
	s.grpcConn = conn
}

func (s *FastHttpServer) routing() *router.Router {
	r := router.New()

	if s.conf.GetGrpc() != nil {
		r.GET("/healthz", s.liveness)
		r.GET("/readyz", s.readiness)
	}

	if s.conf.GetBase().Monitoring.Enabled {
		// metrics
		r.ANY("/metrics", fasthttpadaptor.NewFastHTTPHandler(promhttp.Handler()))
	}

	if s.conf.GetBase().Profiling.Enabled {
		// profiling
		grPprof := r.Group("/debug/pprof")
		grPprof.ANY("/cmdline", fasthttpadaptor.NewFastHTTPHandlerFunc(pprof.Cmdline))
		grPprof.ANY("/profile", fasthttpadaptor.NewFastHTTPHandlerFunc(pprof.Profile))
		grPprof.ANY("/symbol", fasthttpadaptor.NewFastHTTPHandlerFunc(pprof.Symbol))
		grPprof.ANY("/trace", fasthttpadaptor.NewFastHTTPHandlerFunc(pprof.Trace))
		grPprof.ANY("/{path:*}", s.indexPprofRoute)
	}

	return r
}

func (s *FastHttpServer) init() {
	s.server = &fasthttp.Server{
		Name:               s.conf.GetBase().Single.Name,
		Concurrency:        int(s.conf.GetBase().Single.Concurrency),
		TCPKeepalive:       s.conf.GetBase().Single.TCPKeepalive.Enabled,
		TCPKeepalivePeriod: s.conf.GetBase().Single.TCPKeepalive.Period,
		ReadBufferSize:     int(s.conf.GetBase().Single.Buffer.ReadBufferSize),
		WriteBufferSize:    int(s.conf.GetBase().Single.Buffer.WriteBufferSize),
		ReadTimeout:        s.conf.GetBase().Single.HTTPTransport.ReadTimeout,
		WriteTimeout:       s.conf.GetBase().Single.HTTPTransport.WriteTimeout,
		IdleTimeout:        s.conf.GetBase().Single.HTTPTransport.MaxIdleConnDuration,
		Logger:             s.logger,
		Handler:            fasthttp.CompressHandler(s.PanicMiddleware(s.CorsMiddleware(s.routing().Handler))),
	}

	if s.conf.GetBase().IsDebugMode {
		s.server.LogAllErrors = true
	}
}

func (s *FastHttpServer) Start() <-chan error {
	errCh := make(chan error, 1)

	go func() {
		defer close(errCh)
		if s.conf.GetBase().Single != nil && s.conf.GetBase().Single.Enabled {
			s.logger.Info("✔️ FastHttp server started.")
			if err := s.server.ListenAndServe(fmt.Sprintf("%s:%d", s.conf.GetBase().Single.Host, s.conf.GetBase().Single.Port)); err != nil {
				if s.conf.GetBase().IsDebugMode {
					s.logger.Errorw(err.Error(), "serving fasthttp server with error")
				}

				errCh <- err
				return
			}
		}
	}()

	return errCh
}

func (s *FastHttpServer) Stop() (err error) {
	defer func() {
		if s.createConn {
			err1Tmp := err

			if err = s.grpcConn.Close(); err != nil && err1Tmp != nil {
				err = errors.WithMessage(err, err1Tmp.Error())
			}
		}

		if s.conf.GetBase().IsDebugMode {
			s.logger.Info("🛑 FataHttp server stopped.")
		}
	}()

	return s.server.Shutdown()
}

func (s *FastHttpServer) liveness(ctx *fasthttp.RequestCtx) {
	if s.grpcConn.GetState() == connectivity.Shutdown {
		errorPrint(ctx, errors.New("GRPC server connection is closed"), fasthttp.StatusServiceUnavailable)
		return
	}
}

func (s *FastHttpServer) readiness(ctx *fasthttp.RequestCtx) {
	_, err := health.NewHealthClient(s.grpcConn).Check(ctx, &health.HealthCheckRequest{
		Service: "_",
	})

	if err != nil {
		if stat, ok := status.FromError(err); ok && stat.Code() == codes.Unimplemented {
			errorPrint(ctx, errors.New("the GRPC server doesn't implement the grpc health protocol"), fasthttp.StatusNotImplemented)
			return
		}

		errorPrint(ctx, fmt.Errorf("GRPC server rpc failed %s", err), fasthttp.StatusInternalServerError)
		return
	}
}

const (
	corsAllowHeaders     = "authorization"
	corsAllowMethods     = "HEAD,GET,POST,PUT,DELETE,OPTIONS"
	corsAllowOrigin      = "*"
	corsAllowCredentials = "true"
)

func (s *FastHttpServer) CorsMiddleware(next fasthttp.RequestHandler) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {

		ctx.Response.Header.Set("Access-Control-Allow-Credentials", corsAllowCredentials)
		ctx.Response.Header.Set("Access-Control-Allow-Headers", corsAllowHeaders)
		ctx.Response.Header.Set("Access-Control-Allow-Methods", corsAllowMethods)
		ctx.Response.Header.Set("Access-Control-Allow-Origin", corsAllowOrigin)

		next(ctx)
	}
}

func errorPrint(ctx *fasthttp.RequestCtx, err error, statusCode int) {
	ctx.Response.Reset()
	ctx.SetStatusCode(statusCode)
	ctx.SetContentTypeBytes([]byte("application/json"))
	if err1 := json.NewEncoder(ctx).Encode(map[string]string{
		"error": err.Error(),
	}); err1 != nil {
		ctx.Error(err.Error(), fasthttp.StatusInternalServerError)
	}
}

func (s *FastHttpServer) PanicMiddleware(next fasthttp.RequestHandler) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		defer func() {
			if r := recover(); r != nil {
				if err, ok := r.(error); ok {
					s.logger.Errorw("panic middleware detect", err)
					errorPrint(ctx, err, fasthttp.StatusInternalServerError)
				}
			}

			ctx.Done()
		}()

		next(ctx)
	}
}

func (s *FastHttpServer) indexPprofRoute(ctx *fasthttp.RequestCtx) {
	ctx.Response.Header.Set("Content-Type", "text/html")

	for _, v := range rtp.Profiles() {
		ppName := v.Name()
		if strings.HasPrefix(string(ctx.Path()), "/debug/pprof/"+ppName) {
			namedHandler := fasthttpadaptor.NewFastHTTPHandlerFunc(pprof.Handler(ppName).ServeHTTP)
			namedHandler(ctx)
			return
		}
	}

	index := fasthttpadaptor.NewFastHTTPHandlerFunc(pprof.Index)
	index(ctx)
}
