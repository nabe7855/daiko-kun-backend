-- Seed initial data for testing
INSERT INTO companies (id, name, status, commission_rate)
VALUES ('00000000-0000-0000-0000-000000000001', '栃木代行サービス', 'active', 10.0);

INSERT INTO admin_users (id, company_id, username, password_hash, role, name)
VALUES (gen_random_uuid(), '00000000-0000-0000-0000-000000000001', 'company', 'password123', 'company_admin', '栃木代行 社長');
