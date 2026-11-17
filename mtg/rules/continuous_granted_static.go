package rules

import (
	"slices"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/id"
)

// grantedStaticAbility is a static ability a layer-6 effect granted a
// permanent whose continuous effects modify that permanent itself, such as
// Chimeric Mass's "This creature's power and toughness are each equal to the
// number of charge counters on it.", or other objects, such as Mileva, the
// Stalwart's "Other creatures you control have indestructible." Such an
// ability is not a characteristic-defining ability (CR 604.3a), so its effects
// apply in their own layers with the timestamp CR 613.7a gives them.
type grantedStaticAbility struct {
	key          grantedStaticKey
	body         *game.StaticAbility
	sourceCardID id.ID
	timestamp    game.Timestamp
	// scheduled reports that the ability's layer-6 effects have been
	// considered. started reports that one of them applied, so the ability's
	// later-layer effects continue even if it is then removed (CR 613.6).
	scheduled bool
	started   bool
	// removed reports that an ability-removing effect took the ability away
	// after its effects on the permanent began; those effects continue, but its
	// effects on other objects that had not begun no longer exist.
	removed bool
}

// grantedHolderKey identifies a granted static ability on the permanent that
// has it.
type grantedHolderKey struct {
	key    grantedStaticKey
	holder id.ID
}

// grantedStaticKey identifies one granted instance of a static ability.
type grantedStaticKey struct {
	body           *game.StaticAbility
	grantID        id.ID
	grantSourceID  id.ID
	grantTimestamp game.Timestamp
	occurrence     int
}

// grantedSelfStatic returns the static ability body when it affects the
// permanent it is on while that permanent is on the battlefield.
func grantedSelfStatic(body game.Ability) (*game.StaticAbility, bool) {
	static, ok := body.(*game.StaticAbility)
	if !ok || !bodyFunctionsOnBattlefield(static) {
		return nil, false
	}
	return static, len(grantedSelfEffectTemplates(static)) > 0
}

// grantedStaticBody returns the static ability body when, while the permanent
// it is on is on the battlefield, its effects modify that permanent or other
// objects in layer 6 or later.
func grantedStaticBody(body game.Ability) (*game.StaticAbility, bool) {
	static, ok := body.(*game.StaticAbility)
	if !ok || !bodyFunctionsOnBattlefield(static) {
		return nil, false
	}
	return static, len(grantedSelfEffectTemplates(static)) > 0 || len(grantedGroupEffectTemplates(static)) > 0
}

// grantedGroupEffectTemplates returns the continuous effects static generates
// on a group of objects, such as "Other creatures you control have
// indestructible" or "Equipped creature gets +3/+3". Only layer 6 and later
// effects are supported: the type and color layers are resolved before the
// ability-layer evaluation that grants the ability.
func grantedGroupEffectTemplates(static *game.StaticAbility) []game.ContinuousEffect {
	var templates []game.ContinuousEffect
	for _, effect := range static.ContinuousEffects {
		if !effect.AffectedSource && !effect.Group.Empty() && effect.Layer >= game.LayerAbility {
			templates = append(templates, effect)
		}
	}
	return templates
}

// conditionalKeywordStatic reports whether body is a static ability that
// gives the permanent it is on keyword abilities only while its condition
// holds, such as Restless Spire's granted "During your turn, this creature
// has first strike."
func conditionalKeywordStatic(body game.Ability) bool {
	static, ok := body.(*game.StaticAbility)
	return ok && static.Condition.Exists && len(static.KeywordAbilities) > 0 &&
		bodyFunctionsOnBattlefield(static)
}

// grantedSelfEffectTemplates returns the continuous effects static generates
// on the permanent that has it. A conditional static's keyword abilities are
// one such layer-6 effect: the permanent has them only while the condition
// holds (CR 611.3a, 613.1f), so they are not added with the ability itself.
func grantedSelfEffectTemplates(static *game.StaticAbility) []game.ContinuousEffect {
	var templates []game.ContinuousEffect
	for _, effect := range static.ContinuousEffects {
		if effect.AffectedSource && effect.Group.Empty() {
			templates = append(templates, effect)
		}
	}
	if conditionalKeywordStatic(static) {
		keywords := make([]game.Keyword, 0, len(static.KeywordAbilities))
		for _, ability := range static.KeywordAbilities {
			keywords = append(keywords, game.KeywordAbilityKind(ability))
		}
		templates = append(templates, game.ContinuousEffect{
			Duration: game.DurationPermanent, AffectedSource: true, Layer: game.LayerAbility,
			AddKeywords: keywords,
		})
	}
	return templates
}

func recordGrantedStaticAbility(values *permanentEffectiveValues, permanent *game.Permanent, body game.Ability, effect *game.ContinuousEffect, occurrence int) {
	static, ok := grantedStaticBody(body)
	if !ok {
		return
	}
	sourceCardID := effect.SourceCardID
	if sourceCardID == 0 {
		sourceCardID = permanent.CardInstanceID
	}
	values.grantedStatics = append(slices.Clip(values.grantedStatics), grantedStaticAbility{
		key: grantedStaticKey{
			body: static, grantID: effect.ID, grantSourceID: effect.SourceObjectID,
			grantTimestamp: effect.Timestamp, occurrence: occurrence,
		},
		body:         static,
		sourceCardID: sourceCardID,
		// CR 613.7a: the later of the object's timestamp and the timestamp of
		// the effect that created the ability.
		timestamp: max(permanent.Timestamp(), effect.Timestamp),
	})
}

// removeUnstartedGrantedStatics drops the granted statics an ability-removing
// effect takes away, keeping those whose effects on the permanent already
// began (CR 613.6).
func removeUnstartedGrantedStatics(values *permanentEffectiveValues) {
	values.grantedStatics = slices.DeleteFunc(slices.Clone(values.grantedStatics), func(granted grantedStaticAbility) bool {
		return !granted.started
	})
	for i := range values.grantedStatics {
		values.grantedStatics[i].removed = true
	}
}

// appendGrantedStaticEffects appends the effects a granted static ability
// currently generates for layer on the permanent that has it. A conditional
// ability is checked live (CR 611.3a); board supplies the layer-6 values of the
// global ability-layer evaluation when that is in progress.
func appendGrantedStaticEffects(effects []game.ContinuousEffect, g *game.Game, permanent *game.Permanent, values *permanentEffectiveValues, granted grantedStaticAbility, layer game.ContinuousLayer, board map[id.ID]permanentEffectiveValues) []game.ContinuousEffect {
	templates := grantedSelfEffectTemplates(granted.body)
	if !slices.ContainsFunc(templates, func(template game.ContinuousEffect) bool { return template.Layer == layer }) {
		return effects
	}
	if !conditionSatisfied(g, conditionContext{
		controller: values.controller, source: permanent,
		characteristicsBefore: layer, layerValues: board,
	}, granted.body.Condition) {
		return effects
	}
	for _, template := range templates {
		if template.Layer != layer {
			continue
		}
		effect := template
		effect.ID = 0
		effect.AffectedObjectID = permanent.ObjectID
		effect.SourceObjectID = permanent.ObjectID
		effect.SourceCardID = granted.sourceCardID
		effect.Controller = values.controller
		effect.Timestamp = granted.timestamp
		effects = append(effects, effect)
	}
	return effects
}

// applyLayerEffects applies one layer's ordered effects to a permanent. In
// layer 6, each static ability an applied effect grants the permanent adds its
// own layer-6 effects on it to the order by their CR 613.7a timestamp.
func applyLayerEffects(g *game.Game, permanent *game.Permanent, values *permanentEffectiveValues, layer game.ContinuousLayer, ordered []game.ContinuousEffect) {
	for i := range ordered {
		if layer == game.LayerAbility {
			applyDueGrantedStaticAbilityEffects(g, permanent, values, &ordered[i])
		}
		applyContinuousEffect(g, permanent, values, &ordered[i])
	}
	if layer == game.LayerAbility {
		applyDueGrantedStaticAbilityEffects(g, permanent, values, nil)
	}
}

// applyDueGrantedStaticAbilityEffects applies the layer-6 effects of granted
// statics that precede next, or of every remaining one when next is nil. A
// generated effect has no runtime ID, so it precedes a runtime effect with the
// same timestamp.
func applyDueGrantedStaticAbilityEffects(g *game.Game, permanent *game.Permanent, values *permanentEffectiveValues, next *game.ContinuousEffect) {
	for {
		index := -1
		for i, granted := range values.grantedStatics {
			if granted.scheduled || next != nil && granted.timestamp > next.Timestamp {
				continue
			}
			if index == -1 || granted.timestamp < values.grantedStatics[index].timestamp {
				index = i
			}
		}
		if index == -1 {
			return
		}
		values.grantedStatics[index].scheduled = true
		effects := appendGrantedStaticEffects(nil, g, permanent, values, values.grantedStatics[index], game.LayerAbility, nil)
		values.grantedStatics[index].started = len(effects) > 0
		for i := range effects {
			applyContinuousEffect(g, permanent, values, &effects[i])
		}
	}
}

// scheduleGrantedStaticAbilityEffects adds the layer-6 effects of statics just
// granted to permanent to the global ability-layer evaluation, keeping
// remaining in timestamp order.
func scheduleGrantedStaticAbilityEffects(g *game.Game, remaining []abilityLayerEffect, permanent *game.Permanent, values *permanentEffectiveValues) []abilityLayerEffect {
	for i := range values.grantedStatics {
		granted := &values.grantedStatics[i]
		if granted.scheduled {
			continue
		}
		granted.scheduled = true
		for _, template := range grantedSelfEffectTemplates(granted.body) {
			if template.Layer != game.LayerAbility {
				continue
			}
			// One candidate generates all of the ability's layer-6 effects on
			// its permanent.
			remaining = insertAbilityLayerCandidate(remaining, abilityLayerEffect{
				landTypeEffect: landTypeEffect{effect: game.ContinuousEffect{Timestamp: granted.timestamp}},
				granted:        granted.key,
				recipientID:    permanent.ObjectID,
			})
			break
		}
		// Each layer-6 effect on other objects is its own candidate, so the
		// global evaluation orders it and detects its dependencies (CR 613.8).
		for _, template := range grantedGroupEffectTemplates(granted.body) {
			if template.Layer != game.LayerAbility {
				continue
			}
			remaining = insertAbilityLayerCandidate(remaining, abilityLayerEffect{
				landTypeEffect: landTypeEffect{effect: grantedGroupEffect(g, template, permanent, values.controller, *granted)},
				granted:        granted.key,
				holder:         permanent,
			})
		}
	}
	return remaining
}

// insertAbilityLayerCandidate inserts candidate after the remaining candidates
// that do not follow it in timestamp order.
func insertAbilityLayerCandidate(remaining []abilityLayerEffect, candidate abilityLayerEffect) []abilityLayerEffect {
	at, _ := slices.BinarySearchFunc(remaining, candidate, func(a, b abilityLayerEffect) int {
		if compareContinuousEffects(&a.effect, &b.effect) <= 0 {
			return -1
		}
		return 1
	})
	return slices.Insert(remaining, at, candidate)
}

// grantedGroupEffect returns the effect template generates on other objects
// from the static ability granted to holder. The holder is the effect's
// source, and its timestamp is the ability's CR 613.7a timestamp.
func grantedGroupEffect(g *game.Game, template game.ContinuousEffect, holder *game.Permanent, controller game.PlayerID, granted grantedStaticAbility) game.ContinuousEffect {
	effect := template
	effect.ID = 0
	effect.SourceObjectID = holder.ObjectID
	effect.SourceCardID = holder.CardInstanceID
	effect.Controller = controller
	effect.Timestamp = granted.timestamp
	rewriteChosenColorProtectionEffect(&effect, holder)
	rewriteCommanderIdentityProtectionEffect(g, &effect, controller)
	return effect
}

// grantedGroupCandidateApplies reports whether a scheduled layer-6 effect of
// a static ability granted to another permanent applies to recipient on the
// ability-layer board. The holder must still have the ability and its
// condition must hold (CR 611.3a).
func grantedGroupCandidateApplies(g *game.Game, candidate abilityLayerEffect, recipient *game.Permanent, board map[id.ID]permanentEffectiveValues) bool {
	holderValues := board[candidate.holder.ObjectID]
	index := slices.IndexFunc(holderValues.grantedStatics, func(granted grantedStaticAbility) bool {
		return granted.key == candidate.granted
	})
	if index < 0 || holderValues.grantedStatics[index].removed {
		return false
	}
	if !conditionSatisfied(g, conditionContext{
		controller: holderValues.controller, source: candidate.holder,
		characteristicsBefore: game.LayerAbility, layerValues: board,
	}, candidate.granted.body.Condition) {
		return false
	}
	values := board[recipient.ObjectID]
	return continuousEffectApplies(g, recipient, &values, &candidate.effect)
}

// appendGrantedGroupStaticEffects appends the effects in a layer after layer 6
// that static abilities granted to battlefield permanents generate on
// recipient. An effect that began in layer 6 continues to the same objects
// (CR 613.6); one that begins later needs its ability, its condition and its
// affected group at that layer.
func appendGrantedGroupStaticEffects(effects []game.ContinuousEffect, g *game.Game, recipient *game.Permanent, values *permanentEffectiveValues, layer game.ContinuousLayer, sources []staticAbilitySource) []game.ContinuousEffect {
	if !hasGrantedGroupStatic(g, sources) {
		return effects
	}
	board := abilityLossLayer(g, recipient, sources)
	for _, holder := range g.Battlefield {
		holderValues, ok := board.values[holder.ObjectID]
		if !ok {
			continue
		}
		for _, granted := range holderValues.grantedStatics {
			templates := grantedGroupEffectTemplates(granted.body)
			if !slices.ContainsFunc(templates, func(template game.ContinuousEffect) bool { return template.Layer == layer }) {
				continue
			}
			first := templates[0].Layer
			for _, template := range templates[1:] {
				first = min(first, template.Layer)
			}
			continued := first == game.LayerAbility
			if continued {
				if !board.grantedStarted[grantedHolderKey{key: granted.key, holder: holder.ObjectID}][recipient.ObjectID] {
					continue
				}
			} else if granted.removed || !conditionSatisfied(g, conditionContext{
				controller: holderValues.controller, source: holder, characteristicsBefore: layer,
			}, granted.body.Condition) {
				continue
			}
			for _, template := range templates {
				if template.Layer != layer {
					continue
				}
				effect := grantedGroupEffect(g, template, holder, holderValues.controller, granted)
				if continued || continuousEffectApplies(g, recipient, values, &effect) {
					effects = append(effects, effect)
				}
			}
		}
	}
	return effects
}

// hasGrantedGroupStatic reports whether any layer-6 effect grants a static
// ability that affects other objects. Such abilities need the global
// ability-layer evaluation.
func hasGrantedGroupStatic(g *game.Game, sources []staticAbilitySource) bool {
	fc := frameCacheFor(g)
	if fc != nil && fc.grantedGroupBuilt {
		return fc.grantedGroup
	}
	found := slices.ContainsFunc(g.ContinuousEffects, func(effect game.ContinuousEffect) bool {
		return grantsGroupStatic(&effect)
	})
	for _, source := range sources {
		if found {
			break
		}
		for i := range source.face.StaticAbilities {
			body := &source.face.StaticAbilities[i]
			if staticAbilityFunctionsFromSource(body, source) && slices.ContainsFunc(body.ContinuousEffects, func(effect game.ContinuousEffect) bool {
				return grantsGroupStatic(&effect)
			}) {
				found = true
				break
			}
		}
	}
	if fc != nil {
		fc.grantedGroupBuilt, fc.grantedGroup = true, found
	}
	return found
}

// grantsGroupStatic reports whether effect grants a static ability that
// affects other objects, directly or through an ability it grants.
func grantsGroupStatic(effect *game.ContinuousEffect) bool {
	if effect.Layer != game.LayerAbility {
		return false
	}
	for _, body := range effect.AddAbilities {
		static, ok := body.(*game.StaticAbility)
		if !ok || !bodyFunctionsOnBattlefield(static) {
			continue
		}
		if len(grantedGroupEffectTemplates(static)) > 0 {
			return true
		}
		for i := range static.ContinuousEffects {
			if grantsGroupStatic(&static.ContinuousEffects[i]) {
				return true
			}
		}
	}
	return false
}

// grantedStaticCandidateEffects returns the effects a scheduled granted
// static currently generates for its recipient on the ability-layer board,
// or false once the recipient no longer has the ability.
func grantedStaticCandidateEffects(g *game.Game, candidate abilityLayerEffect, recipient *game.Permanent, board map[id.ID]permanentEffectiveValues) ([]game.ContinuousEffect, bool) {
	if recipient.ObjectID != candidate.recipientID {
		return nil, false
	}
	values := board[recipient.ObjectID]
	for _, granted := range values.grantedStatics {
		if granted.key == candidate.granted {
			effects := appendGrantedStaticEffects(nil, g, recipient, &values, granted, game.LayerAbility, board)
			return effects, len(effects) > 0
		}
	}
	return nil, false
}

func applyGrantedStaticCandidate(g *game.Game, recipient *game.Permanent, values *permanentEffectiveValues, candidate abilityLayerEffect, board map[id.ID]permanentEffectiveValues) {
	effects, ok := grantedStaticCandidateEffects(g, candidate, recipient, board)
	if !ok {
		return
	}
	for i := range values.grantedStatics {
		if values.grantedStatics[i].key == candidate.granted {
			values.grantedStatics[i].started = true
		}
	}
	for i := range effects {
		applyContinuousEffect(g, recipient, values, &effects[i])
	}
}

// grantsBoardDependentStatic reports whether effect grants a static ability
// whose effects need the global ability-layer evaluation: one that affects
// other objects, or one whose layer-6 effects on its permanent are
// conditional or may change other effects.
func grantsBoardDependentStatic(effect *game.ContinuousEffect) bool {
	if effect.Layer != game.LayerAbility {
		return false
	}
	if grantsGroupStatic(effect) {
		return true
	}
	for _, body := range effect.AddAbilities {
		static, ok := grantedSelfStatic(body)
		if !ok {
			continue
		}
		templates := grantedSelfEffectTemplates(static)
		for i := range templates {
			granted := &templates[i]
			if granted.Layer == game.LayerAbility && (static.Condition.Exists || abilityEffectMayDepend(granted)) {
				return true
			}
		}
	}
	return false
}
