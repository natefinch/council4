package rules

import (
	"fmt"
	"slices"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// preventionSourceCandidate is one source a one-shot "a source of your choice"
// prevention shield may be bound to (CR 609.7a), identified by the object ID
// its damage events carry.
type preventionSourceCandidate struct {
	objectID   id.ID
	controller game.PlayerID
	label      string
	card       opt.V[game.ChoiceCardInfo]
}

// oneShotPreventionSourceFilter resolves the source filter of a one-shot
// shield as it is created. "Of the chosen color" reads the color the shield's
// source permanent chose as it entered and fixes it now; ok is false when that
// choice is missing, so no shield is created.
func oneShotPreventionSourceFilter(g *game.Game, obj *game.StackObject, prim game.PreventDamage) ([]color.Color, bool) {
	colors := slices.Clone(prim.SourceColors)
	if !prim.SourceColorFromEntryChoice {
		return colors, true
	}
	chosen, ok := sourceEntryChosenManaColor(abilitySourceLastKnownPermanent(g, obj))
	if !ok {
		return nil, false
	}
	entryColor, ok := manaColor(chosen)
	if !ok {
		return nil, false
	}
	if len(colors) > 0 && !slices.Contains(colors, entryColor) {
		return nil, false
	}
	return []color.Color{entryColor}, true
}

// preventionSourceCandidates lists the sources the controller of a one-shot
// shield may choose (CR 609.7a): spells on the stack and the sources of
// abilities on the stack, top first, then permanents on the battlefield, each
// matching the colors and card types required now. The resolving object, which
// remains on the stack until it finishes resolving (CR 608.2n), is listed first
// so a spell may choose itself or an ability its own source. A source is listed
// once under the object ID its damage would carry.
func preventionSourceCandidates(g *game.Game, resolving *game.StackObject, colors []color.Color, cardTypes []types.Card) []preventionSourceCandidate {
	var candidates []preventionSourceCandidate
	seen := make(map[id.ID]bool)
	add := func(candidate preventionSourceCandidate, sourceColors []color.Color, sourceTypes []types.Card) {
		if candidate.objectID == 0 || seen[candidate.objectID] ||
			!preventionSourceCharacteristicsMatch(sourceColors, sourceTypes, colors, cardTypes) {
			return
		}
		seen[candidate.objectID] = true
		candidates = append(candidates, candidate)
	}
	if resolving != nil {
		addStackPreventionSource(g, add, resolving)
	}
	objects := g.Stack.Objects()
	for i := len(objects) - 1; i >= 0; i-- {
		addStackPreventionSource(g, add, objects[i])
	}
	for _, permanent := range g.Battlefield {
		if activeBattlefieldPermanent(permanent) {
			addPermanentPreventionSource(g, add, permanent)
		}
	}
	return candidates
}

// addStackPreventionSource offers a stack object as a prevention source: a
// spell itself, or an ability's source.
func addStackPreventionSource(
	g *game.Game,
	add func(preventionSourceCandidate, []color.Color, []types.Card),
	obj *game.StackObject,
) {
	if obj.Kind == game.StackSpell {
		snapshot := snapshotStackSpell(g, obj)
		name := "Face-down spell"
		info := game.ChoiceCardInfo{Types: slices.Clone(snapshot.Types), Colors: slices.Clone(snapshot.Colors)}
		if def, ok := stackSpellCharacteristicsDef(g, obj); ok && !obj.FaceDown {
			name = def.Name
			info.CardID = obj.SourceID
			info.Name = def.Name
			info.ManaValue = def.ManaValue()
		}
		add(preventionSourceCandidate{
			objectID: obj.ID, controller: obj.Controller,
			label: fmt.Sprintf("%s [%d]", name, obj.ID), card: opt.Val(info),
		}, snapshot.Colors, snapshot.Types)
		return
	}
	sourceID, sourceObjectID := damageSourceIDs(g, obj)
	if permanent, ok := permanentByObjectID(g, sourceObjectID); ok {
		addPermanentPreventionSource(g, add, permanent)
		return
	}
	event := damageEvent{sourceID: sourceID, sourceObjectID: sourceObjectID}
	name := "Ability source"
	var info opt.V[game.ChoiceCardInfo]
	if snapshot, ok := lastKnownObject(g, sourceObjectID); ok {
		if !snapshot.FaceDown && snapshot.Name != "" {
			name = snapshot.Name
		}
		info = opt.Val(game.ChoiceCardInfo{
			CardID: snapshot.CardID, Name: snapshot.Name,
			Types: slices.Clone(snapshot.Types), Colors: slices.Clone(snapshot.Colors),
		})
		if snapshot.FaceDown {
			info.Val.Name = ""
		}
	} else if card := cardChoiceInfo(g, sourceID); card.Exists {
		name, info = card.Val.Name, card
	}
	add(preventionSourceCandidate{
		objectID: sourceObjectID, controller: obj.Controller,
		label: fmt.Sprintf("%s [%d]", name, sourceObjectID), card: info,
	}, damageSourceColors(g, event), damageSourceTypes(g, event))
}

func addPermanentPreventionSource(
	g *game.Game,
	add func(preventionSourceCandidate, []color.Color, []types.Card),
	permanent *game.Permanent,
) {
	values := effectivePermanentValues(g, permanent)
	add(preventionSourceCandidate{
		objectID: permanent.ObjectID, controller: effectiveController(g, permanent),
		label: fmt.Sprintf("%s [%d]", permanentChoiceLabel(g, permanent), permanent.ObjectID),
		card:  permanentChoiceInfo(g, permanent),
	}, values.colors, values.types)
}

// preventionSourceCharacteristicsMatch reports whether a source with
// sourceColors and sourceTypes satisfies a shield's "<color> source" and
// "<type> source" filters; an empty filter matches any source.
func preventionSourceCharacteristicsMatch(sourceColors []color.Color, sourceTypes []types.Card, colors []color.Color, cardTypes []types.Card) bool {
	if len(colors) > 0 && !slices.ContainsFunc(colors, func(c color.Color) bool { return slices.Contains(sourceColors, c) }) {
		return false
	}
	return len(cardTypes) == 0 || slices.ContainsFunc(cardTypes, func(t types.Card) bool { return slices.Contains(sourceTypes, t) })
}

// chooseOneShotPreventionSource has the shield's controller choose the source
// a one-shot shield applies to as the shield is created (CR 609.7a). resolving
// is the stack object creating the shield. The
// deterministic default is the first listed source an opponent controls, else
// the first source. ok is false when no source can be chosen.
func (e *Engine) chooseOneShotPreventionSource(
	g *game.Game,
	agents [game.NumPlayers]PlayerAgent,
	log *TurnLog,
	resolving *game.StackObject,
	chooser game.PlayerID,
	colors []color.Color,
	cardTypes []types.Card,
) (id.ID, bool) {
	candidates := preventionSourceCandidates(g, resolving, colors, cardTypes)
	if len(candidates) == 0 {
		return 0, false
	}
	selected := 0
	if index := slices.IndexFunc(candidates, func(candidate preventionSourceCandidate) bool {
		return candidate.controller != chooser
	}); index >= 0 {
		selected = index
	}
	if len(candidates) > 1 {
		if e == nil {
			e = NewEngine(nil)
		}
		options := make([]game.ChoiceOption, len(candidates))
		for index, candidate := range candidates {
			options[index] = game.ChoiceOption{Index: index, Label: candidate.label, Card: candidate.card}
		}
		selected = e.chooseChoice(g, agents, game.ChoiceRequest{
			Kind: game.ChoicePreventionAllocation, Player: chooser,
			Prompt:  "Choose the source whose next damage to prevent this turn.",
			Options: options, MinChoices: 1, MaxChoices: 1, DefaultSelection: []int{selected},
		}, log)[0]
	}
	return candidates[selected].objectID, true
}

// damageSourceTypes returns the card types of a damage event's source: a
// battlefield source's current types, else its last-known types, else its
// card's printed types.
func damageSourceTypes(g *game.Game, event damageEvent) []types.Card {
	if event.sourceObjectID != 0 {
		if permanent, ok := permanentByObjectID(g, event.sourceObjectID); ok {
			return slices.Clone(effectivePermanentValues(g, permanent).types)
		}
		if snapshot, ok := lastKnownObject(g, event.sourceObjectID); ok {
			return slices.Clone(snapshot.Types)
		}
	}
	if source, ok := damageSourceDef(g, event.sourceID, event.sourceObjectID); ok && source != nil {
		return slices.Clone(source.Types)
	}
	return nil
}

// preventionShieldSourceMatches reports whether a damage event's source still
// satisfies a shield's color and card-type filters. A one-shot shield checks
// its chosen source again as the damage would be dealt (CR 609.7b).
func preventionShieldSourceMatches(g *game.Game, shield *game.PreventionShield, event damageEvent) bool {
	if len(shield.SourceColors) == 0 && len(shield.SourceTypes) == 0 {
		return true
	}
	return preventionSourceCharacteristicsMatch(damageSourceColors(g, event), damageSourceTypes(g, event),
		shield.SourceColors, shield.SourceTypes)
}

// rebindPermanentSpellPreventionShields moves one-shot shields bound to a
// resolving permanent spell onto the permanent it became: an effect that
// prevents damage from a permanent spell also prevents damage from the
// resulting permanent (CR 400.7c, 609.7a).
func rebindPermanentSpellPreventionShields(g *game.Game, obj *game.StackObject, permanent *game.Permanent) {
	if obj == nil || obj.ID == 0 || permanent == nil {
		return
	}
	for i := range g.PreventionShields {
		if g.PreventionShields[i].ChosenSourceObjectID == obj.ID {
			g.PreventionShields[i].ChosenSourceObjectID = permanent.ObjectID
		}
	}
}
