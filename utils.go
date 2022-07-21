package main

import (
	"net/url"
	"path/filepath"
)

const (
	urlSchemaHTTPType int = iota
	urlSchemaSocketType
)

// getURL -
func getURL(urll, token, urlPath string, ttype int) (uu string, err error) {
	u, err := url.Parse(urll)
	if err != nil {
		return
	}

	if token != "" {
		values := u.Query()
		values.Add("token", token)
		u.RawQuery = values.Encode()
	}

	if urlPath != "" {
		u.Path = filepath.Join(u.Path, urlPath)
	}

	switch ttype {
	case urlSchemaHTTPType:
		u.Scheme = "https"
	case urlSchemaSocketType:
		u.Scheme = "wss"
	}

	return u.String(), nil
}

// getStreamURL -
func getStreamURL() (string, error) {
	return getURL(config.Gotify.URL, config.Gotify.Token, "stream", urlSchemaSocketType)
}

// getApplicationURL -
func getApplicationURL() (string, error) {
	return getURL(config.Gotify.URL, config.Gotify.Token, "application", urlSchemaHTTPType)
}

// getImageURL -
func getImageURL(imagePath string) (string, error) {
	return getURL(config.Gotify.URL, "", imagePath, urlSchemaHTTPType)
}
