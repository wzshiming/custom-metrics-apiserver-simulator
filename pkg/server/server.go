package server

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/emicklei/go-restful/v3"
	"github.com/gorilla/handlers"
	"github.com/wzshiming/cmux"
	"github.com/wzshiming/cmux/pattern"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apiserver/pkg/endpoints/discovery"
	"k8s.io/apiserver/pkg/endpoints/filters"
	"k8s.io/client-go/dynamic"
	"k8s.io/klog/v2"
	"net/http/httptest"

	"github.com/wzshiming/custom-metrics-apiserver-simulator/pkg/client/clientset/versioned"
)

// Server is a server that can serve HTTP/HTTPS requests.
type Server struct {
	ctx context.Context

	typedClient   versioned.Interface
	dynamicClient dynamic.Interface
	restMapper    meta.RESTMapper

	discoveryGroupManager discovery.GroupManager

	restfulCont *restful.Container
}

// Config holds configurations needed by the server handlers.
type Config struct {
	TypedClient   versioned.Interface
	DynamicClient dynamic.Interface
	RestMapper    meta.RESTMapper
}

// NewServer creates a new Server.
func NewServer(conf Config) (*Server, error) {
	container := restful.NewContainer()

	s := &Server{
		typedClient:   conf.TypedClient,
		dynamicClient: conf.DynamicClient,
		restMapper:    conf.RestMapper,
		restfulCont:   container,
	}

	return s, nil
}

// Run runs the specified Server.
// This should never exit.
func (s *Server) Run(ctx context.Context, address string, certFile, privateKeyFile string) error {
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return err
	}

	muxListener := cmux.NewMuxListener(listener)
	tlsListener, err := muxListener.MatchPrefix(pattern.Pattern[pattern.TLS]...)
	if err != nil {
		return fmt.Errorf("match tls listener: %w", err)
	}
	unmatchedListener, err := muxListener.Unmatched()
	if err != nil {
		return fmt.Errorf("unmatched listener: %w", err)
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	s.ctx = ctx

	errCh := make(chan error, 1)

	var handler http.Handler = s.restfulCont

	handler = filters.WithRequestInfo(handler, newRequestInfoResolver())

	handler = handlers.CombinedLoggingHandler(os.Stderr, handler)

	if certFile != "" && privateKeyFile != "" {
		go func() {
			klog.InfoS("Starting HTTPS server",
				"address", address,
				"cert", certFile,
				"key", privateKeyFile,
			)
			svc := &http.Server{
				ReadHeaderTimeout: 5 * time.Second,
				BaseContext: func(_ net.Listener) context.Context {
					return ctx
				},
				Addr:    address,
				Handler: handler,
			}
			err = svc.ServeTLS(tlsListener, certFile, privateKeyFile)
			if err != nil {
				errCh <- fmt.Errorf("serve https: %w", err)
			}
		}()
	} else {
		klog.InfoS("Starting test HTTPS server",
			"address", address,
		)
		svc := httptest.Server{
			Listener: tlsListener,
			Config: &http.Server{
				ReadHeaderTimeout: 5 * time.Second,
				BaseContext: func(_ net.Listener) context.Context {
					return ctx
				},
				Addr:    address,
				Handler: handler,
			},
		}
		svc.StartTLS()
	}

	go func() {
		klog.InfoS("Starting HTTP server",
			"address", address,
		)
		svc := &http.Server{
			ReadHeaderTimeout: 5 * time.Second,
			BaseContext: func(_ net.Listener) context.Context {
				return ctx
			},
			Addr:    address,
			Handler: handler,
		}
		err = svc.Serve(unmatchedListener)
		if err != nil {
			errCh <- fmt.Errorf("serve http: %w", err)
		}
	}()

	select {
	case err = <-errCh:
	case <-ctx.Done():
		err = ctx.Err()
	}

	return err
}
