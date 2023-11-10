/*
  Copyright 2020 Amazon.com, Inc. or its affiliates. All Rights Reserved.
  Licensed under the Apache License, Version 2.0 (the "License").
  You may not use this file except in compliance with the License.
  A copy of the License is located at
      http://www.apache.org/licenses/LICENSE-2.0
  or in the "license" file accompanying this file. This file is distributed
  on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either
  express or implied. See the License for the specific language governing
  permissions and limitations under the License.
*/

package main

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"

	"github.com/aws-samples/amazon-ecr-repository-compliance-webhook/pkg/function"
	"github.com/aws-samples/amazon-ecr-repository-compliance-webhook/pkg/webhook"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecr"
	log "github.com/sirupsen/logrus"
	"k8s.io/api/admission/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
)

func init() {
	loglvl := logLevels(os.Getenv("LOG_LEVEL"))
	log.SetFormatter(new(log.JSONFormatter))
	log.Infof("Got log level: `%s`", loglvl)
	log.SetLevel(loglvl)
	log.SetOutput(os.Stdout)
}

var (
	sess, sessError = session.NewSession()
	svc             = ecr.New(sess, &aws.Config{Region: getRegistryRegion()})
	container       = function.NewContainer(svc)

	// decoding
	runtimeScheme = runtime.NewScheme()
	codecs        = serializer.NewCodecFactory(runtimeScheme)
	deserializer  = codecs.UniversalDeserializer()
)

func main() {
	log.Infof("Starting server")

	if sessError != nil {
		log.Errorf("Loading aws session failed %s", sessError)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", handleRequest)

	err := http.ListenAndServe(":3333", mux)
	if err != nil {
		log.Infof("Server failed with error %s", err)
	}
}

func handleRequest(w http.ResponseWriter, r *http.Request) {
	log.Infof("Got request on %s\n", r.URL.Path)

	val, ok := r.Header[http.CanonicalHeaderKey("Content-Type")]
	if !ok {
		http.Error(w, "ErrMissingContentType", http.StatusBadRequest)
		return
	}
	if val[0] != "application/json" {
		http.Error(w, "ErrInvalidContentType", http.StatusBadRequest)
		return
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "ErrMissingBody", http.StatusBadRequest)
		return
	}
	log.Infof("Got body %s %d", body, len(body))

	var review v1beta1.AdmissionReview
	if _, _, err := deserializer.Decode([]byte(body), nil, &review); err != nil {
		http.Error(w, "ErrDecode", http.StatusBadRequest)
		return
	}

	req := &webhook.Request{Admission: review.Request}
	log.Infof("Got parse req %v", *req)

	res, err := container.HandleRequest(context.Background(), req)
	if err != nil {
		http.Error(w, "Could not handle request", http.StatusInternalServerError)
		return
	}
	log.Infof("Sending response %+v", *res)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

func logLevels(lvl string) log.Level {
	loglvl, err := log.ParseLevel(lvl)
	if err != nil {
		return log.InfoLevel
	}

	return loglvl
}

func getRegistryRegion() *string {
	if value, ok := os.LookupEnv("REGISTRY_REGION"); ok {
		return aws.String(value)
	}
	return aws.String(os.Getenv("AWS_REGION"))
}
