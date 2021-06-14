package virtual_security

import (
	"sync"
	"time"
)

// stockPosition - 現物ポジション
type stockPosition struct {
	Code               string    // ポジションコード
	OrderCode          string    // 注文コード
	SymbolCode         string    // 銘柄コード
	Exchange           Exchange  // 市場
	Side               Side      // 売買方向
	ContractedQuantity float64   // 約定数量
	OwnedQuantity      float64   // 保有数量
	HoldQuantity       float64   // 拘束数量
	ContractedAt       time.Time // 約定日時
	mtx                sync.Mutex
}

// exit - 拘束されているポジションをエグジットする
func (p *stockPosition) exit(quantity float64) error {
	p.mtx.Lock()
	defer p.mtx.Unlock()

	if p.OwnedQuantity < quantity {
		return NotEnoughOwnedQuantity
	}
	if p.HoldQuantity < quantity {
		return NotEnoughHoldQuantity
	}
	p.OwnedQuantity -= quantity
	p.HoldQuantity -= quantity
	return nil
}

// hold - ポジションの保有数を拘束する
func (p *stockPosition) hold(quantity float64) error {
	p.mtx.Lock()
	defer p.mtx.Unlock()

	if p.OwnedQuantity < p.HoldQuantity+quantity {
		return NotEnoughOwnedQuantity
	}
	p.HoldQuantity += quantity
	return nil
}

// release - ポジションの拘束数を開放する
func (p *stockPosition) release(quantity float64) error {
	p.mtx.Lock()
	defer p.mtx.Unlock()
	if p.HoldQuantity-quantity < 0 {
		return NotEnoughHoldQuantity
	}
	p.HoldQuantity -= quantity
	return nil
}

func (p *stockPosition) isDied() bool {
	return p.OwnedQuantity <= 0
}
