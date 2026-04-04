package models

import "errors"

var (
	ErrNotFound           = errors.New("not found")
	ErrAlreadyExists      = errors.New("already exists")
	ErrSlotInPast         = errors.New("slot is in the past")
	ErrSlotAlreadyBooked  = errors.New("slot already booked")
	ErrForbidden          = errors.New("forbidden")
	ErrInvalidInput       = errors.New("invalid input")
	ErrScheduleExists     = errors.New("schedule already exists for this room")
	ErrNoSchedule         = errors.New("room has no schedule")
	ErrUserNotFound       = errors.New("user not found")
	ErrInvalidCredentials = errors.New("invalid credentials")
)
