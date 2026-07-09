package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// evelynCovetousPermanent gives playerID a battlefield permanent carrying
// Evelyn, the Covetous' static play/cast-from-exile permission: a paired
// RuleEffectPlayLandsFromZone and RuleEffectCastSpellsFromZone over the exile
// zone, each filtered by a collection counter, restricted to cards the
// controller's own ability exiled (ExileCounterExiledByController), sharing a
// single OncePerTurn use, with the cast grant allowing any-color mana.
func evelynCovetousPermanent(g *game.Game, playerID game.PlayerID) *game.Permanent {
	return addCombatPermanent(g, playerID, &game.CardDef{CardFace: game.CardFace{
		Name: "Test Evelyn",
		StaticAbilities: []game.StaticAbility{{
			RuleEffects: []game.RuleEffect{
				{
					Kind:                           game.RuleEffectPlayLandsFromZone,
					AffectedPlayer:                 game.PlayerYou,
					CastFromZone:                   zone.Exile,
					PermanentTypes:                 []types.Card{types.Land},
					ExileCounterFilter:             opt.Val(counter.Collection),
					ExileCounterExiledByController: true,
					OncePerTurn:                    true,
				},
				{
					Kind:                           game.RuleEffectCastSpellsFromZone,
					AffectedPlayer:                 game.PlayerYou,
					CastFromZone:                   zone.Exile,
					ExileCounterFilter:             opt.Val(counter.Collection),
					ExileCounterExiledByController: true,
					OncePerTurn:                    true,
					SpendAnyMana:                   true,
				},
			},
		}},
	}})
}

// TestExilePlayPermissionProvenanceFilter verifies the
// ExileCounterExiledByController rider limits the permission to cards recorded as
// exiled by an ability the permission's controller controlled: cards exiled by
// another player's ability, or with no recorded exiler, stay out of reach even
// though they carry the same collection counter.
func TestExilePlayPermissionProvenanceFilter(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	evelynCovetousPermanent(g, game.Player1)

	mine := addCardToExile(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name: "Forest", Types: []types.Card{types.Land}}})
	theirs := addCardToExile(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name: "Island", Types: []types.Card{types.Land}}})
	noProvenance := addCardToExile(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name: "Mountain", Types: []types.Card{types.Land}}})
	g.AddExileCounterFromController(mine, counter.Collection, 1, game.Player1)
	g.AddExileCounterFromController(theirs, counter.Collection, 1, game.Player2)
	g.AddExileCounter(noProvenance, counter.Collection, 1)

	if !canPlayLandFromZoneByRuleEffect(g, game.Player1, mine, zone.Exile) {
		t.Fatal("card exiled by the controller's own ability is not playable despite the permission")
	}
	if canPlayLandFromZoneByRuleEffect(g, game.Player1, theirs, zone.Exile) {
		t.Fatal("card exiled by another player's ability is playable through a provenance-filtered permission")
	}
	if canPlayLandFromZoneByRuleEffect(g, game.Player1, noProvenance, zone.Exile) {
		t.Fatal("card with no recorded exiler is playable through a provenance-filtered permission")
	}
}

// TestExilePlayPermissionOncePerTurnSharedAcrossPlayAndCast verifies the
// OncePerTurn rider caps the paired land-play and spell-cast permissions at a
// single shared use per turn: playing a land from exile spends the use, blocking
// a later cast from exile until the per-turn state resets on the next turn.
func TestExilePlayPermissionOncePerTurnSharedAcrossPlayAndCast(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	evelynCovetousPermanent(g, game.Player1)
	landID := addCardToExile(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name: "Forest", Types: []types.Card{types.Land}}})
	spellID := addCardToExile(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name: "Collected Bolt", Types: []types.Card{types.Instant}}})
	g.AddExileCounterFromController(landID, counter.Collection, 1, game.Player1)
	g.AddExileCounterFromController(spellID, counter.Collection, 1, game.Player1)
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone

	if !canPlayLandFromZoneByRuleEffect(g, game.Player1, landID, zone.Exile) {
		t.Fatal("land is not playable before the once-per-turn use is spent")
	}
	if !canCastSpellsFromZoneByRuleEffect(g, game.Player1, spellID, zone.Exile, game.FaceFront) {
		t.Fatal("spell is not castable before the once-per-turn use is spent")
	}

	if !engine.applyAction(g, game.Player1, action.PlayLandFaceFromZone(landID, zone.Exile, game.FaceFront)) {
		t.Fatal("playing the collection-countered land from exile was rejected")
	}

	if canCastSpellsFromZoneByRuleEffect(g, game.Player1, spellID, zone.Exile, game.FaceFront) {
		t.Fatal("spell remained castable after the shared once-per-turn use was spent by the land play")
	}

	engine.advanceToNextTurn(g)
	if !canCastSpellsFromZoneByRuleEffect(g, game.Player1, spellID, zone.Exile, game.FaceFront) {
		t.Fatal("shared once-per-turn use did not reset on the next turn")
	}
}

// TestExilePlayPermissionAnyColorMana verifies the SpendAnyMana rider on the
// cast-from-exile permission lets the controller spend mana of any color to cast
// the permitted card, while a permission without the rider does not.
func TestExilePlayPermissionAnyColorMana(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	evelynCovetousPermanent(g, game.Player1)
	spellID := addCardToExile(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name: "Collected Bolt", Types: []types.Card{types.Instant}}})
	g.AddExileCounterFromController(spellID, counter.Collection, 1, game.Player1)

	if !castFromZoneAllowsAnyMana(g, game.Player1, spellID, zone.Exile, game.FaceFront) {
		t.Fatal("any-color mana permission not honored for the collection-countered exiled spell")
	}

	other := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	playAndCastFromExileWithCroakPermanent(other, game.Player1)
	croakSpell := addCardToExile(other, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name: "Croaked Bolt", Types: []types.Card{types.Instant}}})
	other.AddExileCounter(croakSpell, counter.Croak, 1)
	if castFromZoneAllowsAnyMana(other, game.Player1, croakSpell, zone.Exile, game.FaceFront) {
		t.Fatal("a permission without the any-color rider wrongly allows any-color mana")
	}
}
