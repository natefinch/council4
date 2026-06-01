package payment

import (
	"strings"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
)

// permanentManaOutput derives the mana output of a permanent by checking
// basic land types and simple tap mana abilities.
func permanentManaOutput(s State, permanent *game.Permanent) (color mana.Color, amount int, snow bool, ok bool) {
	if c, ok2 := basicLandManaColor(s, permanent); ok2 {
		return c, 1, s.PermanentHasSupertype(permanent, types.Snow), true
	}
	controller := s.EffectiveController(permanent)
	_, ability, ok2 := simpleTapManaAbility(s, controller, permanent)
	if !ok2 {
		return 0, 0, false, false
	}
	a := ability.Effects[0].Amount
	if a <= 0 {
		a = 1
	}
	return ability.Effects[0].ManaColor, a, s.PermanentHasSupertype(permanent, types.Snow), true
}

func basicLandManaColor(s State, permanent *game.Permanent) (mana.Color, bool) {
	card, ok := s.PermanentCardDef(permanent)
	if !ok || !card.HasType(types.Land) {
		return 0, false
	}
	for _, landType := range basicLandTypes {
		if card.HasSubtype(landType.subtype) || strings.EqualFold(card.Name, string(landType.subtype)) {
			return landType.color, true
		}
	}
	return 0, false
}

var basicLandTypes = []struct {
	subtype types.Sub
	color   mana.Color
}{
	{subtype: types.Plains, color: mana.White},
	{subtype: types.Island, color: mana.Blue},
	{subtype: types.Swamp, color: mana.Black},
	{subtype: types.Mountain, color: mana.Red},
	{subtype: types.Forest, color: mana.Green},
}

func simpleTapManaAbility(s State, playerID game.PlayerID, permanent *game.Permanent) (int, *game.AbilityDef, bool) {
	card, ok := s.PermanentCardDef(permanent)
	if !ok {
		return 0, nil, false
	}
	for i := range card.Abilities {
		ability := &card.Abilities[i]
		if ability.Kind == game.ActivatedAbility &&
			ability.IsManaAbility &&
			hasTapCost(ability) &&
			!ability.ManaCost.Exists &&
			len(ability.Targets) == 0 &&
			len(ability.Effects) == 1 &&
			ability.Effects[0].Type == game.EffectAddMana {
			if s.PermanentHasType(permanent, types.Creature) && permanent.SummoningSick {
				return 0, nil, false
			}
			if !s.ActivationConditionSatisfied(playerID, permanent, ability) {
				continue
			}
			return i, ability, true
		}
	}
	return 0, nil, false
}

func convokeCandidates(s State, playerID game.PlayerID, exclude map[id.ID]bool) []*game.Permanent {
	var nonMana []*game.Permanent
	var manaCreatures []*game.Permanent
	for _, permanent := range s.Battlefield() {
		if !canConvokeWith(s, playerID, permanent, exclude) {
			continue
		}
		if _, _, _, ok := permanentManaOutput(s, permanent); ok {
			manaCreatures = append(manaCreatures, permanent)
			continue
		}
		nonMana = append(nonMana, permanent)
	}
	return append(nonMana, manaCreatures...)
}

func delveCandidates(s State, playerID game.PlayerID, cost *mana.Cost, xValue int, sourceCardID id.ID, sourceZone game.ZoneType) ([]id.ID, int, bool) {
	_, generic, ok := costRequirements(cost, xValue)
	if !ok || generic <= 0 {
		return nil, 0, false
	}
	player, ok := s.Player(playerID)
	if !ok {
		return nil, 0, false
	}
	var exiles []id.ID
	for _, cardID := range player.Graveyard.All() {
		if len(exiles) == generic {
			break
		}
		if sourceZone == game.ZoneGraveyard && cardID == sourceCardID {
			continue
		}
		exiles = append(exiles, cardID)
	}
	if len(exiles) == 0 {
		return nil, 0, false
	}
	return exiles, generic, true
}

func convokePayment(s State, playerID game.PlayerID, cost *mana.Cost, xValue int, exclude map[id.ID]bool) ([]*game.Permanent, *mana.Cost, bool) {
	_, generic, ok := costRequirements(cost, xValue)
	if !ok {
		return nil, cost, false
	}
	candidates := convokeCandidates(s, playerID, exclude)
	paidColored := make(map[int]bool)
	var taps []*game.Permanent
	used := make(map[id.ID]bool)
	if cost != nil {
		for symbolIndex, symbol := range *cost {
			if symbol.Kind != mana.ColoredSymbol {
				continue
			}
			permanent, ok := chooseConvokeColoredCreature(s, candidates, used, symbol.Color)
			if !ok {
				continue
			}
			taps = append(taps, permanent)
			used[permanent.ObjectID] = true
			paidColored[symbolIndex] = true
		}
	}
	genericReduction := 0
	for _, permanent := range candidates {
		if genericReduction == generic {
			break
		}
		if used[permanent.ObjectID] {
			continue
		}
		taps = append(taps, permanent)
		used[permanent.ObjectID] = true
		genericReduction++
	}
	if len(taps) == 0 {
		return nil, cost, false
	}
	return taps, costWithConvokePayments(cost, genericReduction, paidColored), true
}

func chooseConvokeColoredCreature(s State, candidates []*game.Permanent, used map[id.ID]bool, color mana.Color) (*game.Permanent, bool) {
	for _, permanent := range candidates {
		if used[permanent.ObjectID] {
			continue
		}
		for _, permanentColor := range s.PermanentEffectiveColors(permanent) {
			if permanentColor == color {
				return permanent, true
			}
		}
	}
	return nil, false
}

func costWithConvokePayments(cost *mana.Cost, genericReduction int, paidColored map[int]bool) *mana.Cost {
	generic := genericCostAmount(cost) - genericReduction
	if generic < 0 {
		generic = 0
	}
	var modified mana.Cost
	if generic > 0 {
		modified = append(modified, mana.GenericMana(generic))
	}
	if cost != nil {
		for i, symbol := range *cost {
			if symbol.Kind == mana.GenericSymbol || paidColored[i] {
				continue
			}
			modified = append(modified, symbol)
		}
	}
	return &modified
}

func costWithGenericRequirement(cost *mana.Cost, generic int) *mana.Cost {
	if generic < 0 {
		generic = 0
	}
	var modified mana.Cost
	if generic > 0 {
		modified = append(modified, mana.GenericMana(generic))
	}
	if cost != nil {
		for _, symbol := range *cost {
			if symbol.Kind == mana.GenericSymbol || symbol.Kind == mana.VariableSymbol {
				continue
			}
			modified = append(modified, symbol)
		}
	}
	return &modified
}

// availableManaSources groups sources by color. Callers must consume it through
// paymentColors or explicit symbol colors, never by ranging over the map, so
// payment ordering remains deterministic.
func availableManaSources(s State, playerID game.PlayerID, exclude map[id.ID]bool) map[mana.Color][]manaSource {
	available := make(map[mana.Color][]manaSource)
	for _, permanent := range s.Battlefield() {
		if s.EffectiveController(permanent) != playerID || permanent.Tapped || exclude[permanent.ObjectID] {
			continue
		}
		color, amount, snow, ok := permanentManaOutput(s, permanent)
		if !ok {
			continue
		}
		available[color] = append(available[color], manaSource{
			permanent: permanent,
			color:     color,
			amount:    amount,
			snow:      snow,
		})
	}
	return available
}
