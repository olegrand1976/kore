-- Global login uniqueness for active users (cross-tenant). Soft-deleted rows are excluded.
CREATE UNIQUE INDEX IF NOT EXISTS idx_org_users_login_global
	ON org.users (lower(login))
	WHERE deleted_at IS NULL;
