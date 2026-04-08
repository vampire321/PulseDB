package main

import(
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main(){
	ctx := context.Background()
	//these helps to read from environment in real apps
	connStr := "postgres://postgres:password@localhost:5432/postgres"

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
}
