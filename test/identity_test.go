package test

import (
	"testing"

	"github.com/cyberconnecthq/indexer/fetcher"
	"github.com/stretchr/testify/assert"
)

func TestProcessTwitterSybilList(t *testing.T) {
	identity, err := fetcher.NewFetcher().FetchIdentity("0x8AcC5677F98b86c407BFA7861f53857430Ba3904")

	assert.Nil(t, err)
	assert.Len(t, identity.Twitter, 1)
	assert.Equal(t, identity.Twitter[0].Handle, "vintash121")
	assert.Equal(t, identity.Twitter[0].DataSource, fetcher.SYBIL)
}
