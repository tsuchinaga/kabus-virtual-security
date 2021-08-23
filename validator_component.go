package virtual_security

import (
	"fmt"
	"time"
)

func newValidatorComponent() iValidatorComponent {
	return &validatorComponent{}
}

type iValidatorComponent interface {
	isValidMarginOrder(order *marginOrder, now time.Time, positions []*marginPosition) error
}

type validatorComponent struct{}

func (c *validatorComponent) isValidMarginOrder(order *marginOrder, now time.Time, positions []*marginPosition) error {
	if !order.TradeType.isValid() {
		return InvalidTradeTypeError
	}
	if !order.Side.isValid() {
		return InvalidSideError
	}
	if !order.ExecutionCondition.isValid() {
		return InvalidExecutionConditionError
	}
	if order.SymbolCode == "" {
		return InvalidSymbolCodeError
	}
	if order.OrderQuantity <= 0 {
		return InvalidQuantityError
	}
	if order.ExecutionCondition.IsLimitOrder() && order.LimitPrice <= 0 {
		return InvalidLimitPriceError
	}
	if order.ExpiredAt.Before(time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)) {
		return InvalidExpiredError
	}
	if order.ExecutionCondition.IsStop() && (order.StopCondition == nil ||
		order.StopCondition.StopPrice <= 0 ||
		order.StopCondition.ExecutionConditionAfterHit.IsStop() ||
		(order.StopCondition.ExecutionConditionAfterHit.IsLimitOrder() && order.StopCondition.LimitPriceAfterHit <= 0)) {
		return InvalidStopConditionError
	}

	// Exitでエグジットポジションが指定されていなければエラー
	if order.TradeType == TradeTypeExit && (order.ExitPositionList == nil || len(order.ExitPositionList) == 0) {
		return InvalidExitPositionError
	}

	// ExitPositionListで指定された数量とQuantityが一致していなければエラー
	var totalExitQuantity float64
	for _, p := range order.ExitPositionList {
		totalExitQuantity += p.Quantity
	}
	if order.TradeType == TradeTypeExit && order.OrderQuantity != totalExitQuantity {
		return InvalidExitQuantityError
	}

	// Exitで指定されたポジションがエグジットできなければエラー
	positionMap := map[string]*marginPosition{}
	for _, p := range positions {
		if p.OwnedQuantity > 0 {
			positionMap[p.Code] = p
		}
	}
	for _, e := range order.ExitPositionList {
		p, ok := positionMap[e.PositionCode]
		if !ok {
			return fmt.Errorf("position code: %s: %w", e.PositionCode, InvalidExitPositionCodeError)
		}
		if p.OwnedQuantity-p.HoldQuantity < e.Quantity {
			return fmt.Errorf("position code: %s, exitable quantity: %.2f: %w", e.PositionCode, p.OwnedQuantity-p.HoldQuantity, NotEnoughOwnedQuantityError)
		}
	}

	return nil
}
