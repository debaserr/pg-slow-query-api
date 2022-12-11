package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cache"
	"github.com/jackc/pgx/v5/pgxpool"
)

func getQueryParams(c *fiber.Ctx) (*QueryParams, error) {
	var q QueryParams
	if err := c.QueryParser(&q); err != nil {
		return nil, err
	}
	return &q, nil
}

func main() {
	url := fmt.Sprintf("postgres://%s:%s@%s:5432/%s",
		os.Getenv("DB_USER"), os.Getenv("DB_PASSWORD"), os.Getenv("DB_HOST"), os.Getenv("DB_NAME"))
	db, err := pgxpool.New(context.Background(), url)
	if err != nil {
		panic(err)
	}

	repo := NewPostgresRepo(db)
	defer repo.Close()

	app := Setup(repo)

	if err := app.Listen(":3000"); err != nil {
		panic(err)
	}
}

func Setup(repo Repo) *fiber.App {
	app := fiber.New()

	app.Use(cache.New(cache.Config{
		Expiration: 30 * time.Minute,
	}))

	app.Get("/slow-queries", func(c *fiber.Ctx) error {
		params, err := getQueryParams(c)
		if err != nil {
			panic(err)
		}
		res, err := repo.Get(params)
		if err != nil {
			panic(err)
		}
		return c.JSON(res)
	})

	app.Get("/demo/init", func(c *fiber.Ctx) error {
		err := repo.Demo()
		if err != nil {
			return err
		}
		return c.SendString("init demo")
	})

	return app
}
