package cardgen

import (
	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

func lowerLibraryCardClause(ctx contentCtx) (game.AbilityContent, *shared.Diagnostic, bool) {
	effect := ctx.content.Effects[0]
	if effect.Kind != compiler.EffectLookAtLibraryTop && effect.Kind != compiler.EffectReveal &&
		effect.Kind != compiler.EffectPut ||
		effect.CardSource == parser.EffectCardSourceNone ||
		effect.Player != parser.EffectPlayerNone {
		return game.AbilityContent{}, nil, false
	}
	unsupported := func(detail string) (game.AbilityContent, *shared.Diagnostic, bool) {
		return game.AbilityContent{}, contentDiagnostic(ctx, "unsupported observed library card", detail), true
	}
	if effect.CardSource == parser.EffectCardSourcePriorInstructionResult && effect.FromZone == zone.Battlefield {
		if content, ok := lowerReferencedCardMove(ctx); ok {
			return content, nil, true
		}
		return unsupported("the reached card has no exact entered-object movement adapter")
	}
	if !effect.Exact || effect.Negated || effect.DelayedTiming != 0 ||
		effect.Optional || ctx.optional || len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Keywords) != 0 || len(ctx.content.Modes) != 0 {
		return unsupported("the library-card clause has an unmodeled condition, scope, or parameter")
	}
	if effect.CardSource == parser.EffectCardSourceTopOfPlayerLibrary {
		if effect.Kind != compiler.EffectLookAtLibraryTop && effect.Kind != compiler.EffectReveal ||
			!effect.Amount.Known || effect.Amount.Value != 1 || len(ctx.content.References) != 0 {
			return unsupported("observation requires exactly one top card and a modeled library owner")
		}
		player, targets, ok := libraryCardPlayer(ctx)
		if !ok || effect.ClauseID <= 0 {
			return unsupported("observation has no exact library owner or producer clause identity")
		}
		var key game.LinkedKey
		if effect.Kind == compiler.EffectLookAtLibraryTop {
			key = sequenceProductKey(ctx.sequenceEffectIndex)
		}
		if !ctx.sequenceClause && effect.Kind == compiler.EffectLookAtLibraryTop &&
			effect.Context == parser.EffectContextTarget {
			key = lookAtTargetLibraryTopKey
		}
		var primitive game.Primitive = game.LookAtLibraryTop{Player: player, PublishLinked: key}
		if effect.Kind == compiler.EffectReveal {
			primitive = game.Reveal{Amount: game.Fixed(1), Player: player, PublishLinked: key}
		}
		instruction := game.Instruction{Primitive: primitive}
		if key != "" && sequenceLocalProductKey(key) {
			instruction.LocalProducts.Links = []game.LinkedKey{key}
		}
		return game.Mode{Targets: targets, Sequence: []game.Instruction{instruction}}.Ability(), nil, true
	}
	if effect.CardSource != parser.EffectCardSourcePriorInstructionResult ||
		effect.FromZone != zone.Library || len(ctx.content.Targets) != 0 {
		return unsupported("the singular card subject has no parser-owned library observation")
	}
	subject, ok := libraryCardActionSubject(effect, ctx.content.References)
	if !ok {
		return unsupported("the card or destination owner has no exact matching observation provenance")
	}
	card, ok := lowerCardReference(subject, referenceLoweringContext{
		PriorInstruction: ctx.priorInstruction, PriorLinkedKey: ctx.priorLinkedKey,
	})
	if !ok {
		return unsupported("the observed card has no available typed product binding")
	}

	var primitive game.Primitive
	switch effect.Kind {
	case compiler.EffectReveal:
		primitive = game.Reveal{Card: card}
	case compiler.EffectPut:
		if effect.ToZone == zone.Battlefield {
			primitive = game.PutOnBattlefield{
				Source:    game.LinkedBattlefieldSource(ctx.priorLinkedKey),
				Recipient: opt.Val(game.ControllerReference()), EntryTapped: effect.EntersTapped,
			}
		} else {
			switch effect.ToZone {
			case zone.Hand, zone.Graveyard:
				if effect.Destination != parser.EffectDestinationUnspecified {
					return unsupported("a non-library destination cannot carry library placement")
				}
			case zone.Library:
				if effect.Destination != parser.EffectDestinationTop && effect.Destination != parser.EffectDestinationBottom {
					return unsupported("the library destination requires exact top or bottom placement")
				}
			default:
				return unsupported("the observed-card destination is not modeled")
			}
			primitive = game.MoveCard{
				Card: card, FromZone: zone.Library, Destination: effect.ToZone,
				DestinationBottom: effect.Destination == parser.EffectDestinationBottom,
			}
		}
	default:
		return unsupported("the observed-card action is not modeled")
	}
	return game.Mode{Sequence: []game.Instruction{{Primitive: primitive}}}.Ability(), nil, true
}

func libraryCardPlayer(ctx contentCtx) (game.PlayerReference, []game.TargetSpec, bool) {
	switch ctx.content.Effects[0].Context {
	case parser.EffectContextController:
		return game.ControllerReference(), nil, len(ctx.content.Targets) == 0
	case parser.EffectContextTarget:
		if len(ctx.content.Targets) != 1 {
			return game.PlayerReference{}, nil, false
		}
		target, ok := playerTargetSpec(ctx.content.Targets[0])
		if !ok {
			return game.PlayerReference{}, nil, false
		}
		return game.TargetPlayerReference(0), []game.TargetSpec{target}, true
	default:
		return game.PlayerReference{}, nil, false
	}
}
