package configs

import (
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
)

type SignalStopper interface {
	Stop()
}

type SignalStopperWithErr interface {
	Stop() error
}

type SignalCloser interface {
	Close()
}

type SignalCloserWithErr interface {
	Close() error
}

type SingleUseGetter interface {
	SingleEnabled() bool
}

type SingleGetter interface {
	GetBase() *Base
	GetGrpc() *GRPC
	GetClient() *Client
}

type ConfigSetter interface {
	SetConfigs(configs SingleGetter)
}

type LoggerSetter interface {
	SetLogger(logger *zap.SugaredLogger)
}

type GrpcClientConnSetter interface {
	SetGrpcClientConn(conn *grpc.ClientConn)
}

type MainSetters interface {
	ConfigSetter
	LoggerSetter
	GrpcClientConnSetter
}

type Option func(MainSetters) error

func SetConfigs(conf SingleGetter) Option {
	return func(r MainSetters) error {
		if conf != nil {
			r.SetConfigs(conf)
		}
		return nil
	}
}

func SetLogger(logger *zap.SugaredLogger) Option {
	return func(r MainSetters) error {
		if logger != nil {
			r.SetLogger(logger)
		}
		return nil
	}
}

func SetConn(conn *grpc.ClientConn) Option {
	return func(r MainSetters) error {
		if conn != nil && conn.GetState() == connectivity.Ready {
			r.SetGrpcClientConn(conn)
		}
		return nil
	}
}
