package domain

import "errors"

var (
	ErrReportNotFound        = errors.New("report not found")
	ErrReportAlreadyExists   = errors.New("a report already exists for this inspection")
	ErrReportNotGenerated    = errors.New("report PDF has not been generated yet")
	ErrReportFinalized       = errors.New("report is finalized and cannot be modified")
	ErrDeliveryNotFound      = errors.New("delivery not found")
)
