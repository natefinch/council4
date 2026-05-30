package rules

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/mana"
	"strings"
)

type manaOutput struct {
	color  mana.Color
	amount int
	snow   bool
}

func permanentManaOutput(g *game.Game, permanent *game.Permanent) (manaOutput, bool) {
	if color, ok := basicLandManaColor(g, permanent); ok {
		return manaOutput{color: color, amount: 1, snow: permanentIsSnow(g, permanent)}, true
	}
	_, ability, ok := simpleTapManaAbility(g, permanent)
	if !ok {
		return manaOutput{}, false
	}
	amount := ability.Effects[0].Amount
	if amount <= 0 {
		amount = 1
	}
	return manaOutput{color: ability.Effects[0].ManaColor, amount: amount, snow: permanentIsSnow(g, permanent)}, true
}

func basicLandManaColor(g *game.Game, permanent *game.Permanent) (mana.Color, bool) {
	card, ok := permanentCardDef(g, permanent)
	if !ok || !card.HasType(game.TypeLand) {
		return 0, false
	}
	for _, landType := range basicLandTypes {
		if card.HasSubtype(landType.subtype) || strings.EqualFold(card.Name, landType.subtype) {
			return landType.color, true
		}
	}
	return 0, false
}

func permanentIsSnow(g *game.Game, permanent *game.Permanent) bool {
	return permanentHasSupertype(g, permanent, game.Snow)
}

var basicLandTypes = []struct {
	subtype string
	color   mana.Color
}{
	{subtype: "Plains", color: mana.White},
	{subtype: "Island", color: mana.Blue},
	{subtype: "Swamp", color: mana.Black},
	{subtype: "Mountain", color: mana.Red},
	{subtype: "Forest", color: mana.Green},
}

func simpleTapManaAbility(g *game.Game, permanent *game.Permanent) (int, *game.AbilityDef, bool) {
	card, ok := permanentCardDef(g, permanent)
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
			if permanentHasType(g, permanent, game.TypeCreature) && permanent.SummoningSick {
				return 0, nil, false
			}
			return i, ability, true
		}
	}
	return 0, nil, false
}

func convokeCandidates(g *game.Game, playerID game.PlayerID, exclude map[id.ID]bool) []*game.Permanent {
	var nonMana []*game.Permanent
	var manaCreatures []*game.Permanent
	for _, permanent := range g.Battlefield {
		if !canConvokeWith(g, playerID, permanent, exclude) {
			continue
		}
		if _, ok := permanentManaOutput(g, permanent); ok {
			manaCreatures = append(manaCreatures, permanent)
			continue
		}
		nonMana = append(nonMana, permanent)
	}
	return append(nonMana, manaCreatures...)
}

func canConvokeWith(g *game.Game, playerID game.PlayerID, permanent *game.Permanent, exclude map[id.ID]bool) bool {
	if exclude[permanent.ObjectID] || permanent.Tapped || permanent.PhasedOut || effectiveController(g, permanent) != playerID {
		return false
	}
	return permanentHasType(g, permanent, game.TypeCreature)
}

func delveCandidates(g *game.Game, playerID game.PlayerID, cost *mana.Cost, xValue int, sourceCardID id.ID, sourceZone game.ZoneType) ([]id.ID, int, bool) {
	_, generic, ok := costRequirements(cost, xValue)
	if !ok || generic <= 0 {
		return nil, 0, false
	}
	player, ok := playerByID(g, playerID)
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

func convokePayment(g *game.Game, playerID game.PlayerID, cost *mana.Cost, xValue int, exclude map[id.ID]bool) ([]*game.Permanent, *mana.Cost, bool) {
	_, generic, ok := costRequirements(cost, xValue)
	if !ok {
		return nil, cost, false
	}
	candidates := convokeCandidates(g, playerID, exclude)
	paidColored := make(map[int]bool)
	var taps []*game.Permanent
	used := make(map[id.ID]bool)
	if cost != nil {
		for symbolIndex, symbol := range *cost {
			if symbol.Kind != mana.ColoredSymbol {
				continue
			}
			permanent, ok := chooseConvokeColoredCreature(g, candidates, used, symbol.Color)
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

func chooseConvokeColoredCreature(g *game.Game, candidates []*game.Permanent, used map[id.ID]bool, color mana.Color) (*game.Permanent, bool) {
	for _, permanent := range candidates {
		if used[permanent.ObjectID] {
			continue
		}
		for _, permanentColor := range permanentEffectiveColors(g, permanent) {
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
func availableManaSources(g *game.Game, playerID game.PlayerID, exclude map[id.ID]bool) map[mana.Color][]manaSource {
	available := make(map[mana.Color][]manaSource)
	for _, permanent := range g.Battlefield {
		if effectiveController(g, permanent) != playerID || permanent.Tapped || exclude[permanent.ObjectID] {
			continue
		}
		output, ok := permanentManaOutput(g, permanent)
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
