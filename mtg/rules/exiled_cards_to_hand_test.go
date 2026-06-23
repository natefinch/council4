package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

// exiledCardsToHandDef mirrors the two triggered abilities the cardgen lowering
// produces for Wormfang Behemoth: an enters-the-battlefield trigger that moves
// the controller's whole hand to exile and publishes the set under the constant
// exiled-cards-to-hand key, and a separate leaves-the-battlefield trigger that
// returns exactly that linked set to its owners' hands.
func exiledCardsToHandDef() *game.CardDef {
	key := game.LinkedKey("exiled-cards-to-hand")
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
				Content: game.Mode{Sequence: []game.Instruction{{Primitive: game.MoveCard{
					Player:        game.ControllerReference(),
					FromZone:      zone.Hand,
					Destination:   zone.Exile,
					PublishLinked: key,
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
				Content: game.Mode{Sequence: []game.Instruction{{Primitive: game.MoveCard{
					FromLinked:  key,
					FromZone:    zone.Exile,
					Destination: zone.Hand,
				}}}}.Ability(),
			},
		},
	}}
}

// TestExiledCardsToHandRoundTrip exercises the exiled-card back-reference: the
// source exiles the controller's whole hand as a linked set, then returns
// exactly that set to its owner's hand when the source leaves the battlefield.
func TestExiledCardsToHandRoundTrip(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)

	source := addCombatPermanent(g, game.Player1, exiledCardsToHandDef())
	first := addCardToHand(g, game.Player1, greenCreature())
	second := addCardToHand(g, game.Player1, greenCreature())

	obj := linkedSourceObject(source)
	resolveInstruction(engine, g, obj, game.MoveCard{
		Player:        game.ControllerReference(),
		FromZone:      zone.Hand,
		Destination:   zone.Exile,
		PublishLinked: game.LinkedKey("exiled-cards-to-hand"),
	}, nil)

	if g.Players[game.Player1].Hand.Contains(first) || g.Players[game.Player1].Hand.Contains(second) {
		t.Fatal("hand cards were not exiled")
	}
	if !g.Players[game.Player1].Exile.Contains(first) || !g.Players[game.Player1].Exile.Contains(second) {
		t.Fatal("hand cards did not reach exile")
	}

	movePermanentToZone(g, source, zone.Graveyard)
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("leaves-the-battlefield return trigger did not fire")
	}
	engine.resolveTopOfStack(g, &TurnLog{})

	if !g.Players[game.Player1].Hand.Contains(first) || !g.Players[game.Player1].Hand.Contains(second) {
		t.Fatal("exiled cards did not return to their owner's hand")
	}
	if g.Players[game.Player1].Exile.Contains(first) || g.Players[game.Player1].Exile.Contains(second) {
		t.Fatal("exiled cards remained in exile after the source left the battlefield")
	}
}

// TestExiledCardsToHandReturnsToOwnersHand confirms a card whose owner differs
// from the exiling source's controller returns to its own owner's hand, matching
// the "their owner's hand" wording.
func TestExiledCardsToHandReturnsToOwnersHand(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)

	source := addCombatPermanent(g, game.Player1, exiledCardsToHandDef())
	owned := addCardToHand(g, game.Player1, greenCreature())

	obj := linkedSourceObject(source)
	resolveInstruction(engine, g, obj, game.MoveCard{
		Player:        game.ControllerReference(),
		FromZone:      zone.Hand,
		Destination:   zone.Exile,
		PublishLinked: game.LinkedKey("exiled-cards-to-hand"),
	}, nil)

	// Reassign the exiled card's owner to Player2 to prove the return targets the
	// card's owner, not the source's controller.
	g.CardInstances[owned].Owner = game.Player2
	g.Players[game.Player1].Exile.Remove(owned)
	g.Players[game.Player2].Exile.Add(owned)

	resolveInstruction(engine, g, obj, game.MoveCard{
		FromLinked:  game.LinkedKey("exiled-cards-to-hand"),
		FromZone:    zone.Exile,
		Destination: zone.Hand,
	}, nil)

	if !g.Players[game.Player2].Hand.Contains(owned) {
		t.Fatal("exiled card did not return to its owner's hand")
	}
}
