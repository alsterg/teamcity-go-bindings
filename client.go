package teamcity

import "github.com/sethgrid/pester"

func New(url, authtoken string, concurrencyLimit int) *Client {
	if concurrencyLimit == 0 {
		concurrencyLimit = 1000
	}

	http := pester.New()
	http.Concurrency = concurrencyLimit
	http.MaxRetries = 5
	http.Backoff = pester.ExponentialBackoff
	http.KeepLog = true

	client := &Client{
		HTTPClient: http,
		URL:        url,
		Authtoken:  authtoken,
		Flow:       make(chan DataFlow, 10000),
		semaphore:  make(chan bool, concurrencyLimit),
	}

	go client.processDataFlow()

	return client
}

func (c *Client) Close() {
	close(c.Flow)
}
