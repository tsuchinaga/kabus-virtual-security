package virtual_security

// ExchangeType - 市場種別
type ExchangeType string

const (
	ExchangeTypeUnspecified ExchangeType = ""       // 未指定
	ExchangeTypeStock       ExchangeType = "stock"  // 株式現物
	ExchangeTypeMargin      ExchangeType = "margin" // 株式信用
	ExchangeTypeFuture      ExchangeType = "future" // 先物
)

// Exchange - 市場
type Exchange string

const (
	ExchangeUnspecified        Exchange = ""                     // 未指定
	ExchangeToushou            Exchange = "toushou"              // 東証
	ExchangeMeishou            Exchange = "meishou"              // 名証
	ExchangeFukushou           Exchange = "fukushou"             // 福証
	ExchangeSatsushou          Exchange = "satsushou"            // 札証
	ExchangeFutureAllSession   Exchange = "future_all_session"   // 先物日通し
	ExchangeFutureDaySession   Exchange = "future_day_session"   // 先物日中場
	ExchangeFutureNightSession Exchange = "future_night_session" // 先物ナイトセッション
)

// OrderStatus - 注文状態
type OrderStatus string

const (
	OrderStatusUnspecified OrderStatus = ""          // 未指定
	OrderStatusNew         OrderStatus = "new"       // 新規
	OrderStatusWait        OrderStatus = "wait"      // 待機
	OrderStatusInOrder     OrderStatus = "in_order"  // 注文中
	OrderStatusPart        OrderStatus = "part"      // 部分約定
	OrderStatusDone        OrderStatus = "done"      // 全約定
	OrderStatusInCancel    OrderStatus = "in_cancel" // 取消中
	OrderStatusCanceled    OrderStatus = "canceled"  // 取消済み
)

func (e OrderStatus) IsContractable() bool {
	switch e {
	case OrderStatusInOrder, OrderStatusPart:
		return true
	}
	return false
}

func (e OrderStatus) IsFixed() bool {
	switch e {
	case OrderStatusDone, OrderStatusCanceled, OrderStatusUnspecified:
		return true
	}
	return false
}

func (e OrderStatus) IsCancelable() bool {
	switch e {
	case OrderStatusNew, OrderStatusWait, OrderStatusInOrder, OrderStatusPart:
		return true
	}
	return false
}

// StockExecutionCondition - 執行条件
type StockExecutionCondition string

const (
	StockExecutionConditionUnspecified StockExecutionCondition = ""                                  // 未指定
	StockExecutionConditionMO          StockExecutionCondition = "market_order"                      // 成行
	StockExecutionConditionMOMO        StockExecutionCondition = "market_order_on_morning_opening"   // 寄成(前場)
	StockExecutionConditionMOAO        StockExecutionCondition = "market_order_on_afternoon_opening" // 寄成(後場)
	StockExecutionConditionMOMC        StockExecutionCondition = "market_order_on_morning_closing"   // 引成(前場)
	StockExecutionConditionMOAC        StockExecutionCondition = "market_order_on_afternoon_closing" // 引成(後場)
	StockExecutionConditionIOCMO       StockExecutionCondition = "ioc_market_order"                  // IOC成行
	StockExecutionConditionLO          StockExecutionCondition = "limit_order"                       // 指値
	StockExecutionConditionLOMO        StockExecutionCondition = "limit_order_on_morning_opening"    // 寄指(前場)
	StockExecutionConditionLOAO        StockExecutionCondition = "limit_order_on_afternoon_opening"  // 寄指(後場)
	StockExecutionConditionLOMC        StockExecutionCondition = "limit_order_on_morning_closing"    // 引指(前場)
	StockExecutionConditionLOAC        StockExecutionCondition = "limit_order_on_afternoon_closing"  // 引指(後場)
	StockExecutionConditionIOCLO       StockExecutionCondition = "ioc_limit_order"                   // IOC指値
	StockExecutionConditionFunariM     StockExecutionCondition = "funari_on_morning"                 // 不成(前場)
	StockExecutionConditionFunariA     StockExecutionCondition = "funari_on_afternoon"               // 不成(後場)
	StockExecutionConditionStop        StockExecutionCondition = "stop"                              // 逆指値
)

func (e StockExecutionCondition) IsMarketOrder() bool {
	switch e {
	case StockExecutionConditionMO, // 成行
		StockExecutionConditionMOMO,  // 寄成(前場)
		StockExecutionConditionMOAO,  // 寄成(後場)
		StockExecutionConditionMOMC,  // 引成(前場)
		StockExecutionConditionMOAC,  // 引成(後場)
		StockExecutionConditionIOCMO: // IOC成行
		return true
	}
	return false
}

func (e StockExecutionCondition) IsLimitOrder() bool {
	switch e {
	case StockExecutionConditionLO, // 指値
		StockExecutionConditionLOMO,  // 寄指(前場)
		StockExecutionConditionLOAO,  // 寄指(後場)
		StockExecutionConditionLOMC,  // 引指(前場)
		StockExecutionConditionLOAC,  // 引指(後場)
		StockExecutionConditionIOCLO: // IOC指値
		return true
	}
	return false
}

func (e StockExecutionCondition) IsFunari() bool {
	switch e {
	case StockExecutionConditionFunariM, // 不成(前場)
		StockExecutionConditionFunariA: // 不成(後場)
		return true
	}
	return false
}

func (e StockExecutionCondition) IsStop() bool {
	switch e {
	case StockExecutionConditionStop: // 逆指値
		return true
	}
	return false
}

func (e StockExecutionCondition) IsContractableMorningSession() bool {
	switch e {
	case StockExecutionConditionMO,
		StockExecutionConditionMOMO,
		StockExecutionConditionIOCMO,
		StockExecutionConditionLO,
		StockExecutionConditionLOMO,
		StockExecutionConditionIOCLO,
		StockExecutionConditionFunariM,
		StockExecutionConditionFunariA,
		StockExecutionConditionStop:
		return true
	}
	return false
}

func (e StockExecutionCondition) IsContractableMorningSessionClosing() bool {
	switch e {
	case StockExecutionConditionMO,
		StockExecutionConditionMOMO,
		StockExecutionConditionMOMC,
		StockExecutionConditionIOCMO,
		StockExecutionConditionLO,
		StockExecutionConditionLOMO,
		StockExecutionConditionLOMC,
		StockExecutionConditionIOCLO,
		StockExecutionConditionFunariM,
		StockExecutionConditionFunariA,
		StockExecutionConditionStop:
		return true
	}
	return false
}

func (e StockExecutionCondition) IsContractableAfternoonSession() bool {
	switch e {
	case StockExecutionConditionMO,
		StockExecutionConditionMOAO,
		StockExecutionConditionIOCMO,
		StockExecutionConditionLO,
		StockExecutionConditionLOAO,
		StockExecutionConditionIOCLO,
		StockExecutionConditionFunariM,
		StockExecutionConditionFunariA,
		StockExecutionConditionStop:
		return true
	}
	return false
}

func (e StockExecutionCondition) IsContractableAfternoonSessionClosing() bool {
	switch e {
	case StockExecutionConditionMO,
		StockExecutionConditionMOAO,
		StockExecutionConditionMOAC,
		StockExecutionConditionIOCMO,
		StockExecutionConditionLO,
		StockExecutionConditionLOAO,
		StockExecutionConditionLOAC,
		StockExecutionConditionIOCLO,
		StockExecutionConditionFunariM,
		StockExecutionConditionFunariA,
		StockExecutionConditionStop:
		return true
	}
	return false
}

// Side - 売買方向
type Side string

const (
	SideUnspecified Side = ""     // 未指定
	SideBuy         Side = "buy"  // 買い
	SideSell        Side = "sell" // 売り
)

// Session - セッション
type Session string

const (
	SessionUnspecified Session = ""          // 未指定
	SessionMorning     Session = "morning"   // 前場
	SessionAfternoon   Session = "afternoon" // 後場
)

// Timing - タイミング
type Timing string

const (
	TimingUnspecified Timing = ""            // 未指定
	TimingPreOpening  Timing = "pre_opening" // プレオープニング
	TimingOpening     Timing = "opening"     // 寄り
	TimingRegular     Timing = "regular"     // ザラバ
	TimingClosing     Timing = "closing"     // 引け
)

// PriceKind - 価格種別
type PriceKind string

const (
	PriceKindUnspecified PriceKind = ""        // 未指定
	PriceKindOpening     PriceKind = "opening" // 寄り
	PriceKindRegular     PriceKind = "regular" // ザラバ
	PriceKindClosing     PriceKind = "closing" // 引け
)

// ComparisonOperator - 比較演算子
type ComparisonOperator string

const (
	ComparisonOperatorUnspecified                    = ""   // 未指定
	ComparisonOperatorGT          ComparisonOperator = "gt" // より大きい
	ComparisonOperatorGE          ComparisonOperator = "ge" // 以上
	ComparisonOperatorEQ          ComparisonOperator = "eq" // 等しい
	ComparisonOperatorLE          ComparisonOperator = "le" // 以下
	ComparisonOperatorLT          ComparisonOperator = "lt" // 未満
	ComparisonOperatorNE          ComparisonOperator = "ne" // 等しくない
)

func (e ComparisonOperator) CompareFloat64(a, b float64) bool {
	switch e {
	case ComparisonOperatorGT:
		return a > b
	case ComparisonOperatorGE:
		return a >= b
	case ComparisonOperatorEQ:
		return a == b
	case ComparisonOperatorLE:
		return a <= b
	case ComparisonOperatorLT:
		return a < b
	case ComparisonOperatorNE:
		return a != b
	}
	return false
}
