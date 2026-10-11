package cardgen

import "testing"

func TestUnmodeledIteratorDomainsRemainWholeCardRefusals(t *testing.T) {
	t.Parallel()
	for _, body := range []string{
		"Exile target player's graveyard. For each artifact or land card exiled this way, create a 1/1 colorless Insect artifact creature token with flying.",
		"For each Equipment attached to this creature, create a 1/1 white Soldier creature token.",
		"Mill three cards. For each creature card milled this way, create a 2/2 black Necron Warrior artifact creature token.",
		"Exile target opponent's graveyard. For each creature card exiled this way, create a 1/1 colorless Spirit creature token.",
		"Discard two cards, then draw two cards. For each nonland card discarded this way, create a 1/1 red Elemental creature token.",
		"For each nontoken creature you controlled that died this turn, create a 2/2 black Zombie creature token.",
		"Mill two cards. For each creature card milled this way, create a 1/1 green Saproling creature token.",
	} {
		t.Run(body, func(t *testing.T) {
			t.Parallel()
			assertCardUnsupported(t, targetCountTestCard(body), "unsupported")
		})
	}
}
