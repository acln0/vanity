// Copyright 2018 Andrei Tudor Călin
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package vanity provides utilities for building vanity Go import servers.
package vanity

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"net/url"
	"path"
	"strings"
)

// ImportPath defines a mapping between Go import paths.
type ImportPath struct {
	VCS  string
	From string
	To   string
}

// TagFor returns the import tag for the request req.
//
// TagFor verifies that the import path specified by req is either
// identical to ip.From or a sub-path of it.
//
// It returns an ImportTag pointing to the root of the import path, ip.From.
//
// For example, given ip.From == "acln.ro/foo" and a request to
// "acln.ro/foo/bar", the returned ImportTag would have .ImportPath ==
// "acln.ro/foo".
func (ip ImportPath) TagFor(req *http.Request) (*ImportTag, error) {
	p := path.Join(req.Host, req.URL.Path)
	if p != ip.From && !strings.HasPrefix(p, ip.From+"/") {
		return nil, fmt.Errorf("blah")
	}
	return &ImportTag{
		ImportPath: ip.From,
		VCS:        ip.VCS,
		VCSRepo:    ip.To,
	}, nil
}

// WildcardTagFor returns the wildcard import tag for the request req.
//
// TagFor verifies that the import path specified by req is a strict sub-path
// of ip.From.
//
// It returns an ImportTag pointing to the root of the import path, based on
// the first child element of the import path, beyond ip.From.
//
// For example, given ip.From == "acln.ro" and a request to "acln.ro/foo/bar",
// the returned ImportTag would have .ImportPath == "acln.ro/foo".
func (ip ImportPath) WildcardTagFor(req *http.Request) (*ImportTag, error) {
	p := path.Join(req.Host, req.URL.Path)
	if !strings.HasPrefix(p, ip.From+"/") {
		return nil, fmt.Errorf("blah")
	}
	seg := p[len(ip.From)+1:]
	if i := strings.IndexByte(seg, '/'); i >= 0 {
		seg = seg[:i]
	}
	return &ImportTag{
		ImportPath: path.Join(ip.From, seg),
		VCS:        ip.VCS,
		VCSRepo:    path.Join(ip.To, seg),
	}, nil
}

var importTagTemplate = template.Must(template.New("meta").Parse(`
<!DOCTYPE html>
<html>
<head>
	<meta name="go-import" content="{{ .ImportPath }} {{ .VCS }} {{ .VCSRepo }}">
</head>
</html>
`))

// ImportTag represents an HTML go-import meta tag understood by the go tool.
type ImportTag struct {
	ImportPath string
	VCS        string
	VCSRepo    string
}

// Render renders an HTML document to w, containing the go-import meta tag
// represented by t.
func (t *ImportTag) Render(w io.Writer) error {
	return importTagTemplate.Execute(w, t)
}

var redirectTemplate = template.Must(template.New("redirect").Parse(`
<!DOCTYPE html>
<html>
<head>
</head>
<body>
	<a href="{{ . }}">Redirecting to documentation at {{ . }}</a>
</body>
</html>
`))

// RedirectToGodoc redirects req to the corresponding godoc page.
//
// The redirect URL is derived from req.Host and req.URL.Path. For
// example, a request to example.com/foo/bar is redirected to
// godoc.org/example.com/foo/bar.
func RedirectToGodoc(w http.ResponseWriter, req *http.Request) {
	target := &url.URL{
		Scheme: "https",
		Host:   "godoc.org",
		Path:   path.Join(req.Host, req.URL.Path),
	}
	resp := new(bytes.Buffer)
	if err := redirectTemplate.Execute(resp, target.String()); err != nil {
		status := http.StatusInternalServerError
		msg := fmt.Sprintf("internal server error: %v", err)
		http.Error(w, msg, status)
		return
	}
	w.Header().Set("Location", target.String())
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusFound)
	w.Write(resp.Bytes())
}

// IsGoGet returns a boolean indicating whether req is a go get HTTP request.
func IsGoGet(req *http.Request) bool {
	return req.FormValue("go-get") == "1"
}
