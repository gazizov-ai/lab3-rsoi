package service

import "errors"

var ErrHotelNotFound = errors.New("hotel not found")
var ErrServiceUnavailable = errors.New("service unavailable")
