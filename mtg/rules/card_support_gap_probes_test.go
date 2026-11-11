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

// Source review: pinned oracle-cards.json SHA256
// 6da9ced0e19bd384771ab8f14051cc45a8dc90302082b875fa6e039d2399a1b4;
// Killing Glare Oracle ID 7b47891f-130a-4d76-b262-53f5e3bb6b95, input revision
// f7a4d87d-b844-4f20-8b14-4fd32c53dea5. The Scryfall rulings endpoint
// https://api.scryfall.com/cards/f7a4d87d-b844-4f20-8b14-4fd32c53dea5/rulings
// retrieved 2026-11-11 returned WotC's 2013-01-24 ruling: power is checked both
// when targeted and when the spell resolves, and a target above X makes the
// spell fail to resolve (response retained in
// .cardwork/killing-glare-rulings.json).
// Read official CR 2026-09-25 from .cardwork/MagicCompRules-2026-09-25.txt,
// verified SHA256 8d860e451f20f38865b725b42d82feb714c725373dd8f3b32b8652b3eeb070ca:
// 107.3a, 601.2c, and 608.2b. Current TargetSpec has fixed Selection.Power and
// X-derived mana-value bounds, but no X-derived power bound; canonical blocker #3763.
func TestKillingGlareXBoundPowerGapProbe(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	target := addCombatPermanent(g, game.Player2, vanillaCreature("Power three", 3, 5))
	targetSpec := game.TargetSpec{
		MinTargets: 1,
		MaxTargets: 1,
		Allow:      game.TargetAllowPermanent,
		Selection:  opt.Val(game.Selection{RequiredTypes: []types.Card{types.Creature}}),
	}
	def := &game.CardDef{CardFace: game.CardFace{
		Name:     "Killing Glare gap probe",
		ManaCost: opt.Val(cost.Mana{cost.X, cost.B}),
		Types:    []types.Card{types.Instant},
		SpellAbility: opt.Val(game.Mode{
			Targets: []game.TargetSpec{targetSpec},
		}.Ability()),
	}}
	branch := game.CastBranch{XValue: opt.Val(2)}
	if spellTargetsValidForCast(
		g, game.Player1, def, nil, []game.Target{game.PermanentTarget(target.ObjectID)}, 2, 0, branch,
	) {
		t.Fatal("Killing Glare accepted a creature with power greater than announced X")
	}
}

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
