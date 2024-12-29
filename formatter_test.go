package urlf

import (
	"log"
	"net/url"
	"testing"

	"github.com/alecthomas/assert/v2"
)

func init() {
	log.SetFlags(log.Lshortfile | log.LstdFlags)
	log.SetPrefix("🦑")
}

func TestFormatter(t *testing.T) {
	tests := []struct {
		name       string
		actual     func() string
		wantResult string
	}{
		{
			name:       "simple",
			actual:     func() string { return Urlf("http://example.com/{}", 1000).String() },
			wantResult: "http://example.com/1000",
		},
		{
			name:       "domain",
			actual:     func() string { return Urlf("{}://bucket.example.com/file/path", "s3").String() },
			wantResult: "s3://bucket.example.com/file/path",
		},
		{
			name: "domain (string pointer)",
			actual: func() string {
				protocol := "s3"
				return Urlf("{}://bucket.example.com/file/path", &protocol).String()
			},
			wantResult: "s3://bucket.example.com/file/path",
		},
		{
			name:       "protocol-relative URL (static)",
			actual:     func() string { return Urlf(`//bucket.example.com/file/path`).String() },
			wantResult: "//bucket.example.com/file/path",
		},
		{
			name:       "protocol-relative URL (dynamic)",
			actual:     func() string { return Urlf(`{}://bucket.example.com/file/path`, nil).String() },
			wantResult: "//bucket.example.com/file/path",
		},
		{
			name:       "hostname",
			actual:     func() string { return Urlf(`http://{}/to/resource/path`, "api.example.com").String() },
			wantResult: "http://api.example.com/to/resource/path",
		},
		{
			name: "hostname (string pointer)",
			actual: func() string {
				hostname := "api.example.com"
				return Urlf(`http://{}/to/resource/path`, &hostname).String()
			},
			wantResult: "http://api.example.com/to/resource/path",
		},
		{
			name:       "omit hostname (static)",
			actual:     func() string { return Urlf(`/to/resource/path`).String() },
			wantResult: "/to/resource/path",
		},
		{
			name:       "omit hostname (dynamic)",
			actual:     func() string { return Urlf(`http://{}/to/resource/path`, nil).String() },
			wantResult: "/to/resource/path",
		},
		{
			name:       "port",
			actual:     func() string { return Urlf(`http://api.example.com:{}/to/resource/path`, 1000).String() },
			wantResult: "http://api.example.com:1000/to/resource/path",
		},
		{
			name: "port (pointer)",
			actual: func() string {
				port := 1000
				return Urlf(`http://api.example.com:{}/to/resource/path`, &port).String()
			},
			wantResult: "http://api.example.com:1000/to/resource/path",
		},
		{
			name:       "omit port (dynamic)",
			actual:     func() string { return Urlf(`http://api.example.com:{}/to/resource/path`, nil).String() },
			wantResult: "http://api.example.com/to/resource/path",
		},
		{
			name:       "path placeholder - string",
			actual:     func() string { return Urlf(`http://api.example.com/users/{}/`, "bob").String() },
			wantResult: "http://api.example.com/users/bob/",
		},
		{
			name: "path placeholder - string pointer",
			actual: func() string {
				name := "bob"
				return Urlf(`http://api.example.com/users/{}/`, &name).String()
			},
			wantResult: "http://api.example.com/users/bob/",
		},
		{
			name:       "path placeholder - number",
			actual:     func() string { return Urlf(`http://api.example.com/users/{}/`, 1000).String() },
			wantResult: "http://api.example.com/users/1000/",
		},
		{
			name: "path placeholder - number pointer",
			actual: func() string {
				userCode := 1000
				return Urlf(`http://api.example.com/users/{}/`, &userCode).String()
			},
			wantResult: "http://api.example.com/users/1000/",
		},
		{
			name:       "path placeholder - array",
			actual:     func() string { return Urlf(`http://api.example.com/users/{}/`, []any{"a", "b", 1000}).String() },
			wantResult: "http://api.example.com/users/a/b/1000/",
		},
		{
			name:       "path placeholder - string with path separator",
			actual:     func() string { return Urlf(`http://api.example.com/users/{}/`, "a/b/1000").String() },
			wantResult: "http://api.example.com/users/a/b/1000/",
		},
		{
			name:       "path placeholder - string with path separator can escape correctly",
			actual:     func() string { return Urlf(`http://api.example.com/users/{}/`, "a/b/🐙").String() },
			wantResult: "http://api.example.com/users/a/b/%F0%9F%90%99/",
		},
		{
			name:       "path placeholder - array (empty)",
			actual:     func() string { return Urlf(`http://api.example.com/users/{}/`, []any{}).String() },
			wantResult: "http://api.example.com/users/",
		},
		{
			name:       "query placeholder - static",
			actual:     func() string { return Urlf(`http://api.example.com/users/?key=value`).String() },
			wantResult: "http://api.example.com/users/?key=value",
		},
		{
			name:       "query placeholder - static - same keys",
			actual:     func() string { return Urlf(`http://api.example.com/users/?key=value&key=value2`).String() },
			wantResult: "http://api.example.com/users/?key=value&key=value2",
		},
		{
			name:       "query placeholder - dynamic string",
			actual:     func() string { return Urlf(`http://api.example.com/users/?key={}`, "str-value").String() },
			wantResult: "http://api.example.com/users/?key=str-value",
		},
		{
			name: "query placeholder - dynamic string pointer",
			actual: func() string {
				value := "str-value"
				return Urlf(`http://api.example.com/users/?key={}`, &value).String()
			},
			wantResult: "http://api.example.com/users/?key=str-value",
		},
		{
			name:       "query placeholder - null",
			actual:     func() string { return Urlf(`http://api.example.com/users/?key={}`, nil).String() },
			wantResult: "http://api.example.com/users/",
		},
		{
			name: "query placeholder - dynamic array: overwrite existing key",
			actual: func() string {
				return Urlf(`http://api.example.com/users/?key=old&key={}`, []any{"a", "b", "c"}).String()
			},
			wantResult: "http://api.example.com/users/?key=a&key=b&key=c",
		},
		{
			name: "query placeholder - query set via record via url.Values",
			actual: func() string {
				q, err := url.ParseQuery("key=a&key=b&key=c&key2=value")
				if err != nil {
					panic(err)
				}
				return Urlf(`http://api.example.com/users/?key=old&{}`, q).String()
			},
			wantResult: "http://api.example.com/users/?key=a&key=b&key=c&key2=value",
		},
		{
			name:       "hash placeholder - static",
			actual:     func() string { return Urlf(`http://api.example.com/users/#hash`).String() },
			wantResult: "http://api.example.com/users/#hash",
		},
		{
			name:       "hash placeholder - dynamic",
			actual:     func() string { return Urlf(`http://api.example.com/users/#{}`, "hash").String() },
			wantResult: "http://api.example.com/users/#hash",
		},
		{
			name: "hash placeholder - dynamic pointer",
			actual: func() string {
				hash := "hash"
				return Urlf(`http://api.example.com/users/#{}`, &hash).String()
			},
			wantResult: "http://api.example.com/users/#hash",
		},
		{
			name:       "hash placeholder - omit",
			actual:     func() string { return Urlf(`http://api.example.com/users/#{}`, nil).String() },
			wantResult: "http://api.example.com/users/",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.wantResult, tt.actual())
		})
	}
}

func TestCustomFormatter(t *testing.T) {
	tests := []struct {
		name       string
		actual     func() string
		wantResult string
	}{
		{
			name: "credentials",
			actual: func() string {
				url := CustomFormatter(Opt{Username: "user", Password: "pass"})
				return url("http://example.com/{}", 1000).String()
			},
			wantResult: "http://user:pass@example.com/1000",
		},
		{
			name: "scheme",
			actual: func() string {
				url := CustomFormatter(Opt{Protocol: "s3"})
				return url("http://example.com/{}", 1000).String()
			},
			wantResult: "s3://example.com/1000",
		},
		{
			name: "hostname(simple)",
			actual: func() string {
				url := CustomFormatter(Opt{Hostname: "api.example.com"})
				return url("http://api-server/{}", 1000).String()
			},
			wantResult: "http://api.example.com/1000",
		},
		{
			name: "port",
			actual: func() string {
				url := CustomFormatter(Opt{Port: 8080})
				return url("http://example.com/{}", 1000).String()
			},
			wantResult: "http://example.com:8080/1000",
		},
		{
			name: "host with scheme, port",
			actual: func() string {
				url := CustomFormatter(Opt{Hostname: "https://api.example.com:8080"})
				return url("http://example.com/{}", 1000).String()
			},
			wantResult: "https://api.example.com:8080/1000",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.wantResult, tt.actual())
		})
	}
}
