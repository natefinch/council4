package cardgen

import (
	"fmt"
	"go/parser"
	"go/token"
	"slices"
	"strings"
	"testing"

	"github.com/natefinch/council4/cardgen/oracle"
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

func TestLowerNinjutsuAbility(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Ninja",
		Layout:     "normal",
		TypeLine:   "Creature — Human Ninja",
		OracleText: "Ninjutsu {1}{U} ({1}{U}, Return an unblocked attacker you control to hand: Put this card onto the battlefield from your hand tapped and attacking.)",
		Power:      new("2"),
		Toughness:  new("2"),
	})
	if len(face.ActivatedAbilities) != 1 {
		t.Fatalf("got %d activated abilities, want 1", len(face.ActivatedAbilities))
	}
	ability := face.ActivatedAbilities[0]
	if !ability.ManaCost.Exists || !slices.Equal(ability.ManaCost.Val, cost.Mana{cost.O(1), cost.U}) {
		t.Fatalf("mana cost = %#v, want {1}{U}", ability.ManaCost)
	}
	if len(ability.AdditionalCosts) != 1 || ability.AdditionalCosts[0].Kind != cost.AdditionalReturnUnblockedAttacker {
		t.Fatalf("additional costs = %#v, want return unblocked attacker", ability.AdditionalCosts)
	}
	if len(ability.KeywordAbilities) != 1 {
		t.Fatalf("got %d keyword abilities, want 1", len(ability.KeywordAbilities))
	}
	if _, ok := ability.KeywordAbilities[0].(game.NinjutsuKeyword); !ok {
		t.Fatalf("keyword ability = %T, want game.NinjutsuKeyword", ability.KeywordAbilities[0])
	}
}

func TestLowerNinjutsuAbilityRejectsMalformedForms(t *testing.T) {
	t.Parallel()
	for _, oracleText := range []string{
		"Ninjutsu",
		"Ninjutsu {1}{U} extra text",
	} {
		t.Run(oracleText, func(t *testing.T) {
			t.Parallel()
			_, diagnostics := lowerExecutableFaces(&ScryfallCard{
				Name:       "Malformed Ninja",
				Layout:     "normal",
				TypeLine:   "Creature — Ninja",
				OracleText: oracleText,
				Power:      new("2"),
				Toughness:  new("2"),
			})
			if len(diagnostics) == 0 {
				t.Fatal("expected malformed Ninjutsu diagnostic")
			}
		})
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

func TestLowerStandaloneStaticKeywordGrants(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		oracleText string
		domain     game.GroupReferenceDomain
		excluded   bool
		subtypes   []types.Sub
		keywords   []game.Keyword
	}{
		"controlled creatures": {
			oracleText: "Creatures you control have haste and vigilance.",
			domain:     game.GroupDomainObjectControlled,
			keywords:   []game.Keyword{game.Haste, game.Vigilance},
		},
		"other controlled creatures": {
			oracleText: "Other creatures you control have flying.",
			domain:     game.GroupDomainObjectControlled,
			excluded:   true,
			keywords:   []game.Keyword{game.Flying},
		},
		"controlled artifacts": {
			oracleText: "Artifacts you control have indestructible.",
			domain:     game.GroupDomainObjectControlled,
			keywords:   []game.Keyword{game.Indestructible},
		},
		"equipped creature": {
			oracleText: "Equipped creature has shroud and wither.",
			domain:     game.GroupDomainAttachedObject,
			keywords:   []game.Keyword{game.Shroud, game.Wither},
		},
		"controlled subtype": {
			oracleText: "Zombies you control have flying.",
			domain:     game.GroupDomainObjectControlled,
			subtypes:   []types.Sub{types.Zombie},
			keywords:   []game.Keyword{game.Flying},
		},
		"other controlled subtype": {
			oracleText: "Other Dinosaurs you control have haste.",
			domain:     game.GroupDomainObjectControlled,
			excluded:   true,
			subtypes:   []types.Sub{types.Dinosaur},
			keywords:   []game.Keyword{game.Haste},
		},
		"irregular plural subtype": {
			oracleText: "Elves you control have vigilance.",
			domain:     game.GroupDomainObjectControlled,
			subtypes:   []types.Sub{types.Elf},
			keywords:   []game.Keyword{game.Vigilance},
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			face := lowerSingleFace(t, &ScryfallCard{
				Name:       "Test Grant",
				Layout:     "normal",
				TypeLine:   "Enchantment",
				OracleText: test.oracleText,
			})
			if len(face.StaticAbilities) != 1 {
				t.Fatalf("static abilities = %d, want 1", len(face.StaticAbilities))
			}
			effects := face.StaticAbilities[0].Body.ContinuousEffects
			if len(effects) != 1 {
				t.Fatalf("continuous effects = %#v, want 1", effects)
			}
			effect := effects[0]
			if effect.Layer != game.LayerAbility || effect.Group.Domain() != test.domain {
				t.Fatalf("continuous effect = %#v", effect)
			}
			if _, excluded := effect.Group.Exclusion(); excluded != test.excluded {
				t.Fatalf("group exclusion = %v, want %v", excluded, test.excluded)
			}
			if got := effect.Group.Selection().SubtypesAny; !slices.Equal(got, test.subtypes) {
				t.Fatalf("subtypes = %v, want %v", got, test.subtypes)
			}
			if !slices.Equal(effect.AddKeywords, test.keywords) {
				t.Fatalf("keywords = %v, want %v", effect.AddKeywords, test.keywords)
			}
		})
	}
}

func TestRejectUnknownSubtypeStaticKeywordGrant(t *testing.T) {
	t.Parallel()
	_, diagnostics := lowerExecutableFaces(&ScryfallCard{
		Name:       "Test Grant",
		Layout:     "normal",
		TypeLine:   "Enchantment",
		OracleText: "Splorps you control have haste.",
	})
	if len(diagnostics) == 0 || diagnostics[0].Summary != "unsupported static ability" {
		t.Fatalf("diagnostics = %#v, want unsupported static ability", diagnostics)
	}
}

func TestRejectMalformedStandaloneStaticKeywordGrants(t *testing.T) {
	t.Parallel()
	for _, oracleText := range []string{
		"Creatures you control have flying or haste.",
		"Creatures you control have and flying.",
		"Creatures you control have flying and.",
		"Creatures you control have flying haste.",
		"Creatures you control have infect.",
	} {
		_, diagnostics := lowerExecutableFaces(&ScryfallCard{
			Name:       "Test Grant",
			Layout:     "normal",
			TypeLine:   "Enchantment",
			OracleText: oracleText,
		})
		if len(diagnostics) == 0 {
			t.Fatalf("%q lowered without diagnostics", oracleText)
		}
	}
}

func TestLowerSourceConditionalKeywordGrant(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Climber",
		Layout:     "normal",
		TypeLine:   "Creature — Ape",
		OracleText: "As long as you control a Mountain, this creature has menace and vigilance.",
		Power:      new("3"),
		Toughness:  new("3"),
	})
	if len(face.StaticAbilities) != 1 {
		t.Fatalf("static abilities = %d, want 1", len(face.StaticAbilities))
	}
	ability := face.StaticAbilities[0].Body
	if !ability.Condition.Exists {
		t.Fatal("static ability has no condition")
	}
	condition := ability.Condition.Val
	if condition.Text != "As long as you control a Mountain" ||
		!slices.Equal(condition.ControllerControls.SubtypesAny, []types.Sub{types.Mountain}) {
		t.Fatalf("condition = %+v", condition)
	}
	if len(ability.ContinuousEffects) != 1 {
		t.Fatalf("continuous effects = %#v", ability.ContinuousEffects)
	}
	effect := ability.ContinuousEffects[0]
	if effect.Layer != game.LayerAbility ||
		!effect.AffectedSource ||
		!slices.Equal(effect.AddKeywords, []game.Keyword{game.Menace, game.Vigilance}) {
		t.Fatalf("continuous effect = %+v", effect)
	}
}

func TestLowerPostfixSourceConditionalKeywordGrant(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Healer",
		Layout:     "normal",
		TypeLine:   "Creature — Cleric",
		OracleText: "This creature has lifelink as long as you control another Cleric.",
		Power:      new("2"),
		Toughness:  new("2"),
	})
	ability := face.StaticAbilities[0].Body
	condition := ability.Condition.Val
	if !condition.ControllerControls.ExcludeSource ||
		!slices.Equal(condition.ControllerControls.SubtypesAny, []types.Sub{types.Cleric}) {
		t.Fatalf("condition = %+v", condition)
	}
	effect := ability.ContinuousEffects[0]
	if !effect.AffectedSource || !slices.Equal(effect.AddKeywords, []game.Keyword{game.Lifelink}) {
		t.Fatalf("continuous effect = %+v", effect)
	}
}

func TestLowerPostfixLandSubtypeConditionalKeywordGrant(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Sergeant",
		Layout:     "normal",
		TypeLine:   "Creature — Soldier",
		OracleText: "This creature has double strike as long as you control a Gate.",
		Power:      new("3"),
		Toughness:  new("3"),
	})
	condition := face.StaticAbilities[0].Body.Condition.Val
	if !slices.Equal(condition.ControllerControls.SubtypesAny, []types.Sub{types.Gate}) {
		t.Fatalf("condition = %+v", condition)
	}
}

func TestLowerColorQualifiedSourceConditionalKeywordGrants(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		oracleText     string
		types          []types.Card
		colors         []color.Color
		excludedColors []color.Color
	}{
		"one color": {
			oracleText: "This creature has haste as long as you control a red creature.",
			types:      []types.Card{types.Creature},
			colors:     []color.Color{color.Red},
		},
		"either color": {
			oracleText: "This creature has lifelink as long as you control a white or black permanent.",
			colors:     []color.Color{color.White, color.Black},
		},
		"colorless": {
			oracleText: "This creature has haste as long as you control another colorless creature.",
			types:      []types.Card{types.Creature},
			excludedColors: []color.Color{
				color.White,
				color.Blue,
				color.Black,
				color.Red,
				color.Green,
			},
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			face := lowerSingleFace(t, &ScryfallCard{
				Name:       "Test Creature",
				Layout:     "normal",
				TypeLine:   "Creature — Human",
				OracleText: test.oracleText,
				Power:      new("2"),
				Toughness:  new("2"),
			})
			filter := face.StaticAbilities[0].Body.Condition.Val.ControllerControls
			if !slices.Equal(filter.Types, test.types) ||
				!slices.Equal(filter.ColorsAny, test.colors) ||
				!slices.Equal(filter.ExcludedColors, test.excludedColors) {
				t.Fatalf("filter = %+v", filter)
			}
		})
	}
}

func TestGenerateSourceConditionalKeywordGrant(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Test Flier",
		Layout:     "normal",
		TypeLine:   "Creature — Human",
		OracleText: "As long as you control an artifact, this creature has flying.",
		Power:      new("2"),
		Toughness:  new("2"),
	}, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, want := range []string{
		`Condition: opt.Val(game.Condition{`,
		`Text: "As long as you control an artifact"`,
		`Types: []types.Card{types.Artifact}`,
		`AffectedSource: true`,
		`game.Flying`,
	} {
		if !strings.Contains(source, want) {
			t.Fatalf("source missing %q:\n%s", want, source)
		}
	}
}

func TestRejectUnsupportedSourceConditionalKeywordGrant(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Test Attacker",
		Layout:     "normal",
		TypeLine:   "Creature — Human",
		OracleText: "As long as it's attacking, this creature has flying.",
		Power:      new("2"),
		Toughness:  new("2"),
	}, "t")
	if err != nil {
		t.Fatal(err)
	}
	if source != "" {
		t.Fatalf("unexpected source:\n%s", source)
	}
	if len(diagnostics) == 0 {
		t.Fatal("expected unsupported conditional keyword diagnostic")
	}
}

func TestRejectStaticPTBuffWithUnsupportedKeywordText(t *testing.T) {
	t.Parallel()
	for _, oracleText := range []string{
		"Equipped creature gets +2/+2 and has trample or lifelink.\nEquip {3}",
		"Equipped creature gets +2/+2 and has and trample.\nEquip {3}",
		"Equipped creature gets +2/+2 and has trample and.\nEquip {3}",
		"Equipped creature gets +2/+2 and has flying lifelink.\nEquip {3}",
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

func TestLowerOptionalEntryPayments(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		oracleText string
		assert     func(*testing.T, game.ResolutionPayment)
	}{
		{
			name:       "pay life",
			oracleText: "As this land enters, you may pay 2 life. If you don't, it enters tapped.",
			assert: func(t *testing.T, payment game.ResolutionPayment) {
				if len(payment.AdditionalCosts) != 1 ||
					payment.AdditionalCosts[0].Kind != cost.AdditionalPayLife ||
					payment.AdditionalCosts[0].Amount != 2 {
					t.Fatalf("payment = %+v, want pay 2 life", payment)
				}
			},
		},
		{
			name:       "reveal land subtype",
			oracleText: "As this land enters, you may reveal a Mountain or Forest card from your hand. If you don't, this land enters tapped.",
			assert: func(t *testing.T, payment game.ResolutionPayment) {
				if len(payment.AdditionalCosts) != 1 {
					t.Fatalf("payment = %+v, want one reveal cost", payment)
				}
				additional := payment.AdditionalCosts[0]
				if additional.Kind != cost.AdditionalReveal ||
					additional.Source != zone.Hand ||
					additional.SubtypesAny != (cost.SubtypeSet{types.Mountain, types.Forest}) {
					t.Fatalf("additional cost = %+v, want Mountain-or-Forest reveal from hand", additional)
				}
			},
		},
		{
			name:       "reveal creature subtype",
			oracleText: "As this land enters, you may reveal a Giant card from your hand. If you don't, this land enters tapped.",
			assert: func(t *testing.T, payment game.ResolutionPayment) {
				if len(payment.AdditionalCosts) != 1 ||
					payment.AdditionalCosts[0].SubtypesAny != (cost.SubtypeSet{types.Giant}) {
					t.Fatalf("payment = %+v, want Giant reveal", payment)
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
				OracleText: test.oracleText,
			})
			if len(face.ReplacementAbilities) != 1 ||
				!face.ReplacementAbilities[0].UnlessPaid.Exists {
				t.Fatalf("replacement abilities = %+v, want one paid replacement", face.ReplacementAbilities)
			}
			test.assert(t, face.ReplacementAbilities[0].UnlessPaid.Val)
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

func TestLowerWasCastEnterTriggers(t *testing.T) {
	t.Parallel()
	for _, condition := range []string{"if it was cast", "if you cast it"} {
		t.Run(condition, func(t *testing.T) {
			t.Parallel()
			face := lowerSingleFace(t, &ScryfallCard{
				Name:       "Test Construct",
				Layout:     "normal",
				TypeLine:   "Artifact Creature — Construct",
				OracleText: "When this creature enters, " + condition + ", draw a card.",
				Power:      new("2"),
				Toughness:  new("2"),
			})
			trigger := face.TriggeredAbilities[0].Trigger
			if trigger.InterveningIf != condition || !trigger.InterveningIfEventPermanentWasCast {
				t.Fatalf("trigger = %+v, want was-cast intervening-if", trigger)
			}
		})
	}
}

func TestLowerAttackedThisTurnEnterTriggerFailsClosed(t *testing.T) {
	t.Parallel()
	_, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Test Warrior",
		Layout:     "normal",
		TypeLine:   "Creature — Warrior",
		OracleText: "When this creature enters, if this creature attacked this turn, draw a card.",
		Power:      new("2"),
		Toughness:  new("2"),
	}, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) == 0 {
		t.Fatal("attacked-this-turn self-enter condition unexpectedly lowered")
	}
}

func TestLowerControlsPermanentEnterTrigger(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Artificer",
		Layout:     "normal",
		TypeLine:   "Creature — Artificer",
		OracleText: "When this creature enters, if you control an artifact, draw a card.",
		Power:      new("2"),
		Toughness:  new("2"),
	})
	trigger := face.TriggeredAbilities[0].Trigger
	if trigger.InterveningIf != "if you control an artifact" ||
		!trigger.InterveningCondition.Exists {
		t.Fatalf("trigger = %+v, want controls-artifact intervening-if", trigger)
	}
	selection := trigger.InterveningCondition.Val.ControlsMatching
	if !selection.Exists ||
		!slices.Equal(selection.Val.Selection.RequiredTypes, []types.Card{types.Artifact}) {
		t.Fatalf("condition = %+v, want controls an artifact", trigger.InterveningCondition.Val)
	}
}

func TestLowerEnterTriggerRejectsUnsupportedInterveningWording(t *testing.T) {
	t.Parallel()
	_, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Test Handler",
		Layout:     "normal",
		TypeLine:   "Creature — Elf",
		OracleText: "When this creature enters, if you control an Elf, draw a card.",
		Power:      new("2"),
		Toughness:  new("2"),
	}, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) == 0 {
		t.Fatal("unsupported subtype condition unexpectedly lowered")
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

func TestLowerReadAheadSaga(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Read Ahead Saga",
		Layout:     "saga",
		TypeLine:   "Enchantment — Saga",
		OracleText: "Read ahead (Choose a chapter and start with that many lore counters. Add one after your draw step. Skipped chapters don't trigger.)\nI — Draw a card.\nII — Draw a card.",
	})
	if len(face.StaticAbilities) != 1 || !game.BodyHasKeyword(face.StaticAbilities[0].Body, game.ReadAhead) {
		t.Fatalf("static abilities = %#v, want ReadAheadStaticBody", face.StaticAbilities)
	}
	if len(face.ChapterAbilities) != 2 {
		t.Fatalf("chapter abilities = %#v, want two", face.ChapterAbilities)
	}
}

func TestLowerReadAheadRejectsNoncanonicalReminder(t *testing.T) {
	t.Parallel()
	_, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Malformed Read Ahead Saga",
		Layout:     "saga",
		TypeLine:   "Enchantment — Saga",
		OracleText: "Read ahead (Choose whichever chapter you want.)\nI — Draw a card.",
	}, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) == 0 {
		t.Fatal("noncanonical Read ahead reminder unexpectedly lowered")
	}
}

func TestLowerReadAheadRejectsMismatchedSacrificeChapter(t *testing.T) {
	t.Parallel()
	_, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Mismatched Read Ahead Saga",
		Layout:     "saga",
		TypeLine:   "Enchantment — Saga",
		OracleText: "Read ahead (Choose a chapter and start with that many lore counters. Add one after your draw step. Skipped chapters don't trigger. Sacrifice after IV.)\nI — Draw a card.\nII — Draw a card.\nIII — Draw a card.",
	}, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) == 0 {
		t.Fatal("mismatched Read ahead sacrifice chapter unexpectedly lowered")
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

func TestLowerDiesTriggerHadNoPlusPlusCounters(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Undying Bear",
		Layout:     "normal",
		TypeLine:   "Creature — Bear",
		OracleText: "When this creature dies, if it had no +1/+1 counters, draw a card.",
		Power:      new("2"),
		Toughness:  new("2"),
	})
	trigger := face.TriggeredAbilities[0].Trigger
	if trigger.InterveningIf != "if it had no +1/+1 counters" ||
		!trigger.InterveningIfEventPermanentHadNoCounterKind.Exists ||
		trigger.InterveningIfEventPermanentHadNoCounterKind.Val != counter.PlusOnePlusOne {
		t.Fatalf("trigger = %+v, want no +1/+1 counters intervening-if", trigger)
	}
}

func TestLowerDiesTriggerHadNoMinusMinusCounters(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Persist Bear",
		Layout:     "normal",
		TypeLine:   "Creature — Bear",
		OracleText: "When this creature dies, if it had no -1/-1 counters on it, it deals 3 damage to any target.",
		Power:      new("2"),
		Toughness:  new("2"),
	})
	ability := face.TriggeredAbilities[0]
	trigger := ability.Trigger
	if trigger.InterveningIf != "if it had no -1/-1 counters on it" ||
		!trigger.InterveningIfEventPermanentHadNoCounterKind.Exists ||
		trigger.InterveningIfEventPermanentHadNoCounterKind.Val != counter.MinusOneMinusOne {
		t.Fatalf("trigger = %+v, want no -1/-1 counters intervening-if", trigger)
	}
	damage, ok := ability.Content.Modes[0].Sequence[0].Primitive.(game.Damage)
	if !ok || !damage.DamageSource.Exists ||
		damage.DamageSource.Val != game.EventPermanentReference() {
		t.Fatalf("primitive = %+v, want damage from event permanent", ability.Content.Modes[0].Sequence[0].Primitive)
	}
}

func TestLowerDiesTriggerOptional(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Bear",
		Layout:     "normal",
		TypeLine:   "Creature — Bear",
		OracleText: "When this creature dies, you may draw a card.",
		Power:      new("2"),
		Toughness:  new("2"),
	})
	ability := face.TriggeredAbilities[0]
	if !ability.Optional {
		t.Fatal("dies trigger is not optional")
	}
	if _, ok := ability.Content.Modes[0].Sequence[0].Primitive.(game.Draw); !ok {
		t.Fatalf("primitive = %T, want game.Draw", ability.Content.Modes[0].Sequence[0].Primitive)
	}
}

func TestLowerDiesTriggerRejectsAmbiguousCounterAbsence(t *testing.T) {
	t.Parallel()
	for _, condition := range []string{
		"if it had no counters on it",
		"if it had no charge counters on it",
		"if it didn't have a +1/+1 counter on it",
	} {
		t.Run(condition, func(t *testing.T) {
			t.Parallel()
			_, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
				Name:       "Test Bear",
				Layout:     "normal",
				TypeLine:   "Creature — Bear",
				OracleText: "When this creature dies, " + condition + ", draw a card.",
				Power:      new("2"),
				Toughness:  new("2"),
			}, "t")
			if err != nil {
				t.Fatal(err)
			}
			if len(diagnostics) == 0 {
				t.Fatalf("ambiguous or unsupported condition %q unexpectedly lowered", condition)
			}
		})
	}
}

func TestLowerDiesTriggerReturnsEventCardToOwnersHand(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Bear",
		Layout:     "normal",
		TypeLine:   "Creature — Bear",
		OracleText: "When this creature dies, return it to its owner's hand.",
		Power:      new("2"),
		Toughness:  new("2"),
	})
	primitive := face.TriggeredAbilities[0].Content.Modes[0].Sequence[0].Primitive
	move, ok := primitive.(game.MoveCard)
	if !ok {
		t.Fatalf("primitive = %T, want game.MoveCard", primitive)
	}
	if move.Card.Kind != game.CardReferenceEvent ||
		move.FromZone != zone.Graveyard ||
		move.Destination != zone.Hand {
		t.Fatalf("move = %+v, want event card from graveyard to hand", move)
	}
}

func TestLowerDiesTriggerGrantsAdventureCastFromGraveyard(t *testing.T) {
	t.Parallel()
	faces, diagnostics := lowerExecutableFaces(&ScryfallCard{
		Name:   "Test Dreadknight // Test Whispers",
		Layout: "adventure",
		CardFaces: []ScryfallCardFace{
			{
				Name:       "Test Dreadknight",
				ManaCost:   "{1}{G}",
				TypeLine:   "Creature — Human Knight",
				OracleText: "When Test Dreadknight dies, you may cast it from your graveyard as an Adventure until the end of your next turn.",
				Power:      new("2"),
				Toughness:  new("1"),
			},
			{
				Name:       "Test Whispers",
				ManaCost:   "{1}{B}",
				TypeLine:   "Sorcery — Adventure",
				OracleText: "Draw a card.",
			},
		},
	})
	if len(diagnostics) != 0 {
		t.Fatalf("unexpected diagnostics: %#v", diagnostics)
	}
	ability := faces[0].TriggeredAbilities[0]
	if !ability.Optional {
		t.Fatal("cast-permission dies trigger is not optional")
	}
	primitive := ability.Content.Modes[0].Sequence[0].Primitive
	permission, ok := primitive.(game.GrantCastPermission)
	if !ok {
		t.Fatalf("primitive = %T, want game.GrantCastPermission", primitive)
	}
	if permission.Card.Kind != game.CardReferenceEvent ||
		permission.FromZone != zone.Graveyard ||
		permission.Face != game.FaceAlternate ||
		permission.Duration != game.DurationUntilEndOfYourNextTurn {
		t.Fatalf("permission = %+v, want event Adventure cast through next turn", permission)
	}
}

func TestLowerDiesTriggerRejectsAmbiguousEventCardReference(t *testing.T) {
	t.Parallel()
	for _, text := range []string{
		"When this creature dies, return it to the battlefield.",
		"When this creature dies, cast it.",
		"When this creature dies, you may cast it from your graveyard.",
		"When this creature dies, return it to its owner's hand or the battlefield.",
	} {
		t.Run(text, func(t *testing.T) {
			t.Parallel()
			_, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
				Name:       "Test Bear",
				Layout:     "normal",
				TypeLine:   "Creature — Bear",
				OracleText: text,
				Power:      new("2"),
				Toughness:  new("2"),
			}, "t")
			if err != nil {
				t.Fatal(err)
			}
			if len(diagnostics) == 0 {
				t.Fatalf("ambiguous event-card reference unexpectedly lowered: %q", text)
			}
		})
	}
}

func TestLowerDiesTriggerRejectsEnterOnlyInterveningConditions(t *testing.T) {
	t.Parallel()
	for _, condition := range []string{
		"if it was kicked",
		"if it was cast",
		"if you cast it",
		"if this creature attacked this turn",
		"if you control an artifact",
	} {
		t.Run(condition, func(t *testing.T) {
			t.Parallel()
			_, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
				Name:       "Test Bear",
				Layout:     "normal",
				TypeLine:   "Creature — Bear",
				OracleText: "When this creature dies, " + condition + ", draw a card.",
				Power:      new("2"),
				Toughness:  new("2"),
			}, "t")
			if err != nil {
				t.Fatal(err)
			}
			if len(diagnostics) == 0 {
				t.Fatalf("self-dies trigger unexpectedly lowered with %q", condition)
			}
		})
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

func TestLowerDynamicEffectAmounts(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		oracleText string
		quantity   func(game.AbilityContent) game.Quantity
		kind       game.DynamicAmountKind
		multiplier int
		cardType   types.Card
		controller game.ControllerRelation
	}{
		{"controlled creatures damage", "Test Swarm deals damage equal to the number of creatures you control to any target.", func(content game.AbilityContent) game.Quantity {
			primitive, ok := content.Modes[0].Sequence[0].Primitive.(game.Damage)
			if !ok {
				return game.Fixed(0)
			}
			return primitive.Amount
		}, game.DynamicAmountCountSelector, 1, types.Creature, game.ControllerYou},
		{"twice battlefield lands damage", "Test Swarm deals damage equal to twice the number of lands on the battlefield to any target.", func(content game.AbilityContent) game.Quantity {
			primitive, ok := content.Modes[0].Sequence[0].Primitive.(game.Damage)
			if !ok {
				return game.Fixed(0)
			}
			return primitive.Amount
		}, game.DynamicAmountCountSelector, 2, types.Land, game.ControllerAny},
		{"life for opponents", "You gain 2 life for each opponent you have.", func(content game.AbilityContent) game.Quantity {
			primitive, ok := content.Modes[0].Sequence[0].Primitive.(game.GainLife)
			if !ok {
				return game.Fixed(0)
			}
			return primitive.Amount
		}, game.DynamicAmountOpponentCount, 2, "", game.ControllerAny},
		{"controller life", "You gain life equal to your life total.", func(content game.AbilityContent) game.Quantity {
			primitive, ok := content.Modes[0].Sequence[0].Primitive.(game.GainLife)
			if !ok {
				return game.Fixed(0)
			}
			return primitive.Amount
		}, game.DynamicAmountControllerLife, 1, "", game.ControllerAny},
		{"draw for controlled lands", "Draw a card for each land you control.", func(content game.AbilityContent) game.Quantity {
			primitive, ok := content.Modes[0].Sequence[0].Primitive.(game.Draw)
			if !ok {
				return game.Fixed(0)
			}
			return primitive.Amount
		}, game.DynamicAmountCountSelector, 1, types.Land, game.ControllerYou},
		{"power for opponents", "Target creature gets +1/+0 for each opponent you have until end of turn.", func(content game.AbilityContent) game.Quantity {
			primitive, ok := content.Modes[0].Sequence[0].Primitive.(game.ModifyPT)
			if !ok {
				return game.Fixed(0)
			}
			return primitive.PowerDelta
		}, game.DynamicAmountOpponentCount, 1, "", game.ControllerAny},
		{"power after duration", "Target creature gets +1/+0 until end of turn for each opponent you have.", func(content game.AbilityContent) game.Quantity {
			primitive, ok := content.Modes[0].Sequence[0].Primitive.(game.ModifyPT)
			if !ok {
				return game.Fixed(0)
			}
			return primitive.PowerDelta
		}, game.DynamicAmountOpponentCount, 1, "", game.ControllerAny},
		{"counters for controlled lands", "Put X +1/+1 counters on target creature, where X is the number of lands you control.", func(content game.AbilityContent) game.Quantity {
			primitive, ok := content.Modes[0].Sequence[0].Primitive.(game.AddCounter)
			if !ok {
				return game.Fixed(0)
			}
			return primitive.Amount
		}, game.DynamicAmountCountSelector, 1, types.Land, game.ControllerYou},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			face := lowerSingleFace(t, &ScryfallCard{
				Name:       "Test Swarm",
				Layout:     "normal",
				TypeLine:   "Sorcery",
				OracleText: test.oracleText,
			})
			dynamic := test.quantity(face.SpellAbility.Val).DynamicAmount()
			if !dynamic.Exists ||
				dynamic.Val.Kind != test.kind ||
				dynamic.Val.Multiplier != test.multiplier {
				t.Fatalf("dynamic amount = %+v", dynamic)
			}
			if test.cardType != "" {
				selection := dynamic.Val.Group.Selection()
				if len(selection.RequiredTypes) != 1 ||
					selection.RequiredTypes[0] != test.cardType ||
					selection.Controller != test.controller {
					t.Fatalf("selection = %+v", selection)
				}
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

func TestLowerNamedCounterPlacement(t *testing.T) {
	t.Parallel()
	tests := []struct {
		text string
		kind counter.Kind
	}{
		{"Put a charge counter on target artifact.", counter.Charge},
		{"Put two shield counters on target creature you control.", counter.Shield},
		{"Put a first strike counter on target creature.", counter.FirstStrike},
	}
	for _, test := range tests {
		t.Run(test.text, func(t *testing.T) {
			t.Parallel()
			face := lowerSingleFace(t, &ScryfallCard{
				Name:       "Test Counter",
				Layout:     "normal",
				TypeLine:   "Instant",
				OracleText: test.text,
			})
			add, ok := face.SpellAbility.Val.Modes[0].Sequence[0].Primitive.(game.AddCounter)
			if !ok || add.CounterKind != test.kind {
				t.Fatalf("primitive = %+v", face.SpellAbility.Val.Modes[0].Sequence[0].Primitive)
			}
		})
	}
}

func TestLowerKeywordNamedCounterPlacementAbilityShapes(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		card    *ScryfallCard
		content func(loweredFaceAbilities) (game.AbilityContent, bool)
		kind    counter.Kind
	}{
		{
			name: "activated",
			card: &ScryfallCard{
				Name:       "Test Counter",
				Layout:     "normal",
				TypeLine:   "Artifact",
				OracleText: "{T}: Put a flying counter on target creature.",
			},
			content: func(face loweredFaceAbilities) (game.AbilityContent, bool) {
				if len(face.ActivatedAbilities) != 1 {
					return game.AbilityContent{}, false
				}
				return face.ActivatedAbilities[0].Content, true
			},
			kind: counter.Flying,
		},
		{
			name: "loyalty",
			card: &ScryfallCard{
				Name:       "Test Walker",
				Layout:     "normal",
				TypeLine:   "Legendary Planeswalker — Test",
				OracleText: "+1: Put a lifelink counter on target creature.",
				Loyalty:    func() *string { loyalty := "3"; return &loyalty }(),
			},
			content: func(face loweredFaceAbilities) (game.AbilityContent, bool) {
				if len(face.LoyaltyAbilities) != 1 {
					return game.AbilityContent{}, false
				}
				return face.LoyaltyAbilities[0].Content, true
			},
			kind: counter.Lifelink,
		},
		{
			name: "triggered",
			card: &ScryfallCard{
				Name:       "Test Counter",
				Layout:     "normal",
				TypeLine:   "Creature — Test",
				OracleText: "When this creature enters, put a first strike counter on target creature.",
				Power:      new("2"),
				Toughness:  new("2"),
			},
			content: func(face loweredFaceAbilities) (game.AbilityContent, bool) {
				if len(face.TriggeredAbilities) != 1 {
					return game.AbilityContent{}, false
				}
				return face.TriggeredAbilities[0].Content, true
			},
			kind: counter.FirstStrike,
		},
		{
			name: "phase triggered",
			card: &ScryfallCard{
				Name:       "Test Counter",
				Layout:     "normal",
				TypeLine:   "Enchantment",
				OracleText: "At the beginning of your upkeep, put a flying counter on target creature.",
			},
			content: func(face loweredFaceAbilities) (game.AbilityContent, bool) {
				if len(face.TriggeredAbilities) != 1 {
					return game.AbilityContent{}, false
				}
				return face.TriggeredAbilities[0].Content, true
			},
			kind: counter.Flying,
		},
		{
			name: "non-self enter triggered",
			card: &ScryfallCard{
				Name:       "Test Counter",
				Layout:     "normal",
				TypeLine:   "Enchantment",
				OracleText: "Whenever another creature enters, put a lifelink counter on target creature.",
			},
			content: func(face loweredFaceAbilities) (game.AbilityContent, bool) {
				if len(face.TriggeredAbilities) != 1 {
					return game.AbilityContent{}, false
				}
				return face.TriggeredAbilities[0].Content, true
			},
			kind: counter.Lifelink,
		},
		{
			name: "ordered effects",
			card: &ScryfallCard{
				Name:       "Test Counter",
				Layout:     "normal",
				TypeLine:   "Sorcery",
				OracleText: "Put a flying counter on target creature. Draw a card.",
			},
			content: func(face loweredFaceAbilities) (game.AbilityContent, bool) {
				if !face.SpellAbility.Exists {
					return game.AbilityContent{}, false
				}
				return face.SpellAbility.Val, true
			},
			kind: counter.Flying,
		},
		{
			name: "Saga chapter",
			card: &ScryfallCard{
				Name:       "Test Saga",
				Layout:     "saga",
				TypeLine:   "Enchantment — Saga",
				OracleText: "I — Put a lifelink counter on target creature.",
			},
			content: func(face loweredFaceAbilities) (game.AbilityContent, bool) {
				if len(face.ChapterAbilities) != 1 {
					return game.AbilityContent{}, false
				}
				return face.ChapterAbilities[0].Content, true
			},
			kind: counter.Lifelink,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			face := lowerSingleFace(t, test.card)
			content, ok := test.content(face)
			if !ok ||
				len(content.Modes) != 1 ||
				len(content.Modes[0].Sequence) == 0 {
				t.Fatalf("face = %+v, want lowered counter placement", face)
			}
			add, ok := content.Modes[0].Sequence[0].Primitive.(game.AddCounter)
			if !ok || add.CounterKind != test.kind {
				t.Fatalf("primitive = %+v, want %s counter placement", content.Modes[0].Sequence[0].Primitive, test.kind)
			}
		})
	}
}

func TestLowerPlayerCounterPlacement(t *testing.T) {
	t.Parallel()
	tests := []struct {
		text string
		kind counter.Kind
	}{
		{"Put an energy counter on target player.", counter.Energy},
		{"Put two experience counters on target player.", counter.Experience},
		{"Put three poison counters on target opponent.", counter.Poison},
	}
	for _, test := range tests {
		t.Run(test.text, func(t *testing.T) {
			t.Parallel()
			face := lowerSingleFace(t, &ScryfallCard{
				Name:       "Test Counter",
				Layout:     "normal",
				TypeLine:   "Sorcery",
				OracleText: test.text,
			})
			mode := face.SpellAbility.Val.Modes[0]
			add, ok := mode.Sequence[0].Primitive.(game.AddPlayerCounter)
			if !ok ||
				add.CounterKind != test.kind ||
				add.Player != game.TargetPlayerReference(0) ||
				mode.Targets[0].Allow != game.TargetAllowPlayer {
				t.Fatalf("mode = %+v", mode)
			}
		})
	}
}

func TestLowerEveryRecognizedCounterKindOnItsValidTarget(t *testing.T) {
	t.Parallel()
	for kind := counter.PlusOnePlusOne; kind <= counter.Experience; kind++ {
		if kind == counter.Stun || kind == counter.Finality {
			continue
		}
		if !oracle.CounterKindPlacementSupported(kind) {
			t.Fatalf("%s unexpectedly excluded from named placement", kind)
		}
		name := kind.String()
		article := "a"
		if strings.ContainsRune("aeiou", rune(name[0])) {
			article = "an"
		}
		target := "target permanent"
		if kind.PlayerOnly() {
			target = "target player"
		}
		oracleText := fmt.Sprintf("Put %s %s counter on %s.", article, name, target)
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			face := lowerSingleFace(t, &ScryfallCard{
				Name:       "Test Counter",
				Layout:     "normal",
				TypeLine:   "Sorcery",
				OracleText: oracleText,
			})
			primitive := face.SpellAbility.Val.Modes[0].Sequence[0].Primitive
			if kind.PlayerOnly() {
				add, ok := primitive.(game.AddPlayerCounter)
				if !ok || add.CounterKind != kind {
					t.Fatalf("primitive = %+v", primitive)
				}
				return
			}
			add, ok := primitive.(game.AddCounter)
			if !ok || add.CounterKind != kind {
				t.Fatalf("primitive = %+v", primitive)
			}
		})
	}
}

func TestLowerCounterPlacementRejectsMissingRuntimeMechanics(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		name string
		kind counter.Kind
	}{
		{"stun", counter.Stun},
		{"finality", counter.Finality},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			ability := oracle.CompiledAbility{
				Text: "Put a " + test.name + " counter on target creature.",
				Targets: []oracle.CompiledTarget{{
					Text:        "target creature",
					Cardinality: oracle.TargetCardinality{Min: 1, Max: 1},
					Selector:    oracle.CompiledSelector{Kind: oracle.SelectorCreature},
				}},
				Effects: []oracle.CompiledEffect{{
					Kind:             oracle.EffectPut,
					Amount:           oracle.CompiledAmount{Value: 1, Known: true},
					CounterKind:      test.kind,
					CounterKindKnown: true,
				}},
			}
			if _, diagnostic := lowerCounterPlacementSpell(ability); diagnostic == nil {
				t.Fatal("lowering accepted counter kind without runtime mechanics")
			}
		})
	}
}

func TestLowerDynamicNamedCounterPlacement(t *testing.T) {
	t.Parallel()
	tests := []struct {
		text string
		kind game.DynamicAmountKind
	}{
		{"Put X charge counters on target artifact.", game.DynamicAmountX},
		{"Put X poison counters on target player, where X is the number of lands you control.", game.DynamicAmountCountSelector},
		{"Put X energy counters on target player, where X is Test Counter's power.", game.DynamicAmountObjectPower},
	}
	for _, test := range tests {
		face := lowerSingleFace(t, &ScryfallCard{
			Name:       "Test Counter",
			Layout:     "normal",
			TypeLine:   "Sorcery",
			OracleText: test.text,
		})
		primitive := face.SpellAbility.Val.Modes[0].Sequence[0].Primitive
		amount, ok := counterPlacementAmount(primitive)
		if !ok {
			t.Fatalf("%q primitive = %T", test.text, primitive)
		}
		dynamic := amount.DynamicAmount()
		if !dynamic.Exists || dynamic.Val.Kind != test.kind {
			t.Fatalf("%q amount = %+v", test.text, dynamic)
		}
		if test.kind == game.DynamicAmountObjectPower &&
			dynamic.Val.Object != game.SourcePermanentReference() {
			t.Fatalf("%q source reference = %+v", test.text, dynamic.Val.Object)
		}
	}

}

func counterPlacementAmount(primitive game.Primitive) (game.Quantity, bool) {
	switch primitive.Kind() {
	case game.PrimitiveAddCounter:
		add, ok := primitive.(game.AddCounter)
		return add.Amount, ok
	case game.PrimitiveAddPlayerCounter:
		add, ok := primitive.(game.AddPlayerCounter)
		return add.Amount, ok
	default:
		return game.Quantity{}, false
	}
}

func TestRebaseAddPlayerCounterTargetReference(t *testing.T) {
	t.Parallel()
	primitive, ok := rebaseTargetedPrimitive(game.AddPlayerCounter{
		Amount:      game.Fixed(1),
		Player:      game.TargetPlayerReference(0),
		CounterKind: counter.Poison,
	}, 2)
	if !ok {
		t.Fatal("AddPlayerCounter target was not rebased")
	}
	add, ok := primitive.(game.AddPlayerCounter)
	if !ok || add.Player != game.TargetPlayerReference(2) {
		t.Fatalf("rebased primitive = %+v", primitive)
	}
}

func TestLowerCounterPlacementRejectsUnsupportedForms(t *testing.T) {
	t.Parallel()
	for _, oracleText := range []string{
		"Put a quest counter on target permanent.",
		"Put an energy counter on target creature.",
		"Put a charge counter on target player.",
		"Put a charge counter on any target.",
		"Put a +1/+1 counter on each creature you control.",
		"Put a charge and time counter on target artifact.",
		"Put 0 charge counters on target artifact.",
		"Put -1 charge counters on target artifact.",
		"Put a charge counter on target artifact for each land you control.",
	} {
		_, diagnostics := lowerExecutableFaces(&ScryfallCard{
			Name:       "Test Counter",
			Layout:     "normal",
			TypeLine:   "Sorcery",
			OracleText: oracleText,
		})
		if len(diagnostics) == 0 {
			t.Fatalf("%q lowered without diagnostics", oracleText)
		}
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

func TestLowerAtTriggerYourUpkeepDrawCard(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Bear",
		Layout:     "normal",
		TypeLine:   "Creature — Bear",
		OracleText: "At the beginning of your upkeep, draw a card.",
		Power:      new("2"),
		Toughness:  new("2"),
	})
	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("got %d triggered abilities, want 1", len(face.TriggeredAbilities))
	}
	ta := face.TriggeredAbilities[0]
	if ta.Trigger.Type != game.TriggerAt {
		t.Fatalf("trigger type = %v, want TriggerAt", ta.Trigger.Type)
	}
	if ta.Trigger.Pattern.Event != game.EventBeginningOfStep {
		t.Fatalf("event = %v, want EventBeginningOfStep", ta.Trigger.Pattern.Event)
	}
	if ta.Trigger.Pattern.Step != game.StepUpkeep {
		t.Fatalf("step = %v, want StepUpkeep", ta.Trigger.Pattern.Step)
	}
	if ta.Trigger.Pattern.Controller != game.TriggerControllerYou {
		t.Fatalf("controller = %v, want TriggerControllerYou", ta.Trigger.Pattern.Controller)
	}
	draw, ok := ta.Content.Modes[0].Sequence[0].Primitive.(game.Draw)
	if !ok || draw.Amount != game.Fixed(1) {
		t.Fatalf("primitive = %+v, want Draw{Amount: Fixed(1)}", ta.Content.Modes[0].Sequence[0].Primitive)
	}
}

func TestLowerAtTriggerEachOpponentUpkeepDamage(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Pinger",
		Layout:     "normal",
		TypeLine:   "Creature — Goblin",
		OracleText: "At the beginning of each opponent's upkeep, draw a card.",
		Power:      new("1"),
		Toughness:  new("1"),
	})
	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("got %d triggered abilities, want 1", len(face.TriggeredAbilities))
	}
	ta := face.TriggeredAbilities[0]
	if ta.Trigger.Type != game.TriggerAt {
		t.Fatalf("trigger type = %v, want TriggerAt", ta.Trigger.Type)
	}
	if ta.Trigger.Pattern.Event != game.EventBeginningOfStep {
		t.Fatalf("event = %v, want EventBeginningOfStep", ta.Trigger.Pattern.Event)
	}
	if ta.Trigger.Pattern.Step != game.StepUpkeep {
		t.Fatalf("step = %v, want StepUpkeep", ta.Trigger.Pattern.Step)
	}
	if ta.Trigger.Pattern.Controller != game.TriggerControllerOpponent {
		t.Fatalf("controller = %v, want TriggerControllerOpponent", ta.Trigger.Pattern.Controller)
	}
}

func TestLowerAtTriggerEachUpkeepAny(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Watcher",
		Layout:     "normal",
		TypeLine:   "Creature — Human",
		OracleText: "At the beginning of each upkeep, draw a card.",
		Power:      new("1"),
		Toughness:  new("1"),
	})
	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("got %d triggered abilities, want 1", len(face.TriggeredAbilities))
	}
	ta := face.TriggeredAbilities[0]
	if ta.Trigger.Pattern.Controller != game.TriggerControllerAny {
		t.Fatalf("controller = %v, want TriggerControllerAny", ta.Trigger.Pattern.Controller)
	}
}

func TestLowerAtTriggerYourEndStep(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Mystic",
		Layout:     "normal",
		TypeLine:   "Creature — Human Wizard",
		OracleText: "At the beginning of your end step, draw a card.",
		Power:      new("1"),
		Toughness:  new("1"),
	})
	ta := face.TriggeredAbilities[0]
	if ta.Trigger.Pattern.Step != game.StepEnd {
		t.Fatalf("step = %v, want StepEnd", ta.Trigger.Pattern.Step)
	}
	if ta.Trigger.Pattern.Controller != game.TriggerControllerYou {
		t.Fatalf("controller = %v, want TriggerControllerYou", ta.Trigger.Pattern.Controller)
	}
}

func TestLowerAtTriggerBeginningOfCombatYourTurn(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Fighter",
		Layout:     "normal",
		TypeLine:   "Creature — Human Warrior",
		OracleText: "At the beginning of combat on your turn, draw a card.",
		Power:      new("2"),
		Toughness:  new("2"),
	})
	ta := face.TriggeredAbilities[0]
	if ta.Trigger.Pattern.Step != game.StepBeginningOfCombat {
		t.Fatalf("step = %v, want StepBeginningOfCombat", ta.Trigger.Pattern.Step)
	}
	if ta.Trigger.Pattern.Controller != game.TriggerControllerYou {
		t.Fatalf("controller = %v, want TriggerControllerYou", ta.Trigger.Pattern.Controller)
	}
}

func TestLowerAtTriggerYourDrawStep(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Scholar",
		Layout:     "normal",
		TypeLine:   "Creature — Human Wizard",
		OracleText: "At the beginning of your draw step, draw a card.",
		Power:      new("1"),
		Toughness:  new("1"),
	})
	ta := face.TriggeredAbilities[0]
	if ta.Trigger.Pattern.Step != game.StepDraw {
		t.Fatalf("step = %v, want StepDraw", ta.Trigger.Pattern.Step)
	}
	if ta.Trigger.Pattern.Controller != game.TriggerControllerYou {
		t.Fatalf("controller = %v, want TriggerControllerYou", ta.Trigger.Pattern.Controller)
	}
}

func TestLowerAtTriggerEachCombat(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Battler",
		Layout:     "normal",
		TypeLine:   "Creature — Human Soldier",
		OracleText: "At the beginning of each combat, draw a card.",
		Power:      new("2"),
		Toughness:  new("2"),
	})
	ta := face.TriggeredAbilities[0]
	if ta.Trigger.Pattern.Step != game.StepBeginningOfCombat {
		t.Fatalf("step = %v, want StepBeginningOfCombat", ta.Trigger.Pattern.Step)
	}
	if ta.Trigger.Pattern.Controller != game.TriggerControllerAny {
		t.Fatalf("controller = %v, want TriggerControllerAny", ta.Trigger.Pattern.Controller)
	}
}

func TestLowerAtTriggerOptional(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Sage",
		Layout:     "normal",
		TypeLine:   "Creature — Human Wizard",
		OracleText: "At the beginning of your upkeep, you may draw a card.",
		Power:      new("1"),
		Toughness:  new("1"),
	})
	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("got %d triggered abilities, want 1", len(face.TriggeredAbilities))
	}
	ta := face.TriggeredAbilities[0]
	if !ta.Optional {
		t.Fatal("expected Optional = true for 'you may' trigger")
	}
	if ta.Trigger.Pattern.Step != game.StepUpkeep {
		t.Fatalf("step = %v, want StepUpkeep", ta.Trigger.Pattern.Step)
	}
}

func TestLowerAtTriggerPrecombatMainPhaseFailsClosed(t *testing.T) {
	t.Parallel()
	_, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Test Planeswalker",
		Layout:     "normal",
		TypeLine:   "Creature — Human",
		OracleText: "At the beginning of your precombat main phase, draw a card.",
		Power:      new("2"),
		Toughness:  new("2"),
	}, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) == 0 {
		t.Fatal("expected diagnostic for precombat main phase trigger, got none")
	}
	found := false
	for _, d := range diagnostics {
		if strings.Contains(d.Summary, "unsupported phase/step trigger phrase") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected 'unsupported phase/step trigger phrase' diagnostic, got: %v", diagnostics)
	}
}

func TestLowerAtTriggerInterveningIfFailsClosed(t *testing.T) {
	t.Parallel()
	_, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Test Bear",
		Layout:     "normal",
		TypeLine:   "Creature — Bear",
		OracleText: "At the beginning of your upkeep, if you control a creature, draw a card.",
		Power:      new("2"),
		Toughness:  new("2"),
	}, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) == 0 {
		t.Fatal("expected diagnostic for intervening-if on at-trigger, got none")
	}
}

func TestLowerAtTriggerPhraseVariants(t *testing.T) {
	t.Parallel()
	tests := []struct {
		phrase     string
		step       game.Step
		controller game.TriggerControllerFilter
	}{
		{"each upkeep", game.StepUpkeep, game.TriggerControllerAny},
		{"each player's upkeep", game.StepUpkeep, game.TriggerControllerAny},
		{"each opponent's upkeep", game.StepUpkeep, game.TriggerControllerOpponent},
		{"each end step", game.StepEnd, game.TriggerControllerAny},
		{"each player's end step", game.StepEnd, game.TriggerControllerAny},
		{"each combat", game.StepBeginningOfCombat, game.TriggerControllerAny},
	}
	for _, tc := range tests {
		t.Run(tc.phrase, func(t *testing.T) {
			t.Parallel()
			face := lowerSingleFace(t, &ScryfallCard{
				Name:       "Test Card",
				Layout:     "normal",
				TypeLine:   "Creature — Human",
				OracleText: "At the beginning of " + tc.phrase + ", draw a card.",
				Power:      new("1"),
				Toughness:  new("1"),
			})
			if len(face.TriggeredAbilities) != 1 {
				t.Fatalf("got %d triggered abilities, want 1", len(face.TriggeredAbilities))
			}
			ta := face.TriggeredAbilities[0]
			if ta.Trigger.Pattern.Step != tc.step {
				t.Errorf("step = %v, want %v", ta.Trigger.Pattern.Step, tc.step)
			}
			if ta.Trigger.Pattern.Controller != tc.controller {
				t.Errorf("controller = %v, want %v", ta.Trigger.Pattern.Controller, tc.controller)
			}
		})
	}
}
