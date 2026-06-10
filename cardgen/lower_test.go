package cardgen

import (
	"go/parser"
	"go/token"
	"slices"
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

func lowerSingleFace(t *testing.T, card *ScryfallCard) loweredFaceAbilities {
	t.Helper()
	faces, diagnostics := lowerExecutableFaces(card)
	if len(diagnostics) != 0 {
		t.Fatalf("unexpected diagnostics: %#v", diagnostics)
	}
	if len(faces) == 0 {
		t.Fatal("no faces lowered")
	}
	return faces[0]
}

func TestLowerKeywordAbilityStaticBodies(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Bear",
		Layout:     "normal",
		TypeLine:   "Creature — Bear",
		OracleText: "Flying\nVigilance",
		Power:      new("2"),
		Toughness:  new("2"),
	})
	if len(face.StaticAbilities) != 2 {
		t.Fatalf("got %d static abilities, want 2", len(face.StaticAbilities))
	}
	if got := face.StaticAbilities[0].VarName; got != "game.FlyingStaticBody" {
		t.Fatalf("first static VarName = %q", got)
	}
	if got := face.StaticAbilities[1].VarName; got != "game.VigilanceStaticBody" {
		t.Fatalf("second static VarName = %q", got)
	}
}

func TestLowerKeywordAbilityWard(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Bear",
		Layout:     "normal",
		TypeLine:   "Creature — Bear",
		OracleText: "Ward {2}",
		Power:      new("2"),
		Toughness:  new("2"),
	})
	if len(face.StaticAbilities) != 1 {
		t.Fatalf("got %d static abilities, want 1", len(face.StaticAbilities))
	}
	body := face.StaticAbilities[0].Body
	if len(body.KeywordAbilities) != 1 {
		t.Fatalf("got %d keyword abilities, want 1", len(body.KeywordAbilities))
	}
	ward, ok := body.KeywordAbilities[0].(game.WardKeyword)
	if !ok {
		t.Fatalf("keyword ability = %T, want game.WardKeyword", body.KeywordAbilities[0])
	}
	if len(ward.Cost) != 1 || ward.Cost[0].Kind != cost.GenericSymbol || ward.Cost[0].Generic != 2 {
		t.Fatalf("ward cost = %#v, want {2}", ward.Cost)
	}
}

func TestLowerCyclingAbility(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Card",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Cycling {1}{U} ({1}{U}, Discard this card: Draw a card.)",
	})
	if len(face.ActivatedAbilities) != 1 {
		t.Fatalf("got %d activated abilities, want 1", len(face.ActivatedAbilities))
	}
	ability := face.ActivatedAbilities[0]
	if !ability.ManaCost.Exists {
		t.Fatal("cycling ability has no mana cost")
	}
	if len(ability.AdditionalCosts) != 1 || ability.AdditionalCosts[0].Kind != cost.AdditionalDiscard {
		t.Fatalf("additional costs = %#v, want one discard", ability.AdditionalCosts)
	}
	if len(ability.KeywordAbilities) != 1 {
		t.Fatalf("got %d keyword abilities, want 1", len(ability.KeywordAbilities))
	}
	if _, ok := ability.KeywordAbilities[0].(game.CyclingKeyword); !ok {
		t.Fatalf("keyword ability = %T, want game.CyclingKeyword", ability.KeywordAbilities[0])
	}
}

func TestLowerActivatedNonManaCosts(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		oracleText string
		check      func(*testing.T, []cost.Additional)
	}{
		{
			name:       "sacrifice source",
			oracleText: "Sacrifice this artifact: Draw a card.",
			check: func(t *testing.T, costs []cost.Additional) {
				t.Helper()
				if len(costs) != 1 || costs[0].Kind != cost.AdditionalSacrificeSource {
					t.Fatalf("additional costs = %#v, want source sacrifice", costs)
				}
			},
		},
		{
			name:       "typed sacrifice after mana and tap",
			oracleText: "{2}, {T}, Sacrifice a creature: Draw a card.",
			check: func(t *testing.T, costs []cost.Additional) {
				t.Helper()
				if len(costs) != 2 ||
					costs[0].Kind != cost.AdditionalTap ||
					costs[1].Kind != cost.AdditionalSacrifice ||
					!costs[1].MatchPermanentType ||
					costs[1].PermanentType != types.Creature {
					t.Fatalf("additional costs = %#v, want tap and creature sacrifice", costs)
				}
			},
		},
		{
			name:       "typed discard",
			oracleText: "Discard two creature cards: Draw a card.",
			check: func(t *testing.T, costs []cost.Additional) {
				t.Helper()
				if len(costs) != 1 ||
					costs[0].Kind != cost.AdditionalDiscard ||
					costs[0].Amount != 2 ||
					!costs[0].MatchCardType ||
					costs[0].CardType != types.Creature ||
					costs[0].Source != zone.Hand {
					t.Fatalf("additional costs = %#v, want two creature cards discarded", costs)
				}
			},
		},
		{
			name:       "pay life",
			oracleText: "Pay 2 life: Draw a card.",
			check: func(t *testing.T, costs []cost.Additional) {
				t.Helper()
				if len(costs) != 1 || costs[0].Kind != cost.AdditionalPayLife || costs[0].Amount != 2 {
					t.Fatalf("additional costs = %#v, want 2 life", costs)
				}
			},
		},
		{
			name:       "exile source",
			oracleText: "Exile this artifact: Draw a card.",
			check: func(t *testing.T, costs []cost.Additional) {
				t.Helper()
				if len(costs) != 1 ||
					costs[0].Kind != cost.AdditionalExileSource ||
					costs[0].Source != zone.Battlefield {
					t.Fatalf("additional costs = %#v, want battlefield source exile", costs)
				}
			},
		},
		{
			name:       "exile graveyard card",
			oracleText: "Exile a card from your graveyard: Draw a card.",
			check: func(t *testing.T, costs []cost.Additional) {
				t.Helper()
				if len(costs) != 1 ||
					costs[0].Kind != cost.AdditionalExile ||
					costs[0].Amount != 1 ||
					costs[0].Source != zone.Graveyard ||
					costs[0].MatchCardType {
					t.Fatalf("additional costs = %#v, want one graveyard card exile", costs)
				}
			},
		},
		{
			name:       "exile typed graveyard card",
			oracleText: "Exile a creature card from your graveyard: Draw a card.",
			check: func(t *testing.T, costs []cost.Additional) {
				t.Helper()
				if len(costs) != 1 ||
					costs[0].Kind != cost.AdditionalExile ||
					costs[0].Amount != 1 ||
					costs[0].Source != zone.Graveyard ||
					!costs[0].MatchCardType ||
					costs[0].CardType != types.Creature {
					t.Fatalf("additional costs = %#v, want one graveyard creature card exile", costs)
				}
			},
		},
		{
			name:       "exile two graveyard cards",
			oracleText: "Exile two cards from your graveyard: Draw a card.",
			check: func(t *testing.T, costs []cost.Additional) {
				t.Helper()
				if len(costs) != 1 ||
					costs[0].Kind != cost.AdditionalExile ||
					costs[0].Amount != 2 ||
					costs[0].Source != zone.Graveyard {
					t.Fatalf("additional costs = %#v, want two graveyard card exiles", costs)
				}
			},
		},
		{
			name:       "untap source",
			oracleText: "{Q}: Draw a card.",
			check: func(t *testing.T, costs []cost.Additional) {
				t.Helper()
				if len(costs) != 1 ||
					costs[0].Kind != cost.AdditionalUntap ||
					costs[0].Text != "{Q}" {
					t.Fatalf("additional costs = %#v, want untap source", costs)
				}
			},
		},
		{
			name:       "remove source counter",
			oracleText: "Remove a +1/+1 counter from this artifact: Draw a card.",
			check: func(t *testing.T, costs []cost.Additional) {
				t.Helper()
				if len(costs) != 1 ||
					costs[0].Kind != cost.AdditionalRemoveCounter ||
					costs[0].Amount != 1 ||
					costs[0].CounterKind != counter.PlusOnePlusOne {
					t.Fatalf("additional costs = %#v, want source +1/+1 counter removal", costs)
				}
			},
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
			test.check(t, face.ActivatedAbilities[0].AdditionalCosts)
		})
	}
}

func TestLowerActivatedAbilityRejectsAmbiguousExileCost(t *testing.T) {
	t.Parallel()
	faces, diagnostics := lowerExecutableFaces(&ScryfallCard{
		Name:       "Test Engine",
		Layout:     "normal",
		TypeLine:   "Artifact",
		OracleText: "Exile a card: Draw a card.",
	})
	if len(faces) != 1 || len(faces[0].ActivatedAbilities) != 0 {
		t.Fatalf("faces = %#v, want face with no partially lowered ability", faces)
	}
	if len(diagnostics) == 0 {
		t.Fatal("expected unsupported diagnostic")
	}
}

func TestLowerActivatedAbilityRejectsCounterRemovalFromTarget(t *testing.T) {
	t.Parallel()
	faces, diagnostics := lowerExecutableFaces(&ScryfallCard{
		Name:       "Test Engine",
		Layout:     "normal",
		TypeLine:   "Artifact",
		OracleText: "Remove a +1/+1 counter from target creature: Draw a card.",
	})
	if len(faces) != 1 || len(faces[0].ActivatedAbilities) != 0 {
		t.Fatalf("faces = %#v, want face with no partially lowered ability", faces)
	}
	if len(diagnostics) == 0 {
		t.Fatal("expected unsupported diagnostic")
	}
}

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
		{
			"sorcery once per turn",
			"{1}: Draw a card. Activate only as a sorcery. Activate only once each turn.",
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
		len(target.Predicate.PermanentTypes) != 1 ||
		target.Predicate.PermanentTypes[0] != types.Creature {
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

func TestLowerEntersTappedReplacement(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Land",
		Layout:     "normal",
		TypeLine:   "Land",
		OracleText: "This land enters tapped.",
	})
	if len(face.ReplacementAbilities) != 1 {
		t.Fatalf("got %d replacement abilities, want 1", len(face.ReplacementAbilities))
	}
	if !face.ReplacementAbilities[0].Replacement.EntersTapped {
		t.Fatal("replacement is not an enters-tapped replacement")
	}
}

func TestGenerateEquippedCreaturePTBuff(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Test Equipment",
		Layout:     "normal",
		TypeLine:   "Artifact — Equipment",
		OracleText: "Equipped creature gets +2/+0.\nEquip {2}",
	}, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("unexpected diagnostics: %#v", diagnostics)
	}
	if !strings.Contains(source, "LayerPowerToughnessModify") {
		t.Fatalf("source does not contain static PT effect:\n%s", source)
	}
	if !strings.Contains(source, "AttachedObjectGroup") {
		t.Fatalf("source does not contain AttachedObjectGroup:\n%s", source)
	}
	if _, err := parser.ParseFile(token.NewFileSet(), "generated.go", source, parser.AllErrors); err != nil {
		t.Fatalf("generated source does not parse: %v\n%s", err, source)
	}
}

func TestGenerateEquippedCreaturePTBuffWithKeywords(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Test Equipment",
		Layout:     "normal",
		TypeLine:   "Artifact — Equipment",
		OracleText: "Equipped creature gets +2/+2 and has trample and lifelink.\nEquip {3}",
	}, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("unexpected diagnostics: %#v", diagnostics)
	}
	for _, wanted := range []string{
		"game.LayerPowerToughnessModify",
		"game.LayerAbility",
		"AddKeywords: []game.Keyword",
		"game.Trample",
		"game.Lifelink",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
}

func TestGenerateControlledCreaturesPTBuffWithKeyword(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Test Anthem",
		Layout:     "normal",
		TypeLine:   "Enchantment",
		OracleText: "Creatures you control get +1/+1 and have vigilance.",
	}, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("unexpected diagnostics: %#v", diagnostics)
	}
	if !strings.Contains(source, "game.Vigilance") {
		t.Fatalf("source missing vigilance:\n%s", source)
	}
}

func TestRejectStaticPTBuffWithUnsupportedKeywordText(t *testing.T) {
	t.Parallel()
	for _, oracleText := range []string{
		"Equipped creature gets +2/+2 and has trample or lifelink.\nEquip {3}",
		"Equipped creature gets +2/+2 and has and trample.\nEquip {3}",
		"Equipped creature gets +2/+2 and has trample and.\nEquip {3}",
		"Equipped creature gets +2/+2 and has flying lifelink.\nEquip {3}",
		"Equipped creature gets +2/+2 and has shroud.\nEquip {3}",
		"Equipped creature gets +2/+2 and has wither.\nEquip {3}",
	} {
		source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
			Name:       "Test Equipment",
			Layout:     "normal",
			TypeLine:   "Artifact — Equipment",
			OracleText: oracleText,
		}, "t")
		if err != nil {
			t.Fatal(err)
		}
		if source != "" {
			t.Fatalf("unexpected source for %q:\n%s", oracleText, source)
		}
		if len(diagnostics) == 0 {
			t.Fatalf("expected unsupported diagnostic for %q", oracleText)
		}
	}
}

func TestRejectResolvingPTBuffAsStatic(t *testing.T) {
	t.Parallel()
	_, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Test Equipment",
		Layout:     "normal",
		TypeLine:   "Artifact — Equipment",
		OracleText: "Equipped creature gets +2/+0 until end of turn.\nEquip {2}",
	}, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) == 0 {
		t.Fatal("expected rejection of resolving P/T effect, got none")
	}
}

func TestRejectVariablePTBuff(t *testing.T) {
	t.Parallel()
	_, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Test Equipment",
		Layout:     "normal",
		TypeLine:   "Artifact — Equipment",
		OracleText: "Equipped creature gets +1/+0 for each Equipment attached to it.\nEquip {2}",
	}, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) == 0 {
		t.Fatal("expected rejection of variable-amount P/T buff, got none")
	}
}

func TestGenerateExtendedStaticPTBuffSubjects(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		oracleText string
		want       string
	}{
		"walls": {
			oracleText: "Each Wall you control gets +0/+2.",
			want:       `SubtypesAny: []types.Sub{types.Sub("Wall")}`,
		},
		"artifacts": {
			oracleText: "Artifacts you control get +1/+1.",
			want:       "RequiredTypes: []types.Card{types.Artifact}",
		},
		"tokens": {
			oracleText: "Tokens you control get +1/+1.",
			want:       "TokenOnly: true",
		},
		"opponents' creatures": {
			oracleText: "Creatures your opponents control get -1/-0.",
			want:       "Controller: game.ControllerOpponent",
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
				Name:       "Test Anthem",
				Layout:     "normal",
				TypeLine:   "Enchantment",
				OracleText: test.oracleText,
			}, "t")
			if err != nil {
				t.Fatal(err)
			}
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			if !strings.Contains(source, test.want) {
				t.Fatalf("source missing %q:\n%s", test.want, source)
			}
		})
	}
}

func TestLowerConditionalEntersTappedReplacement(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Vista",
		Layout:     "normal",
		TypeLine:   "Land — Forest Plains",
		OracleText: "This land enters tapped unless you control two or more basic lands.",
	})
	if len(face.ReplacementAbilities) != 1 {
		t.Fatalf("got %d replacement abilities, want 1", len(face.ReplacementAbilities))
	}
	repl := face.ReplacementAbilities[0]
	if !repl.Replacement.EntersTapped {
		t.Fatal("replacement is not an enters-tapped replacement")
	}
	if !repl.Replacement.Condition.Exists {
		t.Fatal("conditional replacement has no condition")
	}
	cond := repl.Replacement.Condition.Val
	if !cond.Negate {
		t.Fatal("condition should be negated (unless)")
	}
	filter := cond.ControllerControls
	if len(filter.Types) != 1 || filter.Types[0] != types.Land {
		t.Fatalf("filter types = %#v, want [types.Land]", filter.Types)
	}
	if len(filter.Supertypes) != 1 || filter.Supertypes[0] != types.Basic {
		t.Fatalf("filter supertypes = %#v, want [types.Basic]", filter.Supertypes)
	}
	if filter.MinCount != 2 {
		t.Fatalf("filter MinCount = %d, want 2", filter.MinCount)
	}
}

func TestLowerCommonConditionalEntersTappedReplacements(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name          string
		oracleText    string
		negate        bool
		minCount      int
		excludeSource bool
		subtypes      []types.Sub
	}{
		{
			name:          "two or more other lands",
			oracleText:    "This land enters tapped unless you control two or more other lands.",
			negate:        true,
			minCount:      2,
			excludeSource: true,
		},
		{
			name:          "two or fewer other lands",
			oracleText:    "This land enters tapped unless you control two or fewer other lands.",
			minCount:      3,
			excludeSource: true,
		},
		{
			name:       "basic land subtype pair",
			oracleText: "This land enters tapped unless you control a Plains or an Island.",
			subtypes:   []types.Sub{types.Plains, types.Island},
			negate:     true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			face := lowerSingleFace(t, &ScryfallCard{
				Name:       "Test Land",
				Layout:     "normal",
				TypeLine:   "Land",
				OracleText: test.oracleText,
			})
			condition := face.ReplacementAbilities[0].Replacement.Condition.Val
			filter := condition.ControllerControls
			if condition.Negate != test.negate ||
				filter.MinCount != test.minCount ||
				filter.ExcludeSource != test.excludeSource ||
				!slices.Equal(filter.SubtypesAny, test.subtypes) {
				t.Fatalf("condition = %+v, want negate=%v min=%d exclude=%v subtypes=%v",
					condition, test.negate, test.minCount, test.excludeSource, test.subtypes)
			}
		})
	}
}

func TestLowerLifeAndOpponentConditionalEntersTappedReplacements(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		condition string
		assert    func(*testing.T, game.Condition)
	}{
		{
			name:      "controller life",
			condition: "unless you have 10 or more life",
			assert: func(t *testing.T, condition game.Condition) {
				if condition.ControllerLifeAtLeast != 10 {
					t.Fatalf("ControllerLifeAtLeast = %d, want 10", condition.ControllerLifeAtLeast)
				}
			},
		},
		{
			name:      "any player life",
			condition: "unless a player has 13 or less life",
			assert: func(t *testing.T, condition game.Condition) {
				if condition.AnyPlayerLifeAtMost != 13 {
					t.Fatalf("AnyPlayerLifeAtMost = %d, want 13", condition.AnyPlayerLifeAtMost)
				}
			},
		},
		{
			name:      "opponent count",
			condition: "unless you have two or more opponents",
			assert: func(t *testing.T, condition game.Condition) {
				if condition.OpponentCountAtLeast != 2 {
					t.Fatalf("OpponentCountAtLeast = %d, want 2", condition.OpponentCountAtLeast)
				}
			},
		},
		{
			name:      "one opponent land count",
			condition: "unless an opponent controls two or more lands",
			assert: func(t *testing.T, condition game.Condition) {
				if !condition.AnyOpponentControls.Exists ||
					condition.AnyOpponentControls.Val.MinCount != 2 {
					t.Fatalf("AnyOpponentControls = %+v, want two lands", condition.AnyOpponentControls)
				}
			},
		},
		{
			name:      "collective opponent land count",
			condition: "unless your opponents control eight or more lands",
			assert: func(t *testing.T, condition game.Condition) {
				if !condition.OpponentsControl.Exists ||
					condition.OpponentsControl.Val.MinCount != 8 {
					t.Fatalf("OpponentsControl = %+v, want eight lands", condition.OpponentsControl)
				}
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			face := lowerSingleFace(t, &ScryfallCard{
				Name:       "Test Land",
				Layout:     "normal",
				TypeLine:   "Land",
				OracleText: "This land enters tapped " + test.condition + ".",
			})
			condition := face.ReplacementAbilities[0].Replacement.Condition.Val
			if !condition.Negate {
				t.Fatal("unless condition was not negated")
			}
			test.assert(t, condition)
		})
	}
}

func TestLowerReminderManaAbilitySingleColor(t *testing.T) {
	t.Parallel()
	// Basic lands express their mana ability as a parenthesized reminder.
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Forest",
		Layout:     "normal",
		TypeLine:   "Basic Land — Forest",
		OracleText: "({T}: Add {G}.)",
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

func TestLowerReminderManaAbilityChoice(t *testing.T) {
	t.Parallel()
	// Dual lands express their mana ability as a parenthesized reminder.
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Dual",
		Layout:     "normal",
		TypeLine:   "Land — Mountain Forest",
		OracleText: "({T}: Add {R} or {G}.)",
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

func TestLowerReminderManaAbilityRejectsHybridCost(t *testing.T) {
	t.Parallel()
	// Hybrid-mana cost reminders are not mana abilities; they must be rejected.
	_, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Test Hybrid",
		Layout:     "normal",
		TypeLine:   "Creature — Soldier",
		OracleText: "({R/W} can be paid with either {R} or {W}.)\nFirst strike",
		Power:      new("1"),
		Toughness:  new("1"),
	}, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) == 0 {
		t.Fatal("expected diagnostics for hybrid-cost reminder, got none")
	}
}

func TestLowerReminderManaAbilityRejectsNonMana(t *testing.T) {
	t.Parallel()
	// A parenthesized reminder that is not a mana ability must be rejected.
	_, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Test Card",
		Layout:     "normal",
		TypeLine:   "Creature — Human",
		OracleText: "(This creature can block as though it had flying.)\nFlying",
		Power:      new("1"),
		Toughness:  new("1"),
	}, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) == 0 {
		t.Fatal("expected diagnostics for non-mana reminder, got none")
	}
}

func TestLowerEnterTrigger(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Bear",
		Layout:     "normal",
		TypeLine:   "Creature — Bear",
		OracleText: "When this creature enters, draw a card.",
		Power:      new("2"),
		Toughness:  new("2"),
	})
	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("got %d triggered abilities, want 1", len(face.TriggeredAbilities))
	}

	trigger := face.TriggeredAbilities[0].Trigger
	if trigger.Pattern.Event != game.EventPermanentEnteredBattlefield {
		t.Fatalf("event = %v, want EventPermanentEnteredBattlefield", trigger.Pattern.Event)
	}
	if trigger.Pattern.Source != game.TriggerSourceSelf {
		t.Fatalf("source = %v, want TriggerSourceSelf", trigger.Pattern.Source)
	}
}

func TestLowerKickedEnterTrigger(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Kicker",
		Layout:     "normal",
		ManaCost:   "{2}{U}",
		TypeLine:   "Creature — Wizard",
		OracleText: "Kicker {1}{U}\nWhen this creature enters, if it was kicked, draw two cards.",
		Power:      new("2"),
		Toughness:  new("2"),
	})
	trigger := face.TriggeredAbilities[0].Trigger
	if trigger.InterveningIf != "if it was kicked" ||
		!trigger.InterveningIfEventPermanentWasKicked {
		t.Fatalf("trigger = %+v, want kicked intervening-if", trigger)
	}
	draw, ok := face.TriggeredAbilities[0].Content.Modes[0].Sequence[0].Primitive.(game.Draw)
	if !ok || draw.Amount != game.Fixed(2) {
		t.Fatalf("primitive = %+v, want draw two", face.TriggeredAbilities[0].Content.Modes[0].Sequence[0].Primitive)
	}
}

func TestLowerSagaChapterAbilities(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Saga",
		Layout:     "saga",
		TypeLine:   "Enchantment — Saga",
		OracleText: "I — Draw a card.\nII, III — Draw two cards.",
	})
	if len(face.ChapterAbilities) != 2 {
		t.Fatalf("got %d chapter abilities, want 2", len(face.ChapterAbilities))
	}
	if !slices.Equal(face.ChapterAbilities[0].Chapters, []int{1}) ||
		!slices.Equal(face.ChapterAbilities[1].Chapters, []int{2, 3}) {
		t.Fatalf("chapter numbers = %v, %v", face.ChapterAbilities[0].Chapters, face.ChapterAbilities[1].Chapters)
	}
	draw, ok := face.ChapterAbilities[1].Content.Modes[0].Sequence[0].Primitive.(game.Draw)
	if !ok {
		t.Fatalf("primitive = %T, want game.Draw", face.ChapterAbilities[1].Content.Modes[0].Sequence[0].Primitive)
	}
	if got := draw.Amount; got != game.Fixed(2) {
		t.Fatalf("draw amount = %#v, want 2", got)
	}
}

func TestLowerChapterShapedTextRequiresSagaSubtype(t *testing.T) {
	t.Parallel()
	_, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Not a Saga",
		Layout:     "normal",
		TypeLine:   "Enchantment",
		OracleText: "I — Draw a card.",
	}, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) == 0 {
		t.Fatal("expected non-Saga chapter-shaped text to be rejected")
	}
}

func TestOrdinarySagaReminder(t *testing.T) {
	t.Parallel()
	for _, text := range []string{
		"(As this Saga enters and after your draw step, add a lore counter.)",
		"(As this Saga enters and after your draw step, add a lore counter. Sacrifice after I.)",
		"(As this Saga enters and after your draw step add a lore counter. Sacrifice after III.)",
	} {
		if !isOrdinarySagaReminder(text) {
			t.Errorf("isOrdinarySagaReminder(%q) = false", text)
		}
	}
	for _, text := range []string{
		"Read ahead (Choose a chapter and start with that many lore counters.)",
		"(As this Saga enters and after your draw step, add a lore counter. Sacrifice after VII.)",
		"(As this Saga enters, add a lore counter.)",
	} {
		if isOrdinarySagaReminder(text) {
			t.Errorf("isOrdinarySagaReminder(%q) = true", text)
		}
	}
}

func TestLowerSagaChapterConsumesInlineReminderText(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Saga",
		Layout:     "saga",
		TypeLine:   "Enchantment — Saga",
		OracleText: "I — Proliferate. (Choose any number of permanents and/or players, then give each another counter of each kind already there.)",
	})
	if len(face.ChapterAbilities) != 1 {
		t.Fatalf("got %d chapter abilities, want 1", len(face.ChapterAbilities))
	}
}

func TestLowerDiesTrigger(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Bear",
		Layout:     "normal",
		TypeLine:   "Creature — Bear",
		OracleText: "When this creature dies, draw two cards.",
		Power:      new("2"),
		Toughness:  new("2"),
	})
	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("got %d triggered abilities, want 1", len(face.TriggeredAbilities))
	}
	if face.TriggeredAbilities[0].Trigger.Pattern.Event != game.EventPermanentDied {
		t.Fatalf("event = %v, want EventPermanentDied", face.TriggeredAbilities[0].Trigger.Pattern.Event)
	}
}

func TestLowerSelfDiesDamageTrigger(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Devil",
		Layout:     "normal",
		TypeLine:   "Creature — Devil",
		OracleText: "When this creature dies, it deals 3 damage to any target.",
		Power:      new("2"),
		Toughness:  new("2"),
	})
	mode := face.TriggeredAbilities[0].Content.Modes[0]
	if len(mode.Targets) != 1 {
		t.Fatalf("got %d targets, want 1", len(mode.Targets))
	}
	damage, ok := mode.Sequence[0].Primitive.(game.Damage)
	if !ok ||
		damage.Amount.Value() != 3 ||
		!damage.DamageSource.Exists ||
		damage.DamageSource.Val != game.EventPermanentReference() {
		t.Fatalf("primitive = %+v, want damage from event permanent", mode.Sequence[0].Primitive)
	}
}

func TestLowerManaParameterizedKeywords(t *testing.T) {
	t.Parallel()

	kicker := lowerKeywordForTest(t, "Kicker {1}{G}", game.Kicker)
	kickerKeyword, ok := kicker.(game.KickerKeyword)
	if !ok || kickerKeyword.Cost.String() != "{1}{G}" {
		t.Fatalf("Kicker keyword = %#v, want {1}{G}", kicker)
	}

	madness := lowerKeywordForTest(t, "Madness {2}{B}", game.Madness)
	madnessKeyword, ok := madness.(game.MadnessKeyword)
	if !ok || madnessKeyword.Cost.String() != "{2}{B}" {
		t.Fatalf("Madness keyword = %#v, want {2}{B}", madness)
	}

	morph := lowerKeywordForTest(t, "Morph {3}{U}", game.Morph)
	morphKeyword, ok := morph.(game.MorphKeyword)
	if !ok || morphKeyword.Cost.String() != "{3}{U}" {
		t.Fatalf("Morph keyword = %#v, want {3}{U}", morph)
	}

	disguise := lowerKeywordForTest(t, "Disguise {4}{W}", game.Disguise)
	disguiseKeyword, ok := disguise.(game.DisguiseKeyword)
	if !ok || disguiseKeyword.Cost.String() != "{4}{W}" {
		t.Fatalf("Disguise keyword = %#v, want {4}{W}", disguise)
	}
}

func TestLowerToxicKeyword(t *testing.T) {
	t.Parallel()
	keyword := lowerKeywordForTest(t, "Toxic 2", game.Toxic)
	toxic, ok := keyword.(game.ToxicKeyword)
	if !ok || toxic.Amount != 2 {
		t.Fatalf("Toxic keyword = %#v, want amount 2", keyword)
	}
}

func TestLowerParameterizedKeywordRejectsVariableCost(t *testing.T) {
	t.Parallel()
	_, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Variable Morph",
		Layout:     "normal",
		TypeLine:   "Creature — Test",
		OracleText: "Morph {X}{U}",
		Power:      new("2"),
		Toughness:  new("2"),
	}, "v")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) == 0 || diagnostics[0].Summary != "unsupported parameterized keyword" {
		t.Fatalf("diagnostics = %#v, want unsupported parameterized keyword", diagnostics)
	}
}

func lowerKeywordForTest(t *testing.T, oracleText string, kind game.Keyword) game.KeywordAbility {
	t.Helper()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Parameterized Creature",
		Layout:     "normal",
		TypeLine:   "Creature — Test",
		OracleText: oracleText,
		Power:      new("2"),
		Toughness:  new("2"),
	})
	if len(face.StaticAbilities) != 1 {
		t.Fatalf("static abilities = %d, want 1", len(face.StaticAbilities))
	}
	keyword, ok := game.BodyKeywordAbility(face.StaticAbilities[0].Body, kind)
	if !ok {
		t.Fatalf("%v keyword not found in %#v", kind, face.StaticAbilities[0].Body)
	}
	return keyword
}

func TestLowerSpellDamage(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Bolt",
		Layout:     "normal",
		TypeLine:   "Instant",
		OracleText: "Test Bolt deals 3 damage to any target.",
	})
	if !face.SpellAbility.Exists {
		t.Fatal("spell ability not lowered")
	}
	mode := face.SpellAbility.Val.Modes[0]
	if len(mode.Targets) != 1 {
		t.Fatalf("got %d targets, want 1", len(mode.Targets))
	}
	damage, ok := mode.Sequence[0].Primitive.(game.Damage)
	if !ok {
		t.Fatalf("primitive = %T, want game.Damage", mode.Sequence[0].Primitive)
	}
	if damage.Amount.Value() != 3 {
		t.Fatalf("damage amount = %d, want 3", damage.Amount.Value())
	}
}

func TestLowerSpellDamageQualifiedTarget(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Bolt",
		Layout:     "normal",
		TypeLine:   "Instant",
		OracleText: "Test Bolt deals 3 damage to target attacking or blocking creature.",
	})
	mode := face.SpellAbility.Val.Modes[0]
	if got := mode.Targets[0].Predicate.CombatState; got != game.CombatStateAttackingOrBlocking {
		t.Fatalf("combat state = %v, want attacking or blocking", got)
	}
}

func TestLowerSpellXAmounts(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		cardName   string
		oracleText string
		quantity   func(game.AbilityContent) game.Quantity
	}{
		{
			name:       "damage",
			cardName:   "Test Blaze",
			oracleText: "Test Blaze deals X damage to any target.",
			quantity: func(content game.AbilityContent) game.Quantity {
				primitive, ok := content.Modes[0].Sequence[0].Primitive.(game.Damage)
				if !ok {
					return game.Fixed(0)
				}
				return primitive.Amount
			},
		},
		{
			name:       "draw",
			cardName:   "Test Insight",
			oracleText: "Draw X cards.",
			quantity: func(content game.AbilityContent) game.Quantity {
				primitive, ok := content.Modes[0].Sequence[0].Primitive.(game.Draw)
				if !ok {
					return game.Fixed(0)
				}
				return primitive.Amount
			},
		},
		{
			name:       "life",
			cardName:   "Test Life",
			oracleText: "You gain X life.",
			quantity: func(content game.AbilityContent) game.Quantity {
				primitive, ok := content.Modes[0].Sequence[0].Primitive.(game.GainLife)
				if !ok {
					return game.Fixed(0)
				}
				return primitive.Amount
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			face := lowerSingleFace(t, &ScryfallCard{
				Name:       test.cardName,
				Layout:     "normal",
				TypeLine:   "Sorcery",
				OracleText: test.oracleText,
			})
			dynamic := test.quantity(face.SpellAbility.Val).DynamicAmount()
			if !dynamic.Exists || dynamic.Val.Kind != game.DynamicAmountX {
				t.Fatalf("dynamic amount = %+v, want X", dynamic)
			}
		})
	}
}

func TestLowerSpellDestroyQualifiedTarget(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Destroy",
		Layout:     "normal",
		TypeLine:   "Instant",
		OracleText: "Destroy target tapped creature an opponent controls.",
	})
	target := face.SpellAbility.Val.Modes[0].Targets[0]
	if target.Predicate.Tapped != game.TriTrue ||
		target.Predicate.Controller != game.ControllerOpponent {
		t.Fatalf("predicate = %+v, want tapped creature an opponent controls", target.Predicate)
	}
}

func TestLowerSpellReturnQualifiedTarget(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Return",
		Layout:     "normal",
		TypeLine:   "Instant",
		OracleText: "Return target creature you control to its owner's hand.",
	})
	target := face.SpellAbility.Val.Modes[0].Targets[0]
	if target.Predicate.Controller != game.ControllerYou {
		t.Fatalf("controller = %v, want ControllerYou", target.Predicate.Controller)
	}
}

func TestLowerSpellModifyPTQualifiedTarget(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Growth",
		Layout:     "normal",
		TypeLine:   "Instant",
		OracleText: "Target untapped creature you control gets +2/+2 until end of turn.",
	})
	target := face.SpellAbility.Val.Modes[0].Targets[0]
	if target.Predicate.Tapped != game.TriFalse ||
		target.Predicate.Controller != game.ControllerYou {
		t.Fatalf("predicate = %+v, want untapped creature you control", target.Predicate)
	}
}

func TestLowerOrderedSpellEffects(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Spell",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Destroy target artifact. Draw a card.",
	})
	if !face.SpellAbility.Exists {
		t.Fatal("spell ability not lowered")
	}
	mode := face.SpellAbility.Val.Modes[0]
	if len(mode.Targets) != 1 || len(mode.Sequence) != 2 {
		t.Fatalf("mode = %+v, want one target and two instructions", mode)
	}
	if _, ok := mode.Sequence[0].Primitive.(game.Destroy); !ok {
		t.Fatalf("first primitive = %T, want game.Destroy", mode.Sequence[0].Primitive)
	}
	draw, ok := mode.Sequence[1].Primitive.(game.Draw)
	if !ok || draw.Amount.Value() != 1 {
		t.Fatalf("second primitive = %+v, want draw one", mode.Sequence[1].Primitive)
	}
}

func TestLowerOrderedSpellEffectsWithMultipleTargets(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Spell",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Destroy target artifact. Tap target creature.",
	})
	mode := face.SpellAbility.Val.Modes[0]
	if len(mode.Targets) != 2 || len(mode.Sequence) != 2 {
		t.Fatalf("mode = %+v, want two targets and two instructions", mode)
	}
	destroy, ok := mode.Sequence[0].Primitive.(game.Destroy)
	if !ok || destroy.Object.TargetIndex() != 0 {
		t.Fatalf("first primitive = %+v, want target 0 destroy", mode.Sequence[0].Primitive)
	}
	tap, ok := mode.Sequence[1].Primitive.(game.Tap)
	if !ok || tap.Object.TargetIndex() != 1 {
		t.Fatalf("second primitive = %+v, want target 1 tap", mode.Sequence[1].Primitive)
	}
}

func TestLowerOrderedSpellEffectsRebasesEveryTargetClause(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Spell",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Destroy target artifact. Tap target creature. Target player mills three cards.",
	})
	mode := face.SpellAbility.Val.Modes[0]
	if len(mode.Targets) != 3 || len(mode.Sequence) != 3 {
		t.Fatalf("mode = %+v, want three targets and three instructions", mode)
	}
	destroy, destroyOK := mode.Sequence[0].Primitive.(game.Destroy)
	tap, tapOK := mode.Sequence[1].Primitive.(game.Tap)
	mill, millOK := mode.Sequence[2].Primitive.(game.Mill)
	if !destroyOK || !tapOK || !millOK {
		t.Fatalf(
			"primitives = %T, %T, %T; want game.Destroy, game.Tap, game.Mill",
			mode.Sequence[0].Primitive,
			mode.Sequence[1].Primitive,
			mode.Sequence[2].Primitive,
		)
	}
	if destroy.Object.TargetIndex() != 0 ||
		tap.Object.TargetIndex() != 1 ||
		mill.Player.TargetIndex() != 2 {
		t.Fatalf(
			"target indices = %d, %d, %d; want 0, 1, 2",
			destroy.Object.TargetIndex(),
			tap.Object.TargetIndex(),
			mill.Player.TargetIndex(),
		)
	}
}

func TestLowerSurveilSpell(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Surveil",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Surveil 2. (Look at the top two cards of your library, then put any number of them into your graveyard and the rest on top of your library in any order.)",
	})
	mode := face.SpellAbility.Val.Modes[0]
	surveil, ok := mode.Sequence[0].Primitive.(game.Surveil)
	if !ok ||
		surveil.Amount.Value() != 2 ||
		surveil.Player != game.ControllerReference() {
		t.Fatalf("primitive = %+v, want controller surveils two", mode.Sequence[0].Primitive)
	}
}

func TestLowerInvestigateSpell(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Investigate",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Investigate.",
	})
	mode := face.SpellAbility.Val.Modes[0]
	investigate, ok := mode.Sequence[0].Primitive.(game.Investigate)
	if !ok || investigate.Amount.Value() != 1 {
		t.Fatalf("primitive = %+v, want investigate once", mode.Sequence[0].Primitive)
	}
}

func TestLowerProliferateSpell(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Proliferate",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Proliferate.",
	})
	mode := face.SpellAbility.Val.Modes[0]
	if _, ok := mode.Sequence[0].Primitive.(game.Proliferate); !ok {
		t.Fatalf("primitive = %T, want game.Proliferate", mode.Sequence[0].Primitive)
	}
}

func TestLowerFixedCounterSpell(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Counter",
		Layout:     "normal",
		TypeLine:   "Instant",
		OracleText: "Put two +1/+1 counters on target creature.",
	})
	mode := face.SpellAbility.Val.Modes[0]
	if len(mode.Targets) != 1 || mode.Targets[0].Predicate.PermanentTypes[0] != types.Creature {
		t.Fatalf("targets = %+v, want one creature target", mode.Targets)
	}
	add, ok := mode.Sequence[0].Primitive.(game.AddCounter)
	if !ok ||
		add.Amount.Value() != 2 ||
		add.CounterKind != counter.PlusOnePlusOne ||
		add.Object != game.TargetPermanentReference(0) {
		t.Fatalf("primitive = %+v, want two +1/+1 counters on target 0", mode.Sequence[0].Primitive)
	}
}

func TestLowerRegenerateSpell(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Regenerate",
		Layout:     "normal",
		TypeLine:   "Instant",
		OracleText: "Regenerate target creature.",
	})
	mode := face.SpellAbility.Val.Modes[0]
	regenerate, ok := mode.Sequence[0].Primitive.(game.Regenerate)
	if !ok || regenerate.Object != game.TargetPermanentReference(0) {
		t.Fatalf("primitive = %+v, want regenerate target permanent", mode.Sequence[0].Primitive)
	}
}

func TestLowerFightSpell(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Fight",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Target creature you control fights target creature you don't control.",
	})
	mode := face.SpellAbility.Val.Modes[0]
	if len(mode.Targets) != 2 {
		t.Fatalf("targets = %+v, want two creatures", mode.Targets)
	}
	fight, ok := mode.Sequence[0].Primitive.(game.Fight)
	if !ok ||
		fight.Object != game.TargetPermanentReference(0) ||
		fight.RelatedObject != game.TargetPermanentReference(1) {
		t.Fatalf("primitive = %+v, want targets 0 and 1 fight", mode.Sequence[0].Primitive)
	}
}

func TestLowerLoyaltyAbilityPositiveCost(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Walker",
		Layout:     "normal",
		TypeLine:   "Legendary Planeswalker — Test",
		OracleText: "+1: Draw a card.",
		Loyalty:    func() *string { s := "3"; return &s }(),
	})
	if len(face.LoyaltyAbilities) != 1 {
		t.Fatalf("got %d loyalty abilities, want 1", len(face.LoyaltyAbilities))
	}
	la := face.LoyaltyAbilities[0]
	if la.LoyaltyCost != 1 {
		t.Fatalf("LoyaltyCost = %d, want 1", la.LoyaltyCost)
	}
	if la.Content.IsModal() || len(la.Content.Modes) != 1 {
		t.Fatalf("content = %+v, want single non-modal mode", la.Content)
	}
	draw, ok := la.Content.Modes[0].Sequence[0].Primitive.(game.Draw)
	if !ok || draw.Amount.Value() != 1 || draw.Player != game.ControllerReference() {
		t.Fatalf("primitive = %+v, want controller draws one", la.Content.Modes[0].Sequence[0].Primitive)
	}
}

func TestLowerLoyaltyAbilityNegativeCost(t *testing.T) {
	t.Parallel()
	loyaltyText := "\u22122: Target player mills three cards."
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Walker",
		Layout:     "normal",
		TypeLine:   "Legendary Planeswalker — Test",
		OracleText: loyaltyText,
		Loyalty:    func() *string { s := "4"; return &s }(),
	})
	if len(face.LoyaltyAbilities) != 1 {
		t.Fatalf("got %d loyalty abilities, want 1", len(face.LoyaltyAbilities))
	}
	la := face.LoyaltyAbilities[0]
	if la.LoyaltyCost != -2 {
		t.Fatalf("LoyaltyCost = %d, want -2", la.LoyaltyCost)
	}
	mill, ok := la.Content.Modes[0].Sequence[0].Primitive.(game.Mill)
	if !ok || mill.Amount.Value() != 3 {
		t.Fatalf("primitive = %+v, want mills three", la.Content.Modes[0].Sequence[0].Primitive)
	}
}

func TestLowerLoyaltyAbilityZeroCost(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Walker",
		Layout:     "normal",
		TypeLine:   "Legendary Planeswalker — Test",
		OracleText: "0: Scry 2.",
		Loyalty:    func() *string { s := "3"; return &s }(),
	})
	if len(face.LoyaltyAbilities) != 1 {
		t.Fatalf("got %d loyalty abilities, want 1", len(face.LoyaltyAbilities))
	}
	la := face.LoyaltyAbilities[0]
	if la.LoyaltyCost != 0 {
		t.Fatalf("LoyaltyCost = %d, want 0", la.LoyaltyCost)
	}
	scry, ok := la.Content.Modes[0].Sequence[0].Primitive.(game.Scry)
	if !ok || scry.Amount.Value() != 2 {
		t.Fatalf("primitive = %+v, want scry two", la.Content.Modes[0].Sequence[0].Primitive)
	}
}

func TestLowerLoyaltyAbilityMultiple(t *testing.T) {
	t.Parallel()
	oracleText := "+1: Draw a card.\n\u22122: You lose 3 life."
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Walker",
		Layout:     "normal",
		TypeLine:   "Legendary Planeswalker — Test",
		OracleText: oracleText,
		Loyalty:    func() *string { s := "3"; return &s }(),
	})
	if len(face.LoyaltyAbilities) != 2 {
		t.Fatalf("got %d loyalty abilities, want 2", len(face.LoyaltyAbilities))
	}
	if face.LoyaltyAbilities[0].LoyaltyCost != 1 {
		t.Fatalf("first LoyaltyCost = %d, want 1", face.LoyaltyAbilities[0].LoyaltyCost)
	}
	if face.LoyaltyAbilities[1].LoyaltyCost != -2 {
		t.Fatalf("second LoyaltyCost = %d, want -2", face.LoyaltyAbilities[1].LoyaltyCost)
	}
}

func TestLowerLoyaltyAbilityVariableCostRejected(t *testing.T) {
	t.Parallel()
	_, diagnostics := lowerExecutableFaces(&ScryfallCard{
		Name:       "Test Walker",
		Layout:     "normal",
		TypeLine:   "Legendary Planeswalker — Test",
		OracleText: "\u2212X: Target player mills X cards.",
		Loyalty:    func() *string { s := "3"; return &s }(),
	})
	if len(diagnostics) == 0 {
		t.Fatal("expected diagnostics for variable loyalty cost, got none")
	}
}

func TestLowerModalChooseOneSpell(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Charm",
		Layout:     "normal",
		TypeLine:   "Instant",
		OracleText: "Choose one \u2014\n\u2022 Draw a card.\n\u2022 You gain 3 life.",
	})
	if !face.SpellAbility.Exists {
		t.Fatal("no spell ability")
	}
	content := face.SpellAbility.Val
	if !content.IsModal() {
		t.Fatal("spell ability is not modal")
	}
	if len(content.Modes) != 2 {
		t.Fatalf("got %d modes, want 2", len(content.Modes))
	}
	if content.MinModes != 1 || content.MaxModes != 1 {
		t.Fatalf("MinModes=%d MaxModes=%d, want both 1", content.MinModes, content.MaxModes)
	}
	draw, ok := content.Modes[0].Sequence[0].Primitive.(game.Draw)
	if !ok || draw.Amount.Value() != 1 {
		t.Fatalf("mode 0 primitive = %+v, want draw one", content.Modes[0].Sequence[0].Primitive)
	}
	gain, ok := content.Modes[1].Sequence[0].Primitive.(game.GainLife)
	if !ok || gain.Amount.Value() != 3 {
		t.Fatalf("mode 1 primitive = %+v, want gain 3 life", content.Modes[1].Sequence[0].Primitive)
	}
}

func TestLowerModalChooseOneWithTarget(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Charm",
		Layout:     "normal",
		TypeLine:   "Instant",
		OracleText: "Choose one \u2014\n\u2022 Destroy target artifact.\n\u2022 Draw a card.",
	})
	content := face.SpellAbility.Val
	if !content.IsModal() || len(content.Modes) != 2 {
		t.Fatalf("content = %+v, want modal with 2 modes", content)
	}
	if len(content.Modes[0].Targets) != 1 {
		t.Fatalf("mode 0 targets = %+v, want one target", content.Modes[0].Targets)
	}
	if _, ok := content.Modes[0].Sequence[0].Primitive.(game.Destroy); !ok {
		t.Fatalf("mode 0 primitive = %T, want game.Destroy", content.Modes[0].Sequence[0].Primitive)
	}
}

func TestLowerModalChooseTwoSpell(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Command",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Choose two \u2014\n\u2022 Draw a card.\n\u2022 You gain 3 life.\n\u2022 Proliferate.",
	})
	content := face.SpellAbility.Val
	if content.MinModes != 2 || content.MaxModes != 2 {
		t.Fatalf("MinModes=%d MaxModes=%d, want both 2", content.MinModes, content.MaxModes)
	}
	if len(content.Modes) != 3 {
		t.Fatalf("got %d modes, want 3", len(content.Modes))
	}
}

func TestLowerModalChooseOneOrBoth(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Charm",
		Layout:     "normal",
		TypeLine:   "Instant",
		OracleText: "Choose one or both \u2014\n\u2022 Draw a card.\n\u2022 You gain 3 life.",
	})
	content := face.SpellAbility.Val
	if content.MinModes != 1 || content.MaxModes != 2 {
		t.Fatalf("MinModes=%d MaxModes=%d, want 1 and 2", content.MinModes, content.MaxModes)
	}
	if len(content.Modes) != 2 {
		t.Fatalf("got %d modes, want 2", len(content.Modes))
	}
}

func TestLowerModalChoiceCountExceedsModesRejected(t *testing.T) {
	t.Parallel()
	_, diagnostics := lowerExecutableFaces(&ScryfallCard{
		Name:       "Test Command",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Choose three \u2014\n\u2022 Draw a card.\n\u2022 You gain 3 life.",
	})
	if len(diagnostics) == 0 {
		t.Fatal("expected diagnostics when choice count exceeds modes, got none")
	}
}

func TestLowerModalUnsupportedModeBodyRejected(t *testing.T) {
	t.Parallel()
	_, diagnostics := lowerExecutableFaces(&ScryfallCard{
		Name:       "Test Charm",
		Layout:     "normal",
		TypeLine:   "Instant",
		OracleText: "Choose one \u2014\n\u2022 Draw a card.\n\u2022 Search your library for a card.",
	})
	if len(diagnostics) == 0 {
		t.Fatal("expected diagnostics for unsupported mode body, got none")
	}
}
