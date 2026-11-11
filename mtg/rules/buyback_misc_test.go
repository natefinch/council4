package rules

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/cards"
	"github.com/natefinch/council4/mtg/cards/authored"
	"github.com/natefinch/council4/mtg/cards/cardset"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
)

// Source review: pinned corpus SHA-256 6da9ced0e19bd384771ab8f14051cc45a8dc90302082b875fa6e039d2399a1b4;
// Lab Rats Oracle ID c53d6d3e-64e3-48b9-afab-3bde3164888f, printing revision
// 6fd5d79e-cfd8-484b-bbc8-785177922e30 (TPR 108, 2015-05-06): "Buyback {4}
// (You may pay an additional {4} as you cast this spell. If you do, put this
// card into your hand as it resolves.) Create a 1/1 black Rat creature token."
// Its rulings endpoint https://api.scryfall.com/cards/6fd5d79e-cfd8-484b-bbc8-785177922e30/rulings
// returned 0 records on 2026-11-11. Read CR 2026-09-25 (the corpus rules pin)
// from the official PDF https://media.wizards.com/2026/downloads/MagicCompRules%2020260925.pdf
// (SHA-256 674241213075cd581dd3255567ebeec7b40ffae3a27c6a3bd99218af6b2b6659)
// and the current official TXT https://media.wizards.com/2026/downloads/MagicCompRules%2020261106.txt:
// 111.2-4, 118.8a-b, 601.2b/f-h, 608.2n, and 702.27a require an optional
// additional cast cost, token ownership/characteristics, and hand return only
// after resolution.
func TestLabRatsScenario(t *testing.T) {
	registry, engine := buybackInteractionRuntime(t)
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	for _, manaColor := range []mana.Color{mana.B, mana.C, mana.C, mana.C, mana.C} {
		g.Players[game.Player1].ManaPool.Add(manaColor, 1)
	}

	cardID, spell := castBuybackMisc(t, g, registry, engine, authored.LabRats)
	if spell.ManaSpentToCast != 5 || g.Players[game.Player1].ManaPool.Total() != 0 {
		t.Fatalf("Lab Rats cast spent %d mana with %d remaining; want all five mana paid", spell.ManaSpentToCast, g.Players[game.Player1].ManaPool.Total())
	}

	engine.resolveTopOfStack(g, &TurnLog{})

	assertBuybackRepeaterReturned(t, g, cardID)
	assertBuybackMiscToken(t, g, "Rat Token", color.Black, types.Rat)
}

// Source review: pinned corpus SHA-256 6da9ced0e19bd384771ab8f14051cc45a8dc90302082b875fa6e039d2399a1b4;
// Mystic Speculation Oracle ID a58adf82-d9e2-4b47-933e-68180f280ba2, printing
// revision 92adfcb4-6a0a-46f9-a171-bfa9fade2156 (FUT 41, 2007-05-04):
// "Buyback {2} (You may pay an additional {2} as you cast this spell. If you
// do, put this card into your hand as it resolves.) Scry 3. (Look at the top
// three cards of your library, then put any number of them on the bottom and
// the rest on top in any order.)" Its rulings endpoint
// https://api.scryfall.com/cards/92adfcb4-6a0a-46f9-a171-bfa9fade2156/rulings
// returned 0 records on 2026-11-11. Read CR 2026-09-25 (the corpus rules pin)
// from the official PDF https://media.wizards.com/2026/downloads/MagicCompRules%2020260925.pdf
// and current official TXT https://media.wizards.com/2026/downloads/MagicCompRules%2020261106.txt:
// 118.8a-b, 401.2, 601.2b/f-h, 608.2n, 701.22a, and 702.27a govern the
// optional buyback cost, scry's hidden-library look/top-bottom ordering, and
// resolution return.
func TestMysticSpeculationScenario(t *testing.T) {
	registry, engine := buybackInteractionRuntime(t)
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	for _, manaColor := range []mana.Color{mana.U, mana.C, mana.C} {
		g.Players[game.Player1].ManaPool.Add(manaColor, 1)
	}

	addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Mystic top one"}})
	addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Mystic top two"}})
	addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Mystic top three"}})
	addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Mystic untouched card"}})
	initialLibrary := g.Players[game.Player1].Library.All()
	if len(initialLibrary) != 4 {
		t.Fatal("Mystic Speculation library setup failed")
	}
	cardID, spell := castBuybackMisc(t, g, registry, engine, authored.MysticSpeculation)
	if spell.ManaSpentToCast != 3 || g.Players[game.Player1].ManaPool.Total() != 0 {
		t.Fatalf("Mystic Speculation cast spent %d mana with %d remaining; want all three mana paid", spell.ManaSpentToCast, g.Players[game.Player1].ManaPool.Total())
	}

	var scryCards []id.ID
	var topOrder []id.ID
	agent := &authoredCopyChoiceAgent{choose: func(request game.ChoiceRequest) []int {
		switch request.Kind {
		case game.ChoiceScry:
			if request.Player != game.Player1 || len(request.Options) != 3 {
				t.Fatalf("scry request = %+v, want controller to choose among three cards", request)
			}
			for _, option := range request.Options {
				scryCards = append(scryCards, option.Card.Val.CardID)
			}
			return []int{1}
		case game.ChoiceOrder:
			if request.Player != game.Player1 || len(request.Options) != 2 {
				t.Fatalf("order request = %+v, want controller to order the two cards kept on top", request)
			}
			for _, option := range request.Options {
				topOrder = append(topOrder, option.Card.Val.CardID)
			}
			return []int{1, 0}
		default:
			t.Fatalf("unexpected Mystic Speculation choice: %+v", request)
			return request.DefaultSelection
		}
	}}
	engine.resolveTopOfStackWithChoices(
		g,
		[game.NumPlayers]PlayerAgent{game.Player1: agent},
		&TurnLog{},
	)

	if !slices.Equal(scryCards, initialLibrary[:3]) {
		t.Fatalf("scry options = %v, want the actual top three %v", scryCards, initialLibrary[:3])
	}
	if !slices.Equal(topOrder, []id.ID{scryCards[0], scryCards[2]}) {
		t.Fatalf("top-order options = %v, want the two retained cards %v", topOrder, []id.ID{scryCards[0], scryCards[2]})
	}
	wantLibrary := []id.ID{scryCards[2], scryCards[0], initialLibrary[3], scryCards[1]}
	if got := g.Players[game.Player1].Library.All(); !slices.Equal(got, wantLibrary) {
		t.Fatalf("library after scry = %v, want reordered top and bottom %v", got, wantLibrary)
	}
	assertEvent(t, g.Events, game.EventScry, func(event game.Event) bool {
		return event.Player == game.Player1 && event.Amount == 3
	})
	assertBuybackRepeaterReturned(t, g, cardID)
}

// Source review: pinned corpus SHA-256 6da9ced0e19bd384771ab8f14051cc45a8dc90302082b875fa6e039d2399a1b4;
// Sprout Swarm Oracle ID 8b33d890-7d67-44a8-a253-e2b171d7ca9d, printing revision
// 0b915355-4e98-44df-81bd-961a3d3c86b8 (FUT 138, 2007-05-04): "Convoke (Your
// creatures can help cast this spell. Each creature you tap while casting this
// spell pays for {1} or one mana of that creature's color.) Buyback {3} (You
// may pay an additional {3} as you cast this spell. If you do, put this card
// into your hand as it resolves.) Create a 1/1 green Saproling creature token."
// Its rulings endpoint https://api.scryfall.com/cards/0b915355-4e98-44df-81bd-961a3d3c86b8/rulings
// returned six WotC rulings, all dated 2024-01-12; the relevant ruling confirms
// convoke is applied after total cost, can combine with additional costs, and
// a multicolored creature pays one mana of a color of its controller's choice.
// Read CR 2026-09-25 (the corpus rules pin) from the official PDF
// https://media.wizards.com/2026/downloads/MagicCompRules%2020260925.pdf and
// current official TXT https://media.wizards.com/2026/downloads/MagicCompRules%2020261106.txt:
// 111.2-4, 118.8a-b, 601.2b/f-h, 608.2n, 702.27a, and 702.51a-c govern the
// optional additional cost, convoke payment from total cost, token, and return.
func TestSproutSwarmScenario(t *testing.T) {
	registry, engine := buybackInteractionRuntime(t)
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	creatures := []*game.Permanent{
		addCombatPermanent(g, game.Player1, greenConvokeCreature()),
	}
	for range 4 {
		creatures = append(creatures, addCombatCreaturePermanent(g, game.Player1))
	}

	cardID, spell := castBuybackMisc(t, g, registry, engine, authored.SproutSwarm)
	if !spell.BuybackPaid || g.Players[game.Player1].ManaPool.Total() != 0 {
		t.Fatal("Sprout Swarm did not pay buyback entirely through the available convoke creatures")
	}
	for _, creature := range creatures {
		if !creature.Tapped {
			t.Fatalf("Sprout Swarm did not convoke %s to pay its total cost", permanentEffectiveName(g, creature))
		}
	}

	engine.resolveTopOfStack(g, &TurnLog{})

	assertBuybackRepeaterReturned(t, g, cardID)
	assertBuybackMiscToken(t, g, "Saproling Token", color.Green, types.Saproling)
}

func castBuybackMisc(
	t *testing.T,
	g *game.Game,
	registry *cards.Registry,
	engine *Engine,
	entry cardset.Entry,
) (id.ID, *game.StackObject) {
	t.Helper()
	def, err := registry.LookupOracleID(entry.OracleID)
	if err != nil || def == nil || def.Name != entry.Name {
		t.Fatalf("exact production lookup for %s failed: %v", entry.Name, err)
	}
	cardID := addCardToHand(g, game.Player1, def)
	setMainPhasePriority(g, game.Player1)
	if !engine.applyActionWithChoices(
		g,
		game.Player1,
		action.CastSpell(cardID, nil, 0, nil),
		buybackInteractionChoiceAgent(t, game.Player1, true),
		&TurnLog{},
	) {
		t.Fatalf("%s could not be cast with its printed cost and buyback", entry.Name)
	}
	spell, ok := g.Stack.Peek()
	if !ok || spell.Kind != game.StackSpell || spell.SourceID != cardID || !spell.BuybackPaid {
		t.Fatalf("%s did not record its paid buyback cast: %+v", entry.Name, spell)
	}
	return cardID, spell
}

func assertBuybackMiscToken(t *testing.T, g *game.Game, name string, tokenColor color.Color, subtype types.Sub) {
	t.Helper()
	count := 0
	for _, permanent := range g.Battlefield {
		if !permanent.Token {
			continue
		}
		count++
		def := permanent.TokenDef
		if def.Name != name || permanent.Controller != game.Player1 || permanent.Owner != game.Player1 ||
			permanent.Tapped || !slices.Equal(def.Types, []types.Card{types.Creature}) ||
			!slices.Equal(def.Subtypes, []types.Sub{subtype}) || !slices.Equal(def.Colors, []color.Color{tokenColor}) ||
			!def.Power.Exists || def.Power.Val.Value != 1 || !def.Toughness.Exists || def.Toughness.Val.Value != 1 ||
			def.ManaCost.Exists || len(def.StaticAbilities) != 0 || len(def.TriggeredAbilities) != 0 ||
			len(def.ActivatedAbilities) != 0 {
			t.Fatalf("wrong %s token characteristics or controller: %+v / %+v", name, permanent, def)
		}
	}
	if count != 1 {
		t.Fatalf("%s token count = %d, want 1", name, count)
	}
}
