package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/zone"
)

// TestTargetPlayerHandSizeDamageCountsTargetsHand proves a "deals damage to
// target player equal to the number of cards in that player's hand" effect (Gaze
// of Adamaro, Sudden Impact, Storm Seeker, Toil // Trouble, Vicious Shadows)
// counts the TARGET player's hand. The damage amount's player reference is bound
// to the target the damage is dealt to, so the caster's hand size is irrelevant.
func TestTargetPlayerHandSizeDamageCountsTargetsHand(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)

	// Target (Player2) holds four cards; the caster (Player1) holds one, so a
	// caster-scoped miscount would deal one instead of four.
	for range 4 {
		addCardToHand(g, game.Player2, &game.CardDef{CardFace: game.CardFace{Name: "Target Card"}})
	}
	addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Caster Card"}})
	beforeP2 := g.Players[game.Player2].Life

	targetPlayer := game.TargetPlayerReference(0)
	addInstructionSpellToStackForController(g, game.Player1, []game.Instruction{
		{
			Primitive: game.Damage{
				Amount: game.Dynamic(game.DynamicAmount{
					Kind:       game.DynamicAmountCountCardsInZone,
					Multiplier: 1,
					Player:     &targetPlayer,
					CardZone:   zone.Hand,
					Selection:  &game.Selection{},
				}),
				Recipient: game.AnyTargetDamageRecipient(0),
			},
		},
	}, []game.Target{game.PlayerTarget(game.Player2)})

	engine.resolveTopOfStack(g, &TurnLog{})

	if got := beforeP2 - g.Players[game.Player2].Life; got != 4 {
		t.Fatalf("target player life lost = %d, want 4 (its own hand size)", got)
	}
}
