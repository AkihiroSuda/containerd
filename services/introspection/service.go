package introspection

import (
	api "github.com/containerd/containerd/api/services/introspection"
	"github.com/containerd/containerd/plugin"
	"github.com/containerd/containerd/version"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

var _ api.IntrospectionServer = &Service{}

func init() {
	plugin.Register("introspection-grpc", &plugin.Registration{
		Type: plugin.GRPCPlugin,
		Init: New,
	})
}

func New(ic *plugin.InitContext) (interface{}, error) {
	return &Service{}, nil
}

type Service struct {
}

func (s *Service) Register(server *grpc.Server) error {
	api.RegisterIntrospectionServer(server, s)
	return nil
}

func (s *Service) IntrospectVersion(ctx context.Context, r *api.IntrospectVersionRequest) (*api.Version, error) {
	return &api.Version{
		Version:  version.Version,
		Revision: version.Revision,
	}, nil
}

func (s *Service) IntrospectAll(ctx context.Context, r *api.IntrospectAllRequest) (*api.IntrospectAllResponse, error) {
	v, err := s.IntrospectVersion(ctx, nil)
	if err != nil {
		return nil, err
	}
	return &api.IntrospectAllResponse{Version: v}, nil
}
