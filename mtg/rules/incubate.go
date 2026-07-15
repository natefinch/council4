package rules

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// handleIncubate performs the incubate keyword action (CR 701.55): the recipient
// creates an Incubator token and, when the resolved amount is positive, puts
// that many +1/+1 counters on it. The recipient defaults to the resolving
// object's controller and is otherwise the reference the primitive carries (an
// exiled permanent's last-known controller for "its controller incubates X").
// Incubate 0 still creates the token with no counters (CR 701.55a places no
// minimum), so the token is created regardless of the amount.
func handleIncubate(r *effectResolver, prim game.Incubate) effectResolved {
	amount := r.quantity(prim.Amount)
	amount = max(amount, 0)
	res := effectResolved{accepted: true, amount: amount}
	var recipientRef game.PlayerReference
	if prim.Recipient.Exists {
		recipientRef = prim.Recipient.Val
	}
	recipient, ok := r.recipientController(recipientRef)
	if !ok {
		return res
	}
	created, ok := createTokenPermanentsCollectingWithChoices(r.engine, r.game, recipient, incubatorTokenDef(), 1, false, r.agents, r.log)
	if !ok || len(created) == 0 {
		return res
	}
	if amount > 0 {
		addCountersToPermanentControlledBy(r.game, recipient, created[0], counter.PlusOnePlusOne, amount)
	}
	res.succeeded = true
	return res
}

// incubatorTokenDef builds the Incubator token created by the incubate keyword
// action (CR 701.55a-c): a colorless Incubator artifact token with
// "{2}: Transform this artifact." whose back face is a 0/0 colorless Phyrexian
// artifact creature. The token is created with no counters here; the incubate
// handler adds the +1/+1 counters. The counters carry through the transform, so
// the creature side has power and toughness equal to the counters placed.
func incubatorTokenDef() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:  "Incubator",
			Types: []types.Card{types.Artifact},
			ActivatedAbilities: []game.ActivatedAbility{
				{
					Text:           "{2}: Transform this artifact.",
					ManaCost:       opt.Val(cost.Mana{cost.O(2)}),
					ZoneOfFunction: zone.Battlefield,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.Transform{
									Object: game.SourcePermanentReference(),
								},
							},
						},
					}.Ability(),
				},
			},
		},
		Layout: game.LayoutDoubleFacedToken,
		Back: opt.Val(game.CardFace{
			Name:      "Phyrexian",
			Types:     []types.Card{types.Artifact, types.Creature},
			Subtypes:  []types.Sub{types.Phyrexian},
			Power:     opt.Val(game.PT{Value: 0}),
			Toughness: opt.Val(game.PT{Value: 0}),
		}),
	}
}
