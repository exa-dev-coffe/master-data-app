package menu

type CreateMenuRequest struct {
	Name        string  `db:"name" json:"name" validate:"required,min=3,max=100"`
	Description string  `db:"description" json:"description" validate:"required,min=3"`
	Price       float64 `db:"price" json:"price" validate:"required,gt=0"`
	CategoryID  *int64  `db:"category_id" json:"categoryId"`
	Photo       string  `db:"photo" json:"photo" validate:"required,url"`
	IsAvailable *bool   `db:"is_available" json:"isAvailable"`
	CreatedBy   int64   `db:"created_by" json:"createdBy"`
}

type UpdateMenuRequest struct {
	Id          int     `json:"id" validate:"required"`
	Name        string  `db:"name" json:"name" validate:"required,min=3,max=100"`
	Description string  `db:"description" json:"description" validate:"required,min=3"`
	Price       float64 `db:"price" json:"price" validate:"required,gt=0"`
	CategoryID  *int64  `db:"category_id" json:"categoryId"`
	Photo       string  `db:"photo" json:"photo" validate:"required,url"`
	IsAvailable *bool   `db:"is_available" json:"isAvailable"`
	UpdatedBy   int64   `db:"updated_by" json:"updatedBy"`
}

type SetMenuCategoryRequest struct {
	Id         int64 `db:"id" json:"id" validate:"required"`
	CategoryId int64 `db:"category_id" json:"categoryId" validate:"required"`
	UpdatedBy  int64 `db:"updated_by" json:"updatedBy"`
}

type UpdateMenuAvailabilityRequest struct {
	Id          int   `json:"id" validate:"required"`
	IsAvailable *bool `json:"isAvailable"`
	UpdatedBy   int64 `json:"updatedBy"`
}

type DiscountDetail struct {
	PromotionID   int64   `json:"promotionId"`
	PromotionName string  `json:"promotionName"`
	DiscountType  string  `json:"discountType"`
	DiscountValue float64 `json:"discountValue"`
	Savings       float64 `json:"savings"`
}

type Menu struct {
	Id                 int64           `db:"id" json:"id"`
	Name               string          `db:"name" json:"name"`
	Rating             float64         `db:"rating" json:"rating"`
	Description        string          `db:"description" json:"description"`
	Photo              string          `db:"photo" json:"photo"`
	IsAvailable        bool            `db:"is_available" json:"isAvailable"`
	Price              float64         `db:"price" json:"price"`
	EffectivePrice     float64         `json:"effectivePrice"`
	Discount           *DiscountDetail `json:"discount,omitempty"`
	CategoryId         int64           `db:"category_id" json:"categoryId"`
	CategoryName       string          `db:"category_name" json:"categoryName"`
	PromoID            int64           `db:"promo_id" json:"-"`
	PromoName          string          `db:"promo_name" json:"-"`
	PromoDiscountType  string          `db:"promo_discount_type" json:"-"`
	PromoDiscountValue float64         `db:"promo_discount_value" json:"-"`
	PromoMaxDiscount   float64         `db:"promo_max_discount" json:"-"`
}

func CalculateDiscountHelper(price float64, promoID int64, promoName, discountType string, discountValue, maxDiscount float64) (float64, *DiscountDetail) {
	if promoID <= 0 {
		return price, nil
	}
	var savings float64
	if discountType == "PERCENTAGE" {
		savings = price * (discountValue / 100)
		if maxDiscount > 0 && savings > maxDiscount {
			savings = maxDiscount
		}
	} else {
		savings = discountValue
	}
	if savings > price {
		savings = price
	}
	return price - savings, &DiscountDetail{
		PromotionID:   promoID,
		PromotionName: promoName,
		DiscountType:  discountType,
		DiscountValue: discountValue,
		Savings:       savings,
	}
}

func (m *Menu) CalculateDiscount() {
	m.EffectivePrice, m.Discount = CalculateDiscountHelper(
		m.Price, m.PromoID, m.PromoName, m.PromoDiscountType, m.PromoDiscountValue, m.PromoMaxDiscount,
	)
}

type InternalAvailableMenuResponse struct {
	Id                 int64           `json:"id" db:"id"`
	Price              float64         `json:"price" db:"price"`
	EffectivePrice     float64         `json:"effectivePrice"`
	Discount           *DiscountDetail `json:"discount,omitempty"`
	IsAvailable        bool            `json:"isAvailable" db:"is_available"`
	Name               string          `json:"name" db:"name"`
	PromoID            int64           `db:"promo_id" json:"-"`
	PromoName          string          `db:"promo_name" json:"-"`
	PromoDiscountType  string          `db:"promo_discount_type" json:"-"`
	PromoDiscountValue float64         `db:"promo_discount_value" json:"-"`
	PromoMaxDiscount   float64         `db:"promo_max_discount" json:"-"`
}

func (m *InternalAvailableMenuResponse) CalculateDiscount() {
	m.EffectivePrice, m.Discount = CalculateDiscountHelper(
		m.Price, m.PromoID, m.PromoName, m.PromoDiscountType, m.PromoDiscountValue, m.PromoMaxDiscount,
	)
}

type InternalMenuResponse struct {
	Id                 int64           `json:"id" db:"id"`
	Name               string          `json:"name" db:"name"`
	Description        string          `json:"description" db:"description"`
	Photo              string          `json:"photo" db:"photo"`
	Price              float64         `json:"price" db:"price"`
	EffectivePrice     float64         `json:"effectivePrice"`
	Discount           *DiscountDetail `json:"discount,omitempty"`
	PromoID            int64           `db:"promo_id" json:"-"`
	PromoName          string          `db:"promo_name" json:"-"`
	PromoDiscountType  string          `db:"promo_discount_type" json:"-"`
	PromoDiscountValue float64         `db:"promo_discount_value" json:"-"`
	PromoMaxDiscount   float64         `db:"promo_max_discount" json:"-"`
}

func (m *InternalMenuResponse) CalculateDiscount() {
	m.EffectivePrice, m.Discount = CalculateDiscountHelper(
		m.Price, m.PromoID, m.PromoName, m.PromoDiscountType, m.PromoDiscountValue, m.PromoMaxDiscount,
	)
}

type UpdateRatingAndReviewCountRequest struct {
	Id        int     ` json:"id" validate:"required"`
	Rating    float64 ` json:"rating" validate:"required,gte=0,lte=5"`
	UpdatedBy int64   ` json:"updatedBy"`
}
