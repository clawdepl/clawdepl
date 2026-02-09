//go:build !debug

package cmd

// HasUnsafeToken always returns false in production builds
func HasUnsafeToken() bool {
	return false
}

// HasUnsafeEndpoint always returns false in production builds
func HasUnsafeEndpoint() bool {
	return false
}

// HasUnsafeAuthEndpoint always returns false in production builds
func HasUnsafeAuthEndpoint() bool {
	return false
}

// GetUnsafeUser always returns empty string in production builds
func GetUnsafeUser() string {
	return ""
}
