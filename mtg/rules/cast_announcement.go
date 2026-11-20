package rules

import (
	"fmt"
	"slices"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// chooseCastAnnouncements takes the caster's X, repeatable cost and mode choices
// for cast (CR 601.2b).
func (e *Engine) chooseCastAnnouncements(g *game.Game, player game.PlayerID, def *game.CardDef, cast action.CastSpellAction, free bool, agents [game.NumPlayers]PlayerAgent, log *TurnLog, announced ...func(action.CastSpellAction)) (action.CastSpellAction, bool) {
	publish := func() {
		for _, update := range announced {
			update(cast)
		}
	}
	private := cast.SourceZone == zone.Hand || cast.SourceZone == zone.Library
	manaX := costHasVariableMana(manaCostPtr(def.ManaCost))
	if free && manaX {
		cast.XValue = 0
	} else if manaX || additionalCostsUseX(def.AdditionalCosts) {
		minimum := 0
		if additionalCostsRequirePositiveX(def.AdditionalCosts) {
			minimum = 1
		}
		cast.XValue = e.chooseAnnouncementNumber(g, player, "Choose X.", minimum, max(minimum, cast.XValue), private, agents, log)
	}
	publish()
	if spellHasMultikicker(def) {
		cast.KickerCount = e.chooseAnnouncementNumber(g, player, "Choose how many times to pay multikicker.", 0,
			effectiveKickerCount(cast.KickerPaid, cast.KickerCount), private, agents, log)
		cast.KickerPaid = cast.KickerCount > 0
		publish()
	}
	if costs := def.ReplicateCosts(); len(costs) != 0 {
		counts := make([]int, len(costs))
		copy(counts, cast.ReplicateCounts)
		for i, manaCost := range costs {
			counts[i] = e.chooseAnnouncementNumber(g, player,
				fmt.Sprintf("Choose how many times to pay replicate %d (%s).", i+1, manaCost),
				0, counts[i], private, agents, log)
		}
		cast.ReplicateCounts = counts
		publish()
	}
	branch := castBranchForCast(cast)
	if !modesValidForSpellAtBranch(g, player, def, cast.ChosenModes, branch) {
		modes := modeChoicesForSpellAtBranch(g, player, def, branch)
		if len(modes) == 0 {
			return cast, false
		}
		options := make([]game.ChoiceOption, len(modes))
		for index, choice := range modes {
			options[index] = game.ChoiceOption{Index: index, Label: fmt.Sprintf("Modes %v", choice)}
		}
		selected := e.chooseChoice(g, agents, game.ChoiceRequest{
			Kind: game.ChoiceModal, Player: player, Prompt: "Choose spell modes.", Options: options,
			MinChoices: 1, MaxChoices: 1, DefaultSelection: []int{0}, Private: private,
		}, log)
		cast.ChosenModes = slices.Clone(modes[selected[0]])
		publish()
	}
	return cast, true
}

func (e *Engine) chooseAnnouncementNumber(g *game.Game, player game.PlayerID, prompt string, minimum, fallback int, private bool, agents [game.NumPlayers]PlayerAgent, log *TurnLog) int {
	selected := e.chooseChoice(g, agents, game.ChoiceRequest{
		Kind: game.ChoiceNumber, Player: player, Prompt: prompt, Private: private,
		NumberRange: opt.Val(game.NumberRange{Minimum: minimum}),
		MinChoices:  1, MaxChoices: 1, DefaultSelection: []int{fallback},
	}, log)
	return selected[0]
}
