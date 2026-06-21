package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

func gainLifeQuantity(t *testing.T, content game.AbilityContent) game.Quantity {
	t.Helper()
	primitive, ok := content.Modes[0].Sequence[0].Primitive.(game.GainLife)
	if !ok {
		t.Fatalf("primitive = %T, want game.GainLife", content.Modes[0].Sequence[0].Primitive)
	}
	return primitive.Amount
}

func lowerLifeFace(t *testing.T, oracleText string) loweredFaceAbilities {
	t.Helper()
	return lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Chalice",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: oracleText,
	})
}

func TestLowerDynamicLifeBattlefieldSelectionCounts(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name         string
		oracleText   string
		multiplier   int
		requiredType types.Card
		subtypes     []types.Sub
		supertypes   []types.Super
		colorless    bool
		controller   game.ControllerRelation
		tapped       game.TriState
	}{
		{
			name:       "subtype Shrine",
			oracleText: "You gain 2 life for each Shrine you control.",
			multiplier: 2,
			subtypes:   []types.Sub{types.Shrine},
			controller: game.ControllerYou,
		},
		{
			name:         "tapped creature you control",
			oracleText:   "You gain 1 life for each tapped creature you control.",
			multiplier:   1,
			requiredType: types.Creature,
			controller:   game.ControllerYou,
			tapped:       game.TriTrue,
		},
		{
			name:         "untapped creature you control",
			oracleText:   "You gain 1 life for each untapped creature you control.",
			multiplier:   1,
			requiredType: types.Creature,
			controller:   game.ControllerYou,
			tapped:       game.TriFalse,
		},
		{
			name:         "creature an opponent controls",
			oracleText:   "You gain 1 life for each creature your opponents control.",
			multiplier:   1,
			requiredType: types.Creature,
			controller:   game.ControllerOpponent,
		},
		{
			name:         "colorless creature",
			oracleText:   "You gain 1 life for each colorless creature you control.",
			multiplier:   1,
			requiredType: types.Creature,
			colorless:    true,
			controller:   game.ControllerYou,
		},
		{
			name:         "legendary creature supertype",
			oracleText:   "You gain 1 life for each legendary creature you control.",
			multiplier:   1,
			requiredType: types.Creature,
			supertypes:   []types.Super{types.Legendary},
			controller:   game.ControllerYou,
		},
		{
			name:         "snow land supertype",
			oracleText:   "You gain 1 life for each snow land you control.",
			multiplier:   1,
			requiredType: types.Land,
			supertypes:   []types.Super{types.Snow},
			controller:   game.ControllerYou,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			face := lowerLifeFace(t, test.oracleText)
			dynamic := gainLifeQuantity(t, face.SpellAbility.Val).DynamicAmount()
			if !dynamic.Exists ||
				dynamic.Val.Kind != game.DynamicAmountCountSelector ||
				dynamic.Val.Multiplier != test.multiplier {
				t.Fatalf("dynamic amount = %+v", dynamic)
			}
			selection := dynamic.Val.Group.Selection()
			if selection.Controller != test.controller {
				t.Fatalf("controller = %v, want %v", selection.Controller, test.controller)
			}
			if test.requiredType != "" {
				if len(selection.RequiredTypes) != 1 || selection.RequiredTypes[0] != test.requiredType {
					t.Fatalf("required types = %v, want [%v]", selection.RequiredTypes, test.requiredType)
				}
			} else if len(selection.RequiredTypes) != 0 {
				t.Fatalf("required types = %v, want none", selection.RequiredTypes)
			}
			if !subtypesEqual(selection.SubtypesAny, test.subtypes) {
				t.Fatalf("subtypes = %v, want %v", selection.SubtypesAny, test.subtypes)
			}
			if !supertypesEqual(selection.Supertypes, test.supertypes) {
				t.Fatalf("supertypes = %v, want %v", selection.Supertypes, test.supertypes)
			}
			if selection.Colorless != test.colorless {
				t.Fatalf("colorless = %v, want %v", selection.Colorless, test.colorless)
			}
			if selection.Tapped != test.tapped {
				t.Fatalf("tapped = %v, want %v", selection.Tapped, test.tapped)
			}
		})
	}
}

func TestLowerDynamicLifeCharacteristicFilteredCounts(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		oracleText string
		multiplier int
		op         compare.Op
		value      int
	}{
		{
			name:       "power 4 or greater",
			oracleText: "You gain 4 life for each creature you control with power 4 or greater.",
			multiplier: 4,
			op:         compare.GreaterOrEqual,
			value:      4,
		},
		{
			name:       "power 2 or less",
			oracleText: "You gain 1 life for each creature you control with power 2 or less.",
			multiplier: 1,
			op:         compare.LessOrEqual,
			value:      2,
		},
		{
			name:       "mana value equal to N",
			oracleText: "You gain 1 life for each creature you control with mana value equal to 3.",
			multiplier: 1,
			op:         compare.Equal,
			value:      3,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			face := lowerLifeFace(t, test.oracleText)
			dynamic := gainLifeQuantity(t, face.SpellAbility.Val).DynamicAmount()
			if !dynamic.Exists ||
				dynamic.Val.Kind != game.DynamicAmountCountSelector ||
				dynamic.Val.Multiplier != test.multiplier {
				t.Fatalf("dynamic amount = %+v", dynamic)
			}
			selection := dynamic.Val.Group.Selection()
			if len(selection.RequiredTypes) != 1 || selection.RequiredTypes[0] != types.Creature {
				t.Fatalf("required types = %v, want [Creature]", selection.RequiredTypes)
			}
			filter := selection.Power
			if test.name == "mana value equal to N" {
				filter = selection.ManaValue
			}
			if !filter.Exists || filter.Val.Op != test.op || filter.Val.Value != test.value {
				t.Fatalf("characteristic filter = %+v, want {Op:%v Value:%d}", filter, test.op, test.value)
			}
		})
	}
}

func TestLowerDynamicLifeZoneCounts(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name         string
		oracleText   string
		multiplier   int
		cardZone     zone.Type
		requiredType types.Card
		subtypes     []types.Sub
	}{
		{
			name:       "cards in hand",
			oracleText: "You gain 2 life for each card in your hand.",
			multiplier: 2,
			cardZone:   zone.Hand,
		},
		{
			name:       "Elf cards in graveyard",
			oracleText: "You gain 1 life for each Elf card in your graveyard.",
			multiplier: 1,
			cardZone:   zone.Graveyard,
			subtypes:   []types.Sub{types.Elf},
		},
		{
			name:         "creature cards in graveyard",
			oracleText:   "You gain 1 life for each creature card in your graveyard.",
			multiplier:   1,
			cardZone:     zone.Graveyard,
			requiredType: types.Creature,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			face := lowerLifeFace(t, test.oracleText)
			dynamic := gainLifeQuantity(t, face.SpellAbility.Val).DynamicAmount()
			if !dynamic.Exists ||
				dynamic.Val.Kind != game.DynamicAmountCountCardsInZone ||
				dynamic.Val.Multiplier != test.multiplier ||
				dynamic.Val.CardZone != test.cardZone {
				t.Fatalf("dynamic amount = %+v", dynamic)
			}
			if dynamic.Val.Player == nil {
				t.Fatal("zone count missing player reference")
			}
			if dynamic.Val.Selection == nil {
				t.Fatal("zone count missing selection")
			}
			selection := dynamic.Val.Selection
			if test.requiredType != "" {
				if len(selection.RequiredTypes) != 1 || selection.RequiredTypes[0] != test.requiredType {
					t.Fatalf("required types = %v, want [%v]", selection.RequiredTypes, test.requiredType)
				}
			} else if len(selection.RequiredTypes) != 0 {
				t.Fatalf("required types = %v, want none", selection.RequiredTypes)
			}
			if !subtypesEqual(selection.SubtypesAny, test.subtypes) {
				t.Fatalf("subtypes = %v, want %v", selection.SubtypesAny, test.subtypes)
			}
		})
	}
}

func TestLowerDynamicLifeFailsClosed(t *testing.T) {
	t.Parallel()
	for _, oracleText := range []string{
		"You gain 1 life for each attacking creature you control.",
		"You gain 1 life for each another creature you control.",
		"You gain 1 life for each instant card in your graveyard.",
		"You gain 1 life for each sorcery card in your graveyard.",
		"You gain 1 life for each artifact creature you control.",
	} {
		t.Run(oracleText, func(t *testing.T) {
			t.Parallel()
			source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
				Name:       "Test Chalice",
				Layout:     "normal",
				TypeLine:   "Sorcery",
				OracleText: oracleText,
			}, "t")
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if source != "" || len(diagnostics) == 0 {
				t.Fatalf("source = %q, diagnostics = %#v", source, diagnostics)
			}
			if got := diagnostics[0].Summary; got != "unsupported life spell" {
				t.Fatalf("summary = %q, want unsupported life spell", got)
			}
		})
	}
}

func subtypesEqual(got, want []types.Sub) bool {
	if len(got) != len(want) {
		return false
	}
	for i := range want {
		if got[i] != want[i] {
			return false
		}
	}
	return true
}

func supertypesEqual(got, want []types.Super) bool {
	if len(got) != len(want) {
		return false
	}
	for i := range want {
		if got[i] != want[i] {
			return false
		}
	}
	return true
}
