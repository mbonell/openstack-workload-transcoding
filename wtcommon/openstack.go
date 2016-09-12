package wtcommon

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
	"time"

	"github.com/rackspace/gophercloud"
	"github.com/rackspace/gophercloud/openstack"
	"github.com/rackspace/gophercloud/openstack/objectstorage/v1/objects"
)

// constants with the name of the containers for jobs
const (
	SOURCE_MEDIA_CONTAINER     = "media-source"
	TRANSCODED_MEDIA_CONTAINER = "media-transcoding"
)

// getProvider returns the provider
func GetProvider() (*gophercloud.ProviderClient, error) {
	// Get authentication info
	authOpts, err := openstack.AuthOptionsFromEnv()
	if err != nil {
		return nil, err
	}

	// Get provider
	provider, err := openstack.AuthenticatedClient(authOpts)
	if err != nil {
		return nil, err
	}

	return provider, nil
}

// getProvider returns the object storage
func GetServiceObjectStorage(provider *gophercloud.ProviderClient) (*gophercloud.ServiceClient, error) {
	// Get a service for ObjectStorage
	service, err := openstack.NewObjectStorageV1(provider, gophercloud.EndpointOpts{
		Region: "RegionOne",
	})
	if err != nil {
		return nil, err
	}

	return service, nil
}

// downloadFile downloads a file from an URL into a temp file
func downloadFile(url string) (string, error) {
	// Create a temp file
	tmpfile, err := ioutil.TempFile(os.TempDir(), "media")
	if err != nil {
		return "", err
	}
	defer tmpfile.Close()

	// Download the data
	resp, err := http.Get(url)
	if err != nil {
		return tmpfile.Name(), err
	}
	defer resp.Body.Close()

	// Copy body into tmpfile
	_, err = io.Copy(tmpfile, resp.Body)
	if err != nil {
		return tmpfile.Name(), err
	}

	return tmpfile.Name(), nil
}
func IsValidURL(url2check string) bool {
	// Check correct prefix
	if strings.HasPrefix(url2check, "http://") || strings.HasPrefix(url2check, "https://") {
		// Double check is a valid URL
		_, err := url.Parse(url2check)
		if err == nil {
			return true
		}
	}

	return false
}

// Upload2ObjectStorage uploads the media (url or file) into object storage
func Upload2ObjectStorage(service *gophercloud.ServiceClient, mediaPath string, filename string, containerName string) (string, error) {
	var fn string

	// If is a URL let's download it
	if IsValidURL(mediaPath) {
		// Download file from URL
		fn, err := downloadFile(mediaPath)
		if fn != "" {
			defer os.Remove(fn)
		}

		if err != nil {
			return "", err
		}
	} else { // File, let's verify it exists
		fn = mediaPath

		// Validate file exists
		_, err := os.Stat(fn)
		if err != nil {
			return "", err
		}
	}

	// Open file for reading
	f, err := os.Open(fn)
	if err != nil {
		return "", err
	}
	defer f.Close()

	// Upload to Object Storage
	ext := path.Ext(filename)
	name := fmt.Sprintf("%s-%d%s", filename[:len(filename)-len(ext)], time.Now().UnixNano(), ext)
	res := objects.Create(service, containerName, name, f, nil)
	_, err = res.ExtractHeader()
	if err != nil {
		return "", err
	}

	return name, nil
}

func DownloadFromObjectStorage(service *gophercloud.ServiceClient, objectName, filename string) error {
	// Save object
	res := objects.Download(service, SOURCE_MEDIA_CONTAINER, objectName, nil)
	content, err := res.ExtractContent()

	err = ioutil.WriteFile(filename, []byte(content), 0644)

	return err
}
