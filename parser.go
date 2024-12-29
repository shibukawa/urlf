package urlf

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
)

var ErrParseFailed = errors.New("parse failed")

type partType int

const (
	staticPart partType = iota + 1
	paramPart
)

type part[T comparable] struct {
	partType partType
	index    int
	value    T
}

type queryPart struct {
	key   string
	value part[string]
}

type parseResult struct {
	protocol *part[string]
	hostname *part[string]
	port     *part[uint16]
	paths    []part[string]
	queries  []queryPart
	fragment *part[string]
	username string
	password string
}

type stepType int

const (
	protocol stepType = iota + 1
	hostname
	port
	path
	query
	queryKey
	queryValue
	fragment
	invalid
)

var invalidSeparator = map[stepType]map[string]bool{
	protocol:   {"://": true, "//": false, ":": true, "/": false, "?": true, "=": true, "&": true, "#": true, "@": true},
	path:       {"://": true, "//": true, ":": true, "/": false, "?": false, "=": true, "&": true, "#": false, "@": true},
	query:      {"://": true, "//": true, ":": true, "/": true, "?": false, "=": true, "&": true, "#": false, "@": true},
	queryKey:   {"://": true, "//": true, ":": true, "/": true, "?": true, "=": false, "&": false, "#": false, "@": true},
	queryValue: {"://": true, "//": true, ":": true, "/": true, "?": true, "=": true, "&": false, "#": false, "@": true},
}

var splitterPattern = regexp.MustCompile(`(?::\/\/)|(?:\/\/)|[:/?&=#@]|\{\}`)

type tokenType int

const (
	separator tokenType = iota + 1
	static
	placeholder
)

type token struct {
	tokenType tokenType
	text      string
	index     int
}

func parse(pattern string) (result *parseResult, err error) {
	result = &parseResult{}

	i := 0
	placeholderIndex := 0
	matches := splitterPattern.FindAllStringIndex(pattern, -1)
	tokens := make([]token, 0, len(matches)*2+1)
	for _, m := range matches {
		if i < m[0] {
			tokens = append(tokens, token{tokenType: static, text: pattern[i:m[0]]})
		}
		s := pattern[m[0]:m[1]]
		if s == "{}" {
			tokens = append(tokens, token{tokenType: placeholder, index: placeholderIndex})
			placeholderIndex++
		} else {
			tokens = append(tokens, token{tokenType: separator, text: s})
		}
		i = m[1]
	}
	if i < len(pattern) {
		tokens = append(tokens, token{tokenType: static, text: pattern[i:]})
	}

	appendPath := func(pathString string) {
		if len(result.paths) == 0 {
			result.paths = append(result.paths, part[string]{partType: staticPart, value: pathString})
		} else {
			l := result.paths[len(result.paths)-1] // last
			if l.partType == staticPart {
				result.paths[len(result.paths)-1] = part[string]{partType: staticPart, value: l.value + pathString}
			} else {
				result.paths = append(result.paths, part[string]{partType: staticPart, value: pathString})
			}
		}
	}

	var lastToken string
	step := protocol
	var queryKeyStr string

	for len(tokens) > 0 {
		switch step {
		case protocol:
			{
				p := tokens[0] // protocol
				if len(tokens) > 1 {
					s := tokens[1] // separator
					if s.tokenType == separator && s.text == "://" {
						if p.tokenType == placeholder {
							result.protocol = &part[string]{partType: paramPart, index: p.index}
						} else if p.tokenType == static && p.text == "" {
							return nil, fmt.Errorf("%w: protocol name should not be empty", ErrParseFailed)
						} else {
							result.protocol = &part[string]{partType: staticPart, value: p.text}
						}
						step = hostname
						tokens = tokens[2:]
						lastToken = s.text
						break
					}
					if p.tokenType == separator {
						if invalidSeparator[protocol][p.text] {
							return nil, fmt.Errorf("%w: invalid character: '%s'. only protocol name or //, / are available", ErrParseFailed, p.text)
						}
						if p.text == "//" {
							// Scheme relative URL
							step = hostname
							tokens = tokens[1:]
							lastToken = p.text
						} else {
							step = path
						}
					} else {
						step = path
					}
				}
				break
			}
		case hostname:
			{
				h := tokens[0] // hostname
				if h.tokenType == separator {
					return nil, fmt.Errorf("%w: invalid character: '%s'. after '%s' only hostname string is expected", ErrParseFailed, h.text, lastToken)
				}
				if h.tokenType == placeholder {
					result.hostname = &part[string]{partType: paramPart, index: h.index}
				} else {
					result.hostname = &part[string]{partType: staticPart, value: h.text}
				}
				tokens = tokens[1:]
				lastToken = "hostname"
				step = port
				break
			}
		case port:
			{
				ps := tokens[0] // port separator
				if ps.tokenType == separator && ps.text == ":" {
					if len(tokens) > 1 {
						p := tokens[1] // port
						switch p.tokenType {
						case separator:
							return nil, fmt.Errorf("%w: invalid character: '%s'. after ':' only port number is expected", ErrParseFailed, p.text)
						case placeholder:
							result.port = &part[uint16]{partType: paramPart, index: p.index}
						case static:
							pn, err := strconv.Atoi(p.text)
							if err != nil {
								return nil, fmt.Errorf("%w: port must be a number, but '%s'", ErrParseFailed, p.text)
							}
							if pn < 1 || pn > 65535 {
								return nil, fmt.Errorf("%w: port number must be in range 1-65535, but %d", ErrParseFailed, pn)
							}
							result.port = &part[uint16]{partType: staticPart, value: uint16(pn)}
						}
						lastToken = "port"
						tokens = tokens[2:]
					} else {
						return nil, fmt.Errorf("%w: port number is expected after ':'", ErrParseFailed)
					}
				}
				step = path
				break
			}
		case path:
			{
				s := tokens[0] // separator
				switch s.tokenType {
				case placeholder:
					return nil, fmt.Errorf("%w: invalid placeholder after %s", ErrParseFailed, lastToken)
				case static:
					// if input is relative path like "./path/to/resource" or "path/to/resource", it is ok.
					if (result.protocol != nil || result.hostname != nil) && len(result.paths) == 0 {
						return nil, fmt.Errorf("%w: invalid text after '%s': '%s'", ErrParseFailed, lastToken, s.text)
					}
					appendPath(s.text)
					tokens = tokens[1:]
				case separator:
					if invalidSeparator[path][s.text] {
						return nil, fmt.Errorf("%w: invalid character after %s: '/', '?', '#' are available but '%s'", ErrParseFailed, lastToken, s.text)
					}
					if s.text != "/" {
						step = query
					} else if len(tokens) > 1 {
						p := tokens[1] // path
						switch p.tokenType {
						case separator:
							appendPath("/")
							lastToken = "/"
							step = query
							tokens = tokens[1:]
						case placeholder:
							appendPath("/")
							result.paths = append(result.paths, part[string]{partType: paramPart, index: p.index})
							tokens = tokens[2:]
							lastToken = fmt.Sprintf("{%d}", p.index)
						case static:
							lastToken = "/" + p.text
							appendPath(lastToken)
							tokens = tokens[2:]
						}
					} else {
						// last token
						appendPath("/")
						tokens = tokens[1:]
						step = invalid
					}
				}
			}
		case query:
			{
				s := tokens[0] // separator
				switch s.tokenType {
				case separator:
					if invalidSeparator[query][s.text] {
						return nil, fmt.Errorf("%w: invalid character after %s should be '?', '=', '&', '#' but '%s'", ErrParseFailed, lastToken, s.text)
					}
					switch s.text {
					case "?":
						lastToken = "?"
						step = queryKey
					case "#":
						step = fragment
					}
					tokens = tokens[1:]
				case placeholder:
					return nil, fmt.Errorf("%w: invalid placeholder {%d} after %s. It should be '?' or '#'", ErrParseFailed, s.index, lastToken)
				case static:
					return nil, fmt.Errorf("%w: invalid character after %s should be '?', '#' but '%s'", ErrParseFailed, lastToken, s.text)
				}
			}
		case queryKey:
			{
				qk := tokens[0] // query key
				switch qk.tokenType {
				case separator:
					return nil, fmt.Errorf("%w: query key should be a string or placeholder, but  '%s'", ErrParseFailed, qk.text)
				case placeholder: // query set
					if len(tokens) > 1 {
						s := tokens[1] // splitter
						switch s.tokenType {
						case static:
							return nil, fmt.Errorf("%w: invalid character after query set placeholder {%d}. Only &, # are available, but '%s'", ErrParseFailed, qk.index, s.text)
						case separator:
							if invalidSeparator[queryValue][s.text] {
								return nil, fmt.Errorf("%w: invalid character after query set placeholder {%d}. only &, # are available but '%s'", ErrParseFailed, qk.index, s.text)
							}
							if s.text == "#" {
								step = fragment
							}
						}
						tokens = tokens[2:]
					} else {
						// last token
						step = invalid
						tokens = tokens[1:]
					}
					result.queries = append(result.queries, queryPart{key: "", value: part[string]{partType: paramPart, index: qk.index}})
				case static:
					if len(tokens) > 1 {
						s := tokens[1] // splitter
						if s.tokenType == separator {
							if invalidSeparator[queryKey][s.text] {
								return nil, fmt.Errorf("%w: invalid character after query key '%s'. only =, &, # are available but '%s'", ErrParseFailed, qk.text, s.text)
							}
							switch s.text {
							case "=":
								queryKeyStr = qk.text
								lastToken = qk.text
								step = queryValue
							case "#":
								step = fragment
								fallthrough
							case "&":
								result.queries = append(result.queries, queryPart{key: qk.text, value: part[string]{partType: staticPart, value: ""}})
							}
							tokens = tokens[2:]
						}
					} else {
						result.queries = append(result.queries, queryPart{key: qk.text, value: part[string]{partType: staticPart, value: ""}})
					}
				}
			}
		case queryValue:
			{
				qv := tokens[0] // query value
				switch qv.tokenType {
				case separator:
					return nil, fmt.Errorf("%w: query value of '%s' should be a string or placeholder, but '%s'", ErrParseFailed, queryKeyStr, qv.text)
				case placeholder:
					result.queries = append(result.queries, queryPart{key: queryKeyStr, value: part[string]{partType: paramPart, index: qv.index}})
				case static:
					result.queries = append(result.queries, queryPart{key: queryKeyStr, value: part[string]{partType: staticPart, value: qv.text}})
				}
				if len(tokens) > 1 {
					s := tokens[1] // splitter
					switch s.tokenType {
					case placeholder:
						return nil, fmt.Errorf("%w: invalid placeholder {%d} after query value of '%s'", ErrParseFailed, s.index, queryKeyStr)
					case static:
						return nil, fmt.Errorf("%w: invalid text '%s' after query value of '%s'", ErrParseFailed, s.text, queryKeyStr)
					case separator:
						if invalidSeparator[queryValue][s.text] {
							return nil, fmt.Errorf("%w: invalid character after query value of '%s'. only &, # are available but '%s", ErrParseFailed, queryKeyStr, s.text)
						}
						switch s.text {
						case "&":
							step = queryKey
						case "#":
							step = fragment
						}
						tokens = tokens[2:]
					}
				} else {
					tokens = tokens[1:]
				}
			}
		case fragment:
			{
				f := tokens[0] // fragment
				switch f.tokenType {
				case separator:
					return nil, fmt.Errorf("%w: invalid character after fragment. A static string or placeholder are available but '%s'", ErrParseFailed, f.text)
				case static:
					result.fragment = &part[string]{partType: staticPart, value: f.text}
				case placeholder:
					result.fragment = &part[string]{partType: paramPart, index: f.index}
				}
				tokens = tokens[1:]
				step = invalid // this should be the last step
			}
		case invalid:
			return nil, fmt.Errorf("%w: the url have invalid extra token: [%v]", ErrParseFailed, tokens)
		}
	}

	return result, nil
}
