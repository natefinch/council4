package authored

import (
	"github.com/natefinch/council4/mtg/cardprogram"
	"github.com/natefinch/council4/mtg/cards/cardset"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

const hiveMindOracleID = "f97e405d-4c24-4d4d-9e19-e349c073113c"

var hiveMindCopyProgram = game.ProgramReference{ID: hiveMindOracleID + "/cast/0", Version: "1"}

// HiveMind declares its trigger separately from its ordinary Go copying body.
var HiveMind = cardset.Entry{
	OracleID: hiveMindOracleID, Name: "Hive Mind", New: newHiveMind,
	Programs: []cardprogram.Program{{Reference: hiveMindCopyProgram, Resolve: resolveHiveMindCopy}},
}

func newHiveMind() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Hive Mind", ManaCost: opt.Val(cost.Mana{cost.O(5), cost.U}),
			Types: []types.Card{types.Enchantment}, Colors: []color.Color{color.Blue},
			OracleText: "Whenever a player casts an instant or sorcery spell, each other player copies that spell. Each of those players may choose new targets for their copy.",
			TriggeredAbilities: []game.TriggeredAbility{{
				Trigger: game.TriggerCondition{Pattern: game.TriggerPattern{
					Event:         game.EventSpellCast,
					CardSelection: game.Selection{RequiredTypesAny: []types.Card{types.Instant, types.Sorcery}},
				}},
				Content: game.Mode{Text: "Other players copy the spell with optional new targets", Program: hiveMindCopyProgram}.Ability(),
			}},
		},
	}
}

func resolveHiveMindCopy(ctx cardprogram.Context) error {
	caster := ctx.View().Source().EventPlayer
	if !caster.Exists {
		return nil
	}
	players, err := ctx.View().Players(game.AllPlayersReference())
	if err != nil {
		return err
	}
	for _, player := range players {
		if player == caster.Val {
			continue
		}
		if _, err := ctx.Copy(game.CopyStackObject{
			Object: game.EventStackObjectReference(), MayChooseNewTargets: true,
		}, player); err != nil {
			return err
		}
	}
	return nil
}
