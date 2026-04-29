package grpc

import (
	"context"
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
	subjectpb "github.com/adityakw90/service-access-proto/gen/go/subject"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type Server struct {
	server         *grpc.Server
	listener       net.Listener
	listenerMu     sync.RWMutex
	permHandler    *handler.PermissionHandler
	roleHandler    *handler.RoleHandler
	groupHandler   *handler.GroupHandler
	accessHandler  *handler.AccessHandler
	subjectHandler *handler.SubjectHandler
	m              *monitoring.Monitoring
	regOnce        sync.Once
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
	accessHandler := handler.NewAccessHandler(accessService, validator)
	subjectHandler := handler.NewSubjectHandler(subjectService, validator)

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
		server:         server,
		permHandler:    permHandler,
		roleHandler:    roleHandler,
		groupHandler:   groupHandler,
		accessHandler:  accessHandler,
		subjectHandler: subjectHandler,
		m:              mon,
	}
}

func (s *Server) RegisterServices() {
	// Register service handlers with the gRPC server
	s.regOnce.Do(func() {
		permissionpb.RegisterPermissionServiceServer(s.server, s.permHandler)
		rolepb.RegisterRoleServiceServer(s.server, s.roleHandler)
		grouppb.RegisterGroupServiceServer(s.server, s.groupHandler)
		accesspb.RegisterAccessControlServiceServer(s.server, s.accessHandler)
		subjectpb.RegisterSubjectServiceServer(s.server, s.subjectHandler)
		reflection.Register(s.server)
	})
}

func (s *Server) Start(address string) error {
	var err error
	s.listenerMu.Lock()
	lc := net.ListenConfig{}
	s.listener, err = lc.Listen(context.Background(), "tcp", address)
	s.listenerMu.Unlock()
	if err != nil {
		return err
	}
	s.m.Logger.Info("gRPC server listening", map[string]interface{}{
		"addr": address,
	})

	s.RegisterServices()
	s.m.Logger.Info("register service", map[string]interface{}{
		"addr": address,
	})

	return s.server.Serve(s.listener)
}

func (s *Server) Stop() {
	if s.server != nil {
		s.server.GracefulStop()
	}
	s.listenerMu.RLock()
	defer s.listenerMu.RUnlock()
	if s.listener != nil {
		s.listener.Close()
	}
}

// Addr returns the server address.
func (s *Server) Addr() string {
	s.listenerMu.RLock()
	defer s.listenerMu.RUnlock()
	if s.listener != nil {
		return s.listener.Addr().String()
	}
	return ""
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

func (s *Server) GetSubjectHandler() *handler.SubjectHandler {
	return s.subjectHandler
}
