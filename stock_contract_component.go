package virtual_security

import "time"

func newStockContractComponent() iStockContractComponent {
	return &stockContractComponent{}
}

type iStockContractComponent interface {
	isContractableTime(executionCondition StockExecutionCondition, now time.Time) bool
	confirmMarginOrderContract(order *marginOrder, price *symbolPrice, now time.Time) *confirmContractResult
}

type stockContractComponent struct{}

// isContractableTime - 注文が約定できるタイミングにあるか
func (c *stockContractComponent) isContractableTime(executionCondition StockExecutionCondition, now time.Time) bool {
	return (executionCondition.IsContractableMorningSession() && contractableMorningSessionAuctionTime.between(now)) ||
		(executionCondition.IsContractableMorningSessionClosing() && contractableMorningSessionCloseTime.between(now)) ||
		(executionCondition.IsContractableAfternoonSession() && contractableAfternoonSessionAuctionTime.between(now)) ||
		(executionCondition.IsContractableAfternoonSessionClosing() && contractableAfternoonSessionCloseTime.between(now))
}

// confirmContractItayoseMO - 板寄せ方式での成行注文の約定確認と約定した場合の結果
//   板寄せ方式では、5s以内の現値があれば現値で約定する
//   5s以内の現値がなくても、買い注文で売り気配値があれば売り気配値で約定する
//   5s以内の現値がなくても、売り注文で買い気配値があれば買い気配値で約定する
func (c *stockContractComponent) confirmContractItayoseMO(side Side, price *symbolPrice, now time.Time) *confirmContractResult {
	result := &confirmContractResult{isContracted: false}
	if price == nil {
		return result
	}
	if price.Price > 0 && now.Add(-5*time.Second).Before(price.PriceTime) {
		result.isContracted = true
		result.price = price.Price
		result.contractedAt = now
	} else if side == SideBuy && price.Ask > 0 {
		result.isContracted = true
		result.price = price.Ask
		result.contractedAt = now
	} else if side == SideSell && price.Bid > 0 {
		result.isContracted = true
		result.price = price.Bid
		result.contractedAt = now
	}
	return result
}

// confirmContractAuctionMO - オークション方式での成行注文の約定確認と約定した場合の結果
//   買い注文で売り気配値があれば売り気配値で約定する
//   売り注文で買い気配値があれば買い気配値で約定する
func (c *stockContractComponent) confirmContractAuctionMO(side Side, price *symbolPrice, now time.Time) *confirmContractResult {
	result := &confirmContractResult{isContracted: false}
	if price == nil {
		return result
	}
	if side == SideBuy && price.Ask > 0 {
		result.isContracted = true
		result.price = price.Ask
		result.contractedAt = now
	} else if side == SideSell && price.Bid > 0 {
		result.isContracted = true
		result.price = price.Bid
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
func (c *stockContractComponent) confirmContractItayoseLO(side Side, limitPrice float64, price *symbolPrice, now time.Time) *confirmContractResult {
	result := &confirmContractResult{isContracted: false}
	if price == nil {
		return result
	}
	if price.Price > 0 && now.Add(-5*time.Second).Before(price.PriceTime) {
		if side == SideBuy && limitPrice >= price.Price {
			result.isContracted = true
			result.price = price.Price
			result.contractedAt = now
		} else if side == SideSell && limitPrice <= price.Price {
			result.isContracted = true
			result.price = price.Price
			result.contractedAt = now
		}
	} else {
		if side == SideBuy && price.Ask > 0 && limitPrice >= price.Ask {
			result.isContracted = true
			result.price = price.Ask
			result.contractedAt = now
		} else if side == SideSell && price.Bid > 0 && limitPrice <= price.Bid {
			result.isContracted = true
			result.price = price.Bid
			result.contractedAt = now
		}
	}
	return result
}

// confirmContractAuctionLO - オークション方式での指値注文の約定確認と約定した場合の結果
//   買い注文で売り気配値があり、指値価格より売り気配値が安ければ約定する
//   売り注文で買い気配値があり、指値価格より買い気配値が高ければ約定する
func (c *stockContractComponent) confirmContractAuctionLO(side Side, limitPrice float64, isConfirmed bool, price *symbolPrice, now time.Time) *confirmContractResult {
	result := &confirmContractResult{isContracted: false}
	if price == nil {
		return result
	}
	if side == SideBuy && price.Ask > 0 && limitPrice > price.Ask {
		result.isContracted = true
		result.price = limitPrice
		result.contractedAt = now

		// 初回確認なら板で約定できる
		if !isConfirmed {
			result.price = price.Ask
		}
	} else if side == SideSell && price.Bid > 0 && limitPrice < price.Bid {
		result.isContracted = true
		result.price = limitPrice
		result.contractedAt = now

		// 初回確認なら板で約定できる
		if !isConfirmed {
			result.price = price.Bid
		}
	}
	return result
}

// confirmOrderContract - 注文の要素と価格情報を受け取り、約定可能かのチェックし、約定したらどんな約定状態になるのかを返す
func (c *stockContractComponent) confirmOrderContract(executionCondition StockExecutionCondition, side Side, limitPrice float64, isConfirmed bool, price *symbolPrice, now time.Time) *confirmContractResult {
	// 価格情報がなければ約定しない, 約定可能時間帯じゃなければ約定しない
	if price == nil || !c.isContractableTime(executionCondition, now) {
		return &confirmContractResult{isContracted: false}
	}

	switch executionCondition {
	case StockExecutionConditionMO: // 成行
		// 価格情報が寄りで現在値があれば現在値で約定、現在値がなくても気配値があれば気配値で約定
		// 価格情報が引けで現在値があれば現在値で約定、現在値がなくても気配値があれば気配値で約定
		// 価格情報がザラバなら気配値がある場合に限り気配値で約定
		switch price.kind {
		case PriceKindOpening, PriceKindClosing:
			return c.confirmContractItayoseMO(side, price, now)
		case PriceKindRegular:
			return c.confirmContractAuctionMO(side, price, now)
		}
	case StockExecutionConditionMOMO, StockExecutionConditionMOAO: // 寄成(前場), 寄成(後場)
		// 初回約定確認なら確認をし、初回でなければ何もしない
		//   初回じゃなければ寄りじゃないはず
		// 価格情報が寄りで現在値があれば現在値で約定、現在値がなくても気配値があれば気配値で約定
		if isConfirmed {
			return &confirmContractResult{isContracted: false}
		}

		if price.kind == PriceKindOpening {
			return c.confirmContractItayoseMO(side, price, now)
		}
	case StockExecutionConditionMOMC, StockExecutionConditionMOAC: // 引成(前場), 引成(後場)
		// 初回約定確認なら確認をし、初回でなければ何もしない
		//   初回じゃなければ引けじゃないはず
		// 価格情報が引けで現在値があれば現在値で約定、現在値がなくても気配値があれば気配値で約定
		if isConfirmed {
			return &confirmContractResult{isContracted: false}
		}

		if price.kind == PriceKindClosing {
			return c.confirmContractItayoseMO(side, price, now)
		}
	case StockExecutionConditionIOCMO: // IOC成行
		// 初回約定確認なら確認をし、初回でなければ何もしない
		//   初回じゃなければキャンセルされているはず
		// それ以外は通常の成行と同じ
		if isConfirmed {
			return &confirmContractResult{isContracted: false}
		}

		// 価格情報が寄りで現在値があれば現在値で約定、現在値がなくても気配値があれば気配値で約定
		// 価格情報が引けで現在値があれば現在値で約定、現在値がなくても気配値があれば気配値で約定
		// 価格情報がザラバなら気配値がある場合に限り気配値で約定
		switch price.kind {
		case PriceKindOpening, PriceKindClosing:
			return c.confirmContractItayoseMO(side, price, now)
		case PriceKindRegular:
			return c.confirmContractAuctionMO(side, price, now)
		}
	case StockExecutionConditionLO: // 指値
		// 価格情報が寄りで現在値があり現在値が約定条件を満たしていれば現在値で約定、現在値がなくても気配値があり気配値が約定条件を満たしていれば気配値で約定
		// 価格情報が引けで現在値があり現在値が約定条件を満たしていれば現在値で約定、現在値がなくても気配値があり気配値が約定条件を満たしていれば気配値で約定
		// 価格情報がザラバなら気配値があり気配値が約定条件を満たしていれば指値価格する。ただし、初回チェックの場合は気配値で約定する

		switch price.kind {
		case PriceKindOpening, PriceKindClosing:
			return c.confirmContractItayoseLO(side, limitPrice, price, now)
		case PriceKindRegular:
			return c.confirmContractAuctionLO(side, limitPrice, isConfirmed, price, now)
		}
	case StockExecutionConditionLOMO, StockExecutionConditionLOAO: // 寄指(前場), 寄指(後場)
		// 初回約定確認なら確認をし、初回でなければ何もしない
		//   初回じゃなければ寄りじゃないはず
		// 価格情報が寄りで現在値があり現在値が約定条件を満たしていれば現在値で約定、現在値がなくても気配値があり気配値が約定条件を満たしていれば気配値で約定
		if isConfirmed {
			return &confirmContractResult{isContracted: false}
		}

		if price.kind == PriceKindOpening {
			return c.confirmContractItayoseLO(side, limitPrice, price, now)
		}
	case StockExecutionConditionLOMC, StockExecutionConditionLOAC: // 引指(前場), 引指(後場)
		// 初回約定確認なら確認をし、初回でなければ何もしない
		//   初回じゃなければ寄りじゃないはず
		// 価格情報が寄りで現在値があり現在値が約定条件を満たしていれば現在値で約定、現在値がなくても気配値があり気配値が約定条件を満たしていれば気配値で約定
		if isConfirmed {
			return &confirmContractResult{isContracted: false}
		}

		if price.kind == PriceKindClosing {
			return c.confirmContractItayoseLO(side, limitPrice, price, now)
		}
	case StockExecutionConditionIOCLO: // IOC指値
		// 初回約定確認なら確認をし、初回でなければ何もしない
		//   初回じゃなければキャンセルされているはず
		// それ以外は通常の指値と同じ
		if isConfirmed {
			return &confirmContractResult{isContracted: false}
		}

		switch price.kind {
		case PriceKindOpening, PriceKindClosing:
			return c.confirmContractItayoseLO(side, limitPrice, price, now)
		case PriceKindRegular:
			return c.confirmContractAuctionLO(side, limitPrice, isConfirmed, price, now)
		}
	case StockExecutionConditionFunariM: // 不成(前場)
		// 前場の引けでは引成注文と同じ
		// 前場の引け以外は通常の指値と同じ
		if contractableMorningSessionCloseTime.between(now) {
			return c.confirmContractItayoseMO(side, price, now)
		} else {
			switch price.kind {
			case PriceKindOpening, PriceKindClosing:
				return c.confirmContractItayoseLO(side, limitPrice, price, now)
			case PriceKindRegular:
				return c.confirmContractAuctionLO(side, limitPrice, isConfirmed, price, now)
			}
		}
	case StockExecutionConditionFunariA: // 不成(後場)
		// 後場の引けでは引成注文と同じ
		// 後場の引け以外は通常の指値と同じ
		if contractableAfternoonSessionCloseTime.between(now) {
			return c.confirmContractItayoseMO(side, price, now)
		} else {
			switch price.kind {
			case PriceKindOpening, PriceKindClosing:
				return c.confirmContractItayoseLO(side, limitPrice, price, now)
			case PriceKindRegular:
				return c.confirmContractAuctionLO(side, limitPrice, isConfirmed, price, now)
			}
		}

		// 逆指値での約定確認はない
		//   逆指値発動後は他の執行条件に則った方法で約定する
	}
	return &confirmContractResult{isContracted: false}
}

// confirmMarginOrderContract - 信用注文の約定確認し、約定したらどんな約定状態になるのかを返す
func (c *stockContractComponent) confirmMarginOrderContract(order *marginOrder, price *symbolPrice, now time.Time) *confirmContractResult {
	// 注文がnil, 価格情報がnil, 注文と価格情報の銘柄が一致しない, 注文が約定可能な状態じゃない のいずれかの場合、約定しない
	if order == nil || price == nil || order.SymbolCode != price.SymbolCode || !order.OrderStatus.IsContractable() {
		return &confirmContractResult{isContracted: false}
	}

	return c.confirmOrderContract(order.executionCondition(), order.Side, order.limitPrice(), order.ConfirmingCount > 0, price, now)
}
