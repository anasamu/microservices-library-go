package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/anasamu/microservices-library-go/discovery/gateway"
	"github.com/anasamu/microservices-library-go/discovery/providers/consul"
	"github.com/anasamu/microservices-library-go/discovery/providers/etcd"
	"github.com/anasamu/microservices-library-go/discovery/providers/kubernetes"
	"github.com/anasamu/microservices-library-go/discovery/providers/static"
	"github.com/anasamu/microservices-library-go/discovery/types"
	"github.com/sirupsen/logrus"
)

func main() {
	// Create logger
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)

	// Create discovery manager
	managerConfig := &gateway.ManagerConfig{
		DefaultProvider: "consul",
		RetryAttempts:   3,
		RetryDelay:      time.Second,
		Timeout:         30 * time.Second,
		FallbackEnabled: true,
	}

	manager := gateway.NewDiscoveryManager(managerConfig, logger)

	// Example 1: Using Consul provider
	fmt.Println("=== Consul Provider Example ===")
	consulExample(manager, logger)

	// Example 2: Using Static provider
	fmt.Println("\n=== Static Provider Example ===")
	staticExample(manager, logger)

	// Example 3: Using etcd provider
	fmt.Println("\n=== etcd Provider Example ===")
	etcdExample(manager, logger)

	// Example 4: Using Kubernetes provider
	fmt.Println("\n=== Kubernetes Provider Example ===")
	kubernetesExample(manager, logger)

	// Example 5: Service discovery and watching
	fmt.Println("\n=== Service Discovery and Watching Example ===")
	discoveryExample(manager, logger)

	// Clean up
	manager.Close()
}

func consulExample(manager *gateway.DiscoveryManager, logger *logrus.Logger) {
	// Create Consul provider
	consulConfig := &consul.ConsulConfig{
		Address:    "localhost:8500",
		Datacenter: "dc1",
		Timeout:    30 * time.Second,
	}

	consulProvider, err := consul.NewConsulProvider(consulConfig, logger)
	if err != nil {
		log.Printf("Failed to create Consul provider: %v", err)
		return
	}

	// Register provider
	err = manager.RegisterProvider(consulProvider)
	if err != nil {
		log.Printf("Failed to register Consul provider: %v", err)
		return
	}

	// Connect to Consul
	ctx := context.Background()
	err = consulProvider.Connect(ctx)
	if err != nil {
		log.Printf("Failed to connect to Consul: %v", err)
		return
	}

	// Register a service
	registration := &types.ServiceRegistration{
		ID:       "web-service-1",
		Name:     "web-service",
		Address:  "192.168.1.100",
		Port:     8080,
		Protocol: "http",
		Tags:     []string{"web", "api", "v1"},
		Metadata: map[string]string{
			"version": "1.0.0",
			"env":     "production",
		},
		Health: types.HealthPassing,
		TTL:    30 * time.Second,
	}

	err = manager.RegisterService(ctx, registration)
	if err != nil {
		log.Printf("Failed to register service: %v", err)
		return
	}

	fmt.Printf("Service registered: %s\n", registration.Name)

	// Discover services
	query := &types.ServiceQuery{
		Name: "web-service",
	}

	services, err := manager.DiscoverServices(ctx, query)
	if err != nil {
		log.Printf("Failed to discover services: %v", err)
		return
	}

	fmt.Printf("Discovered %d services\n", len(services))
	for _, service := range services {
		fmt.Printf("  Service: %s, Instances: %d\n", service.Name, len(service.Instances))
		for _, instance := range service.Instances {
			fmt.Printf("    Instance: %s (%s:%d) - %s\n",
				instance.ID, instance.Address, instance.Port, instance.Health)
		}
	}
}

func staticExample(manager *gateway.DiscoveryManager, logger *logrus.Logger) {
	// Create Static provider
	staticConfig := &static.StaticConfig{
		Services: []*types.ServiceRegistration{
			{
				ID:       "db-service-1",
				Name:     "database",
				Address:  "192.168.1.200",
				Port:     5432,
				Protocol: "tcp",
				Tags:     []string{"database", "postgresql"},
				Metadata: map[string]string{
					"version": "13.0",
					"env":     "production",
				},
				Health: types.HealthPassing,
			},
			{
				ID:       "cache-service-1",
				Name:     "cache",
				Address:  "192.168.1.201",
				Port:     6379,
				Protocol: "tcp",
				Tags:     []string{"cache", "redis"},
				Metadata: map[string]string{
					"version": "6.0",
					"env":     "production",
				},
				Health: types.HealthPassing,
			},
		},
		Timeout: 30 * time.Second,
	}

	staticProvider, err := static.NewStaticProvider(staticConfig, logger)
	if err != nil {
		log.Printf("Failed to create Static provider: %v", err)
		return
	}

	// Register provider
	err = manager.RegisterProvider(staticProvider)
	if err != nil {
		log.Printf("Failed to register Static provider: %v", err)
		return
	}

	// Connect to Static provider
	ctx := context.Background()
	err = staticProvider.Connect(ctx)
	if err != nil {
		log.Printf("Failed to connect to Static provider: %v", err)
		return
	}

	// Discover all services
	query := &types.ServiceQuery{}

	services, err := manager.DiscoverServicesWithProvider(ctx, "static", query)
	if err != nil {
		log.Printf("Failed to discover services: %v", err)
		return
	}

	fmt.Printf("Discovered %d services from static provider\n", len(services))
	for _, service := range services {
		fmt.Printf("  Service: %s, Instances: %d\n", service.Name, len(service.Instances))
		for _, instance := range service.Instances {
			fmt.Printf("    Instance: %s (%s:%d) - %s\n",
				instance.ID, instance.Address, instance.Port, instance.Health)
		}
	}
}

func etcdExample(manager *gateway.DiscoveryManager, logger *logrus.Logger) {
	// Create etcd provider
	etcdConfig := &etcd.EtcdConfig{
		Endpoints: []string{"localhost:2379"},
		Namespace: "/services",
		Timeout:   30 * time.Second,
	}

	etcdProvider, err := etcd.NewEtcdProvider(etcdConfig, logger)
	if err != nil {
		log.Printf("Failed to create etcd provider: %v", err)
		return
	}

	// Register provider
	err = manager.RegisterProvider(etcdProvider)
	if err != nil {
		log.Printf("Failed to register etcd provider: %v", err)
		return
	}

	// Connect to etcd
	ctx := context.Background()
	err = etcdProvider.Connect(ctx)
	if err != nil {
		log.Printf("Failed to connect to etcd: %v", err)
		return
	}

	// Register a service
	registration := &types.ServiceRegistration{
		ID:       "message-service-1",
		Name:     "message-service",
		Address:  "192.168.1.300",
		Port:     9090,
		Protocol: "grpc",
		Tags:     []string{"message", "grpc", "v1"},
		Metadata: map[string]string{
			"version": "1.0.0",
			"env":     "production",
		},
		Health: types.HealthPassing,
		TTL:    60 * time.Second,
	}

	err = manager.RegisterServiceWithProvider(ctx, "etcd", registration)
	if err != nil {
		log.Printf("Failed to register service: %v", err)
		return
	}

	fmt.Printf("Service registered with etcd: %s\n", registration.Name)

	// Discover services
	query := &types.ServiceQuery{
		Name: "message-service",
	}

	services, err := manager.DiscoverServicesWithProvider(ctx, "etcd", query)
	if err != nil {
		log.Printf("Failed to discover services: %v", err)
		return
	}

	fmt.Printf("Discovered %d services from etcd\n", len(services))
	for _, service := range services {
		fmt.Printf("  Service: %s, Instances: %d\n", service.Name, len(service.Instances))
		for _, instance := range service.Instances {
			fmt.Printf("    Instance: %s (%s:%d) - %s\n",
				instance.ID, instance.Address, instance.Port, instance.Health)
		}
	}
}

func kubernetesExample(manager *gateway.DiscoveryManager, logger *logrus.Logger) {
	// Create Kubernetes provider
	k8sConfig := &kubernetes.KubernetesConfig{
		Namespace: "default",
		InCluster: false, // Set to true if running inside Kubernetes
		Timeout:   30 * time.Second,
	}

	k8sProvider, err := kubernetes.NewKubernetesProvider(k8sConfig, logger)
	if err != nil {
		log.Printf("Failed to create Kubernetes provider: %v", err)
		return
	}

	// Register provider
	err = manager.RegisterProvider(k8sProvider)
	if err != nil {
		log.Printf("Failed to register Kubernetes provider: %v", err)
		return
	}

	// Connect to Kubernetes
	ctx := context.Background()
	err = k8sProvider.Connect(ctx)
	if err != nil {
		log.Printf("Failed to connect to Kubernetes: %v", err)
		return
	}

	// Discover services
	query := &types.ServiceQuery{}

	services, err := manager.DiscoverServicesWithProvider(ctx, "kubernetes", query)
	if err != nil {
		log.Printf("Failed to discover services: %v", err)
		return
	}

	fmt.Printf("Discovered %d services from Kubernetes\n", len(services))
	for _, service := range services {
		fmt.Printf("  Service: %s, Instances: %d\n", service.Name, len(service.Instances))
		for _, instance := range service.Instances {
			fmt.Printf("    Instance: %s (%s:%d) - %s\n",
				instance.ID, instance.Address, instance.Port, instance.Health)
		}
	}
}

func discoveryExample(manager *gateway.DiscoveryManager, logger *logrus.Logger) {
	ctx := context.Background()

	// Get provider information
	providerInfo := manager.GetProviderInfo()
	fmt.Printf("Registered providers: %d\n", len(providerInfo))
	for name, info := range providerInfo {
		fmt.Printf("  Provider: %s, Connected: %v, Features: %d\n",
			name, info.IsConnected, len(info.SupportedFeatures))
	}

	// List all services from all providers
	allServices, err := manager.ListServices(ctx)
	if err != nil {
		log.Printf("Failed to list services: %v", err)
		return
	}

	fmt.Printf("\nAll services across providers:\n")
	for providerName, services := range allServices {
		fmt.Printf("  %s: %d services\n", providerName, len(services))
		for _, serviceName := range services {
			fmt.Printf("    - %s\n", serviceName)
		}
	}

	// Get statistics
	stats, err := manager.GetStats(ctx)
	if err != nil {
		log.Printf("Failed to get stats: %v", err)
		return
	}

	fmt.Printf("\nProvider statistics:\n")
	for providerName, stat := range stats {
		fmt.Printf("  %s: %d services, %d instances\n",
			providerName, stat.ServicesRegistered, stat.InstancesActive)
	}

	// Example of watching services (if supported by provider)
	watchOptions := &types.WatchOptions{
		ServiceName: "web-service",
		Timeout:     10 * time.Second,
	}

	eventChan, err := manager.WatchServices(ctx, watchOptions)
	if err != nil {
		log.Printf("Watch not supported or failed: %v", err)
		return
	}

	// Listen for events for a short time
	timeout := time.After(5 * time.Second)
	for {
		select {
		case event, ok := <-eventChan:
			if !ok {
				fmt.Println("Watch channel closed")
				return
			}
			fmt.Printf("Service event: %s for service %s\n", event.Type, event.Service.Name)
		case <-timeout:
			fmt.Println("Watch timeout reached")
			return
		}
	}
}
