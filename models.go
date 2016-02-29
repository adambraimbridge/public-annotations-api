package main

type Annotations []annotation

type annotation struct {
	Predicate string   `json:"predicate"`
	ID        string   `json:"id"`
	APIURL    string   `json:"apiUrl`
	PrefLabel string   `json:"prefLabel"`
	LeiCode   string   `json:"leiCode"`
	Types     []string `json:"types"`
}
