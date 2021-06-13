package virtual_security

import "errors"

var (
	NoDataError            = errors.New("no data error")
	ExpiredDataError       = errors.New("expired data error")
	NotEnoughOwnedQuantity = errors.New("not enough owned quantity")
	NotEnoughHoldQuantity  = errors.New("not enough hold quantity")
)
