// Retains the immutable migration for original worker ownership records.
// Current first-login admission runs in admission.go.
package persistence

const claimSchema = `
CREATE TABLE worker_claim_invitation (
 singleton INTEGER PRIMARY KEY CHECK(singleton=1), worker TEXT NOT NULL,
 digest TEXT NOT NULL, expires_at INTEGER NOT NULL
);
CREATE TABLE worker_cluster (
 singleton INTEGER PRIMARY KEY CHECK(singleton=1), cluster TEXT NOT NULL UNIQUE,
 worker TEXT NOT NULL, owner_user TEXT NOT NULL, authority_device TEXT NOT NULL,
 poll_seconds INTEGER NOT NULL DEFAULT 300 CHECK(poll_seconds>0),
 admission_open INTEGER NOT NULL DEFAULT 1 CHECK(admission_open IN (0,1))
);
CREATE TABLE cluster_users (
 cluster TEXT NOT NULL REFERENCES worker_cluster(cluster),
 user_id TEXT NOT NULL, public_profile BLOB NOT NULL,
 admission TEXT NOT NULL CHECK(admission IN ('approved','blocked','pending')),
 admin INTEGER NOT NULL CHECK(admin IN (0,1) AND (admin=0 OR admission='approved')),
 PRIMARY KEY(cluster,user_id)
);
CREATE TABLE cluster_devices_authorized (
 cluster TEXT NOT NULL, user_id TEXT NOT NULL, device_id TEXT NOT NULL, public_binding BLOB NOT NULL,
 PRIMARY KEY(cluster,user_id,device_id),
 FOREIGN KEY(cluster,user_id) REFERENCES cluster_users(cluster,user_id)
);
`
