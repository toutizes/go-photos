package model

import (
  "log"
  "strings"
)

type Person struct {
  name_set map[string]struct{}
}
  
func NewPerson(full_names ...string) *Person {
  p := new(Person)
  p.name_set = make(map[string]struct{}, len(full_names))
  for _, n := range full_names {
    p.name_set[n] = struct{}{}
  }
  return p
}

func HasName(p *Person, name string) bool {
  _, exists := p.name_set[name]
  return exists
}

func PersonQuery(db *Database, p *Person) Query {
  queries := make([]Query, 2 * len(p.name_set))
  i := 0
  for n, _ := range p.name_set {
    queries[i] = KeywordQuery(db, n)
    i += 1
    words := strings.Fields(n)
    and_queries := make([]Query, len(words))
    for j, w := range words {
      and_queries[j] = KeywordQuery(db, w)
    }
    queries[i] = AndQuery(and_queries)
    i += 1
  }
  return OrQuery(queries)
}

var known_people = []*Person{
  NewPerson("sofia moro-devin"),
  NewPerson("pablo moro-devin"),
  NewPerson("clara moro-devin"),
  NewPerson("simon moro-devin"),
  NewPerson("antonio moro-diaz"),
  NewPerson("colombe devin", "colombe moro-devin"),
  NewPerson("joachim baudoin"),
  NewPerson("philomène baudoin"),
  NewPerson("gabriel baudoin"),
  NewPerson("olivier baudoin"),
  NewPerson("priscille devin", "priscille baudoin"),
  NewPerson("marguerite-marie sterlin"),
  NewPerson("mayeul sterlin"),
  NewPerson("amalric sterlin"),
  NewPerson("aude sterlin", "aude tomazo"),
  NewPerson("brunehilde sterlin"),
  NewPerson("ombeline sterlin"),
  NewPerson("hugues sterlin"),
  NewPerson("claire-élise devin", "claire-élise sterlin"),
  NewPerson("pierre devin"),
  NewPerson("léon devin"),
  NewPerson("virginie dubos"),
  NewPerson("anatole devin"),
  NewPerson("joseph devin"),
  NewPerson("marie bitschené"),
  NewPerson("françois devin"),
  NewPerson("zacharie devin"),
  NewPerson("marin devin"),
  NewPerson("julie devin"),
  NewPerson("stéphanie peulmeul", "stéphanie devin"),
  NewPerson("samuel devin"),
  NewPerson("élie devin"),
  NewPerson("leïla devin"),
  NewPerson("catherine doisneau"),
  NewPerson("étienne devin"),
  NewPerson("virgile devin"),
  NewPerson("lucas devin"),
  NewPerson("clémence devin"),
  NewPerson("éloi devin"),
  NewPerson("nathalie verbeck", "nathalie devin"),
  NewPerson("rémi devin"),
  NewPerson("noé devin"),
  NewPerson("lucile devin", "lucile devin-casado", "lucile casado"),
  // TODO: Liam!
  NewPerson("liam devin", "liam devin-casado", "liam casado", 
    "marie devin", "marie devin-casado", "marie casado",),
  NewPerson("katie casado", "katie devin-casado"),
  NewPerson("emmanuel devin"),
  NewPerson("coline devin"),
  NewPerson("julien devin"),
  NewPerson("catherine granger"),
  NewPerson("matthieu devin"),
  NewPerson("agnès devin"),
  NewPerson("bernard devin"),
  NewPerson("marisol devin"),
  NewPerson("cécile moncla", "cécile chartier"),
  NewPerson("bonne maman", "marcelle moncla"),
  NewPerson("bon papa", "robert moncla", "marcel moncla"),
  NewPerson("oncle pierre", "oncle caillou"),
  NewPerson("tante jehanne", "tante farce"),
  NewPerson("brigitte debast", "brigitte nagashima"),
}

func KeywordSynonymsQuery(db *Database, kwd string) Query {
  for _, p := range known_people {
    if HasName(p, kwd) {
      log.Printf("Found a known person: %v", p)
      return PersonQuery(db, p)
    }
  }
  return KeywordQuery(db, kwd)
}


