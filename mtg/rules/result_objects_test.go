package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
)

func TestResultObjectSelectionCharacteristicAtoms(t *testing.T) {
	snapshot := game.ObjectSnapshot{
		Name: "Named Pirate", Types: []types.Card{types.Artifact, types.Creature},
		Supertypes: []types.Super{types.Legendary}, Subtypes: []types.Sub{types.Pirate},
		Colors: []color.Color{color.Blue, color.Black},
	}
	for _, test := range []struct {
		name      string
		selection game.Selection
		matches   bool
	}{
		{"wildcard", game.Selection{}, true},
		{"all types", game.Selection{RequiredTypes: []types.Card{types.Artifact, types.Creature}}, true},
		{"missing type", game.Selection{RequiredTypes: []types.Card{types.Land}}, false},
		{"type union", game.Selection{RequiredTypesAny: []types.Card{types.Land, types.Creature}}, true},
		{"missed type union", game.Selection{RequiredTypesAny: []types.Card{types.Instant, types.Sorcery}}, false},
		{"nonland", game.Selection{ExcludedTypes: []types.Card{types.Land}}, true},
		{"nonartifact", game.Selection{ExcludedTypes: []types.Card{types.Artifact}}, false},
		{"legendary", game.Selection{Supertypes: []types.Super{types.Legendary}}, true},
		{"nonlegendary", game.Selection{ExcludedSupertype: types.Legendary}, false},
		{"Pirate", game.Selection{SubtypesAny: []types.Sub{types.Pirate}}, true},
		{"non-Pirate", game.Selection{ExcludedSubtype: types.Pirate}, false},
		{"Elf", game.Selection{SubtypesAny: []types.Sub{"Elf"}}, false},
		{"blue", game.Selection{ColorsAny: []color.Color{color.Blue}}, true},
		{"nonblue", game.Selection{ExcludedColors: []color.Color{color.Blue}}, false},
		{"colorless", game.Selection{Colorless: true}, false},
		{"multicolored", game.Selection{Multicolored: true}, true},
		{"colored", game.Selection{Colored: true}, true},
		{"name", game.Selection{Name: "Named Pirate"}, true},
		{"wrong name", game.Selection{Name: "Other"}, false},
		{"alternatives", game.Selection{AnyOf: []game.Selection{
			{RequiredTypes: []types.Card{types.Land}}, {SubtypesAny: []types.Sub{types.Pirate}},
		}}, true},
	} {
		t.Run(test.name, func(t *testing.T) {
			if problems := game.ValidateResultObjectSelection(test.selection); len(problems) != 0 {
				t.Fatal(problems)
			}
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			obj := &game.StackObject{Controller: game.Player1}
			if got := resultObjectsMatchSelection(g, obj, []game.ObjectSnapshot{snapshot}, test.selection); got != test.matches {
				t.Fatalf("matched = %t, want %t", got, test.matches)
			}
			if resultObjectsMatchSelection(g, obj, nil, test.selection) {
				t.Fatal("selection matched without an actual result member")
			}
		})
	}
}

func TestResultObjectPublicationClearsFailedRepeat(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Types: []types.Card{types.Land}}})
	obj := &game.StackObject{Controller: game.Player1}
	instruction := &game.Instruction{Primitive: game.Discard{Player: game.ControllerReference(), Amount: game.Fixed(1)}, PublishResult: "discard"}
	engine := NewEngine(nil)
	engine.resolveInstructionWithChoices(g, obj, instruction, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
	if !obj.ResolutionResults["discard"].Succeeded || len(obj.ResolutionResultObjects["discard"]) != 1 {
		t.Fatal("first discard did not publish its actual result")
	}
	engine.resolveInstructionWithChoices(g, obj, instruction, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
	if obj.ResolutionResults["discard"].Succeeded || len(obj.ResolutionResultObjects["discard"]) != 0 {
		t.Fatal("failed repeated publication retained a stale matching result")
	}
}

func TestResultObjectSacrificeDoesNotPublishFallbackDiscards(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addCombatCreaturePermanent(g, game.Player2, game.KeywordNone)
	addCardToHand(g, game.Player3, &game.CardDef{CardFace: game.CardFace{Types: []types.Card{types.Land}}})
	obj := &game.StackObject{Controller: game.Player1}
	NewEngine(nil).resolveInstructionWithChoices(g, obj, &game.Instruction{
		Primitive: game.SacrificePermanents{
			PlayerGroup: game.OpponentsReference(), Amount: game.Fixed(1),
			Selection: game.Selection{RequiredTypes: []types.Card{types.Creature}},
			Fallback:  game.SacrificeFallback{Kind: game.SacrificeFallbackDiscard, Amount: game.Fixed(1)},
		},
		PublishResult: "sacrifice",
	}, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
	if !obj.ResolutionResults["sacrifice"].Succeeded || g.Players[game.Player3].Hand.Size() != 0 {
		t.Fatal("setup did not sacrifice a creature and discard the fallback land")
	}
	snapshots := obj.ResolutionResultObjects["sacrifice"]
	if len(snapshots) != 1 || resultObjectsMatchSelection(g, obj, snapshots, game.Selection{RequiredTypes: []types.Card{types.Land}}) {
		t.Fatal("fallback discard contaminated actual sacrificed result objects")
	}
}
