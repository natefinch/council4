package compiler

import (
	"testing"

	"github.com/natefinch/council4/cardgen/oracle/parser"
)

func TestCompileCorrelatedPlayerExileIsTextBlind(t *testing.T) {
	source := "Each opponent exiles a creature with the greatest power among creatures that player controls.\n" +
		"Spell mastery — If there are two or more instant and/or sorcery cards in your graveyard, Olórin's Searing Light deals damage to each opponent equal to the power of the creature they exiled."
	document, diagnostics := parser.Parse(source, parser.Context{
		CardName:         "Olórin's Searing Light",
		InstantOrSorcery: true,
	})
	if len(diagnostics) != 0 {
		t.Fatalf("parse diagnostics = %#v", diagnostics)
	}
	for i := range document.Abilities {
		document.Abilities[i].Text = "opaque"
		document.Abilities[i].Tokens = nil
		for si := range document.Abilities[i].Sentences {
			document.Abilities[i].Sentences[si].Text = "opaque"
			for ei := range document.Abilities[i].Sentences[si].Effects {
				document.Abilities[i].Sentences[si].Effects[ei].Text = "opaque"
				document.Abilities[i].Sentences[si].Effects[ei].Tokens = nil
			}
		}
	}
	compilation, diagnostics := Compile(document, Context{})
	if len(diagnostics) != 0 {
		t.Fatalf("compile diagnostics = %#v", diagnostics)
	}
	if len(compilation.Abilities) != 2 {
		t.Fatalf("abilities = %d, want 2", len(compilation.Abilities))
	}
	if !compilation.Abilities[0].Content.Effects[0].ExileEachOpponentChoosesGreatestPower {
		t.Fatal("compiled exile lost typed correlated-choice flag")
	}
	if !compilation.Abilities[1].Content.Effects[0].DamageEachOpponentCorrelatedExiledPower {
		t.Fatal("compiled damage lost typed correlated-power flag")
	}
}
