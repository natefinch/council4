package rules

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/cards"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
)

// Source review: pinned Oracle revision 451f317f-f587-4187-9e02-6ac5d7ae9cd8
// (Oracle ID b1fbf818-6699-4f05-9a91-19aa296526bf), exact text:
// "Choose one —
// • Create a 5/5 green Beast creature token.
// • Create five 1/1 green Insect creature tokens.
// Entwine {G}{G}{G} (Choose both if you pay the entwine cost.)"
// Scryfall rulings response
// https://api.scryfall.com/cards/451f317f-f587-4187-9e02-6ac5d7ae9cd8/rulings
// retrieved 2026-11-20 returned an empty list. Official Wizards Comprehensive
// Rules effective 2026-09-25, matching testdata/card-preservation/corpus.json
// rules_edition; https://media.wizards.com/2026/downloads/MagicCompRules%2020260925.txt
// (verified local copy .cardwork/one-dozen-eyes/MagicCompRules_20260925.txt,
// SHA-256 8d860e451f20f38865b725b42d82feb714c725373dd8f3b32b8652b3eeb070ca):
// 111.1-4 require the defined token characteristics and subtype-based names;
// 601.2a-b/f-h and 700.2/700.2a place the mode and additional-cost choices in
// casting; 702.42a-b makes entwine allow both modes for one additional cost and
// resolves them in printed order; 608.2c applies each mode's instructions.
func TestOneDozenEyesModesAndEntwine(t *testing.T) {
	registry := cards.NewDefaultRegistry()
	def, err := registry.LookupOracleID("b1fbf818-6699-4f05-9a91-19aa296526bf")
	if err != nil {
		t.Fatal(err)
	}
	if def.Name != "One Dozen Eyes" {
		t.Fatalf("Oracle ID lookup returned %q", def.Name)
	}

	tests := []struct {
		name          string
		modes         []int
		colorlessMana int
		greenMana     int
		wantCast      bool
		wantTokens    map[string]int
	}{
		{
			name: "Beast mode", modes: []int{0}, colorlessMana: 5, greenMana: 1, wantCast: true,
			wantTokens: map[string]int{"Beast Token": 1},
		},
		{
			name: "Insect mode", modes: []int{1}, colorlessMana: 5, greenMana: 1, wantCast: true,
			wantTokens: map[string]int{"Insect Token": 5},
		},
		{
			name: "Entwine both modes", modes: []int{0, 1}, colorlessMana: 5, greenMana: 4, wantCast: true,
			wantTokens: map[string]int{"Beast Token": 1, "Insect Token": 5},
		},
		{
			name: "Entwine requires three additional green mana", modes: []int{0, 1},
			colorlessMana: 8, greenMana: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			spell := addCardToHand(g, game.Player1, def)
			g.Players[game.Player1].ManaPool.Add(mana.C, tt.colorlessMana)
			g.Players[game.Player1].ManaPool.Add(mana.G, tt.greenMana)
			setMainPhasePriority(g, game.Player1)

			engine := NewEngineWithPrograms(nil, registry.Programs())
			cast := engine.applyAction(g, game.Player1, action.CastSpell(spell, nil, 0, tt.modes))
			if cast != tt.wantCast {
				t.Fatalf("cast = %v, want %v", cast, tt.wantCast)
			}
			if !cast {
				if !g.Players[game.Player1].Hand.Contains(spell) || !g.Stack.IsEmpty() ||
					g.Players[game.Player1].ManaPool.Total() != tt.colorlessMana+tt.greenMana {
					t.Fatal("unpayable entwine cast changed the hand, stack, or mana pool")
				}
				return
			}
			if got := g.Players[game.Player1].ManaPool.Total(); got != 0 {
				t.Fatalf("mana remaining after casting = %d, want 0", got)
			}
			stackObject, ok := g.Stack.Peek()
			if !ok {
				t.Fatal("cast spell was not put on the stack")
			}
			if !slices.Equal(stackObject.ChosenModes, tt.modes) {
				t.Fatalf("stack modes = %v, want %v", stackObject.ChosenModes, tt.modes)
			}

			engine.resolveTopOfStack(g, &TurnLog{})
			assertOneDozenEyesTokens(t, g, tt.wantTokens)
			if !g.Players[game.Player1].Graveyard.Contains(spell) {
				t.Fatal("resolved sorcery did not go to its owner's graveyard")
			}
		})
	}
}

func assertOneDozenEyesTokens(t *testing.T, g *game.Game, want map[string]int) {
	t.Helper()
	got := make(map[string]int)
	for _, permanent := range g.Battlefield {
		if !permanent.Token {
			continue
		}
		def := permanent.TokenDef
		count, expected := want[def.Name]
		if !expected {
			t.Fatalf("unexpected token %q", def.Name)
		}
		var subtype types.Sub
		var size int
		switch def.Name {
		case "Beast Token":
			subtype, size = types.Beast, 5
		case "Insect Token":
			subtype, size = types.Insect, 1
		}
		if permanent.Controller != game.Player1 || permanent.Owner != game.Player1 ||
			!slices.Equal(def.Types, []types.Card{types.Creature}) ||
			!slices.Equal(def.Subtypes, []types.Sub{subtype}) ||
			!slices.Equal(def.Colors, []color.Color{color.Green}) ||
			!def.Power.Exists || def.Power.Val.Value != size ||
			!def.Toughness.Exists || def.Toughness.Val.Value != size {
			t.Fatalf("wrong %s characteristics or controller: %+v / %+v", def.Name, permanent, def)
		}
		got[def.Name]++
		if got[def.Name] > count {
			t.Fatalf("too many %s tokens: got at least %d, want %d", def.Name, got[def.Name], count)
		}
	}
	for name, count := range want {
		if got[name] != count {
			t.Errorf("%s count = %d, want %d", name, got[name], count)
		}
	}
}
