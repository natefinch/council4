package cardgen

import (
	"fmt"

	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/opt"
)

// lowerDelayedTriggerContent lowers a single EffectDelayedTrigger effect
// ("Whenever you cast a spell this turn, ...", "When you next cast a creature
// spell this turn, ...") into a game.CreateDelayedTrigger instruction. The
// nested triggered ability the parser reparsed is compiled and lowered
// recursively; its trigger pattern and content become the event-based delayed
// trigger's pattern and body, scoped to the rest of the turn. It fails closed
// when the effect carries outer targets, references, conditions, keywords, or
// modes, or when the nested ability does not lower to a plain event trigger.
func lowerDelayedTriggerContent(ctx contentCtx) (game.AbilityContent, *shared.Diagnostic) {
	effect := ctx.content.Effects[0]
	unsupported := func(detail string) (game.AbilityContent, *shared.Diagnostic) {
		return game.AbilityContent{}, contentDiagnostic(ctx, "unsupported delayed trigger", detail)
	}
	if effect.DelayedTriggerAbility == nil ||
		effect.Negated ||
		effect.Optional ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Keywords) != 0 ||
		len(ctx.content.Modes) != 0 ||
		len(ctx.content.Targets) != 0 ||
		len(ctx.content.References) != 0 {
		return unsupported("the executable source backend supports only an unconditional, untargeted event-based delayed trigger")
	}
	triggered, ok := lowerDelayedTriggerInner(effect.DelayedTriggerAbility)
	if !ok {
		return unsupported("the nested triggered ability did not lower to a plain event trigger")
	}
	return game.Mode{
		Sequence: []game.Instruction{{
			Primitive: game.CreateDelayedTrigger{Trigger: game.DelayedTriggerDef{
				EventPattern: opt.Val(triggered.Trigger.Pattern),
				OneShot:      effect.DelayedTriggerOneShot,
				Window:       game.DelayedWindowThisTurn,
				Content:      triggered.Content,
			}},
		}},
	}.Ability(), nil
}

// lowerDelayedTriggerInner compiles and lowers the nested triggered ability of a
// delayed trigger, returning the plain triggered ability whose pattern and
// content the delayed trigger reuses. It mirrors attachTokenGrantedAbility's
// recursive compile + lower of an already-parsed inner document, and fails
// closed when the inner document does not compile to exactly one plain triggered
// ability the delayed-trigger runtime can carry (no intervening-if, optional,
// keyword, or per-turn-limit machinery, which DelayedTriggerDef cannot model).
func lowerDelayedTriggerInner(granted *parser.StaticGrantedAbilitySyntax) (game.TriggeredAbility, bool) {
	innerDocument, innerDiags := granted.Inner()
	if len(innerDiags) != 0 {
		return game.TriggeredAbility{}, false
	}
	innerComp, compilerDiags := compiler.Compile(innerDocument, compiler.Context{})
	if len(compilerDiags) != 0 ||
		len(innerComp.Abilities) != 1 ||
		len(innerComp.Syntax.Abilities) != 1 {
		return game.TriggeredAbility{}, false
	}
	lowered, diagnostic := lowerExecutableAbility("", false, nil, innerComp.Abilities[0], &innerComp.Syntax.Abilities[0])
	if diagnostic != nil || !lowered.triggeredAbility.Exists {
		return game.TriggeredAbility{}, false
	}
	triggered := lowered.triggeredAbility.Val
	if triggered.Optional ||
		triggered.MaxTriggersPerTurn != 0 ||
		len(triggered.KeywordAbilities) != 0 ||
		triggered.Trigger.InterveningIf != "" ||
		triggered.Trigger.InterveningCondition.Exists {
		return game.TriggeredAbility{}, false
	}
	return triggered, true
}

// lowerDelayedCombatDamageDrawTrigger lowers a captured-object combat-damage
// rider clause ("... target creature ... Whenever that creature deals combat
// damage to a player this turn, you draw a card.") into the publishing pump and
// a CreateDelayedTrigger whose event source binds to the pumped permanent. The
// prior clause must be a ModifyPT acting on the shared target; this lowerer
// publishes that permanent under a linked key and rebinds the inner self-form
// combat-damage pattern's source to the captured object so the trigger fires
// only on combat damage that specific permanent deals. It mirrors
// lowerDelayedTargetReturn and fails closed on any other shape.
func lowerDelayedCombatDamageDrawTrigger(
	effectIndex int,
	ctx contentCtx,
	sequence []game.Instruction,
) (game.ModifyPT, game.AbilityContent, bool) {
	if effectIndex == 0 ||
		len(sequence) != effectIndex ||
		len(ctx.content.Effects) != 1 ||
		ctx.content.Effects[0].Kind != compiler.EffectDelayedTrigger ||
		!ctx.content.Effects[0].DelayedTriggerBindDamageSource ||
		ctx.content.Effects[0].DelayedTriggerAbility == nil ||
		ctx.content.Effects[0].Negated ||
		ctx.optional ||
		!referencesBindTo(ctx.content.References, compiler.ReferenceBindingTarget, 0) {
		return game.ModifyPT{}, game.AbilityContent{}, false
	}
	previous := sequence[effectIndex-1].Primitive
	if previous.Kind() != game.PrimitiveModifyPT {
		return game.ModifyPT{}, game.AbilityContent{}, false
	}
	modify, ok := previous.(game.ModifyPT)
	if !ok ||
		modify.Object.Kind() != game.ObjectReferenceTargetPermanent ||
		modify.PublishLinked != "" {
		return game.ModifyPT{}, game.AbilityContent{}, false
	}
	triggered, ok := lowerDelayedTriggerInner(ctx.content.Effects[0].DelayedTriggerAbility)
	if !ok {
		return game.ModifyPT{}, game.AbilityContent{}, false
	}
	pattern := triggered.Trigger.Pattern
	if pattern.Event != game.EventDamageDealt ||
		!pattern.RequireCombatDamage ||
		pattern.Source != game.TriggerSourceSelf ||
		pattern.DamageSourceCaptured ||
		!pattern.DamageSourceSelection.Empty() {
		return game.ModifyPT{}, game.AbilityContent{}, false
	}
	pattern.Source = game.TriggerSourceAny
	pattern.Subject = game.TriggerSubjectDefault
	pattern.DamageSourceCaptured = true
	consumed := ctx
	consumed.content.References = nil
	consumed.content.Targets = nil
	if consumed.content.Unconsumed() {
		return game.ModifyPT{}, game.AbilityContent{}, false
	}
	key := game.LinkedKey(fmt.Sprintf("delayed-target-%d", effectIndex))
	object, ok := lowerObjectReference(ctx.content.References[0], referenceLoweringContext{
		TargetLinkedKey: key,
	})
	if !ok {
		return game.ModifyPT{}, game.AbilityContent{}, false
	}
	modify.PublishLinked = key
	delayed := game.CreateDelayedTrigger{Trigger: game.DelayedTriggerDef{
		EventPattern:       opt.Val(pattern),
		Window:             game.DelayedWindowThisTurn,
		Content:            triggered.Content,
		DamageSourceObject: opt.Val(object),
	}}
	return modify, game.Mode{Sequence: []game.Instruction{{Primitive: delayed}}}.Ability(), true
}

// lowerDelayedAttacksMonarchGrant lowers a captured-object attacks-the-monarch
// rider clause ("... put a +1/+1 counter on target creature ... Whenever that
// creature attacks the monarch this turn, it gains double strike and trample
// until end of turn.") into the publishing counter placement and a
// CreateDelayedTrigger whose attacker-declared event binds to that same
// permanent. The prior clause must be an AddCounter placing a +1/+1 counter on
// the shared target; this lowerer publishes that permanent under a linked key
// and rebinds the inner self-form attacks-the-monarch pattern to the captured
// object so the trigger fires only when that specific permanent attacks the
// monarch. It mirrors lowerDelayedCombatDamageDrawTrigger and fails closed on
// any other shape.
func lowerDelayedAttacksMonarchGrant(
	effectIndex int,
	ctx contentCtx,
	sequence []game.Instruction,
) (game.AddCounter, game.AbilityContent, bool) {
	if effectIndex == 0 ||
		len(sequence) != effectIndex ||
		len(ctx.content.Effects) != 1 ||
		ctx.content.Effects[0].Kind != compiler.EffectDelayedTrigger ||
		!ctx.content.Effects[0].DelayedTriggerBindAttacker ||
		ctx.content.Effects[0].DelayedTriggerAbility == nil ||
		ctx.content.Effects[0].Negated ||
		ctx.optional ||
		!referencesBindTo(ctx.content.References, compiler.ReferenceBindingTarget, 0) {
		return game.AddCounter{}, game.AbilityContent{}, false
	}
	previous := sequence[effectIndex-1].Primitive
	if previous.Kind() != game.PrimitiveAddCounter {
		return game.AddCounter{}, game.AbilityContent{}, false
	}
	add, ok := previous.(game.AddCounter)
	if !ok ||
		add.Object.Kind() != game.ObjectReferenceTargetPermanent ||
		add.CounterKind != counter.PlusOnePlusOne ||
		add.PublishLinked != "" ||
		!add.Group.Empty() {
		return game.AddCounter{}, game.AbilityContent{}, false
	}
	triggered, ok := lowerDelayedTriggerInner(ctx.content.Effects[0].DelayedTriggerAbility)
	if !ok {
		return game.AddCounter{}, game.AbilityContent{}, false
	}
	pattern := triggered.Trigger.Pattern
	if pattern.Event != game.EventAttackerDeclared ||
		pattern.Source != game.TriggerSourceSelf ||
		pattern.Player != game.TriggerPlayerMonarch ||
		pattern.AttackRecipient != game.AttackRecipientPlayer ||
		pattern.AttackerCaptured ||
		pattern.AttackAlone {
		return game.AddCounter{}, game.AbilityContent{}, false
	}
	pattern.Source = game.TriggerSourceAny
	pattern.Subject = game.TriggerSubjectDefault
	pattern.AttackerCaptured = true
	consumed := ctx
	consumed.content.References = nil
	consumed.content.Targets = nil
	if consumed.content.Unconsumed() {
		return game.AddCounter{}, game.AbilityContent{}, false
	}
	key := game.LinkedKey(fmt.Sprintf("delayed-target-%d", effectIndex))
	object, ok := lowerObjectReference(ctx.content.References[0], referenceLoweringContext{
		TargetLinkedKey: key,
	})
	if !ok {
		return game.AddCounter{}, game.AbilityContent{}, false
	}
	add.PublishLinked = key
	delayed := game.CreateDelayedTrigger{Trigger: game.DelayedTriggerDef{
		EventPattern:           opt.Val(pattern),
		Window:                 game.DelayedWindowThisTurn,
		Content:                triggered.Content,
		CapturedAttackerObject: opt.Val(object),
	}}
	return add, game.Mode{Sequence: []game.Instruction{{Primitive: delayed}}}.Ability(), true
}
