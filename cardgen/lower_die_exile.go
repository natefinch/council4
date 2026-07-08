package cardgen

import (
	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/zone"
)

// lowerDamageDieToExileSpellAbilities lowers a single-target damage or -X/-X
// spell that carries the rider "If that creature [or planeswalker] would die
// this turn, exile it instead." (Lava Coil, Obliterating Bolt, Magma Spray,
// Flame-Blessed Bolt, Bleed Dry, Ob Nixilis's Cruelty, ...). The whole card is
// one ability the parser classifies as a replacement because it reads
// "would ... instead", but the rider only redirects the spell's single target's
// death to exile for the turn. The lowering strips the trailing
// EffectExileIfWouldDieThisTurn effect, lowers the residual spell (damage or
// -X/-X) through the standard spell path so its target and amount resolution are
// shared with every other spell, then appends a CreateReplacement bound to that
// target. The replacement redirects the target's battlefield-to-graveyard zone
// change (its death) to exile and expires at end of turn. It fails closed for
// every other shape — multiple riders, modal or shared-target spells, a residual
// the spell path cannot lower, or a target the rider does not bind to.
func lowerDamageDieToExileSpellAbilities(cardName string, compilation compiler.Compilation) (game.AbilityContent, bool) {
	if len(compilation.Abilities) != 1 || len(compilation.Syntax.Abilities) != 1 {
		return game.AbilityContent{}, false
	}
	ability := compilation.Abilities[0]
	if ability.Trigger != nil || ability.Cost != nil || ability.Static != nil || ability.Optional {
		return game.AbilityContent{}, false
	}
	content := ability.Content
	effectCount := len(content.Effects)
	if effectCount < 2 || len(content.Modes) != 0 || len(content.Targets) != 1 {
		return game.AbilityContent{}, false
	}
	exile := content.Effects[effectCount-1]
	if !isDieToExileReplacementEffect(&exile) {
		return game.AbilityContent{}, false
	}
	residual := content
	residual.Effects = append([]compiler.CompiledEffect(nil), content.Effects[:effectCount-1]...)
	// The shared ordered-sequence pass marks every effect of a multi-effect
	// ability as requiring ordered lowering. With the exile rider removed a lone
	// residual effect (a single damage or -X/-X clause) is a standalone spell
	// again, so clear the flag to let the single-effect spell path lower it.
	if len(residual.Effects) == 1 {
		residual.Effects[0].RequiresOrderedLowering = false
	}
	residual.References = referencesOutsideSpan(content.References, exile.Span)
	spell, diagnostic := lowerSpellAbilityContent(cardName, residual, ability.Optional, &compilation.Syntax.Abilities[0])
	if diagnostic != nil {
		return game.AbilityContent{}, false
	}
	if spell.IsModal() ||
		len(spell.SharedTargets) != 0 ||
		len(spell.Modes) != 1 ||
		len(spell.Modes[0].Targets) != 1 ||
		len(spell.Modes[0].Sequence) == 0 {
		return game.AbilityContent{}, false
	}
	create := game.CreateReplacement{
		Object:   game.TargetPermanentReference(0),
		Duration: game.DurationThisTurn,
		Replacement: &game.ReplacementEffect{
			MatchEvent:                   game.EventZoneChanged,
			MatchFromZone:                true,
			FromZone:                     zone.Battlefield,
			MatchToZone:                  true,
			ToZone:                       zone.Graveyard,
			ReplaceToZone:                zone.Exile,
			AffectedObjectMustBeCreature: exile.ExileDieSubjectDamagedCreature,
		},
	}
	spell.Modes[0].Sequence = append(spell.Modes[0].Sequence, game.Instruction{Primitive: create})
	return spell, true
}

// isDieToExileReplacementEffect reports whether effect is the exact would-die
// exile rider "If that creature [or planeswalker] would die this turn, exile it
// instead." whose subject binds to the spell's single target, or the burn
// variant "If a creature dealt damage this way would die this turn, exile it
// instead." whose "it" denotes the spell's single damaged target when it is a
// creature (ExileDieSubjectDamagedCreature).
func isDieToExileReplacementEffect(effect *compiler.CompiledEffect) bool {
	if effect.Kind != compiler.EffectExileIfWouldDieThisTurn ||
		effect.Negated ||
		!effect.Exact {
		return false
	}
	if effect.ExileDieSubjectDamagedCreature {
		return true
	}
	return referencesBindTo(effect.References, compiler.ReferenceBindingTarget, 0)
}
