package rules

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/opt"
)

func optCost(cost mana.Cost) opt.V[mana.Cost] {
	return opt.Val(cost)
}

func optPT(pt game.PT) opt.V[game.PT] {
	return opt.Val(pt)
}

func optInt(v int) opt.V[int] {
	return opt.Val(v)
}

func optDynamicValue(v game.DynamicValue) opt.V[game.DynamicValue] {
	return opt.Val(v)
}

func optDynamicAmount(v game.DynamicAmount) opt.V[game.DynamicAmount] {
	return opt.Val(v)
}

func optTrigger(v game.TriggerCondition) opt.V[game.TriggerCondition] {
	return opt.Val(v)
}

func optStateTrigger(v game.StateTriggerCondition) opt.V[game.StateTriggerCondition] {
	return opt.Val(v)
}

func optEffectCondition(v game.EffectCondition) opt.V[game.EffectCondition] {
	return opt.Val(v)
}

func optEffectResultCondition(v game.EffectResultCondition) opt.V[game.EffectResultCondition] {
	return opt.Val(v)
}

func optIntComparison(v compare.Int) opt.V[compare.Int] {
	return opt.Val(v)
}

func optResolutionChoice(v game.ResolutionChoice) opt.V[game.ResolutionChoice] {
	return opt.Val(v)
}

func optResolutionPayment(v game.ResolutionPayment) opt.V[game.ResolutionPayment] {
	return opt.Val(v)
}

func optToken(v *game.CardDef) opt.V[*game.CardDef] {
	return opt.Val(v)
}

func optDelayedTrigger(v game.DelayedTriggerDef) opt.V[game.DelayedTriggerDef] {
	return opt.Val(v)
}

func optReplacement(v game.ReplacementEffect) opt.V[game.ReplacementEffect] {
	return opt.Val(v)
}

func optCopyValues(v game.CopyableValues) opt.V[game.CopyableValues] {
	return opt.Val(v)
}

func optController(v game.PlayerID) opt.V[game.PlayerID] {
	return opt.Val(v)
}
