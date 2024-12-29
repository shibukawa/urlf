package urlf

import (
	"testing"

	"github.com/alecthomas/assert/v2"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name       string
		args       string
		wantResult *parseResult
	}{
		{
			name: "no param: protocol, nostname",
			args: "http://example.com",
			wantResult: &parseResult{
				protocol: &part[string]{partType: staticPart, value: "http"},
				hostname: &part[string]{partType: staticPart, value: "example.com"},
			},
		},
		{
			name: "no param: protocol relative, nostname",
			args: "//example.com",
			wantResult: &parseResult{
				hostname: &part[string]{partType: staticPart, value: "example.com"},
			},
		},
		{
			name: "no param: hostname, port",
			args: `//example.com:8080`,
			wantResult: &parseResult{
				hostname: &part[string]{partType: staticPart, value: "example.com"},
				port:     &part[uint16]{partType: staticPart, value: 8080},
			},
		},
		{
			name: "no param: protocol, hostname, path",
			args: `http://example.com/path/to/resource`,
			wantResult: &parseResult{
				protocol: &part[string]{partType: staticPart, value: "http"},
				hostname: &part[string]{partType: staticPart, value: "example.com"},
				paths:    []part[string]{{partType: staticPart, value: "/path/to/resource"}},
			},
		},
		{
			name: "no param: abs path",
			args: `/path/to/resource`,
			wantResult: &parseResult{
				paths: []part[string]{{partType: staticPart, value: "/path/to/resource"}},
			},
		},
		{
			name: "no param: rel path",
			args: `./path/to/resource`,
			wantResult: &parseResult{
				paths: []part[string]{{partType: staticPart, value: "./path/to/resource"}},
			},
		},
		{
			name: "no param: protocol, hostname, query",
			args: `http://example.com?query1=value1&query2=value2`,
			wantResult: &parseResult{
				protocol: &part[string]{partType: staticPart, value: "http"},
				hostname: &part[string]{partType: staticPart, value: "example.com"},
				queries: []queryPart{
					{key: "query1", value: part[string]{partType: staticPart, value: "value1"}},
					{key: "query2", value: part[string]{partType: staticPart, value: "value2"}},
				},
			},
		},
		{
			name: "no param: protocol, hostname, fragment",
			args: `http://example.com#test`,
			wantResult: &parseResult{
				protocol: &part[string]{partType: staticPart, value: "http"},
				hostname: &part[string]{partType: staticPart, value: "example.com"},
				fragment: &part[string]{partType: staticPart, value: "test"},
			},
		},
		{
			name: "no param: protocol, hostname, port, path",
			args: `http://example.com:8080/path/to/resource`,
			wantResult: &parseResult{
				protocol: &part[string]{partType: staticPart, value: "http"},
				hostname: &part[string]{partType: staticPart, value: "example.com"},
				port:     &part[uint16]{partType: staticPart, value: 8080},
				paths:    []part[string]{{partType: staticPart, value: "/path/to/resource"}},
			},
		},
		{
			name: "no param: protocol, hostname, port, path, query",
			args: `http://example.com:8080/path/to/resource?query1=value1&query2=value2`,
			wantResult: &parseResult{
				protocol: &part[string]{partType: staticPart, value: "http"},
				hostname: &part[string]{partType: staticPart, value: "example.com"},
				port:     &part[uint16]{partType: staticPart, value: 8080},
				paths:    []part[string]{{partType: staticPart, value: "/path/to/resource"}},
				queries: []queryPart{
					{key: "query1", value: part[string]{partType: staticPart, value: "value1"}},
					{key: "query2", value: part[string]{partType: staticPart, value: "value2"}},
				},
			},
		},
		{
			name: "no param: protocol, hostname, port, path, fragment",
			args: `http://example.com:8080/path/to/resource#test`,
			wantResult: &parseResult{
				protocol: &part[string]{partType: staticPart, value: "http"},
				hostname: &part[string]{partType: staticPart, value: "example.com"},
				port:     &part[uint16]{partType: staticPart, value: 8080},
				paths:    []part[string]{{partType: staticPart, value: "/path/to/resource"}},
				fragment: &part[string]{partType: staticPart, value: "test"},
			},
		},
		{
			name: "no param: protocol, hostname, port, path, query, fragment",
			args: `http://example.com:8080/path/to/resource?query1=value1&query2=value2#test`,
			wantResult: &parseResult{
				protocol: &part[string]{partType: staticPart, value: "http"},
				hostname: &part[string]{partType: staticPart, value: "example.com"},
				port:     &part[uint16]{partType: staticPart, value: 8080},
				paths:    []part[string]{{partType: staticPart, value: "/path/to/resource"}},
				queries: []queryPart{
					{key: "query1", value: part[string]{partType: staticPart, value: "value1"}},
					{key: "query2", value: part[string]{partType: staticPart, value: "value2"}},
				},
				fragment: &part[string]{partType: staticPart, value: "test"},
			},
		},
		{
			name: "no param: protocol, hostname, port, query",
			args: `http://example.com:8080?query1=value1&query2=value2`,
			wantResult: &parseResult{
				protocol: &part[string]{partType: staticPart, value: "http"},
				hostname: &part[string]{partType: staticPart, value: "example.com"},
				port:     &part[uint16]{partType: staticPart, value: 8080},
				queries: []queryPart{
					{key: "query1", value: part[string]{partType: staticPart, value: "value1"}},
					{key: "query2", value: part[string]{partType: staticPart, value: "value2"}},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotResult, err := parse(tt.args)
			assert.NoError(t, err)
			assert.Equal(t, tt.wantResult, gotResult)
		})
	}
}
