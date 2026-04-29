package param

// PaginationParam represents the common pagination parameters.
type PaginationParam struct {
	Page    *int
	Limit   *int
	OrderBy *string
	Sort    *string
}
