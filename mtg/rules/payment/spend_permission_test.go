package payment

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// TestSpendPermissionSymbolsDistinguishesAnyColorFromAnyType proves the
// spending permission changes only which mana may pay each requirement
// (CR 609.4b): colored, hybrid, Twobrid, and colored Phyrexian requirements
// become payable with one mana of any type under either permission, {C} stays
// colorless-only under "as though it were mana of any color" (CR 107.4c) but
// not under "mana of any type" (CR 118.14), and generic, X, and snow
// requirements (CR 107.4h) are unchanged. The order is preserved so Phyrexian
// payment choices keep their positions.
func TestSpendPermissionSymbolsDistinguishesAnyColorFromAnyType(t *testing.T) {
	t.Parallel()
	printed := []cost.Symbol{
		cost.O(2), cost.B, cost.C, cost.HybridMana(mana.W, mana.U),
		cost.Twobrid(mana.R), cost.PhyrexianMana(mana.G), cost.X, cost.S,
	}
	if got := spendPermissionSymbols(printed, game.ManaSpendNormal); !slices.Equal(got, printed) {
		t.Fatalf("normal spending = %v, want the cost unchanged", got)
	}
	anyColor := []cost.Symbol{
		cost.O(2), cost.O(1), cost.C, cost.O(1),
		cost.O(1), cost.PhyrexianGeneric(1), cost.X, cost.S,
	}
	if got := spendPermissionSymbols(printed, game.ManaSpendAsAnyColor); !slices.Equal(got, anyColor) {
		t.Fatalf("any-color spending = %v, want %v", got, anyColor)
	}
	anyType := slices.Clone(anyColor)
	anyType[2] = cost.O(1)
	if got := spendPermissionSymbols(printed, game.ManaSpendAsAnyType); !slices.Equal(got, anyType) {
		t.Fatalf("any-type spending = %v, want %v", got, anyType)
	}
	if !slices.Equal(printed, []cost.Symbol{
		cost.O(2), cost.B, cost.C, cost.HybridMana(mana.W, mana.U),
		cost.Twobrid(mana.R), cost.PhyrexianMana(mana.G), cost.X, cost.S,
	}) {
		t.Fatal("the spending permission mutated the cost it was given")
	}
}

// TestSpellRequestSpendPermissionCoversEveryCostOption proves the request's
// spending permission travels with every cost option instead of becoming an
// alternative cost of its own: the normal and evoke options both remain, and
// each is planned with the permission.
func TestSpellRequestSpendPermissionCoversEveryCostOption(t *testing.T) {
	t.Parallel()
	card := &game.CardDef{CardFace: game.CardFace{
		Name:     "Evoke Rider Probe",
		ManaCost: opt.Val(cost.Mana{cost.O(2), cost.W}),
		Types:    []types.Card{types.Creature},
		AlternativeCosts: []cost.Alternative{{
			Label:    "Evoke",
			Mechanic: cost.AlternativeMechanicEvoke,
			ManaCost: opt.Val(cost.Mana{cost.W}),
		}},
	}}
	options := spellCostOptionsForRequest(fakePaymentState{}, SpellRequest{
		PlayerID:     game.Player1,
		Card:         card,
		SourceZone:   zone.Exile,
		SpendAnyMana: game.ManaSpendAsAnyColor,
	})
	if len(options) != 2 {
		t.Fatalf("cost options = %d, want the normal and evoke options", len(options))
	}
	for _, option := range options {
		if option.spendAs != game.ManaSpendAsAnyColor || option.spendContext().spendAs != game.ManaSpendAsAnyColor {
			t.Fatalf("option %q spends as %v, want any color", option.label, option.spendAs)
		}
	}
	if _, ok := spellOptionByLabel(options, "Evoke"); !ok {
		t.Fatal("the spending permission hid the evoke alternative")
	}
}
