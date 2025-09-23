package Categories

type CreateCategoryRequest struct {
	Name string `db:"name" json:"name" validate:"required,min=3,max=100"`
}

type UpdateCategoryRequest struct {
	Name string `db:"name" json:"name" validate:"required,min=3,max=100"`
	Id   int    `json:"id" validate:"required"`
}

type Category struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}
