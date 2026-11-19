//go:build card_support_gap_probe

package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/cardprogram"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

var sickeningDreamsGapProbeProgram = game.ProgramReference{
	ID: "728d2e30-2f36-4054-81db-29b01e376a82/spell/0", Version: "1",
}

func resolveSickeningDreamsGapProbe(ctx cardprogram.Context) error {
	_, err := ctx.Execute(game.Damage{
		Amount: game.Fixed(1),
		Recipient: game.GroupDamageRecipient(game.BattlefieldGroup(
			game.Selection{RequiredTypes: []types.Card{types.Creature}},
		)),
	})
	return err
}

// Source review: pinned oracle-cards.json SHA256
// 6da9ced0e19bd384771ab8f14051cc45a8dc90302082b875fa6e039d2399a1b4;
// Sickening Dreams Oracle ID 728d2e30-2f36-4054-81db-29b01e376a82, input
// revision 9396ac77-9f53-46d6-83d7-a60c15d2035d. The Scryfall rulings endpoint
// https://api.scryfall.com/cards/9396ac77-9f53-46d6-83d7-a60c15d2035d/rulings
// retrieved 2026-11-11 returned no rulings (response retained in
// .cardwork/sickening-dreams-rulings.json).
// Read official CR 2026-09-25 from the verified local copy (SHA256
// 8d860e451f20f38865b725b42d82feb714c725373dd8f3b32b8652b3eeb070ca):
// 107.3a and 601.2h for X and its discard additional cost, 608.2f for processing
// actions on multiple objects and players simultaneously, 120.4b-d for damage,
// and 701.9b for discarding to pay a cost. Recipient is an exclusive union of
// permanent-group or player-group damage; one operation cannot affect both.
// Canonical blocker #3762.
func TestSickeningDreamsMixedDamageGapProbe(t *testing.T) {
	catalog, err := cardprogram.NewCatalog([]cardprogram.Program{{
		Reference: sickeningDreamsGapProbeProgram,
		Resolve:   resolveSickeningDreamsGapProbe,
	}})
	if err != nil {
		t.Fatal(err)
	}
	engine := NewEngineWithPrograms(nil, catalog)
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	creature := addCombatPermanent(g, game.Player2, vanillaCreature("Damage recipient", 5, 5))
	def := &game.CardDef{CardFace: game.CardFace{
		Name:     "Sickening Dreams gap probe",
		ManaCost: opt.Val(cost.Mana{cost.B}),
		Types:    []types.Card{types.Sorcery},
		SpellAbility: opt.Val(game.Mode{
			Program: sickeningDreamsGapProbeProgram,
		}.Ability()),
	}}
	spell := addCardToHand(g, game.Player1, def)
	g.Players[game.Player1].ManaPool.Add(mana.B, 1)
	setMainPhasePriority(g, game.Player1)
	if !engine.applyAction(g, game.Player1, action.CastSpell(spell, nil, 0, nil)) {
		t.Fatal("Sickening Dreams mixed-damage probe could not be cast")
	}
	engine.resolveTopOfStack(g, &TurnLog{})

	if creature.MarkedDamage != 1 || g.Players[game.Player1].Life != 39 {
		t.Fatalf("one mixed-damage operation dealt %d to the creature and left player at %d life; want 1 to each",
			creature.MarkedDamage, g.Players[game.Player1].Life)
	}
}
