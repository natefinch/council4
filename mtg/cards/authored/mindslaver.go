package authored

import (
	"github.com/natefinch/council4/mtg/cardprogram"
	"github.com/natefinch/council4/mtg/cards/cardset"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

const mindslaverOracleID = "a806f4ee-f48a-46c3-8772-5aaabebe4c7e"

var mindslaverControl = game.ProgramReference{ID: mindslaverOracleID + "/activate/0", Version: "1"}

// Mindslaver redirects decisions without changing seats or object control.
var Mindslaver = cardset.Entry{
	OracleID: mindslaverOracleID, Name: "Mindslaver", New: newMindslaver,
	Verification: "mtg/rules/mindslaver_test.go: TestMindslaverCastingActivationCostsTargetsSourceAndCounter; TestMindslaverWholeTurnActionsModesTargetsAndResolvingChoices",
	Programs:     []cardprogram.Program{{Reference: mindslaverControl, Resolve: resolveMindslaver}},
}

func newMindslaver() *game.CardDef {
	const text = "{4}, {T}, Sacrifice Mindslaver: You control target player during that player's next turn. (You see all cards that player could see and make all decisions for them.)"
	return &game.CardDef{CardFace: game.CardFace{
		Name: "Mindslaver", ManaCost: opt.Val(cost.Mana{cost.O(6)}),
		Supertypes: []types.Super{types.Legendary}, Types: []types.Card{types.Artifact},
		OracleText: text,
		ActivatedAbilities: []game.ActivatedAbility{{
			Text:     text,
			ManaCost: opt.Val(cost.Mana{cost.O(4)}),
			AdditionalCosts: []cost.Additional{
				{Kind: cost.AdditionalTap}, {Kind: cost.AdditionalSacrificeSource},
			},
			Content: game.Mode{
				Text:    "Control target player's next actual turn",
				Targets: []game.TargetSpec{{MinTargets: 1, MaxTargets: 1, Allow: game.TargetAllowPlayer}},
				Program: mindslaverControl,
			}.Ability(),
		}},
	}}
}

func resolveMindslaver(ctx cardprogram.Context) error {
	_, err := ctx.ControlNextTurn(game.TargetPlayerReference(0))
	return err
}
