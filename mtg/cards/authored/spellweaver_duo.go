package authored

import (
	"github.com/natefinch/council4/mtg/cardprogram"
	"github.com/natefinch/council4/mtg/cards/cardset"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

const spellweaverDuoOracleID = "1fd4204f-f627-4919-9d48-79b172f4560e"

var spellweaverDuoEnterProgram = game.ProgramReference{
	ID: spellweaverDuoOracleID + "/enter/0", Version: "1",
}

var SpellweaverDuo = cardset.Entry{
	OracleID: spellweaverDuoOracleID,
	Name:     "Spellweaver Duo",
	New:      newSpellweaverDuo,
	Programs: []cardprogram.Program{{
		Reference: spellweaverDuoEnterProgram,
		Resolve:   resolveSpellweaverDuoEnter,
	}},
	Verification: "mtg/rules/spellweaver_duo_test.go: TestSpellweaverDuoAssistAndEntersTrigger; " +
		"CR 2026-09-25 601.2c/f-h, 603.2-3, 608.2b-d, 108.3, 400.7, 702.132a",
}

func newSpellweaverDuo() *game.CardDef {
	const enterText = "When this creature enters, you may return target tapped creature to its owner's hand."
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name:       "Spellweaver Duo",
			ManaCost:   opt.Val(cost.Mana{cost.O(6), cost.U}),
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Human, types.Wizard},
			Colors:     []color.Color{color.Blue},
			Power:      opt.Val(game.PT{Value: 4}),
			Toughness:  opt.Val(game.PT{Value: 4}),
			OracleText: "Assist (Another player can pay up to {6} of this spell's cost.)\n" + enterText,
			StaticAbilities: []game.StaticAbility{
				game.AssistStaticBody,
			},
			TriggeredAbilities: []game.TriggeredAbility{{
				Text: enterText,
				Trigger: game.TriggerCondition{Pattern: game.TriggerPattern{
					Event:  game.EventPermanentEnteredBattlefield,
					Source: game.TriggerSourceSelf,
				}},
				Content: game.Mode{
					Text: enterText,
					Targets: []game.TargetSpec{{
						MinTargets: 1,
						MaxTargets: 1,
						Allow:      game.TargetAllowPermanent,
						Selection: opt.Val(game.Selection{
							RequiredTypes: []types.Card{types.Creature},
							Tapped:        game.TriTrue,
						}),
					}},
					Program: spellweaverDuoEnterProgram,
				}.Ability(),
			}},
		},
	}
}

func resolveSpellweaverDuoEnter(ctx cardprogram.Context) error {
	selected, err := ctx.Ask(game.ChoiceRequest{
		Kind: game.ChoiceMay, Player: ctx.View().Source().Controller,
		Prompt:     "Return the target tapped creature to its owner's hand?",
		Options:    []game.ChoiceOption{{Index: 0, Label: "No"}, {Index: 1, Label: "Yes"}},
		MinChoices: 1, MaxChoices: 1, DefaultSelection: []int{1},
	})
	if err != nil {
		return err
	}
	if selected[0] == 0 {
		return nil
	}
	_, err = ctx.Execute(game.MovePermanent{
		Object:      game.TargetPermanentReference(0),
		Destination: zone.Hand,
	})
	return err
}
