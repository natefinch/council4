package rules

import (
	"slices"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
)

// spellTypeLine is a spell's effective supertypes, card types and subtypes.
type spellTypeLine struct {
	supertypes []types.Super
	types      []types.Card
	subtypes   []types.Sub
}

// spellTypeLineFor returns the type line a spell controlled by controller
// presents to every rule that reads it: targeting, protection, "can't be
// countered", spell-cast events, cost modifiers, cast permissions and
// prohibitions. The base is def (already the face-down, copy-overlaid or
// selected face) as its own characteristic-defining abilities define it
// (CR 604.3), transformed by a bestowed cast (CR 702.103b). Battlefield static
// abilities whose type-changing effects declare a SpellSelection then apply in
// layer order (CR 613.1d, 613.7), each to the spells its controller controls
// that match the selection as of the start of the type layer. Nothing is stored
// on the stack object: the result tracks each source's current presence,
// controller and condition.
func spellTypeLineFor(g *game.Game, controller game.PlayerID, def *game.CardDef, bestowed bool) spellTypeLine {
	face := characteristicDefinedFace(def)
	if face == nil {
		return spellTypeLine{}
	}
	line := spellTypeLine{
		supertypes: slices.Clone(face.Supertypes),
		types:      slices.Clone(face.Types),
		subtypes:   slices.Clone(face.Subtypes),
	}
	if bestowed {
		line.types = game.BestowSpellTypes(line.types)
		line.subtypes = game.BestowSpellSubtypes(line.subtypes)
	}
	if g == nil {
		return line
	}
	effects := spellTypeEffects(g, controller, line)
	if len(effects) == 0 {
		return line
	}
	values := permanentEffectiveValues{
		controller: controller,
		supertypes: line.supertypes,
		types:      line.types,
		subtypes:   line.subtypes,
	}
	ordered := orderContinuousEffects(effects)
	for i := range ordered {
		applyTypeLayer(g, &values, &ordered[i])
	}
	return spellTypeLine{supertypes: values.supertypes, types: values.types, subtypes: values.subtypes}
}

// stackSpellTypeLine returns the effective type line of a spell stack object
// whose base characteristics are def. Non-spell stack objects keep def's own
// characteristic-defined type line.
func stackSpellTypeLine(g *game.Game, obj *game.StackObject, def *game.CardDef) spellTypeLine {
	if obj == nil || obj.Kind != game.StackSpell {
		return spellTypeLineFor(nil, 0, def, false)
	}
	return spellTypeLineFor(g, obj.Controller, def, obj.Bestowed)
}

// withSpellTypeLine returns a shallow copy of def carrying line as its
// supertypes, types and subtypes. def is never mutated.
func withSpellTypeLine(def *game.CardDef, line spellTypeLine) *game.CardDef {
	if def == nil {
		return nil
	}
	effective := *def
	effective.Supertypes = line.supertypes
	effective.Types = line.types
	effective.Subtypes = line.subtypes
	return &effective
}

// spellTypeEffects collects the type-changing effects of battlefield static
// abilities that apply to a spell controlled by controller whose type line at
// the start of the type layer is line. Each effect is stamped with its source,
// controller and timestamp so ordering and source entry choices (Leyline of
// Transformation's chosen creature type) resolve as they do for permanents.
func spellTypeEffects(g *game.Game, controller game.PlayerID, line spellTypeLine) []game.ContinuousEffect {
	if _, owned := beginStaticSourceRead(g); owned {
		defer g.EndStaticSourceFrame()
	}
	subject := selectionSubject{
		kind:      subjectCastSpell,
		g:         g,
		cardTypes: line.types,
		event:     game.Event{CardSupertypes: line.supertypes, CardSubtypes: line.subtypes},
	}
	var effects []game.ContinuousEffect
	for _, source := range staticAbilitySources(g) {
		if source.face == nil {
			continue
		}
		resolvedController := false
		for i := range source.face.StaticAbilities {
			body := &source.face.StaticAbilities[i]
			if !staticAbilityDeclaresSpellTypeEffect(body) || !staticAbilityFunctionsFromSource(body, source) {
				continue
			}
			if !resolvedController {
				if source.permanent != nil {
					source.controller = effectiveController(g, source.permanent)
				}
				resolvedController = true
			}
			if source.controller != controller {
				continue
			}
			if !conditionSatisfied(g, conditionContext{
				controller:            source.controller,
				source:                source.permanent,
				characteristicsBefore: game.LayerType,
			}, body.Condition) {
				continue
			}
			for j := range body.ContinuousEffects {
				template := &body.ContinuousEffects[j]
				if template.Layer != game.LayerType || template.SpellSelection == nil ||
					!matchSelection(&subject, template.SpellSelection) {
					continue
				}
				effect := *template
				effect.SourceObjectID = sourceObjectID(source)
				effect.SourceCardID = source.cardID
				effect.Controller = source.controller
				effect.Timestamp = source.timestamp
				effects = append(effects, effect)
			}
		}
	}
	return effects
}

func staticAbilityDeclaresSpellTypeEffect(body *game.StaticAbility) bool {
	for i := range body.ContinuousEffects {
		if body.ContinuousEffects[i].SpellSelection != nil {
			return true
		}
	}
	return false
}
