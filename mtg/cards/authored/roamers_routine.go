package authored

import (
	"github.com/natefinch/council4/mtg/cards/cardset"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

const roamersRoutineOracleID = "cf2355f3-11d2-4b03-b663-9fad6b07e4b2"
const roamersRoutineOracleText = "Search your library for a basic land card, put it onto the battlefield tapped, then shuffle.\nHarmonize {4}{G} (You may cast this card from your graveyard for its harmonize cost. You may tap a creature you control to reduce that cost by {X}, where X is its power. Then exile this spell.)"

var RoamersRoutine = cardset.Entry{
	OracleID: roamersRoutineOracleID,
	Name:     "Roamer's Routine",
	New:      newRoamersRoutine,
	Verification: "mtg/rules/roamers_routine_test.go: TestRoamersRoutineHarmonizeFindsBasicLandTapped; " +
		"printing fb8c2d5c-ba0c-4d50-8898-5c6574b1e974; Scryfall rulings retrieved " +
		"2026-11-21: 0; CR 2026-11-06 110.5b, 118.7a, 601.2b, 601.2f-h, " +
		"701.23a-b, 701.24a, 701.26a, 702.180a-b",
}

func newRoamersRoutine() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name:       "Roamer's Routine",
			ManaCost:   opt.Val(cost.Mana{cost.O(2), cost.G}),
			Types:      []types.Card{types.Sorcery},
			Colors:     []color.Color{color.Green},
			OracleText: roamersRoutineOracleText,
			AlternativeCosts: []cost.Alternative{
				cost.HarmonizeAlternative(cost.Mana{cost.O(4), cost.G}),
			},
			SpellAbility: opt.Val(game.Mode{Sequence: []game.Instruction{{
				Primitive: game.Search{
					Player: game.ControllerReference(),
					Spec: game.SearchSpec{
						SourceZone:   zone.Library,
						Destination:  zone.Battlefield,
						EntersTapped: true,
						Filter: game.Selection{
							RequiredTypes: []types.Card{types.Land},
							Supertypes:    []types.Super{types.Basic},
						},
					},
					Amount: game.Fixed(1),
				},
			}}}.Ability()),
		},
	}
}
