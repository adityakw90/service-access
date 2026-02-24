package response

import (
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/adityakw90/service-access-proto/gen/go/group"
	"github.com/adityakw90/service-access/internal/core/domain/model"
)

func ToGroupResponse(g *model.Group) *group.Group {
	return &group.Group{
		Uid:         g.UID,
		Name:        g.Name,
		Description: g.Description,
		CreatedAt:   timestamppb.New(g.CreatedAt),
		UpdatedAt:   timestamppb.New(g.UpdatedAt),
	}
}
