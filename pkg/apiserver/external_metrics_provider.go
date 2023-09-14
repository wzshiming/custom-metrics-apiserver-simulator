package apiserver

import (
	"context"
	"fmt"

	"github.com/emicklei/go-restful/v3"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/sets"
	genericapi "k8s.io/apiserver/pkg/endpoints"
	"k8s.io/apiserver/pkg/endpoints/discovery"
	"k8s.io/metrics/pkg/apis/external_metrics"
	specificapi "sigs.k8s.io/custom-metrics-apiserver/pkg/apiserver/installer"
	"sigs.k8s.io/custom-metrics-apiserver/pkg/provider"
	metricstorage "sigs.k8s.io/custom-metrics-apiserver/pkg/registry/external_metrics"

	"github.com/wzshiming/custom-metrics-apiserver-simulator/pkg/apis/v1alpha1"
	"github.com/wzshiming/custom-metrics-apiserver-simulator/pkg/utils/maps"
	"github.com/wzshiming/custom-metrics-apiserver-simulator/pkg/utils/slices"
)

// ExternalMetricProviderConfig is a configuration struct for ExternalMetricProvider
type ExternalMetricProviderConfig struct {
	ExternalMetric Store[*v1alpha1.ExternalMetric]
}

// ExternalMetricsProvider is an implementation of provider.ExternalMetricsProvider
type ExternalMetricsProvider struct {
	externalMetric Store[*v1alpha1.ExternalMetric]
}

// NewExternalMetricsProvider returns a new ExternalMetricsProvider
func NewExternalMetricsProvider(conf ExternalMetricProviderConfig) *ExternalMetricsProvider {
	p := &ExternalMetricsProvider{
		externalMetric: conf.ExternalMetric,
	}
	return p
}

// Install registers the external metrics API and handlers
func (p *ExternalMetricsProvider) Install(container *restful.Container, discoveryGroupManager discovery.GroupManager) error {
	prioritizedVersions := scheme.PrioritizedVersionsForGroup(external_metrics.GroupName)

	for i, groupVersion := range prioritizedVersions {
		resourceStorage := metricstorage.NewREST(p)

		emAPI := &specificapi.MetricsAPIGroupVersion{
			DynamicStorage: resourceStorage,
			APIGroupVersion: &genericapi.APIGroupVersion{
				Root:            discovery.APIGroupPrefix,
				GroupVersion:    groupVersion,
				ParameterCodec:  parameterCodec,
				Serializer:      codecs,
				Creater:         scheme,
				Convertor:       scheme,
				UnsafeConvertor: runtime.UnsafeObjectConvertor(scheme),
				Typer:           scheme,
				Namer:           runtime.Namer(meta.NewAccessor()),
			},
			ResourceLister: provider.NewExternalMetricResourceLister(p),
			Handlers:       &specificapi.EMHandlers{},
		}
		if err := emAPI.InstallREST(container); err != nil {
			return err
		}

		if i == 0 {
			gvfd := metav1.GroupVersionForDiscovery{
				GroupVersion: groupVersion.String(),
				Version:      groupVersion.Version,
			}
			apiGroup := metav1.APIGroup{
				Name:             groupVersion.Group,
				Versions:         []metav1.GroupVersionForDiscovery{gvfd},
				PreferredVersion: gvfd,
			}
			discoveryGroupManager.AddGroup(apiGroup)
			container.Add(discovery.NewAPIGroupHandler(codecs, apiGroup).WebService())
		}
	}
	return nil
}

// GetExternalMetric returns the external metric value for the given metric name and selector
func (p *ExternalMetricsProvider) GetExternalMetric(ctx context.Context, namespace string, metricSelector labels.Selector, info provider.ExternalMetricInfo) (*external_metrics.ExternalMetricValueList, error) {
	em, found := slices.Find(p.externalMetric.List(), func(m *v1alpha1.ExternalMetric) bool {
		name := m.Name
		if m.Spec.Name != "" {
			name = m.Spec.Name
		}
		return name == info.Metric && m.Namespace == namespace
	})
	if !found {
		return nil, fmt.Errorf("no external metric %q found for %q", info.Metric, namespace+"/"+info.Metric)
	}
	if len(em.Spec.Metrics) == 0 {
		return &external_metrics.ExternalMetricValueList{
			Items: []external_metrics.ExternalMetricValue{},
		}, nil
	}

	now := metav1.Now()

	metricValues := slices.Map(em.Spec.Metrics, func(m v1alpha1.ExternalMetricItem) external_metrics.ExternalMetricValue {
		return external_metrics.ExternalMetricValue{
			MetricName:   info.Metric,
			MetricLabels: em.Labels,
			Value:        *m.Value,
			Timestamp:    now,
		}
	})

	return &external_metrics.ExternalMetricValueList{
		Items: metricValues,
	}, nil
}

// ListAllExternalMetrics returns the list of all external metrics
func (p *ExternalMetricsProvider) ListAllExternalMetrics() []provider.ExternalMetricInfo {
	infos := sets.New[provider.ExternalMetricInfo]()

	for _, m := range p.externalMetric.List() {
		name := m.Name
		if m.Spec.Name != "" {
			name = m.Spec.Name
		}
		infos.Insert(provider.ExternalMetricInfo{
			Metric: name,
		})
	}

	return maps.Keys(infos)
}
