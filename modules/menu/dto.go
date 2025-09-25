package menu

type CreateMenuRequest struct {
	Name        string  `db:"name" json:"name" validate:"required,min=3,max=100"`
	Description string  `db:"description" json:"description" validate:"required,min=3"`
	Price       float64 `db:"price" json:"price" validate:"required,gt=0"`
	CategoryID  int64   `db:"category_id" json:"category_id"`
	Photo       string  `db:"photo" json:"photo" validate:"required,url"`
	CreatedBy   int64   `db:"created_by" json:"created_by"`
}

type UpdateMenuRequest struct {
	Id          int     `json:"id" validate:"required"`
	Name        string  `db:"name" json:"name" validate:"required,min=3,max=100"`
	Description string  `db:"description" json:"description" validate:"required,min=3"`
	Price       float64 `db:"price" json:"price" validate:"required,gt=0"`
	CategoryID  int64   `db:"category_id" json:"category_id"`
	Photo       string  `db:"photo" json:"photo" validate:"required,url"`
	IsAvailable bool    `db:"is_available" json:"is_available" validate:"required"`
	UpdatedBy   int64   `db:"updated_by" json:"updated_by"`
}

type Menu struct {
	Id           int64   `db:"id" json:"id"`
	Name         string  `db:"name" json:"name"`
	Description  string  `db:"description" json:"description"`
	Price        float64 `db:"price" json:"price"`
	CategoryId   int64   `db:"category_id" json:"category_id"`
	CategoryName string  `db:"category_name" json:"category_name"`
}
