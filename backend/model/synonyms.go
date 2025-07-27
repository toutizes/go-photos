package model

import (
	"bufio"
	"log"
	"os"
	"path"
	"strings"
)

type Person struct {
	name_set map[string]struct{}
}

func newPerson(full_names ...string) *Person {
	// log.Printf("New person: %v\n", full_names[0])
	p := new(Person)
	p.name_set = make(map[string]struct{}, len(full_names))
	for _, n := range full_names {
		p.name_set[n] = struct{}{}
	}
  for k, v := range p.name_set {
		all_names[k] = v
	}
	return p
}

func hasName(p *Person, name string) bool {
	_, exists := p.name_set[name]
	return exists
}

func personQuery(db *Database, p *Person) Query {
	queries := make([]Query, 2*len(p.name_set))
	i := 0
	for n, _ := range p.name_set {
		queries[i] = FullKeywordQuery(db, n)
		i += 1
		words := strings.Fields(n)
		and_queries := make([]Query, len(words))
		for j, w := range words {
			and_queries[j] = FullKeywordQuery(db, w)
		}
		queries[i] = AndQuery(and_queries)
		i += 1
	}
	return OrQuery(queries)
}

var known_people = []*Person{}
var all_names = map[string]struct{}{}

func LoadSynonyms(root string) {
	syns_path := path.Join(root, "synonyms.txt")
	fi, err := os.Open(syns_path)
	if err != nil {
		log.Printf("Error loading %s: %s\n", syns_path, err.Error())
		return
	}
	defer fi.Close()
	scanner := bufio.NewScanner(fi)
	for scanner.Scan() {
		line := scanner.Text()
		names := strings.Split(line, ",")
		if len(names) == 0 {
			continue
		}
		p := newPerson(names...)
		known_people = append(known_people, p)
	}
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
}

func KeywordSynonymsQuery(db *Database, kwd string) Query {
	for _, p := range known_people {
		if hasName(p, kwd) {
			// log.Printf("Found a known person: %v", p)
			return personQuery(db, p)
		}
	}
	return KeywordQuery(db, kwd)
}

func IsName(db *Database, kwd string) bool {
	_, exists := all_names[kwd]
	return exists
}
