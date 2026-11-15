package rules

import (
	"fmt"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/opt"
)

// manaCostIncludesManaPayment reports whether manaCost includes a mana payment
// once each Phyrexian symbol announced as a life payment is paid with life. A
// {B} that an effect such as K'rrik, Son of Yawgmoth lets the caster pay with
// life remains a mana payment: that effect changes only how the cost is paid
// (CR 601.2h), not the total cost.
func manaCostIncludesManaPayment(manaCost *cost.Mana, announced map[int]bool) bool {
	if manaCost == nil {
		return false
	}
	for i, symbol := range *manaCost {
		if !announced[i] && (cost.Mana{symbol}).ManaValue() > 0 {
			return true
		}
	}
	return false
}

// paymentManaAbilityOptions offers "Finish activating mana abilities" and then
// each of actions, identified by its source: a permanent by its object ID and
// current public characteristics (a face-down one only as face-down, CR 708.2),
// so distinct same-named permanents and tokens stay distinguishable, and a card
// in player's hand by its card instance. A label names the source before the
// ability's text, adding the source's ID when sources share a name.
func paymentManaAbilityOptions(g *game.Game, player game.PlayerID, actions []action.Action) []game.ChoiceOption {
	type source struct {
		name      string
		id        id.ID
		permanent id.ID
		card      opt.V[game.ChoiceCardInfo]
	}
	sources := make([]source, len(actions))
	namedSources := map[string]map[id.ID]bool{}
	for i, candidate := range actions {
		activation, _ := candidate.ActivateAbilityPayload()
		src := source{id: activation.SourceID}
		if permanent, ok := permanentByObjectID(g, activation.SourceID); ok {
			src.name, src.permanent, src.card = permanentChoiceLabel(g, permanent), permanent.ObjectID, permanentChoiceInfo(g, permanent)
		} else {
			src.name, src.card = cardChoiceLabel(g, activation.SourceID), cardChoiceInfo(g, activation.SourceID)
		}
		if namedSources[src.name] == nil {
			namedSources[src.name] = map[id.ID]bool{}
		}
		namedSources[src.name][src.id] = true
		sources[i] = src
	}
	options := []game.ChoiceOption{{Index: 0, Label: "Finish activating mana abilities"}}
	for i, candidate := range actions {
		activation, _ := candidate.ActivateAbilityPayload()
		src := sources[i]
		name := src.name
		if len(namedSources[name]) > 1 {
			name = fmt.Sprintf("%s [%d]", name, src.id)
		}
		label := name + ": " + activatedAbilityText(g, player, activation)
		options = append(options, game.ChoiceOption{Index: len(options), Label: label, Card: src.card, SourcePermanent: src.permanent})
	}
	return options
}
