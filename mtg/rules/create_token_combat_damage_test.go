package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
)

// TestCreateTokenCombatDamageTriggerCreatesTokensEqualToDamage verifies the
// "Whenever a creature you control deals combat damage to a player, create that
// many Treasure tokens." family (Old Gnawbone): the dynamic token count reads
// the triggering combat damage amount.
func TestCreateTokenCombatDamageTriggerCreatesTokensEqualToDamage(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	treasure := &game.CardDef{CardFace: game.CardFace{
		Name:     "Treasure",
		Types:    []types.Card{types.Artifact},
		Subtypes: []types.Sub{types.Treasure},
	}}
	addTriggeredPermanent(g, game.Player1, &game.TriggerPattern{
		Event:               game.EventDamageDealt,
		Controller:          game.TriggerControllerYou,
		Subject:             game.TriggerSubjectDamageSource,
		DamageRecipient:     game.DamageRecipientPlayer,
		RequireCombatDamage: true,
	}, []game.Instruction{{Primitive: game.CreateToken{
		Amount: game.Dynamic(game.DynamicAmount{Kind: game.DynamicAmountEventDamage}),
		Source: game.TokenDef(treasure),
	}}}, nil)

	dealPlayerDamage(g, 0, 0, game.Player1, game.Player2, 4, true)
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("combat-damage trigger was not put on stack")
	}
	engine.resolveTopOfStack(g, &TurnLog{})

	if got := countTokenPermanentsNamed(g, "Treasure"); got != 4 {
		t.Fatalf("Treasure tokens = %d, want 4 (equal to combat damage dealt)", got)
	}
}
