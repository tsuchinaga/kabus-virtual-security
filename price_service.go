package virtual_security

type PriceService interface {
	getBySymbolCode(code string) (*symbolPrice, error)
	set(price *symbolPrice) error
	validation(price RegisterPriceRequest) error
	toSymbolPrice(symbolPrice RegisterPriceRequest) (*symbolPrice, error)
}

type priceService struct {
	clock      Clock
	priceStore PriceStore
}

func (s *priceService) getBySymbolCode(code string) (*symbolPrice, error) {
	return s.priceStore.GetBySymbolCode(code)
}

func (s *priceService) set(price *symbolPrice) error {
	return s.priceStore.Set(price)
}

func (s *priceService) validation(price RegisterPriceRequest) error {
	// ExchangeType が想定外ならエラー
	if price.ExchangeType == ExchangeTypeUnspecified {
		return InvalidExchangeTypeError
	}

	// SymbolCode が空文字ならエラー
	if price.SymbolCode == "" {
		return InvalidSymbolCodeError
	}

	// 時刻情報がなかったらエラー
	if price.PriceTime.IsZero() && price.AskTime.IsZero() && price.BidTime.IsZero() {
		return InvalidTimeError
	}

	return nil
}

func (s *priceService) toSymbolPrice(price RegisterPriceRequest) (*symbolPrice, error) {
	res := &symbolPrice{
		ExchangeType:     price.ExchangeType,
		SymbolCode:       price.SymbolCode,
		Price:            price.Price,
		PriceTime:        price.PriceTime,
		Ask:              price.Ask,
		AskTime:          price.AskTime,
		Bid:              price.Bid,
		BidTime:          price.BidTime,
		session:          s.clock.GetSession(price.ExchangeType, price.PriceTime),
		priceBusinessDay: s.clock.GetBusinessDay(price.ExchangeType, price.PriceTime),
	}

	prevPrice, err := s.priceStore.GetBySymbolCode(price.SymbolCode)
	if err != nil && err != NoDataError {
		return nil, err
	}

	kind := PriceKindUnspecified
	// 前回の価格情報がない、もしくはセッションが違えば始値
	if prevPrice == nil || !prevPrice.priceBusinessDay.Equal(res.priceBusinessDay) || prevPrice.session != res.session {
		kind = PriceKindOpening
	}

	switch res.ExchangeType {
	case ExchangeTypeStock, ExchangeTypeMargin:
		switch {
		case contractableMorningSessionCloseTime.between(res.PriceTime) || contractableAfternoonSessionCloseTime.between(res.PriceTime):
			// その日・そのセッションの引け後の価格は終値
			if kind == PriceKindOpening {
				kind = PriceKindOpeningAndClosing
			} else {
				kind = PriceKindClosing
			}

		case contractableStockPriceTime.between(res.PriceTime):
			if kind != PriceKindOpening {
				// その日・そのセッションのザラバ中の価格は通常値
				kind = PriceKindRegular
			}
		}
	case ExchangeTypeFuture:
		// その日・そのセッションの引け後の価格は終値
		// その日・そのセッションのザラバ中の価格は通常値
		// TODO 先物の種別
	}
	res.kind = kind

	return res, nil
}
