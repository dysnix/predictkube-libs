package faker

import (
	"github.com/google/uuid"
	"gotest.tools/v3/assert"
	"testing"
	"time"

	"github.com/bxcodec/faker/v3"

	pb "github.com/dysnix/predictkube-proto/external/proto/services"
)

func TestMetricsGenerator(t *testing.T) {
	a := pb.ReqSendMetrics{}

	err := MetricsGenerator(time.Hour)
	if err != nil {
		t.Error(err)
		return
	}

	err = faker.FakeData(&a)
	if err != nil {
		t.Error(err)
		return
	}

	t.Logf("%+v", a)
}

func TestHeaderGenerator(t *testing.T) {
	testData := make([]*pb.Header, 10)
	for i, _ := range testData {
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
		err = faker.FakeData(header)
		if err != nil {
			t.Error(err)
			return
		}

		if header.ClusterId != existUUID.String() {
			assert.Equal(t, header.ClusterId, existUUID.String())
		}
	}
}
