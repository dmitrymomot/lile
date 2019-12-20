// Package fromenv provides utilities to create lile options from environment
// variables. fromenv will error with fatal if it cannot resolve or errors
// Source: https://github.com/lileio/fromenv/blob/master/fromenv.go
package fromenv

import (
	"log"
	"os"

	"github.com/lileio/pubsub/v2"
	"github.com/lileio/pubsub/v2/providers/google"
	"github.com/lileio/pubsub/v2/providers/nats"
	opentracing "github.com/opentracing/opentracing-go"
	zipkinot "github.com/openzipkin-contrib/zipkin-go-opentracing"
	"github.com/openzipkin/zipkin-go"
	zipkinhttp "github.com/openzipkin/zipkin-go/reporter/http"
	"github.com/sirupsen/logrus"
)

// Tracer ...
func Tracer(name string) opentracing.Tracer {
	zipkinHost := os.Getenv("USE_ZIPKIN")
	if zipkinHost == "" {
		return opentracing.GlobalTracer()
	}

	addr := "http://zipkin/api/v1/spans"
	if os.Getenv("ZIPKIN_ADDR") != "" {
		addr = os.Getenv("ZIPKIN_ADDR")
	}

	// create our local service endpoint
	endpoint, _ := zipkin.NewEndpoint(name, name)

	logrus.Infof("Using Zipkin HTTP tracer: %s", addr)
	reporter := zipkinhttp.NewReporter(addr)

	// initialize our tracer
	nativeTracer, err := zipkin.NewTracer(reporter, zipkin.WithLocalEndpoint(endpoint))
	if err != nil {
		log.Fatalf("unable to create tracer: %+v\n", err)
	}

	// use zipkin-go-opentracing to wrap our tracer
	tracer := zipkinot.Wrap(nativeTracer)

	// optionally set as Global OpenTracing tracer instance
	opentracing.SetGlobalTracer(tracer)

	return tracer
}

// PubSubProvider ...
func PubSubProvider() pubsub.Provider {
	gpid := os.Getenv("GOOGLE_PUBSUB_PROJECT_ID")
	if gpid != "" {
		gc, err := google.NewGoogleCloud(gpid)
		if err != nil {
			logrus.Fatalf("fronenv: Google Cloud pubsub err: %s", err)
			return nil
		}

		logrus.Infof("Using Google Cloud pubsub: %s", gpid)
		return gc
	}

	natsCluster := os.Getenv("NATS_CLUSTER")
	if natsCluster != "" {
		n, err := nats.NewNats(natsCluster)
		if err != nil {
			logrus.Fatalf("fronenv: Nats Streaming pubsub err: %s", err)
			return nil
		}

		logrus.Infof("Using Nats Streaming pubsub: %s", natsCluster)
		return n
	}

	logrus.Warn("Using noop pubsub provider")
	return pubsub.NoopProvider{}
}
