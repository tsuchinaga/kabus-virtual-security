package virtual_security

import "time"

// SymbolPrice - 銘柄の価格
type SymbolPrice struct {
	SymbolCode string    // 銘柄コード
	Exchange   Exchange  // 市場種別
	Price      float64   // 価格
	PriceTime  time.Time // 価格日時
	Ask        float64   // 買気配値
	AskTime    time.Time // 買気配日時
	Bid        float64   // 売気配値
	BidTime    time.Time // 売気配日時
	kind       PriceKind // 種別
}

func (e *SymbolPrice) MaxTime() time.Time {
	var maxTime time.Time
	if maxTime.Before(e.PriceTime) {
		maxTime = e.PriceTime
	}
	if maxTime.Before(e.BidTime) {
		maxTime = e.BidTime
	}
	if maxTime.Before(e.AskTime) {
		maxTime = e.AskTime
	}
	return maxTime
}

// SessionInfo - セッション情報
type SessionInfo struct {
	Session Session // セッション
	Timing  Timing  // タイミング
}

// UpdatedOrders - 更新された注文
type UpdatedOrders struct {
	SymbolPrice *SymbolPrice   // 登録された銘柄の価格
	Orders      []OrderSummary // 更新された注文
}

// OrderSummary - 注文概要
type OrderSummary struct {
	OrderCode    string
	SymbolCode   string       // 銘柄コード
	ExchangeType ExchangeType // 市場種別
	Status       OrderStatus
}

// StockOrderRequest - 現物注文
type StockOrderRequest struct {
	// TODO 要素
}

// OrderResult - 注文結果
type OrderResult struct {
	OrderCode string // 注文コード
}

// CancelOrder - 注文の取り消し
type CancelOrder struct {
	OrderCode string // 取消対象の注文コード
}

// Contract - 約定
type Contract struct {
	ContractCode string
	OrderCode    string
	PositionCode string
	Price        float64
	Quantity     float64
	ContractedAt time.Time
}

// StockPosition - ポジション
type StockPosition struct {
	Code          string   // ポジションコード
	OrderCode     string   // 注文コード
	SymbolCode    string   // 銘柄コード
	Exchange      Exchange // 市場
	Side          Side     // 売買方向
	OwnedQuantity float64  // 保有数量
	HoldQuantity  float64  // 拘束数量
	ContractedAt  time.Time
}

// confirmContractResult - 約定可能かの結果
type confirmContractResult struct {
	isContracted bool
	price        float64
	contractedAt time.Time
}

// StockStopCondition - 逆指値条件
type StockStopCondition struct {
	StopPrice                  float64                 // 逆指値発動価格
	ComparisonOperator         ComparisonOperator      // 比較方法
	ExecutionConditionAfterHit StockExecutionCondition // 逆指値発動後注文条件
	LimitPriceAfterHit         float64                 // 逆指値発動後指値価格
	ActivatedAt                time.Time               // 逆指値条件が満たされた日時
	IsActivate                 bool
}
