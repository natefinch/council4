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

const terastodonOracleID = "17975a20-2c13-4325-be1e-0bcda5063a2d"

var terastodonEntryProgram = game.ProgramReference{
	ID: terastodonOracleID + "/enter/0", Version: "1",
}

// Terastodon is a complete authored registration; no Oracle compiler admission is required.
var Terastodon = cardset.Entry{
	OracleID: terastodonOracleID, Name: "Terastodon", New: newTerastodon,
	Programs: []cardprogram.Program{{Reference: terastodonEntryProgram, Resolve: resolveTerastodonEntry}},
}

func newTerastodon() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Terastodon", ManaCost: opt.Val(cost.Mana{cost.O(6), cost.G, cost.G}),
			Types: []types.Card{types.Creature}, Subtypes: []types.Sub{types.Elephant},
			Colors: []color.Color{color.Green}, Power: opt.Val(game.PT{Value: 9}), Toughness: opt.Val(game.PT{Value: 9}),
			OracleText: "When this creature enters, you may destroy up to three target noncreature permanents. For each permanent put into a graveyard this way, its controller creates a 3/3 green Elephant creature token.",
			TriggeredAbilities: []game.TriggeredAbility{{
				Trigger: game.TriggerCondition{Pattern: game.TriggerPattern{
					Event: game.EventPermanentEnteredBattlefield, Source: game.TriggerSourceSelf,
				}},
				Content: game.Mode{
					Text: "Optional destruction with tokens only for actual graveyard arrivals",
					Targets: []game.TargetSpec{{
						MinTargets: 0, MaxTargets: 3, Allow: game.TargetAllowPermanent,
						Selection: opt.Val(game.Selection{ExcludedTypes: []types.Card{types.Creature}}),
					}},
					Program: terastodonEntryProgram,
				}.Ability(),
			}},
		},
	}
}

func resolveTerastodonEntry(ctx cardprogram.Context) error {
	selected, err := ctx.Ask(game.ChoiceRequest{
		Kind: game.ChoiceMay, Player: ctx.View().Source().Controller,
		Prompt:     "Destroy the chosen noncreature permanents?",
		Options:    []game.ChoiceOption{{Index: 0, Label: "No"}, {Index: 1, Label: "Yes"}},
		MinChoices: 1, MaxChoices: 1, DefaultSelection: []int{1},
	})
	if err != nil {
		return err
	}
	if selected[0] == 0 {
		return nil
	}
	outcome, err := ctx.Destroy(game.Destroy{Object: game.AllTargetPermanentsReference(0)})
	if err != nil {
		return err
	}
	recipients := make([]cardprogram.TokenRecipient, 0, len(outcome.Members))
	for _, member := range outcome.Members {
		if !member.Moved || member.Destination != zone.Graveyard {
			continue
		}
		recipients = append(recipients, cardprogram.TokenRecipient{Player: member.Before.Controller, Count: 1})
	}
	_, err = ctx.CreateTokens(elephantToken(), recipients...)
	return err
}

func elephantToken() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name: "Elephant", Types: []types.Card{types.Creature}, Subtypes: []types.Sub{types.Elephant},
		Colors: []color.Color{color.Green}, Power: opt.Val(game.PT{Value: 3}), Toughness: opt.Val(game.PT{Value: 3}),
	}}
}
