package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/cards"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

// Source review: pinned Oracle corpus SHA-256
// 6da9ced0e19bd384771ab8f14051cc45a8dc90302082b875fa6e039d2399a1b4,
// Squeaking Pie Sneak Oracle ID 7243b935-73fc-47d2-ac65-ab390ed94f3f,
// printing b7184715-3ed3-4f56-9bf3-ff431cca86be: {1}{B}, 2/2 Goblin Rogue,
// "As an additional cost to cast this spell, reveal a Goblin card from your
// hand or pay {3}. Fear (This creature can't be blocked except by artifact
// creatures and/or black creatures.)" Its Scryfall rulings URI
// https://api.scryfall.com/cards/b7184715-3ed3-4f56-9bf3-ff431cca86be/rulings
// returned an empty list on 2026-11-14. Official WotC Comprehensive Rules TXT,
// effective 2026-11-06, https://media.wizards.com/2026/downloads/MagicCompRules%2020261106.txt:
// 108.2, 118.8-8d, 205.3b/f/m, 402.3, 601.2b/f-h and 701.20a-c require a
// Goblin card in hand to be revealed as the chosen cost branch, kept in its
// zone, and shown through the spell's time on the stack; alternatively {3} is
// added to {1}{B}. Rules 702.36a-c govern the creature's Fear evasion.
func TestSqueakingPieSneakPaysThreeAndHasFear(t *testing.T) {
	g, _, permanent := castRevealOrPayCreature(t, nil, "7243b935-73fc-47d2-ac65-ab390ed94f3f",
		"Squeaking Pie Sneak", types.Goblin, "Pay {3}", mana.B, 1, 3, false)
	if !hasKeyword(g, permanent, game.Fear) {
		t.Fatal("Squeaking Pie Sneak did not have Fear after resolving")
	}
}

// Source review: pinned Oracle corpus SHA-256
// 6da9ced0e19bd384771ab8f14051cc45a8dc90302082b875fa6e039d2399a1b4,
// Sadistic Skymarcher Oracle ID 9e1fe5ec-1669-4587-8e42-b189fd6d39e0,
// printing 1334ee90-c798-42ec-b2be-68dac766a7eb: {2}{B}, 2/2 Vampire Soldier,
// "As an additional cost to cast this spell, reveal a Vampire card from your
// hand or pay {1}. Flying, lifelink." Its Scryfall rulings URI
// https://api.scryfall.com/cards/1334ee90-c798-42ec-b2be-68dac766a7eb/rulings
// returned one WotC ruling dated 2018-01-19 (retrieved 2026-11-14): “If an
// effect refers to a “[subtype] spell” or “[subtype] card,” it refers only to
// a spell or card that has that subtype. For example, March of the Drowned is a
// card that benefits Pirates and features Pirates in its illustration, but it
// isn't a Pirate card.” Official WotC Comprehensive Rules TXT, effective
// 2026-11-06, https://media.wizards.com/2026/downloads/MagicCompRules%2020261106.txt:
// 108.2, 118.8-8d, 205.3b/f/m, 402.3, 601.2b/f-h and 701.20a-c govern the
// subtype-filtered reveal cost and additive {1}; 702.9a-b and 702.15a-b govern
// Flying and lifelink.
func TestSadisticSkymarcherPaysOneAndHasFlyingLifelink(t *testing.T) {
	g, _, permanent := castRevealOrPayCreature(t, nil, "9e1fe5ec-1669-4587-8e42-b189fd6d39e0",
		"Sadistic Skymarcher", types.Vampire, "Pay {1}", mana.B, 2, 1, false)
	if !hasKeyword(g, permanent, game.Flying) || !hasKeyword(g, permanent, game.Lifelink) {
		t.Fatal("Sadistic Skymarcher did not have flying and lifelink after resolving")
	}
}

// Source review: pinned Oracle corpus SHA-256
// 6da9ced0e19bd384771ab8f14051cc45a8dc90302082b875fa6e039d2399a1b4,
// Flamekin Bladewhirl Oracle ID c37770aa-4770-4198-b6a6-f9ff86dbb06b,
// printing 1b6ea1c8-ca58-4240-adb7-71cd49664414: {R}, 2/1 Elemental Warrior,
// "As an additional cost to cast this spell, reveal an Elemental card from your
// hand or pay {3}." Its Scryfall rulings URI
// https://api.scryfall.com/cards/1b6ea1c8-ca58-4240-adb7-71cd49664414/rulings
// returned an empty list on 2026-11-14. Official WotC Comprehensive Rules TXT,
// effective 2026-11-06, https://media.wizards.com/2026/downloads/MagicCompRules%2020261106.txt:
// 108.2, 118.8-8d, 205.3b/f/m, 402.3, 601.2b/f-h and 701.20a-c require an
// Elemental card in hand to be revealed without leaving that zone, or {3} to
// be added to the printed {R} cost.
func TestFlamekinBladewhirlRevealsElemental(t *testing.T) {
	castRevealOrPayCreature(t, nil, "c37770aa-4770-4198-b6a6-f9ff86dbb06b",
		"Flamekin Bladewhirl", types.Elemental, "Reveal an Elemental card from your hand",
		mana.R, 0, 3, true)
}

// Source review: pinned Oracle corpus SHA-256
// 6da9ced0e19bd384771ab8f14051cc45a8dc90302082b875fa6e039d2399a1b4,
// Silvergill Adept Oracle ID cb66c5ec-e7a4-4bd2-b825-6c9b846ba40d, printing
// 45ea8ca1-612d-4984-8572-06710d38bfa6: {1}{U}, 2/1 Merfolk Wizard,
// "As an additional cost to cast this spell, reveal a Merfolk card from your
// hand or pay {3}. When this creature enters, draw a card." Its Scryfall rulings
// URI https://api.scryfall.com/cards/45ea8ca1-612d-4984-8572-06710d38bfa6/rulings
// returned one WotC ruling dated 2018-01-19 (retrieved 2026-11-14): “If an
// effect refers to a “[subtype] spell” or “[subtype] card,” it refers only to
// a spell or card that has that subtype. For example, March of the Drowned is a
// card that benefits Pirates and features Pirates in its illustration, but it
// isn't a Pirate card.” Official WotC Comprehensive Rules TXT, effective
// 2026-11-06, https://media.wizards.com/2026/downloads/MagicCompRules%2020261106.txt:
// 108.2, 118.8-8d, 205.3b/f/m, 402.3, 601.2b/f-h and 701.20a-c govern the
// Merfolk reveal branch and additive {3}; 603.2/603.6a and 121.1-2 require the
// entry trigger to draw one card.
func TestSilvergillAdeptRevealsMerfolkAndDrawsOnEntry(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	drawnID := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Drawn card"}})
	_, engine, _ := castRevealOrPayCreature(t, g, "cb66c5ec-e7a4-4bd2-b825-6c9b846ba40d",
		"Silvergill Adept", types.Merfolk, "Reveal a Merfolk card from your hand", mana.U, 1, 3, true)
	if !engine.putTriggeredAbilitiesOnStackWithChoices(g, simPassPolicies(), &TurnLog{}) {
		t.Fatal("Silvergill Adept did not trigger when it entered")
	}
	engine.resolveTopOfStackWithChoices(g, simPassPolicies(), &TurnLog{})
	if !g.Players[game.Player1].Hand.Contains(drawnID) {
		t.Fatal("Silvergill Adept's entry trigger did not draw the top card")
	}
}

// Source review: pinned Oracle corpus SHA-256
// 6da9ced0e19bd384771ab8f14051cc45a8dc90302082b875fa6e039d2399a1b4,
// Wren's Run Vanquisher Oracle ID fe9cdd15-a390-4a1e-bae9-474ce8d355b8,
// printing f61eae36-9c4a-4d72-9431-6cac34dbb527: {1}{G}, 3/3 Elf Warrior,
// "As an additional cost to cast this spell, reveal an Elf card from your hand
// or pay {3}. Deathtouch (Any amount of damage this deals to a creature is
// enough to destroy it.)" Its Scryfall rulings URI
// https://api.scryfall.com/cards/f61eae36-9c4a-4d72-9431-6cac34dbb527/rulings
// returned an empty list on 2026-11-14. Official WotC Comprehensive Rules TXT,
// effective 2026-11-06, https://media.wizards.com/2026/downloads/MagicCompRules%2020261106.txt:
// 108.2, 118.8-8d, 205.3b/f/m, 402.3, 601.2b/f-h and 701.20a-c require the
// Elf card to be revealed as a cost and remain in hand; {3} is added to {1}{G}.
// Rules 702.2a-b govern Deathtouch.
func TestWrensRunVanquisherRevealsElfAndHasDeathtouch(t *testing.T) {
	g, _, permanent := castRevealOrPayCreature(t, nil, "fe9cdd15-a390-4a1e-bae9-474ce8d355b8",
		"Wren's Run Vanquisher", types.Elf, "Reveal an Elf card from your hand",
		mana.G, 1, 3, true)
	if !hasKeyword(g, permanent, game.Deathtouch) {
		t.Fatal("Wren's Run Vanquisher did not have deathtouch after resolving")
	}
}

type revealOrPayCostAgent struct {
	simPassPolicy
	costLabel string
	chosen    bool
}

func (a *revealOrPayCostAgent) ChooseChoice(_ PlayerObservation, request game.ChoiceRequest) []int {
	if request.Kind == game.ChoicePayment && request.Prompt == "Choose spell cost" {
		for i, option := range request.Options {
			if option.Label == a.costLabel {
				a.chosen = true
				return []int{i}
			}
		}
	}
	return request.DefaultSelection
}

func castRevealOrPayCreature(
	t *testing.T,
	g *game.Game,
	oracleID, name string,
	subtype types.Sub,
	costLabel string,
	coloredMana mana.Color,
	printedGeneric, choiceGeneric int,
	wantReveal bool,
) (*game.Game, *Engine, *game.Permanent) {
	t.Helper()
	if g == nil {
		g = game.NewGame([game.NumPlayers]game.PlayerConfig{})
	}
	engine := NewEngine(nil)
	def, err := cards.NewDefaultRegistry().LookupOracleID(oracleID)
	if err != nil {
		t.Fatalf("default registry lookup for %s failed: %v", name, err)
	}
	if def.Name != name {
		t.Fatalf("Oracle ID %s resolved to %q, want %q", oracleID, def.Name, name)
	}

	spellID := addCardToHand(g, game.Player1, def)
	revealedID := addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name: "Matching subtype card", Types: []types.Card{types.Creature}, Subtypes: []types.Sub{subtype},
	}})
	unrelatedID := addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name: "Unrelated card", Types: []types.Card{types.Sorcery},
	}})
	g.Players[game.Player1].ManaPool.Add(coloredMana, 1)
	if generic := printedGeneric + choiceGeneric; generic > 0 {
		g.Players[game.Player1].ManaPool.Add(mana.C, generic)
	}

	agent := &revealOrPayCostAgent{costLabel: costLabel}
	policies := simPassPolicies()
	policies[game.Player1] = agent
	setMainPhasePriority(g, game.Player1)
	if !engine.applyActionWithChoices(g, game.Player1, action.CastSpell(spellID, nil, 0, nil), policies, &TurnLog{}) {
		t.Fatalf("casting %s with cost choice %q failed", name, costLabel)
	}
	if !agent.chosen {
		t.Fatalf("casting %s did not offer cost choice %q", name, costLabel)
	}
	wantRemaining := choiceGeneric
	if !wantReveal {
		wantRemaining = 0
	}
	if got := g.Players[game.Player1].ManaPool.Amount(mana.C); got != wantRemaining {
		t.Fatalf("%s left %d generic mana, want %d after its selected additional-cost branch", name, got, wantRemaining)
	}
	if !g.Players[game.Player1].Hand.Contains(revealedID) ||
		!g.Players[game.Player1].Hand.Contains(unrelatedID) {
		t.Fatalf("%s moved a hand card while paying its additional cost", name)
	}

	revealEvents := 0
	for _, event := range g.Events {
		if event.Kind != game.EventCardRevealed {
			continue
		}
		revealEvents++
		if event.CardID != revealedID || event.SourceID != spellID ||
			event.Player != game.Player1 || event.FromZone != zone.Hand {
			t.Fatalf("%s revealed an unexpected card for its additional cost: %+v", name, event)
		}
	}
	if wantReveal && revealEvents != 1 {
		t.Fatalf("%s produced %d reveal events, want exactly one matching card", name, revealEvents)
	}
	if !wantReveal && revealEvents != 0 {
		t.Fatalf("%s revealed a hand card after choosing its mana branch", name)
	}

	engine.resolveTopOfStackWithChoices(g, simPassPolicies(), &TurnLog{})
	permanent, ok := findPermanentByCardID(g, spellID)
	if !ok {
		t.Fatalf("%s did not enter the battlefield after its additional cost was paid", name)
	}
	return g, engine, permanent
}
