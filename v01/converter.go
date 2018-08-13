package v01

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"mime"
	"net/http"
	"net/textproto"
	"reflect"
	"strings"
	"time"

	"github.com/dispatchframework/cloudevents-go-sdk"
)

type DefaultHttpRequestExtractor struct {
	converters []cloudevents.HttpCloudEventConverter
}

func NewDefaultHttpRequestExtractor(converters []cloudevents.HttpCloudEventConverter) *DefaultHttpRequestExtractor {
	return &DefaultHttpRequestExtractor{
		converters: converters,
	}
}

func (e *DefaultHttpRequestExtractor) Extract(req *http.Request) (cloudevents.Event, error) {
	if req == nil {
		return nil, errors.New("cannot process nil-request")
	}

	mimeType, _, err := mime.ParseMediaType(req.Header.Get("Content-Type"))
	if err != nil {
		return nil, fmt.Errorf("error parsing request content type: %s", err.Error())
	}

	for _, v := range e.converters {
		if v.CanRead(reflect.TypeOf(Event{}), mimeType) {
			return v.Read(reflect.TypeOf(Event{}), req)
		}
	}
	return nil, cloudevents.ContentTypeNotSupportedError(mimeType)
}

type JsonHttpCloudEventConverter struct {
	supportedMediaTypes      map[string]bool
	supportedMediaTypesSlice []string
}

func NewJsonHttpCloudEventConverter() *JsonHttpCloudEventConverter {
	mediaTypes := map[string]bool{
		"application/cloudevents+json": true,
	}
	var mediaTypesSlice []string
	for k := range mediaTypes {
		mediaTypesSlice = append(mediaTypesSlice, k)
	}
	return &JsonHttpCloudEventConverter{
		supportedMediaTypes:      mediaTypes,
		supportedMediaTypesSlice: mediaTypesSlice,
	}
}

func (j *JsonHttpCloudEventConverter) CanRead(t reflect.Type, mediaType string) bool {
	ptr := reflect.PtrTo(t)
	return ptr.Implements(reflect.TypeOf((*cloudevents.Event)(nil)).Elem()) && j.supportedMediaTypes[mediaType]
}

func (j *JsonHttpCloudEventConverter) CanWrite(t reflect.Type, mediaType string) bool {
	return j.supportedMediaTypes[mediaType]
}

func (j *JsonHttpCloudEventConverter) GetSupportedMediaTypes() []string {
	return j.supportedMediaTypesSlice
}

func (j *JsonHttpCloudEventConverter) Read(t reflect.Type, req *http.Request) (cloudevents.Event, error) {
	e := reflect.New(t).Interface()
	err := json.NewDecoder(req.Body).Decode(e)

	if err != nil {
		return nil, fmt.Errorf("error parsing request: %s", err.Error())
	}
	return e.(cloudevents.Event), nil
}

func (j *JsonHttpCloudEventConverter) Write(t reflect.Type, contentType string, res *http.ResponseWriter) {
	(*res).Header().Set("Content-Type", contentType)
	json.NewEncoder(*res).Encode(t)
}

type BinaryHttpCloudEventConverter struct {
}

func NewBinaryHttpCloudEventConverter() *BinaryHttpCloudEventConverter {
	return &BinaryHttpCloudEventConverter{}
}

func (b *BinaryHttpCloudEventConverter) CanRead(t reflect.Type, mediaType string) bool {
	ptr := reflect.PtrTo(t)
	return ptr.Implements(reflect.TypeOf((*cloudevents.Event)(nil)).Elem())
}

func (b *BinaryHttpCloudEventConverter) CanWrite(t reflect.Type, mediaType string) bool {
	ptr := reflect.PtrTo(t)
	return ptr.Implements(reflect.TypeOf((*cloudevents.Event)(nil)).Elem())
}

func (b *BinaryHttpCloudEventConverter) GetSupportedMediaTypes() []string {
	return []string{}
}

func (b *BinaryHttpCloudEventConverter) Read(t reflect.Type, req *http.Request) (cloudevents.Event, error) {
	e := reflect.New(t)
	numFields := e.Elem().NumField()
	for i := 0; i < numFields; i++ {
		f := t.Field(i)
		tag := f.Tag.Get("cloudevent")
		if tag == "" {
			continue
		}

		ceTagProps := strings.Split(tag, ",")

		props := map[string]bool{}
		for _, v := range ceTagProps[1:] {
			props[v] = true
		}

		name := headerize(f.Name)
		if ceTagProps[0] != "" {
			name = ceTagProps[0]
		}

		if props["map"] {
			extensions := make(map[string]interface{})
			canonicalName := textproto.CanonicalMIMEHeaderKey(name)
			for key, value := range req.Header {
				if strings.HasPrefix(key, canonicalName) {
					extensions[strings.TrimPrefix(key, canonicalName)] = value
				}
			}

			e.Elem().Field(i).Set(reflect.ValueOf(extensions))

			continue
		}

		v := req.Header.Get(name)

		if v == "" && props["required"] {
			return nil, fmt.Errorf("unable to parse event context from request headers: %s", cloudevents.RequiredPropertyError(name))
		}

		if f.Type.AssignableTo(reflect.PtrTo(reflect.TypeOf(time.Time{}))) {
			if v == "" {
				continue
			}
			eventTime, err := time.Parse(time.RFC3339, v)
			if err != nil {
				return nil, fmt.Errorf("error parsing the %s header: %s", name, err.Error())
			}
			e.Elem().Field(i).Set(reflect.ValueOf(&eventTime))
			continue
		}

		e.Elem().Field(i).Set(reflect.ValueOf(v))
	}

	// TODO: implement encoder/decoder registry

	if req.ContentLength == 0 {
		return e.Interface().(cloudevents.Event), nil
	}

	data, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading request body: %s", err.Error())
	}
	e.Elem().FieldByName("Data").Set(reflect.ValueOf(data))
	return e.Interface().(cloudevents.Event), nil
}

func (b *BinaryHttpCloudEventConverter) Write(t reflect.Type, contentType string, res *http.ResponseWriter) {

}
