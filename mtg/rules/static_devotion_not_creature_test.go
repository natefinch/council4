package rules

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// devotionGodDef builds a God-like permanent: a legendary enchantment creature
// carrying the devotion-gated "isn't a creature" static that removes the
// creature card type while its controller's devotion to the given colors is
// below the threshold. It mirrors the runtime shape the cardgen lowering emits
// for the Theros Gods.
func devotionGodDef(name string, manaCost cost.Mana, colors []color.Color, threshold int) *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:       name,
		Supertypes: []types.Super{types.Legendary},
		Types:      []types.Card{types.Enchantment, types.Creature},
		Subtypes:   []types.Sub{types.God},
		ManaCost:   opt.Val(manaCost),
		Power:      opt.Val(game.PT{Value: 6}),
		Toughness:  opt.Val(game.PT{Value: 5}),
		StaticAbilities: []game.StaticAbility{{
			Text: "devotion isn't a creature",
			Condition: opt.Val(game.Condition{
				Aggregates: []game.AggregateComparison{{
					Aggregate: game.AggregateControllerDevotion,
					Colors:    colors,
					Op:        compare.LessThan,
					Value:     threshold,
				}},
			}),
			ContinuousEffects: []game.ContinuousEffect{{
				Layer:          game.LayerType,
				AffectedSource: true,
				RemoveTypes:    []types.Card{types.Creature},
			}},
		}},
	}}
}

// redPipDef builds a vanilla permanent with a single red mana pip so tests can
// raise the controller's red devotion one point at a time.
func redPipDef(name string) *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:     name,
		Types:    []types.Card{types.Enchantment},
		ManaCost: opt.Val(cost.Mana{cost.R}),
	}}
}

func hasCardType(values permanentEffectiveValues, cardType types.Card) bool {
	return slices.Contains(values.types, cardType)
}

// TestDevotionNotCreatureBelowThresholdRemovesCreatureType proves a mono-red God
// with only its own single red pip (devotion 1 < 5) is not a creature but
// remains a legendary enchantment.
func TestDevotionNotCreatureBelowThresholdRemovesCreatureType(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	god := addCombatPermanent(g, game.Player1, devotionGodDef("Purphoros", cost.Mana{cost.O(3), cost.R}, []color.Color{color.Red}, 5))

	values := effectivePermanentValues(g, god)
	if hasCardType(values, types.Creature) {
		t.Fatalf("god is a creature at devotion 1, want non-creature; types=%v", values.types)
	}
	if !hasCardType(values, types.Enchantment) {
		t.Fatalf("god lost its enchantment type; types=%v", values.types)
	}
}

// TestDevotionNotCreatureAtThresholdIsCreature proves reaching devotion equal to
// the threshold (five red pips) makes the God a creature again: the "less than
// five" gate closes at exactly five.
func TestDevotionNotCreatureAtThresholdIsCreature(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	god := addCombatPermanent(g, game.Player1, devotionGodDef("Purphoros", cost.Mana{cost.O(3), cost.R}, []color.Color{color.Red}, 5))
	// The god already contributes one red pip; add four more for devotion five.
	for range 4 {
		addCombatPermanent(g, game.Player1, redPipDef("Mountain Spirit"))
	}

	values := effectivePermanentValues(g, god)
	if !hasCardType(values, types.Creature) {
		t.Fatalf("god is not a creature at devotion 5, want creature; types=%v", values.types)
	}
}

// TestDevotionNotCreatureRecomputesAcrossThreshold proves the type-changing
// static recomputes live as devotion crosses the threshold in both directions:
// the same permanent flips between creature and non-creature as red pips are
// added and removed, with no stale caching.
func TestDevotionNotCreatureRecomputesAcrossThreshold(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	god := addCombatPermanent(g, game.Player1, devotionGodDef("Purphoros", cost.Mana{cost.O(3), cost.R}, []color.Color{color.Red}, 5))

	if hasCardType(effectivePermanentValues(g, god), types.Creature) {
		t.Fatal("god should not be a creature at devotion 1")
	}
	// Raise devotion to five: now a creature.
	added := make([]*game.Permanent, 0, 4)
	for range 4 {
		added = append(added, addCombatPermanent(g, game.Player1, redPipDef("Mountain Spirit")))
	}
	if !hasCardType(effectivePermanentValues(g, god), types.Creature) {
		t.Fatal("god should be a creature at devotion 5")
	}
	// Remove one pip source: devotion drops to four, no longer a creature.
	last := added[len(added)-1]
	g.Battlefield = slices.DeleteFunc(g.Battlefield, func(p *game.Permanent) bool { return p == last })
	if hasCardType(effectivePermanentValues(g, god), types.Creature) {
		t.Fatal("god should stop being a creature when devotion falls to 4")
	}
}

// TestDevotionNotCreatureCountsOnlyControllerPips proves devotion is
// controller-scoped: an opponent's red pips do not raise the God's controller's
// devotion, so the God stays a non-creature.
func TestDevotionNotCreatureCountsOnlyControllerPips(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	god := addCombatPermanent(g, game.Player1, devotionGodDef("Purphoros", cost.Mana{cost.O(3), cost.R}, []color.Color{color.Red}, 5))
	for range 6 {
		addCombatPermanent(g, game.Player2, redPipDef("Enemy Ember"))
	}

	if hasCardType(effectivePermanentValues(g, god), types.Creature) {
		t.Fatal("opponent pips must not raise controller devotion; god should stay a non-creature")
	}
}

// TestDevotionNotCreatureTwoColorThreshold proves the two-color God gate counts
// each qualifying pip once across both colors and uses the seven threshold: six
// mixed white/black pips leave it a non-creature, and the seventh makes it a
// creature.
func TestDevotionNotCreatureTwoColorThreshold(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	// {1}{W}{B}: one white and one black pip from the god itself (devotion 2).
	god := addCombatPermanent(g, game.Player1, devotionGodDef(
		"Athreos", cost.Mana{cost.O(1), cost.W, cost.B}, []color.Color{color.White, color.Black}, 7))
	whitePip := &game.CardDef{CardFace: game.CardFace{Name: "W", Types: []types.Card{types.Enchantment}, ManaCost: opt.Val(cost.Mana{cost.W})}}
	blackPip := &game.CardDef{CardFace: game.CardFace{Name: "B", Types: []types.Card{types.Enchantment}, ManaCost: opt.Val(cost.Mana{cost.B})}}
	// Add four more pips (two white, two black) for devotion 6: still not a creature.
	addCombatPermanent(g, game.Player1, whitePip)
	addCombatPermanent(g, game.Player1, whitePip)
	addCombatPermanent(g, game.Player1, blackPip)
	addCombatPermanent(g, game.Player1, blackPip)
	if hasCardType(effectivePermanentValues(g, god), types.Creature) {
		t.Fatal("god should not be a creature at devotion 6")
	}
	// Seventh pip crosses the threshold.
	addCombatPermanent(g, game.Player1, blackPip)
	if !hasCardType(effectivePermanentValues(g, god), types.Creature) {
		t.Fatal("god should be a creature at devotion 7")
	}
}
