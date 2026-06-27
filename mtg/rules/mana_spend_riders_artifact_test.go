package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/rules/payment"
	"github.com/natefinch/council4/opt"
)

// artifactSpellDef returns a one-mana artifact spell usable as a cast target.
func artifactSpellDef() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:     "Bauble",
		Types:    []types.Card{types.Artifact},
		ManaCost: opt.Val(cost.Mana{cost.O(1)}),
	}}
}

// nonartifactSpellDef returns a one-mana creature spell usable as a cast target.
func nonartifactSpellDef() *game.CardDef {
	def := creatureSpellDef("Goblin", types.Goblin)
	def.ManaCost = opt.Val(cost.Mana{cost.O(1)})
	def.Power = opt.Val(game.PT{Value: 1})
	def.Toughness = opt.Val(game.PT{Value: 1})
	return def
}

// artifactPermanentDef returns a bare artifact permanent definition.
func artifactPermanentDef() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:  "Artifact Source",
		Types: []types.Card{types.Artifact},
	}}
}

// gameWithArtifactRestrictedMana seeds Player1 with one {C} tagged with the
// given artifact restriction condition.
func gameWithArtifactRestrictedMana(condition game.ManaSpendConditionKind) (*game.Game, *game.Player) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	player := g.Players[game.Player1]
	player.ManaPool.Add(mana.C, 1)
	player.ManaRiders = append(player.ManaRiders, game.ManaRiderInstance{
		Unit:       mana.Unit{Color: mana.C},
		Controller: game.Player1,
		Rider: game.ManaSpendRider{
			Condition:   condition,
			Restriction: game.ManaSpendRestrictedToCondition,
		},
	})
	return g, player
}

// castWithArtifactMana attempts to cast def using the restricted mana, returning
// whether the cast succeeded.
func castWithArtifactMana(t *testing.T, g *game.Game, def *game.CardDef) bool {
	t.Helper()
	engine := NewEngine(nil)
	spellID := addCardToHand(g, game.Player1, def)
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone
	return engine.applyAction(g, game.Player1, action.CastSpell(spellID, nil, 0, nil))
}

// payArtifactAbility attempts to pay a {1} ability cost of source with the
// restricted mana, returning whether the payment succeeded.
func payArtifactAbility(g *game.Game, source *game.Permanent) bool {
	manaCost := cost.Mana{cost.O(1)}
	_, ok := paymentOrch.payAbilityCosts(g, payment.AbilityRequest{
		PlayerID: game.Player1,
		Source:   source,
		ManaCost: opt.Val(manaCost),
	})
	return ok
}

// TestArtifactSpellOnlyManaCastRestriction covers Castle Doom / Mishra's
// Workshop: the tagged mana may pay only to cast an artifact spell.
func TestArtifactSpellOnlyManaCastRestriction(t *testing.T) {
	t.Parallel()

	t.Run("rejects nonartifact spell", func(t *testing.T) {
		t.Parallel()
		g, player := gameWithArtifactRestrictedMana(game.ManaSpendCastArtifactSpellOnly)
		if castWithArtifactMana(t, g, nonartifactSpellDef()) {
			t.Fatal("cast nonartifact spell with artifact-spell-only mana succeeded")
		}
		if player.ManaPool.Amount(mana.C) != 1 || len(player.ManaRiders) != 1 {
			t.Fatalf("restricted mana changed after rejected cast: pool=%d riders=%d", player.ManaPool.Amount(mana.C), len(player.ManaRiders))
		}
	})

	t.Run("accepts artifact spell", func(t *testing.T) {
		t.Parallel()
		g, player := gameWithArtifactRestrictedMana(game.ManaSpendCastArtifactSpellOnly)
		if !castWithArtifactMana(t, g, artifactSpellDef()) {
			t.Fatal("cast artifact spell with artifact-spell-only mana failed")
		}
		if player.ManaPool.Amount(mana.C) != 0 {
			t.Fatalf("artifact-spell-only mana not spent: pool=%d", player.ManaPool.Amount(mana.C))
		}
	})

	t.Run("rejects artifact ability payment", func(t *testing.T) {
		t.Parallel()
		g, player := gameWithArtifactRestrictedMana(game.ManaSpendCastArtifactSpellOnly)
		source := addPermanentForSBA(g, game.Player1, artifactPermanentDef())
		if payArtifactAbility(g, source) {
			t.Fatal("artifact-spell-only mana paid an ability cost")
		}
		if player.ManaPool.Amount(mana.C) != 1 || len(player.ManaRiders) != 1 {
			t.Fatalf("restricted mana changed after rejected payment: pool=%d riders=%d", player.ManaPool.Amount(mana.C), len(player.ManaRiders))
		}
	})
}

// TestCastOrActivateArtifactManaRestriction covers Power Depot / Cargo Ship: the
// tagged mana may cast an artifact spell or activate an artifact's ability.
func TestCastOrActivateArtifactManaRestriction(t *testing.T) {
	t.Parallel()

	t.Run("rejects nonartifact spell", func(t *testing.T) {
		t.Parallel()
		g, player := gameWithArtifactRestrictedMana(game.ManaSpendCastOrActivateArtifact)
		if castWithArtifactMana(t, g, nonartifactSpellDef()) {
			t.Fatal("cast nonartifact spell with cast-or-activate-artifact mana succeeded")
		}
		if player.ManaPool.Amount(mana.C) != 1 || len(player.ManaRiders) != 1 {
			t.Fatalf("restricted mana changed after rejected cast: pool=%d riders=%d", player.ManaPool.Amount(mana.C), len(player.ManaRiders))
		}
	})

	t.Run("accepts artifact spell", func(t *testing.T) {
		t.Parallel()
		g, player := gameWithArtifactRestrictedMana(game.ManaSpendCastOrActivateArtifact)
		if !castWithArtifactMana(t, g, artifactSpellDef()) {
			t.Fatal("cast artifact spell with cast-or-activate-artifact mana failed")
		}
		if player.ManaPool.Amount(mana.C) != 0 {
			t.Fatalf("cast-or-activate-artifact mana not spent: pool=%d", player.ManaPool.Amount(mana.C))
		}
	})

	t.Run("accepts artifact ability", func(t *testing.T) {
		t.Parallel()
		g, player := gameWithArtifactRestrictedMana(game.ManaSpendCastOrActivateArtifact)
		source := addPermanentForSBA(g, game.Player1, artifactPermanentDef())
		if !payArtifactAbility(g, source) {
			t.Fatal("cast-or-activate-artifact mana failed to pay an artifact ability")
		}
		if player.ManaPool.Amount(mana.C) != 0 || len(player.ManaRiders) != 0 {
			t.Fatalf("rider not consumed by artifact ability: pool=%d riders=%d", player.ManaPool.Amount(mana.C), len(player.ManaRiders))
		}
	})

	t.Run("rejects nonartifact ability", func(t *testing.T) {
		t.Parallel()
		g, player := gameWithArtifactRestrictedMana(game.ManaSpendCastOrActivateArtifact)
		source := addPermanentForSBA(g, game.Player1, nonartifactSpellDef())
		if payArtifactAbility(g, source) {
			t.Fatal("cast-or-activate-artifact mana paid a nonartifact ability")
		}
		if player.ManaPool.Amount(mana.C) != 1 || len(player.ManaRiders) != 1 {
			t.Fatalf("restricted mana changed after rejected payment: pool=%d riders=%d", player.ManaPool.Amount(mana.C), len(player.ManaRiders))
		}
	})
}

// TestActivateArtifactAbilityManaRestriction covers Soldevi Machinist: the
// tagged mana may only activate an artifact's ability, never cast a spell.
func TestActivateArtifactAbilityManaRestriction(t *testing.T) {
	t.Parallel()

	t.Run("rejects artifact spell", func(t *testing.T) {
		t.Parallel()
		g, player := gameWithArtifactRestrictedMana(game.ManaSpendActivateArtifactAbility)
		if castWithArtifactMana(t, g, artifactSpellDef()) {
			t.Fatal("cast artifact spell with activate-artifact-ability mana succeeded")
		}
		if player.ManaPool.Amount(mana.C) != 1 || len(player.ManaRiders) != 1 {
			t.Fatalf("restricted mana changed after rejected cast: pool=%d riders=%d", player.ManaPool.Amount(mana.C), len(player.ManaRiders))
		}
	})

	t.Run("accepts artifact ability", func(t *testing.T) {
		t.Parallel()
		g, player := gameWithArtifactRestrictedMana(game.ManaSpendActivateArtifactAbility)
		source := addPermanentForSBA(g, game.Player1, artifactPermanentDef())
		if !payArtifactAbility(g, source) {
			t.Fatal("activate-artifact-ability mana failed to pay an artifact ability")
		}
		if player.ManaPool.Amount(mana.C) != 0 || len(player.ManaRiders) != 0 {
			t.Fatalf("rider not consumed by artifact ability: pool=%d riders=%d", player.ManaPool.Amount(mana.C), len(player.ManaRiders))
		}
	})
}

// TestCastArtifactOrActivateAbilityManaRestriction covers Guidelight Optimizer /
// Automated Artificer: cast an artifact spell or activate any ability.
func TestCastArtifactOrActivateAbilityManaRestriction(t *testing.T) {
	t.Parallel()

	t.Run("rejects nonartifact spell", func(t *testing.T) {
		t.Parallel()
		g, player := gameWithArtifactRestrictedMana(game.ManaSpendCastArtifactOrActivateAbility)
		if castWithArtifactMana(t, g, nonartifactSpellDef()) {
			t.Fatal("cast nonartifact spell with cast-artifact-or-activate-ability mana succeeded")
		}
		if player.ManaPool.Amount(mana.C) != 1 || len(player.ManaRiders) != 1 {
			t.Fatalf("restricted mana changed after rejected cast: pool=%d riders=%d", player.ManaPool.Amount(mana.C), len(player.ManaRiders))
		}
	})

	t.Run("accepts artifact spell", func(t *testing.T) {
		t.Parallel()
		g, player := gameWithArtifactRestrictedMana(game.ManaSpendCastArtifactOrActivateAbility)
		if !castWithArtifactMana(t, g, artifactSpellDef()) {
			t.Fatal("cast artifact spell with cast-artifact-or-activate-ability mana failed")
		}
		if player.ManaPool.Amount(mana.C) != 0 {
			t.Fatalf("cast-artifact-or-activate-ability mana not spent: pool=%d", player.ManaPool.Amount(mana.C))
		}
	})

	t.Run("accepts nonartifact ability", func(t *testing.T) {
		t.Parallel()
		g, player := gameWithArtifactRestrictedMana(game.ManaSpendCastArtifactOrActivateAbility)
		source := addPermanentForSBA(g, game.Player1, nonartifactSpellDef())
		if !payArtifactAbility(g, source) {
			t.Fatal("cast-artifact-or-activate-ability mana failed to pay a nonartifact ability")
		}
		if player.ManaPool.Amount(mana.C) != 0 || len(player.ManaRiders) != 0 {
			t.Fatalf("rider not consumed by ability: pool=%d riders=%d", player.ManaPool.Amount(mana.C), len(player.ManaRiders))
		}
	})
}
