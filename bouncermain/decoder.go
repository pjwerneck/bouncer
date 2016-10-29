package bouncermain

import (
	"reflect"
	"strconv"
	"time"

	"github.com/gorilla/schema"
)

func convertDuration(value string) reflect.Value {
	if v, err := strconv.ParseInt(value, 10, 64); err == nil {
		return reflect.ValueOf(time.Duration(v) * time.Millisecond)
	}
	return reflect.Value{}
}

func newDecoder() *schema.Decoder {
	decoder := schema.NewDecoder()
	decoder.RegisterConverter(time.Duration(1), convertDuration)

	return decoder
}
