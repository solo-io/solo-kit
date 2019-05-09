package apiclient_test

import (
	"context"

	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/test/helpers"
	"google.golang.org/grpc"

	"fmt"
	"net"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_zap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	grpc_ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"time"

	"github.com/solo-io/solo-kit/pkg/api/v1/apiserver"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/factory"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/memory"
	v1 "github.com/solo-io/solo-kit/test/mocks/v1"
	"go.uber.org/zap"

	. "github.com/solo-io/solo-kit/pkg/api/v1/clients/apiclient"
	"github.com/solo-io/solo-kit/test/tests/generic"
)

var _ = Describe("Apiclient", func() {

	var (
		port   = 0
		server *grpc.Server
		lis    net.Listener
		client *ResourceClient
		cc     *grpc.ClientConn
	)
	BeforeEach(func() {
		var err error
		lis, err = net.Listen("tcp", fmt.Sprintf(":0"))
		Expect(err).NotTo(HaveOccurred())
		server = grpc.NewServer(grpc.StreamInterceptor(
			grpc_middleware.ChainStreamServer(
				grpc_ctxtags.StreamServerInterceptor(),
				grpc_zap.StreamServerInterceptor(zap.NewNop()),
				func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
					fmt.Fprintf(GinkgoWriter, "%v\n", info.FullMethod)
					return handler(srv, ss)
				},
			)), grpc.UnaryInterceptor(
			grpc_middleware.ChainUnaryServer(
				grpc_ctxtags.UnaryServerInterceptor(),
				grpc_zap.UnaryServerInterceptor(zap.NewNop()),
				func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
					fmt.Fprintf(GinkgoWriter, "%v\n", info.FullMethod)
					return handler(ctx, req)
				},
			)))
		apiserver.NewApiServer(server, nil, &factory.MemoryResourceClientFactory{
			Cache: memory.NewInMemoryResourceCache(),
		}, time.Second, &v1.MockResource{})

		port = lis.Addr().(*net.TCPAddr).Port
		fmt.Fprintf(GinkgoWriter, "grpc listening on %v\n", port)
		go server.Serve(lis)

		// now start the client:

		cc, err = grpc.Dial(fmt.Sprintf("localhost:%v", port), grpc.WithInsecure())
		Expect(err).NotTo(HaveOccurred())
		client = NewResourceClient(cc, "foo", &v1.MockResource{})
	})

	AfterEach(func() {
		server.Stop()
		lis.Close()
	})

	AfterEach(func() {
		cc.Close()
	})
	It("CRUDs resources", func() {
		selector := map[string]string{
			helpers.TestLabel: helpers.RandString(8),
		}
		generic.TestCrudClient("test", client, clients.WatchOpts{
			Selector:    selector,
			Ctx:         context.TODO(),
			RefreshRate: time.Minute,
		})
	})
})
