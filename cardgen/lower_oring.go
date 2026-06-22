package cardgen

import (
	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/zone"
)

// exileUntilLeavesKey is the constant linked key binding an O-Ring exile to the
// source permanent that exiled it. The runtime keys linked objects by source
// card-instance id plus this string, so a fixed key still keeps each prison
// permanent's exiled card distinct. The enters-the-battlefield exile publishes
// it and the synthesized leaves-the-battlefield trigger consumes it to return
// the card.
const exileUntilLeavesKey = game.LinkedKey("exile-until-leaves")

// lowerExileUntilLeavesContent lowers the O-Ring enters-the-battlefield clause
// "exile target <permanent> until <this permanent> leaves the battlefield."
// (Banisher Priest, Banishing Light, Fairgrounds Warden) into a linked exile of
// the single target. The trailing self-reference is the duration anchor naming
// the source permanent, not a second object, so it is consumed rather than
// bound to a target. A paired leaves-the-battlefield return trigger is
// synthesized at the face level (synthesizeExileUntilLeavesReturns) using the
// same linked key.
//
// It returns ok=false for any shape it does not fully consume: a missing or
// non-single target, a non-exile or optional effect, a condition, mode, or
// keyword rider, or a reference that is not the source duration anchor.
func lowerExileUntilLeavesContent(ctx contentCtx) (game.AbilityContent, bool) {
	if ctx.enclosingKind != compiler.AbilityTriggered ||
		ctx.optional ||
		len(ctx.content.Effects) != 1 ||
		len(ctx.content.Targets) != 1 ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Keywords) != 0 ||
		len(ctx.content.Modes) != 0 {
		return game.AbilityContent{}, false
	}
	effect := ctx.content.Effects[0]
	if effect.Kind != compiler.EffectExile ||
		!effect.ExileUntilSourceLeaves ||
		!effect.Exact ||
		effect.Negated ||
		effect.Optional ||
		effect.Context != parser.EffectContextController {
		return game.AbilityContent{}, false
	}
	if !referencesAreOnlySourceAnchors(ctx.content.References) {
		return game.AbilityContent{}, false
	}
	targetSpec, ok := permanentTargetSpec(ctx.content.Targets[0])
	if !ok {
		return game.AbilityContent{}, false
	}
	return game.Mode{
		Targets: []game.TargetSpec{targetSpec},
		Sequence: []game.Instruction{{
			Primitive: game.Exile{
				Object:         game.TargetPermanentReference(0),
				ExileLinkedKey: exileUntilLeavesKey,
			},
		}},
	}.Ability(), true
}

// referencesAreOnlySourceAnchors reports whether every reference (if any) is the
// "this permanent" / source-name duration anchor bound to the source permanent,
// as in "... until this creature leaves the battlefield." Such references name
// the source as the exile's lifetime anchor and carry no resolving object, so
// the O-Ring lowering consumes them in place of a target binding.
func referencesAreOnlySourceAnchors(references []compiler.CompiledReference) bool {
	for _, reference := range references {
		if reference.Binding != compiler.ReferenceBindingSource {
			return false
		}
		if reference.Kind != compiler.ReferenceThisObject &&
			reference.Kind != compiler.ReferenceSelfName {
			return false
		}
	}
	return true
}

// synthesizeExileUntilLeavesReturns appends the paired return-on-leave trigger
// for an O-Ring face. When a lowered trigger exiles a permanent under the
// exile-until-leaves linked key and the face declares no return for that key,
// it adds "When this permanent leaves the battlefield, return the exiled card to
// the battlefield under its owner's control." so the prison releases its captive
// when it leaves play.
func synthesizeExileUntilLeavesReturns(result *loweredFaceAbilities) {
	if !faceExilesUntilLeaves(result) || faceReturnsLinkedBattlefield(result, exileUntilLeavesKey) {
		return
	}
	result.TriggeredAbilities = append(result.TriggeredAbilities, exileUntilLeavesReturnAbility())
}

func faceExilesUntilLeaves(result *loweredFaceAbilities) bool {
	for abilityIndex := range result.TriggeredAbilities {
		content := &result.TriggeredAbilities[abilityIndex].Content
		for modeIndex := range content.Modes {
			for instructionIndex := range content.Modes[modeIndex].Sequence {
				exile, ok := content.Modes[modeIndex].Sequence[instructionIndex].Primitive.(game.Exile)
				if ok && exile.ExileLinkedKey == exileUntilLeavesKey {
					return true
				}
			}
		}
	}
	return false
}

func faceReturnsLinkedBattlefield(result *loweredFaceAbilities, key game.LinkedKey) bool {
	for abilityIndex := range result.TriggeredAbilities {
		content := &result.TriggeredAbilities[abilityIndex].Content
		for modeIndex := range content.Modes {
			for instructionIndex := range content.Modes[modeIndex].Sequence {
				put, ok := content.Modes[modeIndex].Sequence[instructionIndex].Primitive.(game.PutOnBattlefield)
				if !ok {
					continue
				}
				if linked, ok := put.Source.LinkedKey(); ok && linked == key {
					return true
				}
			}
		}
	}
	return false
}

func exileUntilLeavesReturnAbility() game.TriggeredAbility {
	return game.TriggeredAbility{
		Text: "When this permanent leaves the battlefield, return the exiled card to the battlefield under its owner's control.",
		Trigger: game.TriggerCondition{
			Type: game.TriggerWhen,
			Pattern: game.TriggerPattern{
				Event:         game.EventZoneChanged,
				Source:        game.TriggerSourceSelf,
				MatchFromZone: true,
				FromZone:      zone.Battlefield,
			},
		},
		Content: game.Mode{Sequence: []game.Instruction{{
			Primitive: game.PutOnBattlefield{
				Source: game.LinkedBattlefieldSource(exileUntilLeavesKey),
			},
		}}}.Ability(),
	}
}
