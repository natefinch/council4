package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
)

// nonbasicLandDef builds a minimal nonbasic land permanent for the Anathemancer
// count test: it carries no supertype, so the count's ExcludedSupertype filter
// includes it while excluding basic lands built by basicLandDef.
func nonbasicLandDef(name string) *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:  name,
		Types: []types.Card{types.Land},
	}}
}

// TestAnathemancerDamageCountsTargetNonbasicLands resolves Anathemancer's ETB
// "deals damage to target player equal to the number of nonbasic lands that
// player controls" as a Damage whose dynamic count is a player-controlled land
// group anchored to the single player target. The targeted player must take
// damage equal to their own nonbasic lands only: their basic land is excluded
// by the count's ExcludedSupertype filter and another player's nonbasic lands
// do not contribute.
func TestAnathemancerDamageCountsTargetNonbasicLands(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)

	addCombatPermanent(g, game.Player2, nonbasicLandDef("Target Nonbasic 1"))
	addCombatPermanent(g, game.Player2, nonbasicLandDef("Target Nonbasic 2"))
	addCombatPermanent(g, game.Player2, nonbasicLandDef("Target Nonbasic 3"))
	addCombatPermanent(g, game.Player2, basicLandDef(types.Swamp))
	addCombatPermanent(g, game.Player1, nonbasicLandDef("Caster Nonbasic"))

	addEffectSpellToStack(g, game.Player1, game.Damage{
		Amount: game.Dynamic(game.DynamicAmount{
			Kind:       game.DynamicAmountCountSelector,
			Multiplier: 1,
			Group: game.PlayerControlledGroup(game.TargetPlayerReference(0), game.Selection{
				RequiredTypes:     []types.Card{types.Land},
				ExcludedSupertype: types.Basic,
			}),
		}),
		Recipient: game.AnyTargetDamageRecipient(0),
	}, []game.Target{game.PlayerTarget(game.Player2)})
	engine.resolveTopOfStack(g, &TurnLog{})

	if got := g.Players[game.Player2].Life; got != 37 {
		t.Fatalf("targeted player life = %d, want 37 (40 - 3 nonbasic lands)", got)
	}
	if got := g.Players[game.Player1].Life; got != 40 {
		t.Fatalf("caster life = %d, want 40 (untouched)", got)
	}
}
