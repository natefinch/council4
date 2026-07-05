package cardgen

import (
	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/mtg/game"
)

// exileUntilMonarchKey links a permanent exiled "until an opponent becomes the
// monarch" to its source so the synthesized return trigger can put it back.
const exileUntilMonarchKey = game.LinkedKey("exile-until-opponent-monarch")

// lowerExileUntilOpponentBecomesMonarchContent lowers the monarch exile clause
// "exile <target> until an opponent becomes the monarch." (Palace Jailer) into a
// single linked Exile. The paired return — put the exiled card back when an
// opponent becomes the monarch — is synthesized at the face level
// (synthesizeExileUntilOpponentBecomesMonarchReturns). It mirrors the O-Ring
// exile-until-leaves lowering but anchors the return to a monarch change rather
// than the source leaving the battlefield. The target is a single permanent
// (MaxTargets 1), matching the one-card link key; every other exile shape leaves
// the clause unrecognized so lowering fails closed.
func lowerExileUntilOpponentBecomesMonarchContent(ctx contentCtx) (game.AbilityContent, bool) {
	if ctx.enclosingKind != compiler.AbilityTriggered ||
		ctx.optional ||
		len(ctx.content.Effects) != 1 ||
		len(ctx.content.Targets) != 1 ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Keywords) != 0 ||
		len(ctx.content.Modes) != 0 ||
		len(ctx.content.References) != 0 {
		return game.AbilityContent{}, false
	}
	effect := ctx.content.Effects[0]
	if effect.Kind != compiler.EffectExile ||
		!effect.ExileUntilOpponentBecomesMonarch ||
		!effect.Exact ||
		effect.Negated ||
		effect.Optional ||
		effect.Context != parser.EffectContextController {
		return game.AbilityContent{}, false
	}
	targetSpec, ok := permanentTargetSpecWithCardinality(ctx.content.Targets[0])
	if !ok || targetSpec.MaxTargets != 1 {
		return game.AbilityContent{}, false
	}
	return game.Mode{
		Targets: []game.TargetSpec{targetSpec},
		Sequence: []game.Instruction{{
			Primitive: game.Exile{
				Object:         game.TargetPermanentReference(0),
				ExileLinkedKey: exileUntilMonarchKey,
			},
		}},
	}.Ability(), true
}

// synthesizeExileUntilOpponentBecomesMonarchReturns adds the paired return trigger
// for a face that exiles a permanent under exileUntilMonarchKey (Palace Jailer):
// when an opponent becomes the monarch, the exiled card returns to the
// battlefield under its owner's control. It is a no-op when the face carries no
// such exile or already returns the linked card explicitly.
func synthesizeExileUntilOpponentBecomesMonarchReturns(result *loweredFaceAbilities) {
	if !faceExilesUntilOpponentBecomesMonarch(result) ||
		faceReturnsLinkedBattlefield(result, exileUntilMonarchKey) {
		return
	}
	result.TriggeredAbilities = append(result.TriggeredAbilities, exileUntilOpponentBecomesMonarchReturnAbility())
}

func faceExilesUntilOpponentBecomesMonarch(result *loweredFaceAbilities) bool {
	for abilityIndex := range result.TriggeredAbilities {
		if contentExilesUnderLinkKey(&result.TriggeredAbilities[abilityIndex].Content, exileUntilMonarchKey) {
			return true
		}
	}
	return false
}

// contentExilesUnderLinkKey reports whether content exiles a permanent linked
// under key.
func contentExilesUnderLinkKey(content *game.AbilityContent, key game.LinkedKey) bool {
	for modeIndex := range content.Modes {
		for instructionIndex := range content.Modes[modeIndex].Sequence {
			if exile, ok := content.Modes[modeIndex].Sequence[instructionIndex].Primitive.(game.Exile); ok &&
				exile.ExileLinkedKey == key {
				return true
			}
		}
	}
	return false
}

// exileUntilOpponentBecomesMonarchReturnAbility builds the synthesized trigger
// that returns the linked exiled card when an opponent becomes the monarch.
func exileUntilOpponentBecomesMonarchReturnAbility() game.TriggeredAbility {
	return game.TriggeredAbility{
		Text: "When an opponent becomes the monarch, return the exiled card to the battlefield under its owner's control.",
		Trigger: game.TriggerCondition{
			Type: game.TriggerWhen,
			Pattern: game.TriggerPattern{
				Event:  game.EventBecameMonarch,
				Player: game.TriggerPlayerOpponent,
			},
		},
		Content: game.Mode{Sequence: []game.Instruction{{
			Primitive: game.PutOnBattlefield{
				Source: game.LinkedBattlefieldSource(exileUntilMonarchKey),
			},
		}}}.Ability(),
	}
}
