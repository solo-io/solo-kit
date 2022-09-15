package apiclient

import (
	"io"
	"sort"
	"strings"

	"github.com/golang/protobuf/ptypes"
	"github.com/solo-io/solo-kit/pkg/api/v1/apiserver"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources"
	"github.com/solo-io/solo-kit/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"k8s.io/apimachinery/pkg/labels"
)

type ResourceClient struct {
	grpc         apiserver.ApiServerClient
	resourceType resources.ProtoResource
	token        string
	typeUrl      string
}

func NewResourceClient(cc *grpc.ClientConn, token string, resourceType resources.ProtoResource) *ResourceClient {
	return &ResourceClient{
		grpc:         apiserver.NewApiServerClient(cc),
		resourceType: resourceType,
		token:        token,
		typeUrl:      apiserver.TypeUrl(resourceType),
	}
}

var _ clients.ResourceClient = &ResourceClient{}

func (rc *ResourceClient) Kind() string {
	return resources.Kind(rc.resourceType)
}

func (rc *ResourceClient) NewResource() resources.Resource {
	return resources.Clone(rc.resourceType)
}

func (rc *ResourceClient) Register() error {
	return nil
}

func (rc *ResourceClient) RegisterNamespace(namespace string) error {
	return nil
}

func (rc *ResourceClient) Read(namespace, name string, opts clients.ReadOpts) (resources.Resource, error) {
	if err := resources.ValidateName(name); err != nil {
		return nil, errors.Wrapf(err, "validation error")
	}
	opts = opts.WithDefaults()
	opts.Ctx = metadata.AppendToOutgoingContext(opts.Ctx, "authorization", "bearer "+rc.token)
	resp, err := rc.grpc.Read(opts.Ctx, &apiserver.ReadRequest{
		Name:      name,
		Namespace: namespace,
		TypeUrl:   rc.typeUrl,
	})
	if err != nil {
		if stat, ok := status.FromError(err); ok && strings.Contains(stat.Message(), "does not exist") {
			return nil, errors.NewNotExistErr(namespace, name)
		}
		return nil, errors.Wrapf(err, "performing grpc request")
	}
	resource := rc.NewResource()

	protoResource, err := resources.ProtoCast(resource)
	if err != nil {
		return nil, err
	}

	if err := ptypes.UnmarshalAny(resp.Resource, protoResource); err != nil {
		return nil, errors.Wrapf(err, "failed to unmarshal resource %v", rc.Kind())
	}
	return resource, nil
}

func (rc *ResourceClient) Write(resource resources.Resource, opts clients.WriteOpts) (resources.Resource, error) {
	if err := resources.Validate(resource); err != nil {
		return nil, errors.Wrapf(err, "validation error")
	}
	opts = opts.WithDefaults()
	opts.Ctx = metadata.AppendToOutgoingContext(opts.Ctx, "authorization", "bearer "+rc.token)

	protoResource, err := resources.ProtoCast(resource)
	if err != nil {
		return nil, err
	}
	data, err := ptypes.MarshalAny(protoResource)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to marshal resource")
	}

	resp, err := rc.grpc.Write(opts.Ctx, &apiserver.WriteRequest{
		Resource:          data,
		OverwriteExisting: opts.OverwriteExisting,
	})
	if err != nil {
		if stat, ok := status.FromError(err); ok && strings.Contains(stat.Message(), "exists") {
			return nil, errors.NewExistErr(resource.GetMetadata())
		}
		return nil, errors.Wrapf(err, "performing grpc request")
	}
	written := rc.NewResource()

	protoResource, err = resources.ProtoCast(written)
	if err != nil {
		return nil, err
	}

	if err := ptypes.UnmarshalAny(resp.Resource, protoResource); err != nil {
		return nil, errors.Wrapf(err, "failed to unmarshal resource %v", rc.Kind())
	}
	return written, nil
}

func (rc *ResourceClient) Delete(namespace, name string, opts clients.DeleteOpts) error {
	opts = opts.WithDefaults()
	opts.Ctx = metadata.AppendToOutgoingContext(opts.Ctx, "authorization", "bearer "+rc.token)
	_, err := rc.grpc.Delete(opts.Ctx, &apiserver.DeleteRequest{
		Name:           name,
		Namespace:      namespace,
		TypeUrl:        rc.typeUrl,
		IgnoreNotExist: opts.IgnoreNotExist,
	})
	if err != nil {
		if stat, ok := status.FromError(err); ok && strings.Contains(stat.Message(), "does not exist") {
			return errors.NewNotExistErr(namespace, name)
		}
		return errors.Wrapf(err, "deleting resource %v", name)
	}
	return nil
}

func (rc *ResourceClient) List(namespace string, opts clients.ListOpts) (resources.ResourceList, error) {
	opts = opts.WithDefaults()
	opts.Ctx = metadata.AppendToOutgoingContext(opts.Ctx, "authorization", "bearer "+rc.token)
	resp, err := rc.grpc.List(opts.Ctx, &apiserver.ListRequest{
		Namespace: namespace,
		TypeUrl:   rc.typeUrl,
	})
	if err != nil {
		return nil, errors.Wrapf(err, "performing grpc request")
	}

	var resourceList resources.ResourceList
	for _, resourceData := range resp.ResourceList {
		resource := rc.NewResource()
		protoResource, err := resources.ProtoCast(resource)
		if err != nil {
			return nil, err
		}
		if err := ptypes.UnmarshalAny(resourceData, protoResource); err != nil {
			return nil, errors.Wrapf(err, "failed to unmarshal resource %v", rc.Kind())
		}
		if labels.SelectorFromSet(opts.Selector).Matches(labels.Set(resource.GetMetadata().Labels)) {
			resourceList = append(resourceList, resource)
		}
	}

	sort.SliceStable(resourceList, func(i, j int) bool {
		return resourceList[i].GetMetadata().Name < resourceList[j].GetMetadata().Name
	})

	return resourceList, nil
}

func (rc *ResourceClient) Watch(namespace string, opts clients.WatchOpts) (<-chan resources.ResourceList, <-chan error, error) {
	opts = opts.WithDefaults()
	opts.Ctx = metadata.AppendToOutgoingContext(opts.Ctx, "authorization", "bearer "+rc.token)
	resp, err := rc.grpc.Watch(opts.Ctx, &apiserver.WatchRequest{
		Namespace: namespace,
		TypeUrl:   rc.typeUrl,
	})
	if err != nil {
		return nil, nil, errors.Wrapf(err, "performing grpc request")
	}

	resourcesChan := make(chan resources.ResourceList)
	errs := make(chan error)
	go func() {
		list, err := rc.List(namespace, clients.ListOpts{
			Ctx:      opts.Ctx,
			Selector: opts.Selector,
		})
		if err != nil {
			errs <- err
			return
		}
		resourcesChan <- list
	}()
	go func() {
		for {
			select {
			case <-opts.Ctx.Done():
				close(resourcesChan)
				close(errs)
				return
			default:
				resourceDataList, err := resp.Recv()
				if err == io.EOF {
					errs <- errors.Wrapf(err, "grpc stream closed")
					return
				}
				if err != nil {
					errs <- err
					continue
				}
				var resourceList resources.ResourceList
				for _, resourceData := range resourceDataList.ResourceList {
					resource := rc.NewResource()

					protoResource, err := resources.ProtoCast(resource)
					if err != nil {
						errs <- err
						continue
					}
					if err := ptypes.UnmarshalAny(resourceData, protoResource); err != nil {
						errs <- errors.Wrapf(err, "failed to unmarshal resource %v", rc.Kind())
						continue
					}
					if labels.SelectorFromSet(opts.Selector).Matches(labels.Set(resource.GetMetadata().Labels)) {
						resourceList = append(resourceList, resource)
					}
				}

				sort.SliceStable(resourceList, func(i, j int) bool {
					return resourceList[i].GetMetadata().Name < resourceList[j].GetMetadata().Name
				})
				resourcesChan <- resourceList
			}
		}
	}()

	return resourcesChan, errs, nil
}
