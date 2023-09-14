package main

import (
	"fmt"
	"context"
	"os"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/dynamic"
	"github.com/spf13/pflag"

	"github.com/wzshiming/custom-metrics-apiserver-simulator/pkg/server"
	"github.com/wzshiming/custom-metrics-apiserver-simulator/pkg/client/clientset/versioned"
)

func main() {
	ctx := context.Background()
	f := flagpole{
		Address:    ":8443",
		Kubeconfig: os.Getenv("KUBECONFIG"),
	}

	if f.Kubeconfig == "" {
		f.Kubeconfig = os.Getenv("HOME") + "/.kube/config"
	}

	pflag.StringVar(&f.Address, "address", f.Address, "The address to listen on for secure connections.")
	pflag.StringVar(&f.CertFile, "tls-cert-file", "", "File containing the default x509 Certificate for HTTPS.")
	pflag.StringVar(&f.PrivateKeyFile, "tls-private-key-file", "", "File containing the default x509 private key matching --tls-cert-file.")
	pflag.StringVar(&f.Master, "master", "", "The address of the Kubernetes API server (overrides any value in kubeconfig)")
	pflag.StringVar(&f.Kubeconfig, "kubeconfig", f.Kubeconfig, "Path to kubeconfig file with authorization and master location information.")
	pflag.Parse()

	err := run(ctx, f)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}

type flagpole struct {
	Address        string
	CertFile       string
	PrivateKeyFile string
	Master         string
	Kubeconfig     string
}

func run(ctx context.Context, f flagpole) error {
	cfg, err := clientcmd.BuildConfigFromFlags(f.Master, f.Kubeconfig)
	if err != nil {
		return fmt.Errorf("failed to build config: %w", err)
	}

	typedClient, err := versioned.NewForConfig(cfg)
	if err != nil {
		return fmt.Errorf("failed to create typed client: %w", err)
	}

	dynamicClient, err := dynamic.NewForConfig(cfg)
	if err != nil {
		return fmt.Errorf("failed to create dynamic client: %w", err)
	}
	cli, err := client.New(cfg, client.Options{})
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	conf := server.Config{
		TypedClient:   typedClient,
		DynamicClient: dynamicClient,
		RestMapper:    cli.RESTMapper(),
	}
	svc, err := server.NewServer(conf)
	if err != nil {
		return fmt.Errorf("failed to create server: %w", err)
	}
	err = svc.InstallAPIServer(ctx)
	if err != nil {
		return fmt.Errorf("failed to install apiserver: %w", err)
	}

	err = svc.InstallCustomMetricsAPI(ctx)
	if err != nil {
		return fmt.Errorf("failed to install custom metrics: %w", err)
	}

	err = svc.InstallExternalMetricsAPI(ctx)
	if err != nil {
		return fmt.Errorf("failed to install external metrics: %w", err)
	}

	err = svc.InstallOpenAPI(ctx)
	if err != nil {
		return fmt.Errorf("failed to install openapi: %w", err)
	}

	err = svc.Run(ctx, f.Address, f.CertFile, f.PrivateKeyFile)
	if err != nil {
		return err
	}
	return nil
}
