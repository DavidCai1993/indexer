package test

import (
	"testing"

	"github.com/cyberconnecthq/indexer/fetcher"
	"github.com/stretchr/testify/assert"
)

func TestProcessOpenSea(t *testing.T) {
	identity, err := fetcher.NewFetcher().FetchIdentity("0x8AcC5677F98b86c407BFA7861f53857430Ba3904")

	assert.Nil(t, err)
	assert.Len(t, identity.OpenSea, 1)
	assert.Equal(t, identity.OpenSea[0].Username, "Vintash")
	assert.Equal(t, identity.OpenSea[0].DataSource, fetcher.OPENSEA)
}

func TestProcessENS(t *testing.T) {
	identity, err := fetcher.NewFetcher().FetchIdentity("0x56706F118e42AE069F20c5636141B844D1324AE1")

	assert.Nil(t, err)
	assert.Equal(t, identity.Ens, "davidcai.eth")
}

func TestProcessFoundation(t *testing.T) {
	identity, err := fetcher.NewFetcher().FetchIdentity("0x56706F118e42AE069F20c5636141B844D1324AE1")

	assert.Nil(t, err)
	assert.Len(t, identity.Foundation, 1)

	foundationIdentity := identity.Foundation[0]
	assert.Equal(t, foundationIdentity.Username, "davidcai")
	assert.Equal(t, foundationIdentity.Bio, "Learning&& coding")
	assert.Equal(t, foundationIdentity.Tiktok, "DavidCai1111")
	assert.Equal(t, foundationIdentity.Twitch, "DavidCai1111")
	assert.Equal(t, foundationIdentity.Discord, "DavidCai#2943")
	assert.Equal(t, foundationIdentity.Twitter, "DavidCai_1993")
	assert.Equal(t, foundationIdentity.Website, "davidc.ai")
	assert.Equal(t, foundationIdentity.Youtube, "https://www.youtube.com/channel/UCDMwJM0qyiWxeqlDYAPXQ3Q")
	assert.Equal(t, foundationIdentity.Facebook, "DavidCai1111")
	assert.Equal(t, foundationIdentity.Snapchat, "DavidCai1111")
	assert.Equal(t, foundationIdentity.Instagram, "davidcai1111")
	assert.Equal(t, foundationIdentity.DataSource, fetcher.FOUNDATION)
}

func TestProcessRarible(t *testing.T) {
	identity, err := fetcher.NewFetcher().FetchIdentity("0xbc67052a5c7dfa09e9fb9ff00b95d8a722d252c5")

	assert.Nil(t, err)
	assert.NotEmpty(t, identity.Rarible)

	for _, raribleIdentity := range identity.Rarible {
		if raribleIdentity.DataSource != fetcher.RARIBLE {
			continue
		}

		assert.Equal(t, raribleIdentity.Username, "Philipp Kapustin")
		assert.Equal(t, raribleIdentity.Homepage, "https://linktr.ee/Philipp_Kapustin")
		assert.Equal(t, raribleIdentity.Twitter, "Phill_Kapustin")
		assert.Equal(t, raribleIdentity.DataSource, fetcher.RARIBLE)

		return
	}

	assert.Fail(t, "RARIBLE datasource result should be found")
}
func TestProcessZora(t *testing.T) {
	identity, err := fetcher.NewFetcher().FetchIdentity("0x6b8C6E15818C74895c31A1C91390b3d42B336799")

	assert.Nil(t, err)
	assert.Len(t, identity.Zora, 1)
	assert.Equal(t, identity.Zora[0].Username, "juliangilliam")
	assert.Equal(t, identity.Zora[0].Bio, "LOGIK aka Julian Gilliam is a multidisciplinary artist who's creating a world in the physical and digital space.")
	assert.Equal(t, identity.Zora[0].Website, "http://www.juliangilliam.com")
	assert.Equal(t, identity.Zora[0].DataSource, fetcher.ZORA)
}

func TestProcessShowtime(t *testing.T) {
	identity, err := fetcher.NewFetcher().FetchIdentity("0x8318b9c0e640114f7ed71de8b65142654573a152")

	assert.Nil(t, err)
	assert.Len(t, identity.Showtime, 1)
	assert.Equal(t, identity.Showtime[0].Name, "Arman Alipour")
	assert.Equal(t, identity.Showtime[0].Username, "armanalipour")
	assert.Equal(t, identity.Showtime[0].Bio, "Independent Artist , 2d Animator")
	assert.Equal(t, identity.Showtime[0].TwitterHandle, "armanalipour")
	assert.Equal(t, identity.Showtime[0].HicetnuncHandle, "tz/tz1UX2tdWZRMiXVYgrfCzUj2PoYrPpxjmbVX")
	assert.Equal(t, identity.Showtime[0].DataSource, fetcher.SHOWTIME)
}
