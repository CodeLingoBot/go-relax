// Copyright 2014 Codehack.com All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package relax

import (
	"net/http"
)

const (
	headerXHTTPMethodOverride = "X-HTTP-Method-Override"
	queryMethodOverride       = "_method"
)

var (
	// OverrideFilterMethods specifies the methods can be overriden.
	// Format is OverrideFilterMethods[{method}] = {override}
	OverrideFilterMethods = map[string]string{
		"DELETE":  "POST",
		"OPTIONS": "GET",
		"PATCH":   "POST",
		"PUT":     "POST",
	}
)

// OverrideFilter changes the Request.Method if the client specifies
// override via HTTP header or query. This allows clients with limited HTTP
// verbs to send REST requests through GET/POST.
type OverrideFilter struct {
	// Header expected for HTTP Method override
	Header string

	// QueryVar is used if header can't be set
	QueryVar string
}

func (self *OverrideFilter) Run(next HandlerFunc) HandlerFunc {
	if self.Header == "" {
		self.Header = headerXHTTPMethodOverride
	}
	if self.QueryVar == "" {
		self.QueryVar = queryMethodOverride
	}

	return func(rw ResponseWriter, re *Request) {
		if mo := re.URL.Query().Get(self.QueryVar); mo != "" {
			re.Header.Set(self.Header, mo)
		}
		if mo := re.Header.Get(self.Header); mo != "" {
			if mo != re.Method {
				override, ok := OverrideFilterMethods[mo]
				if !ok {
					rw.Error(http.StatusMethodNotAllowed, mo+" method is not overridable.")
					return
				}
				// check that the caller method matches the expected override. e.g., used GET for OPTIONS
				if re.Method != override {
					rw.Error(http.StatusPreconditionFailed, "must use "+override+" to override for "+mo)
					return
				}
				re.Method = override
				re.Header.Del(self.Header)
				re.Info.Set("override.method", override)
			}
		}
		next(rw, re)
	}
}