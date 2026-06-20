package rules

import (
	"fmt"
	"slices"

	"github.com/natefinch/council4/mtg/game/zone"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
)

func (e *Engine) resolveResolutionChoiceValue(g *game.Game, obj *game.StackObject, choice *game.ResolutionChoice, key string, agents [game.NumPlayers]PlayerAgent, log *TurnLog) bool {
	request, values := resolutionChoiceRequest(g, obj, choice)
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
	rememberResolutionChoice(obj, key, result)
	return true
}

// chooseEntryColor prompts the given player to make an entry-time color choice
// and returns the chosen result. It mirrors resolveResolutionChoiceValue but
// targets a permanent rather than a stack object, since entry choices are stored
// on the permanent (CR 614.12) rather than on a resolving stack object.
func (e *Engine) chooseEntryColor(g *game.Game, agents [game.NumPlayers]PlayerAgent, player game.PlayerID, choice *game.ResolutionChoice, log *TurnLog) (game.ResolutionChoiceResult, bool) {
	options, values := resolutionChoiceOptions(g, nil, player, choice)
	if len(values) == 0 {
		return game.ResolutionChoiceResult{}, false
	}
	prompt := choice.Prompt
	if prompt == "" {
		prompt = defaultResolutionChoicePrompt(choice.Kind)
	}
	request := game.ChoiceRequest{
		Kind:             game.ChoiceResolution,
		Player:           player,
		Prompt:           prompt,
		Options:          options,
		MinChoices:       1,
		MaxChoices:       1,
		DefaultSelection: firstResolutionChoiceDefault(options),
	}
	selected := e.chooseChoice(g, agents, request, log)
	if len(selected) != 1 {
		return game.ResolutionChoiceResult{}, false
	}
	result, ok := values[selected[0]]
	if !ok {
		return game.ResolutionChoiceResult{}, false
	}
	return result, true
}

func resolutionChoiceRequest(g *game.Game, obj *game.StackObject, choice *game.ResolutionChoice) (request game.ChoiceRequest, values map[int]game.ResolutionChoiceResult) {
	playerID, ok := resolutionChoicePlayerForStack(g, obj, choice)
	if !ok {
		return game.ChoiceRequest{}, nil
	}
	options, values := resolutionChoiceOptions(g, obj, playerID, choice)
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

func resolutionChoiceOptions(g *game.Game, obj *game.StackObject, playerID game.PlayerID, choice *game.ResolutionChoice) (options []game.ChoiceOption, values map[int]game.ResolutionChoiceResult) {
	values = make(map[int]game.ResolutionChoiceResult)
	add := func(index int, label string, result game.ResolutionChoiceResult) {
		options = append(options, game.ChoiceOption{Index: index, Label: label})
		values[index] = result
	}
	switch choice.Kind {
	case game.ResolutionChoiceMana:
		for i, color := range resolutionChoiceMana(g, obj, playerID, choice) {
			add(i, string(color), game.ResolutionChoiceResult{Kind: choice.Kind, Color: color})
		}
	case game.ResolutionChoiceCardType:
		cardTypes := choice.CardTypes
		if len(cardTypes) == 0 {
			cardTypes = []types.Card{types.Artifact, types.Creature, types.Enchantment, types.Instant, types.Land, types.Planeswalker, types.Sorcery, types.Battle, types.Kindred}
		}
		for i, cardType := range cardTypes {
			add(i, string(cardType), game.ResolutionChoiceResult{Kind: choice.Kind, CardType: cardType})
		}
	case game.ResolutionChoiceSubtype:
		for i, subtype := range types.SubtypesForType(choice.SubtypeOfType) {
			add(i, string(subtype), game.ResolutionChoiceResult{Kind: choice.Kind, Subtype: subtype})
		}
	case game.ResolutionChoicePlayer:
		index := 0
		for player := range game.PlayerID(game.NumPlayers) {
			if !isPlayerAlive(g, player) || !choicePlayerMatches(playerID, player, choice.PlayerRelation) {
				continue
			}
			add(index, fmt.Sprintf("Player %d", player), game.ResolutionChoiceResult{Kind: choice.Kind, Player: player})
			index++
		}
	case game.ResolutionChoiceCard:
		choiceZone := choice.Zone
		if choiceZone == zone.None {
			choiceZone = zone.Hand
		}
		for i, cardID := range resolutionChoiceCardIDs(g, playerID, choiceZone) {
			add(i, cardChoiceLabel(g, cardID), game.ResolutionChoiceResult{Kind: choice.Kind, CardID: cardID})
		}
	case game.ResolutionChoiceNumber:
		index := 0
		for number := choice.MinNumber; number <= choice.MaxNumber; number++ {
			add(index, fmt.Sprint(number), game.ResolutionChoiceResult{Kind: choice.Kind, Number: number})
			index++
		}
	default:
	}
	return options, values
}

func resolutionChoicePlayerForStack(g *game.Game, obj *game.StackObject, choice *game.ResolutionChoice) (game.PlayerID, bool) {
	if choice != nil && choice.PlayerReference != nil {
		return resolvePlayerReference(g, obj, *choice.PlayerReference)
	}
	return resolutionChoicePlayer(stackObjectController(obj), choice), true
}

func resolutionChoicePlayer(controller game.PlayerID, choice *game.ResolutionChoice) game.PlayerID {
	if choice != nil && choice.UsePlayer && choice.Player >= 0 && choice.Player < game.NumPlayers {
		return choice.Player
	}
	return controller
}

func resolutionChoiceMana(g *game.Game, obj *game.StackObject, playerID game.PlayerID, choice *game.ResolutionChoice) []mana.Color {
	if choice == nil {
		return nil
	}
	switch choice.ColorSource {
	case game.ResolutionChoiceColorSourceCommanderIdentity:
		return commanderColorIdentityMana(g, playerID)
	case game.ResolutionChoiceColorSourceFixedOrEntryChosen:
		return fixedOrEntryChosenMana(obj, choice)
	case game.ResolutionChoiceColorSourceLandsProduce:
		return landsProduceMana(g, playerID, choice)
	case game.ResolutionChoiceColorSourceLinkedExileColors:
		return linkedExileColorsMana(g, obj, choice)
	default:
		colors := choice.Colors
		if len(colors) == 0 {
			colors = []mana.Color{mana.W, mana.U, mana.B, mana.R, mana.G, mana.C}
		}
		return colors
	}
}

// fixedOrEntryChosenMana returns the fixed color of a composite "Add {C} or one
// mana of the chosen color." ability together with the color chosen as the
// source permanent entered, read from the stack object's seeded entry choice. The
// entry color is omitted when it was not recorded or duplicates the fixed color.
func fixedOrEntryChosenMana(obj *game.StackObject, choice *game.ResolutionChoice) []mana.Color {
	colors := append([]mana.Color(nil), choice.Colors...)
	result, ok := linkedResolutionChoice(obj, string(choice.EntryChoiceKey))
	if !ok || result.Kind != game.ResolutionChoiceMana {
		return colors
	}
	if slices.Contains(colors, result.Color) {
		return colors
	}
	return append(colors, result.Color)
}

// landsProduceMana returns, in WUBRG order (colorless last), every type of mana
// that a land matching the choice's player relation (relative to the choosing
// playerID) could currently produce (CR 106.7). It scans each battlefield land
// controlled by a matching player and unions the colors that land's mana
// abilities could add; when the choice includes colorless (the "any type"
// wording) it also offers {C} if a matching land could produce colorless. A mana
// ability whose color derives from this same source contributes nothing (handled
// in addInstructionManaColors and abilitiesProduceColorless), matching the
// loop-avoidance ruling for two opposing Exotic Orchards and bounding the scan.
// An empty result leaves the activating ability unactivatable (CR 605.1a).
func landsProduceMana(g *game.Game, playerID game.PlayerID, choice *game.ResolutionChoice) []mana.Color {
	var found colorSet
	colorlessFound := false
	for _, permanent := range g.Battlefield {
		if permanent == nil || permanent.PhasedOut {
			continue
		}
		if !permanentHasType(g, permanent, types.Land) {
			continue
		}
		if !choicePlayerMatches(playerID, effectiveController(g, permanent), choice.PlayerRelation) {
			continue
		}
		values := effectivePermanentValues(g, permanent)
		_, colors := abilitiesManaProduction(values.abilities, permanent.EntryChoices)
		for _, c := range colors {
			found.add(c)
		}
		if choice.IncludeColorless && !colorlessFound &&
			abilitiesProduceColorless(values.abilities, permanent.EntryChoices) {
			colorlessFound = true
		}
	}
	colors := found.ordered()
	manaColors := make([]mana.Color, 0, len(colors)+1)
	for _, c := range colors {
		manaColors = append(manaColors, cost.ManaForColor(c))
	}
	if colorlessFound {
		manaColors = append(manaColors, mana.C)
	}
	return manaColors
}

// linkedExileColorsMana returns, in WUBRG order, the colors of the card linked
// to the source permanent under the choice's LinkID — the imprinted card exiled
// from hand as the permanent entered. It models "Add one mana of any of the
// exiled card's colors." (Chrome Mox). The link is read by the permanent's
// object identity, so a re-entered object with no fresh imprint finds nothing.
// A missing, declined (no link recorded), or colorless imprint yields an empty
// set, leaving the mana ability unactivatable (CR 605.1a, CR 202.2); a
// multicolored imprint yields exactly its colors. Colorless ({C}) is never
// offered because a card's colors are only the five colors (CR 105.2, CR 202.2).
func linkedExileColorsMana(g *game.Game, obj *game.StackObject, choice *game.ResolutionChoice) []mana.Color {
	if obj == nil || choice == nil || choice.LinkID == "" {
		return nil
	}
	var found colorSet
	for _, ref := range linkedObjects(g, linkedObjectByObjectKey(g, obj, choice.LinkID)) {
		cardID := ref.CardID
		if cardID == 0 {
			cardID = ref.ObjectID
		}
		card, ok := g.GetCardInstance(cardID)
		if !ok {
			continue
		}
		faceDef := cardFaceOrDefault(card, game.FaceFront)
		if faceDef == nil {
			continue
		}
		for _, c := range faceDef.Colors {
			found.add(c)
		}
	}
	colors := found.ordered()
	manaColors := make([]mana.Color, 0, len(colors))
	for _, c := range colors {
		manaColors = append(manaColors, cost.ManaForColor(c))
	}
	return manaColors
}

func commanderColorIdentityMana(g *game.Game, playerID game.PlayerID) []mana.Color {
	player, ok := playerByID(g, playerID)
	if !ok || player.CommanderInstanceID == 0 {
		return nil
	}
	card, ok := g.GetCardInstance(player.CommanderInstanceID)
	if !ok || card.Def == nil {
		return nil
	}
	colors := card.Def.ColorIdentity.Colors()
	if len(colors) == 0 {
		return nil
	}
	manaColors := make([]mana.Color, 0, len(colors))
	for _, c := range colors {
		manaColors = append(manaColors, cost.ManaForColor(c))
	}
	return manaColors
}

// commanderColorIdentityCount returns the number of colors in the player's
// commander's color identity (CR 903.4), zero when the player has no modeled
// commander or a colorless one. Partner commanders are not modeled, so it reads
// the single commander instance.
func commanderColorIdentityCount(g *game.Game, playerID game.PlayerID) int {
	player, ok := playerByID(g, playerID)
	if !ok || player.CommanderInstanceID == 0 {
		return 0
	}
	card, ok := g.GetCardInstance(player.CommanderInstanceID)
	if !ok || card.Def == nil {
		return 0
	}
	return card.Def.ColorIdentity.NumColors()
}

func choicePlayerMatches(controller, candidate game.PlayerID, relation game.PlayerRelation) bool {
	switch relation {
	case game.PlayerYou:
		return candidate == controller
	case game.PlayerOpponent, game.PlayerNotYou:
		return candidate != controller
	default:
		return true
	}
}

func resolutionChoiceCardIDs(g *game.Game, playerID game.PlayerID, zoneType zone.Type) []id.ID {
	player, ok := playerByID(g, playerID)
	if !ok {
		return nil
	}
	switch zoneType {
	case zone.Hand:
		return player.Hand.All()
	case zone.Graveyard:
		return player.Graveyard.All()
	case zone.Exile:
		return player.Exile.All()
	case zone.Library:
		return player.Library.All()
	default:
		return nil
	}
}

func defaultResolutionChoicePrompt(kind game.ResolutionChoiceKind) string {
	switch kind {
	case game.ResolutionChoiceMana:
		return "Choose a color."
	case game.ResolutionChoiceCardType:
		return "Choose a card type."
	case game.ResolutionChoiceSubtype:
		return "Choose a creature type."
	case game.ResolutionChoicePlayer:
		return "Choose a player."
	case game.ResolutionChoiceCard:
		return "Choose a card."
	case game.ResolutionChoiceNumber:
		return "Choose a number."
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

// seedEntryChoices copies the source permanent's entry-time choices onto a
// resolving stack object so instructions such as AddMana{EntryChoiceFrom:...}
// can read them through the normal resolution-choice lookup (CR 614.12).
func seedEntryChoices(obj *game.StackObject, permanent *game.Permanent) {
	if obj == nil || permanent == nil || len(permanent.EntryChoices) == 0 {
		return
	}
	if obj.ResolutionChoices == nil {
		obj.ResolutionChoices = make(map[string]game.ResolutionChoiceResult, len(permanent.EntryChoices))
	}
	for key, result := range permanent.EntryChoices {
		obj.ResolutionChoices[string(key)] = result
	}
}

// permanentEntryChoiceAvailable reports whether the source permanent holds a
// recorded entry-time choice under the given key.
func permanentEntryChoiceAvailable(permanent *game.Permanent, key game.ChoiceKey) bool {
	if permanent == nil || len(permanent.EntryChoices) == 0 {
		return false
	}
	_, ok := permanent.EntryChoices[key]
	return ok
}
