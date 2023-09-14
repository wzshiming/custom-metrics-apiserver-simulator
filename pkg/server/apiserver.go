package server

import (
	"context"
	"strings"

	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apiserver/pkg/endpoints/discovery"
	"k8s.io/apiserver/pkg/endpoints/request"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/tools/cache"

	"github.com/wzshiming/custom-metrics-apiserver-simulator/pkg/apis/v1alpha1"
	"github.com/wzshiming/custom-metrics-apiserver-simulator/pkg/apiserver"
)

func newRequestInfoResolver() *request.RequestInfoFactory {
	return &request.RequestInfoFactory{
		APIPrefixes: sets.NewString(strings.Trim(discovery.APIGroupPrefix, "/")),
	}
}

// InstallAPIServer installs the API server for the server.
func (s *Server) InstallAPIServer(ctx context.Context) error {
	s.discoveryGroupManager = apiserver.InstallRootAPIs(s.restfulCont)
	return nil
}

// InstallOpenAPI installs the OpenAPI spec for the server.
func (s *Server) InstallOpenAPI(ctx context.Context) error {
	webServices := s.restfulCont.RegisteredWebServices()

	_, _, err := apiserver.InstallOpenAPIV2(s.restfulCont, webServices)
	if err != nil {
		return err
	}

	_, err = apiserver.InstallOpenAPIV3(s.restfulCont, webServices)
	if err != nil {
		return err
	}

	return nil
}

// InstallCustomMetricsAPI installs the custom metrics API for the server.
func (s *Server) InstallCustomMetricsAPI(ctx context.Context) error {
	cm := s.typedClient.ApisV1alpha1().CustomMetrics()

	store, controller := cache.NewInformer(
		&cache.ListWatch{
			ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
				return cm.List(ctx, options)
			},
			WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
				return cm.Watch(ctx, options)
			},
		},
		&v1alpha1.CustomMetric{},
		0,
		cache.ResourceEventHandlerFuncs{},
	)
	go controller.Run(ctx.Done())

	customMetricsProvider := apiserver.NewCustomMetricsProvider(apiserver.CustomMetricsProviderConfig{
		DynamicClient: s.dynamicClient,
		RESTMapper:    s.restMapper,
		CustomMetric:  apiserver.NewStore[*v1alpha1.CustomMetric](store),
	})
	err := customMetricsProvider.Install(s.restfulCont, s.discoveryGroupManager)
	if err != nil {
		return err
	}

	return nil
}

// InstallExternalMetricsAPI installs the external metrics API for the server.
func (s *Server) InstallExternalMetricsAPI(ctx context.Context) error {
	em := s.typedClient.ApisV1alpha1().ExternalMetrics("")

	store, controller := cache.NewInformer(
		&cache.ListWatch{
			ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
				return em.List(ctx, options)
			},
			WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
				return em.Watch(ctx, options)
			},
		},
		&v1alpha1.ExternalMetric{},
		0,
		cache.ResourceEventHandlerFuncs{},
	)
	go controller.Run(ctx.Done())

	externalMetricsProvider := apiserver.NewExternalMetricsProvider(apiserver.ExternalMetricProviderConfig{
		ExternalMetric: apiserver.NewStore[*v1alpha1.ExternalMetric](store),
	})
	err := externalMetricsProvider.Install(s.restfulCont, s.discoveryGroupManager)
	if err != nil {
		return err
	}

	return nil
}
