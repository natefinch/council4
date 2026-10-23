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

const devoutDecreeOracleID = "92f8c8b1-1fd5-4e80-b8a9-51671937aaf2"

var devoutDecreeSpell = game.ProgramReference{ID: devoutDecreeOracleID + "/spell/0", Version: "1"}

var DevoutDecree = cardset.Entry{
	OracleID: devoutDecreeOracleID,
	Name:     "Devout Decree",
	New:      newDevoutDecree,
	Programs: []cardprogram.Program{{Reference: devoutDecreeSpell, Resolve: resolveDevoutDecree}},
	Verification: "mtg/rules/scry_interaction_test.go: " +
		"TestDevoutDecreeDefaultRegistryPaidCastAndCharacteristicUnions; " +
		"TestDevoutDecreeRevalidatesTargetAndExilesThenScries; " +
		"TestDevoutDecreeIllegalTargetOmitsScry. " +
		"Sources: pinned oracle-cards.json SHA256 6da9ced0e19bd384771ab8f14051cc45a8dc90302082b875fa6e039d2399a1b4; " +
		"FDN 571 printing d174cb01-b1cf-444c-a893-cc61ddf88b8c; full " +
		"https://api.scryfall.com/cards/d174cb01-b1cf-444c-a893-cc61ddf88b8c/rulings " +
		"retrieved 2026-10-22 and retained in the rules test source record. Official CR " +
		"https://media.wizards.com/2026/downloads/MagicCompRules%2020260925.txt, effective 2026-09-25, " +
		"SHA256 8d860e451f20f38865b725b42d82feb714c725373dd8f3b32b8652b3eeb070ca: " +
		"105.2/205.1 and 613.1b/d/e describe current colors, types, and control layers; 115.1a/601.2c " +
		"and 608.2b require the type/color unions and current target legality; 608.2c/614.1 apply ordered, " +
		"replacement-aware exile before scry; 903.9a allows the exiled commander to move to the command " +
		"zone as a state-based action after the spell resolves, not by Sea God's hand-return replacement.",
}

func newDevoutDecree() *game.CardDef {
	const text = "Exile target creature or planeswalker that's black or red. Scry 1. (Look at the top card of your library. You may put that card on the bottom.)"
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name:       "Devout Decree",
			OracleText: text,
			ManaCost:   opt.Val(cost.Mana{cost.O(1), cost.W}),
			Types:      []types.Card{types.Sorcery},
			Colors:     []color.Color{color.White},
			SpellAbility: opt.Val(game.Mode{
				Text: text,
				Targets: []game.TargetSpec{{
					MinTargets: 1,
					MaxTargets: 1,
					Constraint: "target creature or planeswalker that's black or red",
					Allow:      game.TargetAllowPermanent,
					Selection: opt.Val(game.Selection{
						RequiredTypesAny: []types.Card{types.Creature, types.Planeswalker},
						ColorsAny:        []color.Color{color.Black, color.Red},
					}),
				}},
				Program: devoutDecreeSpell,
			}.Ability()),
		},
	}
}

func resolveDevoutDecree(ctx cardprogram.Context) error {
	if _, err := ctx.Execute(game.MovePermanent{
		Object:      game.TargetPermanentReference(0),
		Destination: zone.Exile,
	}); err != nil {
		return err
	}
	_, err := ctx.Execute(game.Scry{Player: game.ControllerReference(), Amount: game.Fixed(1)})
	return err
}
