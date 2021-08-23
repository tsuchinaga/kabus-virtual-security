package virtual_security

import (
	"sync"
	"time"
)

// marginPosition - 信用ポジション
type marginPosition struct {
	Code               string    // ポジションコード
	OrderCode          string    // 注文コード
	SymbolCode         string    // 銘柄コード
	Side               Side      // 方向
	ContractedQuantity float64   // 約定数量
	OwnedQuantity      float64   // 保有数量
	HoldQuantity       float64   // 拘束数量
	Price              float64   // 約定価格
	ContractedAt       time.Time // 約定日時
	mtx                sync.Mutex
}

// デバッグなどで必要になったときに使う
//func (p *marginPosition) String() string {
//	if b, err := json.Marshal(p); err != nil {
//		return err.Error()
//	} else {
//		return string(b)
//	}
//}

// exitable - 拘束されているポジションをエグジットできるかのチェック
func (p *marginPosition) exitable(quantity float64) error {
	if p.OwnedQuantity < quantity {
		return NotEnoughOwnedQuantityError
	}
	if p.HoldQuantity < quantity {
		return NotEnoughHoldQuantityError
	}
	return nil
}

// exit - 拘束されているポジションをエグジットする
func (p *marginPosition) exit(quantity float64) error {
	p.mtx.Lock()
	defer p.mtx.Unlock()

	if err := p.exitable(quantity); err != nil {
		return err
	}
	p.OwnedQuantity -= quantity
	p.HoldQuantity -= quantity
	return nil
}

// holdable - ポジションの保有数を拘束できるかのチェック
func (p *marginPosition) holdable(quantity float64) error {
	if p.OwnedQuantity < p.HoldQuantity+quantity {
		return NotEnoughOwnedQuantityError
	}
	return nil
}

// hold - ポジションの保有数を拘束する
func (p *marginPosition) hold(quantity float64) error {
	p.mtx.Lock()
	defer p.mtx.Unlock()

	if err := p.holdable(quantity); err != nil {
		return err
	}
	p.HoldQuantity += quantity
	return nil
}

// release - ポジションの拘束数を開放する
func (p *marginPosition) release(quantity float64) error {
	p.mtx.Lock()
	defer p.mtx.Unlock()
	if p.HoldQuantity-quantity < 0 {
		return NotEnoughHoldQuantityError
	}
	p.HoldQuantity -= quantity
	return nil
}

func (p *marginPosition) isDied() bool {
	return p.OwnedQuantity <= 0
}

func (p *marginPosition) orderableQuantity() float64 {
	return p.OwnedQuantity - p.HoldQuantity
}
