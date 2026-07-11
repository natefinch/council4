package cardgen

import (
	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// lowerCantBeBlockedSpell lowers the temporary combat-evasion effect "<subject>
// can't be blocked this turn." into ApplyRule instructions that place a
// RuleEffectCantBeBlocked restriction on each affected creature for the turn
// (game.DurationThisTurn, removed during cleanup). It accepts the subject
// shapes the parser recognizes: a target noun phrase with single, plural, or
// optional cardinality ("Up to one target creature can't be blocked this
// turn."), the source itself ("This creature can't be blocked this turn."), a
// prior-subject sequence clause that inherits the source as its subject ("...
// and can't be blocked this turn."), a demonstrative back-reference that names a
// permanent introduced by a preceding clause or the triggering event ("... put a
// +1/+1 counter on this creature. It can't be blocked this turn.", Kappa
// Cannoneer), and the compound "source and up to one other target creature"
// subject (Martha Jones), where the source and each chosen target each gain the
// restriction. Every other recipient, duration, condition, mode, or reference
// fails closed so the broader "can't be blocked this turn" family stays faithful
// and bounded.
func lowerCantBeBlockedSpell(ctx contentCtx) (game.AbilityContent, *shared.Diagnostic) {
	unsupported := func() (game.AbilityContent, *shared.Diagnostic) {
		return game.AbilityContent{}, contentDiagnostic(
			ctx,
			"unsupported can't-be-blocked effect",
			"the executable source backend supports only exact \"<subject> can't be blocked this turn.\"",
		)
	}
	effect := ctx.content.Effects[0]
	if !effect.Exact ||
		effect.Negated ||
		effect.Optional ||
		(effect.Duration != compiler.DurationThisTurn && effect.Duration != compiler.DurationThisCombat) ||
		ctx.optional ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Keywords) != 0 ||
		len(ctx.content.Modes) != 0 {
		return unsupported()
	}
	// "this turn" grants the restriction for the whole turn; "this combat" ends it
	// when the combat phase does (Canal Courier).
	ruleDuration := game.DurationThisTurn
	if effect.Duration == compiler.DurationThisCombat {
		ruleDuration = game.DurationUntilEndOfCombat
	}
	targetSubject := effect.Context == parser.EffectContextTarget &&
		len(ctx.content.Targets) == 1 &&
		len(ctx.content.References) == 0 &&
		creatureTargetSubject(ctx.content.Targets[0])
	sourceSubject := (effect.Context == parser.EffectContextSource ||
		effect.Context == parser.EffectContextPriorSubject) &&
		len(ctx.content.Targets) == 0 &&
		len(ctx.content.References) == 1 &&
		ctx.content.References[0].Binding == compiler.ReferenceBindingSource
	referencedObjectSubject := effect.Context == parser.EffectContextReferencedObject &&
		len(ctx.content.Targets) == 0 &&
		len(ctx.content.References) == 1
	sourceAndTargetSubject := effect.Context == parser.EffectContextTarget &&
		len(ctx.content.Targets) == 1 &&
		ctx.content.Targets[0].Selector.Kind == compiler.SelectorCreature &&
		len(ctx.content.References) == 1 &&
		ctx.content.References[0].Binding == compiler.ReferenceBindingSource
	groupSubject := effect.StaticSubject != compiler.StaticSubjectNone &&
		len(ctx.content.Targets) == 0 &&
		len(ctx.content.References) == 0
	switch {
	case groupSubject:
		// "<creature group> can't be blocked this turn." (Keeper of Keys) applies
		// the restriction to every member of the group for the turn. It lowers to a
		// single group-scoped rule effect (no object anchor) whose affected-
		// permanent filter the runtime evaluates each combat, mirroring the static
		// "<group> can't be blocked" anthem but bounded to this turn.
		affectedController, permanentTypes, ok := cantBeBlockedGroupScope(effect.StaticSubject)
		if !ok {
			return unsupported()
		}
		return game.Mode{
			Sequence: []game.Instruction{cantBeBlockedGroupInstruction(affectedController, permanentTypes, ruleDuration)},
		}.Ability(), nil
	case sourceAndTargetSubject:
		// "<source> and up to one other target creature can't be blocked this
		// turn." (Martha Jones): the source itself plus the chosen target(s) each
		// gain the restriction. The source reference resolves to the source
		// permanent; the target slots resolve to the chosen creatures, and the
		// "other" qualifier excludes the source from being chosen twice.
		object, ok := lowerObjectReference(ctx.content.References[0], referenceLoweringContext{AllowSource: true})
		if !ok {
			return unsupported()
		}
		targetSpec, ok := permanentTargetSpecWithCardinality(ctx.content.Targets[0])
		if !ok {
			return unsupported()
		}
		sequence := make([]game.Instruction, 0, targetSpec.MaxTargets+1)
		sequence = append(sequence, cantBeBlockedInstruction(object, ruleDuration))
		for i := range targetSpec.MaxTargets {
			sequence = append(sequence, cantBeBlockedInstruction(game.TargetPermanentReference(i), ruleDuration))
		}
		return game.Mode{
			Targets:  []game.TargetSpec{targetSpec},
			Sequence: sequence,
		}.Ability(), nil
	case targetSubject:
		targetSpec, ok := permanentTargetSpecWithCardinality(ctx.content.Targets[0])
		if !ok {
			return unsupported()
		}
		sequence := make([]game.Instruction, 0, targetSpec.MaxTargets)
		for i := range targetSpec.MaxTargets {
			sequence = append(sequence, cantBeBlockedInstruction(game.TargetPermanentReference(i), ruleDuration))
		}
		return game.Mode{
			Targets:  []game.TargetSpec{targetSpec},
			Sequence: sequence,
		}.Ability(), nil
	case sourceSubject:
		object, ok := lowerObjectReference(ctx.content.References[0], referenceLoweringContext{AllowSource: true})
		if !ok {
			return unsupported()
		}
		return game.Mode{
			Sequence: []game.Instruction{cantBeBlockedInstruction(object, ruleDuration)},
		}.Ability(), nil
	case referencedObjectSubject:
		// "<back-reference> can't be blocked this turn." ("... put a +1/+1
		// counter on this creature. It can't be blocked this turn.", Kappa
		// Cannoneer) grants the restriction to the permanent a preceding clause or
		// the triggering event introduced. The demonstrative "it"/"that <object>"
		// resolves to that object, which may be the source, the triggering event
		// permanent, or a prior target; any other binding fails closed.
		object, ok := lowerObjectReference(ctx.content.References[0], referenceLoweringContext{
			AllowSource: true,
			AllowEvent:  true,
			AllowTarget: true,
		})
		if !ok {
			return unsupported()
		}
		return game.Mode{
			Sequence: []game.Instruction{cantBeBlockedInstruction(object, ruleDuration)},
		}.Ability(), nil
	default:
		return unsupported()
	}
}

// creatureTargetSubject reports whether the target selector names a creature
// subject that can carry a can't-be-blocked restriction: the bare "target
// creature" selector or a creature subtype noun ("target Detective", "target
// Merfolk"). It rejects non-creature permanent selectors, which the parser does
// not produce for this effect but which would be nonsensical to make unblockable.
func creatureTargetSubject(target compiler.CompiledTarget) bool {
	if target.Selector.Kind == compiler.SelectorCreature {
		return true
	}
	return target.Selector.Kind == compiler.SelectorUnknown &&
		len(target.Selector.SubtypesAny()) > 0
}

// cantBeBlockedInstruction builds the ApplyRule instruction that grants the
// given object a can't-be-blocked restriction for the turn.
func cantBeBlockedInstruction(object game.ObjectReference, duration game.EffectDuration) game.Instruction {
	return game.Instruction{
		Primitive: game.ApplyRule{
			Object: opt.Val(object),
			RuleEffects: []game.RuleEffect{
				{Kind: game.RuleEffectCantBeBlocked},
			},
			Duration: duration,
		},
	}
}

// cantBeBlockedGroupScope maps a static creature-group subject to the runtime
// rule-effect group scope: the affected-controller relation and the required card
// types. It covers the controller's creatures ("Creatures you control ...") and
// every creature ("Creatures ..."). Any other group subject fails closed so the
// group can't-be-blocked path stays bounded to the shapes it renders faithfully.
func cantBeBlockedGroupScope(subject compiler.StaticSubjectKind) (game.ControllerRelation, []types.Card, bool) {
	switch subject {
	case compiler.StaticSubjectControlledCreatures:
		return game.ControllerYou, []types.Card{types.Creature}, true
	case compiler.StaticSubjectAllCreatures:
		return game.ControllerAny, []types.Card{types.Creature}, true
	default:
		return game.ControllerAny, nil, false
	}
}

// cantBeBlockedGroupInstruction builds the ApplyRule instruction that grants a
// can't-be-blocked restriction to every member of a creature group for the turn.
// It carries no object anchor, so the runtime scopes the rule to the affected
// controller and card types rather than a single permanent.
func cantBeBlockedGroupInstruction(controller game.ControllerRelation, permanentTypes []types.Card, duration game.EffectDuration) game.Instruction {
	return game.Instruction{
		Primitive: game.ApplyRule{
			RuleEffects: []game.RuleEffect{{
				Kind:               game.RuleEffectCantBeBlocked,
				AffectedController: controller,
				PermanentTypes:     permanentTypes,
			}},
			Duration: duration,
		},
	}
}
