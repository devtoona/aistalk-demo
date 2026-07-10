-- PostgreSQL 初期化スクリプト
-- 拡張などが必要な場合にここに記述

-- タイムゾーンを日本時間に設定（db_nameはDB_NAME と一致させること）
\set db_name aistalkdb

-- タイムゾーンを日本時間に設定
ALTER DATABASE :db_name SET timezone TO 'Asia/Tokyo';
