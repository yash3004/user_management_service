package internal

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	"github.com/pkg/errors"

	"github.com/yash3004/user_management_service/cmd"
	"github.com/go-jet/jet/v2/mysql"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/golang-migrate/migrate/v4/database/mysql"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"k8s.io/klog/v2"

	"embed"
)

//go:embed migrations/*.sql
var sqlFiles embed.FS

const SQLLogLevel = 6

const transactionContextKey = "UMS:transaction"

func LogStatement(ctx context.Context, statement mysql.Statement) {

	logger := klog.FromContext(ctx)
	if !logger.V(SQLLogLevel).Enabled() {
		return
	}

	stmt, p := statement.Sql()
	logger.Info("executing statement", "statement", stmt, "params", p)
}

func CreateMySqlConnection(cfg cmd.Config) (*sql.DB, error) {
	dsn := cfg.DB.CreateDSN()
	MigrateSQLs(dsn)

	return sql.Open("mysql", dsn)
}

func MigrateSQLs(dsn string) {
	dsn = fmt.Sprintf("%s&multiStatements=true", dsn)

	source, err := iofs.New(sqlFiles, "migrations")
	if err != nil {
		log.Fatalf("cannot prepare migrations source:%v", err)
	}

	klog.V(3).Infof("opening mysql connection for migratoins...")
	mig, err := migrate.NewWithSourceInstance("iofs", source, "mysql://"+dsn)
	if err != nil {
		log.Fatalf("cannot prepare migration instance: %v", err)
	}

	logMigrationVersions("[before migration]", mig)

	if err := mig.Up(); err != nil && err != migrate.ErrNoChange {
		klog.Errorf("up migrations failed: %v", err)

		if err := mig.Down(); err != nil {
			klog.Errorf("migrations rolled back failed: %v", err)
		} else {
			klog.Info("rolled back to last version ")
		}

		panic(err)
	}

	logMigrationVersions("[after migration]", mig)
}

func NewTransactionContext(ctx context.Context, db *sql.DB, opts *sql.TxOptions) (context.Context, *sql.Tx, error) {
	tx, ok := ctx.Value(transactionContextKey).(*sql.Tx)
	if ok {
		return ctx, tx, nil
	}

	tx, err := db.BeginTx(ctx, opts)
	if err != nil {
		return nil, nil, err
	}

	return context.WithValue(ctx, transactionContextKey, tx), tx, nil
}

func GetTransaction(ctx context.Context) (*sql.Tx, error) {
	tx, ok := ctx.Value(transactionContextKey).(*sql.Tx)
	if ok {
		return tx, nil
	}

	return nil, errors.New("no transaction found")
}

func logMigrationVersions(logHead string, mig *migrate.Migrate) {
	if v, d, err := mig.Version(); err != nil && err != migrate.ErrNilVersion {
		klog.Fatalf("cannot load migration version: %v", err)
	} else {
		klog.Infof("%s: current database version is %d (dirty=%v)", logHead, v, d)
	}
}
