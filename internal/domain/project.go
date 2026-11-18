// Defines the owner-local checkout registration and current checkout observation.
// Credentials and working tree contents remain outside the registration.
package domain

type Project struct {
	ID, Owner, Name, Location, Account string
	Branch                             string
	Actions                            []string
	Available                          bool
}
