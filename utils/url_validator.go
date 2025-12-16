package utils

import (
	"errors"
	"net/url"
	"strings"
)

var (
	ErrInvalidURL        = errors.New("invalid URL format")
	ErrSelfShortening    = errors.New("cannot create short URL for this domain")
	ErrUnsupportedScheme = errors.New("only HTTP and HTTPS URLs are supported")
)

func ValidateURL(originalURL string, serviceDomain string) error {

	if originalURL == "" {
		return ErrInvalidURL
	}

	parsedURL, err := url.Parse(originalURL)
	if err != nil {
		return ErrInvalidURL
	}

	if parsedURL.Scheme == "" {
		return ErrUnsupportedScheme
	}

	if parsedURL.Host == "" {
		return ErrInvalidURL
	}

	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return ErrUnsupportedScheme
	}

	if isSameDomain(parsedURL.Host, serviceDomain) {
		return ErrSelfShortening
	}

	return nil
}

func isSameDomain(urlHost, serviceDomain string) bool {
	if serviceDomain == "" {
		return false
	}
	uHost := urlHost
	if parsed, err := url.Parse("http://" + urlHost); err == nil && parsed.Hostname() != "" {
		uHost = parsed.Hostname()
	}

	sHost := serviceDomain
	if parsed, err := url.Parse(serviceDomain); err == nil && parsed.Hostname() != "" {
		sHost = parsed.Hostname()
	}

	return strings.EqualFold(uHost, sHost)
}
