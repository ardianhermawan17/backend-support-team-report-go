package domain

import "errors"

var (
	ErrInvalidReportInput      = errors.New("invalid report input")
	ErrReportNotFound          = errors.New("report not found")
	ErrReportAlreadyExists     = errors.New("report already exists")
	ErrReportScheduleNotFound  = errors.New("report schedule not found")
	ErrReportTopScorerNotFound = errors.New("report top scorer not found")
)
