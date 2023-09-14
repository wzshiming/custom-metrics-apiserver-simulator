package apiserver

import (
	"context"
	"fmt"

	"github.com/emicklei/go-restful/v3"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/sets"
	genericapi "k8s.io/apiserver/pkg/endpoints"
	"k8s.io/apiserver/pkg/endpoints/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/klog/v2"
	"k8s.io/metrics/pkg/apis/custom_metrics"
	specificapi "sigs.k8s.io/custom-metrics-apiserver/pkg/apiserver/installer"
	"sigs.k8s.io/custom-metrics-apiserver/pkg/provider"
	"sigs.k8s.io/custom-metrics-apiserver/pkg/provider/helpers"
	metricstorage "sigs.k8s.io/custom-metrics-apiserver/pkg/registry/custom_metrics"

	"github.com/wzshiming/custom-metrics-apiserver-simulator/pkg/apis/v1alpha1"
	"github.com/wzshiming/custom-metrics-apiserver-simulator/pkg/utils/maps"
	"github.com/wzshiming/custom-metrics-apiserver-simulator/pkg/utils/slices"
)

// CustomMetricsProvider is an implementation of provider.CustomMetricsProvider
type CustomMetricsProvider struct {
	dynamicClient dynamic.Interface
	restMapper    apimeta.RESTMapper
	customMetric  Store[*v1alpha1.CustomMetric]
}

// CustomMetricsProviderConfig is a configuration struct for CustomMetricsProvider
type CustomMetricsProviderConfig struct {
	DynamicClient dynamic.Interface
	RESTMapper    apimeta.RESTMapper
	CustomMetric  Store[*v1alpha1.CustomMetric]
}

// NewCustomMetricsProvider creates a new CustomMetricsProvider
func NewCustomMetricsProvider(conf CustomMetricsProviderConfig) *CustomMetricsProvider {
	p := &CustomMetricsProvider{
		dynamicClient: conf.DynamicClient,
		restMapper:    conf.RESTMapper,
		customMetric:  conf.CustomMetric,
	}
	return p
}

// Install installs the custom metrics API for the server.
func (p *CustomMetricsProvider) Install(container *restful.Container, discoveryGroupManager discovery.GroupManager) error {
	prioritizedVersions := scheme.PrioritizedVersionsForGroup(custom_metrics.GroupName)

	for i, groupVersion := range prioritizedVersions {
		resourceStorage := metricstorage.NewREST(p)

		cmAPI := &specificapi.MetricsAPIGroupVersion{
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
				Namer:           runtime.Namer(apimeta.NewAccessor()),
			},
			ResourceLister: provider.NewCustomMetricResourceLister(p),
			Handlers:       &specificapi.CMHandlers{},
		}
		if err := cmAPI.InstallREST(container); err != nil {
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

// valueFor is a helper function to get just the value of a specific metric
func (p *CustomMetricsProvider) valueFor(info provider.CustomMetricInfo, name types.NamespacedName, metricSelector labels.Selector) (*resource.Quantity, error) {
	cm, err := p.getCustomMetrics(name.Name, name.Namespace, info)
	if err != nil {
		return nil, err
	}

	if cm.Value != nil {
		return cm.Value, nil
	}
	return nil, fmt.Errorf("no value found for metric %s", info.Metric)
}

// metricFor is a helper function which formats a value, metric, and object info into a MetricValue which can be returned by the metrics API
func (p *CustomMetricsProvider) metricFor(value *resource.Quantity, name types.NamespacedName, selector labels.Selector, info provider.CustomMetricInfo, metricSelector *metav1.LabelSelector) (*custom_metrics.MetricValue, error) {
	objRef, err := helpers.ReferenceFor(p.restMapper, name, info)
	if err != nil {
		return nil, err
	}

	metric := &custom_metrics.MetricValue{
		DescribedObject: objRef,
		Metric: custom_metrics.MetricIdentifier{
			Name:     info.Metric,
			Selector: metricSelector,
		},
		WindowSeconds: nil, // TODO: implement this
		Timestamp:     metav1.Now(),
		Value:         *value,
	}

	return metric, nil
}

// metricsFor is a wrapper used by GetMetricBySelector to format several metrics which match a resource selector
func (p *CustomMetricsProvider) metricsFor(namespace string, selector labels.Selector, info provider.CustomMetricInfo, metricSelector labels.Selector) (*custom_metrics.MetricValueList, error) {
	names, err := helpers.ListObjectNames(p.restMapper, p.dynamicClient, namespace, selector, info)
	if err != nil {
		return nil, err
	}

	var metricLabelSelector *metav1.LabelSelector
	if selStr := metricSelector.String(); len(selStr) > 0 {
		metricLabelSelector, err = metav1.ParseToLabelSelector(selStr)
		if err != nil {
			return nil, err
		}
	}

	res := make([]custom_metrics.MetricValue, 0, len(names))
	for _, name := range names {
		namespacedName := types.NamespacedName{Name: name, Namespace: namespace}
		value, err := p.valueFor(info, namespacedName, metricSelector)
		if err != nil {
			if apierrors.IsNotFound(err) {
				continue
			}
			return nil, err
		}

		metric, err := p.metricFor(value, namespacedName, selector, info, metricLabelSelector)
		if err != nil {
			return nil, err
		}
		res = append(res, *metric)
	}

	return &custom_metrics.MetricValueList{
		Items: res,
	}, nil
}

// GetMetricByName returns the value of a single metric for a single object
func (p *CustomMetricsProvider) GetMetricByName(ctx context.Context, name types.NamespacedName, info provider.CustomMetricInfo, metricSelector labels.Selector) (*custom_metrics.MetricValue, error) {
	value, err := p.valueFor(info, name, metricSelector)
	if err != nil {
		return nil, err
	}

	var metricLabelSelector *metav1.LabelSelector
	if selStr := metricSelector.String(); len(selStr) > 0 {
		metricLabelSelector, err = metav1.ParseToLabelSelector(selStr)
		if err != nil {
			return nil, err
		}
	}

	return p.metricFor(value, name, labels.Everything(), info, metricLabelSelector)
}

// GetMetricBySelector returns the value of a metric for all objects which match a resource selector
func (p *CustomMetricsProvider) GetMetricBySelector(ctx context.Context, namespace string, selector labels.Selector, info provider.CustomMetricInfo, metricSelector labels.Selector) (*custom_metrics.MetricValueList, error) {
	return p.metricsFor(namespace, selector, info, metricSelector)
}

// ListAllMetrics returns the list of all metrics provided by this API
func (p *CustomMetricsProvider) ListAllMetrics() []provider.CustomMetricInfo {
	unique := sets.New[provider.CustomMetricInfo]()

	for _, cru := range p.customMetric.List() {
		gv, err := schema.ParseGroupVersion(cru.Spec.ResourceRef.APIGroup)
		if err != nil {
			klog.ErrorS(err, "unable to parse APIGroup for custom metric", "metric", cru.Name)
			continue
		}
		gr := schema.GroupResource{
			Group:    gv.Group,
			Resource: cru.Spec.ResourceRef.Kind,
		}

		m, err := mappingFor(p.restMapper, gr.String())
		if err != nil {
			klog.ErrorS(err, "unable to find GVR for custom metric", "metric", cru.Name)
			continue
		}

		for _, metric := range cru.Spec.Metrics {
			info := provider.CustomMetricInfo{
				GroupResource: m.Resource.GroupResource(),
				Namespaced:    m.Scope.Name() == apimeta.RESTScopeNameNamespace,
				Metric:        metric.Name,
			}
			unique.Insert(info)
		}
	}

	return maps.Keys(unique)
}

func (p *CustomMetricsProvider) getCustomMetrics(name, namespace string, info provider.CustomMetricInfo) (*v1alpha1.CustomMetricItem, error) {
	for _, ccm := range p.customMetric.List() {
		gr := schema.GroupResource{
			Group:    ccm.Spec.ResourceRef.APIGroup,
			Resource: ccm.Spec.ResourceRef.Kind,
		}
		if gr.Group == "v1" {
			gr.Group = ""
		}

		m, err := mappingFor(p.restMapper, gr.String())
		if err != nil {
			klog.ErrorS(err, "unable to find GVR for custom metric")
			continue
		}

		if info.GroupResource != m.Resource.GroupResource() {
			continue
		}

		if ccm.Spec.Selector != nil {
			if len(ccm.Spec.Selector.MatchNames) != 0 && !slices.Contains(ccm.Spec.Selector.MatchNames, name) {
				continue
			}
			if len(ccm.Spec.Selector.MatchNamespaces) != 0 && !slices.Contains(ccm.Spec.Selector.MatchNamespaces, namespace) {
				continue
			}
		}

		item, found := findCustomMetrics(info.Metric, ccm.Spec.Metrics)
		if found {
			return item, nil
		}
	}
	return nil, fmt.Errorf("no custom metric %q found for %q", info.Metric, namespace+"/"+name)
}

func findCustomMetrics(metricName string, items []v1alpha1.CustomMetricItem) (*v1alpha1.CustomMetricItem, bool) {
	var defaultCustomMetricItem *v1alpha1.CustomMetricItem
	for i, item := range items {
		if item.Name == "" {
			defaultCustomMetricItem = &items[i]
			continue
		}
		if item.Name == metricName {
			return &item, true
		}
	}
	return defaultCustomMetricItem, defaultCustomMetricItem != nil
}

// mappingFor returns the RESTMapping for the given resource or kind argument.
func mappingFor(restMapper meta.RESTMapper, resourceOrKindArg string) (*meta.RESTMapping, error) {
	fullySpecifiedGVR, groupResource := schema.ParseResourceArg(resourceOrKindArg)
	gvk := schema.GroupVersionKind{}

	if fullySpecifiedGVR != nil {
		gvk, _ = restMapper.KindFor(*fullySpecifiedGVR)
	}
	if gvk.Empty() {
		gvk, _ = restMapper.KindFor(groupResource.WithVersion(""))
	}
	if !gvk.Empty() {
		return restMapper.RESTMapping(gvk.GroupKind(), gvk.Version)
	}

	fullySpecifiedGVK, groupKind := schema.ParseKindArg(resourceOrKindArg)
	if fullySpecifiedGVK == nil {
		gvk := groupKind.WithVersion("")
		fullySpecifiedGVK = &gvk
	}

	if !fullySpecifiedGVK.Empty() {
		if mapping, err := restMapper.RESTMapping(fullySpecifiedGVK.GroupKind(), fullySpecifiedGVK.Version); err == nil {
			return mapping, nil
		}
	}

	mapping, err := restMapper.RESTMapping(groupKind, gvk.Version)
	if err != nil {
		// if we error out here, it is because we could not match a resource or a kind
		// for the given argument. To maintain consistency with previous behavior,
		// announce that a resource type could not be found.
		// if the error is _not_ a *meta.NoKindMatchError, then we had trouble doing discovery,
		// so we should return the original error since it may help a user diagnose what is actually wrong
		if meta.IsNoMatchError(err) {
			return nil, fmt.Errorf("the server doesn't have a resource type %q", groupResource.Resource)
		}
		return nil, err
	}

	return mapping, nil
}
