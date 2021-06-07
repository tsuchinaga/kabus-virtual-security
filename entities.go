package virtual_security

import (
	"sync"
	"time"
)

// StockOrder - 現物注文
type StockOrder struct {
	Code               string                  // 注文コード
	OrderStatus        OrderStatus             // 状態
	Side               Side                    // 売買方向
	ExecutionCondition StockExecutionCondition // 株式執行条件
	SymbolCode         string                  // 銘柄コード
	Exchange           Exchange                // 市場
	OrderQuantity      float64                 // 注文数量
	ContractedQuantity float64                 // 約定数量
	CanceledQuantity   float64                 // 取消数量
	LimitPrice         float64                 // 指値価格
	ExpiredAt          time.Time               // 有効期限
	StopCondition      *StockStopCondition     // 現物逆指値条件
	OrderedAt          time.Time               // 注文日時
	CanceledAt         time.Time               // 取消日時
	Contracts          []*Contract             // 約定一覧
	ConfirmingCount    int
	mtx                sync.Mutex
}

func (o *StockOrder) isContractableTime(session Session) bool {
	return (o.executionCondition().IsContractableMorningSession() && session == SessionMorning) ||
		(o.executionCondition().IsContractableMorningSessionClosing() && session == SessionMorning) ||
		(o.executionCondition().IsContractableAfternoonSession() && session == SessionAfternoon) ||
		(o.executionCondition().IsContractableAfternoonSessionClosing() && session == SessionAfternoon)
}

// confirmContractItayoseMO - 板寄せ方式での成行注文の約定確認と約定した場合の結果
//   板寄せ方式では、5s以内の現値があれば現値で約定する
//   5s以内の現値がなくても、買い注文で売り気配値があれば売り気配値で約定する
//   5s以内の現値がなくても、売り注文で買い気配値があれば買い気配値で約定する
func (o *StockOrder) confirmContractItayoseMO(price SymbolPrice, now time.Time) *confirmContractResult {
	result := &confirmContractResult{isContracted: false}
	if price.Price > 0 && now.Add(-5*time.Second).Before(price.PriceTime) {
		result.isContracted = true
		result.price = price.Price
		result.contractedAt = now
	} else if o.Side == SideBuy && price.Bid > 0 {
		result.isContracted = true
		result.price = price.Bid
		result.contractedAt = now
	} else if o.Side == SideSell && price.Ask > 0 {
		result.isContracted = true
		result.price = price.Ask
		result.contractedAt = now
	}
	return result
}

// confirmContractAuctionMO - オークション方式での成行注文の約定確認と約定した場合の結果
//   買い注文で売り気配値があれば売り気配値で約定する
//   売り注文で買い気配値があれば買い気配値で約定する
func (o *StockOrder) confirmContractAuctionMO(price SymbolPrice, now time.Time) *confirmContractResult {
	result := &confirmContractResult{isContracted: false}
	if o.Side == SideBuy && price.Bid > 0 {
		result.isContracted = true
		result.price = price.Bid
		result.contractedAt = now
	} else if o.Side == SideSell && price.Ask > 0 {
		result.isContracted = true
		result.price = price.Ask
		result.contractedAt = now
	}
	return result
}

// confirmContractItayoseLO - 板寄せ方式での指値注文の約定確認と約定した場合の結果
//   板寄せ方式では、5s以内の現値があれば現値で約定確認を行なう
//   買い注文の場合、現値が指値価格以下なら約定する
//   売り注文の場合、現値が指値価格以上なら約定する
//   5s以内の現値がなくても、買い注文で売り気配値があり、指値価格より売り気配値が安ければ約定する
//   5s以内の現値がなくても、売り注文で買い気配値があり、指値価格より買い気配値が高ければ約定する
func (o *StockOrder) confirmContractItayoseLO(price SymbolPrice, now time.Time) *confirmContractResult {
	result := &confirmContractResult{isContracted: false}
	if price.Price > 0 && now.Add(-5*time.Second).Before(price.PriceTime) {
		if o.Side == SideBuy && o.limitPrice() >= price.Price {
			result.isContracted = true
			result.price = price.Price
			result.contractedAt = now
		} else if o.Side == SideSell && o.limitPrice() <= price.Price {
			result.isContracted = true
			result.price = price.Price
			result.contractedAt = now
		}
	} else {
		if o.Side == SideBuy && price.Bid > 0 && o.limitPrice() >= price.Bid {
			result.isContracted = true
			result.price = price.Bid
			result.contractedAt = now
		} else if o.Side == SideSell && price.Ask > 0 && o.limitPrice() <= price.Ask {
			result.isContracted = true
			result.price = price.Ask
			result.contractedAt = now
		}
	}
	return result
}

// confirmContractAuctionLO - オークション方式での指値注文の約定確認と約定した場合の結果
//   買い注文で売り気配値があり、指値価格より売り気配値が安ければ約定する
//   売り注文で買い気配値があり、指値価格より買い気配値が高ければ約定する
func (o *StockOrder) confirmContractAuctionLO(price SymbolPrice, now time.Time) *confirmContractResult {
	result := &confirmContractResult{isContracted: false}
	if o.Side == SideBuy && price.Bid > 0 && o.limitPrice() > price.Bid {
		result.isContracted = true
		result.price = o.limitPrice()
		result.contractedAt = now

		if o.ConfirmingCount == 0 {
			result.price = price.Bid
		}
	} else if o.Side == SideSell && price.Ask > 0 && o.limitPrice() < price.Ask {
		result.isContracted = true
		result.price = o.limitPrice()
		result.contractedAt = now

		if o.ConfirmingCount == 0 {
			result.price = price.Ask
		}
	}
	return result
}

func (o *StockOrder) executionCondition() StockExecutionCondition {
	if o.ExecutionCondition.IsStop() && o.OrderStatus != OrderStatusWait && o.StopCondition != nil {
		return o.StopCondition.ExecutionConditionAfterHit
	}
	return o.ExecutionCondition
}

func (o *StockOrder) limitPrice() float64 {
	if o.ExecutionCondition.IsStop() && o.OrderStatus != OrderStatusWait && o.StopCondition != nil {
		return o.StopCondition.LimitPriceAfterHit
	}
	return o.LimitPrice
}

func (o *StockOrder) confirmContract(price SymbolPrice, now time.Time, session Session) *confirmContractResult {
	o.mtx.Lock()
	defer o.mtx.Unlock()

	// TODO 有効期限切れの注文なら状態を更新してfalse

	// 銘柄・市場が同一でなければfalse
	if o.SymbolCode != price.SymbolCode || o.Exchange != price.Exchange {
		return &confirmContractResult{isContracted: false}
	}

	// 約定可能な注文状態でなければfalse
	if !o.OrderStatus.IsContractable() {
		return &confirmContractResult{isContracted: false}
	}

	// 約定可能時間でなければ約定しない
	if !o.isContractableTime(session) {
		return &confirmContractResult{isContracted: false}
	}

	// 執行条件ごとの約定チェック
	//   ここまできたということは、時間条件などはパスしているということ
	switch o.executionCondition() {
	case StockExecutionConditionMO: // 成行
		// 価格情報が寄りで現在値があれば現在値で約定、現在値がなくても気配値があれば気配値で約定
		// 価格情報が引けで現在値があれば現在値で約定、現在値がなくても気配値があれば気配値で約定
		// 価格情報がザラバなら気配値がある場合に限り気配値で約定
		switch price.kind {
		case PriceKindOpening, PriceKindClosing:
			return o.confirmContractItayoseMO(price, now)
		case PriceKindRegular:
			return o.confirmContractAuctionMO(price, now)
		}
	case StockExecutionConditionMOMO, StockExecutionConditionMOAO: // 寄成(前場), 寄成(後場)
		// 初回約定確認なら確認をし、初回でなければ何もしない
		//   初回じゃなければ寄りじゃないはず
		// 価格情報が寄りで現在値があれば現在値で約定、現在値がなくても気配値があれば気配値で約定
		if o.ConfirmingCount > 0 {
			return &confirmContractResult{isContracted: false}
		}

		if price.kind == PriceKindOpening {
			return o.confirmContractItayoseMO(price, now)
		}
	case StockExecutionConditionMOMC, StockExecutionConditionMOAC: // 引成(前場), 引成(後場)
		// 初回約定確認なら確認をし、初回でなければ何もしない
		//   初回じゃなければ引けじゃないはず
		// 価格情報が引けで現在値があれば現在値で約定、現在値がなくても気配値があれば気配値で約定
		if o.ConfirmingCount > 0 {
			return &confirmContractResult{isContracted: false}
		}

		if price.kind == PriceKindClosing {
			return o.confirmContractItayoseMO(price, now)
		}
	case StockExecutionConditionIOCMO: // IOC成行
		// 初回約定確認なら確認をし、初回でなければ何もしない
		//   初回じゃなければキャンセルされているはず
		// それ以外は通常の成行と同じ
		if o.ConfirmingCount > 0 {
			return &confirmContractResult{isContracted: false}
		}

		// 価格情報が寄りで現在値があれば現在値で約定、現在値がなくても気配値があれば気配値で約定
		// 価格情報が引けで現在値があれば現在値で約定、現在値がなくても気配値があれば気配値で約定
		// 価格情報がザラバなら気配値がある場合に限り気配値で約定
		switch price.kind {
		case PriceKindOpening, PriceKindClosing:
			return o.confirmContractItayoseMO(price, now)
		case PriceKindRegular:
			return o.confirmContractAuctionMO(price, now)
		}
	case StockExecutionConditionLO: // 指値
		// 価格情報が寄りで現在値があり現在値が約定条件を満たしていれば現在値で約定、現在値がなくても気配値があり気配値が約定条件を満たしていれば気配値で約定
		// 価格情報が引けで現在値があり現在値が約定条件を満たしていれば現在値で約定、現在値がなくても気配値があり気配値が約定条件を満たしていれば気配値で約定
		// 価格情報がザラバなら気配値があり気配値が約定条件を満たしていれば指値価格する。ただし、初回チェックの場合は気配値で約定する

		switch price.kind {
		case PriceKindOpening, PriceKindClosing:
			return o.confirmContractItayoseLO(price, now)
		case PriceKindRegular:
			return o.confirmContractAuctionLO(price, now)
		}
	case StockExecutionConditionLOMO, StockExecutionConditionLOAO: // 寄指(前場), 寄指(後場)
		// 初回約定確認なら確認をし、初回でなければ何もしない
		//   初回じゃなければ寄りじゃないはず
		// 価格情報が寄りで現在値があり現在値が約定条件を満たしていれば現在値で約定、現在値がなくても気配値があり気配値が約定条件を満たしていれば気配値で約定
		if o.ConfirmingCount > 0 {
			return &confirmContractResult{isContracted: false}
		}

		if price.kind == PriceKindOpening {
			return o.confirmContractItayoseLO(price, now)
		}
	case StockExecutionConditionLOMC, StockExecutionConditionLOAC: // 引指(前場), 引指(後場)
		// 初回約定確認なら確認をし、初回でなければ何もしない
		//   初回じゃなければ寄りじゃないはず
		// 価格情報が寄りで現在値があり現在値が約定条件を満たしていれば現在値で約定、現在値がなくても気配値があり気配値が約定条件を満たしていれば気配値で約定
		if o.ConfirmingCount > 0 {
			return &confirmContractResult{isContracted: false}
		}

		if price.kind == PriceKindClosing {
			return o.confirmContractItayoseLO(price, now)
		}
	case StockExecutionConditionIOCLO: // IOC指値
		// 初回約定確認なら確認をし、初回でなければ何もしない
		//   初回じゃなければキャンセルされているはず
		// それ以外は通常の指値と同じ
		if o.ConfirmingCount > 0 {
			return &confirmContractResult{isContracted: false}
		}

		switch price.kind {
		case PriceKindOpening, PriceKindClosing:
			return o.confirmContractItayoseLO(price, now)
		case PriceKindRegular:
			return o.confirmContractAuctionLO(price, now)
		}
	case StockExecutionConditionFunariM: // 不成(前場)
		// 前場の引けでは引成注文と同じ
		// 前場の引け以外は通常の指値と同じ
		if session == SessionMorning && price.kind == PriceKindClosing {
			return o.confirmContractItayoseMO(price, now)
		} else {
			switch price.kind {
			case PriceKindOpening, PriceKindClosing:
				return o.confirmContractItayoseLO(price, now)
			case PriceKindRegular:
				return o.confirmContractAuctionLO(price, now)
			}
		}
	case StockExecutionConditionFunariA: // 不成(後場)
		// 後場の引けでは引成注文と同じ
		// 後場の引け以外は通常の指値と同じ
		if session == SessionAfternoon && price.kind == PriceKindClosing {
			switch price.kind {
			case PriceKindOpening, PriceKindClosing:
				return o.confirmContractItayoseMO(price, now)
			case PriceKindRegular:
				return o.confirmContractAuctionMO(price, now)
			}
		} else {
			switch price.kind {
			case PriceKindOpening, PriceKindClosing:
				return o.confirmContractItayoseLO(price, now)
			case PriceKindRegular:
				return o.confirmContractAuctionLO(price, now)
			}
		}

		// 逆指値での約定確認はない
		//   逆指値発動後は他の執行条件に則った方法で約定する
	}

	o.ConfirmingCount++
	return &confirmContractResult{isContracted: false}
}

func (o *StockOrder) activate(price SymbolPrice, now time.Time) {
	o.mtx.Lock()
	defer o.mtx.Unlock()

	// 銘柄・市場が同一でなければ何もしない
	if o.SymbolCode != price.SymbolCode || o.Exchange != price.Exchange {
		return
	}

	// 待機注文でなければ何もしない
	if o.OrderStatus != OrderStatusWait {
		return
	}

	// 逆指値注文でなければ何もしない
	if !o.ExecutionCondition.IsStop() {
		return
	}

	// 逆指値条件が設定されていなければ何もしない
	if o.StopCondition == nil {
		return
	}

	// 現在値なし、もしくは現在値が5s以上前なら利用しない
	if price.Price < 1 || !now.Add(-5*time.Second).Before(price.PriceTime) {
		return
	}

	// 逆指値価格と現在値を比較した結果が条件を満たしていれば、注文状態に遷移させる
	if o.StopCondition.ComparisonOperator.CompareFloat64(o.StopCondition.StopPrice, price.Price) {
		o.OrderStatus = OrderStatusInOrder
		o.StopCondition.IsActivate = true
		o.StopCondition.ActivatedAt = now
	}
}

func (o *StockOrder) contract(contract *Contract) {
	if contract == nil {
		return
	}

	o.mtx.Lock()
	defer o.mtx.Unlock()

	o.Contracts = append(o.Contracts, contract)
	o.ContractedQuantity += contract.Quantity
	switch {
	case o.ContractedQuantity == 0:
		o.OrderStatus = OrderStatusInOrder
	case o.OrderQuantity > o.ContractedQuantity:
		o.OrderStatus = OrderStatusPart
	case o.OrderQuantity <= o.ContractedQuantity:
		o.OrderStatus = OrderStatusDone
	}
}

func (o *StockOrder) cancel(canceledAt time.Time) {
	o.mtx.Lock()
	defer o.mtx.Unlock()

	if o.OrderStatus.IsCancelable() {
		o.CanceledAt = canceledAt
		o.OrderStatus = OrderStatusCanceled
	}
}
