package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/cardprogram"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

const bringDownAssistOracleID = "05d2b89b-acc4-4289-95a3-a09ebb23a5d6"

var bringDownAssistProbeProgram = game.ProgramReference{
	ID: bringDownAssistOracleID + "/spell/0", Version: "1",
}

func resolveBringDownAssistProbe(ctx cardprogram.Context) error {
	_, err := ctx.Destroy(game.Destroy{Object: game.TargetPermanentReference(0)})
	return err
}

// Source review: pinned corpus SHA256
// 6da9ced0e19bd384771ab8f14051cc45a8dc90302082b875fa6e039d2399a1b4;
// Bring Down Oracle ID 05d2b89b-acc4-4289-95a3-a09ebb23a5d6, printing
// e18146f9-369c-41c8-8a1d-7737edd2c18e, text "Assist (Another player can pay
// up to {3} of this spell's cost.) Destroy target creature with power 4 or
// greater." Scryfall rulings endpoint
// https://api.scryfall.com/cards/e18146f9-369c-41c8-8a1d-7737edd2c18e/rulings
// returned four WotC rulings dated 2018-06-08: a card may be revealed and
// payment discussed before casting begins, but targets are chosen before the
// assisting player; only generic mana may be assisted; color-changing effects
// do not turn colored costs generic; and cost changes alter the generic amount
// Assist can cover.
// Read the official Wizards CR at
// https://media.wizards.com/2026/downloads/MagicCompRules%2020261106.txt,
// effective November 6, 2026: 208.1 and 613.1g/613.4c require the target's
// current power (including modifiers) to meet 4; 115.1a and 601.2c require a
// legal target on casting; 601.2f–i, 605.3a and 702.132a place Assist after
// cost lock: the helper activates mana abilities, the caster activates theirs,
// and the helper may then pay generic mana before the caster pays the rest; 608.2b
// rechecks target legality; 701.8a-c define destruction and its replacement.
// CR 734.1–2 governs unsuccessful casts; 707.10/.12 distinguish stack
// copies from cast card copies. This unregistered fixture proves #192.
func TestBringDownAssistTextWithSoloPayment(t *testing.T) {
	g, engine, target, spell := bringDownAssistProbeFixture(t)
	g.Players[game.Player1].ManaPool.Add(mana.W, 1)
	g.Players[game.Player1].ManaPool.Add(mana.C, 3)
	setMainPhasePriority(g, game.Player1)

	if !engine.applyAction(g, game.Player1, action.CastSpell(
		spell, []game.Target{game.PermanentTarget(target.ObjectID)}, 0, nil,
	)) {
		t.Fatal("Bring Down rejected its legal target or fully paid cost")
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	if !g.Players[game.Player2].Graveyard.Contains(target.CardInstanceID) ||
		!g.Players[game.Player1].Graveyard.Contains(spell) {
		t.Fatal("Bring Down did not destroy the target and finish in the caster's graveyard")
	}
}

func TestBringDownAssist(t *testing.T) {
	g, engine, target, spell := bringDownAssistProbeFixture(t)
	g.Players[game.Player1].ManaPool.Add(mana.W, 1)
	g.Players[game.Player2].ManaPool.Add(mana.C, 3)
	setMainPhasePriority(g, game.Player1)

	if !engine.applyAction(g, game.Player1, action.CastSpell(
		spell, []game.Target{game.PermanentTarget(target.ObjectID)}, 0, nil,
	)) {
		t.Fatal("Assist should let Player 2 pay the three generic mana after Player 1 chooses the legal 4-power target and pays {W}")
	}
	if g.Players[game.Player2].ManaPool.Total() != 0 || g.Players[game.Player1].ManaPool.Total() != 0 {
		t.Fatal("wrong player paid the generic or white cost")
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	if !g.Players[game.Player2].Graveyard.Contains(target.CardInstanceID) {
		t.Fatal("assisted Bring Down did not destroy its legal target")
	}
}

func bringDownAssistProbeFixture(t *testing.T) (*game.Game, *Engine, *game.Permanent, id.ID) {
	t.Helper()
	catalog, err := cardprogram.NewCatalog([]cardprogram.Program{{
		Reference: bringDownAssistProbeProgram,
		Resolve:   resolveBringDownAssistProbe,
	}})
	if err != nil {
		t.Fatal(err)
	}
	engine := NewEngineWithPrograms(nil, catalog)
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	target := addCombatPermanent(g, game.Player2, vanillaCreature("Four-power target", 4, 4))
	def := &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name:            "Bring Down",
			OracleText:      "Assist (Another player can pay up to {3} of this spell's cost.) Destroy target creature with power 4 or greater.",
			ManaCost:        opt.Val(cost.Mana{cost.O(3), cost.W}),
			Types:           []types.Card{types.Sorcery},
			Colors:          []color.Color{color.White},
			StaticAbilities: []game.StaticAbility{game.AssistStaticBody},
			SpellAbility: opt.Val(game.Mode{
				Text:    "Destroy target creature with power 4 or greater.",
				Program: bringDownAssistProbeProgram,
				Targets: []game.TargetSpec{{
					MinTargets: 1, MaxTargets: 1, Allow: game.TargetAllowPermanent,
					Selection: opt.Val(game.Selection{
						RequiredTypes: []types.Card{types.Creature},
						Power:         opt.Val(compare.Int{Op: compare.GreaterOrEqual, Value: 4}),
					}),
				}},
			}.Ability()),
		},
	}
	spell := addCardToHand(g, game.Player1, def)
	return g, engine, target, spell
}
