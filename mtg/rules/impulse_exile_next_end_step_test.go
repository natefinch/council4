package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

// TestImpulseExileUntilYourNextEndStepGrantsPlay proves that an "until your next
// end step" impulse exile (Inti, Seneschal of the Sun; Yasmin Khan; Wiccan,
// Young Avenger) exiles the top card and records a play-from-exile rule effect
// that expires for the controller.
func TestImpulseExileUntilYourNextEndStepGrantsPlay(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	topID := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Top Card",
		Types: []types.Card{types.Creature},
	}})

	resolveInstruction(engine, g, &game.StackObject{
		Controller:   game.Player1,
		SourceCardID: g.IDGen.Next(),
		SourceID:     g.IDGen.Next(),
	}, game.ImpulseExile{
		Player:   game.ControllerReference(),
		Amount:   game.Fixed(1),
		Duration: game.DurationUntilYourNextEndStep,
	}, &TurnLog{})

	if !g.Players[game.Player1].Exile.Contains(topID) {
		t.Fatal("top card was not exiled")
	}
	var effect *game.RuleEffect
	for i := range g.RuleEffects {
		if g.RuleEffects[i].Kind == game.RuleEffectPlayFromZone &&
			g.RuleEffects[i].AffectedCardID == topID {
			effect = &g.RuleEffects[i]
		}
	}
	if effect == nil {
		t.Fatal("no play-from-zone rule effect was created")
	}
	if effect.Duration != game.DurationUntilYourNextEndStep {
		t.Fatalf("duration = %v, want DurationUntilYourNextEndStep", effect.Duration)
	}
	if effect.ExpiresFor != game.Player1 || effect.CastFromZone != zone.Exile {
		t.Fatalf("rule effect = %+v", *effect)
	}
}

// TestImpulseExileUntilYourNextEndStepExpiry proves that an "until your next end
// step" play permission survives a cleanup on the opponent's turn but is removed
// at the cleanup of the controller's own turn (its next end step).
func TestImpulseExileUntilYourNextEndStepExpiry(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	g.RuleEffects = append(g.RuleEffects, game.RuleEffect{
		ID:           g.IDGen.Next(),
		Kind:         game.RuleEffectPlayFromZone,
		Controller:   game.Player1,
		Duration:     game.DurationUntilYourNextEndStep,
		CreatedTurn:  1,
		CastFromZone: zone.Exile,
		ExpiresFor:   game.Player1,
	})

	// Cleanup on the opponent's turn must not expire a permission that expires
	// for Player1.
	g.Turn.ActivePlayer = game.Player2
	g.Turn.TurnNumber = 2
	expireRuleEffects(g)
	if len(g.RuleEffects) != 1 {
		t.Fatalf("permission expired on opponent's cleanup: %+v", g.RuleEffects)
	}

	// Cleanup on the controller's own turn (its next end step) expires it.
	g.Turn.ActivePlayer = game.Player1
	g.Turn.TurnNumber = 3
	expireRuleEffects(g)
	if len(g.RuleEffects) != 0 {
		t.Fatalf("permission survived its controller's end step: %+v", g.RuleEffects)
	}
}
