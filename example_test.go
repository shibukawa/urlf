package urlf_test

import (
	"fmt"

	"github.com/shibukawa/urlf"
)

func ExampleUrtf() {
	url := urlf.Urlf("http://example.com/{}/", 1000)
	fmt.Println(url)
	// Output: http://example.com/1000/
}

func ExampleCustomFormatter() {
	formatter := urlf.CustomFormatter(urlf.Opt{
		Hostname: "api.example.com",
		Protocol: "https",
	})
	url := formatter("http://api-server/api/users/{}", 1000)
	fmt.Println(url)
	// Output: https://api.example.com/api/users/1000
}
