package domain

import "errors"

var (
	ErrProjectNotFound      = errors.New("project not found")
	ErrProjectAlreadyExists = errors.New("project already exists")
	ErrMediumInvalid        = errors.New("medium is invalid")
	ErrSystemAlreadyExists  = errors.New("system already exists")
	ErrSystemNotFound       = errors.New("system not found")
	ErrShapeInvalid         = errors.New("shape is invalid")
	ErrGeometryInvalid      = errors.New("geometry is invalid")
	ErrLengthInvalid        = errors.New("length is invalid")
	ErrSegmentAlreadyExists = errors.New("segment already exists")
	ErrSegmentNotFound      = errors.New("segment not found")
	ErrFlowRateInvalid      = errors.New("flow rate is invalid")
	ErrUnknownShape         = errors.New("unknown shape")
	ErrAreaInvalid          = errors.New("area is invalid")
)
