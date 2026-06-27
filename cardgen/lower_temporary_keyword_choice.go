package cardgen

import (
	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/opt"
)

// keywordChoiceGrantModes builds one mode per listed keyword, each granting that
// single keyword to the shared object for the given duration. The controller
// picks exactly one mode at resolution, which realizes the "your choice of one of
// the listed keywords" semantics with the existing modal machinery. It backs both
// the indefinite (DurationPermanent) and until-end-of-turn keyword-choice grants.
func keywordChoiceGrantModes(
	keywords []game.Keyword,
	object game.ObjectReference,
	duration game.EffectDuration,
) []game.Mode {
	modes := make([]game.Mode, 0, len(keywords))
	for _, keyword := range keywords {
		modes = append(modes, game.Mode{
			Sequence: []game.Instruction{{
				Primitive: game.ApplyContinuous{
					Object: opt.Val(object),
					ContinuousEffects: []game.ContinuousEffect{{
						Layer:       game.LayerAbility,
						AddKeywords: []game.Keyword{keyword},
					}},
					Duration: duration,
				},
			}},
		})
	}
	return modes
}

// keywordChoiceGrantContent assembles a one-of-N modal keyword-choice grant. The
// listed keywords each become a single mode granting that keyword to object for
// the given duration, the controller selects exactly one, and an optional shared
// target slot is carried across every mode. Abilities (such as a granted
// protection static body) are never produced by a choice list, so a non-empty
// abilities slice is rejected as unrepresentable; a list of fewer than two
// keywords is likewise rejected so a degenerate "choice" never reaches the modal
// machinery.
func keywordChoiceGrantContent(
	keywords []game.Keyword,
	abilities []game.Ability,
	object game.ObjectReference,
	target opt.V[game.TargetSpec],
	duration game.EffectDuration,
) (game.AbilityContent, *shared.Diagnostic) {
	if len(keywords) < 2 || len(abilities) != 0 {
		return game.AbilityContent{}, &shared.Diagnostic{
			Severity: shared.SeverityWarning,
			Summary:  "unsupported keyword choice grant",
			Detail:   "the executable source backend supports only a choice among two or more simple grantable keywords",
		}
	}
	content := game.AbilityContent{
		Modes:    keywordChoiceGrantModes(keywords, object, duration),
		MinModes: 1,
		MaxModes: 1,
	}
	if target.Exists {
		content.SharedTargets = []game.TargetSpec{target.Val}
	}
	return content, nil
}

// lowerTemporaryKeywordChoiceGrant lowers an until-end-of-turn disjunctive
// keyword grant ("This creature gains your choice of vigilance, lifelink, or
// haste until end of turn.", "Target creature gains your choice of lifelink or
// indestructible until end of turn.") into a one-of-N modal grant whose modes
// each add one of the listed keywords until end of turn. It accepts the same
// single-permanent subjects the conjunctive grant does — a single exact target,
// the source permanent, the triggering event permanent, or a referenced target
// object — and fails closed for any group, plural, or quoted-ability shape the
// modal choice cannot represent.
func lowerTemporaryKeywordChoiceGrant(
	ctx contentCtx,
	keywords []game.Keyword,
	abilities []game.Ability,
	targetSubject, sourceSubject, eventPermanentSubject bool,
	unsupported func() (game.AbilityContent, *shared.Diagnostic),
) (game.AbilityContent, *shared.Diagnostic) {
	if targetSubject {
		spec, ok := permanentTargetSpec(ctx.content.Targets[0])
		if !ok {
			return unsupported()
		}
		return keywordChoiceGrantContent(
			keywords,
			abilities,
			game.TargetPermanentReference(0),
			opt.Val(spec),
			game.DurationUntilEndOfTurn,
		)
	}
	var object game.ObjectReference
	var ok bool
	switch {
	case sourceSubject:
		object, ok = lowerObjectReference(ctx.content.References[0], referenceLoweringContext{
			AllowSource:      true,
			SourceCardObject: true,
		})
	case eventPermanentSubject:
		object, ok = lowerObjectReference(ctx.content.References[0], referenceLoweringContext{AllowEvent: true})
	default:
		object, ok = lowerObjectReference(ctx.content.References[0], referenceLoweringContext{AllowTarget: true})
	}
	if !ok {
		return unsupported()
	}
	return keywordChoiceGrantContent(
		keywords,
		abilities,
		object,
		opt.V[game.TargetSpec]{},
		game.DurationUntilEndOfTurn,
	)
}
