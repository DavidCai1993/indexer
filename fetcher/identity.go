package fetcher

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/valyala/fastjson"
	"go.uber.org/zap"
)

const IdentityApiCount = 8

func (f *fetcher) FetchIdentity(address string) (IdentityEntryList, error) {

	var identityArr IdentityEntryList
	ch := make(chan IdentityEntry)

	// Part 1 - Demo data source
	// Context API
	go f.processContext(address, ch)
	// Superrare API
	go f.processSuperrare(address, ch)
	// Part 2 - Add other data source here
	// ENS reverse resolution
	go f.processENS(address, ch)
	// OpenSeaAPI
	go f.processOpenSea(address, ch)
	// Foundation Subgraph
	go f.processFoundation(address, ch)
	// Rarible API
	go f.processRarible(address, ch)
	// Zora API
	go f.processZora(address, ch)
	// Showtime API
	go f.processShowtime(address, ch)

	// Final Part - Merge entry
	for i := 0; i < IdentityApiCount; i++ {
		entry := <-ch
		if entry.Err != nil {
			zap.L().With(zap.Error(entry.Err)).Error("identity api error: " + entry.Msg)
			continue
		}
		if entry.OpenSea != nil {
			identityArr.OpenSea = append(identityArr.OpenSea, *entry.OpenSea)
		}
		if entry.Twitter != nil {
			entry.Twitter.Handle = convertTwitterHandle(entry.Twitter.Handle)
			identityArr.Twitter = append(identityArr.Twitter, *entry.Twitter)
		}
		if entry.Superrare != nil {
			identityArr.Superrare = append(identityArr.Superrare, *entry.Superrare)
		}
		if entry.Rarible != nil {
			identityArr.Rarible = append(identityArr.Rarible, *entry.Rarible)
		}
		if entry.Context != nil {
			identityArr.Context = append(identityArr.Context, *entry.Context)
		}
		if entry.Zora != nil {
			identityArr.Zora = append(identityArr.Zora, *entry.Zora)
		}
		if entry.Foundation != nil {
			identityArr.Foundation = append(identityArr.Foundation, *entry.Foundation)
		}
		if entry.Showtime != nil {
			identityArr.Showtime = append(identityArr.Showtime, *entry.Showtime)
		}
		if entry.Ens != nil {
			identityArr.Ens = entry.Ens.Ens
		}
	}

	return identityArr, nil
}

func (f *fetcher) processContext(address string, ch chan<- IdentityEntry) {
	var result IdentityEntry

	body, err := sendRequest(f.httpClient, RequestArgs{
		url:    fmt.Sprintf(ContextUrl, address),
		method: "GET",
	})
	if err != nil {
		result.Err = err
		result.Msg = "[processContext] fetch identity failed"
		ch <- result
		return
	}
	contextProfile := ContextAppResp{}
	err = json.Unmarshal(body, &contextProfile)
	if err != nil {
		result.Err = err
		result.Msg = "[processContext] identity response json unmarshal failed"
		ch <- result
		return
	}

	if value, ok := contextProfile.Ens[address]; ok {
		result.Ens = &UserEnsIdentity{
			Ens:        value,
			DataSource: CONTEXT,
		}
	}

	for _, profileList := range contextProfile.Profiles {
		for _, entry := range profileList {
			switch entry.Contract {
			case SuperrareContractAddress:
				result.Superrare = &UserSuperrareIdentity{
					Homepage:   entry.Url,
					Username:   entry.Username,
					DataSource: CONTEXT,
				}
			case OpenSeaContractAddress:
				result.OpenSea = &UserOpenSeaIdentity{
					Homepage:   entry.Url,
					Username:   entry.Username,
					DataSource: CONTEXT,
				}
			case RaribleContractAddress:
				result.Rarible = &UserRaribleIdentity{
					Homepage:   entry.Url,
					Username:   entry.Username,
					DataSource: CONTEXT,
				}
			case FoundationContractAddress:
				result.Foundation = &UserFoundationIdentity{
					Website:    entry.Website,
					Username:   entry.Username,
					DataSource: CONTEXT,
				}
			case ZoraContractAddress:
				result.Zora = &UserZoraIdentity{
					Website:    entry.Website,
					Username:   entry.Username,
					DataSource: CONTEXT,
				}
			case ContextContractAddress:
				result.Context = &UserContextIdentity{
					Username:      entry.Username,
					Website:       entry.Website,
					FollowerCount: contextProfile.FollowerCount,
					DataSource:    CONTEXT,
				}
			default:
			}
		}
	}

	ch <- result
	return
}

func (f *fetcher) processSuperrare(address string, ch chan<- IdentityEntry) {
	var result IdentityEntry

	body, err := sendRequest(f.httpClient, RequestArgs{
		url:    fmt.Sprintf(SuperrareUrl, address),
		method: "GET",
	})
	if err != nil {
		result.Err = err
		result.Msg = "[processSuperrare] fetch identity failed"
		ch <- result
		return
	}

	sprProfile := SuperrareProfile{}
	err = json.Unmarshal(body, &sprProfile)
	if err != nil {
		result.Err = err
		result.Msg = "[processSuperrare] identity response json unmarshal failednti"
		ch <- result
		return
	}

	newSprRecord := UserSuperrareIdentity{
		Username:       sprProfile.Result.Username,
		Location:       sprProfile.Result.Location,
		Bio:            sprProfile.Result.Bio,
		InstagramLink:  sprProfile.Result.InstagramLink,
		TwitterLink:    sprProfile.Result.TwitterLink,
		SteemitLink:    sprProfile.Result.SteemitLink,
		Website:        sprProfile.Result.Website,
		SpotifyLink:    sprProfile.Result.SpotifyLink,
		SoundCloudLink: sprProfile.Result.SoundCloudLink,
		DataSource:     SUPERRARE,
	}

	if newSprRecord.Username != "" || newSprRecord.Location != "" || newSprRecord.Bio != "" || newSprRecord.InstagramLink != "" ||
		newSprRecord.TwitterLink != "" || newSprRecord.SteemitLink != "" || newSprRecord.Website != "" ||
		newSprRecord.SpotifyLink != "" || newSprRecord.SoundCloudLink != "" {
		result.Superrare = &newSprRecord
	}

	ch <- result
}

func (f *fetcher) processFoundation(address string, ch chan<- IdentityEntry) {
	var result IdentityEntry

	defer func() { ch <- result }()

	query := map[string]interface{}{
		"query": `
				query UserProfileByPublicKey($publicKey: String!) {
					user: user_by_pk(publicKey: $publicKey) {
						publicKey
						username
						bio
						links
						twitSocialVerifs: socialVerifications(
							where: { isValid: { _eq: true }, service: { _eq: "TWITTER" } }
							limit: 1
						) {
							userId
							username
						}
						instaSocialVerifs: socialVerifications(
							where: { isValid: { _eq: true }, service: { _eq: "INSTAGRAM" } }
							limit: 1
						) {
							userId
							username
						}
					}
				}
		`,
		"variables": map[string]string{"publicKey": address},
	}

	queryBytes, _ := json.Marshal(query)

	body, err := sendRequest(f.httpClient, RequestArgs{
		url:    FoundationUrl,
		method: "POST",
		body:   bytes.NewBuffer(queryBytes).Bytes(),
	})

	if err != nil {
		result.Err = err
		result.Msg = "[processFoundation] fetch identity failed"
		return
	}

	username := fastjson.GetString(body, "data", "user", "username")

	if len(username) != 0 {
		result.Foundation = &UserFoundationIdentity{
			Username:   username,
			Bio:        fastjson.GetString(body, "data", "user", "bio"),
			Tiktok:     fastjson.GetString(body, "data", "user", "links", "tiktok", "handle"),
			Twitch:     fastjson.GetString(body, "data", "user", "links", "twitch", "handle"),
			Discord:    fastjson.GetString(body, "data", "user", "links", "discord", "handle"),
			Twitter:    fastjson.GetString(body, "data", "user", "twitSocialVerifs", "0", "username"),
			Website:    fastjson.GetString(body, "data", "user", "links", "website", "handle"),
			Youtube:    fastjson.GetString(body, "data", "user", "links", "youtube", "handle"),
			Facebook:   fastjson.GetString(body, "data", "user", "links", "facebook", "handle"),
			Snapchat:   fastjson.GetString(body, "data", "user", "links", "snapchat", "handle"),
			Instagram:  fastjson.GetString(body, "data", "user", "instaSocialVerifs", "0", "username"),
			DataSource: FOUNDATION,
		}
	}
}

func (f *fetcher) processRarible(address string, ch chan<- IdentityEntry) {
	var result IdentityEntry

	defer func() { ch <- result }()

	body, err := sendRequest(f.httpClient, RequestArgs{
		url:    fmt.Sprintf(RaribleUrl, address),
		method: "GET",
	})

	if err != nil {
		result.Err = err
		result.Msg = "[processRarible] fetch identity failed"
		return
	}

	username := fastjson.GetString(body, "username")

	if len(username) != 0 {
		result.Rarible = &UserRaribleIdentity{
			Username:   username,
			Homepage:   fastjson.GetString(body, "website"),
			Twitter:    fastjson.GetString(body, "twitterUsername"),
			DataSource: RARIBLE,
		}
	}
}

func (f *fetcher) processZora(address string, ch chan<- IdentityEntry) {
	var result IdentityEntry

	defer func() { ch <- result }()

	body, err := sendRequest(f.httpClient, RequestArgs{
		url:    fmt.Sprintf(ZoraUrl, address),
		method: "GET",
	})

	if err != nil {
		result.Err = err
		result.Msg = "[processZora] fetch identity failed"
		return
	}

	username := fastjson.GetString(body, "0", "username")

	if len(username) != 0 {
		result.Zora = &UserZoraIdentity{
			Username:   username,
			Website:    fastjson.GetString(body, "0", "website"),
			Bio:        fastjson.GetString(body, "0", "bio"),
			DataSource: ZORA,
		}
	}
}

func (f *fetcher) processShowtime(address string, ch chan<- IdentityEntry) {
	var result IdentityEntry

	defer func() { ch <- result }()

	body, err := sendRequest(f.httpClient, RequestArgs{
		url:    fmt.Sprintf(ShowtimeUrl, address),
		method: "GET",
	})

	if err != nil {
		result.Err = err
		result.Msg = "[processShowtime] fetch identity failed"
		return
	}

	rawValue, err := fastjson.ParseBytes(body)

	if err != nil {
		result.Err = err
		result.Msg = "[processShowtime] parse response json failed"
		return
	}

	name := fastjson.GetString(body, "pageProps", "profile", "name")

	if len(name) != 0 {
		result.Showtime = &UserShowtimeIdentity{
			Name:       name,
			Username:   fastjson.GetString(body, "pageProps", "profile", "username"),
			Bio:        fastjson.GetString(body, "pageProps", "profile", "bio"),
			DataSource: SHOWTIME,
		}

		links := rawValue.GetArray("pageProps", "profile", "links")

		if len(links) == 0 {
			return
		}

		for _, link := range links {
			handle := string(link.GetStringBytes("user_input"))

			switch string(link.GetStringBytes("type__name")) {
			case "Twitter":
				result.Showtime.TwitterHandle = handle
			case "Linktree":
				result.Showtime.LinkTreeHandle = handle
			case "CryptoArt.ai":
				result.Showtime.CryptoArtHandle = handle
			case "Foundation":
				result.Showtime.FoundationHandle = handle
			case "hicetnunc.art":
				result.Showtime.HicetnuncHandle = handle
			case "OpenSea":
				result.Showtime.OpenseaHandle = handle
			case "Rarible":
				result.Showtime.RaribleHandle = handle
			}
		}
	}
}

func (f *fetcher) processOpenSea(address string, ch chan<- IdentityEntry) {
	var result IdentityEntry

	defer func() { ch <- result }()

	body, err := sendRequest(f.httpClient, RequestArgs{
		url:    fmt.Sprintf(OpenSeaUrl, address),
		method: "GET",
	})

	if err != nil {
		result.Err = err
		result.Msg = "[processOpenSea] fetch identity failed"
		return
	}

	username := fastjson.GetString(body, "data", "user", "username")

	if len(username) != 0 {
		result.OpenSea = &UserOpenSeaIdentity{
			Username:   username,
			DataSource: OPENSEA,
		}
	}
}

func (f *fetcher) processENS(address string, ch chan<- IdentityEntry) {
	var result IdentityEntry

	defer func() { ch <- result }()

	query := map[string]string{
		"query": fmt.Sprintf(`
			{
				registrations(where: { registrant: "%s" }) {
					id
					domain {
						name
					}
					registrant {
						id
					}
				}
			}
		 `, strings.ToLower(address)),
	}

	queryBytes, _ := json.Marshal(query)

	body, err := sendRequest(f.httpClient, RequestArgs{
		url:    EnsUrl,
		method: "POST",
		body:   bytes.NewBuffer(queryBytes).Bytes(),
	})

	if err != nil {
		result.Err = err
		result.Msg = "[processENS] fetch identity failed"
		return
	}

	domain := fastjson.GetString(body, "data", "registrations", "0", "domain", "name")

	if len(domain) != 0 {
		result.Ens = &UserEnsIdentity{
			Ens:        domain,
			DataSource: ENS,
		}
	}
}
