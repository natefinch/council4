package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// castleGarenbrigCreatureSpell returns a castable {1} creature spell, standing in
// for a creature the six green mana from Castle Garenbrig's "{2}{G}{G}, {T}: Add
// six {G}." ability may pay for.
func castleGarenbrigCreatureSpell() *game.CardDef {
	def := creatureSpellDef("Garenbrig Bear", types.Bear)
	def.ManaCost = opt.Val(cost.Mana{cost.O(1)})
	def.Power = opt.Val(game.PT{Value: 2})
	def.Toughness = opt.Val(game.PT{Value: 2})
	return def
}

// creaturePermanentDef returns a bare creature permanent usable as the source of
// an activated ability whose cost the restricted mana may pay.
func creaturePermanentDef() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:      "Creature Source",
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Bear},
		Power:     opt.Val(game.PT{Value: 1}),
		Toughness: opt.Val(game.PT{Value: 1}),
	}}
}

// gameWithCreatureRestrictedMana seeds Player1 with one {G} tagged with the
// creature cast-or-activate restriction, modeling a unit of the six green mana
// Castle Garenbrig produces.
func gameWithCreatureRestrictedMana() (*game.Game, *game.Player) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	player := g.Players[game.Player1]
	player.ManaPool.Add(mana.G, 1)
	player.ManaRiders = append(player.ManaRiders, game.ManaRiderInstance{
		Unit:       mana.Unit{Color: mana.G},
		Controller: game.Player1,
		Rider: game.ManaSpendRider{
			Condition:   game.ManaSpendCastOrActivateCreature,
			Restriction: game.ManaSpendRestrictedToCondition,
		},
	})
	return g, player
}

// TestCastleGarenbrigCreatureManaRestriction covers Castle Garenbrig: the tagged
// mana may pay to cast a creature spell or to activate an ability of a creature,
// but not to cast a noncreature spell or activate a noncreature's ability.
func TestCastleGarenbrigCreatureManaRestriction(t *testing.T) {
	t.Parallel()

	t.Run("accepts creature spell", func(t *testing.T) {
		t.Parallel()
		g, player := gameWithCreatureRestrictedMana()
		if !castWithArtifactMana(t, g, castleGarenbrigCreatureSpell()) {
			t.Fatal("cast creature spell with creature-restricted mana failed")
		}
		if player.ManaPool.Amount(mana.G) != 0 {
			t.Fatalf("creature-restricted mana not spent: pool=%d", player.ManaPool.Amount(mana.G))
		}
	})

	t.Run("rejects noncreature spell", func(t *testing.T) {
		t.Parallel()
		g, player := gameWithCreatureRestrictedMana()
		if castWithArtifactMana(t, g, monocoloredNoncreatureSpellDef()) {
			t.Fatal("cast noncreature spell with creature-restricted mana succeeded")
		}
		if player.ManaPool.Amount(mana.G) != 1 || len(player.ManaRiders) != 1 {
			t.Fatalf("restricted mana changed after rejected cast: pool=%d riders=%d", player.ManaPool.Amount(mana.G), len(player.ManaRiders))
		}
	})

	t.Run("accepts creature ability", func(t *testing.T) {
		t.Parallel()
		g, player := gameWithCreatureRestrictedMana()
		source := addPermanentForSBA(g, game.Player1, creaturePermanentDef())
		if !payArtifactAbility(g, source) {
			t.Fatal("creature-restricted mana failed to pay a creature ability")
		}
		if player.ManaPool.Amount(mana.G) != 0 || len(player.ManaRiders) != 0 {
			t.Fatalf("rider not consumed by creature ability: pool=%d riders=%d", player.ManaPool.Amount(mana.G), len(player.ManaRiders))
		}
	})

	t.Run("rejects noncreature ability", func(t *testing.T) {
		t.Parallel()
		g, player := gameWithCreatureRestrictedMana()
		source := addPermanentForSBA(g, game.Player1, artifactPermanentDef())
		if payArtifactAbility(g, source) {
			t.Fatal("creature-restricted mana paid a noncreature ability")
		}
		if player.ManaPool.Amount(mana.G) != 1 || len(player.ManaRiders) != 1 {
			t.Fatalf("restricted mana changed after rejected payment: pool=%d riders=%d", player.ManaPool.Amount(mana.G), len(player.ManaRiders))
		}
	})
}
