package odoorpc

// Domain represents an Odoo search domain.
// It is implemented as a slice of conditions, where each condition
// is itself a 3-element slice: [field, operator, value].
type Domain []any

// NewDomain creates a new empty Domain.
func NewDomain() Domain {
	return Domain{}
}

func isLogicalOperator(v any) bool {
	s, ok := v.(string)
	if !ok {
		return false
	}
	return s == "&" || s == "|" || s == "!"
}

func (d Domain) expression() ([]any, bool) {
	if len(d) == 0 {
		return nil, false
	}
	if isLogicalOperator(d[0]) {
		return []any(d), true
	}
	firstClause, ok := d[0].([]any)
	if !ok {
		return nil, false
	}
	if len(d) == 1 {
		return firstClause, true
	}
	expr := firstClause
	for _, item := range d[1:] {
		clause, ok := item.([]any)
		if !ok {
			continue
		}
		expr = []any{"&", expr, clause}
	}
	return expr, true
}

func domainFromExpr(expr []any) Domain {
	if len(expr) == 0 {
		return Domain{}
	}
	if isLogicalOperator(expr[0]) {
		return Domain(expr)
	}
	return Domain{expr}
}

// Equals appends an equality condition to the domain.
func (d Domain) Equals(field string, value any) Domain {
	return append(d, []any{field, "=", value})
}

// NotEquals appends an not equality condition to the domain.
func (d Domain) NotEquals(field string, value any) Domain {
	return append(d, []any{field, "!=", value})
}

// LessThan appends a "<" condition to the domain.
func (d Domain) LessThan(field string, value any) Domain {
	return append(d, []any{field, "<", value})
}

// GreaterThan appends a ">" condition to the domain.
func (d Domain) GreaterThan(field string, value any) Domain {
	return append(d, []any{field, ">", value})
}

// LessThanOrEqual appends a "<=" condition to the domain.
func (d Domain) LessThanOrEqual(field string, value any) Domain {
	return append(d, []any{field, "<=", value})
}

// GreaterThanOrEqual appends a ">=" condition to the domain.
func (d Domain) GreaterThanOrEqual(field string, value any) Domain {
	return append(d, []any{field, ">=", value})
}

// Like appends an match condition %value% to the domain.
func (d Domain) Like(field string, value string) Domain {
	return append(d, []any{field, "like", value})
}

// Like appends an non-casesensitive match condition %value% to the domain.
func (d Domain) Ilike(field string, value string) Domain {
	return append(d, []any{field, "ilike", value})
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

// And combines the current domain with the provided ones using logical AND.
func (d Domain) And(domains ...Domain) Domain {
	if len(domains) == 0 {
		return d
	}
	var exprs [][]any
	if expr, ok := d.expression(); ok {
		exprs = append(exprs, expr)
	}
	for _, other := range domains {
		if expr, ok := other.expression(); ok {
			exprs = append(exprs, expr)
		}
	}
	if len(exprs) == 0 {
		return Domain{}
	}
	result := exprs[0]
	for _, expr := range exprs[1:] {
		result = []any{"&", result, expr}
	}
	return domainFromExpr(result)
}

// Or combines the current domain with the provided ones using logical OR.
func (d Domain) Or(domains ...Domain) Domain {
	if len(domains) == 0 {
		return d
	}
	var exprs [][]any
	if expr, ok := d.expression(); ok {
		exprs = append(exprs, expr)
	}
	for _, other := range domains {
		if expr, ok := other.expression(); ok {
			exprs = append(exprs, expr)
		}
	}
	if len(exprs) == 0 {
		return Domain{}
	}
	result := exprs[0]
	for _, expr := range exprs[1:] {
		result = []any{"|", result, expr}
	}
	return domainFromExpr(result)
}
