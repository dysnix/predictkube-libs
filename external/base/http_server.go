package base

import (
	"encoding/json"
	"fmt"

	"github.com/fasthttp/router"
	"github.com/valyala/fasthttp"
	"go.uber.org/zap"

	libs "github.com/dysnix/predictkube-libs/external/configs"
)

const (
	Root = "root"

	undefinedErr = "undefined server error"
)

type HttpServer struct {
	conf    libs.ServerGetter
	logger  *fastHttpLogger
	server  *fasthttp.Server
	router  *router.Router
	handler fasthttp.RequestHandler

	createConn bool
}

func NewHttpServer(options ...libs.ServerOption) (out *HttpServer, err error) {
	defer func() {
		if out != nil && err == nil {
			out.init()
		}
	}()

	out = &HttpServer{}

	for _, op := range options {
		err := op(out)
		if err != nil {
			return nil, err
		}
	}

	return out, nil
}

func (s *HttpServer) SetConfigs(configs libs.ServerGetter) {
	s.conf = configs
}

func (s *HttpServer) SetLogger(lg *zap.SugaredLogger) {
	s.logger = &fastHttpLogger{
		SugaredLogger: *lg,
	}
}

func (s *HttpServer) SetMiddlewares(middlewares []func(fasthttp.RequestHandler) fasthttp.RequestHandler) {
	baseMiddlewares := []func(fasthttp.RequestHandler) fasthttp.RequestHandler{
		fasthttp.CompressHandler,
		s.panicMiddleware,
		s.corsMiddleware,
	}

	if len(middlewares) > 0 {
		baseMiddlewares = append(baseMiddlewares, middlewares...)
	}

	s.handler = s.router.Handler

	for i := len(baseMiddlewares) - 1; i >= 0; i-- {
		s.handler = baseMiddlewares[i](s.handler)
	}
}

func (s *HttpServer) SetRoutes(routes map[string]map[string]*libs.Route) {
	r := router.New()

	for group, routesMap := range routes {
		newGroup := r.Group(group)

		for path, route := range routesMap {
			switch route.Method {
			case fasthttp.MethodGet:
				if group != Root {
					newGroup.GET(path, route.RequestHandler)
				} else {
					r.GET(path, route.RequestHandler)
				}
			case fasthttp.MethodHead:
				if group != Root {
					newGroup.HEAD(path, route.RequestHandler)
				} else {
					r.HEAD(path, route.RequestHandler)
				}
			case fasthttp.MethodPost:
				if group != Root {
					newGroup.POST(path, route.RequestHandler)
				} else {
					r.POST(path, route.RequestHandler)
				}
			case fasthttp.MethodPut:
				if group != Root {
					newGroup.PUT(path, route.RequestHandler)
				} else {
					r.PUT(path, route.RequestHandler)
				}
			case fasthttp.MethodPatch:
				if group != Root {
					newGroup.PATCH(path, route.RequestHandler)
				} else {
					r.PATCH(path, route.RequestHandler)
				}
			case fasthttp.MethodDelete:
				if group != Root {
					newGroup.DELETE(path, route.RequestHandler)
				} else {
					r.DELETE(path, route.RequestHandler)
				}
			case fasthttp.MethodConnect:
				if group != Root {
					newGroup.CONNECT(path, route.RequestHandler)
				} else {
					r.CONNECT(path, route.RequestHandler)
				}
			case fasthttp.MethodOptions:
				if group != Root {
					newGroup.OPTIONS(path, route.RequestHandler)
				} else {
					r.OPTIONS(path, route.RequestHandler)
				}
			case fasthttp.MethodTrace:
				if group != Root {
					newGroup.TRACE(path, route.RequestHandler)
				} else {
					r.TRACE(path, route.RequestHandler)
				}
			default:
				newGroup.ANY(path, route.RequestHandler)
			}
		}
	}

	s.router = r
}

func (s *HttpServer) init() {
	s.server = &fasthttp.Server{
		Name:               s.conf.GetServerConfigs().Name,
		Concurrency:        int(s.conf.GetServerConfigs().Concurrency),
		TCPKeepalive:       s.conf.GetServerConfigs().TCPKeepalive.Enabled,
		TCPKeepalivePeriod: s.conf.GetServerConfigs().TCPKeepalive.Period,
		ReadBufferSize:     int(s.conf.GetServerConfigs().Buffer.ReadBufferSize),
		WriteBufferSize:    int(s.conf.GetServerConfigs().Buffer.WriteBufferSize),
		ReadTimeout:        s.conf.GetServerConfigs().HTTPTransport.ReadTimeout,
		WriteTimeout:       s.conf.GetServerConfigs().HTTPTransport.WriteTimeout,
		IdleTimeout:        s.conf.GetServerConfigs().HTTPTransport.MaxIdleConnDuration,
		Logger:             s.logger,
		//Handler:            fasthttp.CompressHandler(s.panicMiddleware(s.corsMiddleware(s.router.Handler))),
		Handler:      s.handler,
		LogAllErrors: true,
	}
}

func (s *HttpServer) Start() <-chan error {
	errCh := make(chan error, 1)

	go func() {
		defer close(errCh)
		if s.conf.GetServerConfigs() != nil {
			s.logger.Info("✔️ Http server started.")
			addr := fmt.Sprintf("%s:%d", s.conf.GetServerConfigs().Host, s.conf.GetServerConfigs().Port)
			if err := s.server.ListenAndServe(addr); err != nil {
				s.logger.Errorw(err.Error(), "serving fasthttp server with error")

				errCh <- err
				return
			}
		}
	}()

	return errCh
}

func (s *HttpServer) Stop() (err error) {
	defer func() {
		s.logger.Info("🛑 FataHttp server stopped.")
	}()

	return s.server.Shutdown()
}

func (s *HttpServer) corsMiddleware(next fasthttp.RequestHandler) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {

		ctx.Response.Header.Set("Access-Control-Allow-Credentials", corsAllowCredentials)
		ctx.Response.Header.Set("Access-Control-Allow-Headers", corsAllowHeaders)
		ctx.Response.Header.Set("Access-Control-Allow-Methods", corsAllowMethods)
		ctx.Response.Header.Set("Access-Control-Allow-Origin", corsAllowOrigin)

		next(ctx)
	}
}

func (s *HttpServer) JsonResp(ctx *fasthttp.RequestCtx, resp interface{}) {
	ctx.Response.Reset()
	ctx.SetStatusCode(fasthttp.StatusOK)
	ctx.SetContentTypeBytes([]byte("application/json"))

	if err := json.NewEncoder(ctx).Encode(resp); err != nil {
		s.logger.Debugf("encode json response with error: %3v", err)
		ctx.Error(fmt.Sprintf("encode response body error: %v", err), fasthttp.StatusInternalServerError)
	}
}

func (s *HttpServer) ErrorPrint(ctx *fasthttp.RequestCtx, err error, statusCode int) {
	ctx.Response.Reset()
	ctx.SetStatusCode(statusCode)
	ctx.SetContentTypeBytes([]byte("application/json"))

	resp := make(map[string]interface{})
	if err != nil {
		resp["error"] = err.Error()
	} else {
		resp["error"] = undefinedErr
	}

	if err1 := json.NewEncoder(ctx).Encode(resp); err1 != nil {
		s.logger.Debugf("encode failed response with error: %3v", err1)

		ctx.Error(fmt.Sprintf("encode server error: %v", err1), fasthttp.StatusInternalServerError)
	}
}

func (s *HttpServer) panicMiddleware(next fasthttp.RequestHandler) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		defer func() {
			if r := recover(); r != nil {
				if err, ok := r.(error); ok {
					s.logger.Errorw("panic middleware detect", err)
					errorPrint(ctx, err, fasthttp.StatusInternalServerError)
				}
			}
		}()

		next(ctx)
	}
}
