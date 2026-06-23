package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// plainSpellCard models a vanilla sorcery whose only relevant property for cost
// modification is that it is a spell cast from hand.
func plainSpellCard(name string) *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:     name,
		Types:    []types.Card{types.Sorcery},
		ManaCost: opt.Val(cost.Mana{cost.O(2)}),
	}}
}

// totalGenericIncrease sums the generic increase the active spell cost modifier
// rule effects impose on playerID casting card from hand.
func totalGenericIncrease(g *game.Game, playerID game.PlayerID, card *game.CardDef) int {
	total := 0
	for _, modifier := range staticCostModifiersForContext(g, playerID, card, zone.Hand, nil) {
		total += modifier.GenericIncrease
	}
	return total
}

// TestResolvingSpellCostModifierAppliesUntilYourNextTurn covers the headline
// behavior of issue #1500: a resolved one-shot effect that taxes opponents'
// spells by {2} until the controller's next turn applies the increase only to
// opponents and disappears when the controller's next turn begins.
func TestResolvingSpellCostModifierAppliesUntilYourNextTurn(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	g.Turn.ActivePlayer = game.Player1
	g.Turn.TurnNumber = 4

	g.RuleEffects = append(g.RuleEffects, game.RuleEffect{
		ID:             g.IDGen.Next(),
		Kind:           game.RuleEffectCostModifier,
		Controller:     game.Player1,
		AffectedPlayer: game.PlayerOpponent,
		Duration:       game.DurationUntilYourNextTurn,
		ExpiresFor:     game.Player1,
		CreatedTurn:    g.Turn.TurnNumber,
		CostModifier: game.CostModifier{
			Kind:            game.CostModifierSpell,
			GenericIncrease: 2,
		},
	})

	card := plainSpellCard("Opponent Spell")
	if got := totalGenericIncrease(g, game.Player2, card); got != 2 {
		t.Fatalf("opponent generic increase while active = %d, want 2", got)
	}
	if got := totalGenericIncrease(g, game.Player1, card); got != 0 {
		t.Fatalf("controller generic increase while active = %d, want 0", got)
	}

	// The modifier expires at the start of the controller's next turn.
	g.Turn.ActivePlayer = game.Player1
	g.Turn.TurnNumber = 6
	expireTurnStartDurations(g)

	if len(g.RuleEffects) != 0 {
		t.Fatalf("rule effects after controller's next turn = %d, want 0", len(g.RuleEffects))
	}
	if got := totalGenericIncrease(g, game.Player2, card); got != 0 {
		t.Fatalf("opponent generic increase after expiry = %d, want 0", got)
	}
}

// TestResolvingSpellCostModifierThisTurnExpiresAtCleanup covers the other
// supported lifetime: a controller-scoped reduction that lasts only the turn it
// resolves, applying to the controller's matching spells and clearing at the
// turn's cleanup.
func TestResolvingSpellCostModifierThisTurnExpiresAtCleanup(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	g.Turn.ActivePlayer = game.Player1
	g.Turn.TurnNumber = 3

	g.RuleEffects = append(g.RuleEffects, game.RuleEffect{
		ID:             g.IDGen.Next(),
		Kind:           game.RuleEffectCostModifier,
		Controller:     game.Player1,
		AffectedPlayer: game.PlayerYou,
		Duration:       game.DurationThisTurn,
		CreatedTurn:    g.Turn.TurnNumber,
		CostModifier: game.CostModifier{
			Kind:             game.CostModifierSpell,
			MatchCardType:    true,
			CardType:         types.Artifact,
			GenericReduction: 1,
		},
	})

	artifact := &game.CardDef{CardFace: game.CardFace{
		Name:     "Artifact Spell",
		Types:    []types.Card{types.Artifact},
		ManaCost: opt.Val(cost.Mana{cost.O(3)}),
	}}
	reduction := 0
	for _, modifier := range staticCostModifiersForContext(g, game.Player1, artifact, zone.Hand, nil) {
		reduction += modifier.GenericReduction
	}
	if reduction != 1 {
		t.Fatalf("controller artifact reduction while active = %d, want 1", reduction)
	}

	// A non-artifact spell is unaffected by the card-type-filtered modifier.
	if got := totalGenericIncrease(g, game.Player1, plainSpellCard("Sorcery Spell")); got != 0 {
		t.Fatalf("non-artifact increase = %d, want 0", got)
	}

	expireRuleEffects(g)
	if len(g.RuleEffects) != 0 {
		t.Fatalf("rule effects after cleanup = %d, want 0", len(g.RuleEffects))
	}
}
