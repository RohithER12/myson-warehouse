package models

type ProductAnalytics struct {
	TotalAmounts TotalAmounts      `json:"total_amounts"`
	GodownData   GodownData        `json:"godown_data"`
	ProductsData []ProductWiseData `json:"products_data"`
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
	Amounts      TotalAmounts `json:"amounts"`
	Stock        Stock        `json:"stock"`
	IsFastMoving bool         `json:"is_fast_moving"`
}
