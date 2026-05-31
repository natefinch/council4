package rules

import (
	"slices"
	"strings"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/opt"
)

func ptPtr(v opt.V[game.PT]) *game.PT {
	if !v.Exists {
		return nil
	}
	return &v.Val
}

func dynamicValuePtr(v opt.V[game.DynamicValue]) *game.DynamicValue {
	if !v.Exists {
		return nil
	}
	return &v.Val
}

type permanentEffectiveValues struct {
	name       string
	oracleText string
	colors     []mana.Color
	supertypes []game.Supertype
	types      []game.CardType
	subtypes   []string
	abilities  []game.AbilityDef
	controller game.PlayerID

	power            int
	powerOK          bool
	powerPT          *game.PT
	dynamicPower     *game.DynamicValue
	toughness        int
	toughnessOK      bool
	toughnessPT      *game.PT
	dynamicToughness *game.DynamicValue
	keywords         map[game.Keyword]bool
}

func effectivePower(g *game.Game, permanent *game.Permanent) int {
	values := effectivePermanentValues(g, permanent)
	if !values.powerOK {
		return 0
	}
	return max(0, values.power)
}

func effectiveToughness(g *game.Game, permanent *game.Permanent) (int, bool) {
	values := effectivePermanentValues(g, permanent)
	return values.toughness, values.toughnessOK
}

func hasKeyword(g *game.Game, permanent *game.Permanent, keyword game.Keyword) bool {
	return effectivePermanentValues(g, permanent).keywords[keyword]
}

func permanentHasType(g *game.Game, permanent *game.Permanent, cardType game.CardType) bool {
	return slices.Contains(effectivePermanentValues(g, permanent).types, cardType)
}

func permanentHasSubtype(g *game.Game, permanent *game.Permanent, subtype string) bool {
	return slices.Contains(effectivePermanentValues(g, permanent).subtypes, subtype)
}

func permanentHasSupertype(g *game.Game, permanent *game.Permanent, supertype game.Supertype) bool {
	return slices.Contains(effectivePermanentValues(g, permanent).supertypes, supertype)
}

func permanentEffectiveColors(g *game.Game, permanent *game.Permanent) []mana.Color {
	values := effectivePermanentValues(g, permanent)
	return append([]mana.Color(nil), values.colors...)
}

func permanentEffectiveAbilities(g *game.Game, permanent *game.Permanent) []game.AbilityDef {
	values := effectivePermanentValues(g, permanent)
	return append([]game.AbilityDef(nil), values.abilities...)
}

func permanentEffectiveName(g *game.Game, permanent *game.Permanent) string {
	return effectivePermanentValues(g, permanent).name
}

func effectiveController(g *game.Game, permanent *game.Permanent) game.PlayerID {
	values := basePermanentValues(g, permanent)
	for _, effect := range orderContinuousEffects(continuousEffectsForLayer(g, permanent, &values, game.LayerControl)) {
		if effect.NewController.Exists {
			values.controller = effect.NewController.Val
		}
	}
	return values.controller
}

func effectivePermanentValues(g *game.Game, permanent *game.Permanent) permanentEffectiveValues {
	values := basePermanentValues(g, permanent)
	applyContinuousLayers(g, permanent, &values)
	applyCounterAndTemporaryValues(permanent, &values)
	for _, keyword := range keywordCounters(permanent) {
		values.keywords[keyword] = true
	}
	return values
}

func basePermanentValues(g *game.Game, permanent *game.Permanent) permanentEffectiveValues {
	values := permanentEffectiveValues{keywords: make(map[game.Keyword]bool)}
	values.controller = permanent.Controller
	if permanent.FaceDown {
		values.types = []game.CardType{game.TypeCreature}
		if permanent.FaceDownKind == game.FaceDownDisguise {
			values.abilities = []game.AbilityDef{faceDownDisguiseWardAbility()}
			rebuildKeywords(&values)
		}
		values.power, values.powerOK = 2, true
		values.toughness, values.toughnessOK = 2, true
		return values
	}
	card, ok := permanentCardDef(g, permanent)
	if !ok {
		return values
	}
	values.name = card.Name
	values.oracleText = card.OracleText
	values.colors = append([]mana.Color(nil), card.Colors...)
	values.supertypes = append([]game.Supertype(nil), card.Supertypes...)
	values.types = append([]game.CardType(nil), card.Types...)
	values.subtypes = append([]string(nil), card.Subtypes...)
	values.abilities = append([]game.AbilityDef(nil), card.Abilities...)
	if card.Power.Exists {
		values.powerPT = ptPtr(card.Power)
		values.dynamicPower = dynamicValuePtr(card.DynamicPower)
		values.power, values.powerOK = ptValue(g, values.controller, values.powerPT, values.dynamicPower)
	}
	if card.Toughness.Exists {
		values.toughnessPT = ptPtr(card.Toughness)
		values.dynamicToughness = dynamicValuePtr(card.DynamicToughness)
		values.toughness, values.toughnessOK = ptValue(g, values.controller, values.toughnessPT, values.dynamicToughness)
	}
	rebuildKeywords(&values)
	return values
}

func ptValue(g *game.Game, controller game.PlayerID, pt *game.PT, dynamic *game.DynamicValue) (int, bool) {
	if pt == nil {
		return 0, false
	}
	if !pt.IsStar {
		return pt.Value, true
	}
	if dynamic == nil {
		return 0, false
	}
	return dynamicValue(g, controller, dynamic), true
}

func dynamicValue(g *game.Game, controller game.PlayerID, dynamic *game.DynamicValue) int {
	if dynamic == nil {
		return 0
	}
	switch dynamic.Kind {
	case game.DynamicValueConstant:
		return dynamic.Value
	case game.DynamicValueControllerHandSize:
		if player, ok := playerByID(g, controller); ok {
			return player.Hand.Size()
		}
	case game.DynamicValueControllerGraveyardSize:
		if player, ok := playerByID(g, controller); ok {
			return player.Graveyard.Size()
		}
	case game.DynamicValueControllerCreatureCount:
		return countControlledPermanentsWithType(g, controller, game.TypeCreature)
	case game.DynamicValueControllerLandCount:
		return countControlledPermanentsWithType(g, controller, game.TypeLand)
	case game.DynamicValueControllerArtifactCount:
		return countControlledPermanentsWithType(g, controller, game.TypeArtifact)
	case game.DynamicValueAllBattlefieldCreatureCount:
		count := 0
		for _, permanent := range g.Battlefield {
			if basePermanentHasType(g, permanent, game.TypeCreature) {
				count++
			}
		}
		return count
	}
	return 0
}

func countControlledPermanentsWithType(g *game.Game, controller game.PlayerID, cardType game.CardType) int {
	count := 0
	for _, permanent := range g.Battlefield {
		if permanent.Controller == controller && basePermanentHasType(g, permanent, cardType) {
			count++
		}
	}
	return count
}

func basePermanentHasType(g *game.Game, permanent *game.Permanent, cardType game.CardType) bool {
	if permanent.FaceDown {
		return cardType == game.TypeCreature
	}
	card, ok := permanentCardDef(g, permanent)
	return ok && card.HasType(cardType)
}

func applyContinuousLayers(g *game.Game, permanent *game.Permanent, values *permanentEffectiveValues) {
	for _, layer := range []game.ContinuousLayer{
		game.LayerCopy,
		game.LayerControl,
		game.LayerText,
		game.LayerType,
		game.LayerColor,
		game.LayerAbility,
		game.LayerPowerToughnessSet,
		game.LayerPowerToughnessModify,
		game.LayerPowerToughnessSwitch,
	} {
		effects := continuousEffectsForLayer(g, permanent, values, layer)
		for _, effect := range orderContinuousEffects(effects) {
			applyContinuousEffect(g, permanent, values, effect)
		}
	}
}

func continuousEffectsForLayer(g *game.Game, permanent *game.Permanent, values *permanentEffectiveValues, layer game.ContinuousLayer) []game.ContinuousEffect {
	var effects []game.ContinuousEffect
	for _, effect := range g.ContinuousEffects {
		if effect.Layer == layer && continuousEffectApplies(g, permanent, values, effect) {
			effects = append(effects, effect)
		}
	}
	if layer == game.LayerPowerToughnessModify {
		effects = append(effects, staticPTContinuousEffects(g, permanent, values)...)
	}
	return effects
}

func staticPTContinuousEffects(g *game.Game, permanent *game.Permanent, values *permanentEffectiveValues) []game.ContinuousEffect {
	var effects []game.ContinuousEffect
	for _, source := range g.Battlefield {
		sourceDef, ok := permanentCardDef(g, source)
		if !ok {
			continue
		}
		for i := range sourceDef.Abilities {
			ability := &sourceDef.Abilities[i]
			if ability.Kind != game.StaticAbility || !abilityFunctionsOnBattlefield(ability) {
				continue
			}
			for _, effect := range ability.Effects {
				sourceController := effectiveController(g, source)
				if effect.Type != game.EffectModifyPT || !permanentValuesMatchSelectorForSource(source, sourceController, permanent, values, effect.Selector) {
					continue
				}
				effects = append(effects, game.ContinuousEffect{
					ID:               0,
					SourceObjectID:   source.ObjectID,
					SourceCardID:     source.CardInstanceID,
					Controller:       sourceController,
					Timestamp:        source.Timestamp,
					AffectedObjectID: permanent.ObjectID,
					Layer:            game.LayerPowerToughnessModify,
					PowerDelta:       effect.PowerDelta,
					ToughnessDelta:   effect.ToughnessDelta,
				})
			}
		}
	}
	return effects
}

func continuousEffectApplies(g *game.Game, permanent *game.Permanent, values *permanentEffectiveValues, effect game.ContinuousEffect) bool {
	if effect.AffectedObjectID != 0 {
		return effect.AffectedObjectID == permanent.ObjectID
	}
	if effect.Selector == game.EffectSelectorNone {
		return false
	}
	source, _ := permanentByObjectID(g, effect.SourceObjectID)
	return permanentValuesMatchSelectorForSource(source, effect.Controller, permanent, values, effect.Selector)
}

func permanentValuesMatchSelectorForSource(source *game.Permanent, controller game.PlayerID, permanent *game.Permanent, values *permanentEffectiveValues, selector game.EffectSelector) bool {
	switch selector {
	case game.EffectSelectorAllCreatures:
		return valuesHasType(values, game.TypeCreature)
	case game.EffectSelectorAllArtifacts:
		return valuesHasType(values, game.TypeArtifact)
	case game.EffectSelectorAllEnchantments:
		return valuesHasType(values, game.TypeEnchantment)
	case game.EffectSelectorAllNonlandPermanents:
		return !valuesHasType(values, game.TypeLand)
	case game.EffectSelectorAllPermanents:
		return true
	case game.EffectSelectorCreaturesYouControl:
		return values.controller == controller && valuesHasType(values, game.TypeCreature)
	case game.EffectSelectorOtherCreaturesYouControl:
		return source != nil && permanent.ObjectID != source.ObjectID && values.controller == controller && valuesHasType(values, game.TypeCreature)
	default:
		return false
	}
}

func valuesHasType(values *permanentEffectiveValues, cardType game.CardType) bool {
	return slices.Contains(values.types, cardType)
}

func orderContinuousEffects(effects []game.ContinuousEffect) []game.ContinuousEffect {
	if len(effects) <= 1 {
		return effects
	}
	ordered := append([]game.ContinuousEffect(nil), effects...)
	slices.SortStableFunc(ordered, compareContinuousEffects)
	remaining := append([]game.ContinuousEffect(nil), ordered...)
	result := make([]game.ContinuousEffect, 0, len(ordered))
	applied := make(map[id.ID]bool, len(ordered))
	for len(remaining) > 0 {
		progress := false
		for i := 0; i < len(remaining); {
			if dependenciesSatisfied(remaining[i], applied, remaining) {
				result = append(result, remaining[i])
				if remaining[i].ID != 0 {
					applied[remaining[i].ID] = true
				}
				remaining = append(remaining[:i], remaining[i+1:]...)
				progress = true
				continue
			}
			i++
		}
		if !progress {
			return append(result, remaining...)
		}
	}
	return result
}

func compareContinuousEffects(left, right game.ContinuousEffect) int {
	if left.Timestamp < right.Timestamp {
		return -1
	}
	if left.Timestamp > right.Timestamp {
		return 1
	}
	if left.ID < right.ID {
		return -1
	}
	if left.ID > right.ID {
		return 1
	}
	return 0
}

func dependenciesSatisfied(effect game.ContinuousEffect, applied map[id.ID]bool, remaining []game.ContinuousEffect) bool {
	for _, dependency := range effect.DependsOn {
		if dependency == 0 || applied[dependency] {
			continue
		}
		for _, other := range remaining {
			if other.ID == dependency {
				return false
			}
		}
	}
	return true
}

func applyContinuousEffect(g *game.Game, permanent *game.Permanent, values *permanentEffectiveValues, effect game.ContinuousEffect) {
	switch effect.Layer {
	case game.LayerCopy:
		if effect.CopyValues.Exists {
			applyCopyValues(g, permanent, values, &effect.CopyValues.Val)
		}
	case game.LayerControl:
		if effect.NewController.Exists {
			values.controller = effect.NewController.Val
			recalculateDynamicPT(g, values)
		}
	case game.LayerText:
		if effect.TextFrom != "" {
			values.oracleText = strings.ReplaceAll(values.oracleText, effect.TextFrom, effect.TextTo)
		}
	case game.LayerType:
		applyTypeLayer(values, effect)
	case game.LayerColor:
		if effect.SetColors != nil {
			values.colors = append([]mana.Color(nil), effect.SetColors...)
		}
		values.colors = removeColors(values.colors, effect.RemoveColors)
		values.colors = appendUniqueColors(values.colors, effect.AddColors...)
	case game.LayerAbility:
		values.abilities = append(values.abilities, effect.AddAbilities...)
		for _, ability := range effect.AddAbilities {
			for _, keyword := range ability.Keywords {
				values.keywords[keyword] = true
			}
		}
		for _, keyword := range effect.RemoveKeywords {
			values.keywords[keyword] = false
		}
		for _, keyword := range effect.AddKeywords {
			values.keywords[keyword] = true
		}
	case game.LayerPowerToughnessSet:
		if effect.SetPower.Exists {
			values.powerPT = ptPtr(effect.SetPower)
			values.dynamicPower = nil
			values.power, values.powerOK = ptValue(g, values.controller, values.powerPT, nil)
		}
		if effect.SetToughness.Exists {
			values.toughnessPT = ptPtr(effect.SetToughness)
			values.dynamicToughness = nil
			values.toughness, values.toughnessOK = ptValue(g, values.controller, values.toughnessPT, nil)
		}
	case game.LayerPowerToughnessModify:
		if values.powerOK {
			values.power += effect.PowerDelta
		}
		if values.toughnessOK {
			values.toughness += effect.ToughnessDelta
		}
	case game.LayerPowerToughnessSwitch:
		values.power, values.toughness = values.toughness, values.power
		values.powerOK, values.toughnessOK = values.toughnessOK, values.powerOK
	}
}

func applyCopyValues(g *game.Game, permanent *game.Permanent, values *permanentEffectiveValues, copyValues *game.CopyableValues) {
	if copyValues == nil {
		return
	}
	values.name = copyValues.Name
	values.oracleText = copyValues.OracleText
	values.colors = append([]mana.Color(nil), copyValues.Colors...)
	values.supertypes = append([]game.Supertype(nil), copyValues.Supertypes...)
	values.types = append([]game.CardType(nil), copyValues.Types...)
	values.subtypes = append([]string(nil), copyValues.Subtypes...)
	values.abilities = append([]game.AbilityDef(nil), copyValues.Abilities...)
	values.powerPT = ptPtr(copyValues.Power)
	values.dynamicPower = dynamicValuePtr(copyValues.DynamicPower)
	values.toughnessPT = ptPtr(copyValues.Toughness)
	values.dynamicToughness = dynamicValuePtr(copyValues.DynamicToughness)
	values.power, values.powerOK = ptValue(g, values.controller, values.powerPT, values.dynamicPower)
	values.toughness, values.toughnessOK = ptValue(g, values.controller, values.toughnessPT, values.dynamicToughness)
	rebuildKeywords(values)
}

func recalculateDynamicPT(g *game.Game, values *permanentEffectiveValues) {
	if values == nil {
		return
	}
	// LayerControl runs before P/T layers, so this cannot erase later P/T effects.
	if values.powerPT != nil && values.powerPT.IsStar {
		values.power, values.powerOK = ptValue(g, values.controller, values.powerPT, values.dynamicPower)
	}
	if values.toughnessPT != nil && values.toughnessPT.IsStar {
		values.toughness, values.toughnessOK = ptValue(g, values.controller, values.toughnessPT, values.dynamicToughness)
	}
}

func applyTypeLayer(values *permanentEffectiveValues, effect game.ContinuousEffect) {
	if effect.SetSupertypes != nil {
		values.supertypes = append([]game.Supertype(nil), effect.SetSupertypes...)
	}
	values.supertypes = removeSupertypes(values.supertypes, effect.RemoveSupertypes)
	values.supertypes = appendUniqueSupertypes(values.supertypes, effect.AddSupertypes...)

	if effect.SetTypes != nil {
		values.types = append([]game.CardType(nil), effect.SetTypes...)
	}
	values.types = removeTypes(values.types, effect.RemoveTypes)
	values.types = appendUniqueTypes(values.types, effect.AddTypes...)

	if effect.SetSubtypes != nil {
		values.subtypes = append([]string(nil), effect.SetSubtypes...)
	}
	values.subtypes = removeStrings(values.subtypes, effect.RemoveSubtypes)
	values.subtypes = appendUniqueStrings(values.subtypes, effect.AddSubtypes...)
}

func applyCounterAndTemporaryValues(permanent *game.Permanent, values *permanentEffectiveValues) {
	counterDelta := powerToughnessCounterDelta(permanent)
	if values.powerOK {
		values.power += counterDelta + permanent.TemporaryPowerModifier
	}
	if values.toughnessOK {
		values.toughness += counterDelta + permanent.TemporaryToughnessModifier
	}
}

func rebuildKeywords(values *permanentEffectiveValues) {
	values.keywords = make(map[game.Keyword]bool)
	for i := range values.abilities {
		for _, keyword := range values.abilities[i].Keywords {
			values.keywords[keyword] = true
		}
	}
}

func abilityFunctionsOnBattlefield(ability *game.AbilityDef) bool {
	return ability != nil && (ability.ZoneOfFunction == game.ZoneNone || ability.ZoneOfFunction == game.ZoneBattlefield)
}

func powerToughnessCounterDelta(permanent *game.Permanent) int {
	return permanent.Counters.Get(counter.PlusOnePlusOne) - permanent.Counters.Get(counter.MinusOneMinusOne)
}

func keywordCounters(permanent *game.Permanent) []game.Keyword {
	var keywords []game.Keyword
	for _, keyword := range []game.Keyword{
		game.Deathtouch,
		game.FirstStrike,
		game.Flying,
		game.Hexproof,
		game.Indestructible,
		game.Lifelink,
		game.Menace,
		game.Reach,
		game.Trample,
		game.Vigilance,
	} {
		counterKind, ok := keywordCounterKind(keyword)
		if ok && permanent.Counters.Get(counterKind) > 0 {
			keywords = append(keywords, keyword)
		}
	}
	return keywords
}

func removeColors(colors []mana.Color, remove []mana.Color) []mana.Color {
	return slices.DeleteFunc(colors, func(color mana.Color) bool {
		return slices.Contains(remove, color)
	})
}

func appendUniqueColors(colors []mana.Color, add ...mana.Color) []mana.Color {
	for _, color := range add {
		if !slices.Contains(colors, color) {
			colors = append(colors, color)
		}
	}
	return colors
}

func removeSupertypes(supertypes []game.Supertype, remove []game.Supertype) []game.Supertype {
	return slices.DeleteFunc(supertypes, func(supertype game.Supertype) bool {
		return slices.Contains(remove, supertype)
	})
}

func appendUniqueSupertypes(supertypes []game.Supertype, add ...game.Supertype) []game.Supertype {
	for _, supertype := range add {
		if !slices.Contains(supertypes, supertype) {
			supertypes = append(supertypes, supertype)
		}
	}
	return supertypes
}

func removeTypes(types []game.CardType, remove []game.CardType) []game.CardType {
	return slices.DeleteFunc(types, func(cardType game.CardType) bool {
		return slices.Contains(remove, cardType)
	})
}

func appendUniqueTypes(types []game.CardType, add ...game.CardType) []game.CardType {
	for _, cardType := range add {
		if !slices.Contains(types, cardType) {
			types = append(types, cardType)
		}
	}
	return types
}

func removeStrings(values []string, remove []string) []string {
	return slices.DeleteFunc(values, func(value string) bool {
		return slices.Contains(remove, value)
	})
}

func appendUniqueStrings(values []string, add ...string) []string {
	for _, value := range add {
		if !slices.Contains(values, value) {
			values = append(values, value)
		}
	}
	return values
}
