package fission

import (
	"net/http"

	"errors"
	"io/ioutil"

	"encoding/json"

	"github.com/fission/fission-workflow/pkg/types"
	"github.com/fission/fission-workflow/pkg/types/typedvalues"
	"github.com/sirupsen/logrus"
)

func ParseRequest(r *http.Request, target map[string]*types.TypedValue) error {
	contentType := r.Header.Get("Content-Type")
	logrus.WithField("url", r.URL).WithField("content-type", contentType).Info("Request content-type")
	// Map Inputs to function parameters
	body, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		panic(err)
	}

	var i interface{} = body
	err = json.Unmarshal(body, &i)
	if err != nil {
		logrus.Info("Input is not json: %v", err)
		i = body
	}

	parsedInput, err := typedvalues.Parse(i)
	if err != nil {
		logrus.Errorf("Failed to parse body: %v", err)
		return errors.New("failed to parse body")
	}

	logrus.WithField(types.INPUT_MAIN, parsedInput).Info("Parsed body")
	target[types.INPUT_MAIN] = parsedInput
	return nil
}
