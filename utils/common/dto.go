package common

type ParamsListRequest struct {
	Search     Search // field, value
	Sort       Sort   // field, order
	Size       int
	Page       int
	NoPaginate bool
}
type Search struct {
	Field string
	Value string
}

type Sort struct {
	Field string
	Order string
}

type DeleteRequest struct {
	Id int `query:"id" validate:"required"`
}

type GetByIDRequest struct {
	Id int `json:"id" validate:"required"`
}
