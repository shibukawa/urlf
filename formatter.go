package urlf

import (
	"errors"
	"fmt"
	"net/url"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"sync"
)

var ErrFormatFailed = errors.New("format failed")

// Opt is a struct for custom formatter options.
type Opt struct {
	Hostname string
	Port     uint16
	Protocol string
	Username string
	Password string
}

// CustomFormatter is a custom formatter function.
//
// Generated function has same signature with Urlf.
// But it can be customized by Opt.
//
// It is a "Must" version of TryCustomFormatter.
func CustomFormatter(o Opt) func(format string, args ...any) string {
	f := TryCustomFormatter(o)
	return func(format string, args ...any) string {
		if result, err := f(format, args...); err != nil {
			panic(err)
		} else {
			return result
		}
	}
}

var cache = sync.Map{}

// TryCustomFormatter generates a custom formatter function that returns an empty string.
func TryCustomFormatter(o Opt) func(format string, args ...any) (string, error) {
	return func(format string, args ...any) (string, error) {
		var ot *parseResult // original template
		if v, ok := cache.Load(format); ok {
			ot = v.(*parseResult)
		} else {
			var err error
			ot, err = parse(format)
			if err != nil {
				return "", err
			}
			cache.Store(format, ot)
		}
		t, err := overwrite(ot, o)
		if err != nil {
			return "", err
		}
		r := &url.URL{}

		// Scheme
		if t.protocol != nil {
			if t.protocol.partType == staticPart {
				r.Scheme = t.protocol.value
			} else {
				switch v := args[t.protocol.index].(type) {
				case string:
					r.Scheme = v
				case *string:
					r.Scheme = *v
				case nil:
					// do nothing
				default:
					return "", fmt.Errorf("%w: invalid protocol value. only string param is available, but '%v'", ErrFormatFailed, args[t.protocol.index])
				}
			}
		}

		// Host
		if t.hostname != nil {
			if t.hostname.partType == staticPart {
				r.Host = t.hostname.value
			} else {
				switch v := args[t.hostname.index].(type) {
				case string:
					r.Host = v
				case *string:
					r.Host = *v
				case nil: // omit scheme too
					r.Scheme = ""
				default:
					return "", fmt.Errorf("%w: invalid hostname value. only string param is available, but '%v'", ErrFormatFailed, args[t.hostname.index])
				}
			}
		}

		// Port
		if t.port != nil && r.Host != "" {
			if t.port.partType == staticPart {
				r.Host += ":" + strconv.Itoa(int(t.port.value))
			} else {
				switch v := args[t.port.index].(type) {
				case int:
					r.Host += ":" + strconv.Itoa(v)
				case *int:
					r.Host += ":" + strconv.Itoa(*v)
				case nil:
					// do nothing
				default:
					return "", fmt.Errorf("%w: invalid port value. only int param is available, but '%v'", ErrFormatFailed, args[t.port.index])
				}
			}
		}

		// Path
		var paths []string
		for _, p := range t.paths {
			if p.partType == staticPart {
				paths = append(paths, p.value)
			} else {
				// todo error check
				v := args[p.index]
				switch v2 := v.(type) {
				case string:
					paths = append(paths, v2)
				case *string:
					paths = append(paths, *v2)
				case int:
					paths = append(paths, strconv.Itoa(v2))
				case *int:
					paths = append(paths, strconv.Itoa(*v2))
				case nil:
					// do nothing
				default:
					rv := reflect.ValueOf(v)
					if rv.Kind() == reflect.Slice {
						for i := 0; i < rv.Len(); i++ {
							switch ev := rv.Index(i).Interface().(type) {
							case string:
								paths = append(paths, "/"+ev)
							case *string:
								paths = append(paths, "/"+*ev)
							case int:
								paths = append(paths, "/"+strconv.Itoa(ev))
							case *int:
								paths = append(paths, "/"+strconv.Itoa(*ev))
							case nil:
								// do nothing
							}
						}
					}
				}

			}
		}

		// Query
		query := url.Values{}

		updateQuery := func(key string, value any) error {
			switch v := value.(type) {
			case string:
				query.Add(key, v)
			case *string:
				query.Add(key, *v)
			case int:
				query.Add(key, strconv.Itoa(v))
			case *int:
				query.Add(key, strconv.Itoa(*v))
			case nil:
			default:
				rv := reflect.ValueOf(v)
				if rv.Kind() == reflect.Slice {
					for i := 0; i < rv.Len(); i++ {
						switch ev := rv.Index(i).Interface().(type) {
						case string:
							if i == 0 {
								query.Set(key, ev)
							} else {
								query.Add(key, ev)
							}
						case *string:
							if i == 0 {
								query.Set(key, *ev)
							} else {
								query.Add(key, *ev)
							}
						case int:
							if i == 0 {
								query.Set(key, strconv.Itoa(ev))
							} else {
								query.Add(key, strconv.Itoa(ev))
							}
						case *int:
							if i == 0 {
								query.Set(key, strconv.Itoa(*ev))
							} else {
								query.Add(key, strconv.Itoa(*ev))
							}
						case nil:
							// do nothing
						}
					}
				} else {
					return fmt.Errorf("%w: query value must be string, int, nil, [](string|int), but '%v'", ErrFormatFailed, value)
				}
			}
			return nil
		}
		for _, q := range t.queries {
			if q.value.partType == staticPart {
				query.Add(q.key, q.value.value)
			} else if q.key != "" {
				if err := updateQuery(q.key, args[q.value.index]); err != nil {
					return "", err
				}
			} else if vs, ok := args[q.value.index].(url.Values); ok {
				for key, values := range vs {
					if err := updateQuery(key, values); err != nil {
						return "", err
					}
				}
			} else {
				return "", fmt.Errorf("%w: query set must be url.Values, but '%v'", ErrFormatFailed, args[q.value.index])
			}
		}
		r.RawQuery = query.Encode()

		if t.fragment != nil {
			if t.fragment.partType == staticPart {
				r.Fragment = t.fragment.value
			} else {
				switch v := args[t.fragment.index].(type) {
				case string:
					r.Fragment = v
				case *string:
					r.Fragment = *v
				case nil:
					// do nothing
				default:
					return "", fmt.Errorf("%w: fragment must be a string, but '%v'", ErrFormatFailed, args[t.fragment.index])
				}
			}
		}

		for _, p := range paths {
			if strings.HasSuffix(r.Path, "/") && strings.HasPrefix(p, "/") {
				r.Path = r.Path + p[1:]
			} else {
				r.Path += p
			}
		}

		if t.username != "" {
			r.User = url.UserPassword(t.username, t.password)
		}

		return r.String(), nil
	}
}

// Urlf is a default formatter function.
//
// It is a "Must" version of TryUrlf. It assumes URL template string is written as a static string literal
// and the template is not created dynamically.
// So, it raise panic instead of return error.
//
// If you want to get parsing error, use TryUrlf, instead.
func Urlf(format string, args ...any) string {
	return CustomFormatter(Opt{})(format, args...)
}

// TryUrlf is a similar function to Urlf, but it returns an error if the format is invalid.
func TryUrlf(format string, args ...any) (string, error) {
	return TryCustomFormatter(Opt{})(format, args...)
}

var hostPattern = regexp.MustCompile(`^(?P<protocol>\w+:\/\/)?(?P<hostname>[^:]+)(?P<port>:\d+)?`)

func overwrite(src *parseResult, opt Opt) (result *parseResult, err error) {
	result = &parseResult{
		protocol: src.protocol,
		hostname: src.hostname,
		port:     src.port,
		paths:    src.paths,
		queries:  src.queries,
		fragment: src.fragment,
	}

	if opt.Hostname != "" {
		match := hostPattern.FindStringSubmatch(opt.Hostname)
		if match[1] != "" {
			result.protocol = &part[string]{partType: staticPart, value: match[1][:len(match[1])-3]}
		}
		if match[2] != "" {
			result.hostname = &part[string]{partType: staticPart, value: match[2]}
		}
		if match[3] != "" {
			p, err := strconv.ParseUint(match[3][1:], 10, 16)
			if err != nil {
				return nil, fmt.Errorf("%w: invalid port number '%s'", ErrParseFailed, match[3])
			}
			result.port = &part[uint16]{partType: staticPart, value: uint16(p)}
		}
	}
	if opt.Protocol != "" {
		result.protocol = &part[string]{partType: staticPart, value: opt.Protocol}
	}
	if opt.Port != 0 {
		result.port = &part[uint16]{partType: staticPart, value: opt.Port}
	}
	if opt.Username != "" && opt.Password != "" {
		result.username = opt.Username
		result.password = opt.Password
	} else if opt.Username != "" || opt.Password != "" {
		return nil, fmt.Errorf("%w: both username and password must be set", ErrParseFailed)
	}
	return result, nil
}
