package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/cardprogram"
	"github.com/natefinch/council4/mtg/cards/authored"
	"github.com/natefinch/council4/mtg/cards/cardset"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

func existingParagraphRemovalEngine(t *testing.T, entry cardset.Entry) *Engine {
	t.Helper()
	catalog, err := cardprogram.NewCatalog(entry.Programs)
	if err != nil {
		t.Fatal(err)
	}
	return NewEngineWithPrograms(nil, catalog)
}

func existingParagraphCreature(name string, manaValue int) *game.CardDef {
	def := vanillaCreature(name, 2, 10)
	def.ManaCost = opt.Val(cost.Mana{cost.O(manaValue)})
	return def
}

// Source review: pinned oracle-cards.json SHA256
// 6da9ced0e19bd384771ab8f14051cc45a8dc90302082b875fa6e039d2399a1b4;
// Forced March Oracle ID 457dc141-bc72-42ad-b482-c78183462e34, input revision
// 36eae0e1-7100-449d-a259-7abfcd429117. Scryfall rulings endpoint
// https://api.scryfall.com/cards/36eae0e1-7100-449d-a259-7abfcd429117/rulings
// retrieved 2026-11-11 returned no rulings (response retained in
// .cardwork/forced-march-rulings.json).
// Read the official Comprehensive Rules effective 2026-09-25 from the verified
// local copy .cardwork/MagicCompRules-2026-09-25.txt (SHA256
// 8d860e451f20f38865b725b42d82feb714c725373dd8f3b32b8652b3eeb070ca):
// 107.3a for announced X, 202.3 for mana value, 608.2c for resolving the written
// instruction, and 701.8 for destruction.
func TestForcedMarchScenario(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := existingParagraphRemovalEngine(t, authored.ForcedMarch)
	setMainPhasePriority(g, game.Player1)

	one := addCombatPermanent(g, game.Player1, existingParagraphCreature("One mana creature", 1))
	two := addCombatPermanent(g, game.Player2, existingParagraphCreature("Two mana creature", 2))
	three := addCombatPermanent(g, game.Player3, existingParagraphCreature("Three mana creature", 3))
	artifactDef := &game.CardDef{CardFace: game.CardFace{
		Name: "Two mana artifact", ManaCost: opt.Val(cost.Mana{cost.O(2)}),
		Types: []types.Card{types.Artifact},
	}}
	artifact := addCombatPermanent(g, game.Player4, artifactDef)

	spell := addCardToHand(g, game.Player1, authored.ForcedMarch.New())
	g.Players[game.Player1].ManaPool.Add(mana.C, 2)
	g.Players[game.Player1].ManaPool.Add(mana.B, 3)
	if !engine.applyAction(g, game.Player1, action.CastSpell(spell, nil, 2, nil)) {
		t.Fatal("Forced March could not be cast for X=2 and {B}{B}{B}")
	}
	engine.resolveTopOfStack(g, &TurnLog{})

	for _, permanent := range []*game.Permanent{one, two} {
		if permanentByCardID(g, permanent.CardInstanceID) != nil {
			t.Fatal("Forced March left a creature with mana value X or less on the battlefield")
		}
	}
	if permanentByCardID(g, three.CardInstanceID) != three {
		t.Fatal("Forced March destroyed a creature whose mana value exceeded X")
	}
	if permanentByCardID(g, artifact.CardInstanceID) != artifact {
		t.Fatal("Forced March destroyed a noncreature permanent")
	}
}

type incineratingBlastChoicePolicy struct {
	simPassPolicy
	discard bool
}

func (p incineratingBlastChoicePolicy) ChooseChoice(_ PlayerObservation, request game.ChoiceRequest) []int {
	if request.Kind == game.ChoiceMay {
		if p.discard {
			return []int{1}
		}
		return []int{0}
	}
	return request.DefaultSelection
}

// Source review: pinned oracle-cards.json SHA256
// 6da9ced0e19bd384771ab8f14051cc45a8dc90302082b875fa6e039d2399a1b4;
// Incinerating Blast Oracle ID 49f0da4a-e046-4220-b90e-e105a5f2ef00, input
// revision d58e20ab-c5ca-4295-884d-78efdaa83243. The Scryfall rulings endpoint
// https://api.scryfall.com/cards/d58e20ab-c5ca-4295-884d-78efdaa83243/rulings
// retrieved 2026-11-11 returned WotC's 2024-11-08 ruling that an illegal target
// prevents the spell and all effects, including its discard/draw rider
// (response retained in .cardwork/incinerating-blast-rulings.json). Read official
// CR 2026-09-25 from the verified local copy (SHA256
// 8d860e451f20f38865b725b42d82feb714c725373dd8f3b32b8652b3eeb070ca): 601.2c,
// 608.2b-c, 120.4b-d, 701.9, and 121.2.
func TestIncineratingBlastScenario(t *testing.T) {
	for _, chooseDiscard := range []bool{false, true} {
		name := "declines discard"
		if chooseDiscard {
			name = "discards then draws"
		}
		t.Run(name, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			engine := existingParagraphRemovalEngine(t, authored.IncineratingBlast)
			setMainPhasePriority(g, game.Player1)

			target := addCombatPermanent(g, game.Player2, vanillaCreature("Blast target", 2, 10))
			discarded := addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Discard candidate"}})
			drawn := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Draw candidate"}})
			spell := addCardToHand(g, game.Player1, authored.IncineratingBlast.New())
			g.Players[game.Player1].ManaPool.Add(mana.C, 4)
			g.Players[game.Player1].ManaPool.Add(mana.R, 1)

			if !engine.applyAction(g, game.Player1, action.CastSpell(
				spell, []game.Target{game.PermanentTarget(target.ObjectID)}, 0, nil,
			)) {
				t.Fatal("Incinerating Blast could not be cast targeting a creature for {4}{R}")
			}
			agents := [game.NumPlayers]PlayerAgent{}
			agents[game.Player1] = incineratingBlastChoicePolicy{discard: chooseDiscard}
			engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

			if target.MarkedDamage != 6 {
				t.Fatalf("target marked damage = %d, want 6", target.MarkedDamage)
			}
			if chooseDiscard {
				if !g.Players[game.Player1].Graveyard.Contains(discarded) ||
					!g.Players[game.Player1].Hand.Contains(drawn) {
					t.Fatal("choosing to discard did not discard one card and draw one card")
				}
			} else if !g.Players[game.Player1].Hand.Contains(discarded) ||
				!g.Players[game.Player1].Library.Contains(drawn) {
				t.Fatal("declining the discard changed the hand or drew a card")
			}
		})
	}
}
