package repository

import (
	"database/sql"
	"fmt"
)

type MigrationRepository interface {
	CreateUserTable() error
	CreateDeletedUserTable() error
	CreateStoriesTable() error
	CreateMessageTable() error
	CreateGroupTable() error
	CreateUserGroupRelationTable() error
	CreateGroupMessageTable() error
	CreateAdminTable() error
}

type repository struct {
	db *sql.DB
}

func NewMigrationRepository(db *sql.DB) MigrationRepository {
	return &repository{db}
}

func (r *repository) CreateUserTable() error {
	_, err := r.db.Exec(`CREATE TABLE IF NOT EXISTS users(user_ph_no TEXT PRIMARY KEY NOT NULL, user_name TEXT NOT NULL, user_avatar TEXT, user_about TEXT NOT NULL, user_password TEXT NOT NULL, is_blocked BOOLEAN NOT NULL, blocked_duration TEXT, block_list TEXT);`)
	return err
}

func (r *repository) CreateDeletedUserTable() error {
	_, err := r.db.Exec(`CREATE TABLE IF NOT EXISTS deleted_users(user_ph_no TEXT PRIMARY KEY NOT NULL, user_avatar TEXT, user_about TEXT NOT NULL, is_blocked BOOLEAN NOT NULL, blocked_duration TEXT, block_list TEXT, delete_time TEXT);`)
	return err
}

func (r *repository) CreateStoriesTable() error {
	_, err := r.db.Exec(`CREATE TABLE IF NOT EXISTS stories(user_id TEXT NOT NULL, story_url TEXT NOT NULL, story_update_time TEXT NOT NULL, viewers TEXT NOT NULL, is_active TEXT NOT NULL);`)
	return err
}

func (r *repository) CreateMessageTable() error {
	_, err := r.db.Exec(`CREATE TABLE IF NOT EXISTS messages(msg_id BIGSERIAL PRIMARY KEY NOT NULL, from_user_id TEXT NOT NULL, to_user_id TEXT NOT NULL, content TEXT NOT NULL, content_type TEXT NOT NULL, sent_time TEXT NOT NULL, status TEXT NOT NULL, is_recent BOOLEAN NOT NULL);`)
	return err
}

func (r *repository) CreateGroupTable() error {
	_, err := r.db.Exec(`CREATE TABLE IF NOT EXISTS groups(group_id BIGSERIAL PRIMARY KEY NOT NULL, group_name TEXT NOT NULL, group_avatar TEXT NOT NULL, group_about TEXT NOT NULL, group_creator TEXT NOT NULL, group_created_date TEXT NOT NULL, group_total_members BIGINT NOT NULL, is_banned BOOLEAN NOT NULL, banned_time TEXT);`)
	return err
}

func (r *repository) CreateUserGroupRelationTable() error {
	_, err := r.db.Exec(`CREATE TABLE IF NOT EXISTS user_group_relations(id BIGSERIAL PRIMARY KEY NOT NULL, group_id BIGINT NOT NULL, user_id TEXT NOT NULL, user_role TEXT NOT NULL);`)
	return err
}

func (r *repository) CreateGroupMessageTable() error {
	_, err := r.db.Exec(`CREATE TABLE IF NOT EXISTS group_messages(msg_id BIGSERIAL PRIMARY KEY NOT NULL, group_id BIGINT NOT NULL, sender_id TEXT NOT NULL, message_content TEXT NOT NULL, sent_time TEXT NOT NULL, is_recent BOOLEAN NOT NULL, status TEXT NOT NULL, content_type TEXT);`)
	return err
}

func (r *repository) CreateAdminTable() error {
	_, err := r.db.Exec(`CREATE TABLE IF NOT EXISTS admins(admin_id BIGSERIAL PRIMARY KEY NOT NULL, admin_name TEXT DEFAULT 'root' NOT NULL, admin_pass TEXT DEFAULT 'toor' NOT NULL);`)
	if err != nil {
		return err
	}

	fmt.Println("Admin table created successfully!")

	var count int
	err = r.db.QueryRow("SELECT COUNT(*) FROM admins;").Scan(&count)
	if err != nil {
		return err
	}

	if count == 0 {
		_, err := r.db.Exec(`INSERT INTO admins(admin_name, admin_pass) VALUES(DEFAULT, DEFAULT);`)
		if err != nil {
			return err
		}
	}

	return nil
}
