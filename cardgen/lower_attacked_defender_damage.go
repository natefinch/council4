package cardgen

import (
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game"
)

// lowerAttackedDefenderDamageSpell lowers a "deals N damage to the player or
// planeswalker it's attacking" effect whose recipient is the defending player
// of the triggering attack, as in "Whenever this creature attacks, it deals 1
// damage to the player or planeswalker it's attacking." (Scorch Spitter,
// Cavalcade of Calamity). The recipient resolves to the triggering attack
// event's defending player through DefendingPlayerReference, so lowering gates
// the enclosing trigger to an attack event the runtime populates with a
// defending player. The amount is a fixed value, X, or a resolvable
// source-power dynamic amount. It emits one Damage instruction with the
// defending-player recipient and no target spec, failing closed for any shape
// outside that template (a recipient that is not the attacked defender, a
// non-attack trigger, an unresolvable amount, any target, recipient selector,
// condition, keyword, or mode).
func lowerAttackedDefenderDamageSpell(ctx contentCtx) (game.AbilityContent, *shared.Diagnostic) {
	effect := ctx.content.Effects[0]
	assertUndividedRecipientDamageDispatch(ctx, parser.DamageRecipientReferenceAttackedDefender)
	unsupported := func() (game.AbilityContent, *shared.Diagnostic) {
		return game.AbilityContent{}, contentDiagnostic(
			ctx,
			"unsupported damage spell",
			"the executable source backend supports only exact fixed, X, or source-power damage to the attacked player or planeswalker",
		)
	}
	if !effect.Exact ||
		effect.Negated ||
		len(ctx.content.Targets) != 0 ||
		len(effect.DamageRecipient.GroupSelectors) != 0 ||
		len(ctx.content.Conditions) != 0 ||
		len(abilityKeywordsExcludingSelectorPredicates(ctx.content)) != 0 ||
		len(ctx.content.Modes) != 0 {
		return unsupported()
	}
	if !attackDefendingPlayerEvent(ctx.triggerEvent) {
		return unsupported()
	}
	// The damage subject heads the reference list (the attacking source dealing
	// the damage); any trailing reference is the recipient anaphor ("that
	// creature is attacking") naming the attacker again, which the typed
	// defending-player recipient already resolves from the event. Only the
	// leading subject reference feeds the damage-source and amount exactness
	// checks, so the recipient anaphor cannot bleed into them.
	references := ctx.content.References
	if len(references) == 0 || !exactDamageSourceSyntax(references[:1]) {
		return unsupported()
	}
	amount, ok := eventPlayerDamageAmount(effect, references[:1])
	if !ok {
		return unsupported()
	}
	damage := game.Damage{
		Amount:       amount,
		Recipient:    game.PlayerDamageRecipient(game.DefendingPlayerReference()),
		DamageSource: primaryDamageSource(references[:1]),
	}
	return game.Mode{
		Sequence: []game.Instruction{{Primitive: damage}},
	}.Ability(), nil
}

// attackDefendingPlayerEvent reports whether the trigger event kind is one whose
// triggering event carries a defending player the runtime resolves through
// DefendingPlayerReference: the attack-declared, became-blocked, and
// became-unblocked combat events. It fails closed for every other event so the
// attacked-defender recipient lowers only where the defending player resolves.
func attackDefendingPlayerEvent(kind game.EventKind) bool {
	return kind == game.EventAttackerDeclared ||
		kind == game.EventAttackerBecameBlocked ||
		kind == game.EventAttackerBecameUnblocked
}
