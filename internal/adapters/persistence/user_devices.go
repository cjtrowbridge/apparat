// Stores verified public profiles and user-signed device bindings in a single transaction.
// A public binding identifies its owner; service authorization requires separate policy.
package persistence

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"

	"github.com/cjtrowbridge/apparat/internal/identity"
)

const personalSchema = `
CREATE TABLE personal_profiles (
    user_id TEXT PRIMARY KEY, revision INTEGER NOT NULL CHECK(revision>0), signed_profile BLOB NOT NULL
);
CREATE TABLE personal_device_bindings (
    user_id TEXT NOT NULL REFERENCES personal_profiles(user_id),
    device_id TEXT NOT NULL, signed_binding BLOB NOT NULL, PRIMARY KEY(user_id,device_id)
);
`

func (db *DB) RememberDevice(ctx context.Context, profile identity.SignedProfile, binding identity.SignedBinding) error {
	if !profile.Valid() || !binding.Valid() || !bytes.Equal(profile.Profile.Public, binding.Binding.User) {
		return errors.New("invalid personal credentials")
	}
	user := identity.Fingerprint(profile.Profile.Public)
	device := identity.Fingerprint(binding.Binding.Device)
	public, _ := json.Marshal(profile)
	signed, _ := json.Marshal(binding)
	tx, err := db.sql.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var old []byte
	var revision uint64
	err = tx.QueryRowContext(ctx, "SELECT revision,signed_profile FROM personal_profiles WHERE user_id=?", user).Scan(&revision, &old)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return err
	}
	if revision == profile.Profile.Revision && !bytes.Equal(old, public) {
		return errors.New("conflicting personal profile revision")
	}
	if revision < profile.Profile.Revision {
		_, err = tx.ExecContext(ctx, "INSERT INTO personal_profiles(user_id,revision,signed_profile) VALUES(?,?,?) ON CONFLICT(user_id) DO UPDATE SET revision=excluded.revision,signed_profile=excluded.signed_profile", user, profile.Profile.Revision, public)
		if err != nil {
			return err
		}
	}
	_, err = tx.ExecContext(ctx, "INSERT INTO personal_device_bindings(user_id,device_id,signed_binding) VALUES(?,?,?) ON CONFLICT(user_id,device_id) DO NOTHING", user, device, signed)
	if err != nil {
		return err
	}
	return tx.Commit()
}
