package cardgen

import (
	"slices"

	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/zone"
)

// isReanimationAuraFace reports whether a face is a graveyard-reanimation Aura
// of the Animate Dead / Dance of the Dead family: an Aura whose Enchant
// restriction is a creature card in a graveyard and which carries the "when
// this Aura enters ... return ... and attach it ... when this Aura leaves ...
// sacrifice" reanimation trigger. Both structural signals must be present so
// unrelated Auras and bare graveyard-card enchants fail closed. The recognition
// is text-blind: it keys off the compiled enchant restriction and trigger
// shape, never a card name.
func isReanimationAuraFace(parsedType ParsedTypeLine, compilation compiler.Compilation) bool {
	if !slices.Contains(parsedType.Subtypes, "Aura") {
		return false
	}
	hasGraveyardEnchant := false
	hasEnterTrigger := false
	for _, ability := range compilation.Abilities {
		if reanimationEnchantTarget(ability) {
			hasGraveyardEnchant = true
		}
		if isReanimationEnterTrigger(ability) {
			hasEnterTrigger = true
		}
	}
	return hasGraveyardEnchant && hasEnterTrigger
}

// reanimationEnchantTarget reports whether an ability is the graveyard-card
// Enchant static ("Enchant creature card in a graveyard") of a reanimation Aura.
func reanimationEnchantTarget(ability compiler.CompiledAbility) bool {
	if len(ability.Content.Keywords) != 1 ||
		ability.Content.Keywords[0].Kind != parser.KeywordEnchant {
		return false
	}
	return ability.Content.Keywords[0].EnchantTarget.InGraveyard
}

// isReanimationEnterTrigger reports whether an ability is the "when this Aura
// enters ..." reanimation trigger. Its whole body (lose/gain enchant, return
// the enchanted card under your control, attach this Aura, and the delayed
// leaves-battlefield sacrifice) is modeled by the runtime reanimation
// resolution handler plus the emitted leaves-battlefield sacrifice trigger, so
// the recognizer consumes the ability and re-emits only the leaves trigger.
func isReanimationEnterTrigger(ability compiler.CompiledAbility) bool {
	if ability.Kind != compiler.AbilityTriggered || ability.Trigger == nil {
		return false
	}
	pattern, ok := lowerTriggerPattern(&ability.Trigger.Pattern)
	if !ok {
		return false
	}
	return pattern.Source == game.TriggerSourceSelf &&
		pattern.Event == game.EventPermanentEnteredBattlefield
}

// reanimationLeavesSacrificeTrigger builds the leaves-battlefield trigger of a
// reanimation Aura ("When this Aura leaves the battlefield, that creature's
// controller sacrifices it"). It sacrifices the reanimated creature recorded as
// the Aura's linked object at resolution; if the creature already left, the
// linked reference resolves to nothing and the sacrifice is a no-op.
func reanimationLeavesSacrificeTrigger() game.TriggeredAbility {
	return game.TriggeredAbility{
		Trigger: game.TriggerCondition{
			Type: game.TriggerWhen,
			Pattern: game.TriggerPattern{
				Event:         game.EventZoneChanged,
				Source:        game.TriggerSourceSelf,
				MatchFromZone: true,
				FromZone:      zone.Battlefield,
			},
		},
		Content: game.Mode{
			Sequence: []game.Instruction{
				{
					Primitive: game.Sacrifice{
						Object:          game.LinkedObjectReference(game.ReanimationLinkID),
						ByItsController: true,
					},
				},
			},
		}.Ability(),
	}
}
