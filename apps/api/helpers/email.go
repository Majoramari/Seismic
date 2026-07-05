package helpers

import "strings"

// NormalizeEmail applies Gmail-specific normalization rules
// so dot variations and +tags don't create duplicate accounts.
// Other providers are only lowercased and trimmed.
func NormalizeEmail(email string) string {
	email = strings.TrimSpace(strings.ToLower(email))

	parts := strings.SplitN(email, "@", 2)
	if len(parts) != 2 {
		return email
	}

	local, domain := parts[0], parts[1]

	if domain == "gmail.com" || domain == "googlemail.com" {
		// Remove everything after + (Gmail ignores it)
		if idx := strings.Index(local, "+"); idx != -1 {
			local = local[:idx]
		}
		// Remove all dots (Gmail ignores them)
		local = strings.ReplaceAll(local, ".", "")
		domain = "gmail.com" // normalize googlemail.com too
	}

	return local + "@" + domain
}
