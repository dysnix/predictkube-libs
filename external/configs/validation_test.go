package configs

import (
	"context"
	"testing"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
)

type tmpStruct struct {
	SomeTypeFirst  bool          //`validate:"required"`
	SomeTypeSecond time.Duration `validate:"require_if_not_nil_or_empty=SomeTypeFirst"`
	SomeCron       CronStr       `validate:"cron"`
	Profiling      Profiling
	Single         *Single
}

func (t tmpStruct) SingleEnabled() bool {
	return t.Single != nil && t.Single.Enabled
}

func TestValidateRequiredIfNotEmpty(t *testing.T) {

	var cases = []struct {
		name string
		opts []*tmpStruct
		want string
	}{
		{
			name: "1. valid validation",
			opts: []*tmpStruct{
				{
					SomeTypeFirst:  false,
					SomeTypeSecond: 0,
					SomeCron:       "* * * * *",
					Profiling: Profiling{
						Enabled: true,
						//Host:    "",
						//Port:    0,
					},
					Single: &Single{
						Enabled:     false,
						Host:        "localhost",
						Port:        8097,
						Name:        "pprof/monitoring server",
						Concurrency: 100000,
						TCPKeepalive: &TCPKeepalive{
							Enabled: true,
							Period:  time.Second,
						},
						Buffer: &Buffer{
							ReadBufferSize:  4 << 20,
							WriteBufferSize: 4 << 20,
						},
						HTTPTransport: HTTPTransport{
							ReadTimeout:         7 * time.Second,
							WriteTimeout:        7 * time.Second,
							MaxIdleConnDuration: 15 * time.Second,
						},
					},
				},
				{
					SomeTypeFirst:  true,
					SomeTypeSecond: 15,
					SomeCron:       "@every 1m",
				},
			},
			want: "",
		},
		{
			name: "2. invalid validation of 'require_if_not_nil_or_empty' tag",
			opts: []*tmpStruct{
				{
					SomeTypeFirst:  true,
					SomeTypeSecond: 0,
					SomeCron:       "* * * * *",
				},
			},
			want: `Key: 'tmpStruct.SomeTypeSecond' Error:Field validation for 'SomeTypeSecond' failed on the 'require_if_not_nil_or_empty' tag`,
		},
		{
			name: "3. invalid validation of 'cron' tag",
			opts: []*tmpStruct{
				{
					SomeTypeFirst:  false,
					SomeTypeSecond: 0,
					SomeCron:       "* & * * *",
				},
			},
			want: `Key: 'tmpStruct.SomeCron' Error:Field validation for 'SomeCron' failed on the 'cron' tag`,
		},
		{
			name: "4. invalid validation of 'host_if_enabled' tag",
			opts: []*tmpStruct{
				{
					SomeTypeFirst:  false,
					SomeTypeSecond: 0,
					SomeCron:       "* * * * *",
					Profiling: Profiling{
						Enabled: true,
						Host:    "",
						Port:    8080,
					},
				},
			},
			want: `Key: 'tmpStruct.Profiling.Host' Error:Field validation for 'Host' failed on the 'host_if_enabled' tag`,
		},
		{
			name: "5. invalid validation of 'port_if_enabled' tag",
			opts: []*tmpStruct{
				{
					SomeTypeFirst:  false,
					SomeTypeSecond: 0,
					SomeCron:       "* * * * *",
					Profiling: Profiling{
						Enabled: true,
						Host:    "localhost",
						Port:    0,
					},
				},
			},
			want: `Key: 'tmpStruct.Profiling.Port' Error:Field validation for 'Port' failed on the 'port_if_enabled' tag`,
		},
	}

	testValidator := validator.New()

	//assert.NoError(t, testValidator.Var(Client{ClusterID: "6900cfdc-38a5-11ec-9742-acde48001122"}, "uuid"))

	for i := range cases {
		tc := cases[i]
		t.Run(tc.name, func(t *testing.T) {
			for j := range tc.opts {
				err := RegisterCustomValidationsTags(context.Background(), testValidator, nil, tc.opts[j])
				assert.NoError(t, err)

				err = testValidator.Struct(tc.opts[j])
				if valErr, ok := err.(validator.ValidationErrors); ok {
					for _, err := range valErr {
						assert.Equal(t, tc.want, err.Error())
					}
				}
			}
		})
	}
}
