package errors

import "errors"

var (
	ErrUnknownAbiFieldType         = errors.New("unknown abi field type")
	ErrUnknownGoFieldType          = errors.New("unknown go field type")
	ErrUnnamedInput                = errors.New("no name input")
	ErrMixedNamedAndUnnamedOutputs = errors.New("mixed named and unnamed outputs")
	ErrNotImplemented              = errors.New("not implemented")
	ErrNotMultiversX               = errors.New("not a MultiversX ABI")
)
