package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// TestGrantCastPermissionTargetGraveyardCardUntilEndOfTurn verifies the targeted
// graveyard cast-permission primitive "you may cast target <card> from your
// graveyard this turn" (Norika Yamazaki, the Poet): resolving the grant lets the
// controller cast the chosen graveyard card's front face until end of turn, and
// the permission clears at that turn's cleanup.
func TestGrantCastPermissionTargetGraveyardCardUntilEndOfTurn(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	g.Turn.TurnNumber = 3
	g.Turn.ActivePlayer = game.Player1

	targetID := addCardToGraveyard(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Graveyard Aura",
		Types: []types.Card{types.Enchantment},
	}})
	sourceID := addEffectSpellToStack(g, game.Player1, game.GrantCastPermission{
		Card:     game.CardReference{Kind: game.CardReferenceTarget},
		FromZone: zone.Graveyard,
		Face:     game.FaceFront,
		Duration: game.DurationUntilEndOfTurn,
	}, []game.Target{currentCardTarget(t, g, targetID)})
	card, ok := g.GetCardInstance(sourceID)
	if !ok {
		t.Fatal("source card instance not found")
	}
	card.Def.SpellAbility.Val.Modes[0].Targets = []game.TargetSpec{{
		MinTargets: 1,
		MaxTargets: 1,
		Allow:      game.TargetAllowCard,
		TargetZone: zone.Graveyard,
		Selection:  opt.Val(game.Selection{RequiredTypes: []types.Card{types.Enchantment}, Controller: game.ControllerYou}),
	}}

	engine.resolveTopOfStack(g, &TurnLog{})

	var granted *game.RuleEffect
	for i := range g.RuleEffects {
		if g.RuleEffects[i].Kind == game.RuleEffectCastFromZone {
			granted = &g.RuleEffects[i]
		}
	}
	if granted == nil {
		t.Fatal("no cast-from-graveyard permission created")
	}
	if granted.AffectedCardID != targetID {
		t.Fatalf("granted.AffectedCardID = %v, want %v (the targeted card)", granted.AffectedCardID, targetID)
	}
	if granted.CastFromZone != zone.Graveyard {
		t.Fatalf("granted.CastFromZone = %v, want graveyard", granted.CastFromZone)
	}
	if !granted.CastFace.Exists || granted.CastFace.Val != game.FaceFront {
		t.Fatalf("granted.CastFace = %+v, want front face", granted.CastFace)
	}
	if granted.Duration != game.DurationUntilEndOfTurn {
		t.Fatalf("granted.Duration = %v, want until end of turn", granted.Duration)
	}

	expireRuleEffects(g)
	for i := range g.RuleEffects {
		if g.RuleEffects[i].Kind == game.RuleEffectCastFromZone {
			t.Fatal("cast-from-graveyard permission survived end-of-turn cleanup")
		}
	}
}
