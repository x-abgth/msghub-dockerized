package logic

import "github.com/x-abgth/msghub-dockerized/msghub-server/repository"

type MigrationLogic interface {
	MigrateUserTable() error
	MigrateDeletedUserTable() error
	MigrateStoriesTable() error
	MigrateMessageTable() error
	MigrateGroupTable() error
	MigrateUserGroupRelationTable() error
	MigrateGroupMessageTable() error
	MigrateAdminTable() error
}

type migrationLogic struct {
	migrationRepository repository.MigrationRepository
}

func NewMigrateLogic(migrationRepository repository.MigrationRepository) MigrationLogic {
	return &migrationLogic{migrationRepository: migrationRepository}
}

func (m *migrationLogic) MigrateUserTable() error {
	return m.migrationRepository.CreateUserTable()
}

func (m *migrationLogic) MigrateDeletedUserTable() error {
	return m.migrationRepository.CreateDeletedUserTable()
}

func (m *migrationLogic) MigrateStoriesTable() error {
	return m.migrationRepository.CreateStoriesTable()
}

func (m *migrationLogic) MigrateMessageTable() error {
	return m.migrationRepository.CreateMessageTable()
}

func (m *migrationLogic) MigrateGroupTable() error {
	return m.migrationRepository.CreateGroupTable()
}

func (m *migrationLogic) MigrateUserGroupRelationTable() error {
	return m.migrationRepository.CreateUserGroupRelationTable()
}

func (m *migrationLogic) MigrateGroupMessageTable() error {
	return m.migrationRepository.CreateGroupMessageTable()
}

func (m *migrationLogic) MigrateAdminTable() error {
	return m.migrationRepository.CreateAdminTable()
}
