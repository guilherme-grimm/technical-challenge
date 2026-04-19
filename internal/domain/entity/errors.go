package entity

import "errors"

var (
	ErrDeviceNotFound         = errors.New("device not found")
	ErrVersionConflict        = errors.New("version conflict")
	ErrDeviceInUse            = errors.New("device in use")
	ErrDeviceInUseImmutable   = errors.New("name and brand cannot be changed while device is in use")
	ErrDuplicateID            = errors.New("duplicate id")
	ErrEmptyClient            = errors.New("empty database")
	ErrEmptyLogger            = errors.New("empty logger")
	ErrEmptyDeviceCreateInput = errors.New("empty device create input")
	ErrEmptyDeviceUpdateInput = errors.New("empty device update input")
	ErrEmptyDevicePatchInput  = errors.New("empty device patch input")
	ErrEmptyDeviceName        = errors.New("empty device name")
	ErrEmptyDeviceBrand       = errors.New("empty device brand")
	ErrEmptyDeviceID          = errors.New("empty device id")
	ErrEmptyListFilter        = errors.New("empty list filter")
	ErrInvalidDeviceState     = errors.New("invalid device state")
)
