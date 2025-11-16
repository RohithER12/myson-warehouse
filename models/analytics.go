package models

type ProductAnalytics struct {
	TotalAmounts   TotalAmounts      `json:"total_amounts"`
	GodownData     GodownData        `json:"godown_data"`
	ProductsData   []ProductWiseData `json:"products_data"`
	TopTenProducts []ProductCount    `json:"top_ten_products"`
}

type ProductCount struct {
	ProductInfo ProductData `json:"product_info"`
	Stock       Stock       `json:"stock"`
}

type TotalAmounts struct {
	OnBoardingAmount  float64 `json:"on_boarding_amount"`
	OffBoardingAmount float64 `json:"off_boarding_amount"`
	InStockAmount     float64 `json:"in_stock_amount"`
	ProfitAmount      float64 `json:"profit_amount"`
	NetProfitAmount   float64 `json:"net_profit_amount"`
	ExpenseAmount     float64 `json:"expense_amount"`
}

type GodownData struct {
	GodownID            uint    `json:"godown_id"`
	GodownName          string  `json:"godown_name"`
	TotalSpace          float64 `json:"total_space"`
	AvailableSpace      float64 `json:"available_space"`
	UsedSpace           float64 `json:"used_space"`
	UsedSpacePercentage float64 `json:"used_space_percentage"`
}

type Stock struct {
	OnBoardCount  int `json:"on_board_count"`
	OffBoardCount int `json:"off_board_count"`
	InStockCount  int `json:"in_stock_count"`
}

type ProductWiseData struct {
	ProductInfo  ProductData         `json:"product_info"`
	Amounts      TotalProductAmounts `json:"amounts"`
	Stock        Stock               `json:"stock"`
	IsFastMoving bool                `json:"is_fast_moving"`
}
type ProductWiseAnalyticsData struct {
	ProductInfo  ProductData         `json:"product_info"`
	Amounts      TotalProductAmounts `json:"amounts"`
	Stock        Stock               `json:"stock"`
	IsFastMoving bool                `json:"is_fast_moving"`
}

type TotalProductAmounts struct {
	ProductOnBoardingAmount  float64 `json:"product_on_boarding_amount"`
	ProductOffBoardingAmount float64 `json:"product_off_boarding_amount"`
	ProductInStockAmount     float64 `json:"product_in_stock_amount"`
	ProductProfitAmount      float64 `json:"product_profit_amount"`
	ProductNetProfitAmount   float64 `json:"product_net_profit_amount"`
	ProductExpenseAmount     float64 `json:"product_expense_amount"`
}
