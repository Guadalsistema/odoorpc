package odoorpc

import (
	"fmt"
	"strconv"
	"strings"
	"text/scanner"
)

// ParseDomain parses a Python-style domain string into a Domain.
// Example input: "[('name', '=', 'test'), ('active', '=', True)]"
func ParseDomain(s string) (Domain, error) {
	p := &domainParser{}
	p.scanner.Init(strings.NewReader(s))
	// Don't scan chars (single quotes) - we handle Python-style strings manually
	p.scanner.Mode = scanner.ScanIdents | scanner.ScanInts | scanner.ScanFloats | scanner.ScanStrings
	p.scanner.Error = func(s *scanner.Scanner, msg string) {
		p.err = fmt.Errorf("scan error at %s: %s", s.Position, msg)
	}
	p.next()
	return p.parse()
}

type domainParser struct {
	scanner scanner.Scanner
	tok     rune
	lit     string
	err     error
}

func (p *domainParser) next() {
	p.tok = p.scanner.Scan()
	p.lit = p.scanner.TokenText()
}

func (p *domainParser) parse() (Domain, error) {
	if p.err != nil {
		return nil, p.err
	}
	if p.tok == scanner.EOF {
		return Domain{}, nil
	}
	if p.tok != '[' {
		return nil, fmt.Errorf("domain must start with '[', got %q at %s", p.lit, p.scanner.Position)
	}
	items, err := p.parseList()
	if err != nil {
		return nil, err
	}
	if p.tok != scanner.EOF {
		return nil, fmt.Errorf("unexpected token %q after domain at %s", p.lit, p.scanner.Position)
	}
	return Domain(items), nil
}

func (p *domainParser) parseList() ([]any, error) {
	if p.tok != '[' {
		return nil, fmt.Errorf("expected '[', got %q at %s", p.lit, p.scanner.Position)
	}
	p.next()

	var items []any
	for p.tok != ']' && p.tok != scanner.EOF {
		item, err := p.parseValue()
		if err != nil {
			return nil, err
		}
		items = append(items, item)
		if p.tok == ',' {
			p.next()
		}
	}
	if p.tok != ']' {
		return nil, fmt.Errorf("expected ']' at %s", p.scanner.Position)
	}
	p.next()
	return items, nil
}

func (p *domainParser) parseTuple() ([]any, error) {
	if p.tok != '(' {
		return nil, fmt.Errorf("expected '(', got %q at %s", p.lit, p.scanner.Position)
	}
	p.next()

	var items []any
	for p.tok != ')' && p.tok != scanner.EOF {
		item, err := p.parseValue()
		if err != nil {
			return nil, err
		}
		items = append(items, item)
		if p.tok == ',' {
			p.next()
		}
	}
	if p.tok != ')' {
		return nil, fmt.Errorf("expected ')' at %s", p.scanner.Position)
	}
	p.next()
	return items, nil
}

func (p *domainParser) parseValue() (any, error) {
	if p.err != nil {
		return nil, p.err
	}

	switch p.tok {
	case '(':
		return p.parseTuple()
	case '[':
		return p.parseList()
	case '\'':
		return p.parseSingleQuotedString()
	case scanner.Int:
		i, err := strconv.ParseInt(p.lit, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid integer %q: %w", p.lit, err)
		}
		p.next()
		return i, nil
	case scanner.Float:
		f, err := strconv.ParseFloat(p.lit, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid float %q: %w", p.lit, err)
		}
		p.next()
		return f, nil
	case scanner.String, scanner.RawString, scanner.Char:
		s, err := strconv.Unquote(p.lit)
		if err != nil {
			// Handle single-quoted strings (Python style)
			if len(p.lit) >= 2 && p.lit[0] == '\'' && p.lit[len(p.lit)-1] == '\'' {
				s = p.lit[1 : len(p.lit)-1]
			} else {
				return nil, fmt.Errorf("invalid string %q: %w", p.lit, err)
			}
		}
		p.next()
		return s, nil
	case scanner.Ident:
		switch strings.ToLower(p.lit) {
		case "true":
			p.next()
			return true, nil
		case "false":
			p.next()
			return false, nil
		case "none":
			p.next()
			return nil, nil
		default:
			return nil, fmt.Errorf("unknown keyword %q at %s", p.lit, p.scanner.Position)
		}
	case '-':
		p.next()
		if p.tok == scanner.Int {
			i, err := strconv.ParseInt(p.lit, 10, 64)
			if err != nil {
				return nil, fmt.Errorf("invalid integer %q: %w", p.lit, err)
			}
			p.next()
			return -i, nil
		} else if p.tok == scanner.Float {
			f, err := strconv.ParseFloat(p.lit, 64)
			if err != nil {
				return nil, fmt.Errorf("invalid float %q: %w", p.lit, err)
			}
			p.next()
			return -f, nil
		}
		return nil, fmt.Errorf("expected number after '-' at %s", p.scanner.Position)
	default:
		return nil, fmt.Errorf("unexpected token %q at %s", p.lit, p.scanner.Position)
	}
}

func (p *domainParser) parseSingleQuotedString() (string, error) {
	// Current token is single quote, read string content manually
	var sb strings.Builder
	for {
		ch := p.scanner.Next()
		if ch == scanner.EOF {
			return "", fmt.Errorf("unterminated string at %s", p.scanner.Position)
		}
		if ch == '\\' {
			next := p.scanner.Next()
			switch next {
			case '\\':
				sb.WriteByte('\\')
			case '\'':
				sb.WriteByte('\'')
			case '"':
				sb.WriteByte('"')
			case 'n':
				sb.WriteByte('\n')
			case 't':
				sb.WriteByte('\t')
			case 'r':
				sb.WriteByte('\r')
			default:
				sb.WriteByte('\\')
				sb.WriteRune(next)
			}
			continue
		}
		if ch == '\'' {
			p.next() // Move to next token after the closing quote
			return sb.String(), nil
		}
		sb.WriteRune(ch)
	}
}
