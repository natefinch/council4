package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// selfReferentialGreatestPower is a creature whose static ability sets its own
// power to the greatest power among creatures its controller controls — a group
// that includes itself. Computing its effective power therefore requires knowing
// the greatest power among your creatures, which requires its own effective
// power: a characteristic-dependency loop (CR 613.8) of the shape that real
// Commander boards reach (a "* = greatest power you control" creature alongside
// other such creatures). Before the re-entry guard in effectivePermanentValues
// this recursed without bound and overflowed the stack.
func selfReferentialGreatestPower(g *game.Game, controller game.PlayerID) *game.Permanent {
	return addCombatPermanent(g, controller, &game.CardDef{CardFace: game.CardFace{
		Name:      "Mirror Colossus",
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(game.PT{Value: 3}),
		Toughness: opt.Val(game.PT{Value: 3}),
		StaticAbilities: []game.StaticAbility{{
			ContinuousEffects: []game.ContinuousEffect{{
				Layer:          game.LayerPowerToughnessSet,
				AffectedSource: true,
				SetPowerDynamic: opt.Val(game.DynamicAmount{
					Kind: game.DynamicAmountGreatestPowerInGroup,
					Group: game.BattlefieldGroup(game.Selection{
						RequiredTypes: []types.Card{types.Creature},
						Controller:    game.ControllerYou,
					}),
				}),
			}},
		}},
	}})
}

// TestEffectivePermanentValuesBreaksCharacteristicDependencyLoop confirms a
// characteristic-defining static ability whose value depends on the very
// characteristic it defines terminates instead of overflowing the stack. Real
// four-player games reached this and crashed the engine (and exhausted memory in
// the WebAssembly playtester); the fix is a re-entry guard that falls back to base
// values, breaking the loop (CR 613.8c: a dependency loop applies in timestamp
// order, without re-entering).
func TestEffectivePermanentValuesBreaksCharacteristicDependencyLoop(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	colossus := selfReferentialGreatestPower(g, game.Player1)

	// Alone, the greatest power among your creatures is the colossus's own base
	// power: the loop resolves to 3 rather than recursing forever.
	if got := effectivePermanentValues(g, colossus); !got.powerOK || got.power != 3 {
		t.Fatalf("self-referential greatest-power creature resolved to power %d (ok=%v), want 3",
			got.power, got.powerOK)
	}

	// A bigger independent creature raises the greatest power, and the colossus's
	// power follows it — evaluating that still terminates.
	addCombatCreaturePermanentWithPower(g, game.Player1, 7)
	if got := effectivePermanentValues(g, colossus); !got.powerOK || got.power != 7 {
		t.Fatalf("colossus power with a 7-power creature in play = %d (ok=%v), want 7",
			got.power, got.powerOK)
	}
}
