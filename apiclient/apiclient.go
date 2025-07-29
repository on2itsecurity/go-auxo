package apiclient

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"golang.org/x/net/http2"
)

// apiConfig contains the config for accessing the API
type apiConfig struct {
	address      string
	token        string
	tr           *http2.Transport
	debug        bool
	contentType  string
	timeoutInSec int
}

// APIClient server details
type APIClient struct {
	config apiConfig
}

// NewAPIClient - returns a new APIClient object
func NewAPIClient(address, token string, debug bool) (*APIClient, error) {
	//Init / Constructor
	APIClient := new(APIClient)

	//Set the configuration
	APIClient.config.token = token
	APIClient.config.address = address
	APIClient.config.debug = debug
	APIClient.config.contentType = "application/json"
	APIClient.config.tr = &http2.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: false},
	}
	APIClient.config.timeoutInSec = 5

	return APIClient, nil
}

// ApiCall - makes an API call to the ZeroTrust API
// Returns the response body or an error
func (c *APIClient) ApiCall(ctx context.Context, uri string, method string, data string) ([]byte, error) {
	if ctx == nil {
		// Use default timeout if no context provided
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), time.Duration(c.config.timeoutInSec)*time.Second)
		defer cancel()
	}

	// Remove timeout from client - let context handle it
	client := &http.Client{Transport: c.config.tr}
	debug := c.config.debug

	address := "https://" + c.config.address + uri

	if debug {
		fmt.Println("Method: " + method)
		fmt.Println("Address: " + address)
	}

	req, err := http.NewRequestWithContext(ctx, method, address, bytes.NewBuffer([]byte(data)))

	if debug {
		fmt.Println("Data:" + data)
	}

	if err != nil {
		return []byte{}, err
	}

	req.Header.Add("Content-Type", c.config.contentType)
	req.Header.Add("Authorization", "Bearer "+c.config.token)
	//req.Header.Set("Connection", "close") //Optional

	response, err := client.Do(req)

	var body []byte
	if err != nil {
		if debug {
			fmt.Println("Error", err)
		}
		return []byte{}, err
	}

	body, err = ioutil.ReadAll(response.Body)

	if err != nil {
		if debug {
			fmt.Println("Error", err)
		}
		return body, err
	}

	defer req.Body.Close()
	defer response.Body.Close()

	if debug {
		fmt.Printf("Response code: %d\n", response.StatusCode)
	}

	if (response.StatusCode != 200) && (response.StatusCode != 201) {
		if debug {
			fmt.Println("Error", err)
		}
		return body, fmt.Errorf("Not 200 or 201 ok, but %d, with body %s", response.StatusCode, string(body))
	}

	if debug {
		fmt.Println("Return: " + string(body))
	}

	return body, nil
}

// SetTimeout - sets the default timeout used when context is nil
func (c *APIClient) SetTimeout(seconds int) {
	c.config.timeoutInSec = seconds
}
