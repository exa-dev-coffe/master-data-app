package promotion

type Promotion struct {
	ID            int64    `json:"id" db:"id"`
	Name          string   `json:"name" db:"name"`
	TargetType    string   `json:"targetType" db:"target_type"` // PRODUCT, CATEGORY, ALL
	TargetID      *int64   `json:"targetId" db:"target_id"`
	TargetName    *string  `json:"targetName,omitempty" db:"target_name"`
	DiscountType  string   `json:"discountType" db:"discount_type"` // PERCENTAGE, FIXED
	DiscountValue float64  `json:"discountValue" db:"discount_value"`
	MaxDiscount   float64  `json:"maxDiscount" db:"max_discount"`
	MinPurchase   float64  `json:"minPurchase" db:"min_purchase"`
	StartAt       string   `json:"startAt" db:"start_at"`
	EndAt         string   `json:"endAt" db:"end_at"`
	IsActive      bool     `json:"isActive" db:"is_active"`
	CreatedAt     string   `json:"createdAt" db:"created_at"`
	CreatedBy     *int64   `json:"createdBy" db:"created_by"`
	UpdatedAt     string   `json:"updatedAt" db:"updated_at"`
	UpdatedBy     *int64   `json:"updatedBy" db:"updated_by"`
	DeletedAt     *string  `json:"deletedAt,omitempty" db:"deleted_at"`
}

type CreatePromotionRequest struct {
	Name          string  `json:"name" validate:"required,min=3,max=100"`
	TargetType    string  `json:"targetType" validate:"required,oneof=PRODUCT CATEGORY ALL"`
	TargetID      *int64  `json:"targetId"`
	DiscountType  string  `json:"discountType" validate:"required,oneof=PERCENTAGE FIXED"`
	DiscountValue float64 `json:"discountValue" validate:"required,gt=0"`
	MaxDiscount   float64 `json:"maxDiscount"`
	MinPurchase   float64 `json:"minPurchase"`
	StartAt       string  `json:"startAt" validate:"required"`
	EndAt         string  `json:"endAt" validate:"required"`
	CreatedBy     int64   `json:"createdBy"`
}

type UpdatePromotionRequest struct {
	ID            int64   `json:"id" validate:"required"`
	Name          string  `json:"name" validate:"required,min=3,max=100"`
	TargetType    string  `json:"targetType" validate:"required,oneof=PRODUCT CATEGORY ALL"`
	TargetID      *int64  `json:"targetId"`
	DiscountType  string  `json:"discountType" validate:"required,oneof=PERCENTAGE FIXED"`
	DiscountValue float64 `json:"discountValue" validate:"required,gt=0"`
	MaxDiscount   float64 `json:"maxDiscount"`
	MinPurchase   float64 `json:"minPurchase"`
	StartAt       string  `json:"startAt" validate:"required"`
	EndAt         string  `json:"endAt" validate:"required"`
	UpdatedBy     int64   `json:"updatedBy"`
}

type UpdatePromotionStatusRequest struct {
	IsActive bool `json:"isActive"`
}

type MenuDiscountInfo struct {
	PromotionID   int64   `json:"promotionId" db:"promotion_id"`
	PromotionName string  `json:"promotionName" db:"promotion_name"`
	DiscountType  string  `json:"discountType" db:"discount_type"`
	DiscountValue float64 `json:"discountValue" db:"discount_value"`
	MaxDiscount   float64 `json:"maxDiscount" db:"max_discount"`
	Savings       float64 `json:"savings" db:"savings"`
}
