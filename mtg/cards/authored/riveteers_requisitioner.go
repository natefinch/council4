package authored

import (
	"github.com/natefinch/council4/mtg/cards/cardset"
	tokenT "github.com/natefinch/council4/mtg/cards/tokens/t"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

const (
	riveteersRequisitionerOracleID   = "469f25de-c180-47e9-9b1b-e23269139bbd"
	riveteersRequisitionerDiesText   = `When this creature dies, create a Treasure token. (It's an artifact with "{T}, Sacrifice this token: Add one mana of any color.")`
	riveteersRequisitionerOracleText = `When this creature dies, create a Treasure token. (It's an artifact with "{T}, Sacrifice this token: Add one mana of any color.")
Blitz {2}{R} (If you cast this spell for its blitz cost, it gains haste and "When this creature dies, draw a card." Sacrifice it at the beginning of the next end step.)`
)

var RiveteersRequisitioner = cardset.Entry{
	OracleID:     riveteersRequisitionerOracleID,
	Name:         "Riveteers Requisitioner",
	New:          newRiveteersRequisitioner,
	Verification: "mtg/rules/riveteers_requisitioner_test.go: TestRiveteersRequisitionerBlitzDiesCreatesTreasureAndDraws",
}

func newRiveteersRequisitioner() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name:             "Riveteers Requisitioner",
			ManaCost:         opt.Val(cost.Mana{cost.O(1), cost.R}),
			AlternativeCosts: []cost.Alternative{cost.BlitzAlternative(cost.Mana{cost.O(2), cost.R})},
			Types:            []types.Card{types.Creature},
			Subtypes:         []types.Sub{"Lizard", "Rogue"},
			Colors:           []color.Color{color.Red},
			Power:            opt.Val(game.PT{Value: 3}),
			Toughness:        opt.Val(game.PT{Value: 1}),
			OracleText:       riveteersRequisitionerOracleText,
			TriggeredAbilities: []game.TriggeredAbility{{
				Text: riveteersRequisitionerDiesText,
				Trigger: game.TriggerCondition{Type: game.TriggerWhen, Pattern: game.TriggerPattern{
					Event: game.EventPermanentDied, Source: game.TriggerSourceSelf,
					SubjectSelection: game.Selection{RequiredTypes: []types.Card{types.Creature}},
				}},
				Content: game.Mode{Text: riveteersRequisitionerDiesText, Sequence: []game.Instruction{{
					Primitive: game.CreateToken{
						Amount: game.Fixed(1),
						Source: game.TokenDef(tokenT.TreasureToken3c5493746c3742e08d88a8555d46732d),
					},
				}}}.Ability(),
			}},
			StaticAbilities: []game.StaticAbility{game.BlitzStaticAbility()},
		},
	}
}
