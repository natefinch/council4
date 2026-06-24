package cardgen

import (
	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/opt"
)

// lowerThresholdInsteadManaSpellAbilities lowers the Threshold conditional-mana
// cycle: a base mana production followed by a larger "Add <more> instead if
// <condition>" alternative (Cabal Ritual: "Add {B}{B}{B}.\nThreshold — Add
// {B}{B}{B}{B}{B} instead if there are seven or more cards in your
// graveyard."). The two paragraphs compile to two separate add-mana spell
// abilities; the second carries the parser's Instead flag and a single
// effect-gate condition. They fuse into one spell whose base output resolves
// only when the condition fails and whose larger output resolves only when it
// holds, so exactly one production happens.
func lowerThresholdInsteadManaSpellAbilities(cardName string, compilation compiler.Compilation) (game.AbilityContent, bool) {
	_ = cardName
	if len(compilation.Abilities) != 2 ||
		len(compilation.Syntax.Abilities) != 2 {
		return game.AbilityContent{}, false
	}
	base := compilation.Abilities[0]
	alternative := compilation.Abilities[1]
	if !isPlainSpellShell(base) || !isPlainSpellShell(alternative) {
		return game.AbilityContent{}, false
	}
	baseEffect, ok := soleFixedAddManaEffect(base.Content, false)
	if !ok {
		return game.AbilityContent{}, false
	}
	alternativeEffect, ok := soleFixedAddManaEffect(alternative.Content, true)
	if !ok {
		return game.AbilityContent{}, false
	}
	condition, ok := lowerCondition(alternative.Content.Conditions[0], conditionContextEffectGate)
	if !ok {
		return game.AbilityContent{}, false
	}
	notCondition := condition
	notCondition.Negate = !notCondition.Negate
	baseSeq := gatedFixedAddMana(baseEffect.Mana.Colors, &notCondition)
	alternativeSeq := gatedFixedAddMana(alternativeEffect.Mana.Colors, &condition)
	return game.Mode{Sequence: append(baseSeq, alternativeSeq...)}.Ability(), true
}

// isPlainSpellShell reports whether an ability is an unconditional spell with no
// trigger, cost, static rider, or optional wrapper (an ability word such as
// "Threshold" is allowed and carries no semantics).
func isPlainSpellShell(ability compiler.CompiledAbility) bool {
	return ability.Kind == compiler.AbilitySpell &&
		ability.Trigger == nil &&
		ability.Cost == nil &&
		ability.Static == nil &&
		ability.AlternativeCost == nil &&
		!ability.Optional
}

// soleFixedAddManaEffect validates that an ability's content is a single
// fixed-color add-mana effect to the controller and returns it. wantInstead
// selects between the base production (no Instead flag, no conditions) and the
// conditional alternative (Instead flag set, exactly one condition).
func soleFixedAddManaEffect(content compiler.AbilityContent, wantInstead bool) (compiler.CompiledEffect, bool) {
	wantConditions := 0
	if wantInstead {
		wantConditions = 1
	}
	if len(content.Effects) != 1 ||
		len(content.Conditions) != wantConditions ||
		len(content.Targets) != 0 ||
		len(content.Keywords) != 0 ||
		len(content.Modes) != 0 ||
		len(content.References) != 0 {
		return compiler.CompiledEffect{}, false
	}
	effect := content.Effects[0]
	if !fixedAddManaToController(effect, wantInstead) {
		return compiler.CompiledEffect{}, false
	}
	return effect, true
}

// fixedAddManaToController reports whether an effect adds a fixed, known-color
// amount of mana to its controller with no targets or references. wantInstead
// selects the parser's Instead flag (the larger conditional alternative) versus
// the base production.
func fixedAddManaToController(effect compiler.CompiledEffect, wantInstead bool) bool {
	if effect.Kind != compiler.EffectAddMana ||
		effect.Negated ||
		effect.Optional ||
		effect.DelayedTiming != 0 ||
		effect.Duration != compiler.DurationNone ||
		effect.Context != parser.EffectContextController ||
		effect.Payment.Form != parser.EffectPaymentFormUnknown ||
		len(effect.Targets) != 0 ||
		len(effect.References) != 0 {
		return false
	}
	manaEffect := effect.Mana
	return manaEffect.Instead == wantInstead &&
		manaEffect.ColorsKnown &&
		!manaEffect.Choice &&
		!manaEffect.AnyColor &&
		len(manaEffect.Colors) != 0
}

// tronConditionalManaContent detects a single activated mana ability that adds a
// base fixed amount of mana and, gated on an effect-condition, a different fixed
// amount "instead" (the Urza tron lands: "{T}: Add {C}. If you control an
// Urza's Power-Plant and an Urza's Tower, add {C}{C} instead."). It returns a
// stripped ability that produces only the base mana — for ordinary activation
// shell lowering — and the gated content that resolves exactly one production
// depending on whether the condition holds.
func tronConditionalManaContent(ability compiler.CompiledAbility) (compiler.CompiledAbility, game.AbilityContent, bool) {
	content := ability.Content
	if len(content.Effects) != 2 ||
		len(content.Conditions) != 1 ||
		len(content.Targets) != 0 ||
		len(content.Keywords) != 0 ||
		len(content.Modes) != 0 ||
		len(content.References) != 0 {
		return compiler.CompiledAbility{}, game.AbilityContent{}, false
	}
	baseEffect := content.Effects[0]
	altEffect := content.Effects[1]
	if !fixedAddManaToController(baseEffect, false) || !fixedAddManaToController(altEffect, true) {
		return compiler.CompiledAbility{}, game.AbilityContent{}, false
	}
	condition, ok := lowerCondition(content.Conditions[0], conditionContextEffectGate)
	if !ok {
		return compiler.CompiledAbility{}, game.AbilityContent{}, false
	}
	notCondition := condition
	notCondition.Negate = !notCondition.Negate
	baseSeq := gatedFixedAddMana(baseEffect.Mana.Colors, &notCondition)
	altSeq := gatedFixedAddMana(altEffect.Mana.Colors, &condition)
	gated := game.Mode{Sequence: append(baseSeq, altSeq...)}.Ability()

	stripped := ability
	stripped.Content = content
	strippedBase := baseEffect
	strippedBase.RequiresOrderedLowering = false
	stripped.Content.Effects = []compiler.CompiledEffect{strippedBase}
	stripped.Content.Conditions = nil
	return stripped, gated, true
}

// counterConditionalMultiplierManaContent detects a single activated mana
// ability that adds a base production and, gated on a self-counter state
// condition, multiplies that same production "instead" (Incubation Druid:
// "{T}: Add one mana of any type that a land you control could produce. If this
// creature has a +1/+1 counter on it, add three mana of that type instead."). The
// conditional alternative reuses the base production's chosen type as an empty
// "<n> mana of that type" backref, so the runtime keeps the base production's
// setup (such as the lands-produce Choose) and gates each resulting AddMana into
// a Fixed(1) production when the counter is absent and a Fixed(n) production when
// it is present, leaving exactly one production. It returns a stripped ability
// that produces only the base mana — for ordinary activation shell lowering — and
// the gated content. Unlike tronConditionalManaContent, the base and the
// alternative need not be fixed colors: any base whose runtime content is a
// setup-then-Fixed(1)-AddMana production is multiplied. It fails closed unless
// the rider is an empty same-type backref, the gate is a self-counter state, and
// every base production unit is a Fixed(1) AddMana.
func counterConditionalMultiplierManaContent(ability compiler.CompiledAbility) (compiler.CompiledAbility, game.AbilityContent, bool) {
	content := ability.Content
	if len(content.Effects) != 2 ||
		len(content.Conditions) != 1 ||
		len(content.Targets) != 0 ||
		len(content.Keywords) != 0 ||
		len(content.Modes) != 0 ||
		!referencesAllSource(content.References) {
		return compiler.CompiledAbility{}, game.AbilityContent{}, false
	}
	multiplied, ok := manaTypeMultiplierRiderAmount(content.Effects[1])
	if !ok {
		return compiler.CompiledAbility{}, game.AbilityContent{}, false
	}
	if !isSelfCounterCondition(content.Conditions[0]) {
		return compiler.CompiledAbility{}, game.AbilityContent{}, false
	}
	condition, ok := lowerCondition(content.Conditions[0], conditionContextEffectGate)
	if !ok {
		return compiler.CompiledAbility{}, game.AbilityContent{}, false
	}
	baseContent, ok := baseSingleManaProductionContent(content.Effects[0])
	if !ok {
		return compiler.CompiledAbility{}, game.AbilityContent{}, false
	}
	gated, ok := gateManaProductionByCount(baseContent, condition, multiplied)
	if !ok {
		return compiler.CompiledAbility{}, game.AbilityContent{}, false
	}
	stripped := ability
	stripped.Content = content
	strippedBase := content.Effects[0]
	strippedBase.RequiresOrderedLowering = false
	stripped.Content.Effects = []compiler.CompiledEffect{strippedBase}
	stripped.Content.Conditions = nil
	stripped.Content.References = nil
	return stripped, gated, true
}

// baseSingleManaProductionContent validates that an effect is a single
// unconditional add-mana production of one mana to its controller and returns its
// runtime content. The base production may be any typed mana body the backend can
// lower (a fixed color, a choice, or a lands-produce choice); the counter rider
// multiplies whatever single mana it produces. It fails closed on a negated,
// optional, delayed, lasting, targeted, referenced, paid, replacement, or
// non-unit-amount add-mana effect — including one that itself carries an
// "instead" replacement, which only the rider may.
func baseSingleManaProductionContent(effect compiler.CompiledEffect) (game.AbilityContent, bool) {
	if effect.Kind != compiler.EffectAddMana ||
		effect.Negated ||
		effect.Optional ||
		effect.DelayedTiming != 0 ||
		effect.Duration != compiler.DurationNone ||
		effect.Context != parser.EffectContextController ||
		effect.Payment.Form != parser.EffectPaymentFormUnknown ||
		effect.HasUnrecognizedSibling ||
		effect.Mana.Instead ||
		effect.Replacement.Kind != parser.EffectReplacementNone ||
		len(effect.Targets) != 0 ||
		len(effect.References) != 0 {
		return game.AbilityContent{}, false
	}
	if !effect.Amount.Known ||
		effect.Amount.VariableX ||
		effect.Amount.DynamicKind != compiler.DynamicAmountNone ||
		effect.Amount.Value != 1 {
		return game.AbilityContent{}, false
	}
	return typedManaEffectContent(effect.Mana)
}

// manaTypeMultiplierRiderAmount reports the multiplied amount of a counter rider
// "add <n> mana of that type instead" — an add-mana effect to the controller
// whose mana body is the empty same-type backref (no symbols, colors, choice, or
// typed production of its own), that carries the "instead" replacement marker,
// and whose fixed amount is two or more. The empty body marks it as multiplying
// the base production's chosen type rather than adding a distinct one, and the
// "instead" marker confirms the rider replaces the base production rather than
// stacking with it (so the mutually exclusive gating is well-defined). It fails
// closed on any other add-mana shape, including an additive "add <n> mana of that
// type" with no "instead".
func manaTypeMultiplierRiderAmount(effect compiler.CompiledEffect) (int, bool) {
	if effect.Kind != compiler.EffectAddMana ||
		effect.Negated ||
		effect.Optional ||
		effect.DelayedTiming != 0 ||
		effect.Duration != compiler.DurationNone ||
		(effect.Context != parser.EffectContextController &&
			effect.Context != parser.EffectContextPriorSubject) ||
		effect.Payment.Form != parser.EffectPaymentFormUnknown ||
		effect.HasUnrecognizedSibling ||
		effect.Replacement.Kind != parser.EffectReplacementInstead ||
		len(effect.Targets) != 0 ||
		len(effect.References) != 0 {
		return 0, false
	}
	if !manaBodyEmpty(effect.Mana) {
		return 0, false
	}
	if !effect.Amount.Known ||
		effect.Amount.VariableX ||
		effect.Amount.DynamicKind != compiler.DynamicAmountNone ||
		effect.Amount.Value < 2 {
		return 0, false
	}
	return effect.Amount.Value, true
}

// manaBodyEmpty reports whether a compiled mana body carries no type production
// of its own — every discriminant is unset — which marks an "<n> mana of that
// type" backref that multiplies a sibling production's chosen type.
func manaBodyEmpty(m compiler.CompiledEffectMana) bool {
	return len(m.Symbols) == 0 &&
		len(m.Colors) == 0 &&
		!m.ColorsKnown &&
		!m.Choice &&
		!m.AnyColor &&
		!m.ChosenColor &&
		!m.ChosenColorFixedKnown &&
		!m.ChosenColorDevotion &&
		!m.ChosenColorDynamic &&
		!m.CommanderIdentity &&
		!m.DynamicColorless &&
		!m.LegacyBodyExact &&
		!m.FilterPair &&
		len(m.FilterColors) == 0 &&
		!m.LandsProduce &&
		!m.LinkedExileColors &&
		!m.ColorsAmongControlled &&
		m.ColorsAmongSelector == nil &&
		!m.EachColorAmongControlled &&
		!m.AnyOneColorDynamic &&
		m.AnyColorCount == 0 &&
		!m.Instead &&
		!m.TriggerLandProducedType
}

// referencesAllSource reports whether every content reference binds to the
// source permanent — the inert self-references a self-counter gate introduces
// ("this creature has a +1/+1 counter on it" yields "this creature" and "it").
// They describe the source rather than targeting or wiring extra objects, so the
// multiplier may safely ignore them; any non-source reference fails closed.
func referencesAllSource(references []compiler.CompiledReference) bool {
	for _, reference := range references {
		if reference.Binding != compiler.ReferenceBindingSource {
			return false
		}
	}
	return true
}

// isSelfCounterCondition reports whether a compiled condition tests the source
// permanent's own counters: an object-match on the source binding whose selection
// requires a specific counter kind ("has a +1/+1 counter on it") or any counter
// ("has counters on it"). It scopes the counter-conditional mana multiplier to a
// self-counter gate, matching the construct's wording family.
func isSelfCounterCondition(condition compiler.CompiledCondition) bool {
	return condition.Predicate == compiler.ConditionPredicateObjectMatches &&
		condition.ObjectBinding == compiler.ReferenceBindingSource &&
		(condition.Selection.CounterKindKnown || condition.Selection.AnyCounter)
}

// gateManaProductionByCount rebuilds a base mana production so each unit it adds
// resolves as a Fixed(1) production when condition is false and a Fixed(multiplied)
// production when it is true, leaving exactly one production. Setup instructions
// (such as the lands-produce Choose) are preserved ungated; every produced AddMana
// must be an unconditional Fixed(1) add so the multiplier is well-defined. It fails
// closed on modal content, a pre-gated instruction, or a non-unit production.
func gateManaProductionByCount(base game.AbilityContent, condition game.Condition, multiplied int) (game.AbilityContent, bool) {
	if base.IsModal() {
		return game.AbilityContent{}, false
	}
	notCondition := condition
	notCondition.Negate = !notCondition.Negate
	source := base.Modes[0].Sequence
	seq := make([]game.Instruction, 0, len(source)+1)
	gated := false
	for _, instruction := range source {
		if !instructionIsPlainPrimitive(instruction) {
			return game.AbilityContent{}, false
		}
		add, ok := instruction.Primitive.(game.AddMana)
		if !ok {
			seq = append(seq, instruction)
			continue
		}
		if add.Amount.IsDynamic() || add.Amount.Value() != 1 {
			return game.AbilityContent{}, false
		}
		baseAdd := add
		baseAdd.Amount = game.Fixed(1)
		multipliedAdd := add
		multipliedAdd.Amount = game.Fixed(multiplied)
		seq = append(seq,
			gatedInstruction(baseAdd, notCondition),
			gatedInstruction(multipliedAdd, condition))
		gated = true
	}
	if !gated {
		return game.AbilityContent{}, false
	}
	return game.Mode{Sequence: seq}.Ability(), true
}

// instructionIsPlainPrimitive reports whether an instruction carries only its
// primitive, with no existing gate, optional flow, or result wiring, so the
// multiplier may safely re-gate it.
func instructionIsPlainPrimitive(instruction game.Instruction) bool {
	return !instruction.Condition.Exists &&
		!instruction.CardCondition.Exists &&
		!instruction.ResultGate.Exists &&
		!instruction.Optional &&
		!instruction.OptionalActor.Exists &&
		instruction.PublishResult == "" &&
		instruction.Description == ""
}

// gatedInstruction wraps a primitive in an instruction gated on the supplied
// condition so it resolves only when the condition holds.
func gatedInstruction(primitive game.Primitive, condition game.Condition) game.Instruction {
	return game.Instruction{
		Primitive: primitive,
		Condition: opt.Val(game.EffectCondition{Condition: opt.Val(condition)}),
	}
}

// gatedFixedAddMana builds one fixed-color AddMana instruction per color, each
// gated on the supplied condition so the production resolves only when the
// condition holds.
func gatedFixedAddMana(colors []mana.Color, condition *game.Condition) []game.Instruction {
	gate := opt.Val(game.EffectCondition{Condition: opt.Val(*condition)})
	seq := make([]game.Instruction, 0, len(colors))
	for _, c := range colors {
		seq = append(seq, game.Instruction{
			Primitive: game.AddMana{Amount: game.Fixed(1), ManaColor: c},
			Condition: gate,
		})
	}
	return seq
}
