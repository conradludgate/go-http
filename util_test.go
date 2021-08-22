package http_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-test/deep"
	"github.com/jarcoal/httpmock"
)

type Verifier func(req *http.Request) error

func VerifyJSONBody(expected interface{}) Verifier {
	return func(req *http.Request) error {
		if req.Body == nil {
			return errors.New("no body")
		}

		if req.Header.Get("content-type") != "application/json" {
			return errors.New("no json body")
		}

		reqBody := make(map[string]interface{})
		if err := json.NewDecoder(req.Body).Decode(&reqBody); err != nil {
			return errors.New("could not read body")
		}

		expJson, err := json.Marshal(expected)
		if err != nil {
			return errors.New("could not encode expected json")
		}

		expBody := make(map[string]interface{})
		if err := json.Unmarshal(expJson, &expBody); err != nil {
			return errors.New("could not decode expected json")
		}

		if diff := deep.Equal(reqBody, expBody); diff != nil {
			return fmt.Errorf("unexpected request:\n\t%s", strings.Join(diff, "\n\t"))
		}

		return nil
	}
}

func VerifyHeader(key string, values ...string) Verifier {
	return func(req *http.Request) error {
		vs := req.Header.Values(key)
		if len(vs) != len(values) {
			return fmt.Errorf("invalid headers. expected %+v, got %+v", values, vs)
		}
		for i, v := range vs {
			if v != values[i] {
				return fmt.Errorf("invalid headers. expected %+v, got %+v", values, vs)
			}
		}

		return nil
	}
}

func RespondWith(responder httpmock.Responder, verifiers ...Verifier) httpmock.Responder {
	return func(req *http.Request) (*http.Response, error) {
		for _, verifier := range verifiers {
			if err := verifier(req); err != nil {
				return httpmock.NewStringResponse(http.StatusInternalServerError, err.Error()), nil
			}
		}

		return responder(req)
	}
}

func JSON(response interface{}) httpmock.Responder {
	return func(req *http.Request) (*http.Response, error) {
		return httpmock.NewJsonResponse(http.StatusOK, response)
	}
}
