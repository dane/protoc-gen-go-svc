package main

import (
	"testing"

	overridev1 "github.com/dane/protoc-gen-go-svc/example/override/v1"
	service "github.com/dane/protoc-gen-go-svc/example/proto/go/service"
	servicev1 "github.com/dane/protoc-gen-go-svc/example/proto/go/service/v1"
	testingv1 "github.com/dane/protoc-gen-go-svc/example/proto/go/service/v1/testing"
	testingv2 "github.com/dane/protoc-gen-go-svc/example/proto/go/service/v2/testing"
)

func TestV2(t *testing.T) {
	tests := []struct {
		Fn      testingv2.TestFunc
		Params  testingv2.Params
		Options []service.Option
	}{
		{
			Fn: testingv2.NewCreateConversionTest,
			Params: testingv2.Params{
				PublicInput:   "testdata/conversions/v2/create-request.json",
				PublicOutput:  "testdata/conversions/v2/create-response.json",
				PrivateInput:  "testdata/conversions/private/create-request-v2.json",
				PrivateOutput: "testdata/conversions/private/create-response-v2.json",
			},
		},
		{
			Fn: testingv2.NewBatchConversionTest,
			Params: testingv2.Params{
				PublicInput:   "testdata/conversions/v2/batch-request.json",
				PublicOutput:  "testdata/conversions/v2/batch-response.json",
				PrivateInput:  "testdata/conversions/private/batch-request-v2.json",
				PrivateOutput: "testdata/conversions/private/batch-response-v2.json",
			},
		},
	}

	for _, test := range tests {
		test.Fn(t, test.Params, test.Options)
	}
}

func TestV1(t *testing.T) {
	tests := []struct {
		Fn      testingv1.TestFunc
		Params  testingv1.Params
		Options []service.Option
	}{
		{
			Fn: testingv1.NewCreateConversionTest,
			Params: testingv1.Params{
				PublicInput:   "testdata/conversions/v1/create-request.json",
				PublicOutput:  "testdata/conversions/v1/create-response.json",
				PrivateInput:  "testdata/conversions/private/create-request-all.json",
				PrivateOutput: "testdata/conversions/private/create-response-all.json",
			},
			Options: []service.Option{
				overridev1.Converter{servicev1.NewConverter()},
			},
		},
	}

	for _, test := range tests {
		test.Fn(t, test.Params, test.Options)
	}
}
