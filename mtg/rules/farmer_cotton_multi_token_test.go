package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
)

// countTokensNamed counts battlefield token permanents controlled by controller
// whose token definition carries name.
func countTokensNamed(g *game.Game, name string, controller game.PlayerID) int {
	n := 0
	for _, permanent := range g.Battlefield {
		if permanent.Token && permanent.Controller == controller &&
			permanent.TokenDef != nil && permanent.TokenDef.Name == name {
			n++
		}
	}
	return n
}

func farmerCottonEntersContent() game.AbilityContent {
	halfling := &game.CardDef{CardFace: game.CardFace{
		Name: "Halfling", Types: []types.Card{types.Creature}, Subtypes: []types.Sub{types.Halfling},
	}}
	food := &game.CardDef{CardFace: game.CardFace{
		Name: "Food", Types: []types.Card{types.Artifact}, Subtypes: []types.Sub{types.Food},
	}}
	return game.Mode{Sequence: []game.Instruction{
		{Primitive: game.CreateToken{Amount: game.Dynamic(game.DynamicAmount{Kind: game.DynamicAmountX}), Source: game.TokenDef(halfling)}},
		{Primitive: game.CreateToken{Amount: game.Dynamic(game.DynamicAmount{Kind: game.DynamicAmountX}), Source: game.TokenDef(food)}},
	}}.Ability()
}

// resolveFarmerCottonEnters resolves Farmer Cotton's ETB trigger with the given
// captured X. The trigger's variable X reads from the resolving stack object's
// XValue, the value the permanent captured when it was cast.
func resolveFarmerCottonEnters(t *testing.T, g *game.Game, xValue int) {
	t.Helper()
	engine := NewEngine(nil)
	obj := &game.StackObject{
		ID: g.IDGen.Next(), Kind: game.StackTriggeredAbility, Controller: game.Player1, XValue: xValue,
	}
	engine.resolveAbilityContentWithChoices(g, obj, farmerCottonEntersContent(), [game.NumPlayers]PlayerAgent{}, &TurnLog{})
}

// TestFarmerCottonCreatesBothTokenTypesAtX proves the shared-X multi-token create
// makes exactly X Halfling creature tokens and X Food tokens from one captured X.
func TestFarmerCottonCreatesBothTokenTypesAtX(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	resolveFarmerCottonEnters(t, g, 3)
	if got := countTokensNamed(g, "Halfling", game.Player1); got != 3 {
		t.Errorf("Halfling tokens = %d, want 3", got)
	}
	if got := countTokensNamed(g, "Food", game.Player1); got != 3 {
		t.Errorf("Food tokens = %d, want 3", got)
	}
}

// TestFarmerCottonXZeroCreatesNoTokens proves that with X=0 neither token type is
// created — the dynamic count is honored rather than floored up to one each.
func TestFarmerCottonXZeroCreatesNoTokens(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	resolveFarmerCottonEnters(t, g, 0)
	if got := countTokensNamed(g, "Halfling", game.Player1); got != 0 {
		t.Errorf("Halfling tokens at X=0 = %d, want 0", got)
	}
	if got := countTokensNamed(g, "Food", game.Player1); got != 0 {
		t.Errorf("Food tokens at X=0 = %d, want 0", got)
	}
}

// TestFarmerCottonReplacementDoublesBothTokenTypes proves a token-creation
// doubling replacement (Anointed Procession) applies to each multi-token spec
// independently, so both the Halflings and the Food are doubled.
func TestFarmerCottonReplacementDoublesBothTokenTypes(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addReplacementPermanent(t, g, game.Player1, tokenDoublingReplacementCardDef())
	resolveFarmerCottonEnters(t, g, 2)
	if got := countTokensNamed(g, "Halfling", game.Player1); got != 4 {
		t.Errorf("Halfling tokens = %d, want 4 (X=2 doubled)", got)
	}
	if got := countTokensNamed(g, "Food", game.Player1); got != 4 {
		t.Errorf("Food tokens = %d, want 4 (X=2 doubled)", got)
	}
}
