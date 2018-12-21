package v02_test

import (
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	cloudevents "github.com/cloudevents/sdk-go"
	"github.com/cloudevents/sdk-go/v02"
)

func TestBuildSuccess(t *testing.T) {
	now := time.Now()
	principal := v02.NewCloudEventBuilder()
	event, err := principal.
		SpecVersion(cloudevents.Version02).
		ID("1234-1234-1234").
		Source(url.URL{
			Scheme: "http",
			Host:   "example.com",
			Path:   "/cloudevent",
		}).
		Type("com.example.cloudevent").
		ContentType("application/cloudevents+json").
		SchemaURL(url.URL{
			Scheme: "http",
			Host:   "example.com",
			Path:   "/cloudevent",
		}).
		Time(now).
		Data(map[string]interface{}{
			"key1": "val1",
			"key2": "val2",
		}).
		Build()

	assert.Nil(t, err)
	assert.Equal(t, event.SpecVersion, cloudevents.Version02)
	assert.Equal(t, event.ID, "1234-1234-1234")
	assert.Equal(t, event.Type, "com.example.cloudevent")
	assert.Equal(t, event.Source.String(), "http://example.com/cloudevent")

	assert.Equal(t, event.ContentType, "application/cloudevents+json")
	assert.Equal(t, event.Data, map[string]interface{}{
		"key1": "val1",
		"key2": "val2",
	})
	assert.Equal(t, event.SchemaURL.String(), "http://example.com/cloudevent")
	assert.Equal(t, event.Time, &now)
}

func TestDefaultSpecVersion(t *testing.T) {
	principal := v02.NewCloudEventBuilder()
	event, err := principal.
		ID("1234-1234-1234").
		Source(url.URL{
			Scheme: "http",
			Host:   "example.com",
			Path:   "/cloudevent",
		}).
		Type("com.example.cloudevent").
		Build()

	assert.Nil(t, err)
	assert.Equal(t, event.SpecVersion, cloudevents.Version02)
	assert.Equal(t, event.ID, "1234-1234-1234")
	assert.Equal(t, event.Type, "com.example.cloudevent")
	assert.Equal(t, event.Source.String(), "http://example.com/cloudevent")
}

func TestBuildMissingRequiredProperty(t *testing.T) {
	principal := v02.NewCloudEventBuilder()
	event, err := principal.
		SpecVersion(cloudevents.Version02).
		Build()

	assert.Error(t, err)
	assert.Zero(t, event)

	event, err = principal.
		Type("com.example.cloudevent").
		Build()

	assert.Error(t, err)
	assert.Zero(t, event)

	event, err = principal.
		ID("1234-1234-1234").
		Build()

	assert.Error(t, err)
	assert.Zero(t, event)

	event, err = principal.
		Source(url.URL{
			Scheme: "http",
			Host:   "example.com",
			Path:   "/cloudevent",
		}).
		Build()

	assert.NoError(t, err)
	assert.Equal(t, event.ID, "1234-1234-1234")
	assert.Equal(t, event.Source.String(), "http://example.com/cloudevent")
	assert.Equal(t, event.Type, "com.example.cloudevent")
}

func TestBuildWithExtensions(t *testing.T) {
	principal := v02.NewCloudEventBuilder()
	event, err := principal.
		ID("1234-1234-1234").
		Type("com.example.cloudevent").
		Source(url.URL{
			Scheme: "http",
			Host:   "example.com",
			Path:   "/cloudevent",
		}).
		Extension("myextension", "myvalue").
		Build()

	assert.NoError(t, err)

	val, ok := event.Get("myextension")
	assert.True(t, ok)
	assert.Equal(t, val, "myvalue")
}
