package apiclient_test

import (
	"context"
	"fmt"
	"net"
	"time"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_zap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	grpc_ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/solo-io/solo-kit/pkg/api/v1/apiserver"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	. "github.com/solo-io/solo-kit/pkg/api/v1/clients/apiclient"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/factory"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/memory"
	"github.com/solo-io/solo-kit/test/helpers"
	v1 "github.com/solo-io/solo-kit/test/mocks/v1"
	"github.com/solo-io/solo-kit/test/tests/generic"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

var _ = Describe("Apiclient", func() {

	var (
		ctx    context.Context
		cancel context.CancelFunc

		port   = 0
		server *grpc.Server
		lis    net.Listener
		client *ResourceClient
		cc     *grpc.ClientConn
	)

	BeforeEach(func() {
		ctx, cancel = context.WithCancel(context.Background())

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
		_, err = fmt.Fprintf(GinkgoWriter, "grpc listening on %v\n", port)
		Expect(err).NotTo(HaveOccurred())

		go func() {
			defer GinkgoRecover()

			_ = server.Serve(lis)
		}()

		// now start the client:

		cc, err = grpc.Dial(fmt.Sprintf("localhost:%v", port), grpc.WithInsecure())
		Expect(err).NotTo(HaveOccurred())
		client = NewResourceClient(cc, "foo", &v1.MockResource{})
	})

	AfterEach(func() {
		cancel()
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
		generic.TestCrudClient("test1", "test2", client, clients.WatchOpts{
			Selector:    selector,
			Ctx:         ctx,
			RefreshRate: time.Minute,
		})
	})
})
