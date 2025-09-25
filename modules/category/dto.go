package category

type CreateCategoryRequest struct {
	Name      string `db:"name" json:"name" validate:"required,min=3,max=100"`
	CreatedBy int64  `db:"created_by" json:"created_by"`
}

type Category struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}
