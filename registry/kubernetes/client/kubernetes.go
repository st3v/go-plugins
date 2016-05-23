package client

import "github.com/micro/go-plugins/registry/kubernetes/client/watch"

// Kubernetes ...
type Kubernetes interface {
	UpdatePod(podName string, pod *Pod) (*Pod, error)

	CreateService(s *Service) (*Service, error)
	UpdateService(name string, s *Service) (*Service, error)

	ListEndpoints(labels map[string]string) (*EndpointsList, error)
	GetEndpoints(name string) (*Endpoints, error)
	UpdateEndpoints(name string, s *Endpoints) (*Endpoints, error)
	WatchEndpoints(labels map[string]string) (watch.Watch, error)
}

// Meta ...
type Meta struct {
	Name        string             `json:"name,omitempty"`
	Labels      map[string]*string `json:"labels,omitempty"`
	Annotations map[string]*string `json:"annotations,omitempty"`
}

// Endpoints ...
type Endpoints struct {
	Metadata *Meta    `json:"metadata"`
	Subsets  []Subset `json:"subsets,omitempty"`
}

// EndpointsList ...
type EndpointsList struct {
	Items []Endpoints `json:"items"`
}

// Service ...
type Service struct {
	Metadata *Meta        `json:"metadata"`
	Spec     *ServiceSpec `json:"spec,omitempty"`
}

// ServiceSpec ...
type ServiceSpec struct {
	Ports     []Port            `json:"ports"`
	Selector  map[string]string `json:"selector,omitempty"`
	ClusterIP string            `json:"clusterIP"`
}

// Port ...
type Port struct {
	Name       string      `json:"name,omitempty"`
	Port       int         `json:"port"`
	TargetPort interface{} `json:"targetPort"`
}

// Subset ...
type Subset struct {
	Addresses []Address    `json:"addresses"`
	Ports     []SubsetPort `json:"ports"`
}

// Address ...
type Address struct {
	TargetRef ObjectRef `json:"targetRef"`
	IP        string    `json:"ip"`
}

// ObjectRef ... used as part of the endpoint address to identify pod
type ObjectRef struct {
	Name string
}

// SubsetPort ...
type SubsetPort struct {
	Name string `json:"name,omitempty"`
	Port int    `json:"port"`
}

// PodList ...
type PodList struct {
	Items []Pod `json:"items"`
}

// Pod is the top level item for a pod
type Pod struct {
	Metadata *Meta   `json:"metadata"`
	Status   *Status `json:"status"`
}

// Status ...
type Status struct {
	PodIP string `json:"podIP"`
}
