package configs

import (
	"encoding/json"
	"fmt"
	"time"

	str2duration "github.com/xhit/go-str2duration/v2"

	"github.com/dysnix/predictkube-libs/external/enums"
	tc "github.com/dysnix/predictkube-libs/external/types_convertation"
)

type Base struct {
	IsDebugMode bool `yaml:"debugMode" json:"debug_mode"`
	Profiling   Profiling
	Monitoring  Monitoring
	Single      *Single `yaml:"single,omitempty" json:"single,omitempty"`
}

type Client struct {
	ClusterID string `yaml:"clusterId" json:"cluster_id" validate:"uuid"`
	Name      string
	Token     string `validate:"jwt"`
}

type Single struct {
	Enabled       bool
	Host          string `validate:"host_if_enabled"`
	Port          uint16 `validate:"port_if_enabled"`
	Name          string
	Concurrency   uint          `validate:"required_if=Enabled true,gt=1,lte=1000000"`
	Buffer        *Buffer       `yaml:"buffer,omitempty" json:"buffer,omitempty" validate:"required_if=Enabled true"`
	TCPKeepalive  *TCPKeepalive `yaml:"tcpKeepalive,omitempty" json:"tcpKeepalive,omitempty" validate:"required_if=Enabled true"`
	HTTPTransport HTTPTransport
}

type Buffer struct {
	ReadBufferSize  uint `yaml:"readBufferSize" json:"read_buffer_size" validate:"gte=4096"`
	WriteBufferSize uint `yaml:"writeBufferSize" json:"write_buffer_size" validate:"gte=4096"`
}

func (b *Buffer) MarshalJSON() ([]byte, error) {
	type alias struct {
		ReadBufferSize  string `yaml:"readBufferSize" json:"read_buffer_size"`
		WriteBufferSize string `yaml:"writeBufferSize" json:"write_buffer_size"`
	}

	if b == nil {
		*b = Buffer{}
	}

	return json.Marshal(alias{
		ReadBufferSize:  tc.BytesSize(float64(b.ReadBufferSize)),
		WriteBufferSize: tc.BytesSize(float64(b.WriteBufferSize)),
	})
}

func (b *Buffer) UnmarshalJSON(data []byte) (err error) {
	type alias struct {
		ReadBufferSize  string `yaml:"readBufferSize" json:"read_buffer_size"`
		WriteBufferSize string `yaml:"writeBufferSize" json:"write_buffer_size"`
	}
	var tmp alias
	if err = json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	if b == nil {
		*b = Buffer{}
	}

	var tmpB int64
	if tmpB, err = tc.RAMInBytes(tmp.ReadBufferSize); err != nil {
		return err
	}

	b.ReadBufferSize = uint(tmpB)

	if tmpB, err = tc.RAMInBytes(tmp.WriteBufferSize); err != nil {
		return err
	}

	b.WriteBufferSize = uint(tmpB)

	return nil
}

func (b *Buffer) MarshalYAML() (interface{}, error) {
	type alias struct {
		ReadBufferSize  string `yaml:"readBufferSize" json:"read_buffer_size"`
		WriteBufferSize string `yaml:"writeBufferSize" json:"write_buffer_size"`
	}

	if b == nil {
		*b = Buffer{}
	}

	return alias{
		ReadBufferSize:  tc.BytesSize(float64(b.ReadBufferSize)),
		WriteBufferSize: tc.BytesSize(float64(b.WriteBufferSize)),
	}, nil
}

func (b *Buffer) UnmarshalYAML(unmarshal func(interface{}) error) error {
	type alias struct {
		ReadBufferSize  string `yaml:"readBufferSize" json:"read_buffer_size"`
		WriteBufferSize string `yaml:"writeBufferSize" json:"write_buffer_size"`
	}

	var tmp alias
	err := unmarshal(&tmp)
	if err != nil {
		return err
	}

	if b == nil {
		*b = Buffer{}
	}

	var tmpB int64
	if tmpB, err = tc.RAMInBytes(tmp.ReadBufferSize); err != nil {
		return err
	}

	b.ReadBufferSize = uint(tmpB)

	if tmpB, err = tc.RAMInBytes(tmp.WriteBufferSize); err != nil {
		return err
	}

	b.WriteBufferSize = uint(tmpB)

	return nil
}

type TCPKeepalive struct {
	Enabled bool
	Period  time.Duration `validate:"required,gt=0"`
}

func (k *TCPKeepalive) MarshalJSON() ([]byte, error) {
	type alias struct {
		Enabled bool
		Period  string
	}

	if k == nil {
		*k = TCPKeepalive{}
	}

	return json.Marshal(alias{
		Enabled: k.Enabled,
		Period:  HumanDuration(k.Period),
	})
}

func (k *TCPKeepalive) UnmarshalJSON(data []byte) (err error) {
	type alias struct {
		Enabled bool
		Period  string
	}

	var tmp alias
	if err = json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	if k == nil {
		*k = TCPKeepalive{}
	}

	k.Enabled = tmp.Enabled

	k.Period, err = str2duration.ParseDuration(tmp.Period)
	if err != nil {
		return err
	}

	return nil
}

func (k *TCPKeepalive) MarshalYAML() (interface{}, error) {
	type alias struct {
		Enabled bool
		Period  string
	}

	if k == nil {
		*k = TCPKeepalive{}
	}

	return alias{
		Enabled: k.Enabled,
		Period:  HumanDuration(k.Period),
	}, nil
}

func (k *TCPKeepalive) UnmarshalYAML(unmarshal func(interface{}) error) error {
	type alias struct {
		Enabled bool
		Period  string
	}

	var tmp alias
	err := unmarshal(&tmp)
	if err != nil {
		return err
	}

	if k == nil {
		*k = TCPKeepalive{}
	}

	k.Enabled = tmp.Enabled

	k.Period, err = str2duration.ParseDuration(tmp.Period)
	if err != nil {
		return err
	}

	return nil
}

type Monitoring struct {
	Enabled bool
	Host    string `yaml:"host,omitempty" json:"host,omitempty" validate:"host_if_enabled"`
	Port    uint16 `yaml:"port,omitempty" json:"port,omitempty" validate:"port_if_enabled"`
}

type Profiling struct {
	Enabled bool
	Host    string `yaml:"host,omitempty" json:"host,omitempty" validate:"host_if_enabled"`
	Port    uint16 `yaml:"port,omitempty" json:"port,omitempty" validate:"port_if_enabled"`
}

type CronStr string

type Informer struct {
	Resource string        `yaml:"resource" json:"resource" validate:"required"`
	Interval time.Duration `yaml:"interval" json:"interval" validate:"required,gt=0"`
}

type K8sCloudWatcher struct {
	CtxPath   string     `yaml:"kubeConfigPath" json:"kube_config_path" validate:"required,file"`
	Informers []Informer `yaml:"informers" json:"informers" validate:"required,gt=0"`
}

type GRPC struct {
	Enabled       bool
	UseReflection bool        `yaml:"useReflection" json:"use_reflection"`
	Compression   Compression `yaml:"compression" json:"compression"`
	Conn          *Connection `yaml:"connection" json:"connection" validate:"required"`
	Keepalive     *Keepalive  `yaml:"keepalive" json:"keepalive"`
}

type Compression struct {
	Enabled bool                  `yaml:"enabled" json:"enabled"`
	Type    enums.CompressionType `yaml:"type" json:"type"`
}

type Connection struct {
	Host            string        `yaml:"host" json:"host" validate:"grpc_host"`
	Port            uint16        `yaml:"port" json:"port" validate:"required,gt=0"`
	ReadBufferSize  uint          `yaml:"readBufferSize" json:"read_buffer_size" validate:"required,gte=4096"`
	WriteBufferSize uint          `yaml:"writeBufferSize" json:"write_buffer_size" validate:"required,gte=4096"`
	MaxMessageSize  uint          `yaml:"maxMessageSize" json:"max_message_size" validate:"required,gte=2048"`
	Insecure        bool          `yaml:"insecure" json:"insecure"`
	Timeout         time.Duration `yaml:"timeout" json:"timeout" validate:"gte=0"`
}

func (c *Connection) MarshalJSON() ([]byte, error) {
	type alias struct {
		Host            string  `yaml:"host" json:"host"`
		Port            uint16  `yaml:"port" json:"port"`
		ReadBufferSize  string  `yaml:"readBufferSize" json:"read_buffer_size"`
		WriteBufferSize string  `yaml:"writeBufferSize" json:"write_buffer_size"`
		MaxMessageSize  string  `yaml:"maxMessageSize" json:"max_message_size"`
		Insecure        bool    `yaml:"insecure" json:"insecure"`
		Timeout         *string `yaml:"timeout,omitempty" json:"timeout,omitempty"`
	}

	if c == nil {
		*c = Connection{}
	}

	return json.Marshal(alias{
		Host:            c.Host,
		Port:            c.Port,
		ReadBufferSize:  tc.BytesSize(float64(c.ReadBufferSize)),
		WriteBufferSize: tc.BytesSize(float64(c.WriteBufferSize)),
		MaxMessageSize:  tc.BytesSize(float64(c.MaxMessageSize)),
		Insecure:        c.Insecure,
		Timeout:         tc.String(ConvertDurationToStr(c.Timeout)),
	})
}

func (c *Connection) UnmarshalJSON(data []byte) (err error) {
	type alias struct {
		Host            string  `yaml:"host" json:"host"`
		Port            uint16  `yaml:"port" json:"port"`
		ReadBufferSize  string  `yaml:"readBufferSize" json:"read_buffer_size"`
		WriteBufferSize string  `yaml:"writeBufferSize" json:"write_buffer_size"`
		MaxMessageSize  string  `yaml:"maxMessageSize" json:"max_message_size"`
		Insecure        bool    `yaml:"insecure" json:"insecure"`
		Timeout         *string `yaml:"timeout,omitempty" json:"timeout,omitempty"`
	}
	var tmp alias
	if err = json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	if c == nil {
		*c = Connection{}
	}

	if tmp.Timeout != nil {
		c.Timeout, err = str2duration.ParseDuration(*tmp.Timeout)
		if err != nil {
			return err
		}
	}

	c.Host = tmp.Host
	c.Port = tmp.Port

	var tmpB int64
	if tmpB, err = tc.RAMInBytes(tmp.ReadBufferSize); err != nil {
		return err
	}

	c.ReadBufferSize = uint(tmpB)

	if tmpB, err = tc.RAMInBytes(tmp.WriteBufferSize); err != nil {
		return err
	}

	c.WriteBufferSize = uint(tmpB)

	if tmpB, err = tc.RAMInBytes(tmp.MaxMessageSize); err != nil {
		return err
	}

	c.MaxMessageSize = uint(tmpB)
	c.Insecure = tmp.Insecure

	return nil
}

func (c *Connection) MarshalYAML() (interface{}, error) {
	type alias struct {
		Host            string  `yaml:"host" json:"host"`
		Port            uint16  `yaml:"port" json:"port"`
		ReadBufferSize  string  `yaml:"readBufferSize" json:"read_buffer_size"`
		WriteBufferSize string  `yaml:"writeBufferSize" json:"write_buffer_size"`
		MaxMessageSize  string  `yaml:"maxMessageSize" json:"max_message_size"`
		Insecure        bool    `yaml:"insecure" json:"insecure"`
		Timeout         *string `yaml:"timeout,omitempty" json:"timeout,omitempty"`
	}

	if c == nil {
		*c = Connection{}
	}

	return alias{
		Host:            c.Host,
		Port:            c.Port,
		ReadBufferSize:  tc.BytesSize(float64(c.ReadBufferSize)),
		WriteBufferSize: tc.BytesSize(float64(c.WriteBufferSize)),
		MaxMessageSize:  tc.BytesSize(float64(c.MaxMessageSize)),
		Insecure:        c.Insecure,
		Timeout:         tc.String(ConvertDurationToStr(c.Timeout)),
	}, nil
}

func (c *Connection) UnmarshalYAML(unmarshal func(interface{}) error) error {
	type alias struct {
		Host            string  `yaml:"host" json:"host"`
		Port            uint16  `yaml:"port" json:"port"`
		ReadBufferSize  string  `yaml:"readBufferSize" json:"read_buffer_size"`
		WriteBufferSize string  `yaml:"writeBufferSize" json:"write_buffer_size"`
		MaxMessageSize  string  `yaml:"maxMessageSize" json:"max_message_size"`
		Insecure        bool    `yaml:"insecure" json:"insecure"`
		Timeout         *string `yaml:"timeout,omitempty" json:"timeout,omitempty"`
	}
	var tmp alias
	err := unmarshal(&tmp)
	if err != nil {
		return err
	}

	if c == nil {
		*c = Connection{}
	}

	if tmp.Timeout != nil {
		c.Timeout, err = str2duration.ParseDuration(*tmp.Timeout)
		if err != nil {
			return err
		}
	}

	c.Host = tmp.Host
	c.Port = tmp.Port

	var tmpB int64
	if tmpB, err = tc.RAMInBytes(tmp.ReadBufferSize); err != nil {
		return err
	}

	c.ReadBufferSize = uint(tmpB)

	if tmpB, err = tc.RAMInBytes(tmp.WriteBufferSize); err != nil {
		return err
	}

	c.WriteBufferSize = uint(tmpB)

	if tmpB, err = tc.RAMInBytes(tmp.MaxMessageSize); err != nil {
		return err
	}

	c.MaxMessageSize = uint(tmpB)
	c.Insecure = tmp.Insecure

	return nil
}

type Keepalive struct {
	Time              time.Duration      `yaml:"time" json:"time" validate:"required,gt=0"`
	Timeout           time.Duration      `yaml:"timeout" json:"timeout" validate:"required,gt=0"`
	EnforcementPolicy *EnforcementPolicy `yaml:"enforcementPolicy" json:"enforcement_policy"`
}

func (ka *Keepalive) MarshalJSON() ([]byte, error) {
	type alias struct {
		Time              string             `yaml:"time" json:"time"`
		Timeout           string             `yaml:"timeout" json:"timeout"`
		EnforcementPolicy *EnforcementPolicy `yaml:"enforcementPolicy" json:"enforcement_policy"`
	}

	if ka == nil {
		*ka = Keepalive{}
	}

	return json.Marshal(alias{
		Time:              ConvertDurationToStr(ka.Time),
		Timeout:           ConvertDurationToStr(ka.Timeout),
		EnforcementPolicy: ka.EnforcementPolicy,
	})
}

func (ka *Keepalive) UnmarshalJSON(data []byte) (err error) {
	type alias struct {
		Time              string             `yaml:"time" json:"time"`
		Timeout           string             `yaml:"timeout" json:"timeout"`
		EnforcementPolicy *EnforcementPolicy `yaml:"enforcementPolicy" json:"enforcement_policy"`
	}
	var tmp alias
	if err = json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	if ka == nil {
		*ka = Keepalive{}
	}

	ka.Time, err = str2duration.ParseDuration(tmp.Time)
	if err != nil {
		return err
	}

	ka.Timeout, err = str2duration.ParseDuration(tmp.Timeout)
	if err != nil {
		return err
	}

	ka.EnforcementPolicy = tmp.EnforcementPolicy

	return nil
}

func (ka *Keepalive) MarshalYAML() (interface{}, error) {
	type alias struct {
		Time              string             `yaml:"time" json:"time"`
		Timeout           string             `yaml:"timeout" json:"timeout"`
		EnforcementPolicy *EnforcementPolicy `yaml:"enforcementPolicy" json:"enforcement_policy"`
	}

	if ka == nil {
		*ka = Keepalive{}
	}

	return alias{
		Time:              ConvertDurationToStr(ka.Time),
		Timeout:           ConvertDurationToStr(ka.Timeout),
		EnforcementPolicy: ka.EnforcementPolicy,
	}, nil
}

func (ka *Keepalive) UnmarshalYAML(unmarshal func(interface{}) error) error {
	type alias struct {
		Time              string             `yaml:"time" json:"time"`
		Timeout           string             `yaml:"timeout" json:"timeout"`
		EnforcementPolicy *EnforcementPolicy `yaml:"enforcementPolicy" json:"enforcement_policy"`
	}
	var tmp alias
	err := unmarshal(&tmp)
	if err != nil {
		return err
	}

	if ka == nil {
		*ka = Keepalive{}
	}

	ka.Time, err = str2duration.ParseDuration(tmp.Time)
	if err != nil {
		return err
	}

	ka.Timeout, err = str2duration.ParseDuration(tmp.Timeout)
	if err != nil {
		return err
	}

	ka.EnforcementPolicy = tmp.EnforcementPolicy

	return nil
}

type EnforcementPolicy struct {
	MinTime             time.Duration `yaml:"minTime" json:"min_time" validate:"required,gt=0"`
	PermitWithoutStream bool          `yaml:"permitWithoutStream" json:"permit_without_stream"`
}

func (ep *EnforcementPolicy) MarshalJSON() ([]byte, error) {
	type alias struct {
		MinTime             string `yaml:"minTime" json:"min_time"`
		PermitWithoutStream bool   `yaml:"permitWithoutStream" json:"permit_without_stream"`
	}

	if ep == nil {
		*ep = EnforcementPolicy{}
	}

	return json.Marshal(alias{
		MinTime:             ConvertDurationToStr(ep.MinTime),
		PermitWithoutStream: ep.PermitWithoutStream,
	})
}

func (ep *EnforcementPolicy) UnmarshalJSON(data []byte) (err error) {
	type alias struct {
		MinTime             string `yaml:"minTime" json:"min_time"`
		PermitWithoutStream bool   `yaml:"permitWithoutStream" json:"permit_without_stream"`
	}
	var tmp alias
	if err = json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	if ep == nil {
		*ep = EnforcementPolicy{}
	}

	ep.PermitWithoutStream = tmp.PermitWithoutStream
	ep.MinTime, err = str2duration.ParseDuration(tmp.MinTime)

	return err
}

func (ep *EnforcementPolicy) MarshalYAML() (interface{}, error) {
	type alias struct {
		MinTime             string `yaml:"minTime" json:"min_time"`
		PermitWithoutStream bool   `yaml:"permitWithoutStream" json:"permit_without_stream"`
	}

	if ep == nil {
		*ep = EnforcementPolicy{}
	}

	return alias{
		MinTime:             ConvertDurationToStr(ep.MinTime),
		PermitWithoutStream: ep.PermitWithoutStream,
	}, nil
}

func (ep *EnforcementPolicy) UnmarshalYAML(unmarshal func(interface{}) error) error {
	type alias struct {
		MinTime             string `yaml:"minTime" json:"min_time"`
		PermitWithoutStream bool   `yaml:"permitWithoutStream" json:"permit_without_stream"`
	}
	var tmp alias
	err := unmarshal(&tmp)
	if err != nil {
		return err
	}

	if ep == nil {
		*ep = EnforcementPolicy{}
	}

	ep.PermitWithoutStream = tmp.PermitWithoutStream
	ep.MinTime, err = str2duration.ParseDuration(tmp.MinTime)

	return err
}

type HTTPTransport struct {
	MaxIdleConnDuration time.Duration `yaml:"maxIdleConnDuration" json:"max_idle_conn_duration" validate:"required,gt=0"`
	ReadTimeout         time.Duration `yaml:"readTimeout" json:"read_timeout" validate:"required,gt=0"`
	WriteTimeout        time.Duration `yaml:"writeTimeout" json:"write_timeout" validate:"required,gt=0"`
}

func (t *HTTPTransport) GetTransportConfigs() *HTTPTransport {
	return t
}

func (t *HTTPTransport) MarshalJSON() ([]byte, error) {
	type alias struct {
		MaxIdleConnDuration string `yaml:"maxIdleConnDuration" json:"max_idle_conn_duration"`
		ReadTimeout         string `yaml:"readTimeout" json:"read_timeout"`
		WriteTimeout        string `yaml:"writeTimeout" json:"write_timeout"`
	}

	if t == nil {
		*t = HTTPTransport{}
	}

	return json.Marshal(alias{
		MaxIdleConnDuration: HumanDuration(t.MaxIdleConnDuration),
		ReadTimeout:         HumanDuration(t.ReadTimeout),
		WriteTimeout:        HumanDuration(t.WriteTimeout),
	})
}

func (t *HTTPTransport) UnmarshalJSON(data []byte) (err error) {
	type alias struct {
		MaxIdleConnDuration string `yaml:"maxIdleConnDuration" json:"max_idle_conn_duration"`
		ReadTimeout         string `yaml:"readTimeout" json:"read_timeout"`
		WriteTimeout        string `yaml:"writeTimeout" json:"write_timeout"`
	}
	var tmp alias
	if err = json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	if t == nil {
		*t = HTTPTransport{}
	}

	t.MaxIdleConnDuration, err = str2duration.ParseDuration(tmp.MaxIdleConnDuration)
	if err != nil {
		return err
	}

	t.ReadTimeout, err = str2duration.ParseDuration(tmp.ReadTimeout)
	if err != nil {
		return err
	}

	t.WriteTimeout, err = str2duration.ParseDuration(tmp.WriteTimeout)
	if err != nil {
		return err
	}

	return nil
}

func (t *HTTPTransport) MarshalYAML() (interface{}, error) {
	type alias struct {
		MaxIdleConnDuration string `yaml:"maxIdleConnDuration" json:"max_idle_conn_duration"`
		ReadTimeout         string `yaml:"readTimeout" json:"read_timeout"`
		WriteTimeout        string `yaml:"writeTimeout" json:"write_timeout"`
	}

	if t == nil {
		*t = HTTPTransport{}
	}

	return alias{
		MaxIdleConnDuration: HumanDuration(t.MaxIdleConnDuration),
		ReadTimeout:         HumanDuration(t.ReadTimeout),
		WriteTimeout:        HumanDuration(t.WriteTimeout),
	}, nil
}

func (t *HTTPTransport) UnmarshalYAML(unmarshal func(interface{}) error) error {
	type alias struct {
		MaxIdleConnDuration string `yaml:"maxIdleConnDuration" json:"max_idle_conn_duration"`
		ReadTimeout         string `yaml:"readTimeout" json:"read_timeout"`
		WriteTimeout        string `yaml:"writeTimeout" json:"write_timeout"`
	}
	var tmp alias
	err := unmarshal(&tmp)
	if err != nil {
		return err
	}

	if t == nil {
		*t = HTTPTransport{}
	}

	t.MaxIdleConnDuration, err = str2duration.ParseDuration(tmp.MaxIdleConnDuration)
	if err != nil {
		return err
	}

	t.ReadTimeout, err = str2duration.ParseDuration(tmp.ReadTimeout)
	if err != nil {
		return err
	}

	t.WriteTimeout, err = str2duration.ParseDuration(tmp.WriteTimeout)
	if err != nil {
		return err
	}

	return nil
}

type Postgres struct {
	Username string        `validate:"ascii"`
	Password string        `validate:"ascii"`
	Database string        `validate:"required,ascii"`
	Host     string        `validate:"grpc_host"`
	Port     uint16        `validate:"required,gt=0"`
	Schema   string        `validate:"alphanum"`
	SSLMode  enums.SSLMode `validate:"required"`
	Pool     *Pool         `yaml:"pool" json:"pool" validate:"required"`
}

func (pg *Postgres) Dsn() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", pg.Host, pg.Port, pg.Username, pg.Password, pg.Database)
}

type Pool struct {
	MaxIdleConns    int           `yaml:"maxIdleConns" json:"max_idle_conns" validate:"required,gt=0"`
	MaxOpenConns    int           `yaml:"maxOpenConns" json:"max_open_conns" validate:"required,gt=0"`
	ConnMaxLifetime time.Duration `yaml:"connMaxLifetime" json:"conn_max_lifetime" validate:"required,gt=0"`
}

func (p *Pool) MarshalYAML() (interface{}, error) {
	type alias struct {
		MaxIdleConns    int    `yaml:"maxIdleConns" json:"max_idle_conns"`
		MaxOpenConns    int    `yaml:"maxOpenConns" json:"max_open_conns"`
		ConnMaxLifetime string `yaml:"connMaxLifetime" json:"conn_max_lifetime"`
	}

	if p == nil {
		*p = Pool{}
	}

	return alias{
		MaxIdleConns:    p.MaxIdleConns,
		MaxOpenConns:    p.MaxOpenConns,
		ConnMaxLifetime: HumanDuration(p.ConnMaxLifetime),
	}, nil
}

func (p *Pool) UnmarshalYAML(unmarshal func(interface{}) error) error {
	type alias struct {
		MaxIdleConns    int    `yaml:"maxIdleConns" json:"max_idle_conns"`
		MaxOpenConns    int    `yaml:"maxOpenConns" json:"max_open_conns"`
		ConnMaxLifetime string `yaml:"connMaxLifetime" json:"conn_max_lifetime"`
	}
	var tmp alias
	err := unmarshal(&tmp)
	if err != nil {
		return err
	}

	if p == nil {
		*p = Pool{}
	}

	p.MaxIdleConns = tmp.MaxIdleConns
	p.MaxOpenConns = tmp.MaxOpenConns

	p.ConnMaxLifetime, err = str2duration.ParseDuration(tmp.ConnMaxLifetime)
	if err != nil {
		return err
	}

	return nil
}

func (p *Pool) MarshalJSON() ([]byte, error) {
	type alias struct {
		MaxIdleConns    int    `yaml:"maxIdleConns" json:"max_idle_conns"`
		MaxOpenConns    int    `yaml:"maxOpenConns" json:"max_open_conns"`
		ConnMaxLifetime string `yaml:"connMaxLifetime" json:"conn_max_lifetime"`
	}

	if p == nil {
		*p = Pool{}
	}

	return json.Marshal(alias{
		MaxIdleConns:    p.MaxIdleConns,
		MaxOpenConns:    p.MaxOpenConns,
		ConnMaxLifetime: HumanDuration(p.ConnMaxLifetime),
	})
}

func (p *Pool) UnmarshalJSON(data []byte) (err error) {
	type alias struct {
		MaxIdleConns    int    `yaml:"maxIdleConns" json:"max_idle_conns"`
		MaxOpenConns    int    `yaml:"maxOpenConns" json:"max_open_conns"`
		ConnMaxLifetime string `yaml:"connMaxLifetime" json:"conn_max_lifetime"`
	}

	var tmp alias
	if err = json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	if p == nil {
		*p = Pool{}
	}

	p.MaxIdleConns = tmp.MaxIdleConns
	p.MaxOpenConns = tmp.MaxOpenConns

	p.ConnMaxLifetime, err = str2duration.ParseDuration(tmp.ConnMaxLifetime)
	if err != nil {
		return err
	}

	return nil
}
