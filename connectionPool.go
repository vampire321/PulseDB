package main

import(
	"context"
	"fmt"
	"os"
	"time"
	"errors"
	"github.com/jackc/pgx/v5"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main(){
	ctx := context.Background()
	//these helps to read from environment in real apps
	connStr := "postgres://postgres:secret@localhost:5433/pulsedb"
	//pgxpool.New creates the pool - does Not open connection yet
	pool, err := pgxpool.New(ctx, connStr)
	if err != nil{
		fmt.Println("failed  to craete pool", err)
		os.Exit(1)
	}
	defer pool.Close()

	//ping opens one connection to verify it works
	if err := pool.Ping(ctx); 
		err != nil{
		fmt.Println("failed to ping database", err)
		os.Exit(1)
	}
	fmt.Println("connected to POstgreSQL")

	//QuerryRow for single row
	var version string
	err = pool.QueryRow(ctx, "SELECT version()").Scan(&version)
	if err != nil{
		//handles error
		fmt.Println("Query failed", err)
		os.Exit(1)
	}
	fmt.Println("PostgreSQL version:", version)
// Pattern 1 — INSERT
var id, status string
var createdAt, updatedAt time.Time

err = pool.QueryRow(ctx,
    `INSERT INTO monitors (name, url, interval_s)
     VALUES ($1, $2, $3)
     RETURNING id, status, created_at, updated_at`,
    "Google", "https://google.com", 30,
).Scan(&id, &status, &createdAt, &updatedAt)

if err != nil {
    fmt.Println("Insert failed:", err)
    os.Exit(1)
}
fmt.Println("Inserted:", id)

//pattern-2 SELECT
var name, url string
var intervalS int

err = pool.QueryRow(ctx,
    `SELECT name, url, interval_s FROM monitors WHERE id = $1`,
    id,
).Scan(&name, &url, &intervalS)

if errors.Is(err, pgx.ErrNoRows) {
    fmt.Println("not found")
}else{
	fmt.Println("fetched monitor:", name, url, intervalS)
}

//pattern -3 SELECT multiple row
rows, err := pool.Query(ctx, `SELECT name, url, interval_s FROM monitors`) //it sends sql to database and db returns multiple rows
if err != nil{
	panic(err) //stop progeam if querry fails
}
defer rows.Close() //close rows when done(can  cause memory leak if not closed)
for rows.Next(){
	var id , name , url string
	rows.Scan(&id,&name,&url) //scan each row into variables
	fmt.Println("monitor:", id, name, url)
}
}
