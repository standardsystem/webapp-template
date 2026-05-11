-- 依存関係順に削除（user_providers が users を参照しているため先に drop）
DROP TABLE IF EXISTS user_providers;
DROP TABLE IF EXISTS users;
