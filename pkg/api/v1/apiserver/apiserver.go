package apiserver

import (
	"context"
	"time"

	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/auth"

	"github.com/gogo/protobuf/types"
	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/factory"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources"
	"github.com/solo-io/solo-kit/pkg/errors"
	"google.golang.org/grpc"
)

type Callbacks interface {
	OnRegister(ctx context.Context, req *RegisterRequest) (*RegisterResponse, error)
	OnRead(ctx context.Context, req *ReadRequest) (*ReadResponse, error)
	OnWrite(ctx context.Context, req *WriteRequest) (*WriteResponse, error)
	OnDelete(ctx context.Context, req *DeleteRequest) (*DeleteResponse, error)
	OnList(ctx context.Context, req *ListRequest) (*ListResponse, error)
	OnWatch(req *WatchRequest, watch ApiServer_WatchServer) error
}

type ApiServer struct {
	callbacks            Callbacks
	resourceTypes        map[string]resources.ProtoResource
	factory              factory.ResourceClientFactory
	watchRefreshDuration time.Duration
}

func TypeUrl(m resources.ProtoResource) string {
	data, err := types.MarshalAny(m)
	if err != nil {
		panic("failed to marshal resource " + err.Error())
	}
	return data.TypeUrl
}

// TODO(yuval-k): can we get rid of watchRefreshDuration ? sounds like this should be the ResourceClientFactory
// responsibility
func NewApiServer(s *grpc.Server, callbacks Callbacks, factory factory.ResourceClientFactory, watchRefreshDuration time.Duration, resourceTypes ...resources.ProtoResource) ApiServerServer {
	mapped := make(map[string]resources.ProtoResource)
	for _, resource := range resourceTypes {
		mapped[TypeUrl(resource)] = resource
	}
	srv := &ApiServer{
		callbacks:            callbacks,
		resourceTypes:        mapped,
		factory:              factory,
		watchRefreshDuration: watchRefreshDuration,
	}
	RegisterApiServerServer(s, srv)
	return srv
}

func tokenFromCtx(ctx context.Context) (string, error) {
	return grpc_auth.AuthFromMD(ctx, "bearer")
}

func (s *ApiServer) resourceClient(ctx context.Context, resourceKind string) (clients.ResourceClient, error) {
	token, err := tokenFromCtx(ctx)
	if err != nil {
		return nil, errors.Wrapf(err, "retrieving auth token from request")
	}
	if token == "" {
		return nil, errors.Errorf("auth token cannot be empty")
	}
	resourceType, ok := s.resourceTypes[resourceKind]
	if !ok {
		return nil, errors.Errorf("no resource type registered for kind %s", resourceKind)
	}
	return s.factory.NewResourceClient(factory.NewResourceClientParams{
		Token:        token,
		ResourceType: resourceType,
	})
}

func (s *ApiServer) Register(ctx context.Context, req *RegisterRequest) (*RegisterResponse, error) {
	if s.callbacks != nil {
		resp, err := s.callbacks.OnRegister(ctx, req)
		if err != nil {
			return resp, err
		}
	}
	for kind := range s.resourceTypes {
		rc, err := s.resourceClient(ctx, kind)
		if err != nil {
			return nil, err
		}
		if err := rc.Register(); err != nil {
			return nil, errors.Wrapf(err, "failed to register client %v", rc.Kind())
		}
	}
	return &RegisterResponse{}, nil
}

func (s *ApiServer) Read(ctx context.Context, req *ReadRequest) (*ReadResponse, error) {
	if s.callbacks != nil {
		resp, err := s.callbacks.OnRead(ctx, req)
		if err != nil {
			return resp, err
		}
	}
	rc, err := s.resourceClient(ctx, req.TypeUrl)
	if err != nil {
		return nil, err
	}
	resource, err := rc.Read(req.Namespace, req.Name, clients.ReadOpts{
		Ctx: contextutils.WithLogger(ctx, "apiserver.read"),
	})
	if err != nil {
		return nil, err
	}
	protoResource, err := resources.ProtoCast(resource)
	if err != nil {
		return nil, err
	}
	data, err := types.MarshalAny(protoResource)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to marshal resource")
	}
	return &ReadResponse{
		Resource: data,
	}, nil
}

func (s *ApiServer) Write(ctx context.Context, req *WriteRequest) (*WriteResponse, error) {
	if s.callbacks != nil {
		resp, err := s.callbacks.OnWrite(ctx, req)
		if err != nil {
			return resp, err
		}
	}
	rc, err := s.resourceClient(ctx, req.Resource.TypeUrl)
	if err != nil {
		return nil, err
	}
	resource := rc.NewResource()
	protoResource, err := resources.ProtoCast(resource)
	if err != nil {
		return nil, err
	}
	if err := types.UnmarshalAny(req.Resource, protoResource); err != nil {
		return nil, errors.Wrapf(err, "failed to unmarshal resource %v", rc.Kind())
	}
	resource, err = rc.Write(resource, clients.WriteOpts{
		OverwriteExisting: req.OverwriteExisting,
		Ctx:               contextutils.WithLogger(ctx, "apiserver.write"),
	})
	if err != nil {
		return nil, err
	}
	protoResource, err = resources.ProtoCast(resource)
	if err != nil {
		return nil, err
	}

	data, err := types.MarshalAny(protoResource)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to marshal resource")
	}

	return &WriteResponse{
		Resource: data,
	}, nil
}

func (s *ApiServer) Delete(ctx context.Context, req *DeleteRequest) (*DeleteResponse, error) {
	if s.callbacks != nil {
		resp, err := s.callbacks.OnDelete(ctx, req)
		if err != nil {
			return resp, err
		}
	}
	rc, err := s.resourceClient(ctx, req.TypeUrl)
	if err != nil {
		return nil, err
	}
	if err := rc.Delete(req.Namespace, req.Name, clients.DeleteOpts{
		IgnoreNotExist: req.IgnoreNotExist,
		Ctx:            contextutils.WithLogger(ctx, "apiserver.delete"),
	}); err != nil {
		return nil, errors.Wrapf(err, "failed to delete resource %v", req.TypeUrl)
	}
	return &DeleteResponse{}, nil
}

func (s *ApiServer) List(ctx context.Context, req *ListRequest) (*ListResponse, error) {
	if s.callbacks != nil {
		resp, err := s.callbacks.OnList(ctx, req)
		if err != nil {
			return resp, err
		}
	}
	rc, err := s.resourceClient(ctx, req.TypeUrl)
	if err != nil {
		return nil, err
	}
	resourceList, err := rc.List(req.Namespace, clients.ListOpts{
		Ctx: contextutils.WithLogger(ctx, "apiserver.read"),
	})
	var resourceListResponse []*types.Any
	for _, resource := range resourceList {
		protoResource, err := resources.ProtoCast(resource)
		if err != nil {
			return nil, err
		}
		data, err := types.MarshalAny(protoResource)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to marshal resource %v", req.TypeUrl)
		}

		resourceListResponse = append(resourceListResponse, data)
	}
	return &ListResponse{
		ResourceList: resourceListResponse,
	}, nil
}

func (s *ApiServer) Watch(req *WatchRequest, watch ApiServer_WatchServer) error {
	if s.callbacks != nil {
		err := s.callbacks.OnWatch(req, watch)
		if err != nil {
			return err
		}
	}
	rc, err := s.resourceClient(watch.Context(), req.TypeUrl)
	if err != nil {
		return err
	}
	ctx := contextutils.WithLogger(watch.Context(), "apiserver.read")

	resourceWatch, errs, err := rc.Watch(req.Namespace, clients.WatchOpts{
		RefreshRate: s.watchRefreshDuration,
		Ctx:         ctx,
	})
	for {
		select {
		case resourceList := <-resourceWatch:
			var resourceListResponse []*types.Any
			for _, resource := range resourceList {
				protoResource, err := resources.ProtoCast(resource)
				if err != nil {
					return err
				}
				data, err := types.MarshalAny(protoResource)
				if err != nil {
					return errors.Wrapf(err, "failed to marshal resource %v", req.TypeUrl)
				}

				resourceListResponse = append(resourceListResponse, data)
			}
			if err := watch.Send(&ListResponse{
				ResourceList: resourceListResponse,
			}); err != nil {
				return errors.Wrapf(err, "failed to send list response on watch")
			}
		case err := <-errs:
			return errors.Wrapf(err, "error during %v watch", req.TypeUrl)
		case <-ctx.Done():
			return nil
		}
	}
}
