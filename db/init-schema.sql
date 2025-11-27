-- ========================================
-- Kickoff NFL Database Schema
-- ========================================
-- Description: Database schema for NFL prediction game microservices
-- Author: Diego + Claude
-- Date: 2025-11-26
-- ========================================

-- ========================================
-- EXTENSIONS
-- ========================================
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- ========================================
-- TABLES
-- ========================================

-- Users table
CREATE TABLE IF NOT EXISTS users (
    id VARCHAR(255) PRIMARY KEY,
    username VARCHAR(100) UNIQUE NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    full_name VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    active BOOLEAN DEFAULT true,
    CONSTRAINT users_username_not_empty CHECK (username <> ''),
    CONSTRAINT users_email_not_empty CHECK (email <> ''),
    CONSTRAINT users_email_format CHECK (email ~* '^[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}$')
);

-- Teams table
CREATE TABLE IF NOT EXISTS teams (
    id VARCHAR(255) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    city VARCHAR(255) NOT NULL,
    abbreviation VARCHAR(10) UNIQUE NOT NULL,
    conference VARCHAR(50) NOT NULL,
    division VARCHAR(50) NOT NULL,
    logo_url TEXT,
    stadium VARCHAR(255),
    CONSTRAINT teams_conference_valid CHECK (conference IN ('AFC', 'NFC')),
    CONSTRAINT teams_division_valid CHECK (division IN (
        'AFC_EAST', 'AFC_NORTH', 'AFC_SOUTH', 'AFC_WEST',
        'NFC_EAST', 'NFC_NORTH', 'NFC_SOUTH', 'NFC_WEST'
    ))
);

-- Games table
CREATE TABLE IF NOT EXISTS games (
    id VARCHAR(255) PRIMARY KEY,
    home_team_id VARCHAR(255) NOT NULL REFERENCES teams(id) ON DELETE RESTRICT,
    away_team_id VARCHAR(255) NOT NULL REFERENCES teams(id) ON DELETE RESTRICT,
    week INT NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'SCHEDULED',
    home_score INT DEFAULT 0,
    away_score INT DEFAULT 0,
    scheduled_at TIMESTAMP WITH TIME ZONE NOT NULL,
    started_at TIMESTAMP WITH TIME ZONE,
    completed_at TIMESTAMP WITH TIME ZONE,
    CONSTRAINT games_week_valid CHECK (week >= 1 AND week <= 18),
    CONSTRAINT games_status_valid CHECK (status IN ('SCHEDULED', 'IN_PROGRESS', 'COMPLETED', 'POSTPONED', 'CANCELED')),
    CONSTRAINT games_scores_valid CHECK (home_score >= 0 AND away_score >= 0),
    CONSTRAINT games_different_teams CHECK (home_team_id <> away_team_id)
);

-- Predictions table
CREATE TABLE IF NOT EXISTS predictions (
    id VARCHAR(255) PRIMARY KEY,
    user_id VARCHAR(255) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    game_id VARCHAR(255) NOT NULL REFERENCES games(id) ON DELETE CASCADE,
    predicted_winner_id VARCHAR(255) NOT NULL REFERENCES teams(id) ON DELETE RESTRICT,
    status VARCHAR(50) NOT NULL DEFAULT 'PENDING',
    points INT DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT predictions_status_valid CHECK (status IN ('PENDING', 'CORRECT', 'INCORRECT', 'VOID')),
    CONSTRAINT predictions_points_valid CHECK (points >= 0),
    CONSTRAINT predictions_unique_user_game UNIQUE (user_id, game_id)
);

-- ========================================
-- INDEXES
-- ========================================

-- Users indexes
CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_active ON users(active);
CREATE INDEX IF NOT EXISTS idx_users_created_at ON users(created_at DESC);

-- Teams indexes
CREATE INDEX IF NOT EXISTS idx_teams_abbreviation ON teams(abbreviation);
CREATE INDEX IF NOT EXISTS idx_teams_conference ON teams(conference);
CREATE INDEX IF NOT EXISTS idx_teams_division ON teams(division);

-- Games indexes
CREATE INDEX IF NOT EXISTS idx_games_week ON games(week);
CREATE INDEX IF NOT EXISTS idx_games_status ON games(status);
CREATE INDEX IF NOT EXISTS idx_games_scheduled_at ON games(scheduled_at);
CREATE INDEX IF NOT EXISTS idx_games_home_team ON games(home_team_id);
CREATE INDEX IF NOT EXISTS idx_games_away_team ON games(away_team_id);
CREATE INDEX IF NOT EXISTS idx_games_completed_at ON games(completed_at DESC);

-- Predictions indexes
CREATE INDEX IF NOT EXISTS idx_predictions_user_id ON predictions(user_id);
CREATE INDEX IF NOT EXISTS idx_predictions_game_id ON predictions(game_id);
CREATE INDEX IF NOT EXISTS idx_predictions_status ON predictions(status);
CREATE INDEX IF NOT EXISTS idx_predictions_created_at ON predictions(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_predictions_user_status ON predictions(user_id, status);

-- ========================================
-- SEED DATA - NFL Teams
-- ========================================

-- AFC East
INSERT INTO teams (id, name, city, abbreviation, conference, division, stadium) VALUES
    ('BUF', 'Bills', 'Buffalo', 'BUF', 'AFC', 'AFC_EAST', 'Highmark Stadium'),
    ('MIA', 'Dolphins', 'Miami', 'MIA', 'AFC', 'AFC_EAST', 'Hard Rock Stadium'),
    ('NE', 'Patriots', 'New England', 'NE', 'AFC', 'AFC_EAST', 'Gillette Stadium'),
    ('NYJ', 'Jets', 'New York', 'NYJ', 'AFC', 'AFC_EAST', 'MetLife Stadium')
ON CONFLICT (id) DO NOTHING;

-- AFC North
INSERT INTO teams (id, name, city, abbreviation, conference, division, stadium) VALUES
    ('BAL', 'Ravens', 'Baltimore', 'BAL', 'AFC', 'AFC_NORTH', 'M&T Bank Stadium'),
    ('CIN', 'Bengals', 'Cincinnati', 'CIN', 'AFC', 'AFC_NORTH', 'Paycor Stadium'),
    ('CLE', 'Browns', 'Cleveland', 'CLE', 'AFC', 'AFC_NORTH', 'Cleveland Browns Stadium'),
    ('PIT', 'Steelers', 'Pittsburgh', 'PIT', 'AFC', 'AFC_NORTH', 'Acrisure Stadium')
ON CONFLICT (id) DO NOTHING;

-- AFC South
INSERT INTO teams (id, name, city, abbreviation, conference, division, stadium) VALUES
    ('HOU', 'Texans', 'Houston', 'HOU', 'AFC', 'AFC_SOUTH', 'NRG Stadium'),
    ('IND', 'Colts', 'Indianapolis', 'IND', 'AFC', 'AFC_SOUTH', 'Lucas Oil Stadium'),
    ('JAX', 'Jaguars', 'Jacksonville', 'JAX', 'AFC', 'AFC_SOUTH', 'TIAA Bank Field'),
    ('TEN', 'Titans', 'Tennessee', 'TEN', 'AFC', 'AFC_SOUTH', 'Nissan Stadium')
ON CONFLICT (id) DO NOTHING;

-- AFC West
INSERT INTO teams (id, name, city, abbreviation, conference, division, stadium) VALUES
    ('DEN', 'Broncos', 'Denver', 'DEN', 'AFC', 'AFC_WEST', 'Empower Field at Mile High'),
    ('KC', 'Chiefs', 'Kansas City', 'KC', 'AFC', 'AFC_WEST', 'Arrowhead Stadium'),
    ('LV', 'Raiders', 'Las Vegas', 'LV', 'AFC', 'AFC_WEST', 'Allegiant Stadium'),
    ('LAC', 'Chargers', 'Los Angeles', 'LAC', 'AFC', 'AFC_WEST', 'SoFi Stadium')
ON CONFLICT (id) DO NOTHING;

-- NFC East
INSERT INTO teams (id, name, city, abbreviation, conference, division, stadium) VALUES
    ('DAL', 'Cowboys', 'Dallas', 'DAL', 'NFC', 'NFC_EAST', 'AT&T Stadium'),
    ('NYG', 'Giants', 'New York', 'NYG', 'NFC', 'NFC_EAST', 'MetLife Stadium'),
    ('PHI', 'Eagles', 'Philadelphia', 'PHI', 'NFC', 'NFC_EAST', 'Lincoln Financial Field'),
    ('WAS', 'Commanders', 'Washington', 'WAS', 'NFC', 'NFC_EAST', 'FedExField')
ON CONFLICT (id) DO NOTHING;

-- NFC North
INSERT INTO teams (id, name, city, abbreviation, conference, division, stadium) VALUES
    ('CHI', 'Bears', 'Chicago', 'CHI', 'NFC', 'NFC_NORTH', 'Soldier Field'),
    ('DET', 'Lions', 'Detroit', 'DET', 'NFC', 'NFC_NORTH', 'Ford Field'),
    ('GB', 'Packers', 'Green Bay', 'GB', 'NFC', 'NFC_NORTH', 'Lambeau Field'),
    ('MIN', 'Vikings', 'Minnesota', 'MIN', 'NFC', 'NFC_NORTH', 'U.S. Bank Stadium')
ON CONFLICT (id) DO NOTHING;

-- NFC South
INSERT INTO teams (id, name, city, abbreviation, conference, division, stadium) VALUES
    ('ATL', 'Falcons', 'Atlanta', 'ATL', 'NFC', 'NFC_SOUTH', 'Mercedes-Benz Stadium'),
    ('CAR', 'Panthers', 'Carolina', 'CAR', 'NFC', 'NFC_SOUTH', 'Bank of America Stadium'),
    ('NO', 'Saints', 'New Orleans', 'NO', 'NFC', 'NFC_SOUTH', 'Caesars Superdome'),
    ('TB', 'Buccaneers', 'Tampa Bay', 'TB', 'NFC', 'NFC_SOUTH', 'Raymond James Stadium')
ON CONFLICT (id) DO NOTHING;

-- NFC West
INSERT INTO teams (id, name, city, abbreviation, conference, division, stadium) VALUES
    ('ARI', 'Cardinals', 'Arizona', 'ARI', 'NFC', 'NFC_WEST', 'State Farm Stadium'),
    ('LAR', 'Rams', 'Los Angeles', 'LAR', 'NFC', 'NFC_WEST', 'SoFi Stadium'),
    ('SF', '49ers', 'San Francisco', 'SF', 'NFC', 'NFC_WEST', 'Levi''s Stadium'),
    ('SEA', 'Seahawks', 'Seattle', 'SEA', 'NFC', 'NFC_WEST', 'Lumen Field')
ON CONFLICT (id) DO NOTHING;

-- ========================================
-- SEED DATA - Sample Users (for testing)
-- ========================================

INSERT INTO users (id, username, email, full_name) VALUES
    ('user_1', 'john_doe', 'john.doe@example.com', 'John Doe'),
    ('user_2', 'jane_smith', 'jane.smith@example.com', 'Jane Smith'),
    ('user_3', 'bob_wilson', 'bob.wilson@example.com', 'Bob Wilson')
ON CONFLICT (id) DO NOTHING;

-- ========================================
-- SEED DATA - Sample Games (Week 1)
-- ========================================

INSERT INTO games (id, home_team_id, away_team_id, week, status, scheduled_at) VALUES
    ('game_1', 'KC', 'BUF', 1, 'SCHEDULED', CURRENT_TIMESTAMP + INTERVAL '1 week'),
    ('game_2', 'SF', 'DAL', 1, 'SCHEDULED', CURRENT_TIMESTAMP + INTERVAL '1 week'),
    ('game_3', 'PHI', 'NYG', 1, 'SCHEDULED', CURRENT_TIMESTAMP + INTERVAL '1 week')
ON CONFLICT (id) DO NOTHING;

-- ========================================
-- FUNCTIONS & TRIGGERS
-- ========================================

-- Trigger para actualizar updated_at automáticamente
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Aplicar trigger a predictions
DROP TRIGGER IF EXISTS update_predictions_updated_at ON predictions;
CREATE TRIGGER update_predictions_updated_at
    BEFORE UPDATE ON predictions
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- ========================================
-- VIEWS (para reporting)
-- ========================================

-- Vista de leaderboard
CREATE OR REPLACE VIEW leaderboard_view AS
SELECT
    u.id AS user_id,
    u.username,
    u.full_name,
    COUNT(p.id) AS total_predictions,
    COUNT(p.id) FILTER (WHERE p.status = 'CORRECT') AS correct_predictions,
    COUNT(p.id) FILTER (WHERE p.status = 'INCORRECT') AS incorrect_predictions,
    COALESCE(SUM(p.points), 0) AS total_points,
    CASE
        WHEN COUNT(p.id) FILTER (WHERE p.status IN ('CORRECT', 'INCORRECT')) > 0
        THEN ROUND(
            (COUNT(p.id) FILTER (WHERE p.status = 'CORRECT')::DECIMAL /
             COUNT(p.id) FILTER (WHERE p.status IN ('CORRECT', 'INCORRECT'))::DECIMAL) * 100,
            2
        )
        ELSE 0
    END AS accuracy_percentage
FROM users u
LEFT JOIN predictions p ON u.id = p.user_id
WHERE u.active = true
GROUP BY u.id, u.username, u.full_name
ORDER BY total_points DESC, correct_predictions DESC;

-- ========================================
-- GRANTS (para el usuario de la aplicación)
-- ========================================

GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public TO kickoff_user;
GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA public TO kickoff_user;
GRANT SELECT ON leaderboard_view TO kickoff_user;

-- ========================================
-- FINAL MESSAGE
-- ========================================

-- Verificar que todo se creó correctamente
DO $$
DECLARE
    table_count INTEGER;
    index_count INTEGER;
    team_count INTEGER;
BEGIN
    SELECT COUNT(*) INTO table_count FROM pg_tables WHERE schemaname = 'public';
    SELECT COUNT(*) INTO index_count FROM pg_indexes WHERE schemaname = 'public';
    SELECT COUNT(*) INTO team_count FROM teams;

    RAISE NOTICE '========================================';
    RAISE NOTICE 'Database initialized successfully!';
    RAISE NOTICE '========================================';
    RAISE NOTICE 'Tables created: %', table_count;
    RAISE NOTICE 'Indexes created: %', index_count;
    RAISE NOTICE 'NFL Teams seeded: %', team_count;
    RAISE NOTICE '========================================';
END $$;
