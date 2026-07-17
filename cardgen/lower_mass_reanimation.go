package cardgen

import (
	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/zone"
)

// massReanimationChosenKey labels the cards a ChooseCardFromEachGraveyard picks
// from every player's graveyard so the paired ReanimateLinkedCards puts exactly
// those cards onto the battlefield.
const massReanimationChosenKey = game.LinkedKey("mass-reanimation-chosen")

// lowerMassReanimationFromEachGraveyardSequence lowers the mass reanimation base
// "[Each player mills N cards.] For each player, choose a creature [or
// planeswalker] card in that player's graveyard. Put those cards onto the
// battlefield under your control. [Then each creature you control becomes a
// Phyrexian in addition to its other types.]" (Breach the Multiverse). The
// optional leading group mill fills every graveyard, one chooser picks a matching
// card in each player's graveyard, and the chosen cards enter the battlefield at
// once under the controller's control. An optional trailing controlled-creatures
// type grant (the #3138 static-subject rider) permanently adds a type to the
// creatures the controller controls once the reanimated cards have entered.
//
// It owns any sequence carrying an EffectChooseFromEachGraveyard so the new
// effect never falls through to the generic per-effect lowering. Any surrounding
// shape it does not fully model — a target, condition, mode, keyword, an
// unexpected effect, or an unsupported rider — fails closed with a diagnostic
// rather than lowering a partial reanimation.
func lowerMassReanimationFromEachGraveyardSequence(ctx contentCtx) (game.AbilityContent, *shared.Diagnostic, bool) {
	chooseIdx := -1
	for i := range ctx.content.Effects {
		if ctx.content.Effects[i].Kind == compiler.EffectChooseFromEachGraveyard {
			if chooseIdx >= 0 {
				return game.AbilityContent{},
					unsupportedEffectSequenceDiagnostic(ctx, "structural — multiple per-player graveyard choices"),
					true
			}
			chooseIdx = i
		}
	}
	if chooseIdx < 0 {
		return game.AbilityContent{}, nil, false
	}
	unsupported := func(reason string) (game.AbilityContent, *shared.Diagnostic, bool) {
		return game.AbilityContent{}, unsupportedEffectSequenceDiagnostic(ctx, reason), true
	}
	if ctx.optional ||
		len(ctx.content.Targets) != 0 ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Keywords) != 0 ||
		len(ctx.content.Modes) != 0 {
		return unsupported("structural — mass reanimation carries unsupported riders")
	}
	// The choose must be preceded by at most the group mill and followed by the
	// put and at most one rider, so any other effect layout — including a choose
	// with no following put — fails closed rather than indexing past the effects.
	if chooseIdx > 1 ||
		len(ctx.content.Effects) < chooseIdx+2 ||
		len(ctx.content.Effects) > chooseIdx+3 {
		return unsupported("structural — unexpected mass reanimation effect layout")
	}
	var sequence []game.Instruction
	if chooseIdx == 1 {
		mill := ctx.content.Effects[0]
		if mill.Kind != compiler.EffectMill ||
			mill.Context != parser.EffectContextEachPlayer ||
			!mill.Exact ||
			mill.Negated ||
			mill.Optional ||
			mill.DelayedTiming != 0 ||
			len(mill.References) != 0 ||
			!mill.Amount.Known ||
			mill.Amount.Value < 1 {
			return unsupported("structural — unsupported mass reanimation mill")
		}
		millAmount, ok := cardCountQuantity(mill.Amount, false)
		if !ok {
			return unsupported("structural — unsupported mass reanimation mill amount")
		}
		sequence = append(sequence, game.Instruction{Primitive: game.Mill{
			Amount:      millAmount,
			PlayerGroup: game.AllPlayersReference(),
		}})
	}
	choose := ctx.content.Effects[chooseIdx]
	if choose.Context != parser.EffectContextEachPlayer ||
		!choose.Exact ||
		choose.Negated ||
		len(choose.References) != 0 {
		return unsupported("structural — unsupported per-player graveyard choice")
	}
	selection, ok := SelectionForSelector(choose.Selector)
	if !ok {
		return unsupported("structural — unsupported per-player graveyard choice filter")
	}
	put := ctx.content.Effects[chooseIdx+1]
	if put.Kind != compiler.EffectPut ||
		put.Context != parser.EffectContextController ||
		put.ToZone != zone.Battlefield ||
		put.Negated ||
		put.Optional ||
		put.DelayedTiming != 0 ||
		len(put.References) != 1 ||
		put.References[0].Pronoun != compiler.ReferencePronounThose {
		return unsupported("structural — unsupported mass reanimation put")
	}

	var riderInstruction *game.Instruction
	if len(ctx.content.Effects) == chooseIdx+3 {
		rider := ctx.content.Effects[chooseIdx+2]
		instruction, ok := lowerMassReanimationRider(rider)
		if !ok {
			return unsupported("structural — unsupported mass reanimation rider")
		}
		riderInstruction = &instruction
	}

	sequence = append(sequence,
		game.Instruction{Primitive: game.ChooseCardFromEachGraveyard{
			Chooser:   game.ControllerReference(),
			Players:   game.AllPlayersReference(),
			Selection: selection,
			Optional:  choose.Optional,
			LinkedKey: massReanimationChosenKey,
		}},
		game.Instruction{Primitive: game.ReanimateLinkedCards{
			Controller: game.ControllerReference(),
			LinkedKey:  massReanimationChosenKey,
		}},
	)
	if riderInstruction != nil {
		sequence = append(sequence, *riderInstruction)
	}
	return game.Mode{Sequence: sequence}.Ability(), nil, true
}

// lowerMassReanimationRider lowers the optional trailing characteristic grant a
// mass reanimation applies once the chosen cards have entered. It handles the
// controlled-creatures static-subject grant "Then each creature you control
// becomes a <type> in addition to its other types." (Breach the Multiverse, the
// #3138 rider), returning the ApplyContinuous instruction. The grant snapshots
// the controller's creatures itself, so it needs no published returned group.
// Any other rider shape (for example a plural returned-group rider such as
// "They're Zombies in addition to their other types.") fails closed.
func lowerMassReanimationRider(rider compiler.CompiledEffect) (game.Instruction, bool) {
	if rider.Kind != compiler.EffectBecomeType ||
		rider.StaticSubject != compiler.StaticSubjectControlledCreatures ||
		rider.BecomeTypeUntilEndOfTurn ||
		rider.Negated ||
		rider.Optional ||
		(len(rider.BecomeTypeAddTypes) == 0 && len(rider.BecomeTypeAddSubtypes) == 0) {
		return game.Instruction{}, false
	}
	group, ok := resolvingStaticSubjectGroup(&rider)
	if !ok {
		return game.Instruction{}, false
	}
	continuousEffects := becomeTypeContinuousEffects(&rider)
	for i := range continuousEffects {
		continuousEffects[i].Group = group
	}
	return game.Instruction{Primitive: game.ApplyContinuous{
		ContinuousEffects: continuousEffects,
		Duration:          game.DurationPermanent,
	}}, true
}
