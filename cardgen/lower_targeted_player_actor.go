package cardgen

import (
	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game"
)

type targetedPlayerActorProjection struct {
	body   compiler.AbilityContent
	effect compiler.CompiledEffect
	order  shared.SourceOrder
}

func targetedPlayerActorScope(body compiler.AbilityContent, effect compiler.CompiledEffect) *targetedPlayerActorProjection {
	target, ok := body.TargetedPlayerActor(effect)
	if !ok {
		return nil
	}
	if _, ok := targetPlayersGroupSpec(target); !ok {
		return nil
	}
	return &targetedPlayerActorProjection{body: body, effect: effect, order: target.Order}
}

func (scope *targetedPlayerActorProjection) valid() bool {
	if scope == nil {
		return false
	}
	target, ok := scope.body.TargetedPlayerActor(scope.effect)
	return ok && target.Order == scope.order
}

func targetedPlayerActor(primitive game.Primitive) bool {
	if value, ok := primitive.(game.Draw); ok {
		return value.PlayerGroup.Kind == game.PlayerGroupReferenceTargetedPlayers
	}
	if value, ok := primitive.(game.Discard); ok {
		return value.PlayerGroup.Kind == game.PlayerGroupReferenceTargetedPlayers
	}
	if value, ok := primitive.(game.MoveTopOfLibrary); ok {
		return value.PlayerGroup.Kind == game.PlayerGroupReferenceTargetedPlayers
	}
	if value, ok := primitive.(game.GainLife); ok {
		return value.PlayerGroup.Kind == game.PlayerGroupReferenceTargetedPlayers
	}
	if value, ok := primitive.(game.LoseLife); ok {
		return value.PlayerGroup.Kind == game.PlayerGroupReferenceTargetedPlayers
	}
	if value, ok := primitive.(game.SacrificePermanents); ok {
		return value.PlayerGroup.Kind == game.PlayerGroupReferenceTargetedPlayers
	}
	return false
}

func transformGroupActorQuantityTargets(
	player *game.PlayerReference, group game.PlayerGroupReference, amount *game.Quantity,
	transform targetIndexTransform, scope *targetedPlayerActorProjection,
) bool {
	if group.Kind == game.PlayerGroupReferenceTargetedPlayers && !scope.valid() {
		return false
	}
	return transformPlayerQuantityTargets(player, amount, transform)
}

func transformTargetedActorSequence(sequence []game.Instruction, transform targetIndexTransform, scope *targetedPlayerActorProjection) bool {
	for i := range sequence {
		primitive, ok := transformPrimitiveTargetIndicesWithActorScope(sequence[i].Primitive, transform, scope)
		if !ok {
			return false
		}
		sequence[i].Primitive = primitive
	}
	return true
}
