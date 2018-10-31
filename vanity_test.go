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

package vanity_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"acln.ro/vanity"
)

func TestSimpleImportPath(t *testing.T) {
	p := vanity.ImportPath{
		VCS:  "git",
		From: "acln.ro/foo",
		To:   "https://github.com/acln0/foo",
	}
	tests := []struct {
		path    string
		want    vanity.ImportTag
		wantErr bool
	}{
		{
			path: "acln.ro/foo",
			want: vanity.ImportTag{
				ImportPath: "acln.ro/foo",
				VCS:        "git",
				VCSRepo:    "https://github.com/acln0/foo",
			},
		},
		{
			path: "acln.ro/foo/bar",
			want: vanity.ImportTag{
				ImportPath: "acln.ro/foo",
				VCS:        "git",
				VCSRepo:    "https://github.com/acln0/foo",
			},
		},
		{
			path: "acln.ro/foo/bar/baz",
			want: vanity.ImportTag{
				ImportPath: "acln.ro/foo",
				VCS:        "git",
				VCSRepo:    "https://github.com/acln0/foo",
			},
		},
		{
			path:    "acln.ro/",
			wantErr: true,
		},
		{
			path:    "acln.ro/bar",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		req, err := http.NewRequest("GET", "https://"+tt.path, nil)
		if err != nil {
			t.Fatal(err)
		}
		tag, err := p.TagFor(req)
		if err != nil && !tt.wantErr {
			t.Errorf("%s: %v", tt.path, err)
		}
		if err == nil && tt.wantErr {
			t.Errorf("%s: got %#v, want error", tt.path, tag)
		}
	}
}

func TestWildcardImportPath(t *testing.T) {
	p := vanity.ImportPath{
		VCS:  "git",
		From: "acln.ro",
		To:   "https://github.com/acln0",
	}
	tests := []struct {
		path    string
		want    vanity.ImportTag
		wantErr bool
	}{
		{
			path: "acln.ro/foo",
			want: vanity.ImportTag{
				ImportPath: "acln.ro/foo",
				VCS:        "git",
				VCSRepo:    "https://github.com/acln0/foo",
			},
		},
		{
			path: "acln.ro/foo/bar",
			want: vanity.ImportTag{
				ImportPath: "acln.ro/foo",
				VCS:        "git",
				VCSRepo:    "https://github.com/acln0/foo",
			},
		},
		{
			path: "acln.ro/foo/bar/baz",
			want: vanity.ImportTag{
				ImportPath: "acln.ro/foo",
				VCS:        "git",
				VCSRepo:    "https://github.com/acln0/foo",
			},
		},
		{
			path:    "acln.ro",
			wantErr: true,
		},
		{
			path:    "acln.ro/",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		req, err := http.NewRequest("GET", "https://"+tt.path, nil)
		if err != nil {
			t.Fatal(err)
		}
		tag, err := p.WildcardTagFor(req)
		if err != nil && !tt.wantErr {
			t.Errorf("%s: %v", tt.path, err)
		}
		if err == nil && tt.wantErr {
			t.Errorf("%s: got %#v, want error", tt.path, tag)
		}
		if tag != nil && !reflect.DeepEqual(*tag, tt.want) {
			t.Errorf("%s: got %#v, want %#v", tt.path, *tag, tt.want)
		}
	}
}

func TestImportTagRender(t *testing.T) {
	tag := vanity.ImportTag{
		ImportPath: "acln.ro/foo",
		VCS:        "git",
		VCSRepo:    "https://github.com/acln0/foo",
	}
	buf := new(bytes.Buffer)
	if err := tag.Render(buf); err != nil {
		t.Fatal(err)
	}
	want := `<meta name="go-import" content="acln.ro/foo git https://github.com/acln0/foo">`
	if !strings.Contains(buf.String(), want) {
		t.Fatalf("%s\ndoes not contain\n%s", buf.String(), want)
	}
}

func TestRedirectToGodoc(t *testing.T) {
	importPath := "acln.ro/foo"
	r, err := http.NewRequest("GET", importPath, nil)
	if err != nil {
		t.Fatal(err)
	}
	w := httptest.NewRecorder()
	vanity.RedirectToGodoc(w, r)
	resp := w.Result()
	if resp.StatusCode != http.StatusFound {
		t.Errorf("got %d want %d", resp.StatusCode, http.StatusFound)
	}
	gotLocation := resp.Header.Get("Location")
	wantLocation := "https://godoc.org/" + importPath
	if gotLocation != wantLocation {
		t.Errorf("got location %s want %s", gotLocation, wantLocation)
	}
}
