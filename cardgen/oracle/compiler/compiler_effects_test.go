package compiler

import (
	"testing"

	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

func TestCompileModalAbility(t *testing.T) {
	t.Parallel()
	source := "Choose one —\n• Destroy target creature.\n• Draw two cards."
	compilation, diagnostics := compileSource(source, pipelineContext{InstantOrSorcery: true})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	ability := compilation.Abilities[0]
	if len(ability.Content.Modes) != 2 {
		t.Fatalf("modes = %#v", ability.Content.Modes)
	}
	if ability.Content.Modes[0].Content.Effects[0].Kind != EffectDestroy ||
		len(ability.Content.Modes[0].Content.Targets) != 1 ||
		ability.Content.Modes[1].Content.Effects[0].Kind != EffectDraw {
		t.Fatalf("compiled modes = %#v", ability.Content.Modes)
	}
}

func TestCompileKeywordsAndReminder(t *testing.T) {
	t.Parallel()
	source := "First strike (This creature deals combat damage before other creatures.)\nEquip {2} ({2}: Attach to target creature you control.)"
	compilation, diagnostics := compileSource(source, pipelineContext{})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	if got := compilation.Abilities[0].Content.Keywords; len(got) != 1 || got[0].Name != "First strike" {
		t.Fatalf("first strike = %#v", got)
	}
	equip := compilation.Abilities[1]
	if len(equip.Content.Keywords) != 1 || equip.Content.Keywords[0].Name != "Equip" ||
		equip.Content.Keywords[0].Parameter != "{2}" {
		t.Fatalf("equip = %#v", equip.Content.Keywords)
	}
	if len(equip.Content.Effects) != 0 || len(equip.Content.Targets) != 0 {
		t.Fatalf("reminder text leaked semantics: %#v", equip)
	}
}

func TestCompileDevoidAndReminder(t *testing.T) {
	t.Parallel()
	source := "Devoid (This card has no color.)"
	compilation, diagnostics := compileSource(source, pipelineContext{})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	ability := compilation.Abilities[0]
	if len(ability.Content.Keywords) != 1 ||
		ability.Content.Keywords[0].Name != "Devoid" ||
		ability.Content.Keywords[0].Text != "Devoid" {
		t.Fatalf("keywords = %#v", ability.Content.Keywords)
	}
	if len(ability.Content.Effects) != 0 || len(ability.Content.References) != 0 {
		t.Fatalf("reminder text leaked semantics: %#v", ability)
	}
}

func TestCompileReadAheadAndReminder(t *testing.T) {
	t.Parallel()
	source := "Read ahead (Choose a chapter and start with that many lore counters. Add one after your draw step. Skipped chapters don't trigger.)"
	compilation, diagnostics := compileSource(source, pipelineContext{})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	if len(compilation.Abilities) != 1 {
		t.Fatalf("abilities = %#v, want one", compilation.Abilities)
	}
	ability := compilation.Abilities[0]
	if len(ability.Content.Keywords) != 1 ||
		ability.Content.Keywords[0].Name != "Read ahead" ||
		ability.Content.Keywords[0].Text != "Read ahead" {
		t.Fatalf("keywords = %#v", ability.Content.Keywords)
	}
	if len(ability.Content.Effects) != 0 || len(ability.Content.References) != 0 {
		t.Fatalf("reminder text leaked semantics: %#v", ability)
	}
}

func TestCompileEnchantKeywordParameter(t *testing.T) {
	t.Parallel()
	compilation, diagnostics := compileSource("Enchant creature", pipelineContext{})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	keywords := compilation.Abilities[0].Content.Keywords
	if len(keywords) != 1 {
		t.Fatalf("keywords = %#v", keywords)
	}
	if keywords[0].Name != "Enchant" ||
		keywords[0].Parameter != "creature" ||
		keywords[0].Text != "Enchant creature" ||
		keywords[0].Span.Start.Offset != 0 ||
		keywords[0].Span.End.Offset != len("Enchant creature") {
		t.Fatalf("enchant keyword = %#v", keywords[0])
	}
}

func TestCompileProtectionKeywordParameter(t *testing.T) {
	t.Parallel()
	compilation, diagnostics := compileSource("Protection from red", pipelineContext{})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	keywords := compilation.Abilities[0].Content.Keywords
	if len(keywords) != 1 {
		t.Fatalf("keywords = %#v", keywords)
	}
	if keywords[0].Name != "Protection" ||
		keywords[0].Parameter != "red" ||
		keywords[0].Text != "Protection from red" ||
		keywords[0].Span.Start.Offset != 0 ||
		keywords[0].Span.End.Offset != len("Protection from red") {
		t.Fatalf("protection keyword = %#v", keywords[0])
	}
}

func TestCompileProtectionKeywordMultipleColors(t *testing.T) {
	t.Parallel()
	compilation, diagnostics := compileSource("Protection from black and from red", pipelineContext{})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	keywords := compilation.Abilities[0].Content.Keywords
	if len(keywords) != 1 ||
		keywords[0].Parameter != "black,red" ||
		keywords[0].Text != "Protection from black and from red" ||
		keywords[0].Span.End.Offset != len("Protection from black and from red") {
		t.Fatalf("protection keyword = %#v", keywords)
	}
}

func TestCompileTargetsAndReferences(t *testing.T) {
	t.Parallel()
	source := "Legolas deals damage to up to one target creature you don't control. It gains trample until end of turn."
	compilation, diagnostics := compileSource(source, pipelineContext{
		CardName: "Legolas",
	})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	ability := compilation.Abilities[0]
	if len(ability.Content.Targets) != 1 ||
		ability.Content.Targets[0].Cardinality != (TargetCardinality{Min: 0, Max: 1}) ||
		ability.Content.Targets[0].Selector.Controller != ControllerNotYou {
		t.Fatalf("targets = %#v", ability.Content.Targets)
	}
	if len(ability.Content.References) != 2 ||
		ability.Content.References[0].Kind != ReferenceSelfName ||
		ability.Content.References[1].Kind != ReferencePronoun {
		t.Fatalf("references = %#v", ability.Content.References)
	}
}

func TestCompileControlledPermanentTemporaryKeywordGrant(t *testing.T) {
	t.Parallel()
	compilation, diagnostics := compileSource(
		"Permanents you control gain hexproof and indestructible until end of turn.",
		pipelineContext{InstantOrSorcery: true},
	)
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	content := compilation.Abilities[0].Content
	if len(content.Effects) != 1 {
		t.Fatalf("effects = %#v, want one", content.Effects)
	}
	effect := content.Effects[0]
	if effect.Kind != EffectGain ||
		effect.StaticSubject != StaticSubjectControlledPermanents ||
		effect.Duration != DurationUntilEndOfTurn ||
		!effect.Exact {
		t.Fatalf("effect = %#v, want exact controlled-permanent keyword grant", effect)
	}
	if len(content.Keywords) != 2 ||
		content.Keywords[0].Kind != parser.KeywordHexproof ||
		content.Keywords[1].Kind != parser.KeywordIndestructible {
		t.Fatalf("keywords = %#v, want hexproof and indestructible", content.Keywords)
	}
}

func TestCompileExactTargetCardinalityAndPluralSelector(t *testing.T) {
	t.Parallel()
	compilation, diagnostics := compileSource("Tap two target creatures.", pipelineContext{InstantOrSorcery: true})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	targets := compilation.Abilities[0].Content.Targets
	if len(targets) != 1 ||
		targets[0].Cardinality != (TargetCardinality{Min: 2, Max: 2}) ||
		targets[0].Selector.Kind != SelectorCreature {
		t.Fatalf("targets = %#v", targets)
	}
}

func TestCompileThirdPersonEffects(t *testing.T) {
	t.Parallel()
	tests := map[string]EffectKind{
		"Each opponent discards a card.":        EffectDiscard,
		"Target player draws two cards.":        EffectDraw,
		"Its controller sacrifices a creature.": EffectSacrifice,
		"That player searches their library.":   EffectSearch,
		"That creature transforms.":             EffectTransform,
	}
	for source, want := range tests {
		t.Run(source, func(t *testing.T) {
			t.Parallel()
			compilation, diagnostics := compileSource(source, pipelineContext{InstantOrSorcery: true})
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			effects := compilation.Abilities[0].Content.Effects
			if len(effects) != 1 || effects[0].Kind != want {
				t.Fatalf("effects = %#v, want %v", effects, want)
			}
		})
	}
}

func TestCompileFixedEffectValues(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		context pipelineContext
		kind    EffectKind
		amount  int
		symbol  string
	}{
		"Draw two cards.": {
			context: pipelineContext{InstantOrSorcery: true},
			kind:    EffectDraw,
			amount:  2,
		},
		"Shock deals 3 damage to any target.": {
			context: pipelineContext{CardName: "Shock", InstantOrSorcery: true},
			kind:    EffectDealDamage,
			amount:  3,
		},
		"{T}: Add {G}.": {
			kind:   EffectAddMana,
			amount: 1,
			symbol: "{G}",
		},
	}

	for source, test := range tests {
		t.Run(source, func(t *testing.T) {
			t.Parallel()
			compilation, diagnostics := compileSource(source, test.context)
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			effects := compilation.Abilities[0].Content.Effects
			if len(effects) != 1 {
				t.Fatalf("effects = %#v", effects)
			}
			effect := effects[0]
			if effect.Kind != test.kind ||
				!effect.Amount.Known ||
				effect.Amount.Value != test.amount ||
				effect.Symbol() != test.symbol {
				t.Fatalf("effect = %#v", effect)
			}
		})
	}
}

func TestCompileDelayedEffectTiming(t *testing.T) {
	t.Parallel()
	tests := []struct {
		source string
		kind   EffectKind
		timing game.DelayedTriggerTiming
	}{
		{
			source: "Draw a card at the beginning of the next turn's upkeep.",
			kind:   EffectDraw,
			timing: game.DelayedAtBeginningOfNextUpkeep,
		},
		{
			source: "Exile it at the beginning of the next end step.",
			kind:   EffectExile,
			timing: game.DelayedAtBeginningOfNextEndStep,
		},
		{
			source: "Sacrifice it at the beginning of the next end step.",
			kind:   EffectSacrifice,
			timing: game.DelayedAtBeginningOfNextEndStep,
		},
		{
			source: "Return it to its owner's hand at the beginning of the next end step.",
			kind:   EffectReturn,
			timing: game.DelayedAtBeginningOfNextEndStep,
		},
	}

	for _, tt := range tests {
		t.Run(tt.source, func(t *testing.T) {
			t.Parallel()
			compilation, diagnostics := compileSource(tt.source, pipelineContext{InstantOrSorcery: true})
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			effects := compilation.Abilities[0].Content.Effects
			if len(effects) != 1 || effects[0].Kind != tt.kind || effects[0].DelayedTiming != tt.timing {
				t.Fatalf("effects = %#v, want kind %v timing %v", effects, tt.kind, tt.timing)
			}
		})
	}
}

func TestCompileArcaneDenialDelayedDraws(t *testing.T) {
	t.Parallel()
	source := "Counter target spell. Its controller may draw up to two cards at the beginning of the next turn's upkeep.\nYou draw a card at the beginning of the next turn's upkeep."
	compilation, diagnostics := compileSource(source, pipelineContext{InstantOrSorcery: true})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	var effects []CompiledEffect
	for _, ability := range compilation.Abilities {
		effects = append(effects, ability.Content.Effects...)
	}
	if len(effects) != 3 {
		t.Fatalf("effects = %#v, want three", effects)
	}
	targetDraw := effects[1]
	if targetDraw.Context != parser.EffectContextReferencedObjectController ||
		targetDraw.DelayedTiming != game.DelayedAtBeginningOfNextUpkeep ||
		!targetDraw.Optional ||
		!targetDraw.Amount.RangeKnown ||
		targetDraw.Amount.Minimum != 0 ||
		targetDraw.Amount.Maximum != 2 ||
		len(targetDraw.References) != 1 ||
		targetDraw.References[0].Binding != ReferenceBindingTarget ||
		targetDraw.References[0].Occurrence != 0 {
		t.Fatalf("target-controller draw = %#v", targetDraw)
	}
	controllerDraw := effects[2]
	if controllerDraw.Context != parser.EffectContextController ||
		controllerDraw.DelayedTiming != game.DelayedAtBeginningOfNextUpkeep ||
		controllerDraw.Optional ||
		!controllerDraw.Amount.Known ||
		controllerDraw.Amount.Value != 1 {
		t.Fatalf("controller draw = %#v", controllerDraw)
	}
}

func TestCompileDelayedBlinkEffects(t *testing.T) {
	t.Parallel()
	compilation, diagnostics := compileSource(
		"Exile target creature. Return that card to the battlefield under its owner's control at the beginning of the next end step.",
		pipelineContext{InstantOrSorcery: true},
	)
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	effects := compilation.Abilities[0].Content.Effects
	if len(effects) != 2 ||
		effects[0].Kind != EffectExile ||
		effects[0].DelayedTiming != 0 ||
		effects[1].Kind != EffectReturn ||
		effects[1].DelayedTiming != game.DelayedAtBeginningOfNextEndStep {
		t.Fatalf("effects = %#v, want immediate exile followed by delayed return", effects)
	}
}

func TestCompileDynamicEffectAmounts(t *testing.T) {
	t.Parallel()
	tests := []struct {
		source     string
		context    pipelineContext
		kind       DynamicAmountKind
		form       DynamicAmountForm
		multiplier int
		selector   SelectorKind
		controller ControllerKind
		text       string
	}{
		{"Swarm deals damage equal to the number of creatures you control to any target.", pipelineContext{CardName: "Swarm", InstantOrSorcery: true}, DynamicAmountCount, DynamicAmountEqual, 1, SelectorCreature, ControllerYou, "equal to the number of creatures you control"},
		{"Swarm deals damage equal to twice the number of lands on the battlefield to any target.", pipelineContext{CardName: "Swarm", InstantOrSorcery: true}, DynamicAmountCount, DynamicAmountEqual, 2, SelectorLand, ControllerAny, "equal to twice the number of lands on the battlefield"},
		{"You gain 2 life for each opponent you have.", pipelineContext{InstantOrSorcery: true}, DynamicAmountOpponentCount, DynamicAmountForEach, 2, SelectorUnknown, ControllerAny, "for each opponent you have"},
		{"You gain life equal to your life total.", pipelineContext{InstantOrSorcery: true}, DynamicAmountControllerLife, DynamicAmountEqual, 1, SelectorUnknown, ControllerAny, "equal to your life total"},
		{"You gain X life, where X is your life total.", pipelineContext{InstantOrSorcery: true}, DynamicAmountControllerLife, DynamicAmountWhereX, 1, SelectorUnknown, ControllerAny, "where X is your life total"},
		{"When this creature dies, it deals damage equal to its power to any target.", pipelineContext{CardName: "Devil"}, DynamicAmountSourcePower, DynamicAmountEqual, 1, SelectorUnknown, ControllerAny, "equal to its power"},
		{"{T}: Put X +1/+1 counters on target creature, where X is Druid's power.", pipelineContext{CardName: "Druid"}, DynamicAmountSourcePower, DynamicAmountWhereX, 1, SelectorUnknown, ControllerAny, "where X is Druid's power"},
		{"{T}: Put X +1/+1 counters on target creature, where X is Fight Bear's power.", pipelineContext{CardName: "Fight Bear"}, DynamicAmountSourcePower, DynamicAmountWhereX, 1, SelectorUnknown, ControllerAny, "where X is Fight Bear's power"},
		{"You gain 2 life for each basic land type among lands you control.", pipelineContext{InstantOrSorcery: true}, DynamicAmountBasicLandTypes, DynamicAmountForEach, 2, SelectorUnknown, ControllerAny, "for each basic land type among lands you control"},
		{"Flames deals damage equal to the number of basic land types among lands you control to any target.", pipelineContext{CardName: "Flames", InstantOrSorcery: true}, DynamicAmountBasicLandTypes, DynamicAmountEqual, 1, SelectorUnknown, ControllerAny, "equal to the number of basic land types among lands you control"},
		{"Flames deals X damage to any target, where X is the number of basic land types among lands you control.", pipelineContext{CardName: "Flames", InstantOrSorcery: true}, DynamicAmountBasicLandTypes, DynamicAmountWhereX, 1, SelectorUnknown, ControllerAny, "where X is the number of basic land types among lands you control"},
		{"Swarm deals damage equal to the number of cards in your hand to any target.", pipelineContext{CardName: "Swarm", InstantOrSorcery: true}, DynamicAmountCount, DynamicAmountEqual, 1, SelectorCard, ControllerYou, "equal to the number of cards in your hand"},
		{"Sacrifice a creature: Target player mills cards equal to the sacrificed creature's power.", pipelineContext{CardName: "Altar"}, DynamicAmountSacrificedPower, DynamicAmountEqual, 1, SelectorUnknown, ControllerAny, "equal to the sacrificed creature's power"},
		{"Sacrifice a creature: Target player mills cards equal to the sacrificed creature's toughness.", pipelineContext{CardName: "Altar"}, DynamicAmountSacrificedToughness, DynamicAmountEqual, 1, SelectorUnknown, ControllerAny, "equal to the sacrificed creature's toughness"},
		{"Sacrifice a creature: Target player mills cards equal to the sacrificed creature's mana value.", pipelineContext{CardName: "Altar"}, DynamicAmountSacrificedManaValue, DynamicAmountEqual, 1, SelectorUnknown, ControllerAny, "equal to the sacrificed creature's mana value"},
	}

	for _, test := range tests {
		t.Run(test.source, func(t *testing.T) {
			t.Parallel()
			compilation, diagnostics := compileSource(test.source, test.context)
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			amount := compilation.Abilities[0].Content.Effects[0].Amount
			if amount.DynamicKind != test.kind ||
				amount.DynamicForm != test.form ||
				amount.Multiplier != test.multiplier ||
				amount.Selector().Kind != test.selector ||
				amount.Selector().Controller != test.controller ||
				amount.Text != test.text {
				t.Fatalf("amount = %#v tokens = %#v", amount, compilation.Syntax.Abilities[0].Tokens)
			}
			if test.kind == DynamicAmountSourcePower && amount.ReferenceSpan == (shared.Span{}) {
				t.Fatal("source-power amount has no reference span")
			}
		})
	}
}

// TestCompileSubtypeCountAmounts verifies that "the number of <subtype>" count
// phrases (the plural-headed "equal to"/"where X is" forms) resolve to a
// DynamicAmountCount carrying the subtype in the count selection. The singular
// head ("the number of Goblin") is ungrammatical and must not resolve, keeping
// the count fail-closed.
func TestCompileSubtypeCountAmounts(t *testing.T) {
	t.Parallel()
	tests := []struct {
		source  string
		form    DynamicAmountForm
		subtype types.Sub
	}{
		{"Swarm deals damage equal to the number of Goblins you control to any target.", DynamicAmountEqual, types.Goblin},
		{"Swarm deals damage equal to the number of Mountains you control to any target.", DynamicAmountEqual, types.Mountain},
		{"Swarm deals X damage to any target, where X is the number of Wizards you control.", DynamicAmountWhereX, types.Wizard},
	}
	for _, test := range tests {
		t.Run(test.source, func(t *testing.T) {
			t.Parallel()
			compilation, diagnostics := compileSource(test.source, pipelineContext{CardName: "Swarm", InstantOrSorcery: true})
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			amount := compilation.Abilities[0].Content.Effects[0].Amount
			if amount.DynamicKind != DynamicAmountCount ||
				amount.DynamicForm != test.form {
				t.Fatalf("amount = %#v, want subtype count %v", amount, test.form)
			}
			subtypes := amount.Selector().SubtypesAny()
			if len(subtypes) != 1 || subtypes[0] != test.subtype ||
				amount.Selector().Controller != ControllerYou {
				t.Fatalf("selector = %#v, want subtype %v controlled by you", amount.Selector(), test.subtype)
			}
		})
	}
}

func TestCompileSubtypeCountSingularHeadFailsClosed(t *testing.T) {
	t.Parallel()
	compilation, _ := compileSource(
		"Swarm deals damage equal to the number of Goblin you control to any target.",
		pipelineContext{CardName: "Swarm", InstantOrSorcery: true},
	)
	amount := compilation.Abilities[0].Content.Effects[0].Amount
	if amount.DynamicKind == DynamicAmountCount {
		t.Fatalf("singular count head resolved to a count amount: %#v", amount)
	}
}

func TestCompileWithCyclingTargetSelector(t *testing.T) {
	t.Parallel()
	source := "Return up to two target cards with cycling from your graveyard to your hand."
	compilation, diagnostics := compileSource(source, pipelineContext{InstantOrSorcery: true})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	target := compilation.Abilities[0].Content.Targets[0]
	if target.Cardinality.Min != 0 || target.Cardinality.Max != 2 {
		t.Fatalf("cardinality = %#v, want up to two", target.Cardinality)
	}
	if target.Selector.Kind != SelectorCard || target.Selector.Keyword != parser.KeywordCycling {
		t.Fatalf("selector = %#v, want card with Cycling", target.Selector)
	}
}

func TestCompileDynamicCardCountWithCyclingInGraveyard(t *testing.T) {
	t.Parallel()
	source := "Flare deals X damage to any target, where X is the number of cards with a cycling ability in your graveyard."
	compilation, diagnostics := compileSource(source, pipelineContext{CardName: "Flare", InstantOrSorcery: true})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	amount := compilation.Abilities[0].Content.Effects[0].Amount
	if amount.DynamicKind != DynamicAmountCount ||
		amount.DynamicForm != DynamicAmountWhereX ||
		amount.Selector().Kind != SelectorCard ||
		amount.Selector().Keyword != parser.KeywordCycling ||
		amount.Selector().Zone != zone.Graveyard ||
		amount.Selector().Controller != ControllerYou {
		t.Fatalf("amount = %#v, want count of cards with Cycling in your graveyard", amount)
	}
}

func TestCompileNamedCounterKinds(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		kind counter.Kind
	}{
		{"+1/+1", counter.PlusOnePlusOne},
		{"charge", counter.Charge},
		{"first strike", counter.FirstStrike},
		{"poison", counter.Poison},
		{"experience", counter.Experience},
		{"stun", counter.Stun},
	}
	for _, test := range tests {
		source := "Put a " + test.name + " counter on target permanent."
		compilation, diagnostics := compileSource(source, pipelineContext{InstantOrSorcery: true})
		if len(diagnostics) != 0 {
			t.Fatalf("%q diagnostics = %#v", source, diagnostics)
		}
		effect := compilation.Abilities[0].Content.Effects[0]
		if !effect.CounterKindKnown || effect.CounterKind != test.kind {
			t.Fatalf("%q counter kind = %v, %v", source, effect.CounterKind, effect.CounterKindKnown)
		}
	}

	compilation, diagnostics := compileSource(
		"Put a fade counter on target permanent.",
		pipelineContext{InstantOrSorcery: true},
	)
	if len(diagnostics) != 0 {
		t.Fatalf("unknown counter diagnostics = %#v", diagnostics)
	}
	if compilation.Abilities[0].Content.Effects[0].CounterKindKnown {
		t.Fatal("unknown counter kind was recognized")
	}
}

func TestCompileEntersWithCounterKind(t *testing.T) {
	t.Parallel()
	compilation, diagnostics := compileSource(
		"This creature enters with three +1/+1 counters on it.",
		pipelineContext{},
	)
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	effect := compilation.Abilities[0].Content.Effects[0]
	if effect.Kind != EffectEnterTapped ||
		!effect.CounterKindKnown ||
		effect.CounterKind != counter.PlusOnePlusOne ||
		!effect.Amount.Known ||
		effect.Amount.Value != 3 {
		t.Fatalf("effect = %#v, want fixed +1/+1 ETB counters", effect)
	}
}

// TestCompileNamedCounterKindsRejectsMissingRuntimeMechanics verifies that
// counter kinds whose placement has no complete runtime semantics stay
// unrecognized for placement. Finality counters remain unsupported because
// their death-replacement behavior is not yet modeled; stun counters, by
// contrast, are now recognized (see TestCompileNamedCounterKinds).
func TestCompileNamedCounterKindsRejectsMissingRuntimeMechanics(t *testing.T) {
	t.Parallel()
	for _, name := range []string{"finality"} {
		source := "Put a " + name + " counter on target creature."
		compilation, diagnostics := compileSource(source, pipelineContext{InstantOrSorcery: true})
		if len(diagnostics) != 0 {
			t.Fatalf("%q diagnostics = %#v", source, diagnostics)
		}
		effect := compilation.Abilities[0].Content.Effects[0]
		if effect.CounterKindKnown {
			t.Fatalf("%q counter kind was accepted for placement", source)
		}
	}
}

func TestCompileDynamicEffectAmountOffset(t *testing.T) {
	t.Parallel()
	tests := []struct {
		source string
		offset int
		text   string
	}{
		{"Draw cards equal to the number of cards in your hand plus one.", 1, "equal to the number of cards in your hand plus one"},
		{"Swarm deals damage equal to the number of creatures you control plus two to any target.", 2, "equal to the number of creatures you control plus two"},
	}
	for _, test := range tests {
		t.Run(test.source, func(t *testing.T) {
			t.Parallel()
			compilation, diagnostics := compileSource(test.source, pipelineContext{CardName: "Swarm", InstantOrSorcery: true})
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			amount := compilation.Abilities[0].Content.Effects[0].Amount
			if amount.DynamicKind != DynamicAmountCount ||
				amount.Addend != test.offset ||
				amount.Text != test.text {
				t.Fatalf("amount = %#v", amount)
			}
		})
	}
}

func TestCompileDynamicEffectAmountsRejectsAmbiguousSubjects(t *testing.T) {
	t.Parallel()
	for _, source := range []string{
		"You gain 2 life for each opponent and creature.",
		"Swarm deals damage equal to creatures you control to any target.",
		"You gain X life, where X is opponent.",
	} {
		t.Run(source, func(t *testing.T) {
			t.Parallel()
			compilation, diagnostics := compileSource(source, pipelineContext{
				CardName:         "Swarm",
				InstantOrSorcery: true,
			})
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			if amount := compilation.Abilities[0].Content.Effects[0].Amount; amount.DynamicKind != DynamicAmountNone || amount.Known {
				t.Fatalf("amount = %#v, want unsupported", amount)
			}
		})
	}
}

func TestCompileDynamicEffectAmountsRejectsNumberDisagreement(t *testing.T) {
	t.Parallel()
	for _, source := range []string{
		"Draw a card for each creatures you control.",
		"Swarm deals damage equal to the number of creature you control to any target.",
	} {
		t.Run(source, func(t *testing.T) {
			t.Parallel()
			compilation, diagnostics := compileSource(source, pipelineContext{
				CardName:         "Swarm",
				InstantOrSorcery: true,
			})
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			if amount := compilation.Abilities[0].Content.Effects[0].Amount; amount.DynamicKind != DynamicAmountNone || amount.Known {
				t.Fatalf("amount = %#v, want unsupported", amount)
			}
		})
	}
}

func TestCompileEffectAmountsAreClauseLocal(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		source string
		check  func(*testing.T, []CompiledEffect)
	}{
		{
			name:   "fixed then dynamic effect",
			source: "You gain 2 life, then draw a card for each creature you control.",
			check: func(t *testing.T, effects []CompiledEffect) {
				t.Helper()
				assertFixedEffectAmount(t, effects, EffectGain, 2)
				assertDynamicEffectAmount(t, effects, EffectDraw, DynamicAmountCount)
			},
		},
		{
			name:   "dynamic then fixed effect",
			source: "Draw a card for each creature you control, then you gain 2 life.",
			check: func(t *testing.T, effects []CompiledEffect) {
				t.Helper()
				assertDynamicEffectAmount(t, effects, EffectDraw, DynamicAmountCount)
				assertFixedEffectAmount(t, effects, EffectGain, 2)
			},
		},
		{
			name:   "and separates effects",
			source: "Draw a card for each creature you control and gain 2 life.",
			check: func(t *testing.T, effects []CompiledEffect) {
				t.Helper()
				assertDynamicEffectAmount(t, effects, EffectDraw, DynamicAmountCount)
				assertFixedEffectAmount(t, effects, EffectGain, 2)
			},
		},
		{
			name:   "fixed before condition formula",
			source: "You gain 2 life if the number of creatures you control is greater than 3.",
			check: func(t *testing.T, effects []CompiledEffect) {
				t.Helper()
				assertFixedEffectAmount(t, effects, EffectGain, 2)
			},
		},
		{
			name:   "dynamic before condition amount",
			source: "Draw a card for each creature you control unless your life total is 2.",
			check: func(t *testing.T, effects []CompiledEffect) {
				t.Helper()
				assertDynamicEffectAmount(t, effects, EffectDraw, DynamicAmountCount)
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			compilation, diagnostics := compileSource(test.source, pipelineContext{InstantOrSorcery: true})
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			test.check(t, compilation.Abilities[0].Content.Effects)
		})
	}
}

func assertFixedEffectAmount(t *testing.T, effects []CompiledEffect, kind EffectKind, value int) {
	t.Helper()
	for i := range effects {
		effect := &effects[i]
		if effect.Kind == kind {
			if !effect.Amount.Known ||
				effect.Amount.Value != value ||
				effect.Amount.DynamicKind != DynamicAmountNone {
				t.Fatalf("%v amount = %#v, want fixed %d", kind, effect.Amount, value)
			}
			return
		}
	}
	t.Fatalf("effects = %#v, missing %v", effects, kind)
}

func assertDynamicEffectAmount(t *testing.T, effects []CompiledEffect, kind EffectKind, dynamicKind DynamicAmountKind) {
	t.Helper()
	for i := range effects {
		effect := &effects[i]
		if effect.Kind == kind {
			if effect.Amount.Known || effect.Amount.DynamicKind != dynamicKind {
				t.Fatalf("%v amount = %#v, want dynamic %v", kind, effect.Amount, dynamicKind)
			}
			return
		}
	}
	t.Fatalf("effects = %#v, missing %v", effects, kind)
}
