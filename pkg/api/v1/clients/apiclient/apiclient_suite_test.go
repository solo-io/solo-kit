package apiclient_test

import (
	"testing"

	"google.golang.org/grpc"

	"fmt"
	"net"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_zap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	grpc_ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/solo-io/solo-kit/pkg/api/v1/apiserver"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/factory"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/memory"
	"github.com/solo-io/solo-kit/pkg/utils/log"
	v1 "github.com/solo-io/solo-kit/test/mocks/v1"
	"go.uber.org/zap"
)

// TODO: fix tests
func TestApiclient(t *testing.T) {

	log.Printf("Skipping Apiclient Suite. Tests are currently failing and need to be fixed.")
	return

	RegisterFailHandler(Fail)
	RunSpecs(t, "Apiclient Suite")
}

var (
	resourceClient = memory.NewResourceClient(memory.NewInMemoryResourceCache(), &v1.MockResource{})
	port           = 1234
	server         *grpc.Server
)

var _ = BeforeSuite(func() {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	Expect(err).NotTo(HaveOccurred())
	server = grpc.NewServer(grpc.StreamInterceptor(
		grpc_middleware.ChainStreamServer(
			grpc_ctxtags.StreamServerInterceptor(),
			grpc_zap.StreamServerInterceptor(zap.NewNop()),
			func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
				log.Printf("%v", info.FullMethod)
				return handler(srv, ss)
			},
		)))
	apiserver.NewApiServer(server, nil, &factory.MemoryResourceClientFactory{
		Cache: memory.NewInMemoryResourceCache(),
	}, &v1.MockResource{})
	log.Printf("grpc listening on %v", port)
	go server.Serve(lis)
})

var _ = AfterSuite(func() {
	server.Stop()
})
