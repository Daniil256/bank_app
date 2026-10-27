-- +goose Up
INSERT INTO users (email, password_hash, role)
VALUES (
           'admin@bank.com',
--         12345678
           '$argon2id$v=19$m=65536,t=1,p=4$Bz52ZJXN5atPXdN/ZVQsfg$wev9mcBPdsEEkP+mwY7ytf/SeYS/Tp0UyXGIN3gVaMw',
           'ADMIN'
       )
ON CONFLICT (email) DO NOTHING;

-- +goose Down
DELETE FROM users WHERE email = 'admin@bank.com' AND role = 'ADMIN';