package rules

import (
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/cards"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

func castForFreeInstruction() *game.Instruction {
	return &game.Instruction{
		Primitive: game.CastForFree{
			Player: game.ControllerReference(),
			Selection: game.Selection{
				ExcludedTypes: []types.Card{types.Land},
				ManaValue:     opt.Val(compare.Int{Op: compare.LessOrEqual, Value: 5}),
			},
			Zone: zone.Hand,
		},
	}
}

// CR 601.2a/c/f/h, 118.9d (2026-11-06): Lightning Axe still needs its
// additional cost. The unpayable {5} branch is never offered, and choosing a
// target observes the same announced object that becomes cast.
func TestFreeLightningAxeCompletableAdditionalCost(t *testing.T) {
	for _, multiple := range []bool{false, true} {
		t.Run(map[bool]string{false: "per card", true: "multiple cards"}[multiple], func(t *testing.T) {
			e, g := newEffectiveCascadeGame()
			reg := cards.NewDefaultRegistry()
			grove := addCombatPermanent(g, game.Player1, reg.Lookup("Waterlogged Grove"))
			bears := addCombatPermanent(g, game.Player2, reg.Lookup("Grizzly Bears"))
			forest := addCardToHand(g, game.Player1, reg.Lookup("Forest"))
			from := zone.Exile
			axe := addCardToExile(g, game.Player1, reg.Lookup("Lightning Axe"))
			if multiple {
				g.Players[game.Player1].Exile.Remove(axe)
				g.Players[game.Player1].Hand.Add(axe)
				from = zone.Hand
			}
			life, priority := g.Players[game.Player1].Life, g.Turn.PriorityPlayer
			var announced *game.StackObject
			targets := 0
			agent := &controlScript{choice: func(_ PlayerObservation, request game.ChoiceRequest) []int {
				for _, option := range request.Options {
					if strings.Contains(option.Label, "Pay {5}") {
						t.Fatal("unpayable additional-cost alternative was offered")
					}
				}
				if request.Kind == game.ChoiceTarget {
					targets++
					top, ok := g.Stack.Peek()
					if !ok || top.SourceID != axe || !top.Casting || castSourceContains(g.Players[game.Player1], axe, from) {
						t.Fatalf("target choice did not observe the announced Axe: %+v", top)
					}
					announced = top
					for _, option := range request.Options {
						if len(option.Targets) == 1 && option.Targets[0] == game.PermanentTarget(bears.ObjectID) {
							return []int{option.Index}
						}
					}
					t.Fatal("legal Bears target not offered")
				}
				return request.DefaultSelection
			}}
			agents := reboundKickerAgents(agent)
			if multiple {
				e.resolveInstructionWithChoices(g, &game.StackObject{Controller: game.Player1}, castForFreeInstruction(), agents, &TurnLog{})
			} else if !e.chooseAndCastFreeSpellFromSourceMay(g, g.Players[game.Player1], game.Player1, axe, from,
				game.Selection{}, opt.V[int]{}, false, true, agents, &TurnLog{}) {
				t.Fatal("legal discard route was not cast")
			}
			top, _ := g.Stack.Peek()
			if targets != 1 || top != announced || top.Casting || g.Stack.Size() != 1 || spellCastEventCount(g) != 1 ||
				!g.Players[game.Player1].Graveyard.Contains(forest) || g.Players[game.Player1].Hand.Contains(forest) ||
				grove.Tapped || g.Players[game.Player1].Life != life || g.Turn.PriorityPlayer != priority {
				t.Fatalf("cast=%+v targets=%d: want one completed discard-paid cast without mana or priority", top, targets)
			}
		})
	}
}

// TestCastForFreeCastsChosenSpell verifies the controller's chosen matching
// spell leaves hand for the stack while a land and an over-costed spell stay.
func TestCastForFreeCastsChosenSpell(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Source",
		Types: []types.Card{types.Artifact},
	}})
	obj := triggeredObjFor(source)

	spell := addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:     "Cheap Bolt",
		Types:    []types.Card{types.Instant},
		ManaCost: opt.Val(cost.Mana{cost.O(3)}),
	}})
	land := addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Forest",
		Types: []types.Card{types.Land},
	}})
	expensive := addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:     "Big Spell",
		Types:    []types.Card{types.Sorcery},
		ManaCost: opt.Val(cost.Mana{cost.O(7)}),
	}})

	agents := [game.NumPlayers]PlayerAgent{game.Player1: defaultChoiceAgent{}}
	engine.resolveInstructionWithChoices(g, obj, castForFreeInstruction(), agents, &TurnLog{})

	if g.Players[game.Player1].Hand.Contains(spell) {
		t.Fatal("chosen spell still in hand")
	}
	if !g.Players[game.Player1].Hand.Contains(land) {
		t.Fatal("land was removed from hand")
	}
	if !g.Players[game.Player1].Hand.Contains(expensive) {
		t.Fatal("over-costed spell was removed from hand")
	}
	if g.Stack.Size() != 1 {
		t.Fatalf("stack size = %d, want 1 (the free-cast spell)", g.Stack.Size())
	}
}

// TestCastForFreeWithNoEligibleSpellDoesNothing verifies that with only an
// ineligible (over-costed) spell in hand, nothing is cast.
func TestCastForFreeWithNoEligibleSpellDoesNothing(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Source",
		Types: []types.Card{types.Artifact},
	}})
	obj := triggeredObjFor(source)

	expensive := addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:     "Big Spell",
		Types:    []types.Card{types.Sorcery},
		ManaCost: opt.Val(cost.Mana{cost.O(7)}),
	}})

	agents := [game.NumPlayers]PlayerAgent{game.Player1: defaultChoiceAgent{}}
	engine.resolveInstructionWithChoices(g, obj, castForFreeInstruction(), agents, &TurnLog{})

	if !g.Players[game.Player1].Hand.Contains(expensive) {
		t.Fatal("over-costed spell was removed from hand")
	}
	if g.Stack.Size() != 0 {
		t.Fatalf("stack size = %d, want 0", g.Stack.Size())
	}
}
