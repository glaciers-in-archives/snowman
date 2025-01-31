package rdf

import (
	"errors"
	"fmt"
	"strings"
)

type Term interface {
	String() string
	Type() TermType
}

type TermType int

const (
	TermBlank TermType = iota
	TermIRI
	TermLiteral
)

type Blank struct {
	id string
}

func (b Blank) Type() TermType {
	return TermBlank
}

func (b Blank) String() string {
	return b.id[2:]
}

func NewBlank(id string) (Blank, error) {
	if len(strings.TrimSpace(id)) == 0 {
		return Blank{}, errors.New("blank id")
	}
	return Blank{id: "_:" + id}, nil
}

type IRI struct {
	str string
}

func (u IRI) Type() TermType {
	return TermIRI
}

func (u IRI) String() string {
	return u.str
}

func (u IRI) Split() (namespace, localName string) {
	if hashIndex := strings.LastIndex(u.str, "#"); hashIndex != -1 {
		return u.str[:hashIndex+1], u.str[hashIndex+1:]
	}

	if schemeIndex := strings.Index(u.str, "://"); schemeIndex != -1 {
		startIndex := schemeIndex + 3
		if slashIndex := strings.LastIndex(u.str[startIndex:], "/"); slashIndex != -1 {
			actualIndex := startIndex + slashIndex
			return u.str[:actualIndex+1], u.str[actualIndex+1:]
		}
	} else {
		if slashIndex := strings.LastIndex(u.str, "/"); slashIndex != -1 {
			return u.str[:slashIndex+1], u.str[slashIndex+1:]
		}
	}

	// No hash or slash found
	return "", ""
}

func NewIRI(iri string) (IRI, error) {
	if len(iri) == 0 {
		return IRI{}, errors.New("empty IRI")
	}
	for _, r := range iri {
		if r >= '\x00' && r <= '\x20' {
			return IRI{}, fmt.Errorf("disallowed character: %q", r)
		}
		switch r {
		case '<', '>', '"', '{', '}', '|', '^', '`', '\\':
			return IRI{}, fmt.Errorf("disallowed character: %q", r)
		}
	}
	return IRI{str: iri}, nil
}

type Literal struct {
	str      string
	lang     string
	DataType IRI
}

func (l Literal) Type() TermType {
	return TermLiteral
}

func (l Literal) Lang() string {
	return l.lang
}

func (l Literal) String() string {
	return l.str
}

func NewLangLiteral(v, lang string) (Literal, error) {
	afterDash := false
	if len(lang) >= 1 && lang[0] == '-' {
		return Literal{}, errors.New("invalid language tag: must start with a letter")
	}
	for _, r := range lang {
		switch {
		case (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z'):
			continue
		case r == '-':
			if afterDash {
				return Literal{}, errors.New("invalid language tag: only one '-' allowed")
			}
			afterDash = true
		case r >= '0' && r <= '9':
			if afterDash {
				continue
			}
			fallthrough
		default:
			return Literal{}, fmt.Errorf("invalid language tag: unexpected character: %q", r)
		}
	}
	if lang[len(lang)-1] == '-' {
		return Literal{}, errors.New("invalid language tag: trailing '-' disallowed")
	}
	return Literal{str: v, lang: lang, DataType: IRI{str: "http://www.w3.org/1999/02/22-rdf-syntax-ns#langString"}}, nil
}

func NewTypedLiteral(v string, dt IRI) Literal {
	return Literal{str: v, DataType: dt}
}
