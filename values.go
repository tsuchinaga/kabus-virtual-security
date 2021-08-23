package virtual_security

import (
	"time"
)

// RegisterPriceRequest - 銘柄価格のリクエスト
type RegisterPriceRequest struct {
	ExchangeType ExchangeType // 市場種別
	SymbolCode   string       // 銘柄コード
	Price        float64      // 価格
	PriceTime    time.Time    // 価格日時
	Bid          float64      // 買気配値
	BidTime      time.Time    // 買気配日時
	Ask          float64      // 売気配値
	AskTime      time.Time    // 売気配日時
}

// symbolPrice - 銘柄の価格
type symbolPrice struct {
	ExchangeType     ExchangeType // 市場種別
	SymbolCode       string       // 銘柄コード
	Price            float64      // 価格
	PriceTime        time.Time    // 価格日時
	Bid              float64      // 買気配値
	BidTime          time.Time    // 買気配日時
	Ask              float64      // 売気配値
	AskTime          time.Time    // 売気配日時
	kind             PriceKind    // 種別
	session          Session      // セッション
	priceBusinessDay time.Time    // 価格日時の営業日
}

func (e *symbolPrice) maxTime() time.Time {
	var maxTime time.Time
	if maxTime.Before(e.PriceTime) {
		maxTime = e.PriceTime
	}
	if maxTime.Before(e.AskTime) {
		maxTime = e.AskTime
	}
	if maxTime.Before(e.BidTime) {
		maxTime = e.BidTime
	}
	return maxTime
}

// UpdatedOrders - 更新された注文
type UpdatedOrders struct {
	Orders []OrderSummary // 更新された注文
}

// OrderSummary - 注文概要
type OrderSummary struct {
	OrderCode    string
	SymbolCode   string       // 銘柄コード
	ExchangeType ExchangeType // 市場種別
	Status       OrderStatus
}

// StockOrder - 現物注文
type StockOrder struct {
	Code               string                  // 注文コード
	OrderStatus        OrderStatus             // 状態
	Side               Side                    // 売買方向
	ExecutionCondition StockExecutionCondition // 株式執行条件
	SymbolCode         string                  // 銘柄コード
	OrderQuantity      float64                 // 注文数量
	ContractedQuantity float64                 // 約定数量
	CanceledQuantity   float64                 // 取消数量
	LimitPrice         float64                 // 指値価格
	ExpiredAt          time.Time               // 有効期限
	StopCondition      *StockStopCondition     // 現物逆指値条件
	OrderedAt          time.Time               // 注文日時
	CanceledAt         time.Time               // 取消日時
	Contracts          []*Contract             // 約定一覧
	Message            string                  // メッセージ
}

// StockOrderRequest - 現物注文リクエスト
type StockOrderRequest struct {
	Side               Side                    // 売買方向
	ExecutionCondition StockExecutionCondition // 株式執行条件
	SymbolCode         string                  // 銘柄コード
	Quantity           float64                 // 注文数量
	LimitPrice         float64                 // 指値価格
	ExpiredAt          time.Time               // 有効期限
	StopCondition      *StockStopCondition     // 現物逆指値条件
}

// OrderResult - 注文結果
type OrderResult struct {
	OrderCode string // 注文コード
}

// CancelOrderRequest - 注文の取り消しリクエスト
type CancelOrderRequest struct {
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

// StockPosition - 現物ポジション
type StockPosition struct {
	Code               string    // ポジションコード
	OrderCode          string    // 注文コード
	SymbolCode         string    // 銘柄コード
	Side               Side      // 売買方向
	ContractedQuantity float64   // 約定数量
	OwnedQuantity      float64   // 保有数量
	HoldQuantity       float64   // 拘束数量
	Price              float64   // 約定価格
	ContractedAt       time.Time // 約定日時
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
	isActivate                 bool
}

// MarginOrderRequest - 信用注文リクエスト
type MarginOrderRequest struct {
	TradeType          TradeType               // 取引区分
	Side               Side                    // 売買方向
	ExecutionCondition StockExecutionCondition // 株式執行条件
	SymbolCode         string                  // 銘柄コード
	Quantity           float64                 // 注文数量 (複数のポジションを指定してエグジットする場合は、合計数量を入れる)
	LimitPrice         float64                 // 指値価格
	ExpiredAt          time.Time               // 有効期限
	StopCondition      *StockStopCondition     // 現物逆指値条件
	ExitPositionList   []ExitPosition          // エグジットポジションリスト
}

type ExitPosition struct {
	PositionCode string  // ポジションコード
	Quantity     float64 // 数量
}

// MarginOrder - 信用注文
type MarginOrder struct {
	Code               string                  // 注文コード
	OrderStatus        OrderStatus             // 状態
	TradeType          TradeType               // 取引区分
	Side               Side                    // 売買方向
	ExecutionCondition StockExecutionCondition // 株式執行条件
	SymbolCode         string                  // 銘柄コード
	OrderQuantity      float64                 // 注文数量
	ContractedQuantity float64                 // 約定数量
	CanceledQuantity   float64                 // 取消数量
	LimitPrice         float64                 // 指値価格
	ExpiredAt          time.Time               // 有効期限
	StopCondition      *StockStopCondition     // 現物逆指値条件
	ExitPositionList   []ExitPosition          // エグジットポジションリスト
	OrderedAt          time.Time               // 注文日時
	CanceledAt         time.Time               // 取消日時
	Contracts          []*Contract             // 約定一覧
	Message            string                  // メッセージ
}

// MarginPosition - 信用ポジション
type MarginPosition struct {
	Code               string    // ポジションコード
	OrderCode          string    // 注文コード
	SymbolCode         string    // 銘柄コード
	Side               Side      // 売買方向
	ContractedQuantity float64   // 約定数量
	OwnedQuantity      float64   // 保有数量
	HoldQuantity       float64   // 拘束数量
	Price              float64   // 約定価格
	ContractedAt       time.Time // 約定日時
}
