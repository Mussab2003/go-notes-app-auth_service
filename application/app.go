package application

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type App struct {
	router http.Handler
	pgdb *pgxpool.Pool
}

func New(ctx context.Context) *App {
	user := os.Getenv("DB_USER")
	pass := os.Getenv("DB_PASSWORD")
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	dbname := os.Getenv("DB_NAME")
	
	connString := fmt.Sprintf("postgres://%s:%s@%s:%s/%s",
		user, pass, host, port, dbname,
	)
	
	dbpool, err := pgxpool.New(ctx, connString)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}
	app := &App{
		router: loadRoutes(),
		pgdb: dbpool,
	}
	return app
}

func (a *App) Start(ctx context.Context) error {
	server := &http.Server{
		Addr : ":8000",
		Handler: a.router,
	}

	if err := a.pgdb.Ping(ctx); err != nil {
		return fmt.Errorf("unable to connect to postgres: %w", err)
	}

	defer func() {
		a.pgdb.Close()
	}()

	ch := make(chan error, 1)

	go func(){
		err := server.ListenAndServe()
		if err != nil{
			ch <- fmt.Errorf("unable to start server, %w", err)
		}
		close(ch)
	}()
	
	select{
	case err := <- ch:
		return err
	case <- ctx.Done():
		timeout, cancel := context.WithTimeout(context.Background(), time.Second * 3)
		defer cancel()
		return server.Shutdown(timeout)
	}
}
