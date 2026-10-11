package rules

import (
	"testing"

	"github.com/natefinch/council4/cardgen"
	"github.com/natefinch/council4/mtg/game"
)

func referenceSubjectOriginals() []cardgen.ScryfallCard {
	return []cardgen.ScryfallCard{
		{
			ID: "c808a8ec-895c-4777-9e11-e83ce34eddef", OracleID: "bbf90463-280a-413e-b8a3-4577dcf11e54",
			Name: "Killer Instinct", Layout: "normal", ManaCost: "{4}{R}{G}", TypeLine: "Enchantment",
			Colors:     []string{"G", "R"},
			OracleText: "At the beginning of your upkeep, reveal the top card of your library. If it's a creature card, put it onto the battlefield. That creature gains haste until end of turn. Sacrifice it at the beginning of the next end step.",
		},
		{
			ID: "628d3058-d32e-438c-9a86-3e9c2be73c4a", OracleID: "e446bfa6-54b8-4dfa-98d3-023d12b7543d",
			Name: "Callous Deceiver", Layout: "normal", ManaCost: "{2}{U}", TypeLine: "Creature \u2014 Spirit",
			Colors: []string{"U"},
			Power:  new("1"), Toughness: new("3"),
			OracleText: "{1}: Look at the top card of your library.\n{2}: Reveal the top card of your library. If it's a land card, this creature gets +1/+0 and gains flying until end of turn. Activate only once each turn.",
		},
		{
			ID: "6c49d705-9b0c-4c2e-9c27-46eec35deb92", OracleID: "c3b8ea58-d84f-4c60-9f72-e4ec389eb93c",
			Name: "Feral Deceiver", Layout: "normal", ManaCost: "{3}{G}", TypeLine: "Creature \u2014 Spirit",
			Colors: []string{"G"},
			Power:  new("3"), Toughness: new("2"),
			OracleText: "{1}: Look at the top card of your library.\n{2}: Reveal the top card of your library. If it's a land card, this creature gets +2/+2 and gains trample until end of turn. Activate only once each turn.",
		},
		{
			ID: "e3d532b3-4bd6-4c1d-974d-789976117497", OracleID: "578e0577-0ad8-4bfb-a317-9684c508770a",
			Name: "Brutal Deceiver", Layout: "normal", ManaCost: "{2}{R}", TypeLine: "Creature \u2014 Spirit",
			Colors: []string{"R"},
			Power:  new("2"), Toughness: new("2"),
			OracleText: "{1}: Look at the top card of your library.\n{2}: Reveal the top card of your library. If it's a land card, this creature gets +1/+0 and gains first strike until end of turn. Activate only once each turn.",
		},
		{
			ID: "faa0c280-36e1-468b-8f88-61de394f250c", OracleID: "37042a3e-c377-4282-80f9-068e4955f8a1",
			Name: "Doombot Harbinger", Layout: "normal", ManaCost: "{2}{B}", TypeLine: "Artifact Creature \u2014 Robot Villain",
			Colors: []string{"B"},
			Power:  new("2"), Toughness: new("1"),
			OracleText: "Flying\nWhen this creature enters, you may mill four cards. (You may put the top four cards of your library into your graveyard.)\nWhen this creature dies, you may exile this card. When you do, return target creature card from your graveyard to your hand.",
		},
		{
			ID: "9184bc57-0a16-4abf-a13a-6a03a175c28a", OracleID: "24dbddad-998b-4755-b356-4c8aca3592b1",
			Name: "Flitting Guerrilla", Layout: "normal", ManaCost: "{2}{B}", TypeLine: "Creature \u2014 Faerie Rogue",
			Colors: []string{"B"},
			Power:  new("2"), Toughness: new("2"),
			OracleText: "Flying\nWhen this creature dies, each player mills two cards. Then you may exile this card. When you do, put target creature or battle card from your graveyard on top of your library. (To mill two cards, a player puts the top two cards of their library into their graveyard.)",
		},
	}
}

func compiledReferenceSubjectOriginal(t *testing.T, name string) *game.CardDef {
	t.Helper()
	for _, card := range referenceSubjectOriginals() {
		if card.Name == name {
			return compileUnlessCard(t, card)
		}
	}
	t.Fatalf("original fixture missing: %s", name)
	return nil
}

func referenceSubjectCreature(t *testing.T) *game.CardDef {
	t.Helper()
	return compileUnlessCard(t, cardgen.ScryfallCard{
		Name: "Grizzly Bears", Layout: "normal", ManaCost: "{1}{G}", TypeLine: "Creature \u2014 Bear",
		Power: new("2"), Toughness: new("2"),
	})
}
