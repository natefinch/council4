package rules

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

func TestCompiledIndependentSubtypeUnionScope(t *testing.T) {
	def := compiledPermanentUnion(t, "Tap target tapped Merfolk creature or tapped Kithkin creature.")
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	for _, tc := range []struct {
		kind   types.Card
		sub    types.Sub
		tapped bool
		want   bool
	}{
		{types.Creature, types.Merfolk, true, true},
		{types.Creature, types.Kithkin, true, true},
		{types.Creature, types.Kithkin, false, false},
		{types.Artifact, types.Merfolk, true, false},
		{types.Artifact, types.Kithkin, true, false},
	} {
		target := addCombatPermanent(g, game.Player3, &game.CardDef{CardFace: game.CardFace{
			Name: "Scope Candidate", Types: []types.Card{tc.kind}, Subtypes: []types.Sub{tc.sub},
		}})
		target.Tapped = tc.tapped
		choices := targetChoicesForSpell(g, game.Player1, def, nil, game.CastBranch{})
		offered := slices.ContainsFunc(choices.choices, func(choice []game.Target) bool {
			return slices.Contains(choice, game.PermanentTarget(target.ObjectID))
		})
		if offered != tc.want {
			t.Errorf("kind=%v subtype=%v tapped=%v offered=%v want=%v", tc.kind, tc.sub, tc.tapped, offered, tc.want)
		}
	}
}
func TestCompiledIndependentUnionManaValueScope(t *testing.T) {
	def := compiledPermanentUnion(t, "Tap target creature with mana value 3 or less or Vehicle with mana value 3 or less.")
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	for _, tc := range []struct {
		kind types.Card
		sub  types.Sub
		mv   int
	}{
		{types.Creature, "", 3},
		{types.Creature, "", 4},
		{types.Artifact, types.Vehicle, 3},
		{types.Artifact, types.Vehicle, 4},
	} {
		card := &game.CardDef{CardFace: game.CardFace{
			Name: "Bound Candidate", Types: []types.Card{tc.kind}, ManaCost: opt.Val(cost.Mana{cost.O(tc.mv)}),
		}}
		if tc.sub != "" {
			card.Subtypes = []types.Sub{tc.sub}
		}
		target := addCombatPermanent(g, game.Player3, card)
		choices := targetChoicesForSpell(g, game.Player1, def, nil, game.CastBranch{})
		offered := slices.ContainsFunc(choices.choices, func(choice []game.Target) bool {
			return slices.Contains(choice, game.PermanentTarget(target.ObjectID))
		})
		if offered != (tc.mv <= 3) {
			t.Errorf("kind=%v subtype=%v MV=%d offered=%v", tc.kind, tc.sub, tc.mv, offered)
		}
	}
}
