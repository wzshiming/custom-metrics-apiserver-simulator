package apiserver

import (
	"github.com/emicklei/go-restful/v3"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apiserver/pkg/endpoints/discovery"
	cminstall "k8s.io/metrics/pkg/apis/custom_metrics/install"
	eminstall "k8s.io/metrics/pkg/apis/external_metrics/install"
	"sigs.k8s.io/custom-metrics-apiserver/pkg/apiserver/installer"
)

var (
	scheme         = runtime.NewScheme()
	codecs         = serializer.NewCodecFactory(scheme)
	parameterCodec = runtime.NewParameterCodec(scheme)
)

func init() {
	cminstall.Install(scheme)
	eminstall.Install(scheme)
	utilruntime.Must(installer.RegisterConversions(scheme))
}

// InstallRootAPIs installs the root APIs for the apiserver.
func InstallRootAPIs(container *restful.Container) discovery.GroupManager {
	handler := discovery.NewRootAPIsHandler(discovery.CIDRRule{}, codecs)
	container.Handle(discovery.APIGroupPrefix, handler)
	return handler
}
