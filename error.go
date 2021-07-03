package virtual_security

import "errors"

var (
	NilArgumentError               = errors.New("nil argument error")
	NoDataError                    = errors.New("no data error")
	ExpiredDataError               = errors.New("expired data error")
	NotEnoughOwnedQuantityError    = errors.New("not enough owned quantity error")
	NotEnoughHoldQuantityError     = errors.New("not enough hold quantity error")
	UncancellableOrderError        = errors.New("uncancellable order error")
	InvalidSideError               = errors.New("invalid side error")
	InvalidExecutionConditionError = errors.New("invalid execution condition error")
	InvalidSymbolCodeError         = errors.New("invalid symbol code error")
	InvalidQuantityError           = errors.New("invalid quantity error")
	InvalidLimitPriceError         = errors.New("invalid limit price error")
	InvalidExpiredError            = errors.New("invalid expired error")
	InvalidStopConditionError      = errors.New("invalid stop condition error")
	InvalidTimeError               = errors.New("invalid time error")
	InvalidExchangeTypeError       = errors.New("invalid exchange type error")
)
