-- +goose Up
-- Вставляем администратора, если его еще нет.
-- Пароль нужно будет сгенерировать или задать через переменную окружения.
INSERT INTO users (email, password_hash, role)
VALUES (
           'admin@bank.com',
--         123456
           '$argon2id$v=19$m=65536,t=3,p=2$bG5LdzNNenhQZW0wejN4SnR2ZHJmT0FwUU55T3hUejc$1qvyedS7IKCNU8frLiF54kI0VVv1bwrtuyjvKp4krjU',
           'ADMIN'
       )
ON CONFLICT (email) DO NOTHING;

-- +goose Down
DELETE FROM users WHERE email = 'admin@bank.com' AND role = 'ADMIN';