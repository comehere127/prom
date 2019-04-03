package main

import (
	"fmt"
	"github.com/btnguyen2k/prom"
	_ "github.com/denisenkom/go-mssqldb"
	"math/rand"
	"reflect"
	"strconv"
	"strings"
	"time"
)

var timezoneMssql = "Asia/Kabul"

// construct an 'prom.MongoConnect' instance
func createSqlConnectMssql() *prom.SqlConnect {
	driver := "sqlserver"
	dsn := "sqlserver://sa:test@localhost:1433?database=tempdb"
	sqlConnect, err := prom.NewSqlConnectWithFlavor(driver, dsn, 10000, nil, prom.FlavorMsSql)
	if sqlConnect == nil || err != nil {
		if err != nil {
			fmt.Println("Error:", err)
		}
		if sqlConnect == nil {
			panic("error creating [prom.SqlConnect] instance")
		}
	}
	loc, _ := time.LoadLocation(timezoneMssql)
	sqlConnect.SetLocation(loc)
	return sqlConnect
}

var colsMssql = []string{"id", "username", "email",
	"data_bool", "data_int", "data_float",
	"data_time", "data_timez",
	"data_date", "data_datez",
	"data_datetime", "data_datetimez",
	"data_timestamp", "data_timestampz"}

func printRowMssql(row map[string]interface{}) {
	id := row["id"]
	fmt.Printf("\t\tRow [%v]\n", id)
	for _, n := range colsMssql {
		v := row[n]
		fmt.Println("\t\t\t", n, "[", reflect.TypeOf(v), "] = ", v)
	}
}

func main() {
	rand.Seed(time.Now().UnixNano())
	SEP := "======================================================================"
	sqlConnect := createSqlConnectMssql()
	defer sqlConnect.Close()
	loc, _ := time.LoadLocation(timezoneMssql)
	fmt.Println("Timezone:", loc)

	{
		fmt.Println("-== Database & Ping info ==-")

		// get the database object and send ping command
		db := sqlConnect.GetDB()
		fmt.Println("\tCurrent database:", db)
		fmt.Println("\tIs connected    :", sqlConnect.IsConnected())
		err := sqlConnect.Ping(nil)
		if err != nil {
			fmt.Println("\tPing error      :", err)
		} else {
			fmt.Println("\tPing ok")
		}

		fmt.Println(SEP)
	}

	{
		fmt.Println("-== Drop/Create Table ==-")

		// setting up
		sql := "DROP TABLE IF EXISTS tbl_demo"
		_, err := sqlConnect.GetDB().Exec(sql)
		if err != nil {
			fmt.Printf("\tError while executing query [%s]: %s\n", sql, err)
		} else {
			fmt.Println("\tDropped table [tbl_demo]")

			types := []string{"INT", "NVARCHAR(64)", "NVARCHAR(128)",
				"NCHAR(1)", "INT", "FLOAT(53)",
				"TIME", "TIME",
				"DATE", "DATE",
				"DATETIME", "DATETIMEOFFSET",
				"DATETIME2", "DATETIMEOFFSET"}

			sql := "CREATE TABLE tbl_demo ("
			for i := range colsMssql {
				sql += colsMssql[i] + " " + types[i] + ","
			}
			sql += "PRIMARY KEY(id))"
			fmt.Println("\tQuery:" + sql)

			_, err := sqlConnect.GetDB().Exec(sql)
			if err != nil {
				fmt.Printf("\tError while executing query: %s\n", err)
			} else {
				fmt.Println("\tCreated table [tbl_demo]")
			}
		}

		fmt.Println(SEP)
	}

	{
		fmt.Println("-== Insert Rows to Table ==-")

		// insert some rows
		sql := "INSERT INTO tbl_demo ("
		sql += strings.Join(colsMssql, ",")
		sql += ") VALUES ("
		for k := range colsMssql {
			sql += "@p" + strconv.Itoa(k+1) + ","
		}
		sql = sql[0 : len(sql)-1]
		sql += ")"

		n := 100
		fmt.Printf("\tInserting %d rows to table [tbl_demo]\n", n)
		for i := 1; i <= n; i++ {
			t := time.Unix(int64(rand.Int31()), rand.Int63()%1000000000).In(loc)
			id := i
			username := t.String()
			email := strconv.Itoa(int(rand.Int31n(int32(n)))) + "@" + strconv.Itoa(int(rand.Int31n(9999))) + ".com"
			dataInt := rand.Int31()
			dataBool := strconv.Itoa(int(dataInt % 2))
			dataFloat := rand.Float64()
			dataTime := t
			dataTimez := t
			dataDate := t
			dataDatez := t
			dataDatetime := t
			dataDatetimez := t
			dataTimestamp := t
			dataTimestampz := t
			_, err := sqlConnect.GetDB().Exec(sql, id, username, email, dataBool, dataInt, dataFloat,
				dataTime, dataTimez, dataDate, dataDatez, dataDatetime, dataDatetimez, dataTimestamp, dataTimestampz)
			if err != nil {
				fmt.Println("\t\tError:", err)
			}
		}

		fmt.Println(SEP)
	}

	{
		fmt.Println("-== Query Single Row from Table ==-")

		// query single row
		sql := "SELECT * FROM tbl_demo WHERE id=@p1"

		id := rand.Intn(100) + 1
		fmt.Printf("\tFetching row id %d from table [tbl_demo]\n", id)
		dbRow := sqlConnect.GetDB().QueryRow(sql, id)
		data, err := sqlConnect.FetchRow(dbRow, len(colsMssql))
		if err != nil {
			fmt.Printf("\t\tError fetching row %d from table [tbl_demo]: %s\n", id, err)
		} else if data == nil {
			fmt.Println("\t\tRow not found")
		} else {
			for _, v := range data {
				switch v.(type) {
				case []byte:
					fmt.Println("\t\t", reflect.TypeOf(v), string(v.([]byte)))
				default:
					fmt.Println("\t\t", reflect.TypeOf(v), v)
				}
			}
		}

		id = 999
		fmt.Printf("\tFetching row id %d from table [tbl_demo]\n", id)
		dbRow = sqlConnect.GetDB().QueryRow(sql, id)
		data, err = sqlConnect.FetchRow(dbRow, len(colsMssql))
		if err != nil {
			fmt.Printf("\t\tError fetching row %d from table [tbl_demo]: %s\n", id, err)
		} else if data == nil {
			fmt.Println("\t\tNo row matches query")
		} else {
			for _, v := range data {
				switch v.(type) {
				case []byte:
					fmt.Println("\t\t", reflect.TypeOf(v), string(v.([]byte)))
				default:
					fmt.Println("\t\t", reflect.TypeOf(v), v)
				}
			}
		}

		fmt.Println(SEP)
	}

	{
		fmt.Println("-== Query Multiple Rows from Table ==-")

		// query multiple rows: https://docs.microsoft.com/en-us/sql/t-sql/queries/select-order-by-clause-transact-sql?view=sql-server-2017
		// (SQL Server 2012+ only)
		sql := "SELECT * FROM tbl_demo WHERE id>=@p1 ORDER BY id OFFSET 0 ROWS FETCH NEXT 4 ROWS ONLY"

		id := rand.Intn(100) + 1
		fmt.Printf("\tFetching rows starting at %d from table [tbl_demo]\n", id)
		dbRows1, err := sqlConnect.GetDB().Query(sql, id)
		defer dbRows1.Close()
		if err != nil {
			fmt.Printf("\tE\trror while executing query: %s\n", err)
		} else {
			rows, err := sqlConnect.FetchRows(dbRows1)
			if err != nil {
				fmt.Printf("\t\tError while fetching rows from table [tbl_demo]: %s\n", err)
			} else if len(rows) > 0 {
				for _, r := range rows {
					printRowMssql(r)
				}
			} else {
				fmt.Println("\t\tNo row matches query")
			}
		}

		id = 999
		fmt.Printf("\tFetching rows starting at %d from table [tbl_demo]\n", id)
		dbRows2, err := sqlConnect.GetDB().Query(sql, id)
		defer dbRows2.Close()
		if err != nil {
			fmt.Printf("\t\tError while executing query: %s\n", err)
		} else {
			rows, err := sqlConnect.FetchRows(dbRows2)
			if err != nil {
				fmt.Printf("\t\tError while fetching rows from table [tbl_demo]: %s\n", err)
			} else if len(rows) > 0 {
				for _, r := range rows {
					printRowMssql(r)
				}
			} else {
				fmt.Println("\t\tNo row matches query")
			}
		}

		fmt.Println(SEP)
	}

	{
		fmt.Println("-== Query Multiple Rows from Table (using callback) ==-")

		// query multiple rows with callback function
		sql := "SELECT * FROM tbl_demo WHERE id>=@p1 ORDER BY id OFFSET 0 ROWS FETCH NEXT 4 ROWS ONLY"
		callback := func(row map[string]interface{}, err error) bool {
			if err != nil {
				fmt.Printf("\t\tError while fetching rows from table [tbl_demo]: %s\n", err)
			} else {
				printRowMssql(row)
			}
			return true
		}

		id := rand.Intn(100) + 1
		fmt.Printf("\tFetching rows starting at %d from table [tbl_demo]\n", id)
		dbRows1, err := sqlConnect.GetDB().Query(sql, id)
		defer dbRows1.Close()
		if err != nil {
			fmt.Printf("\t\tError while executing query: %s\n", err)
		} else {
			sqlConnect.FetchRowsCallback(dbRows1, callback)
		}

		id = 999
		fmt.Printf("\tFetching rows starting at %d from table [tbl_demo]\n", id)
		dbRows2, err := sqlConnect.GetDB().Query(sql, id)
		defer dbRows2.Close()
		if err != nil {
			fmt.Printf("\t\tError while executing query: %s\n", err)
		} else {
			sqlConnect.FetchRowsCallback(dbRows2, callback)
		}

		fmt.Println(SEP)
	}
}
