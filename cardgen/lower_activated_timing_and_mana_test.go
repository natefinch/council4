package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
)

func TestLowerActivatedAbilityTiming(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		oracleText string
		want       game.TimingRestriction
	}{
		{"sorcery", "{1}: Draw a card. Activate only as a sorcery.", game.SorceryOnly},
		{"once per turn", "{1}: Draw a card. Activate only once each turn.", game.OncePerTurn},
		{"combat", "{1}: Draw a card. Activate only during combat.", game.DuringCombat},
		{"upkeep", "{1}: Draw a card. Activate only during your upkeep.", game.DuringUpkeep},
		{"during your turn", "{1}: Draw a card. Activate only during your turn.", game.DuringYourTurn},
		{"sorcery speed variant", "{1}: Draw a card. Activate only at sorcery speed.", game.SorceryOnly},
		{"cast sorcery variant", "{1}: Draw a card. Activate only any time you could cast a sorcery.", game.SorceryOnly},
		{"per turn variant", "{1}: Draw a card. Activate only once per turn.", game.OncePerTurn},
		{"each combat variant", "{1}: Draw a card. Activate only during each combat.", game.DuringCombat},
		{"each controller upkeep variant", "{1}: Draw a card. Activate only during each of your upkeeps.", game.DuringUpkeep},
		{
			"sorcery once per turn",
			"{1}: Draw a card. Activate only as a sorcery. Activate only once each turn.",
			game.SorceryOncePerTurn,
		},
		{
			"sorcery once per turn conjoined",
			"{1}: Draw a card. Activate only as a sorcery and only once each turn.",
			game.SorceryOncePerTurn,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			face := lowerSingleFace(t, &ScryfallCard{
				Name:       "Test Engine",
				Layout:     "normal",
				TypeLine:   "Artifact",
				OracleText: test.oracleText,
			})
			if len(face.ActivatedAbilities) != 1 {
				t.Fatalf("activated abilities = %d, want 1", len(face.ActivatedAbilities))
			}
			if got := face.ActivatedAbilities[0].Timing; got != test.want {
				t.Fatalf("timing = %v, want %v", got, test.want)
			}
		})
	}
}

func TestLowerManaAbilityTiming(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Engine",
		Layout:     "normal",
		TypeLine:   "Artifact",
		OracleText: "{T}: Add {G}. Activate only during combat.",
	})
	if len(face.ManaAbilities) != 1 {
		t.Fatalf("mana abilities = %d, want 1", len(face.ManaAbilities))
	}
	if got := face.ManaAbilities[0].Timing; got != game.DuringCombat {
		t.Fatalf("timing = %v, want %v", got, game.DuringCombat)
	}
}

func TestLowerUntapManaAbility(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Engine",
		Layout:     "normal",
		TypeLine:   "Artifact",
		OracleText: "{Q}: Add {G}.",
	})
	if len(face.ManaAbilities) != 1 {
		t.Fatalf("mana abilities = %d, want 1", len(face.ManaAbilities))
	}
	costs := face.ManaAbilities[0].AdditionalCosts
	if len(costs) != 1 || costs[0].Kind != cost.AdditionalUntap {
		t.Fatalf("additional costs = %#v, want untap source", costs)
	}
}

func TestLowerEquipAbility(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Equipment",
		Layout:     "normal",
		TypeLine:   "Artifact — Equipment",
		OracleText: "Equip {2}",
	})
	if len(face.ActivatedAbilities) != 1 {
		t.Fatalf("got %d activated abilities, want 1", len(face.ActivatedAbilities))
	}
	ability := face.ActivatedAbilities[0]
	equipCost, ok := game.ActivatedBodyEquipCost(&ability)
	if !ok || len(equipCost) != 1 || equipCost[0] != cost.O(2) {
		t.Fatalf("equip cost = %#v, %v; want {2}", equipCost, ok)
	}
}

func TestLowerEnchantCreatureAbility(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Aura",
		Layout:     "normal",
		TypeLine:   "Enchantment — Aura",
		OracleText: "Enchant creature",
	})
	if len(face.StaticAbilities) != 1 {
		t.Fatalf("got %d static abilities, want 1", len(face.StaticAbilities))
	}
	target, ok := game.StaticBodyEnchantTarget(&face.StaticAbilities[0].Body)
	if !ok ||
		target.MinTargets != 1 ||
		target.MaxTargets != 1 ||
		target.Allow != game.TargetAllowPermanent ||
		len(target.Selection.Val.RequiredTypesAny) != 1 ||
		target.Selection.Val.RequiredTypesAny[0] != types.Creature {
		t.Fatalf("enchant target = %+v, %v; want one creature", target, ok)
	}
}

func TestLowerProtectionFromColorAbility(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Bear",
		Layout:     "normal",
		TypeLine:   "Creature — Bear",
		OracleText: "Protection from red",
		Power:      new("2"),
		Toughness:  new("2"),
	})
	if len(face.StaticAbilities) != 1 {
		t.Fatalf("got %d static abilities, want 1", len(face.StaticAbilities))
	}
	protected := game.StaticBodyProtectionColors(&face.StaticAbilities[0].Body)
	if len(protected) != 1 || protected[0] != color.Red {
		t.Fatalf("protection colors = %v, want red", protected)
	}
}

func TestLowerProtectionFromColorWithSimpleKeyword(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Bear",
		Layout:     "normal",
		TypeLine:   "Creature — Bear",
		OracleText: "Protection from red, haste",
		Power:      new("2"),
		Toughness:  new("2"),
	})
	if len(face.StaticAbilities) != 2 {
		t.Fatalf("got %d static abilities, want 2", len(face.StaticAbilities))
	}
	protected := game.StaticBodyProtectionColors(&face.StaticAbilities[0].Body)
	if len(protected) != 1 || protected[0] != color.Red {
		t.Fatalf("protection colors = %v, want red", protected)
	}
	if face.StaticAbilities[1].VarName != "game.HasteStaticBody" {
		t.Fatalf("second ability = %+v, want haste", face.StaticAbilities[1])
	}
}

func TestLowerProtectionFromMultipleColors(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Bear",
		Layout:     "normal",
		TypeLine:   "Creature — Bear",
		OracleText: "Protection from black and from red",
		Power:      new("2"),
		Toughness:  new("2"),
	})
	protected := game.StaticBodyProtectionColors(&face.StaticAbilities[0].Body)
	if len(protected) != 2 || protected[0] != color.Black || protected[1] != color.Red {
		t.Fatalf("protection colors = %v, want black and red", protected)
	}
}

func TestLowerEnchantedCreaturePTBuffAlongsideEnchant(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Aura",
		Layout:     "normal",
		TypeLine:   "Enchantment — Aura",
		OracleText: "Enchant creature\nEnchanted creature gets +2/+2.",
	})
	if len(face.StaticAbilities) != 2 {
		t.Fatalf("got %d static abilities, want 2", len(face.StaticAbilities))
	}
	body := face.StaticAbilities[1].Body
	if len(body.ContinuousEffects) != 1 {
		t.Fatalf("got %d continuous effects, want 1", len(body.ContinuousEffects))
	}
	ce := body.ContinuousEffects[0]
	if ce.Layer != game.LayerPowerToughnessModify {
		t.Fatalf("layer = %v, want LayerPowerToughnessModify", ce.Layer)
	}
	if ce.Group.Domain() != game.GroupDomainAttachedObject {
		t.Fatalf("group domain = %v, want GroupDomainAttachedObject", ce.Group.Domain())
	}
	if ce.PowerDelta != 2 || ce.ToughnessDelta != 2 {
		t.Fatalf("PT delta = %d/%d, want 2/2", ce.PowerDelta, ce.ToughnessDelta)
	}
}

func TestLowerEquippedCreaturePTBuff(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Equipment",
		Layout:     "normal",
		TypeLine:   "Artifact — Equipment",
		OracleText: "Equipped creature gets +2/+0.\nEquip {2}",
	})
	if len(face.StaticAbilities) != 1 {
		t.Fatalf("got %d static abilities, want 1", len(face.StaticAbilities))
	}
	body := face.StaticAbilities[0].Body
	if len(body.ContinuousEffects) != 1 {
		t.Fatalf("got %d continuous effects, want 1", len(body.ContinuousEffects))
	}
	ce := body.ContinuousEffects[0]
	if ce.Layer != game.LayerPowerToughnessModify {
		t.Fatalf("layer = %v, want LayerPowerToughnessModify", ce.Layer)
	}
	if ce.Group.Domain() != game.GroupDomainAttachedObject {
		t.Fatalf("group domain = %v, want GroupDomainAttachedObject", ce.Group.Domain())
	}
	if ce.PowerDelta != 2 || ce.ToughnessDelta != 0 {
		t.Fatalf("PT delta = %d/%d, want 2/0", ce.PowerDelta, ce.ToughnessDelta)
	}
}

func TestLowerCreaturesYouControlPTBuff(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Anthem",
		Layout:     "normal",
		TypeLine:   "Creature — Soldier",
		OracleText: "Creatures you control get +1/+1.",
		Power:      new("2"),
		Toughness:  new("2"),
	})
	body := face.StaticAbilities[0].Body
	ce := body.ContinuousEffects[0]
	if ce.Group.Domain() != game.GroupDomainObjectControlled {
		t.Fatalf("group domain = %v, want GroupDomainObjectControlled", ce.Group.Domain())
	}
	selection := ce.Group.Selection()
	if len(selection.RequiredTypes) != 1 || selection.RequiredTypes[0] != types.Creature {
		t.Fatalf("selection = %+v, want creature requirement", selection)
	}
	if _, excluded := ce.Group.Exclusion(); excluded {
		t.Fatal("group exclusion unexpectedly set")
	}
	if ce.PowerDelta != 1 || ce.ToughnessDelta != 1 {
		t.Fatalf("PT delta = %d/%d, want 1/1", ce.PowerDelta, ce.ToughnessDelta)
	}
}

func TestLowerOtherCreaturesYouControlPTBuff(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Lord",
		Layout:     "normal",
		TypeLine:   "Creature — Soldier",
		OracleText: "Other creatures you control get +1/+0.",
		Power:      new("2"),
		Toughness:  new("2"),
	})
	body := face.StaticAbilities[0].Body
	ce := body.ContinuousEffects[0]
	if ce.Group.Domain() != game.GroupDomainObjectControlled {
		t.Fatalf("group domain = %v, want GroupDomainObjectControlled", ce.Group.Domain())
	}
	if _, excluded := ce.Group.Exclusion(); !excluded {
		t.Fatal("group exclusion missing")
	}
	if ce.PowerDelta != 1 || ce.ToughnessDelta != 0 {
		t.Fatalf("PT delta = %d/%d, want 1/0", ce.PowerDelta, ce.ToughnessDelta)
	}
}

func TestLowerTapManaAbilityFixedColor(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Land",
		Layout:     "normal",
		TypeLine:   "Land",
		OracleText: "{T}: Add {G}.",
	})
	if len(face.ManaAbilities) != 1 {
		t.Fatalf("got %d mana abilities, want 1", len(face.ManaAbilities))
	}
	mode := face.ManaAbilities[0].Content.Modes[0]
	if len(mode.Sequence) != 1 {
		t.Fatalf("got %d instructions, want 1", len(mode.Sequence))
	}
	addMana, ok := mode.Sequence[0].Primitive.(game.AddMana)
	if !ok {
		t.Fatalf("primitive = %T, want game.AddMana", mode.Sequence[0].Primitive)
	}
	if addMana.ManaColor != mana.G {
		t.Fatalf("mana color = %q, want mana.G", addMana.ManaColor)
	}
}

func TestLowerTapManaAbilityChoice(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Land",
		Layout:     "normal",
		TypeLine:   "Land",
		OracleText: "{T}: Add {R} or {G}.",
	})
	if len(face.ManaAbilities) != 1 {
		t.Fatalf("got %d mana abilities, want 1", len(face.ManaAbilities))
	}
	mode := face.ManaAbilities[0].Content.Modes[0]
	choose, ok := mode.Sequence[0].Primitive.(game.Choose)
	if !ok {
		t.Fatalf("primitive = %T, want game.Choose", mode.Sequence[0].Primitive)
	}
	if choose.Choice.Kind != game.ResolutionChoiceMana {
		t.Fatalf("choice kind = %v, want ResolutionChoiceMana", choose.Choice.Kind)
	}
	if len(choose.Choice.Colors) != 2 {
		t.Fatalf("choice colors = %#v, want two colors", choose.Choice.Colors)
	}
}

// TestLowerManaAbilityMultiSymbolOutput verifies that "{T}: Add {G}{W}." is
// lowered to a mana ability with two sequential AddMana instructions, one for
// each mana symbol. This is the single-tap / two-color-output shape shared by
// dual-color tap lands (e.g. Sungrass Prairie).
func TestLowerManaAbilityMultiSymbolOutput(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Land",
		Layout:     "normal",
		TypeLine:   "Land",
		OracleText: "{T}: Add {G}{W}.",
	})
	if len(face.ManaAbilities) != 1 {
		t.Fatalf("got %d mana abilities, want 1", len(face.ManaAbilities))
	}
	ab := face.ManaAbilities[0]
	if ab.ManaCost.Exists {
		t.Fatalf("ManaCost = %v, want none", ab.ManaCost)
	}
	if len(ab.AdditionalCosts) != 1 || ab.AdditionalCosts[0].Kind != cost.AdditionalTap {
		t.Fatalf("AdditionalCosts = %#v, want [tap]", ab.AdditionalCosts)
	}
	mode := ab.Content.Modes[0]
	if len(mode.Sequence) != 2 {
		t.Fatalf("sequence length = %d, want 2", len(mode.Sequence))
	}
	first, ok := mode.Sequence[0].Primitive.(game.AddMana)
	if !ok {
		t.Fatalf("sequence[0] = %T, want game.AddMana", mode.Sequence[0].Primitive)
	}
	second, ok := mode.Sequence[1].Primitive.(game.AddMana)
	if !ok {
		t.Fatalf("sequence[1] = %T, want game.AddMana", mode.Sequence[1].Primitive)
	}
	if first.ManaColor != mana.G {
		t.Fatalf("first mana color = %q, want G", first.ManaColor)
	}
	if second.ManaColor != mana.W {
		t.Fatalf("second mana color = %q, want W", second.ManaColor)
	}
}

// TestLowerManaAbilityManaCostAndTap verifies that "{1}, {T}: Add {G}{W}." is
// lowered to a mana ability with ManaCost {1} and AdditionalCosts [tap], plus
// two sequential AddMana instructions. This is the Signet / mana-cost-tap-dual
// shape (e.g. Selesnya Signet, Sungrass Prairie variant).
func TestLowerManaAbilityManaCostAndTap(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Signet",
		Layout:     "normal",
		TypeLine:   "Artifact",
		OracleText: "{1}, {T}: Add {G}{W}.",
	})
	if len(face.ManaAbilities) != 1 {
		t.Fatalf("got %d mana abilities, want 1", len(face.ManaAbilities))
	}
	ab := face.ManaAbilities[0]
	if !ab.ManaCost.Exists {
		t.Fatal("ManaCost missing, want {1}")
	}
	if len(ab.ManaCost.Val) != 1 {
		t.Fatalf("ManaCost symbols = %d, want 1", len(ab.ManaCost.Val))
	}
	if len(ab.AdditionalCosts) != 1 || ab.AdditionalCosts[0].Kind != cost.AdditionalTap {
		t.Fatalf("AdditionalCosts = %#v, want [tap]", ab.AdditionalCosts)
	}
	mode := ab.Content.Modes[0]
	if len(mode.Sequence) != 2 {
		t.Fatalf("sequence length = %d, want 2", len(mode.Sequence))
	}
	first, ok := mode.Sequence[0].Primitive.(game.AddMana)
	if !ok || first.ManaColor != mana.G {
		t.Fatalf("first AddMana = %v, want G", mode.Sequence[0].Primitive)
	}
	second, ok := mode.Sequence[1].Primitive.(game.AddMana)
	if !ok || second.ManaColor != mana.W {
		t.Fatalf("second AddMana = %v, want W", mode.Sequence[1].Primitive)
	}
}

// TestLowerManaAbilityTapPayLife verifies that "{T}, Pay 1 life: Add {U} or {R}."
// is lowered with a tap additional cost, a pay-life additional cost, and a
// two-color mana choice. This is the pain-land / filter-land shape.
func TestLowerManaAbilityTapPayLife(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Pain Land",
		Layout:     "normal",
		TypeLine:   "Land",
		OracleText: "{T}, Pay 1 life: Add {U} or {R}.",
	})
	if len(face.ManaAbilities) != 1 {
		t.Fatalf("got %d mana abilities, want 1", len(face.ManaAbilities))
	}
	ab := face.ManaAbilities[0]
	if ab.ManaCost.Exists {
		t.Fatalf("ManaCost = %v, want none", ab.ManaCost)
	}
	if len(ab.AdditionalCosts) != 2 {
		t.Fatalf("AdditionalCosts count = %d, want 2", len(ab.AdditionalCosts))
	}
	if ab.AdditionalCosts[0].Kind != cost.AdditionalTap {
		t.Fatalf("AdditionalCosts[0].Kind = %v, want AdditionalTap", ab.AdditionalCosts[0].Kind)
	}
	if ab.AdditionalCosts[1].Kind != cost.AdditionalPayLife || ab.AdditionalCosts[1].Amount != 1 {
		t.Fatalf("AdditionalCosts[1] = %#v, want AdditionalPayLife amount=1", ab.AdditionalCosts[1])
	}
	mode := ab.Content.Modes[0]
	choose, ok := mode.Sequence[0].Primitive.(game.Choose)
	if !ok || choose.Choice.Kind != game.ResolutionChoiceMana || len(choose.Choice.Colors) != 2 {
		t.Fatalf("sequence[0] = %v, want mana choice of 2 colors", mode.Sequence[0].Primitive)
	}
}

// TestLowerManaAbilitySacrificeSelf verifies that "Sacrifice this creature: Add {C}."
// is lowered with an AdditionalSacrificeSource cost and a fixed colorless mana output.
func TestLowerManaAbilitySacrificeSelf(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Scion",
		Layout:     "normal",
		TypeLine:   "Creature — Eldrazi Scion",
		OracleText: "Sacrifice this creature: Add {C}.",
		Power:      new("1"),
		Toughness:  new("1"),
	})
	if len(face.ManaAbilities) != 1 {
		t.Fatalf("got %d mana abilities, want 1", len(face.ManaAbilities))
	}
	ab := face.ManaAbilities[0]
	if ab.ManaCost.Exists {
		t.Fatalf("ManaCost = %v, want none", ab.ManaCost)
	}
	if len(ab.AdditionalCosts) != 1 || ab.AdditionalCosts[0].Kind != cost.AdditionalSacrificeSource {
		t.Fatalf("AdditionalCosts = %#v, want [sacrifice source]", ab.AdditionalCosts)
	}
	mode := ab.Content.Modes[0]
	if len(mode.Sequence) != 1 {
		t.Fatalf("sequence length = %d, want 1", len(mode.Sequence))
	}
	addMana, ok := mode.Sequence[0].Primitive.(game.AddMana)
	if !ok || addMana.ManaColor != mana.C {
		t.Fatalf("sequence[0] = %v, want AddMana{C}", mode.Sequence[0].Primitive)
	}
}

// TestLowerManaAbilityPureManaAnyColor verifies that "{G}: Add one mana of any
// color." is lowered with a mana cost {G}, no additional costs, and a five-color
// choice output. This is the Orochi Leafcaller / Nomadic Elf shape.
func TestLowerManaAbilityPureManaAnyColor(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Leafcaller",
		Layout:     "normal",
		TypeLine:   "Creature — Snake Shaman",
		OracleText: "{G}: Add one mana of any color.",
		Power:      new("1"),
		Toughness:  new("1"),
	})
	if len(face.ManaAbilities) != 1 {
		t.Fatalf("got %d mana abilities, want 1", len(face.ManaAbilities))
	}
	ab := face.ManaAbilities[0]
	if !ab.ManaCost.Exists {
		t.Fatal("ManaCost missing, want {G}")
	}
	if len(ab.AdditionalCosts) != 0 {
		t.Fatalf("AdditionalCosts = %#v, want empty", ab.AdditionalCosts)
	}
	mode := ab.Content.Modes[0]
	choose, ok := mode.Sequence[0].Primitive.(game.Choose)
	if !ok || choose.Choice.Kind != game.ResolutionChoiceMana || len(choose.Choice.Colors) != 5 {
		t.Fatalf("sequence[0] = %v, want any-color mana choice", mode.Sequence[0].Primitive)
	}
}

// TestLowerManaAbilityPureManaFixed verifies that "{R}: Add {B}." is lowered
// with a mana cost {R}, no additional costs, and a single AddMana{B} instruction.
// This is the Agent of Stromgald / mana-conversion shape.
func TestLowerManaAbilityPureManaFixed(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Agent",
		Layout:     "normal",
		TypeLine:   "Creature — Human Cleric",
		OracleText: "{R}: Add {B}.",
		Power:      new("1"),
		Toughness:  new("1"),
	})
	if len(face.ManaAbilities) != 1 {
		t.Fatalf("got %d mana abilities, want 1", len(face.ManaAbilities))
	}
	ab := face.ManaAbilities[0]
	if !ab.ManaCost.Exists {
		t.Fatal("ManaCost missing, want {R}")
	}
	if len(ab.AdditionalCosts) != 0 {
		t.Fatalf("AdditionalCosts = %#v, want empty", ab.AdditionalCosts)
	}
	mode := ab.Content.Modes[0]
	addMana, ok := mode.Sequence[0].Primitive.(game.AddMana)
	if !ok || addMana.ManaColor != mana.B {
		t.Fatalf("sequence[0] = %v, want AddMana{B}", mode.Sequence[0].Primitive)
	}
}

// TestLowerManaAbilityDiscardCost verifies that "Discard a card: Add {B}." is
// lowered with an AdditionalDiscard cost and a single AddMana{B} output.
// This is the Skirge Familiar family shape (mana ability with discard cost).
func TestLowerManaAbilityDiscardCost(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Skirge",
		Layout:     "normal",
		TypeLine:   "Creature — Imp",
		OracleText: "Discard a card: Add {B}.",
		Power:      new("3"),
		Toughness:  new("1"),
	})
	if len(face.ManaAbilities) != 1 {
		t.Fatalf("got %d mana abilities, want 1", len(face.ManaAbilities))
	}
	ab := face.ManaAbilities[0]
	if ab.ManaCost.Exists {
		t.Fatalf("ManaCost = %v, want none", ab.ManaCost)
	}
	if len(ab.AdditionalCosts) != 1 || ab.AdditionalCosts[0].Kind != cost.AdditionalDiscard {
		t.Fatalf("AdditionalCosts = %#v, want [discard]", ab.AdditionalCosts)
	}
	if ab.AdditionalCosts[0].Amount != 1 {
		t.Fatalf("discard amount = %d, want 1", ab.AdditionalCosts[0].Amount)
	}
	mode := ab.Content.Modes[0]
	addMana, ok := mode.Sequence[0].Primitive.(game.AddMana)
	if !ok || addMana.ManaColor != mana.B {
		t.Fatalf("sequence[0] = %v, want AddMana{B}", mode.Sequence[0].Primitive)
	}
}

// TestLowerManaAbilityTypedSacrifice verifies that "Sacrifice a creature: Add {C}{C}."
// is lowered with an AdditionalSacrifice cost targeting creatures and a two-instruction
// colorless mana output. This is the Ashnod's Altar shape.
func TestLowerManaAbilityTypedSacrifice(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Altar",
		Layout:     "normal",
		TypeLine:   "Artifact",
		OracleText: "Sacrifice a creature: Add {C}{C}.",
	})
	if len(face.ManaAbilities) != 1 {
		t.Fatalf("got %d mana abilities, want 1", len(face.ManaAbilities))
	}
	ab := face.ManaAbilities[0]
	if ab.ManaCost.Exists {
		t.Fatalf("ManaCost = %v, want none", ab.ManaCost)
	}
	if len(ab.AdditionalCosts) != 1 {
		t.Fatalf("AdditionalCosts count = %d, want 1", len(ab.AdditionalCosts))
	}
	sacCost := ab.AdditionalCosts[0]
	if sacCost.Kind != cost.AdditionalSacrifice || sacCost.Amount != 1 ||
		!sacCost.MatchPermanentType || sacCost.PermanentType != types.Creature {
		t.Fatalf("AdditionalCosts[0] = %#v, want sacrifice-a-creature", sacCost)
	}
	mode := ab.Content.Modes[0]
	if len(mode.Sequence) != 2 {
		t.Fatalf("sequence length = %d, want 2", len(mode.Sequence))
	}
	for i, instr := range mode.Sequence {
		addMana, ok := instr.Primitive.(game.AddMana)
		if !ok || addMana.ManaColor != mana.C {
			t.Fatalf("sequence[%d] = %v, want AddMana{C}", i, instr.Primitive)
		}
	}
}

// TestLowerManaAbilityRejectsComplexBody verifies that mana abilities with body
// patterns outside the three supported shapes (fixed, choice, any-color) are
// rejected. "Three mana in any combination" requires Amount > 1 with a
// repeated-choice mechanism that is not yet supported.
func TestLowerManaAbilityRejectsComplexBody(t *testing.T) {
	t.Parallel()
	_, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Test Goblin",
		Layout:     "normal",
		TypeLine:   "Creature — Goblin",
		OracleText: "{T}, Sacrifice a Forest: Add three mana in any combination of {R} and/or {G}.",
		Power:      new("2"),
		Toughness:  new("2"),
	}, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) == 0 {
		t.Fatal("expected unsupported diagnostic for complex mana body")
	}
}
