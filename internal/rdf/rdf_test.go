package rdf

import (
	"testing"
)

func TestNewBlank(t *testing.T) {
	tests := []struct {
		name    string
		id      string
		want    string
		wantErr bool
	}{
		{
			name:    "valid blank node",
			id:      "node1",
			want:    "node1",
			wantErr: false,
		},
		{
			name:    "empty id",
			id:      "",
			wantErr: true,
		},
		{
			name:    "whitespace only id",
			id:      "   ",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewBlank(tt.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewBlank() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got.String() != tt.want {
				t.Errorf("NewBlank() = %v, want %v", got.String(), tt.want)
			}
			if !tt.wantErr && got.Type() != TermBlank {
				t.Errorf("NewBlank() type = %v, want %v", got.Type(), TermBlank)
			}
		})
	}
}

func TestNewIRI(t *testing.T) {
	tests := []struct {
		name    string
		iri     string
		want    string
		wantErr bool
	}{
		{
			name:    "valid IRI",
			iri:     "http://example.org/resource",
			want:    "http://example.org/resource",
			wantErr: false,
		},
		{
			name:    "empty IRI",
			iri:     "",
			wantErr: true,
		},
		{
			name:    "IRI with illegal character <",
			iri:     "http://example.org/<resource>",
			wantErr: true,
		},
		{
			name:    "IRI with illegal character >",
			iri:     "http://example.org/>resource",
			wantErr: true,
		},
		{
			name:    "IRI with control character",
			iri:     "http://example.org/resource\n",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewIRI(tt.iri)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewIRI() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got.String() != tt.want {
				t.Errorf("NewIRI() = %v, want %v", got.String(), tt.want)
			}
			if !tt.wantErr && got.Type() != TermIRI {
				t.Errorf("NewIRI() type = %v, want %v", got.Type(), TermIRI)
			}
		})
	}
}

func TestIRISplit(t *testing.T) {
	tests := []struct {
		name       string
		iri        string
		wantPrefix string
		wantSuffix string
	}{
		{
			name:       "split on slash",
			iri:        "http://example.org/resource",
			wantPrefix: "http://example.org/",
			wantSuffix: "resource",
		},
		{
			name:       "split on hash",
			iri:        "http://example.org#resource",
			wantPrefix: "http://example.org#",
			wantSuffix: "resource",
		},
		{
			name:       "no split character",
			iri:        "http://example.org",
			wantPrefix: "",
			wantSuffix: "",
		},
		{
			name:       "no split characters non http scheme",
			iri:        "file://path",
			wantPrefix: "",
			wantSuffix: "",
		},
		{
			name:       "multiple split characters",
			iri:        "http://example.org/path#fragment",
			wantPrefix: "http://example.org/path#",
			wantSuffix: "fragment",
		},
		{
			name:       "multiple split characters non http scheme",
			iri:        "ssh://example.org/path#fragment",
			wantPrefix: "ssh://example.org/path#",
			wantSuffix: "fragment",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			iri, _ := NewIRI(tt.iri)
			gotPrefix, gotSuffix := iri.Split()
			if gotPrefix != tt.wantPrefix {
				t.Errorf("IRI.Split() prefix = %v, want %v", gotPrefix, tt.wantPrefix)
			}
			if gotSuffix != tt.wantSuffix {
				t.Errorf("IRI.Split() suffix = %v, want %v", gotSuffix, tt.wantSuffix)
			}
		})
	}
}

func TestNewLangLiteral(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		lang    string
		want    string
		wantErr bool
	}{
		{
			name:    "valid language tag",
			value:   "Hello",
			lang:    "en",
			want:    "Hello",
			wantErr: false,
		},
		{
			name:    "valid language tag with region",
			value:   "Hello",
			lang:    "en-US",
			want:    "Hello",
			wantErr: false,
		},
		{
			name:    "invalid language tag - starts with dash",
			value:   "Hello",
			lang:    "-en",
			wantErr: true,
		},
		{
			name:    "invalid language tag - multiple dashes",
			value:   "Hello",
			lang:    "en--US",
			wantErr: true,
		},
		{
			name:    "invalid language tag - ends with dash",
			value:   "Hello",
			lang:    "en-",
			wantErr: true,
		},
		{
			name:    "invalid language tag - special character",
			value:   "Hello",
			lang:    "en$US",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewLangLiteral(tt.value, tt.lang)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewLangLiteral() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if got.String() != tt.want {
					t.Errorf("NewLangLiteral() = %v, want %v", got.String(), tt.want)
				}
				if got.Lang() != tt.lang {
					t.Errorf("NewLangLiteral() lang = %v, want %v", got.Lang(), tt.lang)
				}
				if got.Type() != TermLiteral {
					t.Errorf("NewLangLiteral() type = %v, want %v", got.Type(), TermLiteral)
				}
			}
		})
	}
}

func TestNewTypedLiteral(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		datatype string
		want     string
	}{
		{
			name:     "string literal",
			value:    "Hello",
			datatype: "http://www.w3.org/2001/XMLSchema#string",
			want:     "Hello",
		},
		{
			name:     "integer literal",
			value:    "42",
			datatype: "http://www.w3.org/2001/XMLSchema#integer",
			want:     "42",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dt, _ := NewIRI(tt.datatype)
			got := NewTypedLiteral(tt.value, dt)
			if got.String() != tt.want {
				t.Errorf("NewTypedLiteral() = %v, want %v", got.String(), tt.want)
			}
			if got.DataType.String() != tt.datatype {
				t.Errorf("NewTypedLiteral() datatype = %v, want %v", got.DataType.String(), tt.datatype)
			}
			if got.Type() != TermLiteral {
				t.Errorf("NewTypedLiteral() type = %v, want %v", got.Type(), TermLiteral)
			}
		})
	}
}
