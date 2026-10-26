package catalog

import (
	"bufio"
	"log"
	"os"
	"path/filepath"
	"strings"
)

var frenchAccents = map[rune]byte{
	'à': 'a',
	'â': 'a',
	'é': 'e',
	'è': 'e',
	'ê': 'e',
	'ë': 'e',
	'î': 'i',
	'ï': 'i',
	'ô': 'o',
	'ù': 'u',
	'û': 'u',
	'ü': 'u',
	'ç': 'c',
}

// DropAccents converts s to lowercase and strips French diacritics.
// If cache is non-nil, results are memoized by original string s.
func DropAccents(s string, cache map[string]string) string {
	if cache != nil {
		if cached, ok := cache[s]; ok {
			return cached
		}
	}
	lower := strings.ToLower(s)
	if lower == s {
		lower = s
	}
	changed := false
	bs := make([]byte, 0, len(lower))
	for _, r := range lower {
		if repl, ok := frenchAccents[r]; ok {
			changed = true
			bs = append(bs, repl)
		} else {
			bs = append(bs, byte(r))
		}
	}
	res := lower
	if changed {
		res = string(bs)
	}
	if cache != nil {
		cache[s] = res
	}
	return res
}

// Person represents a set of synonymous names/nicknames for a person.
type Person struct {
	Names []string
	set   map[string]struct{}
}

// Synonyms holds the loaded person synonym groups from synonyms.txt.
type Synonyms struct {
	people   []*Person
	byName   map[string]*Person
	allNames map[string]struct{}
}

// NewSynonyms creates an empty Synonyms table.
func NewSynonyms() *Synonyms {
	return &Synonyms{
		byName:   make(map[string]*Person),
		allNames: make(map[string]struct{}),
	}
}

// LoadSynonyms reads <staticRoot>/synonyms.txt and returns a populated Synonyms table.
// Missing synonyms.txt logs a warning and returns an empty table.
func LoadSynonyms(staticRoot string) *Synonyms {
	syns := NewSynonyms()
	if staticRoot == "" {
		return syns
	}
	synsPath := filepath.Join(staticRoot, "synonyms.txt")
	fi, err := os.Open(synsPath)
	if err != nil {
		log.Printf("Error loading %s: %v", synsPath, err)
		return syns
	}
	defer fi.Close()

	scanner := bufio.NewScanner(fi)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		rawNames := strings.Split(line, ",")
		names := make([]string, 0, len(rawNames))
		set := make(map[string]struct{}, len(rawNames))
		for _, n := range rawNames {
			n = strings.TrimSpace(n)
			if n == "" {
				continue
			}
			if _, exists := set[n]; !exists {
				set[n] = struct{}{}
				names = append(names, n)
			}
		}
		if len(names) == 0 {
			continue
		}
		p := &Person{Names: names, set: set}
		syns.people = append(syns.people, p)
		for n := range set {
			syns.allNames[n] = struct{}{}
			if _, exists := syns.byName[n]; !exists {
				syns.byName[n] = p
			}
		}
	}
	if err := scanner.Err(); err != nil {
		log.Printf("Error scanning %s: %v", synsPath, err)
	}
	return syns
}

// Lookup returns the Person matching name, or nil if none.
func (s *Synonyms) Lookup(name string) *Person {
	if s == nil {
		return nil
	}
	return s.byName[name]
}

// IsName reports whether kwd is a registered person name or alias.
func (s *Synonyms) IsName(kwd string) bool {
	if s == nil {
		return false
	}
	_, ok := s.allNames[kwd]
	return ok
}
