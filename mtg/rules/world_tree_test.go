package rules

import (
	"reflect"
	"slices"
	"testing"

	cards "github.com/natefinch/council4/mtg/cards/t"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

// TestWorldTreeConditionalLandGrantTogglesAtSixLands proves The World Tree's
// conditional static grant — "As long as you control six or more lands, lands
// you control have '{T}: Add one mana of any color.'" — using the real generated
// card. The grant is inactive while its controller controls five lands, becomes
// active for every land they control (including The World Tree itself, since the
// group is "lands you control", not "other lands") once a sixth land arrives,
// and switches back off when land control drops below six. Lands an opponent
// controls and the controller's nonland permanents never receive the ability.
func TestWorldTreeConditionalLandGrantTogglesAtSixLands(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	worldTree := addCombatPermanent(g, game.Player1, cards.TheWorldTree())
	forests := make([]*game.Permanent, 0, 5)
	for range 4 {
		forests = append(forests, addLandPermanent(g, game.Player1, "Forest", types.Forest))
	}

	// Five lands controlled (The World Tree + four Forests): grant inactive.
	for _, forest := range forests {
		if got := countAnyColorManaAbilities(g, forest); got != 0 {
			t.Fatalf("controlled land any-color grants at five lands = %d, want 0", got)
		}
	}
	if got := countAnyColorManaAbilities(g, worldTree); got != 0 {
		t.Fatalf("The World Tree self any-color grants at five lands = %d, want 0", got)
	}

	// A sixth controlled land activates the grant for every land the controller
	// controls, The World Tree included.
	sixth := addLandPermanent(g, game.Player1, "Forest", types.Forest)
	for _, forest := range append(slices.Clone(forests), sixth) {
		if got := countAnyColorManaAbilities(g, forest); got != 1 {
			t.Fatalf("controlled land any-color grants at six lands = %d, want 1", got)
		}
	}
	if got := countAnyColorManaAbilities(g, worldTree); got != 1 {
		t.Fatalf("The World Tree self any-color grants at six lands = %d, want 1", got)
	}

	// An opponent's land and a controlled nonland are outside the group.
	opponentLand := addLandPermanent(g, game.Player2, "Island", types.Island)
	if got := countAnyColorManaAbilities(g, opponentLand); got != 0 {
		t.Fatalf("opponent land any-color grants = %d, want 0", got)
	}
	controlledNonland := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Bear",
		Types: []types.Card{types.Creature},
	}})
	if got := countAnyColorManaAbilities(g, controlledNonland); got != 0 {
		t.Fatalf("controlled nonland any-color grants = %d, want 0", got)
	}

	// Dropping back below six lands deactivates the grant again: it tracks land
	// control dynamically rather than latching on.
	g.Battlefield = slices.DeleteFunc(g.Battlefield, func(permanent *game.Permanent) bool {
		return permanent == sixth
	})
	for _, forest := range forests {
		if got := countAnyColorManaAbilities(g, forest); got != 0 {
			t.Fatalf("controlled land any-color grants after dropping below six = %d, want 0", got)
		}
	}
	if got := countAnyColorManaAbilities(g, worldTree); got != 0 {
		t.Fatalf("The World Tree self any-color grants after dropping below six = %d, want 0", got)
	}
}

// TestWorldTreePrintedAndActivatedAbilities verifies the non-grant halves of the
// real generated card: it enters tapped, taps for {G}, and carries the full
// activation cost ({W}{W}{U}{U}{B}{B}{R}{R}{G}{G}, tap, sacrifice this land) for
// the "any number of God cards" battlefield search.
func TestWorldTreePrintedAndActivatedAbilities(t *testing.T) {
	def := cards.TheWorldTree()

	wantEntersTapped := game.EntersTappedReplacement("This land enters tapped.")
	if !slices.ContainsFunc(def.ReplacementAbilities, func(replacement game.ReplacementAbility) bool {
		return reflect.DeepEqual(replacement, wantEntersTapped)
	}) {
		t.Errorf("The World Tree is missing its enters-tapped replacement; has %+v", def.ReplacementAbilities)
	}

	wantBaseMana := game.TapManaAbility(mana.G)
	if !slices.ContainsFunc(def.ManaAbilities, func(ability game.ManaAbility) bool {
		return reflect.DeepEqual(ability, wantBaseMana)
	}) {
		t.Errorf("The World Tree is missing its {T}: Add {G} mana ability; has %+v", def.ManaAbilities)
	}

	if len(def.ActivatedAbilities) != 1 {
		t.Fatalf("The World Tree activated abilities = %d, want 1", len(def.ActivatedAbilities))
	}
	activation := def.ActivatedAbilities[0]

	wantMana := cost.Mana{cost.W, cost.W, cost.U, cost.U, cost.B, cost.B, cost.R, cost.R, cost.G, cost.G}
	if !activation.ManaCost.Exists || !reflect.DeepEqual(activation.ManaCost.Val, wantMana) {
		t.Errorf("activation mana cost = %+v (exists=%v), want %+v", activation.ManaCost.Val, activation.ManaCost.Exists, wantMana)
	}
	if !slices.ContainsFunc(activation.AdditionalCosts, func(additional cost.Additional) bool {
		return additional.Kind == cost.AdditionalTap
	}) {
		t.Errorf("activation is missing its tap cost; has %+v", activation.AdditionalCosts)
	}
	if !slices.ContainsFunc(activation.AdditionalCosts, func(additional cost.Additional) bool {
		return additional.Kind == cost.AdditionalSacrificeSource
	}) {
		t.Errorf("activation is missing its sacrifice-this-land cost; has %+v", activation.AdditionalCosts)
	}

	if len(activation.Content.Modes) != 1 || len(activation.Content.Modes[0].Sequence) != 1 {
		t.Fatalf("activation content shape = %d modes, want a single one-instruction mode", len(activation.Content.Modes))
	}
	search, ok := activation.Content.Modes[0].Sequence[0].Primitive.(game.Search)
	if !ok {
		t.Fatalf("activation instruction = %T, want game.Search", activation.Content.Modes[0].Sequence[0].Primitive)
	}
	if !search.Spec.AnyNumber {
		t.Error("activation search is not marked AnyNumber")
	}
	if search.Spec.SourceZone != zone.Library || search.Spec.Destination != zone.Battlefield {
		t.Errorf("activation search zones = %v -> %v, want library -> battlefield", search.Spec.SourceZone, search.Spec.Destination)
	}
	if !slices.Contains(search.Spec.Filter.SubtypesAny, types.God) {
		t.Errorf("activation search filter subtypes = %+v, want to include God", search.Spec.Filter.SubtypesAny)
	}
}
