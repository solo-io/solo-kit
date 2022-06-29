package crd

import (
	"log"
	"sync"

	"github.com/rotisserie/eris"

	"github.com/solo-io/solo-kit/pkg/api/v1/clients/kube/crd/client/clientset/versioned/scheme"
	v1 "github.com/solo-io/solo-kit/pkg/api/v1/clients/kube/crd/solo.io/v1"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources"
	"github.com/solo-io/solo-kit/pkg/utils/kubeutils"
	"github.com/solo-io/solo-kit/pkg/utils/protoutils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// TODO(ilackarms): evaluate this fix for concurrent map access in k8s.io/apimachinery/pkg/runtime.SchemaBuider
var (
	registerLock sync.Mutex

	MarshalErr = func(err error, detail string) error {
		return eris.Wrapf(err, "internal error: failed to marshal %s", detail)
	}
)

type CrdMeta struct {
	Plural        string
	Group         string
	KindName      string
	ShortName     string
	ClusterScoped bool
}

type Version struct {
	Version string
	Type    runtime.Object
}

type Crd struct {
	CrdMeta
	Version Version
}

type MultiVersionCrd struct {
	CrdMeta
	Versions []Version
}

func NewCrd(
	plural string,
	group string,
	version string,
	kindName string,
	shortName string,
	clusterScoped bool,
	objType runtime.Object) Crd {
	c := Crd{
		CrdMeta: CrdMeta{
			Plural:        plural,
			Group:         group,
			KindName:      kindName,
			ShortName:     shortName,
			ClusterScoped: clusterScoped,
		},
		Version: Version{
			Version: version,
			Type:    objType,
		},
	}
	if err := c.AddToScheme(scheme.Scheme); err != nil {
		log.Panicf("error while adding [%v] CRD to scheme: %v", c.FullName(), err)
	}
	return c
}

func (d Crd) KubeResource(resource resources.InputResource) (*v1.Resource, error) {
	var (
		spec   v1.Spec
		status v1.Status
	)

	if customResource, ok := resource.(resources.CustomInputResource); ok {
		// Handle custom spec/status marshalling

		var err error
		spec, err = customResource.MarshalSpec()
		if err != nil {
			return nil, MarshalErr(err, "resource spec to map")
		}
		status, _ = customResource.MarshalStatus()

	} else {
		// Default marshalling

		data, err := protoutils.MarshalMap(resource)
		if err != nil {
			return nil, MarshalErr(err, "resource to map")
		}

		delete(data, "metadata")
		delete(data, "status")
		spec = data

		if resource.GetStatus() != nil {
			statusProto := resource.GetStatus()
			statusMap, _ := protoutils.MarshalMapFromProtoWithEnumsAsInts(statusProto)
			status = statusMap
		}
	}

	// Since v1.Status is a map, it can be nil when sent to kube api server.
	// In order to avoid this scenario, make sure to set it in the case that it is nil
	if status == nil {
		status = v1.Status{}
	}

	return &v1.Resource{
		TypeMeta:   d.TypeMeta(),
		ObjectMeta: kubeutils.ToKubeMetaMaintainNamespace(resource.GetMetadata()),
		Status:     status,
		Spec:       &spec,
	}, nil
}

func (d CrdMeta) FullName() string {
	return d.Plural + "." + d.Group
}

func (d Crd) TypeMeta() metav1.TypeMeta {
	return metav1.TypeMeta{
		Kind:       d.KindName,
		APIVersion: d.Group + "/" + d.Version.Version,
	}
}

// GroupVersion is group version used to register these objects
func (d Crd) GroupVersion() schema.GroupVersion {
	return schema.GroupVersion{Group: d.Group, Version: d.Version.Version}
}

// GroupVersionKing is the unique id of this crd
func (d Crd) GroupVersionKind() schema.GroupVersionKind {
	return schema.GroupVersionKind{Group: d.Group, Version: d.Version.Version, Kind: d.KindName}
}

// Kind takes an unqualified kind and returns back a Group qualified GroupKind
func (d Crd) Kind(kind string) schema.GroupKind {
	return d.GroupVersion().WithKind(kind).GroupKind()
}

func (d CrdMeta) GroupKind() schema.GroupKind {
	return schema.GroupKind{Group: d.Group, Kind: d.KindName}
}

// Resource takes an unqualified resource and returns a Group qualified GroupResource
func (d Crd) Resource(resource string) schema.GroupResource {
	return d.GroupVersion().WithResource(resource).GroupResource()
}

func (d Crd) SchemeBuilder() runtime.SchemeBuilder {
	return runtime.NewSchemeBuilder(func(scheme *runtime.Scheme) error {
		scheme.AddKnownTypeWithName(d.GroupVersion().WithKind(d.KindName), &v1.Resource{})
		scheme.AddKnownTypeWithName(d.GroupVersion().WithKind(d.KindName+"List"), &v1.ResourceList{})

		metav1.AddToGroupVersion(scheme, d.GroupVersion())
		return nil
	})
}

func (d Crd) AddToScheme(s *runtime.Scheme) error {
	registerLock.Lock()
	defer registerLock.Unlock()
	builder := d.SchemeBuilder()
	return (&builder).AddToScheme(s)
}
