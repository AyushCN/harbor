package database

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type User struct {
	ID           uuid.UUID
	GithubID     int64
	Username     string
	Name         *string
	AvatarURL    *string
	EncryptedToken *string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type Workspace struct {
	ID        uuid.UUID
	OwnerID   uuid.UUID
	GitURL    string
	GitBranch string
	Name      *string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Environment struct {
	ID              uuid.UUID
	WorkspaceID     uuid.UUID
	HostID          uuid.UUID
	Status          string
	ContainerID     *string
	Port            *int
	PublicURL       *string
	LastActivityAt  *time.Time
	CreatedAt       time.Time
	UpdatedAt       time.Time
	DeletedAt       *time.Time
}

type EnvironmentMember struct {
	ID             uuid.UUID
	EnvironmentID  uuid.UUID
	UserID         uuid.UUID
	Role           string
	InvitedBy      *uuid.UUID
	CreatedAt      time.Time
}

type ShareLink struct {
	ID           uuid.UUID
	EnvironmentID uuid.UUID
	CreatedBy     uuid.UUID
	Token         string
	Role          string
	ExpiresAt     *time.Time
	MaxUses       *int
	UseCount      int
	CreatedAt     time.Time
}

type MemberWithUser struct {
	EnvironmentMember
	User User
}

func (q *Queries) GetUserByID(ctx context.Context, id uuid.UUID) (*User, error) {
	query := `SELECT id, github_id, username, name, avatar_url, encrypted_token, created_at, updated_at FROM users WHERE id = $1`
	row := q.pool.QueryRow(ctx, query, id)
	var u User
	err := row.Scan(&u.ID, &u.GithubID, &u.Username, &u.Name, &u.AvatarURL, &u.EncryptedToken, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (q *Queries) GetUserByGithubID(ctx context.Context, githubID int64) (*User, error) {
	query := `SELECT id, github_id, username, name, avatar_url, encrypted_token, created_at, updated_at FROM users WHERE github_id = $1`
	row := q.pool.QueryRow(ctx, query, githubID)
	var u User
	err := row.Scan(&u.ID, &u.GithubID, &u.Username, &u.Name, &u.AvatarURL, &u.EncryptedToken, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (q *Queries) GetUserByUsername(ctx context.Context, username string) (*User, error) {
	query := `SELECT id, github_id, username, name, avatar_url, encrypted_token, created_at, updated_at FROM users WHERE username = $1`
	row := q.pool.QueryRow(ctx, query, username)
	var u User
	err := row.Scan(&u.ID, &u.GithubID, &u.Username, &u.Name, &u.AvatarURL, &u.EncryptedToken, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (q *Queries) CreateUser(ctx context.Context, githubID int64, username, name, avatarURL, encryptedToken string) (*User, error) {
	query := `
		INSERT INTO users (github_id, username, name, avatar_url, encrypted_token)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, github_id, username, name, avatar_url, encrypted_token, created_at, updated_at
	`
	row := q.pool.QueryRow(ctx, query, githubID, username, name, avatarURL, encryptedToken)
	var u User
	err := row.Scan(&u.ID, &u.GithubID, &u.Username, &u.Name, &u.AvatarURL, &u.EncryptedToken, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (q *Queries) UpdateUserToken(ctx context.Context, id uuid.UUID, encryptedToken string) error {
	query := `UPDATE users SET encrypted_token = $1, updated_at = NOW() WHERE id = $2`
	_, err := q.pool.Exec(ctx, query, encryptedToken, id)
	return err
}

type Queries struct {
	pool *pgxpool.Pool
}

func NewQueries(pool *pgxpool.Pool) *Queries {
	return &Queries{pool: pool}
}

func (q *Queries) CreateWorkspace(ctx context.Context, ownerID uuid.UUID, gitURL, gitBranch, name string) (*Workspace, error) {
	query := `
		INSERT INTO workspaces (owner_id, git_url, git_branch, name)
		VALUES ($1, $2, $3, $4)
		RETURNING id, owner_id, git_url, git_branch, name, created_at, updated_at
	`
	row := q.pool.QueryRow(ctx, query, ownerID, gitURL, gitBranch, name)
	var w Workspace
	err := row.Scan(&w.ID, &w.OwnerID, &w.GitURL, &w.GitBranch, &w.Name, &w.CreatedAt, &w.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &w, nil
}

func (q *Queries) GetWorkspaceByID(ctx context.Context, id uuid.UUID) (*Workspace, error) {
	query := `SELECT id, owner_id, git_url, git_branch, name, created_at, updated_at FROM workspaces WHERE id = $1`
	row := q.pool.QueryRow(ctx, query, id)
	var w Workspace
	err := row.Scan(&w.ID, &w.OwnerID, &w.GitURL, &w.GitBranch, &w.Name, &w.CreatedAt, &w.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &w, nil
}

func (q *Queries) CreateEnvironment(ctx context.Context, workspaceID, hostID uuid.UUID) (*Environment, error) {
	query := `
		INSERT INTO environments (workspace_id, host_id, status)
		VALUES ($1, $2, 'CREATED')
		RETURNING id, workspace_id, host_id, status, container_id, port, public_url, last_activity_at, created_at, updated_at, deleted_at
	`
	row := q.pool.QueryRow(ctx, query, workspaceID, hostID)
	var e Environment
	err := row.Scan(&e.ID, &e.WorkspaceID, &e.HostID, &e.Status, &e.ContainerID, &e.Port, &e.PublicURL, &e.LastActivityAt, &e.CreatedAt, &e.UpdatedAt, &e.DeletedAt)
	if err != nil {
		return nil, err
	}
	return &e, nil
}

func (q *Queries) AddEnvironmentMember(ctx context.Context, envID, userID uuid.UUID, role string, invitedBy *uuid.UUID) (*EnvironmentMember, error) {
	query := `
		INSERT INTO environment_members (environment_id, user_id, role, invited_by)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (environment_id, user_id) DO UPDATE SET role = EXCLUDED.role
		RETURNING id, environment_id, user_id, role, invited_by, created_at
	`
	row := q.pool.QueryRow(ctx, query, envID, userID, role, invitedBy)
	var m EnvironmentMember
	err := row.Scan(&m.ID, &m.EnvironmentID, &m.UserID, &m.Role, &m.InvitedBy, &m.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (q *Queries) GetEnvironmentByID(ctx context.Context, id uuid.UUID) (*Environment, error) {
	query := `SELECT id, workspace_id, host_id, status, container_id, port, public_url, last_activity_at, created_at, updated_at, deleted_at FROM environments WHERE id = $1 AND deleted_at IS NULL`
	row := q.pool.QueryRow(ctx, query, id)
	var e Environment
	err := row.Scan(&e.ID, &e.WorkspaceID, &e.HostID, &e.Status, &e.ContainerID, &e.Port, &e.PublicURL, &e.LastActivityAt, &e.CreatedAt, &e.UpdatedAt, &e.DeletedAt)
	if err != nil {
		return nil, err
	}
	return &e, nil
}

func (q *Queries) GetEnvironmentsForUser(ctx context.Context, userID uuid.UUID) ([]EnvironmentWithRole, error) {
	query := `
		SELECT e.id, e.workspace_id, e.host_id, e.status, e.container_id, e.port, e.public_url, e.last_activity_at, e.created_at, e.updated_at, e.deleted_at,
		       em.role
		FROM environments e
		JOIN environment_members em ON em.environment_id = e.id
		WHERE em.user_id = $1 AND e.deleted_at IS NULL
		ORDER BY e.updated_at DESC
	`
	rows, err := q.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var envs []EnvironmentWithRole
	for rows.Next() {
		var e EnvironmentWithRole
		err := rows.Scan(&e.ID, &e.WorkspaceID, &e.HostID, &e.Status, &e.ContainerID, &e.Port, &e.PublicURL, &e.LastActivityAt, &e.CreatedAt, &e.UpdatedAt, &e.DeletedAt, &e.Role)
		if err != nil {
			return nil, err
		}
		envs = append(envs, e)
	}
	return envs, rows.Err()
}

type EnvironmentWithRole struct {
	Environment
	Role string
}

func (q *Queries) GetEnvironmentMembers(ctx context.Context, envID uuid.UUID) ([]MemberWithUser, error) {
	query := `
		SELECT em.id, em.environment_id, em.user_id, em.role, em.invited_by, em.created_at,
		       u.id, u.github_id, u.username, u.name, u.avatar_url, u.encrypted_token, u.created_at, u.updated_at
		FROM environment_members em
		JOIN users u ON u.id = em.user_id
		WHERE em.environment_id = $1
		ORDER BY 
			CASE em.role WHEN 'OWNER' THEN 1 WHEN 'COLLABORATOR' THEN 2 WHEN 'VIEWER' THEN 3 ELSE 4 END,
			em.created_at
	`
	rows, err := q.pool.Query(ctx, query, envID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var members []MemberWithUser
	for rows.Next() {
		var m MemberWithUser
		err := rows.Scan(
			&m.ID, &m.EnvironmentID, &m.UserID, &m.Role, &m.InvitedBy, &m.CreatedAt,
			&m.User.ID, &m.User.GithubID, &m.User.Username, &m.User.Name, &m.User.AvatarURL, &m.User.EncryptedToken, &m.User.CreatedAt, &m.User.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		members = append(members, m)
	}
	return members, rows.Err()
}

func (q *Queries) GetUserRoleInEnvironment(ctx context.Context, envID, userID uuid.UUID) (string, error) {
	query := `SELECT role FROM environment_members WHERE environment_id = $1 AND user_id = $2`
	var role string
	err := q.pool.QueryRow(ctx, query, envID, userID).Scan(&role)
	if err != nil {
		return "", err
	}
	return role, nil
}

func (q *Queries) UpdateMemberRole(ctx context.Context, envID, userID uuid.UUID, role string) error {
	query := `UPDATE environment_members SET role = $1 WHERE environment_id = $2 AND user_id = $3`
	result, err := q.pool.Exec(ctx, query, role, envID, userID)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (q *Queries) RemoveMember(ctx context.Context, envID, userID uuid.UUID) error {
	query := `DELETE FROM environment_members WHERE environment_id = $1 AND user_id = $2`
	result, err := q.pool.Exec(ctx, query, envID, userID)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (q *Queries) CreateShareLink(ctx context.Context, envID, createdBy uuid.UUID, token, role string, expiresAt *time.Time, maxUses *int) (*ShareLink, error) {
	query := `
		INSERT INTO share_links (environment_id, created_by, token, role, expires_at, max_uses)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, environment_id, created_by, token, role, expires_at, max_uses, use_count, created_at
	`
	row := q.pool.QueryRow(ctx, query, envID, createdBy, token, role, expiresAt, maxUses)
	var s ShareLink
	err := row.Scan(&s.ID, &s.EnvironmentID, &s.CreatedBy, &s.Token, &s.Role, &s.ExpiresAt, &s.MaxUses, &s.UseCount, &s.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (q *Queries) GetShareLinkByToken(ctx context.Context, token string) (*ShareLink, error) {
	query := `SELECT id, environment_id, created_by, token, role, expires_at, max_uses, use_count, created_at FROM share_links WHERE token = $1`
	row := q.pool.QueryRow(ctx, query, token)
	var s ShareLink
	err := row.Scan(&s.ID, &s.EnvironmentID, &s.CreatedBy, &s.Token, &s.Role, &s.ExpiresAt, &s.MaxUses, &s.UseCount, &s.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (q *Queries) IncrementShareLinkUseCount(ctx context.Context, token string) error {
	query := `UPDATE share_links SET use_count = use_count + 1 WHERE token = $1`
	_, err := q.pool.Exec(ctx, query, token)
	return err
}

func (q *Queries) UpdateEnvironmentStatus(ctx context.Context, envID uuid.UUID, status string) error {
	query := `UPDATE environments SET status = $1, updated_at = NOW() WHERE id = $2`
	_, err := q.pool.Exec(ctx, query, status, envID)
	return err
}

func (q *Queries) UpdateEnvironmentContainer(ctx context.Context, envID uuid.UUID, containerID string, port int, publicURL string) error {
	query := `UPDATE environments SET container_id = $1, port = $2, public_url = $3, status = 'RUNNING', updated_at = NOW() WHERE id = $4`
	_, err := q.pool.Exec(ctx, query, containerID, port, publicURL, envID)
	return err
}

func (q *Queries) UpdateEnvironmentLastActivity(ctx context.Context, envID uuid.UUID) error {
	query := `UPDATE environments SET last_activity_at = NOW() WHERE id = $1`
	_, err := q.pool.Exec(ctx, query, envID)
	return err
}

func (q *Queries) SoftDeleteEnvironment(ctx context.Context, envID uuid.UUID) error {
	query := `UPDATE environments SET deleted_at = NOW(), status = 'STOPPED', updated_at = NOW() WHERE id = $1`
	_, err := q.pool.Exec(ctx, query, envID)
	return err
}

func (q *Queries) GetUserEnvironmentsWithHost(ctx context.Context, userID uuid.UUID) ([]EnvironmentWithHost, error) {
	query := `
		SELECT e.id, e.workspace_id, e.host_id, e.status, e.container_id, e.port, e.public_url, e.last_activity_at, e.created_at, e.updated_at, e.deleted_at,
		       em.role,
		       u.id, u.username, u.avatar_url,
		       w.git_url, w.git_branch, w.name
		FROM environments e
		JOIN environment_members em ON em.environment_id = e.id
		JOIN users u ON u.id = e.host_id
		LEFT JOIN workspaces w ON w.id = e.workspace_id
		WHERE em.user_id = $1 AND e.deleted_at IS NULL
		ORDER BY e.updated_at DESC
	`
	rows, err := q.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var envs []EnvironmentWithHost
	for rows.Next() {
		var e EnvironmentWithHost
		err := rows.Scan(
			&e.ID, &e.WorkspaceID, &e.HostID, &e.Status, &e.ContainerID, &e.Port, &e.PublicURL, &e.LastActivityAt, &e.CreatedAt, &e.UpdatedAt, &e.DeletedAt,
			&e.Role,
			&e.Host.ID, &e.Host.Username, &e.Host.AvatarURL,
			&e.WorkspaceGitURL, &e.WorkspaceGitBranch, &e.WorkspaceName,
		)
		if err != nil {
			return nil, err
		}
		envs = append(envs, e)
	}
	return envs, rows.Err()
}

type EnvironmentWithHost struct {
	Environment
	Role                string
	Host                UserInfo
	WorkspaceGitURL     string
	WorkspaceGitBranch  string
	WorkspaceName       string
}

type UserInfo struct {
	ID        uuid.UUID
	Username  string
	AvatarURL *string
}

// ============================================
// BUILD LOGS
// ============================================
type BuildLog struct {
	ID            uuid.UUID
	EnvironmentID uuid.UUID
	UserID        *uuid.UUID
	Sequence      int
	Message       string
	Level         string
	CreatedAt     time.Time
}

type BuildLogLevel string

const (
	BuildLogLevelInfo    BuildLogLevel = "info"
	BuildLogLevelWarn    BuildLogLevel = "warn"
	BuildLogLevelError   BuildLogLevel = "error"
	BuildLogLevelSuccess BuildLogLevel = "success"
)

func (q *Queries) AddBuildLog(ctx context.Context, environmentID uuid.UUID, userID *uuid.UUID, message string, level BuildLogLevel) error {
	query := `
		INSERT INTO build_logs (environment_id, user_id, sequence, message, level)
		SELECT $1, $2, COALESCE(MAX(sequence), 0) + 1, $3, $4
		FROM build_logs
		WHERE environment_id = $1
	`
	_, err := q.pool.Exec(ctx, query, environmentID, userID, string(level), string(level))
	return err
}

func (q *Queries) GetBuildLogs(ctx context.Context, environmentID uuid.UUID, limit, offset int) ([]BuildLog, error) {
	query := `
		SELECT id, environment_id, user_id, sequence, message, level, created_at
		FROM build_logs
		WHERE environment_id = $1
		ORDER BY sequence ASC
		LIMIT $2 OFFSET $3
	`
	rows, err := q.pool.Query(ctx, query, environmentID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []BuildLog
	for rows.Next() {
		var log BuildLog
		var userID *uuid.UUID
		err := rows.Scan(&log.ID, &log.EnvironmentID, &userID, &log.Sequence, &log.Message, &log.Level, &log.CreatedAt)
		if err != nil {
			return nil, err
		}
		log.UserID = userID
		logs = append(logs, log)
	}
	return logs, rows.Err()
}

func (q *Queries) GetBuildLogCount(ctx context.Context, environmentID uuid.UUID) (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM build_logs WHERE environment_id = $1`
	err := q.pool.QueryRow(ctx, query, environmentID).Scan(&count)
	return count, err
}

func (q *Queries) ClearBuildLogs(ctx context.Context, environmentID uuid.UUID) error {
	query := `DELETE FROM build_logs WHERE environment_id = $1`
	_, err := q.pool.Exec(ctx, query, environmentID)
	return err
}