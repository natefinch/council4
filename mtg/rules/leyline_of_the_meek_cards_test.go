package rules

import (
	"testing"

	cards "github.com/natefinch/council4/mtg/cards/l"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// leylineMeekCreatureTokenPermanent puts a creature token onto the battlefield
// under controller with the given base power/toughness.
func leylineMeekCreatureTokenPermanent(g *game.Game, controller game.PlayerID, power, toughness int) *game.Permanent {
	permanent := &game.Permanent{
		ObjectID:   g.IDGen.Next(),
		Owner:      controller,
		Controller: controller,
		Token:      true,
		TokenDef: &game.CardDef{CardFace: game.CardFace{
			Name:      "Soldier Token",
			Types:     []types.Card{types.Creature},
			Power:     opt.Val(game.PT{Value: power}),
			Toughness: opt.Val(game.PT{Value: toughness}),
		}},
	}
	g.Battlefield = append(g.Battlefield, permanent)
	return permanent
}

// TestLeylineOfTheMeekBuffsCreatureTokens proves the real card's "Creature
// tokens get +1/+1." raises a token creature's power and toughness by one.
func TestLeylineOfTheMeekBuffsCreatureTokens(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addCombatPermanent(g, game.Player1, cards.LeylineOfTheMeek())
	token := leylineMeekCreatureTokenPermanent(g, game.Player1, 1, 1)

	if got := effectivePower(g, token); got != 2 {
		t.Fatalf("token power = %d, want 2", got)
	}
	toughness, ok := effectiveToughness(g, token)
	if !ok || toughness != 2 {
		t.Fatalf("token toughness = %d (ok=%v), want 2", toughness, ok)
	}
}

// TestLeylineOfTheMeekIgnoresNontokenCreatures proves the +1/+1 buff applies
// only to token creatures, leaving printed (nontoken) creatures untouched.
func TestLeylineOfTheMeekIgnoresNontokenCreatures(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addCombatPermanent(g, game.Player1, cards.LeylineOfTheMeek())
	nontoken := addCombatCreaturePermanentWithPower(g, game.Player1, 2)

	if got := effectivePower(g, nontoken); got != 2 {
		t.Fatalf("nontoken creature power = %d, want 2 (unbuffed)", got)
	}
	toughness, ok := effectiveToughness(g, nontoken)
	if !ok || toughness != 2 {
		t.Fatalf("nontoken creature toughness = %d (ok=%v), want 2 (unbuffed)", toughness, ok)
	}
}

// TestLeylineOfTheMeekBuffsEveryPlayersTokens proves the anthem is not scoped to
// the controller: "Creature tokens get +1/+1." buffs an opponent's tokens too.
func TestLeylineOfTheMeekBuffsEveryPlayersTokens(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addCombatPermanent(g, game.Player1, cards.LeylineOfTheMeek())
	opponentToken := leylineMeekCreatureTokenPermanent(g, game.Player2, 1, 1)

	if got := effectivePower(g, opponentToken); got != 2 {
		t.Fatalf("opponent token power = %d, want 2", got)
	}
	toughness, ok := effectiveToughness(g, opponentToken)
	if !ok || toughness != 2 {
		t.Fatalf("opponent token toughness = %d (ok=%v), want 2", toughness, ok)
	}
}
