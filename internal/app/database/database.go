package database

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/binary"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5"
	buslock "github.com/matheustavarestrindade/rinha-de-backend-2024-q1.git/internal/app/bus"
)

var conn *pgx.Conn

func Connect() {
	// Connect to the database
	var err error

	conn, err = pgx.Connect(context.Background(), "postgres://admin:123@db:5432/rinha")

	if err != nil {
		log.Fatal(err)
	}

	err = conn.Ping(context.Background())
	if err != nil {
		panic(err.Error())
	}
	fmt.Println("Connected to the database")
	conn.Exec(context.Background(), `CREATE OR REPLACE FUNCTION 
            update_balance_and_insert_transaction(
                _clientId INT,
                _value INT,
                _type CHAR,
                _description VARCHAR(10)
            ) RETURNS TABLE(wl INT, fb INT) AS $$
                BEGIN
                    SELECT balance, withdraw_limit INTO fb, wl FROM clients WHERE id = _clientId;
                    IF (fb + _value) < -wl AND _type = 'd' THEN
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
	conn.Close(context.Background())
}

func GetClientByIdWithTransactions(clientId int) (*bytes.Buffer, error) {
	okSignal, done, ctx := buslock.Get().GetLock()
	defer done()
	<-okSignal

	query := `SELECT c.withdraw_limit, 
                     c.balance, 
                     t.value, 
                     t.type, 
                     t.description, 
                     t.created_at
             FROM clients c
             LEFT JOIN transaction t ON c.id = t.client_id
             WHERE c.id = $1 
             ORDER BY t.created_at DESC LIMIT 10;`

	rows, err := conn.Query(ctx, query, clientId)
	defer rows.Close()

	if err != nil {
		return nil, err
	}

	var buffer bytes.Buffer
	var isFirst = true
	var rawResults [10][6][]byte = [10][6][]byte{}
	var index = 0

	for rows.Next() {
		rawData := rows.RawValues()
		for i, raw := range rawData {
            // Copy the raw data to the rawResults array
            rawResults[index][i] = make([]byte, len(raw))
            copy(rawResults[index][i], raw)
		}
		index++
	}
    rows.Close()
	done()

	if len(rawResults) == 0 {
		return nil, err
	}

	for _, rawResult := range rawResults {
		if isFirst {
			buffer.Write([]byte(`{"saldo":{"limite":`))
			// Convert int32 to string
			buffer.WriteString(FastStr(toInt(rawResult[0])))
			buffer.Write([]byte(`,"total":`))
			buffer.WriteString(FastStr(toInt(rawResult[1])))
			buffer.Write([]byte(`,"data_extrato":"`))
			buffer.Write([]byte(time.Now().Format(time.RFC3339)))
			buffer.Write([]byte(`"},"ultimas_transacoes":[`))

		}
		if len(rawResult[2]) > 0 {
			if !isFirst {
				buffer.Write([]byte(`,`))
			}
			buffer.Write([]byte(`{"valor":`))
			buffer.WriteString(FastStr(toInt(rawResult[2])))
			buffer.Write([]byte(`,"tipo":"`))
			buffer.Write(rawResult[3])
			buffer.Write([]byte(`","descricao":"`))
			buffer.Write(rawResult[4])
			buffer.Write([]byte(`","realizada_em":"`))
			buffer.WriteString(timestampBytesToRFC3339(rawResult[5]))
			buffer.Write([]byte(`"}`))
		}
		isFirst = false
	}
	buffer.Write([]byte(`] }`))

	return &buffer, nil
}

func CreateClientTransaction(clientId int, transactionType string, transactionValue int, transactionDescription string) (string, string, bool) {
	okSignal, done, ctx := buslock.Get().GetLock()
	<-okSignal

	var finalBalance sql.NullInt32
	var withdrawLimit sql.NullInt32
	err := conn.QueryRow(ctx, `SELECT * FROM update_balance_and_insert_transaction($1, $2, $3, $4);`,
		clientId,
		transactionValue,
		transactionType,
		transactionDescription).Scan(&withdrawLimit, &finalBalance)

	done()
	if err != nil || !withdrawLimit.Valid || !finalBalance.Valid {
		return "", "", true
	}

	return FastStr(withdrawLimit.Int32), FastStr(finalBalance.Int32), false
}

type Number interface {
	int | int32
}

func FastStr[number Number](n number) string {
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
func toInt(bytes []byte) int32 {
    return int32(binary.BigEndian.Uint32(bytes))
}

func timestampBytesToRFC3339(buf []byte) string {
	// Ensure buffer is at least 4 bytes (Unix timestamp size), otherwise return an empty string
	if len(buf) < 4 {
		return ""
	}

	// Convert bytes to uint32 assuming big endian and then to int64 for Unix timestamp
	timestamp := int64(binary.BigEndian.Uint32(buf))

	// Convert Unix timestamp to time.Time
	t := time.Unix(timestamp, 0).UTC()

	// Format time.Time to RFC 3339 format
	return t.Format(time.RFC3339)
}
