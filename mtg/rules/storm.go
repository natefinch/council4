package rules

import "github.com/natefinch/council4/mtg/game"

func stormCopyCount(g *game.Game, spell *game.CardDef) int {
	if spell == nil || !spell.HasKeyword(game.Storm) {
		return 0
	}
	count := 0
	for _, event := range g.EventsThisTurn() {
		if event.Kind == game.EventSpellCast {
			count++
		}
	}
	return count
}

func createStormCopies(g *game.Game, original *game.StackObject, count int) {
	for range count {
		g.Stack.Push(&game.StackObject{
			ID:                  g.IDGen.Next(),
			Kind:                game.StackSpell,
			SourceID:            original.SourceID,
			Face:                original.Face,
			Controller:          original.Controller,
			Targets:             append([]game.Target(nil), original.Targets...),
			TargetCounts:        append([]int(nil), original.TargetCounts...),
			ChosenModes:         append([]int(nil), original.ChosenModes...),
			XValue:              original.XValue,
			KickerPaid:          original.KickerPaid,
			Flashback:           original.Flashback,
			Suspend:             original.Suspend,
			Copy:                true,
			AdditionalCostsPaid: append([]string(nil), original.AdditionalCostsPaid...),
		})
	}
}
