package faker

import (
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/google/uuid"
	"gotest.tools/v3/assert"

	pb "github.com/dysnix/predictkube-proto/external/proto/services"
)

func TestMetricsGenerator(t *testing.T) {
	a := pb.ReqSendMetrics{}

	err := MetricsGenerator(time.Hour)
	if err != nil {
		t.Error(err)
		return
	}

	gofakeit.Struct(&a)
	t.Logf("%+v", a)
}

func TestHeaderGenerator(t *testing.T) {
	testData := make([]*pb.Header, 10)
	for i := range testData {
		testData[i] = &pb.Header{}
	}

	existUUID, err := uuid.NewUUID()
	if err != nil {
		t.Error(err)
		return
	}

	err = MetricsGenerator(time.Hour, existUUID.String())
	if err != nil {
		t.Error(err)
		return
	}

	for _, header := range testData {
		gofakeit.Struct(header)
		header.ClusterId = existUUID.String()
		assert.Equal(t, header.ClusterId, existUUID.String())
	}
}
