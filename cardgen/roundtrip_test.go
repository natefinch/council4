package cardgen

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// roundTripCards are representative cards exercising the mana, static, and spell
// ability categories through the full typed pipeline.
var roundTripCards = []*ScryfallCard{
	{
		Name: "Ashling", Layout: "normal", TypeLine: "Legendary Creature - Elemental Shaman",
		OracleText: ashlingRemovedCounterText, Power: new("1"), Toughness: new("1"),
	},
	{
		Name: "RT Counter Quantity", Layout: "normal", TypeLine: "Artifact",
		OracleText: "{1}: Remove all charge counters from this artifact. Draw that many cards.",
	},
	{
		Name: "RT Counter Quantity Spell", Layout: "normal", TypeLine: "Sorcery",
		OracleText: "Remove all +1/+1 counters from target creature. Draw that many cards.",
	},
	{
		Name: "RT Reached Target Card", Layout: "normal", TypeLine: "Creature",
		OracleText: "{1}: Exile target card from a graveyard. If it was a permanent card, put a +1/+1 counter on this creature.",
		Power:      new("2"), Toughness: new("2"),
	},
	{
		Name: "RT Countered Spell Condition", Layout: "normal", TypeLine: "Instant",
		OracleText: "Counter target spell. If that spell's mana value was 3 or less, proliferate.",
	},
	{
		Name: "RT Event Spell Condition", Layout: "normal", TypeLine: "Creature",
		OracleText: "Whenever you cast an instant or sorcery spell, you gain 1 life. If that spell has mana value 5 or greater, put a +1/+1 counter on this creature.",
		Power:      new("2"), Toughness: new("2"),
	},
	{
		Name: "RT Spell After Return", Layout: "normal", TypeLine: "Creature",
		OracleText: "Whenever you cast an instant or sorcery spell, return target creature card from your graveyard to the battlefield. You gain 1 life. If that spell has mana value 5 or greater, you gain 2 life.",
		Power:      new("2"), Toughness: new("2"),
	},
	{
		Name: "RT Numeric Unless", Layout: "normal", TypeLine: "Creature",
		OracleText: "When this creature enters, draw a card. Then discard a card unless you control another creature with power 2 or less.",
		Power:      new("2"), Toughness: new("2"),
	},
	{
		Name:       "RT Bear",
		Layout:     "normal",
		TypeLine:   "Creature — Bear",
		ManaCost:   "{1}{G}",
		Colors:     []string{"G"},
		OracleText: "Flying\nVigilance",
		Power:      new("2"),
		Toughness:  new("2"),
	},
	{
		Name:       "RT Land",
		Layout:     "normal",
		TypeLine:   "Land",
		OracleText: "{T}: Add {G}.",
	},
	{
		Name:       "RT Bolt",
		Layout:     "normal",
		TypeLine:   "Instant",
		ManaCost:   "{R}",
		Colors:     []string{"R"},
		OracleText: "RT Bolt deals 3 damage to any target.",
	},
	{
		Name: "RT Condition Group", Layout: "normal", TypeLine: "Instant",
		OracleText: "If you have no cards in hand, draw a card, then draw a card.",
	},
	{
		Name: "RT Condition Ability", Layout: "normal", TypeLine: "Enchantment",
		OracleText: "{1}: If you have no cards in hand, draw a card, then draw a card.",
	},
	{
		Name: "RT Returned Subject", Layout: "normal", TypeLine: "Sorcery",
		OracleText: "Return target creature card from your graveyard to the battlefield. If it's an Elf, put a +1/+1 counter on it.",
	},
	{
		Name: "RT Blinked Subject", Layout: "normal", TypeLine: "Instant",
		OracleText: "Exile target creature you control, then return it to the battlefield under its owner's control. If that creature is a Bird, Frog, Otter, or Rat, draw a card.",
	},
	{
		Name: "RT Modal Subject", Layout: "normal", TypeLine: "Instant",
		OracleText: "Choose one —\n• Exile target creature you control, then return it to the battlefield tapped under its owner's control. If it's an Elf, untap it.\n• Draw a card.",
	},
	{
		Name: "RT Ordinal Modes", Layout: "normal", TypeLine: "Artifact",
		OracleText: "{1}: Choose one \u2014\n\u2022 If this is the second time this ability has resolved this turn, draw a card, then draw a card.\n\u2022 You gain 2 life.",
	},
	{
		Name: "RT Result Count", Layout: "normal", TypeLine: "Sorcery",
		OracleText: "Discard three cards. If you discarded two land cards this way, draw a card. Otherwise, you gain 2 life.",
	},
	{
		Name: "RT Result Ability", Layout: "normal", TypeLine: "Enchantment",
		OracleText: "{1}: Discard three cards. If two land cards were discarded this way, draw a card.",
	},
	{
		Name: "RT Ordinal Mana", Layout: "normal", TypeLine: "Artifact",
		OracleText: flamekinBody,
	},
	{
		Name: "RT Gated Mana", Layout: "normal", TypeLine: "Instant",
		OracleText: "If you have no cards in hand, draw a card, then add {R}{G}.",
	},
	{
		Name: "RT Sacrifice Cost Subject", Layout: "normal", TypeLine: "Artifact",
		OracleText: "Sacrifice a creature: You gain 1 life. If the sacrificed creature was red, draw a card.",
	},
	{
		Name: "RT Discard Cost Subject", Layout: "normal", TypeLine: "Sorcery",
		OracleText: "As an additional cost to cast this spell, discard a card.\nDraw a card. If the discarded card wasn't a land card, you gain 2 life.",
	},
	{
		Name: "RT Numeric Cost Subject", Layout: "normal", TypeLine: "Artifact",
		OracleText: "{T}, Sacrifice a creature: Create a Food token. If the sacrificed creature had toughness 4 or greater, create two Food tokens instead.",
	},
	{
		Name:       "RT Bog",
		Layout:     "normal",
		TypeLine:   "Land",
		OracleText: "When this land enters, exile target player's graveyard.",
	},
	{
		Name:       "RT Ozolith",
		Layout:     "normal",
		TypeLine:   "Legendary Artifact",
		OracleText: "Whenever a creature you control leaves the battlefield, if it had counters on it, put those counters on RT Ozolith.\nAt the beginning of combat on your turn, if RT Ozolith has counters on it, you may move all counters from RT Ozolith onto target creature.",
	},
	{
		Name:       "RT Nesting Grounds",
		Layout:     "normal",
		TypeLine:   "Land",
		OracleText: "{T}: Add {C}.\n{1}, {T}: Move a counter from target permanent you control onto a second target permanent. Activate only as a sorcery.",
	},
	{
		Name:       "RT Reaver Cleaver",
		Layout:     "normal",
		TypeLine:   "Legendary Artifact — Equipment",
		OracleText: "Equipped creature gets +1/+1 and has trample and \"Whenever this creature deals combat damage to a player or planeswalker, create that many Treasure tokens.\"\nEquip {3}",
	},
	{
		Name:       "RT Trailblazer's Boots",
		Layout:     "normal",
		TypeLine:   "Artifact — Equipment",
		OracleText: "Equipped creature has nonbasic landwalk. (It can't be blocked as long as defending player controls a nonbasic land.)\nEquip {2}",
	},
	{
		// Exercises MassReturnFromGraveyard, whose rendered Destination zone
		// literal requires the zone import (regression guard for #995).
		Name:       "RT Replenish",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		ManaCost:   "{3}{W}",
		Colors:     []string{"W"},
		OracleText: "Return all enchantment cards from your graveyard to the battlefield.",
	},
	{
		Name:       "RT Brotherhood Regalia",
		Layout:     "normal",
		TypeLine:   "Artifact — Equipment",
		OracleText: "Equipped creature has ward {2}, is an Assassin in addition to its other types, and can't be blocked.\nEquip legendary creature {1}\nEquip {3}",
	},
	{
		// Exercises the noncreature-exclusion spell cost modifier (chapter II)
		// and the reanimation counter-kind choice (chapter III) introduced for
		// Elspeth Conquers Death (#1500). The chapter III AddCounter renders a
		// KindChoices literal on a linked object, requiring the counter import.
		Name:     "RT Elspeth Conquers Death",
		Layout:   "saga",
		TypeLine: "Enchantment — Saga",
		ManaCost: "{3}{W}{W}",
		OracleText: "(As this Saga enters and after your draw step, add a lore counter. Sacrifice after III.)\n" +
			"I — Exile target permanent an opponent controls with mana value 3 or greater.\n" +
			"II — Noncreature spells your opponents cast cost {2} more to cast until your next turn.\n" +
			"III — Return target creature or planeswalker card from your graveyard to the battlefield. Put a +1/+1 counter or a loyalty counter on it.",
	},
}

// writeRoundTripPackage generates source for roundTripCards into a fresh package
// directory inside the module and returns the directory and package name.
func writeRoundTripPackage(t *testing.T) (dir, pkgName string) {
	t.Helper()
	suffix := filepath.Base(t.TempDir())
	dir = filepath.Join(".", "roundtrippkg"+suffix)
	if err := os.Mkdir(dir, 0o700); err != nil {
		t.Fatalf("Mkdir: %v", err)
	}
	t.Cleanup(func() {
		_ = os.RemoveAll(dir)
	})

	pkgName = filepath.Base(dir)
	for _, card := range roundTripCards {
		source, diagnostics, err := GenerateExecutableCardSource(card, pkgName)
		if err != nil {
			t.Fatalf("GenerateExecutableCardSource(%q): %v", card.Name, err)
		}
		if len(diagnostics) != 0 {
			t.Fatalf("GenerateExecutableCardSource(%q) diagnostics: %#v", card.Name, diagnostics)
		}
		file := filepath.Join(dir, CardNameToVarName(card.Name)+".go")
		if err := os.WriteFile(file, []byte(source), 0o600); err != nil {
			t.Fatalf("WriteFile: %v", err)
		}
	}
	return dir, pkgName
}

// TestRoundTripCompiles generates executable card source, writes it to a fresh
// package directory inside the module, and runs `go build` to verify the
// rendered output is valid, compilable Go.
func TestRoundTripCompiles(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping go build round-trip in short mode")
	}

	dir, _ := writeRoundTripPackage(t)
	cmd := exec.CommandContext(context.Background(), "go", "build", "./"+filepath.Base(dir))
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("go build failed: %v\n%s", err, out)
	}
}

// TestRoundTripSemantic generates executable card source, writes a semantic test
// alongside it in the same package, and runs `go test` so the generated vars are
// checked for the actual typed structure they must round-trip to — not merely
// that they compile.
func TestRoundTripSemantic(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping go test round-trip in short mode")
	}

	dir, pkgName := writeRoundTripPackage(t)
	testFile := filepath.Join(dir, "semantic_test.go")
	if err := os.WriteFile(testFile, []byte(semanticTestSource(pkgName)), 0o600); err != nil {
		t.Fatalf("WriteFile semantic test: %v", err)
	}

	cmd := exec.CommandContext(context.Background(), "go", "test", "-count=1", "./"+filepath.Base(dir))
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("go test failed: %v\n%s", err, out)
	}
}

// semanticTestSource returns the source of a test file, in package pkgName, that
// directly inspects the generated vars to confirm they round-trip to the
// expected typed structure.
func semanticTestSource(pkgName string) string {
	return fmt.Sprintf(`package %s

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
)

func TestRTRemovedCounterQuantitySemantic(t *testing.T) {
	seq := RTCounterQuantity().ActivatedAbilities[0].Content.Modes[0].Sequence
	remove := seq[0].Primitive.(game.RemoveCounter)
	draw := seq[1].Primitive.(game.Draw)
	if remove.AllKinds || remove.CounterKind != counter.Charge ||
		remove.Amount.DynamicAmount().Val.Kind != game.DynamicAmountObjectCounters ||
		seq[0].PublishResult == "" || !seq[1].ResultGate.Val.AmountAvailable ||
		draw.Amount.DynamicAmount().Val.ResultKey != seq[0].PublishResult {
		t.Fatal("named-kind actual quantity publication did not round-trip")
	}
	seq = RTCounterQuantitySpell().SpellAbility.Val.Modes[0].Sequence
	remove = seq[0].Primitive.(game.RemoveCounter)
	draw = seq[1].Primitive.(game.Draw)
	if remove.AllKinds || remove.CounterKind != counter.PlusOnePlusOne ||
		!seq[1].ResultGate.Val.AmountAvailable ||
		draw.Amount.DynamicAmount().Val.ResultKey != seq[0].PublishResult {
		t.Fatal("spell scalar availability did not round-trip")
	}
	ability := Ashling().ActivatedAbilities[0]
	seq = ability.Content.Modes[0].Sequence
	remove = seq[1].Primitive.(game.RemoveCounter)
	if !ability.CountsResolutionsThisTurn || len(seq) != 4 ||
		remove.AllKinds || remove.CounterKind != counter.PlusOnePlusOne ||
		seq[1].Condition.Val.Condition.Val.SourceAbilityResolutionOrdinalThisTurn != 3 ||
		seq[1].PublishCondition == "" {
		t.Fatal("ordinal named-kind removal did not round-trip")
	}
	for _, instruction := range seq[2:] {
		damage := instruction.Primitive.(game.Damage)
		if instruction.ConditionGate != seq[1].PublishCondition ||
			!instruction.ResultGate.Val.AmountAvailable ||
			instruction.ResultGate.Val.Key != seq[1].PublishResult ||
			damage.Amount.DynamicAmount().Val.ResultKey != seq[1].PublishResult {
			t.Fatal("expanded damage scalar and condition identities did not round-trip")
		}
	}
}

func TestRTNumericConditionSemantic(t *testing.T) {
	past := RTCounteredSpellCondition().SpellAbility.Val.Modes[0].Sequence[1].Condition.Val.Condition.Val
	if !past.UseCounteredSpellManaValue ||
		past.Object.Val.Kind() != game.ObjectReferenceTargetStackObject ||
		past.ObjectMatches.Val.ManaValue.Val.Op != compare.LessOrEqual ||
		past.ObjectMatches.Val.ManaValue.Val.Value != 3 {
		t.Fatal("past target-spell information did not round-trip")
	}
	event := RTEventSpellCondition().TriggeredAbilities[0].Content.Modes[0].Sequence[1].Condition.Val.Condition.Val
	if event.UseCounteredSpellManaValue || event.Object.Val.Kind() != game.ObjectReferenceEventStackObject ||
		event.ObjectMatches.Val.ManaValue.Val.Op != compare.GreaterOrEqual ||
		event.ObjectMatches.Val.ManaValue.Val.Value != 5 {
		t.Fatal("exact event-spell numeric reference did not round-trip")
	}
	returnedCompetitor := RTSpellAfterReturn().TriggeredAbilities[0].Content.Modes[0].Sequence[2].Condition.Val.Condition.Val
	if returnedCompetitor.Object.Val.Kind() != game.ObjectReferenceEventStackObject ||
		returnedCompetitor.ObjectMatches.Val.ManaValue.Val.Value != 5 {
		t.Fatal("returned permanent captured the event-spell condition")
	}
	unless := RTNumericUnless().TriggeredAbilities[0].Content.Modes[0].Sequence[1].Condition.Val.Condition.Val
	if !unless.Negate || !unless.ControlsMatching.Val.Selection.ExcludeSource ||
		unless.ControlsMatching.Val.Selection.Power.Val.Op != compare.LessOrEqual ||
		unless.ControlsMatching.Val.Selection.Power.Val.Value != 2 {
		t.Fatal("source-excluding numeric Unless did not round-trip")
	}
	seq := RTReachedTargetCard().ActivatedAbilities[0].Content.Modes[0].Sequence
	card := seq[1].Condition.Val.Condition.Val
	if card.Object.Val.Kind() != game.ObjectReferenceTargetCard ||
		card.TargetCardResultKey == "" || card.TargetCardResultKey != seq[0].PublishResult {
		t.Fatal("actual reached-card publication did not round-trip")
	}
}

func TestRTLandSemantic(t *testing.T) {
	if RTLand().CardFace.Name != "RT Land" {
		t.Fatalf("name = %%q", RTLand().CardFace.Name)
	}
	if len(RTLand().CardFace.ManaAbilities) != 1 {
		t.Fatalf("mana abilities = %%d", len(RTLand().CardFace.ManaAbilities))
	}
	prim := RTLand().CardFace.ManaAbilities[0].Content.Modes[0].Sequence[0].Primitive
	add, ok := prim.(game.AddMana)
	if !ok {
		t.Fatalf("primitive type = %%T", prim)
	}
	if add.ManaColor != mana.G {
		t.Fatalf("mana color = %%q", add.ManaColor)
	}
}

func TestRTPublishedSubjectSemantic(t *testing.T) {
	returned := RTReturnedSubject().SpellAbility.Val.Modes[0].Sequence
	put := returned[0].Primitive.(game.PutOnBattlefield)
	want := game.LinkedObjectReference(string(put.PublishLinked))
	if put.PublishLinked == "" || returned[1].Condition.Val.Condition.Val.Object.Val != want ||
		returned[1].Primitive.(game.AddCounter).Object != want {
		t.Fatal("returned subject publication did not round-trip")
	}
	blinked := RTBlinkedSubject().SpellAbility.Val.Modes[0].Sequence
	put = blinked[1].Primitive.(game.PutOnBattlefield)
	input, ok := put.Source.LinkedKey()
	condition := blinked[2].Condition.Val.Condition.Val
	if !ok || put.PublishLinked == "" || input == put.PublishLinked ||
		condition.Object.Val != game.LinkedObjectReference(string(put.PublishLinked)) ||
		len(condition.ObjectMatches.Val.SubtypesAny) != 4 {
		t.Fatal("blink input, output or full subject selection did not round-trip")
	}
	modal := RTModalSubject().SpellAbility.Val.Modes[0].Sequence
	put = modal[1].Primitive.(game.PutOnBattlefield)
	want = game.LinkedObjectReference(string(put.PublishLinked))
	if put.PublishLinked == "" || modal[2].Primitive.(game.Untap).Object != want ||
		modal[2].Condition.Val.Condition.Val.Object.Val != want {
		t.Fatal("modal subject and consequence identity did not round-trip")
	}
}

func TestRTBearSemantic(t *testing.T) {
	if len(RTBear().CardFace.StaticAbilities) != 2 {
		t.Fatalf("static abilities = %%d", len(RTBear().CardFace.StaticAbilities))
	}
	keywords := RTBear().CardFace.StaticAbilities[0].KeywordAbilities
	if len(keywords) != 1 {
		t.Fatalf("keyword abilities = %%d", len(keywords))
	}
	keyword, ok := keywords[0].(game.SimpleKeyword)
	if !ok || keyword.Kind != game.Flying {
		t.Fatalf("keyword[0] = %%#v", keywords[0])
	}
}

func TestRTBoltSemantic(t *testing.T) {
	if !RTBolt().CardFace.SpellAbility.Exists {
		t.Fatal("spell ability missing")
	}
	mode := RTBolt().CardFace.SpellAbility.Val.Modes[0]
	if len(mode.Targets) != 1 {
		t.Fatalf("targets = %%d", len(mode.Targets))
	}
	damage, ok := mode.Sequence[0].Primitive.(game.Damage)
	if !ok {
		t.Fatalf("primitive type = %%T", mode.Sequence[0].Primitive)
	}
	if damage.Amount.Value() != 3 {
		t.Fatalf("damage amount = %%d", damage.Amount.Value())
	}
}

func TestRTConditionGroupSemantic(t *testing.T) {
	seq := RTConditionGroup().SpellAbility.Val.Modes[0].Sequence
	if len(seq) != 2 || !seq[0].Condition.Exists ||
		!seq[0].Condition.Val.Condition.Val.ControllerHandEmpty ||
		seq[0].PublishCondition == "" ||
		seq[1].Condition.Exists || seq[1].ConditionGate != seq[0].PublishCondition {
		t.Fatal("scoped condition evaluation did not round-trip")
	}
}

func TestRTConditionAbilitySemantic(t *testing.T) {
	seq := RTConditionAbility().ActivatedAbilities[0].Content.Modes[0].Sequence
	if len(seq) != 2 || !seq[0].Condition.Exists ||
		seq[0].PublishCondition == "" ||
		seq[1].Condition.Exists || seq[1].ConditionGate != seq[0].PublishCondition {
		t.Fatal("nested scoped condition evaluation did not round-trip")
	}
}

func TestRTOrdinalModesSemantic(t *testing.T) {
	ability := RTOrdinalModes().ActivatedAbilities[0]
	if !ability.CountsResolutionsThisTurn || ability.ActivationCondition.Exists ||
		len(ability.Content.Modes) != 2 {
		t.Fatal("resolution-count shell metadata did not round-trip")
	}
	seq := ability.Content.Modes[0].Sequence
	if len(seq) != 2 || !seq[0].Condition.Exists ||
		seq[0].Condition.Val.Condition.Val.SourceAbilityResolutionOrdinalThisTurn != 2 ||
		seq[0].PublishCondition == "" || seq[1].ConditionGate != seq[0].PublishCondition {
		t.Fatal("modal resolution ordinal and grouped consumer did not round-trip")
	}
}

func TestRTResultCountSemantic(t *testing.T) {
	for _, seq := range [][]game.Instruction{
		RTResultCount().SpellAbility.Val.Modes[0].Sequence,
		RTResultAbility().ActivatedAbilities[0].Content.Modes[0].Sequence,
	} {
		gate := seq[1].ResultGate.Val
		if seq[0].PublishResult == "" || gate.Key != seq[0].PublishResult ||
			!gate.ObjectSelection.Exists || !gate.CardOnly || !gate.ObjectCountRange.Exists ||
			gate.ObjectCountRange.Val.Min != 2 || gate.ObjectCountRange.Val.Max != 2 {
			t.Fatal("actual matching-result count did not round-trip")
		}
	}
	complement := RTResultCount().SpellAbility.Val.Modes[0].Sequence[2].ResultGate.Val
	if !complement.Negate || complement.ObjectCountRange.Val.Min != 2 {
		t.Fatal("count predicate complement did not round-trip")
	}
}

func TestRTOrdinaryManaSemantic(t *testing.T) {
	ability := RTOrdinalMana().ActivatedAbilities[0]
	seq := ability.Content.Modes[0].Sequence
	if !ability.CountsResolutionsThisTurn || ability.ActivationCondition.Exists ||
		len(RTOrdinalMana().ManaAbilities) != 0 || len(seq) != 2 || seq[0].Condition.Exists {
		t.Fatal("ordinary fixed-mana shell did not round-trip")
	}
	add, ok := seq[1].Primitive.(game.AddMana)
	if !ok || add.Amount.Value() != 8 || add.ManaColor != mana.R || !seq[1].Optional ||
		seq[1].Condition.Val.Condition.Val.SourceAbilityResolutionOrdinalThisTurn != 3 {
		t.Fatal("one optional fixed-mana action did not round-trip")
	}
	group := RTGatedMana().SpellAbility.Val.Modes[0].Sequence
	if len(group) != 3 || group[0].PublishCondition == "" ||
		group[1].ConditionGate != group[0].PublishCondition || group[2].ConditionGate != group[0].PublishCondition {
		t.Fatal("expanded fixed-mana group decision did not round-trip")
	}
	for i, color := range []mana.Color{mana.R, mana.G} {
		add, ok := group[i+1].Primitive.(game.AddMana)
		if !ok || add.Amount.Value() != 1 || add.ManaColor != color {
			t.Fatal("fixed-mana color or amount did not round-trip")
		}
	}
}

func TestRTPaidCostSubjectSemantic(t *testing.T) {
	ability := RTSacrificeCostSubject().ActivatedAbilities[0]
	key := ability.AdditionalCosts[0].SubjectKey
	ref := ability.Content.Modes[0].Sequence[1].Condition.Val.Condition.Val.Object.Val
	if key == "" || ref != game.PaidCostReference(key, game.PaidCostSacrifice, types.Creature) {
		t.Fatal("sacrifice component identity, domain and noun did not round-trip")
	}
	card := RTDiscardCostSubject()
	key = card.AdditionalCosts[0].SubjectKey
	condition := card.SpellAbility.Val.Modes[0].Sequence[1].Condition.Val.Condition.Val
	if key == "" || condition.Object.Val != game.PaidCostReference(key, game.PaidCostDiscard, "") || !condition.Negate {
		t.Fatalf("discard cost key=%%q reference key=%%q negated=%%v", key, condition.Object.Val.CostKey(), condition.Negate)
	}
	numeric := RTNumericCostSubject().ActivatedAbilities[0]
	key = numeric.AdditionalCosts[1].SubjectKey
	sequence := numeric.Content.Modes[0].Sequence
	condition = sequence[0].Condition.Val.Condition.Val
	if key == "" || condition.Object.Val != game.PaidCostReference(key, game.PaidCostSacrifice, types.Creature) ||
		condition.ObjectMatches.Val.Toughness.Val != (compare.Int{Op: compare.GreaterOrEqual, Value: 4}) ||
		!condition.Negate || sequence[0].PublishCondition == "" ||
		sequence[1].ConditionGate != sequence[0].PublishCondition || !sequence[1].ConditionGateNegate {
		t.Fatal("numeric paid-cost predicate and exclusive replacement envelope did not round-trip")
	}
}
`, pkgName)
}
