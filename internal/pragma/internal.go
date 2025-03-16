package pragma

// DoNotImplement can be embedded in an interface to prevent trivial
// implementations of the interface.
//
// This is useful to prevent unauthorized implementations of an interface.
type DoNotImplement interface{ Custom(DoNotImplement) }
