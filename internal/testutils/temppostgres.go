// Package testutils
package testutils

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"

	"github.com/jackc/pgx/v5"
)

type TempPostgresInstance struct {
	Connection *pgx.Conn
	dataDir    string
	logger     *slog.Logger
	dbname     string
	user       string
	port       int
}

func NewTempPostgresInstance(logger *slog.Logger, dbname string, user string, port int) *TempPostgresInstance {
	dataDir, err := os.MkdirTemp("", "pgtest")
	if err != nil {
		panic(err)
	}

	return &TempPostgresInstance{
		dataDir: dataDir,
		logger:  logger,
		dbname:  dbname,
		user:    user,
		port:    port,
	}
}

func (pg *TempPostgresInstance) Setup() error {
	if pg.isRunning() {
		_, err := pg.connect()
		return err
	}

	if !pg.isInitialized() {
		if err := pg.initialize(); err != nil {
			pg.logger.Error("failed to initialize database", "error", err)
			return err
		}
	}

	if err := pg.start(); err != nil {
		pg.logger.Error("failed to start database", "error", err)
		return err
	}

	if err := pg.createDatabase(); err != nil {
		pg.logger.Error("failed to create database", "error", err)
		return err
	}

	if _, err := pg.connect(); err != nil {
		pg.logger.Error("failed to connect to database", "error", err)
		return err
	}

	return nil
}

func (pg *TempPostgresInstance) Cleanup() error {
	if err := pg.close(); err != nil {
		pg.logger.Error("failed to close connection to database", "error", err)
		return err
	}

	if pg.isRunning() {
		if err := pg.stop(); err != nil {
			pg.logger.Error("failed to stop database", "error", err)
			return err
		}
	}

	if _, err := os.Stat(pg.dataDir); err == nil {
		if err := os.RemoveAll(pg.dataDir); err != nil {
			pg.logger.Error("failed to remove tmp directory", "dataDir", pg.dataDir, "error", err)
			return err
		}
	}

	return nil
}

func (pg *TempPostgresInstance) isRunning() bool {
	cmd := exec.Command("pg_ctl", "-D", pg.dataDir, "status")
	err := cmd.Run()
	return err == nil
}

func (pg *TempPostgresInstance) connect() (*pgx.Conn, error) {
	connStr := fmt.Sprintf(
		"postgres://%s@localhost:%d/%s?sslmode=disable",
		pg.user,
		pg.port,
		pg.dbname,
	)

	conn, err := pgx.Connect(context.Background(), connStr)
	pg.Connection = conn
	return conn, err
}

func (pg *TempPostgresInstance) close() error {
	if pg.Connection != nil {
		if err := pg.Connection.Close(context.Background()); err != nil {
			return err
		}

		pg.Connection = nil
	}

	return nil
}

func (pg *TempPostgresInstance) isInitialized() bool {
	path := filepath.Join(pg.dataDir, "PG_VERSION")
	_, err := os.Stat(path)
	return err == nil
}

func (pg *TempPostgresInstance) initialize() error {
	cmd := newCommand("initdb",
		"-D", pg.dataDir,
		"-U", pg.user,
		"--no-locale",
		"--auth-local=trust",
		"--auth-host=trust",
	)
	return cmd.Run()
}

func (pg *TempPostgresInstance) start() error {
	logfilePath := filepath.Join(pg.dataDir, "logfile")
	socketPath := filepath.Join(pg.dataDir, "sockets")
	options := fmt.Sprintf("-k %s -p %d", socketPath, pg.port)

	if err := os.Mkdir(socketPath, 0700); err != nil {
		return err
	}

	cmd := newCommand("pg_ctl",
		"-D", pg.dataDir,
		"-l", logfilePath,
		"-o", options,
		"start",
	)
	return cmd.Run()
}

func (pg *TempPostgresInstance) createDatabase() error {
	cmd := newCommand("createdb",
		"-h", "localhost",
		"-p", strconv.Itoa(pg.port),
		"-U", pg.user,
		"-O", pg.user,
		pg.dbname,
	)
	return cmd.Run()
}

func (pg *TempPostgresInstance) stop() error {
	cmd := newCommand("pg_ctl",
		"-D", pg.dataDir,
		"stop",
	)
	return cmd.Run()
}

func newCommand(name string, arg ...string) *exec.Cmd {
	cmd := exec.Command(name, arg...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd
}
