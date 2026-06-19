package rules

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
)

// forestGrantingStaticPermanent adds a permanent whose static ability makes
// every land on the battlefield a Forest in addition to its other land types,
// mirroring the continuous "Each land is a <basic land type>" cluster (Yavimaya,
// Cradle of Growth). The source is an enchantment so the static itself is not a
// land, isolating the effect on the separately added lands under test.
func forestGrantingStaticPermanent(g *game.Game, controller game.PlayerID) *game.Permanent {
	return addCombatPermanent(g, controller, &game.CardDef{CardFace: game.CardFace{
		Name:  "Verdant Mantle",
		Types: []types.Card{types.Enchantment},
		StaticAbilities: []game.StaticAbility{{
			ContinuousEffects: []game.ContinuousEffect{{
				Layer: game.LayerType,
				Group: game.BattlefieldGroup(game.Selection{
					RequiredTypes: []types.Card{types.Land},
				}),
				AddSubtypes: []types.Sub{types.Forest},
			}},
		}},
	}})
}

// addBasicLandWithManaPermanent adds a printed basic land carrying its intrinsic mana
// ability, matching how basic lands enter from card faces.
func addBasicLandWithManaPermanent(g *game.Game, controller game.PlayerID, name string, subtype types.Sub, manaColor mana.Color) *game.Permanent {
	return addCombatPermanent(g, controller, &game.CardDef{CardFace: game.CardFace{
		Name:          name,
		Supertypes:    []types.Super{types.Basic},
		Types:         []types.Card{types.Land},
		Subtypes:      []types.Sub{subtype},
		ManaAbilities: []game.ManaAbility{game.TapManaAbility(manaColor)},
	}})
}

func greenManaAbilityIndex(g *game.Game, permanent *game.Permanent) (int, bool) {
	for idx, ability := range permanentEffectiveAbilities(g, permanent) {
		body, ok := ability.(*game.ManaAbility)
		if !ok {
			continue
		}
		_, colors := abilitiesManaProduction([]game.Ability{body}, nil)
		if slices.Contains(colors, color.Green) {
			return idx, true
		}
	}
	return 0, false
}

// TestAddedForestSubtypeGrantsGreenManaAbility proves the runtime models the
// land-type-adding continuous static end to end: a Plains affected by a
// Forest-granting static gains the Forest subtype and the intrinsic green mana
// ability that subtype confers, so the Plains can actually tap for green while
// still tapping for its printed white.
func TestAddedForestSubtypeGrantsGreenManaAbility(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	plains := addBasicLandWithManaPermanent(g, game.Player1, "Plains", types.Plains, mana.W)
	forestGrantingStaticPermanent(g, game.Player1)

	values := effectivePermanentValues(g, plains)
	if !slices.Contains(values.subtypes, types.Forest) {
		t.Fatalf("effective subtypes = %v, want to contain Forest", values.subtypes)
	}
	if !slices.Contains(values.subtypes, types.Plains) {
		t.Fatalf("effective subtypes = %v, want to retain Plains", values.subtypes)
	}

	_, colors := abilitiesManaProduction(permanentEffectiveAbilities(g, plains), nil)
	if !slices.Contains(colors, color.White) {
		t.Fatalf("effective mana colors = %v, want to retain white", colors)
	}
	if !slices.Contains(colors, color.Green) {
		t.Fatalf("effective mana colors = %v, want to gain green", colors)
	}

	idx, ok := greenManaAbilityIndex(g, plains)
	if !ok {
		t.Fatal("no green mana ability found on the affected Plains")
	}
	act := action.ActivateAbility(plains.ObjectID, idx, nil, 0)
	if containsAction(engine.legalActions(g, game.Player1), act) {
		t.Fatalf("payment-only green mana ability activation %+v was exposed as a standalone action", act)
	}
	if !engine.applyAction(g, game.Player1, act) {
		t.Fatal("applyAction(granted green mana ability) = false, want true")
	}
	if got := g.Players[game.Player1].ManaPool.Amount(mana.G); got != 1 {
		t.Fatalf("green mana = %d, want 1", got)
	}
}

// TestAddedForestSubtypeDoesNotDuplicatePrintedForestMana confirms the added-only
// scoping: a printed Forest already carries its green mana ability, so the same
// Forest-granting static must not append a second one. The Forest keeps exactly
// its single printed mana ability.
func TestAddedForestSubtypeDoesNotDuplicatePrintedForestMana(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	forest := addBasicLandWithManaPermanent(g, game.Player1, "Forest", types.Forest, mana.G)
	forestGrantingStaticPermanent(g, game.Player1)

	manaAbilities := 0
	for _, ability := range permanentEffectiveAbilities(g, forest) {
		if _, ok := ability.(*game.ManaAbility); ok {
			manaAbilities++
		}
	}
	if manaAbilities != 1 {
		t.Fatalf("printed Forest mana abilities = %d, want 1 (no duplicate)", manaAbilities)
	}
}

// TestForestGrantingStaticIgnoresNonLands confirms the static fails closed for
// permanents outside its land group: a creature gains neither the Forest subtype
// nor a green mana ability.
func TestForestGrantingStaticIgnoresNonLands(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	creature := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Plain Bear",
		Types: []types.Card{types.Creature},
	}})
	forestGrantingStaticPermanent(g, game.Player1)

	values := effectivePermanentValues(g, creature)
	if slices.Contains(values.subtypes, types.Forest) {
		t.Fatalf("creature effective subtypes = %v, want no Forest", values.subtypes)
	}
	if _, ok := greenManaAbilityIndex(g, creature); ok {
		t.Fatal("creature unexpectedly gained a green mana ability")
	}
}
