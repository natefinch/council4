package cardgen

import (
	"go/parser"
	"go/token"
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
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
