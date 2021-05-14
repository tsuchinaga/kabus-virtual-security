package virtual_security

type Security interface {
	RegisterPrice(symbolPrice SymbolPrice) (*UpdatedOrders, error) // 銘柄価格の登録
	StockOrder(order *StockOrderRequest) (*OrderResult, error)     // 現物注文
	CancelOrder(cancelOrder *CancelOrder) (*OrderResult, error)    // 注文の取り消し
	StockOrders() ([]StockOrder, error)                            // 現物注文一覧
	StockPositions() ([]StockPosition, error)                      // ポジション一覧
}

type security struct {
}
