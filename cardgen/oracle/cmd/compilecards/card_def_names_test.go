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

var SaberAnts = newSaberAnts

func newSaberAnts() *game.CardDef {
	return &game.CardDef{}
}

var saberAntsToken = newSaberAntsToken()

func newSaberAntsToken() *game.CardDef {
	return &game.CardDef{}
}

var GiantSpider = newGiantSpider

func newGiantSpider() *game.CardDef {
	return &game.CardDef{}
}
`
	file, err := parser.ParseFile(token.NewFileSet(), "a.go", source, 0)
	if err != nil {
		t.Fatal(err)
	}
	got := cardDefNames(file)
	slices.Sort(got)
	// Card builders are returned; the token builder (referenced by a call) is not.
	want := []string{"newGiantSpider", "newSaberAnts"}
	if !slices.Equal(got, want) {
		t.Fatalf("cardDefNames = %v, want %v", got, want)
	}
}
