ALTER TABLE "users" DROP COLUMN "password_hash";
ALTER TABLE "users" ADD COLUMN "password" varchar(50) NOT NULL;
