package shared

import "errors"

var (
	ErrorConflict  = errors.New("conflicting version signatures")
	ErrorDuplicate = errors.New("duplicate version signatures")
	ErrorNoChanges = errors.New("no changes needed")
)

// Identity is a shared data type which applies to an entity in a remote or local system. It's made to compare signatures between different sources. A source will typically need an Auth string to give access to the entity located on the provided Path and then return its Hash.
type Identity struct {
	Id      string // Id if there is a singular key integer to access the resource, it should be accessible for further requests
	Hash    string // Hash is a sha256 string which represents the signature of the Identity
	Auth    string // Auth is the base64 encoded Private Authentication Token to access the Identity
	Path    string // Path is a URL string representing the location of the Identity
	Version string // Version is a revision string for comparing the same entities at different revisions
}

func (i Identity) HasIdentifiers() bool {
	return i.Id != "" && i.Hash != ""
}

// Next returns the current definition of
func (i Identity) Next(next *Identity) (*Identity, error) {
	if next.Hash == i.Hash {
		if next.Version == i.Version {
			return nil, ErrorNoChanges
		} else {
			return nil, ErrorDuplicate
		}
	} else {
		if next.Version == i.Version {
			return nil, ErrorConflict
		}
	}

	return next, nil
}
