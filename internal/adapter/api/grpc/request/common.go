package request

import (
	"strings"

	common "github.com/adityakw90/service-access-proto/gen/go/common"
	"github.com/adityakw90/service-access/internal/core/domain/param"
)

type PaginationRequest struct {
	Page    int    `validate:"required,min=1"`
	Limit   int    `validate:"required,min=1,max=100"`
	Sort    string `validate:"omitempty,oneof=asc desc"`
	OrderBy string `validate:"omitempty"`
}

func (pr *PaginationRequest) ToPaginationParam() *param.PaginationParam {
	return &param.PaginationParam{
		Page:    &pr.Page,
		Limit:   &pr.Limit,
		Sort:    &pr.Sort,
		OrderBy: &pr.OrderBy,
	}
}

func PaginationRequestFromPb(req *common.Pagination) *PaginationRequest {
	return &PaginationRequest{
		Page:    int(req.Page),
		Limit:   int(req.Limit),
		Sort:    strings.TrimSpace(req.Sort),
		OrderBy: strings.TrimSpace(req.OrderBy),
	}
}

// ToPaginationParam converts proto pagination to domain param directly.
func ToPaginationParam(p *common.Pagination) *param.PaginationParam {
	if p == nil {
		return DefaultPagination()
	}
	page := int(p.Page)
	limit := int(p.Limit)
	sort := strings.TrimSpace(p.Sort)
	orderBy := strings.TrimSpace(p.OrderBy)
	return &param.PaginationParam{
		Page:    &page,
		Limit:   &limit,
		Sort:    &sort,
		OrderBy: &orderBy,
	}
}

// DefaultPagination returns default pagination parameters.
func DefaultPagination() *param.PaginationParam {
	page := 1
	limit := 10
	sort := "asc"
	orderBy := "created_at"
	return &param.PaginationParam{
		Page:    &page,
		Limit:   &limit,
		Sort:    &sort,
		OrderBy: &orderBy,
	}
}
