package rules

import (
	"fmt"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/mana"
)

func (e *Engine) resolveResolutionChoice(g *game.Game, obj *game.StackObject, effect game.Effect, agents [game.NumPlayers]PlayerAgent, log *TurnLog) bool {
	if effect.Choice == nil {
		return true
	}
	request, values := resolutionChoiceRequest(g, obj, effect.Choice)
	if len(values) == 0 {
		return false
	}
	selected := e.chooseChoice(g, agents, request, log)
	if len(selected) != 1 {
		return false
	}
	result, ok := values[selected[0]]
	if !ok {
		return false
	}
	rememberResolutionChoice(obj, effect.LinkID, result)
	return true
}

func resolutionChoiceRequest(g *game.Game, obj *game.StackObject, choice *game.ResolutionChoice) (game.ChoiceRequest, map[int]game.ResolutionChoiceResult) {
	playerID := stackObjectController(obj)
	if choice.UsePlayer && choice.Player >= 0 && choice.Player < game.NumPlayers {
		playerID = choice.Player
	}
	options, values := resolutionChoiceOptions(g, playerID, choice)
	prompt := choice.Prompt
	if prompt == "" {
		prompt = defaultResolutionChoicePrompt(choice.Kind)
	}
	return game.ChoiceRequest{
		Kind:             game.ChoiceResolution,
		Player:           playerID,
		Prompt:           prompt,
		Options:          options,
		MinChoices:       1,
		MaxChoices:       1,
		DefaultSelection: firstResolutionChoiceDefault(options),
	}, values
}

func resolutionChoiceOptions(g *game.Game, playerID game.PlayerID, choice *game.ResolutionChoice) ([]game.ChoiceOption, map[int]game.ResolutionChoiceResult) {
	values := make(map[int]game.ResolutionChoiceResult)
	var options []game.ChoiceOption
	add := func(index int, label string, result game.ResolutionChoiceResult) {
		options = append(options, game.ChoiceOption{Index: index, Label: label})
		values[index] = result
	}
	switch choice.Kind {
	case game.ResolutionChoiceColor:
		colors := choice.Colors
		if len(colors) == 0 {
			colors = append(mana.AllColors(), mana.Colorless)
		}
		for i, color := range colors {
			add(i, color.String(), game.ResolutionChoiceResult{Kind: choice.Kind, Color: color})
		}
	case game.ResolutionChoiceCardType:
		types := choice.CardTypes
		if len(types) == 0 {
			types = []game.CardType{game.TypeArtifact, game.TypeCreature, game.TypeEnchantment, game.TypeInstant, game.TypeLand, game.TypePlaneswalker, game.TypeSorcery, game.TypeBattle, game.TypeKindred}
		}
		for i, cardType := range types {
			add(i, cardType.String(), game.ResolutionChoiceResult{Kind: choice.Kind, CardType: cardType})
		}
	case game.ResolutionChoicePlayer:
		index := 0
		for player := game.PlayerID(0); player < game.NumPlayers; player++ {
			if !isPlayerAlive(g, player) || !choicePlayerMatches(playerID, player, choice.PlayerRelation) {
				continue
			}
			add(index, fmt.Sprintf("Player %d", player), game.ResolutionChoiceResult{Kind: choice.Kind, Player: player})
			index++
		}
	case game.ResolutionChoiceCard:
		zone := choice.Zone
		if zone == game.ZoneNone {
			zone = game.ZoneHand
		}
		for i, cardID := range resolutionChoiceCardIDs(g, playerID, zone) {
			add(i, cardChoiceLabel(g, cardID), game.ResolutionChoiceResult{Kind: choice.Kind, CardID: cardID})
		}
	}
	return options, values
}

func choicePlayerMatches(controller, candidate game.PlayerID, relation game.PlayerRelation) bool {
	switch relation {
	case game.PlayerYou:
		return candidate == controller
	case game.PlayerOpponent:
		return candidate != controller
	case game.PlayerNotYou:
		return candidate != controller
	default:
		return true
	}
}

func resolutionChoiceCardIDs(g *game.Game, playerID game.PlayerID, zone game.ZoneType) []id.ID {
	player := playerByID(g, playerID)
	if player == nil {
		return nil
	}
	switch zone {
	case game.ZoneHand:
		return player.Hand.All()
	case game.ZoneGraveyard:
		return player.Graveyard.All()
	case game.ZoneExile:
		return player.Exile.All()
	case game.ZoneLibrary:
		return player.Library.All()
	default:
		return nil
	}
}

func defaultResolutionChoicePrompt(kind game.ResolutionChoiceKind) string {
	switch kind {
	case game.ResolutionChoiceColor:
		return "Choose a color."
	case game.ResolutionChoiceCardType:
		return "Choose a card type."
	case game.ResolutionChoicePlayer:
		return "Choose a player."
	case game.ResolutionChoiceCard:
		return "Choose a card."
	default:
		return "Choose."
	}
}

func firstResolutionChoiceDefault(options []game.ChoiceOption) []int {
	if len(options) == 0 {
		return nil
	}
	return []int{options[0].Index}
}

func rememberResolutionChoice(obj *game.StackObject, linkID string, result game.ResolutionChoiceResult) {
	if obj == nil || linkID == "" {
		return
	}
	if obj.ResolutionChoices == nil {
		obj.ResolutionChoices = make(map[string]game.ResolutionChoiceResult)
	}
	obj.ResolutionChoices[linkID] = result
}

func linkedResolutionChoice(obj *game.StackObject, linkID string) (game.ResolutionChoiceResult, bool) {
	if obj == nil || linkID == "" || obj.ResolutionChoices == nil {
		return game.ResolutionChoiceResult{}, false
	}
	result, ok := obj.ResolutionChoices[linkID]
	return result, ok
}
