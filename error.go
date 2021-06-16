package virtual_security

import "errors"

var (
	NilArgumentError            = errors.New("nil argument error")
	NoDataError                 = errors.New("no data error")
	ExpiredDataError            = errors.New("expired data error")
	NotEnoughOwnedQuantityError = errors.New("not enough owned quantity error")
	NotEnoughHoldQuantityError  = errors.New("not enough hold quantity error")
	UncancellableOrderError     = errors.New("uncancellable order error")
)
