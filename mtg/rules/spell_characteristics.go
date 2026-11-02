package rules

import (
	"slices"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
)

// spellTypeLine is a spell's effective supertypes, card types and subtypes.
type spellTypeLine struct {
	supertypes []types.Super
	types      []types.Card
	subtypes   []types.Sub
}

// spellTypeScope names whose type-changing effects reach a spell. Effects that
// apply to "spells you control" reach it through controller. A spell is a card
// on the stack (CR 112.1), so effects that apply to "cards you own that aren't
// on the battlefield" reach it through the owner of card, its CardInstance ID.
// card is 0 for a copy, which isn't a card (CR 112.1a, 707.10), and for a
// spell with no card.
type spellTypeScope struct {
	controller game.PlayerID
	card       id.ID
}

// castSpellTypeScope is the scope of caster's spell that is, or will be, card.
func castSpellTypeScope(caster game.PlayerID, card id.ID) spellTypeScope {
	return spellTypeScope{controller: caster, card: card}
}

// stackSpellTypeScope is the scope of obj, a spell on the stack.
func stackSpellTypeScope(obj *game.StackObject) spellTypeScope {
	scope := spellTypeScope{controller: obj.Controller}
	if obj.Kind == game.StackSpell && !obj.Copy {
		scope.card = obj.SourceID
	}
	return scope
}

// spellTypeLineFor returns the type line a spell in scope presents to every
// rule that reads it: targeting, protection, "can't be countered", spell-cast
// events, cost modifiers, cast permissions and prohibitions. The base is def
// (already the face-down, copy-overlaid or selected face) as its own
// characteristic-defining abilities define it (CR 604.3), transformed by a
// bestowed cast (CR 702.103b). Battlefield static abilities whose type-changing
// effects declare a SpellSelection (for spells their controller controls) or an
// OwnedCardSelection (for a card-backed spell their controller owns) then apply
// in layer order (CR 613.1d, 613.7), each matched against the spell's type line
// as of the start of the type layer. An effect that reaches the spell through
// both clauses applies once. Nothing is stored on the stack object: the result
// tracks each source's current presence, controller and condition.
func spellTypeLineFor(g *game.Game, scope spellTypeScope, def *game.CardDef, bestowed bool) spellTypeLine {
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
	reaches := []typeLineReach{{player: scope.controller, selectionOf: spellTypeSelection}}
	if owner, ok := spellCardOwner(g, scope.card); ok {
		reaches = append(reaches, typeLineReach{player: owner, selectionOf: ownedCardTypeSelection})
	}
	return applyTypeLineEffects(g, scope.controller, line, typeLineEffects(g, line, reaches...))
}

// spellCardOwner returns the owner of card, the card a spell is or is being
// cast as. A card on the battlefield is not one: it is the permanent whose
// spell copy is cast (a prepared permanent), and the copy isn't a card.
func spellCardOwner(g *game.Game, card id.ID) (game.PlayerID, bool) {
	if card == 0 {
		return 0, false
	}
	instance, ok := g.GetCardInstance(card)
	if !ok {
		return 0, false
	}
	for _, permanent := range g.Battlefield {
		if permanent != nil && permanent.CardInstanceID == card {
			return 0, false
		}
	}
	return instance.Owner, true
}

// applyTypeLineEffects applies the collected layer 4 effects to line in
// timestamp order (CR 613.7) through the battlefield type layer. line must be
// owned by the caller because the type layer may append to its slices.
func applyTypeLineEffects(g *game.Game, controller game.PlayerID, line spellTypeLine, effects []game.ContinuousEffect) spellTypeLine {
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
		return spellTypeLineFor(nil, spellTypeScope{}, def, false)
	}
	return spellTypeLineFor(g, stackSpellTypeScope(obj), def, obj.Bestowed)
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

// typeLineSelection picks the selection that extends a battlefield type-changing
// effect to an object outside the battlefield (a spell or an owned card).
type typeLineSelection func(*game.ContinuousEffect) *game.Selection

func spellTypeSelection(effect *game.ContinuousEffect) *game.Selection {
	return effect.SpellSelection
}

func ownedCardTypeSelection(effect *game.ContinuousEffect) *game.Selection {
	return effect.OwnedCardSelection
}

// typeLineReach extends battlefield type-changing effects controlled by player
// to an object outside the battlefield through the selection selectionOf picks.
type typeLineReach struct {
	player      game.PlayerID
	selectionOf typeLineSelection
}

// typeLineEffects collects the type-changing effects of battlefield static
// abilities that reach an object whose type line at the start of the type layer
// is line: an effect reaches it when, for some reach, the effect's controller is
// that reach's player and the selection that reach picks matches line. Each
// effect is collected once however many reaches it matches, stamped with its
// source, controller and timestamp so ordering and source entry choices (Leyline
// of Transformation's chosen creature type) resolve as they do for permanents.
func typeLineEffects(g *game.Game, line spellTypeLine, reaches ...typeLineReach) []game.ContinuousEffect {
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
			if !staticAbilityDeclaresTypeLineEffect(body, reaches) || !staticAbilityFunctionsFromSource(body, source) {
				continue
			}
			if !resolvedController {
				if source.permanent != nil {
					source.controller = effectiveController(g, source.permanent)
				}
				resolvedController = true
			}
			if !slices.ContainsFunc(reaches, func(reach typeLineReach) bool { return reach.player == source.controller }) {
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
				if template.Layer != game.LayerType || !typeLineEffectReaches(template, source.controller, &subject, reaches) {
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

func typeLineEffectReaches(template *game.ContinuousEffect, controller game.PlayerID, subject *selectionSubject, reaches []typeLineReach) bool {
	for _, reach := range reaches {
		if reach.player != controller {
			continue
		}
		if selection := reach.selectionOf(template); selection != nil && matchSelection(subject, selection) {
			return true
		}
	}
	return false
}

func staticAbilityDeclaresTypeLineEffect(body *game.StaticAbility, reaches []typeLineReach) bool {
	for i := range body.ContinuousEffects {
		for _, reach := range reaches {
			if reach.selectionOf(&body.ContinuousEffects[i]) != nil {
				return true
			}
		}
	}
	return false
}
