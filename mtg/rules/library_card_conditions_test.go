package rules

import (
	"fmt"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

func TestLibraryCardSelectionConsumesActualObservedSubject(t *testing.T) {
	for _, tc := range []struct {
		name      string
		selection game.Selection
		qualify   func(*game.CardDef)
	}{
		{"snow", game.Selection{Supertypes: []types.Super{types.Snow}}, func(def *game.CardDef) {
			def.Supertypes = []types.Super{types.Snow}
		}},
		{"legendary", game.Selection{Supertypes: []types.Super{types.Legendary}}, func(def *game.CardDef) {
			def.Supertypes = []types.Super{types.Legendary}
		}},
		{"colorless", game.Selection{Colorless: true}, func(def *game.CardDef) {
			def.Colors = nil
		}},
		{"nonland", game.Selection{ExcludedTypes: []types.Card{types.Land}}, func(def *game.CardDef) {
			def.Types = []types.Card{types.Creature}
		}},
		{"creature or land", game.Selection{RequiredTypesAny: []types.Card{types.Creature, types.Land}}, func(def *game.CardDef) {
			def.Types = []types.Card{types.Land}
		}},
	} {
		for _, outcome := range []string{"matching", "wrong card", "empty library", "missing definition"} {
			t.Run(tc.name+"/"+outcome, func(t *testing.T) {
				g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
				g.Players[game.Player1].Life = 20
				def := &game.CardDef{CardFace: game.CardFace{
					Name: "Observed", Types: []types.Card{types.Artifact, types.Land}, Colors: []color.Color{color.Green},
				}}
				tc.qualify(def)
				addCardToLibrary(g, game.Player1, def)
				var observed id.ID
				if outcome != "empty library" {
					if outcome == "wrong card" {
						def = &game.CardDef{CardFace: game.CardFace{
							Name: "Wrong Subject", Types: []types.Card{types.Artifact, types.Land}, Colors: []color.Color{color.Green},
						}}
						if tc.name == "creature or land" {
							def.Types = []types.Card{types.Artifact}
						}
					}
					observed = addCardToLibrary(g, game.Player3, def)
				}
				obj := &game.StackObject{
					ID: g.IDGen.Next(), Controller: game.Player1,
					Targets: []game.Target{game.PlayerTarget(game.Player2), game.PlayerTarget(game.Player3)},
				}
				resolver := newEffectResolver(NewEngine(nil), g, obj, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
				resolver.resolveInstruction(&game.Instruction{Primitive: game.LookAtLibraryTop{
					Player: game.TargetPlayerReference(1), PublishLinked: "observed",
				}})
				if outcome == "missing definition" {
					g.CardInstances[observed].Def = nil
				}
				resolver.resolveInstruction(&game.Instruction{
					Primitive: game.GainLife{Player: game.ControllerReference(), Amount: game.Fixed(2)},
					Condition: opt.Val(game.EffectCondition{Condition: opt.Val(game.Condition{
						Object: opt.Val(game.LinkedObjectReference("observed")), ObjectMatches: opt.Val(tc.selection),
					})}),
				})
				want := 20
				if outcome == "matching" {
					want += 2
				}
				if got := g.Players[game.Player1].Life; got != want {
					t.Fatalf("life=%d, want %d: condition acquired a decoy, wrong, or unavailable card", got, want)
				}
			})
		}
	}
}

func TestLibraryCardGroupCapturesOnceButIndependentConditionReevaluates(t *testing.T) {
	for _, version := range []uint64{0, 7} {
		t.Run(fmt.Sprintf("version-%d", version), func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			g.Players[game.Player1].Life = 20
			cardID := addCardToLibrary(g, game.Player1, vanillaCreature("Observed", 2, 3))
			g.CardInstances[cardID].ZoneVersion = version
			obj := &game.StackObject{ID: g.IDGen.Next(), Controller: game.Player1}
			resolver := newEffectResolver(NewEngine(nil), g, obj, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
			resolver.resolveInstruction(&game.Instruction{Primitive: game.Reveal{
				Player: game.ControllerReference(), Amount: game.Fixed(1), PublishLinked: "observed",
			}})
			condition := opt.Val(game.EffectCondition{Condition: opt.Val(game.Condition{
				Object:        opt.Val(game.LinkedObjectReference("observed")),
				ObjectMatches: opt.Val(game.Selection{RequiredTypes: []types.Card{types.Creature}}),
			})})
			resolver.resolveInstruction(&game.Instruction{
				Primitive: game.MoveCard{
					Card:     game.CardReference{Kind: game.CardReferenceLinked, LinkID: "observed"},
					FromZone: zone.Library, Destination: zone.Hand,
				},
				Condition: condition, PublishCondition: "card-group",
			})
			resolver.resolveInstruction(&game.Instruction{
				Primitive:     game.GainLife{Player: game.ControllerReference(), Amount: game.Fixed(2)},
				ConditionGate: "card-group",
			})
			resolver.resolveInstruction(&game.Instruction{
				Primitive: game.GainLife{Player: game.ControllerReference(), Amount: game.Fixed(4)},
				Condition: condition,
			})
			if !g.Players[game.Player1].Hand.Contains(cardID) {
				t.Fatal("first group member did not move the actual observed card")
			}
			if got := g.Players[game.Player1].Life; got != 22 {
				t.Fatalf("life=%d, want 22: group must retain its decision; independent predicate must reject the old incarnation", got)
			}
		})
	}
}

func TestLibraryCardUnavailableSubjectDoesNotPublishOtherwise(t *testing.T) {
	for _, outcome := range []string{"matching", "nonmatching", "empty", "missing definition", "reincarnated"} {
		t.Run(outcome, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			g.Players[game.Player1].Life = 20
			var cardID id.ID
			if outcome != "empty" {
				def := vanillaCreature("Observed", 2, 3)
				if outcome == "nonmatching" {
					def.Types = []types.Card{types.Artifact}
				}
				cardID = addCardToLibrary(g, game.Player1, def)
			}
			obj := &game.StackObject{ID: g.IDGen.Next(), Controller: game.Player1}
			resolver := newEffectResolver(NewEngine(nil), g, obj, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
			resolver.resolveInstruction(&game.Instruction{Primitive: game.Reveal{
				Player: game.ControllerReference(), Amount: game.Fixed(1), PublishLinked: "observed",
			}})
			switch outcome {
			case "matching", "nonmatching", "empty":
			case "missing definition":
				g.CardInstances[cardID].Def = nil
			case "reincarnated":
				g.CardInstances[cardID].ZoneVersion++
			default:
				t.Fatalf("unknown observation outcome %q", outcome)
			}
			resolver.resolveInstruction(&game.Instruction{
				Primitive: game.GainLife{Player: game.ControllerReference(), Amount: game.Fixed(2)},
				Condition: opt.Val(game.EffectCondition{Condition: opt.Val(game.Condition{
					Object:        opt.Val(game.LinkedObjectReference("observed")),
					ObjectMatches: opt.Val(game.Selection{RequiredTypes: []types.Card{types.Creature}}),
				})}),
				PublishCondition: "card-group",
			})
			resolver.resolveInstruction(&game.Instruction{
				Primitive:     game.GainLife{Player: game.ControllerReference(), Amount: game.Fixed(4)},
				ConditionGate: "card-group", ConditionGateNegate: true,
			})
			want := 20
			switch outcome {
			case "matching":
				want = 22
			case "nonmatching":
				want = 24
			default:
			}
			if got := g.Players[game.Player1].Life; got != want {
				t.Fatalf("life=%d, want %d: unavailable subject is not a false observed predicate", got, want)
			}
		})
	}
}
