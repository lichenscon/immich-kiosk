package immich

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/charmbracelet/log"
	"github.com/damongolding/immich-kiosk/internal/cache"
	"github.com/google/go-querystring/query"
)

// RandomImage fetches a random image from the Immich API while handling caching and retries.
//
// This function performs the following:
// - Makes an API request to get random images based on configured parameters
// - Caches results to optimize subsequent requests
// - Filters images based on type, trash/archive status, and aspect ratio
// - Retries up to MaxRetries times if no suitable images are found
// - Updates the cache to remove used images
//
// Parameters:
//   - requestID: Unique identifier for tracking and logging
//   - deviceID: ID of the device making the request
//   - isPrefetch: Indicates if this is a prefetch request
//
// Returns an error if no suitable image is found after retries or if there
// are any issues with API calls, caching, or image processing.
func (i *ImmichAsset) RandomImage(requestID, deviceID string, isPrefetch bool) error {

	if isPrefetch {
		log.Debug(requestID, "PREFETCH", deviceID, "Getting Random image", true)
	} else {
		log.Debug(requestID + " Getting Random image")
	}

	for retries := 0; retries < MaxRetries; retries++ {

		var immichAssets []ImmichAsset

		u, err := url.Parse(requestConfig.ImmichUrl)
		if err != nil {
			_, _, err = immichApiFail(immichAssets, err, nil, "")
			return err
		}

		requestBody := ImmichSearchRandomBody{
			Type:       string(ImageType),
			WithExif:   true,
			WithPeople: true,
			Size:       requestConfig.Kiosk.FetchedAssetsSize,
		}

		if requestConfig.ShowArchived {
			requestBody.WithArchived = true
		}

		// convert body to queries so url is unique and can be cached
		queries, _ := query.Values(requestBody)

		apiUrl := url.URL{
			Scheme:   u.Scheme,
			Host:     u.Host,
			Path:     "api/search/random",
			RawQuery: fmt.Sprintf("kiosk=%x", sha256.Sum256([]byte(queries.Encode()))),
		}

		jsonBody, err := json.Marshal(requestBody)
		if err != nil {
			_, _, err = immichApiFail(immichAssets, err, nil, "")
			return err
		}

		immichApiCall := immichApiCallDecorator(i.immichApiCall, requestID, deviceID, immichAssets)
		apiBody, err := immichApiCall("POST", apiUrl.String(), jsonBody)
		if err != nil {
			_, _, err = immichApiFail(immichAssets, err, apiBody, apiUrl.String())
			return err
		}

		err = json.Unmarshal(apiBody, &immichAssets)
		if err != nil {
			_, _, err = immichApiFail(immichAssets, err, apiBody, apiUrl.String())
			return err
		}

		apiCacheKey := cache.ApiCacheKey(apiUrl.String(), deviceID)

		if len(immichAssets) == 0 {
			log.Debug(requestID + " No images left in cache. Refreshing and trying again")
			cache.Delete(apiCacheKey)
			continue
		}

		for immichAssetIndex, img := range immichAssets {

			// We only want images and that are not trashed or archived (unless wanted by user)
			isInvalidType := img.Type != ImageType
			isTrashed := img.IsTrashed
			isArchived := img.IsArchived && !requestConfig.ShowArchived
			isInvalidRatio := !i.ratioCheck(&img)

			if isInvalidType || isTrashed || isArchived || isInvalidRatio {
				continue
			}

			if requestConfig.Kiosk.Cache {
				// Remove the current image from the slice
				immichAssetsToCache := append(immichAssets[:immichAssetIndex], immichAssets[immichAssetIndex+1:]...)
				jsonBytes, err := json.Marshal(immichAssetsToCache)
				if err != nil {
					log.Error("Failed to marshal immichAssetsToCache", "error", err)
					return err
				}

				// replace cwith cache minus used image
				err = cache.Replace(apiCacheKey, jsonBytes)
				if err != nil {
					log.Debug("cache not found!")
				}
			}

			*i = img
			return nil
		}

		log.Debug(requestID + " No viable images left in cache. Refreshing and trying again")
		cache.Delete(apiCacheKey)
	}
	return fmt.Errorf("No images found for random. Max retries reached.")
}
