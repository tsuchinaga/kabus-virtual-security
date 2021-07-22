package virtual_security

func newPriceService(clock iClock, priceStore iPriceStore) iPriceService {
	return &priceService{
		clock:      clock,
		priceStore: priceStore,
	}
}

type iPriceService interface {
	getBySymbolCode(code string) (*symbolPrice, error)
	set(price *symbolPrice) error
	validation(price RegisterPriceRequest) error
	toSymbolPrice(symbolPrice RegisterPriceRequest) (*symbolPrice, error)
}

type priceService struct {
	clock      iClock
	priceStore iPriceStore
}

func (s *priceService) getBySymbolCode(code string) (*symbolPrice, error) {
	return s.priceStore.getBySymbolCode(code)
}

func (s *priceService) set(price *symbolPrice) error {
	return s.priceStore.set(price)
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
	if price.PriceTime.IsZero() && price.BidTime.IsZero() && price.AskTime.IsZero() {
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
		Bid:              price.Bid,
		BidTime:          price.BidTime,
		Ask:              price.Ask,
		AskTime:          price.AskTime,
		session:          s.clock.getSession(price.ExchangeType, price.PriceTime),
		priceBusinessDay: s.clock.getBusinessDay(price.ExchangeType, price.PriceTime),
	}

	prevPrice, err := s.priceStore.getBySymbolCode(price.SymbolCode)
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
