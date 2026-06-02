package payment

import (
	"strings"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
)

type permanentManaOutputResult struct {
	color  mana.Color
	amount int
	snow   bool
}

// permanentManaOutput derives the mana output of a permanent by checking
// basic land types and simple tap mana abilities.
func permanentManaOutput(s State, permanent *game.Permanent) (permanentManaOutputResult, bool) {
	if c, ok := basicLandManaColor(s, permanent); ok {
		return permanentManaOutputResult{color: c, amount: 1, snow: s.PermanentHasSupertype(permanent, types.Snow)}, true
	}
	controller := s.EffectiveController(permanent)
	_, ability, ok := simpleTapManaAbility(s, controller, permanent)
	if !ok {
		return permanentManaOutputResult{}, false
	}
	amount := ability.Effects[0].Amount
	if amount <= 0 {
		amount = 1
	}
	return permanentManaOutputResult{color: ability.Effects[0].ManaColor, amount: amount, snow: s.PermanentHasSupertype(permanent, types.Snow)}, true
}

func basicLandManaColor(s State, permanent *game.Permanent) (mana.Color, bool) {
	card, ok := s.PermanentCardDef(permanent)
	if !ok || !card.HasType(types.Land) {
		return "", false
	}
	for _, landType := range basicLandTypes {
		if card.HasSubtype(landType.subtype) || strings.EqualFold(card.Name, string(landType.subtype)) {
			return landType.color, true
		}
	}
	return "", false
}

var basicLandTypes = []struct {
	subtype types.Sub
	color   mana.Color
}{
	{subtype: types.Plains, color: mana.W},
	{subtype: types.Island, color: mana.U},
	{subtype: types.Swamp, color: mana.B},
	{subtype: types.Mountain, color: mana.R},
	{subtype: types.Forest, color: mana.G},
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
		if _, ok := permanentManaOutput(s, permanent); ok {
			manaCreatures = append(manaCreatures, permanent)
			continue
		}
		nonMana = append(nonMana, permanent)
	}
	return append(nonMana, manaCreatures...)
}

func delveCandidates(s State, playerID game.PlayerID, cost *cost.Mana, xValue int, sourceCardID id.ID, sourceZone game.ZoneType) ([]id.ID, int, bool) {
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

func convokePayment(s State, playerID game.PlayerID, manaCost *cost.Mana, xValue int, exclude map[id.ID]bool) ([]*game.Permanent, *cost.Mana, bool) {
	_, generic, ok := costRequirements(manaCost, xValue)
	if !ok {
		return nil, manaCost, false
	}
	candidates := convokeCandidates(s, playerID, exclude)
	paidColored := make(map[int]bool)
	var taps []*game.Permanent
	used := make(map[id.ID]bool)
	if manaCost != nil {
		for symbolIndex, symbol := range *manaCost {
			for _, color := range symbol.Colors() {
				permanent, ok := chooseConvokeColoredCreature(s, candidates, used, color)
				if !ok {
					continue
				}
				taps = append(taps, permanent)
				used[permanent.ObjectID] = true
				paidColored[symbolIndex] = true
				break
			}
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
		return nil, manaCost, false
	}
	return taps, costWithConvokePayments(manaCost, genericReduction, paidColored), true
}

func chooseConvokeColoredCreature(s State, candidates []*game.Permanent, used map[id.ID]bool, m mana.Color) (*game.Permanent, bool) {
	if m == mana.C {
		// can't pay for colorless pips via convoke.
		return nil, false
	}
	for _, permanent := range candidates {
		if used[permanent.ObjectID] {
			continue
		}
		for _, c := range s.PermanentEffectiveColors(permanent) {
			if cost.ManaForColor(c) == m {
				return permanent, true
			}
		}
	}
	return nil, false
}

func costWithConvokePayments(manaCost *cost.Mana, genericReduction int, paidColored map[int]bool) *cost.Mana {
	generic := max(genericCostAmount(manaCost)-genericReduction, 0)
	var modified cost.Mana
	if generic > 0 {
		modified = append(modified, cost.O(generic))
	}
	if manaCost != nil {
		for i, symbol := range *manaCost {
			if symbol.Kind == cost.GenericSymbol || paidColored[i] {
				continue
			}
			modified = append(modified, symbol)
		}
	}
	return &modified
}

func costWithGenericRequirement(manaCost *cost.Mana, generic int) *cost.Mana {
	if generic < 0 {
		generic = 0
	}
	var modified cost.Mana
	if generic > 0 {
		modified = append(modified, cost.O(generic))
	}
	if manaCost != nil {
		for _, symbol := range *manaCost {
			if symbol.Kind == cost.GenericSymbol || symbol.Kind == cost.VariableSymbol {
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
		output, ok := permanentManaOutput(s, permanent)
		if !ok {
			continue
		}
		available[output.color] = append(available[output.color], manaSource{
			permanent: permanent,
			color:     output.color,
			amount:    output.amount,
			snow:      output.snow,
		})
	}
	return available
}
