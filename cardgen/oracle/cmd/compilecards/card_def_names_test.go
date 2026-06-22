package main

import (
	"go/parser"
	"go/token"
	"slices"
	"testing"
)

func TestCardDefNamesRecognizesBuilderFunctions(t *testing.T) {
	t.Parallel()
	source := `package a

import "github.com/natefinch/council4/mtg/game"

var SaberAnts = newSaberAnts()

func newSaberAnts() *game.CardDef {
	return &game.CardDef{}
}

var saberAntsToken = newSaberAntsToken()

func newSaberAntsToken() *game.CardDef {
	return &game.CardDef{}
}

var LegacyLiteral = &game.CardDef{}
`
	file, err := parser.ParseFile(token.NewFileSet(), "a.go", source, 0)
	if err != nil {
		t.Fatal(err)
	}
	got := cardDefNames(file)
	slices.Sort(got)
	want := []string{"LegacyLiteral", "SaberAnts"}
	if !slices.Equal(got, want) {
		t.Fatalf("cardDefNames = %v, want %v", got, want)
	}
}
