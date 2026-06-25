package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// whipReanimationInstructions mirrors the lowering of "Return target creature
// card from your graveyard to the battlefield. It gains haste. Exile it at the
// beginning of the next end step." The reanimation publishes the entered
// permanent under a linked key; the haste grant reads that linked object and
// republishes it under a second key; and the delayed end-step trigger exiles
// that linked object.
func whipReanimationInstructions() []game.Instruction {
	const (
		hasteKey game.LinkedKey = "gain-keyword-1"
		exileKey game.LinkedKey = "delayed-exile-2"
	)
	return []game.Instruction{
		{
			Primitive: game.PutOnBattlefield{
				Source:        game.CardBattlefieldSource(game.CardReference{Kind: game.CardReferenceTarget}),
				Recipient:     opt.Val(game.ControllerReference()),
				PublishLinked: hasteKey,
			},
		},
		{
			Primitive: game.ApplyContinuous{
				Object: opt.Val(game.LinkedObjectReference(string(hasteKey))),
				ContinuousEffects: []game.ContinuousEffect{{
					Layer:       game.LayerAbility,
					AddKeywords: []game.Keyword{game.Haste},
				}},
				Duration:      game.DurationPermanent,
				PublishLinked: exileKey,
			},
		},
		{
			Primitive: game.CreateDelayedTrigger{Trigger: game.DelayedTriggerDef{
				Timing: game.DelayedAtBeginningOfNextEndStep,
				Content: game.Mode{
					Sequence: []game.Instruction{{
						Primitive: game.Exile{Object: game.LinkedObjectReference(string(exileKey))},
					}},
				}.Ability(),
			}},
		},
	}
}

func addWhipReanimationSpell(g *game.Game, target game.Target) id.ID {
	sourceID := addInstructionSpellToStackForController(g, game.Player1, whipReanimationInstructions(), []game.Target{target})
	card, _ := g.GetCardInstance(sourceID)
	card.Def.SpellAbility.Val.Modes[0].Targets = []game.TargetSpec{{
		MinTargets: 1,
		MaxTargets: 1,
		Allow:      game.TargetAllowCard,
		TargetZone: zone.Graveyard,
		Selection: opt.Val(game.Selection{
			RequiredTypes: []types.Card{types.Creature},
		}),
	}}
	return sourceID
}

func TestWhipReanimationGrantsHasteAndExilesAtNextEndStep(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)

	cardID := addCardToGraveyard(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:      "Reanimation Target",
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(game.PT{Value: 2}),
		Toughness: opt.Val(game.PT{Value: 2}),
	}})

	addWhipReanimationSpell(g, currentCardTarget(t, g, cardID))
	engine.resolveTopOfStack(g, &TurnLog{})

	permanent, ok := reanimatedPermanent(g, cardID)
	if !ok {
		t.Fatal("reanimation target was not returned to the battlefield")
	}
	if !hasKeyword(g, permanent, game.Haste) {
		t.Fatal("reanimated permanent did not gain haste")
	}
	if len(g.DelayedTriggers) != 1 {
		t.Fatalf("delayed triggers = %d; want 1 end-step exile", len(g.DelayedTriggers))
	}
	if g.DelayedTriggers[0].Timing != game.DelayedAtBeginningOfNextEndStep {
		t.Fatalf("delayed trigger timing = %v; want next end step", g.DelayedTriggers[0].Timing)
	}

	engine.runEndingPhase(g, [game.NumPlayers]PlayerAgent{})

	if _, stillOnBattlefield := permanentByObjectID(g, permanent.ObjectID); stillOnBattlefield {
		t.Fatal("reanimated permanent remained on the battlefield after the end step")
	}
	if !g.Players[game.Player1].Exile.Contains(cardID) {
		t.Fatal("reanimated permanent was not exiled at the next end step")
	}
}
