package cardgen

import (
	"reflect"

	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

const (
	imprintCreatedTokenLink = game.LinkedKey("imprint-created-token")
	imprintTokenResultKey   = game.ResultKey("imprint-token-created")
)

// lowerReplaceLinkedExiledCardTrigger lowers the typed single-current-imprint
// sequence. It reads only the semantic exact-sequence kind and trigger pattern;
// the parser owns every Oracle word.
func lowerReplaceLinkedExiledCardTrigger(
	ability compiler.CompiledAbility,
) (game.TriggeredAbility, *shared.Diagnostic) {
	const summary = "unsupported linked-exile replacement trigger"
	if ability.Trigger == nil ||
		ability.Trigger.Pattern.Kind != compiler.TriggerWhenever ||
		!ability.Optional ||
		ability.Trigger.Condition != nil ||
		ability.Trigger.MaxTriggersPerTurn != 0 ||
		ability.Content.Unconsumed() {
		return game.TriggeredAbility{}, executableDiagnostic(
			ability,
			summary,
			"the linked-exile replacement requires an optional bare whenever trigger with no other content",
		)
	}
	pattern, ok := lowerTriggerPattern(&ability.Trigger.Pattern)
	wantSelection := game.Selection{
		RequiredTypes: []types.Card{types.Creature},
		NonToken:      true,
	}
	if !ok ||
		pattern.Event != game.EventPermanentDied ||
		pattern.Controller != game.TriggerControllerAny ||
		pattern.Source != game.TriggerSourceAny ||
		!reflect.DeepEqual(pattern.SubjectSelection, wantSelection) {
		return game.TriggeredAbility{}, executableDiagnostic(
			ability,
			summary,
			"the linked-exile replacement requires an unrestricted nontoken-creature dies trigger",
		)
	}
	return game.TriggeredAbility{
		Text: ability.Text,
		Trigger: game.TriggerCondition{
			Type:    game.TriggerWhenever,
			Pattern: pattern,
		},
		Content: game.Mode{Sequence: []game.Instruction{{
			Optional: true,
			Primitive: game.ReplaceLinkedExiledCard{
				Card:     game.CardReference{Kind: game.CardReferenceEvent},
				FromZone: zone.Graveyard,
				LinkID:   game.LinkedKey(imprintLinkKey),
			},
		}}}.Ability(),
	}, nil
}

// lowerLinkedExiledCopyTokenActivation lowers the typed imprint payoff into a
// linked-card token copy and a delayed group exile. Capturing the whole created
// group preserves token-doubling replacements and same-turn repeated activations.
func lowerLinkedExiledCopyTokenActivation(
	ability compiler.CompiledAbility,
	bodyText string,
) (game.AbilityContent, *shared.Diagnostic) {
	if ability.Optional ||
		ability.Content.Unconsumed() {
		return game.AbilityContent{}, executableDiagnostic(
			ability,
			"unsupported linked-exile token-copy activation",
			"the linked-exile token copy requires an unconditional nonoptional activation body",
		)
	}
	create := game.Instruction{
		Primitive: game.CreateToken{
			Amount: game.Fixed(1),
			Source: game.TokenCopyOf(game.TokenCopySpec{
				Source:      game.TokenCopySourceLinkedExiledCard,
				LinkID:      game.LinkedKey(imprintLinkKey),
				AddKeywords: []game.Keyword{game.Haste},
			}),
			PublishLinked: imprintCreatedTokenLink,
		},
		PublishResult: imprintTokenResultKey,
	}
	cleanup := game.Instruction{
		ResultGate: opt.Val(game.InstructionResultGate{
			Key:       imprintTokenResultKey,
			Succeeded: game.TriTrue,
		}),
		Primitive: game.CreateDelayedTrigger{Trigger: game.DelayedTriggerDef{
			Timing:              game.DelayedAtBeginningOfNextEndStep,
			CapturedObjectGroup: opt.Val(game.LinkedObjectReference(string(imprintCreatedTokenLink))),
			Content: game.Mode{Sequence: []game.Instruction{{
				Primitive: game.Exile{Group: game.CapturedObjectsGroup()},
			}}}.Ability(),
		}},
	}
	return game.Mode{
		Text:     bodyText,
		Sequence: []game.Instruction{create, cleanup},
	}.Ability(), nil
}
