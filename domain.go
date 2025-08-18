package odoorpc

// Domain represents an Odoo search domain.
// It is implemented as a slice of conditions, where each condition
// is itself a 3-element slice: [field, operator, value].
type Domain []any

// NewDomain creates a new empty Domain.
func NewDomain() Domain {
	return Domain{}
}

// Equals appends an equality condition to the domain.
func (d Domain) Equals(field string, value any) Domain {
	return append(d, []any{field, "=", value})
}

// In appends an "in" condition for a slice of int64 values.
func (d Domain) In(field string, values []int64) Domain {
	vals := make([]any, len(values))
	for i, v := range values {
		vals[i] = v
	}
	return append(d, []any{field, "in", vals})
}

// ChildOf appends a "child_of" condition.
func (d Domain) ChildOf(field string, value any) Domain {
	return append(d, []any{field, "child_of", value})
}
