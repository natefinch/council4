package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// TestCreateTokenDynamicSizeOverride verifies that a create-token instruction
// carrying a dynamic power and toughness override ("create an X/X ... token,
// where X is the amount of life you gained this turn.", Tivash, Gloom Summoner)
// fixes the created token's printed power and toughness to the resolved amount
// while leaving the source definition's unset printed P/T untouched.
func TestCreateTokenDynamicSizeOverride(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	obj := &game.StackObject{
		ID:         g.IDGen.Next(),
		Kind:       game.StackSpell,
		Controller: game.Player1,
	}
	r := &effectResolver{engine: NewEngine(nil), game: g, obj: obj, log: &TurnLog{}}

	def := &game.CardDef{CardFace: game.CardFace{
		Name:     "Demon",
		Types:    []types.Card{types.Creature},
		Subtypes: []types.Sub{types.Demon},
	}}
	resolved := handleCreateToken(r, game.CreateToken{
		Amount:    game.Fixed(1),
		Source:    game.TokenDef(def),
		Power:     opt.Val(game.Fixed(4)),
		Toughness: opt.Val(game.Fixed(4)),
	})
	if !resolved.succeeded {
		t.Fatal("handleCreateToken did not succeed")
	}

	var token *game.Permanent
	for _, permanent := range g.Battlefield {
		if permanent.Token {
			token = permanent
			break
		}
	}
	if token == nil || token.TokenDef == nil {
		t.Fatal("no token created")
	}
	if !token.TokenDef.Power.Exists || token.TokenDef.Power.Val.Value != 4 ||
		!token.TokenDef.Toughness.Exists || token.TokenDef.Toughness.Val.Value != 4 {
		t.Fatalf("token P/T = %+v/%+v, want 4/4 from override", token.TokenDef.Power, token.TokenDef.Toughness)
	}
	if def.Power.Exists || def.Toughness.Exists {
		t.Fatalf("source definition P/T mutated: %+v/%+v, want unset", def.Power, def.Toughness)
	}
}
