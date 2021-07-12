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
	Side               Side      // 方向
	ContractedQuantity float64   // 約定数量
	OwnedQuantity      float64   // 保有数量
	HoldQuantity       float64   // 拘束数量
	Price              float64   // 約定価格
	ContractedAt       time.Time // 約定日時
	mtx                sync.Mutex
}

// デバッグなどで必要になったときに使う
//func (p *stockPosition) String() string {
//	if b, err := json.Marshal(p); err != nil {
//		return err.Error()
//	} else {
//		return string(b)
//	}
//}

// exit - 拘束されているポジションをエグジットする
func (p *stockPosition) exit(quantity float64) error {
	p.mtx.Lock()
	defer p.mtx.Unlock()

	if p.OwnedQuantity < quantity {
		return NotEnoughOwnedQuantityError
	}
	if p.HoldQuantity < quantity {
		return NotEnoughHoldQuantityError
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
		return NotEnoughOwnedQuantityError
	}
	p.HoldQuantity += quantity
	return nil
}

// release - ポジションの拘束数を開放する
func (p *stockPosition) release(quantity float64) error {
	p.mtx.Lock()
	defer p.mtx.Unlock()
	if p.HoldQuantity-quantity < 0 {
		return NotEnoughHoldQuantityError
	}
	p.HoldQuantity -= quantity
	return nil
}

func (p *stockPosition) isDied() bool {
	return p.OwnedQuantity <= 0
}
