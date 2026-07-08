package cardgen

import "github.com/natefinch/council4/mtg/game"

// selfExileLinkKey is the canonical linked-object key a permanent uses to tie
// the cards its own "exile ... for as long as it remains exiled, its owner may
// play it" ability places into exile to the "whenever a player plays a card
// exiled with <this permanent>" trigger that pays them off (Prowl, Stoic
// Strategist). The exile-for-play primitive publishes each exiled card under
// this key, and the play trigger matches cards in the same source's pool, so
// both faces of the mechanic agree on one stable, source-scoped link.
const selfExileLinkKey = game.LinkedKey("exiled-with-source")
