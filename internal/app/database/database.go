package database

import (
	"bytes"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"time"

	_ "github.com/lib/pq"
	buslock "github.com/matheustavarestrindade/rinha-de-backend-2024-q1.git/internal/app/bus"
)

var db *sql.DB

func Connect() {
	// Connect to the database
	var err error

	db, err = sql.Open("postgres", "host=db user=admin password=123 dbname=rinha sslmode=disable")
	db.SetMaxOpenConns(30)

	if err != nil {
		log.Fatal(err)
	}

	err = db.Ping()
	if err != nil {
		panic(err.Error())
	}
	fmt.Println("Connected to the database")

	db.Exec(`CREATE OR REPLACE FUNCTION 
            update_balance_and_insert_transaction(
                _clientId INT,
                _value INT,
                _type CHAR,
                _description VARCHAR(10)
            ) RETURNS TABLE(wl INT, fb INT) AS $$
                DECLARE
                    bal INT;
                BEGIN

                    SELECT balance, withdraw_limit INTO bal, wl FROM clients WHERE id = _clientId;

                    IF (bal + _value) < -wl AND _type = 'd' THEN
                        RAISE NOTICE 'Insufficient funds';
                    ELSE
                        UPDATE clients SET balance = balance + _value WHERE id = _clientId RETURNING balance INTO fb; 
                        INSERT INTO transaction (client_id, value, type, description) VALUES (_clientId, ABS(_value), _type, _description);
                        RETURN NEXT;
                    END IF;
                END;
            $$ LANGUAGE plpgsql;
    `)
}

func Close() {
	db.Close()
}

func GetClientByIdWithTransactions(clientId int) (*bytes.Buffer, error) {
	okSignal, done := buslock.Get().GetLock()
	defer done()
	<-okSignal

	query := `SELECT c.id AS client_id, 
                     c.withdraw_limit, 
                     c.balance, 
                     t.id AS transaction_id, 
                     t.value, 
                     t.type, 
                     t.description, 
                     t.created_at
             FROM clients c
             LEFT JOIN transaction t ON c.id = t.client_id
             WHERE c.id = $1 
             ORDER BY t.created_at DESC;`

	rows, err := db.Query(query, clientId)
	defer rows.Close()

	if err != nil {
		return nil, err
	}

	var clientWithdrawLimit sql.NullInt32
	var clientBalance sql.NullInt32
	var transaction_id, transaction_value sql.NullInt32
	var transaction_type sql.NullString
	var transaction_description sql.NullString
	var transaction_created_at sql.NullTime

	var buffer bytes.Buffer
	var isFirst = true

	for rows.Next() {
		err = rows.Scan(&clientId, &clientWithdrawLimit, &clientBalance, &transaction_id, &transaction_value, &transaction_type, &transaction_description, &transaction_created_at)
		if err != nil {
			return nil, err
		}
		if isFirst {
			buffer.Write([]byte(`{"saldo":{"limite":` + FastStr(clientWithdrawLimit.Int32) + `,"total":` + FastStr(clientBalance.Int32) + `,"data_extrato":"` + time.Now().Format(time.RFC3339) + `"},"ultimas_transacoes":[`))
		}

		if transaction_id.Valid && transaction_value.Valid && transaction_type.Valid && transaction_created_at.Valid {
			if !isFirst {
				buffer.Write([]byte(`,`))
			}

			buffer.Write([]byte(`{"valor":` + FastStr(transaction_value.Int32) + `,"tipo":"` + transaction_type.String + `","descricao":"` + transaction_description.String + `","realizada_em":"` + transaction_created_at.Time.Format(time.RFC3339) + `"}`))
			isFirst = false
		}
	}
	buffer.Write([]byte(`] }`))

	return &buffer, nil
}

func CreateClientTransaction(clientId int, transactionType string, transactionValue int, transactionDescription string) (int, int, error) {
	okSignal, done := buslock.Get().GetLock()
	defer done()
	<-okSignal

	var finalBalance sql.NullInt32
	var withdrawLimit sql.NullInt32
	err := db.QueryRow(`SELECT * FROM update_balance_and_insert_transaction($1, $2, $3, $4);`,
		clientId,
		transactionValue,
		transactionType,
		transactionDescription).Scan(&withdrawLimit, &finalBalance)

	if err != nil || !withdrawLimit.Valid || !finalBalance.Valid {
		return 0, 0, errors.New("Insufficient funds")
	}

	return int(withdrawLimit.Int32), int(finalBalance.Int32), nil
}

func FastStr(n int32) string {
	buf := [11]byte{}
	pos := len(buf)
	i := int64(n)
	signed := i < 0
	if signed {
		i = -i
	}
	for {
		pos--
		buf[pos], i = '0'+byte(i%10), i/10
		if i == 0 {
			if signed {
				pos--
				buf[pos] = '-'
			}
			return string(buf[pos:])
		}
	}
}
