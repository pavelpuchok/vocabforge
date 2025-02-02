package telegram

import (
	"database/sql"
	"errors"
	"fmt"
	"log/slog"

	"gopkg.in/telebot.v4"
)

const SQLTxContextKey = "__sqlDBTx"

func NewSQLTxMiddleware(db *sql.DB, log *slog.Logger) telebot.MiddlewareFunc {
	return func(hf telebot.HandlerFunc) telebot.HandlerFunc {
		return func(c telebot.Context) error {
			tx, err := db.Begin()
			if err != nil {
				return fmt.Errorf("SQLTxMiddleware failed to start transaction. %w", err)
			}

			defer func() {
				err := tx.Rollback()
				if err != nil {
					log.Error(fmt.Sprintf("SQLTxMiddleware failed to rollback transaction. %s", err))
				}
			}()

			c.Set(SQLTxContextKey, tx)

			if err := hf(c); err != nil {
				return err
			}

			err = tx.Commit()
			if err != nil {
				return fmt.Errorf("SQLTxMiddleware failed to commit transaction. %w", err)
			}

			return nil
		}
	}
}

func GetSQLTx(ctx telebot.Context) (*sql.Tx, error) {
	tx, ok := ctx.Get(SQLTxContextKey).(*sql.Tx)
	if !ok {
		return nil, errors.New("unable to get SQL Transaction from telebot context")
	}
	return tx, nil
}
