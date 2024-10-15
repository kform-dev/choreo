/*
Copyright 2024 Nokia.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package git

import (
	"context"
	"net/http"
	"time"

	"github.com/henderiw/logger/log"
)

const url = "https://api.github.com"

func CheckOnline(ctx context.Context) bool {
	log := log.FromContext(ctx)
	client := http.Client{
		Timeout: 5 * time.Second,
	}

	// Create a new request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Error("failed to create request", "url", url, "err", err)
		return false
	}

	// Set cache-control headers
	req.Header.Set("Cache-Control", "no-cache, no-store, must-revalidate")
	req.Header.Set("Pragma", "no-cache") // HTTP/1.0 caches might not understand Cache-Control
	req.Header.Set("Expires", "0")       // Proxies

	// Send the request
	resp, err := client.Do(req)
	if err != nil {
		log.Error("failed to reach url", "url", url, "err", err)
		return false
	}
	defer resp.Body.Close() //

	// Check if the HTTP status code is in the 200-299 range
	return resp.StatusCode >= 200 && resp.StatusCode < 300
}
