package grpc

import (
	"net"
	"sync"

	"github.com/adityakw90/go-monitoring"
	"github.com/adityakw90/service-access/internal/adapter/api/grpc/handler"
	"github.com/adityakw90/service-access/internal/adapter/api/grpc/middleware"
	"github.com/adityakw90/service-access/internal/adapter/api/grpc/validator"
	"github.com/adityakw90/service-access/internal/core/port/service"

	accesspb "github.com/adityakw90/service-access-proto/gen/go/access"
	grouppb "github.com/adityakw90/service-access-proto/gen/go/group"
	permissionpb "github.com/adityakw90/service-access-proto/gen/go/permission"
	rolepb "github.com/adityakw90/service-access-proto/gen/go/role"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type Server struct {
	server        *grpc.Server
	permHandler   *handler.PermissionHandler
	roleHandler   *handler.RoleHandler
	groupHandler  *handler.GroupHandler
	accessHandler *handler.AccessHandler
	m             *monitoring.Monitoring
	regOnce       sync.Once
}

func NewServer(
	permService service.PermissionService,
	roleService service.RoleService,
	groupService service.GroupService,
	accessService service.AccessService,
	subjectService service.SubjectService,
	mon *monitoring.Monitoring,
) *Server {
	validator := validator.New()

	// Create handlers
	permHandler := handler.NewPermissionHandler(permService, validator)
	roleHandler := handler.NewRoleHandler(roleService, validator)
	groupHandler := handler.NewGroupHandler(groupService, validator)
	accessHandler := handler.NewAccessHandler(accessService, subjectService, validator)

	// Create gRPC server
	server := grpc.NewServer(
		grpc.UnaryInterceptor(
			middleware.ChainUnaryInterceptors(
				middleware.UnaryRequestInterceptor(mon),
			),
		),
		grpc.StreamInterceptor(
			middleware.ChainStreamInterceptors(
				middleware.StreamRequestInterceptor(mon),
			),
		),
	)

	return &Server{
		server:        server,
		permHandler:   permHandler,
		roleHandler:   roleHandler,
		groupHandler:  groupHandler,
		accessHandler: accessHandler,
	}
}

func (s *Server) RegisterServices() {
	// Register service handlers with the gRPC server
	s.regOnce.Do(func() {
		permissionpb.RegisterPermissionServiceServer(s.server, s.permHandler)
		rolepb.RegisterRoleServiceServer(s.server, s.roleHandler)
		grouppb.RegisterGroupServiceServer(s.server, s.groupHandler)
		accesspb.RegisterAccessControlServiceServer(s.server, s.accessHandler)
		reflection.Register(s.server)
	})
}

func (s *Server) Start(address string) error {
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return err
	}
	s.m.Logger.Info("gRPC server listening", map[string]interface{}{
		"addr": address,
	})

	s.RegisterServices()

	return s.server.Serve(listener)
}

func (s *Server) Stop() {
	s.server.GracefulStop()
}

func (s *Server) GetPermissionHandler() *handler.PermissionHandler {
	return s.permHandler
}

func (s *Server) GetRoleHandler() *handler.RoleHandler {
	return s.roleHandler
}

func (s *Server) GetGroupHandler() *handler.GroupHandler {
	return s.groupHandler
}

func (s *Server) GetAccessHandler() *handler.AccessHandler {
	return s.accessHandler
}
