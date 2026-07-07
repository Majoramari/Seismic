-- 005_add_monthly_goal_period.sql
ALTER TABLE goals DROP CONSTRAINT goals_period_check;
ALTER TABLE goals
    ADD CONSTRAINT goals_period_check CHECK (period IN ('daily', 'weekly', 'monthly'));