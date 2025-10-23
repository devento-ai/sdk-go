package devento

import "encoding/json"

// UpdateField represents a JSON field in a PATCH request that can be explicitly
// set to a value, set to null, or left untouched. When marshaled with the
// "omitempty" tag, an unset field is omitted from the payload while a null value
// is serialized as JSON null.
type UpdateField[T any] struct {
	value *T
	set   bool
}

// NewUpdateField returns an UpdateField whose value is set to the provided
// value.
func NewUpdateField[T any](value T) UpdateField[T] {
	return UpdateField[T]{
		value: &value,
		set:   true,
	}
}

// NullUpdateField returns an UpdateField that is explicitly set to JSON null.
func NullUpdateField[T any]() UpdateField[T] {
	return UpdateField[T]{
		value: nil,
		set:   true,
	}
}

// Unset marks the field so it will be omitted from the serialized payload.
func (f *UpdateField[T]) Unset() {
	f.value = nil
	f.set = false
}

// IsSet reports whether the field should be included in the serialized payload.
func (f UpdateField[T]) IsSet() bool {
	return f.set
}

// IsNull reports whether the field is explicitly set to JSON null.
func (f UpdateField[T]) IsNull() bool {
	return f.set && f.value == nil
}

// Value returns the underlying value and whether it has been set (and is not null).
func (f UpdateField[T]) Value() (T, bool) {
	if !f.set || f.value == nil {
		var zero T
		return zero, false
	}
	return *f.value, true
}

// MarshalJSON implements json.Marshaler, emitting either the stored value or
// null. Unset fields rely on the "omitempty" struct tag to stay omitted.
func (f UpdateField[T]) MarshalJSON() ([]byte, error) {
	if !f.set || f.value == nil {
		return []byte("null"), nil
	}
	return json.Marshal(*f.value)
}

// IsZero allows the json package to treat an unset field as empty when
// "omitempty" is used in struct tags.
func (f UpdateField[T]) IsZero() bool {
	return !f.set
}
