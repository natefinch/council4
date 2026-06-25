package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

// wormfangCardDef mirrors the two triggered abilities the cardgen lowering
// produces for Wormfang Behemoth: an enters trigger that exiles the controller's
// entire hand under the constant exile-hand-return key, and a separate
// leaves-the-battlefield trigger that returns that exact linked set to its
// owners' hands.
func wormfangCardDef() *game.CardDef {
	key := game.LinkedKey("exile-hand-return")
	return &game.CardDef{CardFace: game.CardFace{
		Name:  "Wormfang Behemoth",
		Types: []types.Card{types.Creature},
		TriggeredAbilities: []game.TriggeredAbility{
			{
				Trigger: game.TriggerCondition{
					Type: game.TriggerWhen,
					Pattern: game.TriggerPattern{
						Event:  game.EventPermanentEnteredBattlefield,
						Source: game.TriggerSourceSelf,
					},
				},
				Content: game.Mode{Sequence: []game.Instruction{{Primitive: game.ExileEntireHand{
					Player:    game.ControllerReference(),
					LinkedKey: key,
				}}}}.Ability(),
			},
			{
				Trigger: game.TriggerCondition{
					Type: game.TriggerWhen,
					Pattern: game.TriggerPattern{
						Event:         game.EventZoneChanged,
						Source:        game.TriggerSourceSelf,
						MatchFromZone: true,
						FromZone:      zone.Battlefield,
					},
				},
				Content: game.Mode{Sequence: []game.Instruction{{Primitive: game.ReturnExiledCardsToHand{
					LinkedKey: key,
				}}}}.Ability(),
			},
		},
	}}
}

func TestExileEntireHandThenReturnOnSourceLeaving(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addCombatPermanent(g, game.Player1, wormfangCardDef())

	handCard1 := addCardToHand(g, game.Player1, evidenceCard("Plains", 0))
	handCard2 := addCardToHand(g, game.Player1, evidenceCard("Island", 0))

	obj := linkedSourceObject(source)
	resolveInstruction(engine, g, obj, game.ExileEntireHand{
		Player:    game.ControllerReference(),
		LinkedKey: game.LinkedKey("exile-hand-return"),
	}, nil)

	if g.Players[game.Player1].Hand.Size() != 0 {
		t.Fatalf("hand size = %d, want 0 after entire-hand exile", g.Players[game.Player1].Hand.Size())
	}
	if !g.Players[game.Player1].Exile.Contains(handCard1) || !g.Players[game.Player1].Exile.Contains(handCard2) {
		t.Fatal("exiled hand cards did not reach the owner's exile zone")
	}

	movePermanentToZone(g, source, zone.Graveyard)
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("leaves-the-battlefield return trigger did not fire")
	}
	engine.resolveTopOfStack(g, &TurnLog{})

	if !g.Players[game.Player1].Hand.Contains(handCard1) || !g.Players[game.Player1].Hand.Contains(handCard2) {
		t.Fatalf("hand = %+v, want both exiled cards returned", g.Players[game.Player1].Hand.All())
	}
	if g.Players[game.Player1].Exile.Contains(handCard1) || g.Players[game.Player1].Exile.Contains(handCard2) {
		t.Fatal("cards remained in exile after the source left the battlefield")
	}
}
