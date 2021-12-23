package server

import (
	"reflect"
	"testing"

	pb "github.com/dysnix/predictkube-proto/external/proto/services"
)

func TestInjectClientMetadataInterceptor(t *testing.T) {
	var req interface{} = &pb.ReqSendMetrics{}
	st := reflect.TypeOf(req)
	_, ok := st.MethodByName("GetHeader")
	if ok {
		var b interface{} = pb.Header{ClusterId: "bsc-1"}
		field := reflect.New(reflect.TypeOf(b))
		field.Elem().Set(reflect.ValueOf(b))
		reflect.ValueOf(req).Elem().FieldByName("Header").Set(field)
	}

	t.Log(req.(*pb.ReqSendMetrics).Header.ClusterId)
}
