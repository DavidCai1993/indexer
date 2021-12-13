package fetcher

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"github.com/neilotoole/errgroup"
	"github.com/valyala/fastjson"
	"go.uber.org/zap"
)

const (
	FetchRateLimit   = 100
	IdentityApiCount = 3
)

func (f *fetcher) FetchIdentity(address string) (IdentityEntryList, error) {

	var identityArr IdentityEntryList
	ch := make(chan IdentityEntry)

	// Part 1 - Demo data source
	// Context API
	go f.processContext(address, ch)
	// Superrare API
	go f.processSuperrare(address, ch)
	// Part 2 - Add other data source here
	// Poap API
	go f.processPoap(address, ch)

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
		if entry.Poap != nil {
			identityArr.Poap = append(identityArr.Poap, entry.Poap...)
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

func (f *fetcher) processPoap(address string, ch chan<- IdentityEntry) {
	var result IdentityEntry

	defer func() { ch <- result }()

	body, err := sendRequest(f.httpClient, RequestArgs{
		url:    fmt.Sprintf(PoapScanUrl, address),
		method: "GET",
	})

	if err != nil {
		result.Err = err
		result.Msg = "[processPoap] fetch identity failed"
		return
	}

	rawValue, err := fastjson.ParseBytes(body)

	if err != nil {
		result.Err = err
		result.Msg = "[processPoap] parse response json failed"
		return
	}

	tokens := rawValue.GetArray()

	// Use errorgroup to fetch the recommendations parallelly at a
	// safe request rate and catch the request errors
	g, _ := errgroup.WithContextN(context.Background(), FetchRateLimit, len(tokens))

	for i := range tokens {
		func(token *fastjson.Value) {
			g.Go(func() error {
				event := token.Get("event")

				poapIdentity := UserPoapIdentity{
					EventID:    event.GetInt("id"),
					Supply:     event.GetInt("supply"),
					Year:       event.GetInt("year"),
					TokenID:    string(token.GetStringBytes("tokenId")),
					Owner:      string(token.GetStringBytes("owner")),
					EventDesc:  string(event.GetStringBytes("description")),
					FancyID:    string(event.GetStringBytes("fancy_id")),
					EventName:  string(event.GetStringBytes("name")),
					EventUrl:   string(event.GetStringBytes("event_url")),
					ImageUrl:   string(event.GetStringBytes("image_url")),
					Country:    string(event.GetStringBytes("country")),
					City:       string(event.GetStringBytes("city")),
					StartDate:  string(event.GetStringBytes("start_date")),
					EndDate:    string(event.GetStringBytes("end_date")),
					ExpiryDate: string(event.GetStringBytes("expiry_date")),
					DataSource: POAP,
				}

				recommendations, err := f.getPoapRecommendation(event.GetInt("id"))

				if err != nil {
					return err
				}

				poapIdentity.Recommendations = recommendations

				result.Poap = append(result.Poap, poapIdentity)

				return nil
			})
		}(tokens[i])
	}

	if err := g.Wait(); err != nil {
		result.Err = err
		result.Msg = "[processPoap] get recommendations failed"
		return
	}

	ch <- result
}

func (f *fetcher) getPoapRecommendation(eventID int) ([]PoapRecommendation, error) {
	var results []PoapRecommendation

	query := map[string]string{
		"query": fmt.Sprintf(`
			{
				event(id: "%d") {
					tokens {
						id
						owner {
							id
						}
					}
				}
			}
		 `, eventID),
	}

	queryBytes, _ := json.Marshal(query)

	body, err := sendRequest(f.httpClient, RequestArgs{
		url:    PoapSubgraphUrl,
		method: "POST",
		body:   bytes.NewBuffer(queryBytes).Bytes(),
	})

	if err != nil {
		return nil, err
	}

	rawValue, err := fastjson.ParseBytes(body)

	if err != nil {
		return nil, err
	}

	for _, token := range rawValue.GetArray("data", "event", "tokens") {
		results = append(results, PoapRecommendation{
			TokenID: string(token.GetStringBytes("id")),
			Address: string(token.GetStringBytes("owner", "id")),
			EventID: eventID,
		})
	}

	return results, nil
}
