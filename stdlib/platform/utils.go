package platform

import (
	"github.com/pkg/browser"
)

// LaunchURL() pushes an OS-level event to open the given URL using the user's default / primary browser.
//
// For future-proofing, LaunchURL() should be used instead of calling browser.OpenURL, allowing its implementation to be switched out.
func LaunchURL(url string) error {
	return browser.OpenURL(url)
}
