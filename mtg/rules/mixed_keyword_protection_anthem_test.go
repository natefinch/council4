package rules

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// akromasMemorial mirrors the static ability the executable backend generates
// for "Creatures you control have flying, first strike, vigilance, trample,
// haste, and protection from black and from red." — a single LayerAbility
// continuous effect whose grant list mixes simple keywords (carried in
// AddKeywords) with an ability-backed keyword (protection from colors, carried
// in AddAbilities).
func akromasMemorial(g *game.Game, controller game.PlayerID) *game.Permanent {
	protection := game.ProtectionFromColorsStaticAbility(color.Black, color.Red)
	return addCombatPermanent(g, controller, &game.CardDef{CardFace: game.CardFace{
		Name:  "Akroma's Memorial",
		Types: []types.Card{types.Artifact},
		StaticAbilities: []game.StaticAbility{{
			ContinuousEffects: []game.ContinuousEffect{{
				Layer: game.LayerAbility,
				Group: game.ObjectControlledGroup(
					game.SourcePermanentReference(),
					game.Selection{RequiredTypes: []types.Card{types.Creature}},
				),
				AddKeywords:  []game.Keyword{game.Flying, game.Vigilance},
				AddAbilities: []game.Ability{&protection},
			}},
		}},
	}})
}

func permanentHasGrantedProtectionFromColor(g *game.Game, permanent *game.Permanent, c color.Color) bool {
	for _, ability := range permanentEffectiveAbilities(g, permanent) {
		body, ok := ability.(*game.StaticAbility)
		if !ok {
			continue
		}
		prot, ok := game.StaticBodyProtectionKeyword(body)
		if !ok {
			continue
		}
		if slices.Contains(prot.FromColors, c) {
			return true
		}
	}
	return false
}

// TestMixedKeywordProtectionAnthemGrantsBoth confirms a single continuous grant
// that mixes simple keywords with protection applies BOTH to the controlled
// creatures: a creature you control gains the simple keywords and protection
// from both colors, while an opponent's creature is unaffected.
func TestMixedKeywordProtectionAnthemGrantsBoth(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	akromasMemorial(g, game.Player1)

	mine := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:      "Friendly Bear",
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(game.PT{Value: 2}),
		Toughness: opt.Val(game.PT{Value: 2}),
	}})
	theirs := addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
		Name:      "Rival Bear",
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(game.PT{Value: 2}),
		Toughness: opt.Val(game.PT{Value: 2}),
	}})

	view := findPermanentView(t, observe(g, game.Player1), mine.ObjectID)
	if !view.HasKeyword(game.Flying) || !view.HasKeyword(game.Vigilance) {
		t.Fatalf("controlled creature keywords = flying:%v vigilance:%v, want both true",
			view.HasKeyword(game.Flying), view.HasKeyword(game.Vigilance))
	}
	if !permanentHasGrantedProtectionFromColor(g, mine, color.Black) ||
		!permanentHasGrantedProtectionFromColor(g, mine, color.Red) {
		t.Fatal("controlled creature lacks granted protection from black and red")
	}

	opponentView := findPermanentView(t, observe(g, game.Player1), theirs.ObjectID)
	if opponentView.HasKeyword(game.Flying) {
		t.Fatal("opponent creature should not gain flying")
	}
	if permanentHasGrantedProtectionFromColor(g, theirs, color.Black) {
		t.Fatal("opponent creature should not gain protection")
	}
}
