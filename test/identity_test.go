package test

import (
	"testing"

	"github.com/cyberconnecthq/indexer/fetcher"
	"github.com/stretchr/testify/assert"
)

func TestProcessPoap(t *testing.T) {
	identity, err := fetcher.NewFetcher().FetchIdentity("0x3ab56c8a5E4B307A60b6A769B1C083EE165d6dd6")

	assert.Nil(t, err)
	assert.NotEmpty(t, identity.Poap)

	for _, poapIdentity := range identity.Poap {
		if poapIdentity.TokenID != "2009095" {
			continue
		}

		assert.Equal(t, poapIdentity.Owner, "0x3ab56c8a5E4B307A60b6A769B1C083EE165d6dd6")
		assert.Equal(t, poapIdentity.DataSource, fetcher.POAP)
		assert.Equal(t, poapIdentity.EventID, 9653)
		assert.Equal(t, poapIdentity.TokenID, "2009095")
		assert.Equal(t, poapIdentity.FancyID, "ethereals-moon-mission-poap-4-2021")
		assert.Equal(t, poapIdentity.EventName, "Ethereals Moon Mission POAP #4")
		assert.Equal(t, poapIdentity.EventUrl, "https://discord.gg/etherealswtf")
		assert.Equal(t, poapIdentity.ImageUrl, "https://assets.poap.xyz/ethereals-moon-mission-poap-4-2021-logo-1633505479706.png")
		assert.Equal(t, poapIdentity.Country, "United States")
		assert.Equal(t, poapIdentity.City, "denver")
		assert.Equal(t, poapIdentity.Year, 2021)
		assert.Equal(t, poapIdentity.StartDate, "06-Oct-2021")
		assert.Equal(t, poapIdentity.EndDate, "09-Oct-2021")
		assert.Equal(t, poapIdentity.ExpiryDate, "09-Nov-2021")
		assert.GreaterOrEqual(t, poapIdentity.Supply, 23055)
		assert.Equal(t, poapIdentity.EventDesc, "HOLDERS MUST BE IN ETHEREALS DISCORD IN ORDER TO QUALIFY FOR CONTEST!!! \n\nEthereals Moon Mission is officially engaged and this token will enter you into your chance to win an Ethereal. 5 Winners will be announced on reveal day and the holders must be in The Ethereals Discord in order to qualify for contest. If you are not in discord you will have no way to redeem the prize!")

		assert.NotEmpty(t, poapIdentity.Recommendations)

		assert.Equal(t, poapIdentity.Recommendations[0].TokenID, "2027770")
		assert.Equal(t, poapIdentity.Recommendations[0].Address, "0x2c863b892e10eb7c2b6e527abaaa1e9be47a35d9")
		assert.Equal(t, poapIdentity.Recommendations[0].EventID, 9653)

		return
	}

	assert.Fail(t, "Token 2009095 should be found")
}
