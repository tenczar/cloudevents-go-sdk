package cloudevents

import (
	"fmt"
	"net/http"
	"reflect"
)

type HttpRequestExtractor interface {
	Extract(*http.Request) (Event, error)
}

type HttpCloudEventConverter interface {
	CanRead(t reflect.Type, mediaType string) bool
	CanWrite(t reflect.Type, mediaType string) bool
	GetSupportedMediaTypes() []string
	Read(t reflect.Type, req *http.Request) (Event, error)
	Write(t reflect.Type, contentType string, res *http.ResponseWriter)
}

type Converter interface {
	CanRead(t reflect.Type, mediaType string) bool
	CanWrite(t reflect.Type, mediaType string) bool
	Convert(in interface{}, out reflect.Type) error
}

// FromHTTPRequest parses a CloudEvent from any known encoding.
func FromHTTPRequest(req *http.Request, t reflect.Type) (Event, error) {
	// TODO: this should check the version of incoming CloudVersion header and create an appropriate event structure.
	eventType := reflect.TypeOf((*Event)(nil)).Elem()
	ptr := reflect.PtrTo(t)
	if ok := ptr.Implements(eventType); ok {
		e := reflect.New(t)
		version := e.MethodByName("CloudEventVersion").Call([]reflect.Value{})[0].Interface().(string)
		println(version)
		if req.Header.Get("CE-") == "" {

		}
		rets := e.MethodByName("FromHTTPRequest").Call([]reflect.Value{reflect.ValueOf(req)})
		var err error
		if !rets[0].IsNil() {
			err = rets[0].Interface().(error)
		}
		return e.Interface().(Event), err
	}

	return nil, fmt.Errorf("%v does not implement %v", t, eventType)

}
