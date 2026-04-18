package entity

import "errors"

var (
	ErrDeviceNotFound         = errors.New("device not found")
	ErrVersionConflict        = errors.New("version conflict")
	ErrDeviceInUse            = errors.New("device in use")
	ErrEmptyDatabase          = errors.New("empty database")
	ErrEmptyLogger            = errors.New("empty logger")
	ErrEmptyDeviceCreateInput = errors.New("empty device create input")
	ErrEmptyDeviceName        = errors.New("empty device name")
	ErrEmptyDeviceBrand       = errors.New("empty device brand")
	ErrInvalidDeviceState     = errors.New("invalid device state")
	ErrEmptyListFilter        = errors.New("empty list filter")
	ErrEmptyDeviceID          = errors.New("empty device id")
	ErrEmptyDeviceUpdateInput = errors.New("empty device update input")
	ErrEmptyDevicePatchInput  = errors.New("empty device patch input")
)
