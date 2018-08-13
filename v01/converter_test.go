package v01_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	cloudevents "github.com/dispatchframework/cloudevents-go-sdk"
	"github.com/dispatchframework/cloudevents-go-sdk/v01"
)

func TestDefaultHttpRequestExtractorJsonSuccess(t *testing.T) {
	jsonConverter := v01.NewJsonHttpCloudEventConverter()
	extractor := v01.NewDefaultHttpRequestExtractor([]cloudevents.HttpCloudEventConverter{jsonConverter})

	event := v01.Event{
		EventType:        "dispatch",
		EventTypeVersion: "0.1",
		EventID:          "00001",
		Source:           "dispatch",
	}

	var buffer bytes.Buffer
	json.NewEncoder(&buffer).Encode(event)
	req := httptest.NewRequest("GET", "/", &buffer)
	req.Header = http.Header{}
	req.Header.Set("Content-Type", "application/cloudevents+json")
	req.Header.Set("CE-eventType", "dispatch")
	req.Header.Set("CE-source", "dispatch")
	req.Header.Set("CE-eventID", "00001")

	e, err := extractor.Extract(req)
	if err != nil {
		t.Errorf("Failed converting to json with error: %v", err)
	}

	t.Errorf("Got event: %+v", e)
}

func TestDefaultHttpRequestExtractorBinarySuccess(t *testing.T) {
	binaryConverter := v01.NewBinaryHttpCloudEventConverter()
	extractor := v01.NewDefaultHttpRequestExtractor([]cloudevents.HttpCloudEventConverter{binaryConverter})

	header := map[string][]string{
		"Content-Type":           []string{"text/plain"},
		"Ce-Eventtype":           []string{"dispatch"},
		"Ce-Source":              []string{"dispatch"},
		"Ce-Eventid":             []string{"00001"},
		"Ce-X-My-Extension":      []string{"myvalue"},
		"Ce-X-Another-Extension": []string{"anothervalue"},
		"Ce-Eventtime":           []string{"2018-08-08T15:00:00-07:00"},
	}

	req := &http.Request{
		Header: header,
	}

	e, err := extractor.Extract(req)
	if err != nil {
		t.Errorf("Failed converting to binary with error: %v", err)
	}

	t.Errorf("Got event: %+v", e)
}
