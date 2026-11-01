package rules

// Source review: Engine gap #3442. Cascade and discover judge an exiled split
// card by its combined characteristics. Pinned corpus
// 6da9ced0e19bd384771ab8f14051cc45a8dc90302082b875fa6e039d2399a1b4 and Scryfall
// rulings retrieved for this unit:
//   - Bloodbraid Elf e2f12f6f-9383-47e6-a44f-2834ad130e51 {2}{R}{G}, mana value
//     4: "Haste" / "Cascade (When you cast this spell, exile cards from the top
//     of your library until you exile a nonland card that costs less. You may
//     cast it without paying its mana cost. Put the exiled cards on the bottom in
//     a random order.)". Ruling 2021-06-18: "The mana value of a split card is
//     determined by the combined mana cost of its two halves. If cascade allows
//     you to cast a split card, you may cast either half but not both halves."
//   - Assault // Battery 9b2ca0a2-0a18-4d9e-b953-52018af3c65b {R} // {3}{G}, mana
//     value 5: Assault (Sorcery) "Assault deals 2 damage to any target." /
//     Battery (Sorcery) "Create a 3/3 green Elephant creature token." Rulings
//     2022-12-08: a split card's characteristics are a combination of its two
//     halves while it is not on the stack (mana value 5 in a library); an effect
//     that lets you cast a spell with certain characteristics considers only the
//     half being cast.
//   - Llanowar Elves 6a0b230b-d391-4998-a3f7-7b158a0ec2cd {G}: "{T}: Add {G}."
// Comprehensive Rules (verified copy effective 2026-09-25): 709.4 (in every zone
// except the stack, a split card's characteristics are its two halves
// combined), 709.4b (mana value from the combined mana cost), 709.4c (each card
// type of either half), 709.3 (a player chooses which half to cast), 702.85a
// (cascade exiles until a nonland card whose mana value is less, and the
// resulting spell's mana value must also be less) and 701.57a (discover N exiles
// until a nonland card with mana value N or less; the resulting spell's mana
// value must also be N or less).

import (
	"slices"
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/mana"
)

// assaultBattery returns the registered Assault // Battery split card, whose
// front half is registered under the name Assault.
func (f cascadeTriggerFixture) assaultBattery(t *testing.T) *game.CardDef {
	t.Helper()
	def := f.def(t, "Assault")
	if def.Layout != game.LayoutSplit || !def.Alternate.Exists || def.Alternate.Val.Name != "Battery" ||
		def.ManaValue() != 1 || def.Alternate.Val.ManaValue() != 4 {
		t.Fatalf("Assault definition = %+v, want split card Assault // Battery", def)
	}
	return def
}

// splitHalfChoiceAgent answers the free-cast choice with the first option
// whose label starts with half, and accepts every "may" choice.
type splitHalfChoiceAgent struct {
	half     string
	requests []game.ChoiceRequest
}

func (*splitHalfChoiceAgent) ChooseAction(PlayerObservation, []action.Action) action.Action {
	return action.Pass()
}

func (a *splitHalfChoiceAgent) ChooseChoice(_ PlayerObservation, request game.ChoiceRequest) []int {
	if request.Kind == game.ChoiceResolution && strings.Contains(request.Prompt, "without paying its mana cost") &&
		!strings.HasSuffix(request.Prompt, "?") {
		a.requests = append(a.requests, request)
		for _, option := range request.Options {
			if strings.HasPrefix(option.Label, a.half) {
				return []int{option.Index}
			}
		}
		return nil
	}
	if strings.HasSuffix(request.Prompt, "?") {
		return []int{1}
	}
	return request.DefaultSelection
}

func optionLabelsWithPrefix(request game.ChoiceRequest, prefix string) int {
	count := 0
	for _, option := range request.Options {
		if strings.HasPrefix(option.Label, prefix) {
			count++
		}
	}
	return count
}

// TestBloodbraidElfCascadeSkipsSplitCardByCombinedManaValue is the #3442
// scenario: Assault // Battery has mana value 5 in exile (CR 709.4b), not
// Assault's 1, so Bloodbraid Elf's cascade (mana value 4) passes over it and
// stops at Llanowar Elves, which is cast; Assault // Battery is bottomed.
func TestBloodbraidElfCascadeSkipsSplitCardByCombinedManaValue(t *testing.T) {
	f := newCascadeTriggerFixture(t)
	g := f.game
	elvesID := addCardToLibrary(g, game.Player1, f.def(t, "Llanowar Elves"))
	splitID := addCardToLibrary(g, game.Player1, f.assaultBattery(t))
	bloodbraidID, _ := f.castBloodbraid(t)
	agent := &cascadeCastChoiceAgent{}
	agents := [game.NumPlayers]PlayerAgent{game.Player1: agent}

	resolveCascadeTrigger(t, f.engine, g, &agents)

	if len(agent.requests) != 1 || optionLabelsWithPrefix(agent.requests[0], "Llanowar Elves") != 1 ||
		optionLabelsWithPrefix(agent.requests[0], "Assault") != 0 || optionLabelsWithPrefix(agent.requests[0], "Battery") != 0 {
		t.Fatalf("cascade requests = %+v, want one offer to cast Llanowar Elves only", agent.requests)
	}
	objects := g.Stack.Objects()
	if len(objects) != 2 || objects[0].SourceID != bloodbraidID || objects[1].SourceID != elvesID {
		t.Fatalf("stack = %+v, want Llanowar Elves cast above Bloodbraid Elf", objects)
	}
	if got := g.Players[game.Player1].Library.All(); !sameCardIDs(got, []id.ID{splitID}) {
		t.Fatalf("library = %v, want only Assault // Battery bottomed", got)
	}
	if g.Players[game.Player1].Exile.Size() != 0 {
		t.Fatal("cascade left cards in exile")
	}
}

// TestCascadeHitsSplitCardAndCastsEitherHalf proves a split card whose combined
// mana value is less than the cascade spell's is the hit, and the caster may
// cast either half (but not both) as the resulting spell (CR 702.85a, 709.3).
func TestCascadeHitsSplitCardAndCastsEitherHalf(t *testing.T) {
	for _, half := range []string{"Assault", "Battery"} {
		t.Run(half, func(t *testing.T) {
			f := newCascadeTriggerFixture(t)
			g := f.game
			below := addCardToLibrary(g, game.Player1, f.def(t, "Llanowar Elves"))
			splitID := addCardToLibrary(g, game.Player1, f.assaultBattery(t))
			cascadeID := addCardToHand(g, game.Player1, cascadeSpell(6))
			g.Players[game.Player1].ManaPool.Add(mana.C, 6)
			if !f.engine.applyAction(g, game.Player1, action.CastSpell(cascadeID, nil, 0, nil)) {
				t.Fatal("cascade spell cast failed")
			}
			agent := &splitHalfChoiceAgent{half: half}
			agents := [game.NumPlayers]PlayerAgent{game.Player1: agent}

			resolveCascadeTrigger(t, f.engine, g, &agents)

			if len(agent.requests) != 1 || optionLabelsWithPrefix(agent.requests[0], "Battery") != 1 ||
				optionLabelsWithPrefix(agent.requests[0], "Assault") == 0 {
				t.Fatalf("cascade requests = %+v, want offers to cast Assault or Battery", agent.requests)
			}
			top, ok := g.Stack.Peek()
			wantFace := game.FaceFront
			if half == "Battery" {
				wantFace = game.FaceAlternate
			}
			if !ok || g.Stack.Size() != 2 || top.SourceID != splitID || top.Face != wantFace {
				t.Fatalf("stack = %+v, want %s cast from Assault // Battery above the cascade spell", g.Stack.Objects(), half)
			}
			if manaValue, ok := stackObjectManaValue(g, top); !ok || manaValue != map[string]int{"Assault": 1, "Battery": 4}[half] {
				t.Fatalf("cast %s mana value = %d, %v; want only that half's mana value", half, manaValue, ok)
			}
			if got := g.Players[game.Player1].Library.All(); !slices.Equal(got, []id.ID{below}) || g.Players[game.Player1].Exile.Size() != 0 {
				t.Fatalf("library = %v, want Llanowar Elves untouched below the split hit", got)
			}
		})
	}
}

// TestDiscoverJudgesSplitCardByCombinedManaValue proves discover uses the same
// exiled-card characteristics (CR 701.57a, 709.4b): discover 4 passes over
// Assault // Battery (mana value 5) to Llanowar Elves, while discover 5 stops
// at it and may cast Battery, whose own mana value 4 is at most 5.
func TestDiscoverJudgesSplitCardByCombinedManaValue(t *testing.T) {
	t.Run("discover 4 skips", func(t *testing.T) {
		f := newCascadeTriggerFixture(t)
		g := f.game
		elvesID := addCardToLibrary(g, game.Player1, f.def(t, "Llanowar Elves"))
		splitID := addCardToLibrary(g, game.Player1, f.assaultBattery(t))
		addEffectSpellToStack(g, game.Player1, game.DiscoverCards{Amount: game.Fixed(4)}, nil)
		agents := [game.NumPlayers]PlayerAgent{game.Player1: &choiceOnlyAgent{choices: [][]int{{0}}}}

		f.engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

		if !g.Players[game.Player1].Hand.Contains(elvesID) || g.Stack.Size() != 0 {
			t.Fatal("discover 4 did not stop at Llanowar Elves past Assault // Battery")
		}
		if got := g.Players[game.Player1].Library.All(); !slices.Equal(got, []id.ID{splitID}) {
			t.Fatalf("library = %v, want Assault // Battery bottomed", got)
		}
	})
	t.Run("discover 5 casts Battery", func(t *testing.T) {
		f := newCascadeTriggerFixture(t)
		g := f.game
		elvesID := addCardToLibrary(g, game.Player1, f.def(t, "Llanowar Elves"))
		splitID := addCardToLibrary(g, game.Player1, f.assaultBattery(t))
		addEffectSpellToStack(g, game.Player1, game.DiscoverCards{Amount: game.Fixed(5)}, nil)
		agent := &splitHalfChoiceAgent{half: "Battery"}
		agents := [game.NumPlayers]PlayerAgent{game.Player1: agent}

		f.engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

		top, ok := g.Stack.Peek()
		if !ok || g.Stack.Size() != 1 || top.SourceID != splitID || top.Face != game.FaceAlternate {
			t.Fatalf("stack = %+v, want Battery discovered and cast", g.Stack.Objects())
		}
		if len(agent.requests) != 1 || optionLabelsWithPrefix(agent.requests[0], "Assault") == 0 {
			t.Fatalf("discover requests = %+v, want Assault also offered", agent.requests)
		}
		if got := g.Players[game.Player1].Library.All(); !slices.Equal(got, []id.ID{elvesID}) {
			t.Fatalf("library = %v, want Llanowar Elves untouched", got)
		}
	})
}
