package virtual_security

import (
	"sync"
	"time"
)

// marginOrder - 信用注文
type marginOrder struct {
	Code               string                  // 注文コード
	OrderStatus        OrderStatus             // 状態
	TradeType          TradeType               // 取引区分
	Side               Side                    // 売買方向
	ExecutionCondition StockExecutionCondition // 株式執行条件
	SymbolCode         string                  // 銘柄コード
	OrderQuantity      float64                 // 注文数量 (複数のポジションを指定してエグジットする場合は、エグジットする合計数量)
	ContractedQuantity float64                 // 約定数量
	CanceledQuantity   float64                 // 取消数量
	LimitPrice         float64                 // 指値価格
	ExpiredAt          time.Time               // 有効期限
	StopCondition      *StockStopCondition     // 現物逆指値条件
	ExitPositionList   []ExitPosition          // エグジットポジションリスト
	OrderedAt          time.Time               // 注文日時
	CanceledAt         time.Time               // 取消日時
	Contracts          []*Contract             // 約定一覧
	ConfirmingCount    int                     // 約定確認回数
	Message            string                  // メッセージ
	mtx                sync.Mutex
}

// デバッグなどで必要になったときに使う
//func (o *marginOrder) String() string {
//	if b, err := json.Marshal(o); err != nil {
//		return err.Error()
//	} else {
//		return string(b)
//	}
//}

func (o *marginOrder) lock() {
	o.mtx.Lock()
}

func (o *marginOrder) unlock() {
	o.mtx.Unlock()
}

// isDied - 約定やキャンセルで終了した注文が一定時間経過して保持する必要がなくなっているかどうか
func (o *marginOrder) isDied(now time.Time) bool {
	// 未終了のステータスなら死んでいない
	if !o.OrderStatus.IsFixed() {
		return false
	}

	border := now.AddDate(0, 0, -1)

	// キャンセル日時があって、1日以上前なら死んだ注文
	if !o.CanceledAt.IsZero() && o.CanceledAt.Before(border) {
		return true
	}
	// 約定情報があって、最後の約定情報が1日以上前なら死んだ注文
	if o.Contracts != nil && len(o.Contracts) > 0 && o.Contracts[len(o.Contracts)-1].ContractedAt.Before(border) {
		return true
	}
	// キャンセル日時も約定情報もない終了している注文は死んだものとする
	if o.CanceledAt.IsZero() && (o.Contracts == nil || len(o.Contracts) < 1) {
		return true
	}
	return false
}

// executionCondition - 逆指値なども加味した、注文の執行条件
func (o *marginOrder) executionCondition() StockExecutionCondition {
	if o.ExecutionCondition.IsStop() && o.OrderStatus != OrderStatusWait && o.StopCondition != nil {
		return o.StopCondition.ExecutionConditionAfterHit
	}
	return o.ExecutionCondition
}

// limitPrice - 逆指値なども加味した、注文の指値価格
func (o *marginOrder) limitPrice() float64 {
	if o.ExecutionCondition.IsStop() && o.OrderStatus != OrderStatusWait && o.StopCondition != nil {
		return o.StopCondition.LimitPriceAfterHit
	}
	return o.LimitPrice
}

// activate - 未有効な注文を有効な注文に変える
//   逆指値のようなトリガーで発動する注文を想定
func (o *marginOrder) activate(price *symbolPrice, now time.Time) {
	if price == nil {
		return
	}

	o.mtx.Lock()
	defer o.mtx.Unlock()

	// 銘柄が同一でなければ何もしない
	if o.SymbolCode != price.SymbolCode {
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
		o.StopCondition.isActivate = true
		o.StopCondition.ActivatedAt = now
	}
}

// expired - 有効期限切れなら注文をキャンセル済みにする
func (o *marginOrder) expired(now time.Time) {
	o.mtx.Lock()
	defer o.mtx.Unlock()

	// 有効期限がゼロ値なら有効期限なしで何もしない
	if o.ExpiredAt.IsZero() {
		return
	}

	// 期限切れの注文なら状態を更新してfalse
	if now.After(o.ExpiredAt) {
		o.CanceledAt = now
		o.OrderStatus = OrderStatusCanceled
		o.Message = "expired"
	}
}

// contract - 約定情報を追加し、約定の進捗に合わせてステータスを更新する
func (o *marginOrder) contract(contract *Contract) {
	if contract == nil {
		return
	}

	o.mtx.Lock()
	defer o.mtx.Unlock()

	if o.Contracts == nil {
		o.Contracts = []*Contract{}
	}
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

// cancel - 注文を取消状態にする
func (o *marginOrder) cancel(canceledAt time.Time) {
	o.mtx.Lock()
	defer o.mtx.Unlock()

	if o.OrderStatus.IsCancelable() {
		o.CanceledAt = canceledAt
		o.OrderStatus = OrderStatusCanceled
	}
}
