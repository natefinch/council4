package rules

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// handleEmpowerJace performs the empower Jace keyword action (CR 701.71a). If
// the resolving controller controls no Jace planeswalker token, they create the
// predefined Jace token with 0 loyalty. They then choose among the Jace
// planeswalker tokens they control and put Amount loyalty counters on it.
// Candidates are gathered after creation so a replacement that created more
// than one Jace token (or none) is honored. No state-based actions are checked
// between creation and placement, so the 0-loyalty token survives to receive
// the counters.
func handleEmpowerJace(r *effectResolver, prim game.EmpowerJace) effectResolved {
	amount := max(r.quantity(prim.Amount), 0)
	res := effectResolved{accepted: true, amount: amount}
	controller := stackObjectController(r.obj)
	if len(controlledJacePlaneswalkerTokens(r.game, controller)) == 0 {
		createTokenPermanentsCollectingWithChoices(r.engine, r.game, controller, empowerJaceTokenDef(), 1, false, r.agents, r.log)
	}
	candidates := controlledJacePlaneswalkerTokens(r.game, controller)
	if len(candidates) == 0 {
		return res
	}
	chosen := candidates[0]
	if len(candidates) > 1 {
		var ok bool
		chosen, ok = r.chooseOnePermanentAmong(candidates, "Choose a Jace planeswalker token to empower")
		if !ok {
			return res
		}
	}
	res.succeeded = addCountersToPermanentControlledBy(r.game, controller, chosen, counter.Loyalty, amount)
	return res
}

// controlledJacePlaneswalkerTokens returns the Jace planeswalker tokens
// controller controls, judged by current characteristics. Nontoken Jace
// planeswalkers never qualify (CR 701.71a), and phased-out permanents are
// treated as though they don't exist (CR 702.26b).
func controlledJacePlaneswalkerTokens(g *game.Game, controller game.PlayerID) []*game.Permanent {
	var tokens []*game.Permanent
	for _, permanent := range g.Battlefield {
		if permanent.Token && !permanent.PhasedOut &&
			effectiveController(g, permanent) == controller &&
			permanentHasType(g, permanent, types.Planeswalker) &&
			permanentHasSubtype(g, permanent, types.Jace) {
			tokens = append(tokens, permanent)
		}
	}
	return tokens
}

// empowerJaceTokenDef builds the token defined by CR 701.71a: a blue Jace
// planeswalker token with 0 loyalty, "[−1]: Surveil 1," and "[−3]: Draw a
// card."
func empowerJaceTokenDef() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name:     "Jace",
			Colors:   []color.Color{color.Blue},
			Types:    []types.Card{types.Planeswalker},
			Subtypes: []types.Sub{types.Jace},
			Loyalty:  opt.Val(0),
			LoyaltyAbilities: []game.LoyaltyAbility{
				{
					Text:        "−1: Surveil 1.",
					LoyaltyCost: -1,
					Content: game.Mode{
						Sequence: []game.Instruction{{
							Primitive: game.Surveil{Amount: game.Fixed(1), Player: game.ControllerReference()},
						}},
					}.Ability(),
				},
				{
					Text:        "−3: Draw a card.",
					LoyaltyCost: -3,
					Content: game.Mode{
						Sequence: []game.Instruction{{
							Primitive: game.Draw{Amount: game.Fixed(1), Player: game.ControllerReference()},
						}},
					}.Ability(),
				},
			},
		},
	}
}
