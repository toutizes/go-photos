package model

import (
  "log"
)

type Person struct {
  first_name string
  last_name1 string
  last_name2 string
  keywords map[string]struct{}
}
  
func NewPerson(first_name string, last_name1 string,  last_name2 string) Person {
  keywords := map[string]struct{}{}
  keywords[first_name + " " + last_name1] = struct{}{}
  if last_name2 !=  "" {
    keywords[first_name + " " + last_name2] = struct{}{}
  }
  return Person{
    first_name: first_name,
    last_name1: last_name1,
    last_name2: last_name2,
    keywords: keywords,
  }
}

func PersonQuery(db *Database, p *Person) Query {
  queries := []Query{
    KeywordQuery(db, p.first_name + " " + p.last_name1), 
    AndQuery([]Query{
      KeywordQuery(db, p.first_name), 
      KeywordQuery(db, p.last_name1),
    }),
  }
  if p.last_name2 != "" {
    queries = append(queries, 
      KeywordQuery(db, p.first_name + " " + p.last_name2),
      AndQuery([]Query{
        KeywordQuery(db, p.first_name), 
        KeywordQuery(db, p.last_name2),
      }),
    )
  }
  return OrQuery(queries)
}

var known_people = []Person{
  NewPerson("sofia", "moro-devin", ""),
  NewPerson("pablo", "moro-devin", ""),
  NewPerson("clara", "moro-devin", ""),
  NewPerson("simon", "moro-devin", ""),
  NewPerson("antonio", "moro-diaz", ""),
  NewPerson("colombe", "devin", "moro-devin"),
  NewPerson("joachim", "baudoin", ""),
  NewPerson("philomène", "baudoin", ""),
  NewPerson("gabriel", "baudoin", ""),
  NewPerson("olivier", "baudoin", ""),
  NewPerson("priscille", "devin", "baudoin"),
  NewPerson("marguerite-marie", "sterlin", ""),
  NewPerson("mayeul", "sterlin", ""),
  NewPerson("amalric", "sterlin", ""),
  NewPerson("aude", "sterlin", "tomazo"),
  NewPerson("brunehilde", "sterlin", ""),
  NewPerson("ombeline", "sterlin", ""),
  NewPerson("hugues", "sterlin", ""),
  NewPerson("claire-élise", "devin", "sterlin"),
  NewPerson("pierre", "devin", ""),
  NewPerson("léon", "devin", ""),
  NewPerson("virginie", "dubos", ""),
  NewPerson("anatole", "devin", ""),
  NewPerson("joseph", "devin", ""),
  NewPerson("marie", "bitschené", ""),
  NewPerson("françois", "devin", ""),
  NewPerson("zacharie", "devin", ""),
  NewPerson("marin", "devin", ""),
  NewPerson("julie", "devin", ""),
  NewPerson("stéphanie", "peulmeul", "devin"),
  NewPerson("samuel", "devin", ""),
  NewPerson("élie", "devin", ""),
  NewPerson("leïla", "devin", ""),
  NewPerson("catherine", "doisneau", ""),
  NewPerson("étienne", "devin", ""),
  NewPerson("virgile", "devin", ""),
  NewPerson("lucas", "devin", ""),
  NewPerson("clémence", "devin", ""),
  NewPerson("éloi", "devin", ""),
  NewPerson("nathalie", "verbeck", "devin"),
  NewPerson("rémi", "devin", ""),
  NewPerson("noé", "devin", ""),
  NewPerson("lucile", "devin", "devin-casado"),
  // TODO: Liam!
  NewPerson("marie", "devin", "devin-casado"),
  NewPerson("katie", "casado", "devin-casado"),
  NewPerson("emmanuel", "devin", ""),
  NewPerson("coline", "devin", ""),
  NewPerson("julien", "devin", ""),
  NewPerson("catherine", "granger", ""),
  NewPerson("matthieu", "devin", ""),
  NewPerson("agnès", "devin", ""),
  NewPerson("bernard", "devin", ""),
  NewPerson("marisol", "devin", ""),
}

func KeywordSynonymsQuery(db *Database, kwd string) Query {
  for _, p := range known_people {
    if _, exists := p.keywords[kwd]; exists {
      log.Printf("Found a known person: %v", p)
      return PersonQuery(db, &p)
    }
  }
  return KeywordQuery(db, kwd)
}


