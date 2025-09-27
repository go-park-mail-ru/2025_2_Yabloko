package db

import (
	"context"
	"fmt"
	"log"
	"math/rand/v2"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

var restaurants = []AppendInfo{
	{
		Name:        "Суши Мастер",
		Description: "Современный японский ресторан с широким выбором суши и роллов.",
		Address:     "ул. Тверская, 12",
		CardImg:     "/stores/1.jpg",
		Rating:      4.5,
		OpenAt:      time.Date(0, 1, 1, 10, 0, 0, 0, time.Local),
		ClosedAt:    time.Date(0, 1, 1, 23, 0, 0, 0, time.Local),
	},
	{
		Name:        "Sushi House",
		Description: "Аутентичные японские суши и сашими, свежие ингредиенты каждый день.",
		Address:     "Невский пр., 45",
		CardImg:     "/stores/2.jpg",
		Rating:      4.3,
		OpenAt:      time.Date(0, 1, 1, 10, 0, 0, 0, time.Local),
		ClosedAt:    time.Date(0, 1, 1, 23, 0, 0, 0, time.Local),
	},
	{
		Name:        "Pizza Roma",
		Description: "Итальянская пицца на тонком тесте, приготовленная в дровяной печи.",
		Address:     "ул. Ленина, 78",
		CardImg:     "/stores/3.jpg",
		Rating:      4.6,
		OpenAt:      time.Date(0, 1, 1, 10, 0, 0, 0, time.Local),
		ClosedAt:    time.Date(0, 1, 1, 23, 0, 0, 0, time.Local),
	},
	{
		Name:        "Trattoria Italia",
		Description: "Уютная итальянская траттория с пастой, ризотто и авторской пиццей.",
		Address:     "ул. Советская, 21",
		CardImg:     "/stores/4.jpg",
		Rating:      4.4,
		OpenAt:      time.Date(0, 1, 1, 10, 0, 0, 0, time.Local),
		ClosedAt:    time.Date(0, 1, 1, 23, 0, 0, 0, time.Local),
	},
	{
		Name:        "La Bella Italia",
		Description: "Ресторан итальянской кухни с атмосферой Рима и Венеции.",
		Address:     "ул. Большая Покровская, 32",
		CardImg:     "/stores/5.jpg",
		Rating:      4.7,
		OpenAt:      time.Date(0, 1, 1, 10, 0, 0, 0, time.Local),
		ClosedAt:    time.Date(0, 1, 1, 23, 0, 0, 0, time.Local),
	},
	{
		Name:        "Pho Vietnam",
		Description: "Вьетнамские супы, спринг-роллы и свежие салаты в центре города.",
		Address:     "ул. Вайнера, 10",
		CardImg:     "/stores/6.jpg",
		Rating:      4.5,
		OpenAt:      time.Date(0, 1, 1, 10, 0, 0, 0, time.Local),
		ClosedAt:    time.Date(0, 1, 1, 23, 0, 0, 0, time.Local),
	},
	{
		Name:        "Indian Spice",
		Description: "Аутентичная индийская кухня с карри, самосой и традиционными специями.",
		Address:     "ул. Арбат, 18",
		CardImg:     "/stores/7.jpg",
		Rating:      4.6,
		OpenAt:      time.Date(0, 1, 1, 10, 0, 0, 0, time.Local),
		ClosedAt:    time.Date(0, 1, 1, 23, 0, 0, 0, time.Local),
	},
	{
		Name:        "Tandoori House",
		Description: "Индийский ресторан с блюдами из тандура и вегетарианскими опциями.",
		Address:     "ул. Малая Конюшенная, 5",
		CardImg:     "/stores/8.jpg",
		Rating:      4.4,
		OpenAt:      time.Date(0, 1, 1, 10, 0, 0, 0, time.Local),
		ClosedAt:    time.Date(0, 1, 1, 23, 0, 0, 0, time.Local),
	},
	{
		Name:        "Ristorante Italia",
		Description: "Итальянские паста, пицца и десерты в уютной атмосфере.",
		Address:     "ул. Куйбышева, 40",
		CardImg:     "/stores/9.jpg",
		Rating:      4.5,
		OpenAt:      time.Date(0, 1, 1, 10, 0, 0, 0, time.Local),
		ClosedAt:    time.Date(0, 1, 1, 23, 0, 0, 0, time.Local),
	},
	{
		Name:        "Burger Time",
		Description: "Классические и авторские бургеры с картошкой и соусами.",
		Address:     "ул. Пролетарская, 11",
		CardImg:     "/stores/10.jpg",
		Rating:      4.3,
		OpenAt:      time.Date(0, 1, 1, 10, 0, 0, 0, time.Local),
		ClosedAt:    time.Date(0, 1, 1, 23, 0, 0, 0, time.Local),
	},
	{
		Name:        "Fries & Burgers",
		Description: "Бургеры, картофель фри и напитки на любой вкус.",
		Address:     "ул. Рождественская, 25",
		CardImg:     "/stores/11.jpg",
		Rating:      4.2,
		OpenAt:      time.Date(0, 1, 1, 10, 0, 0, 0, time.Local),
		ClosedAt:    time.Date(0, 1, 1, 23, 0, 0, 0, time.Local),
	},
	{
		Name:        "Meat Lovers",
		Description: "Мясные бургеры с сочной котлетой и хрустящей булочкой.",
		Address:     "ул. Белинского, 14",
		CardImg:     "/stores/12.jpg",
		Rating:      4.6,
		OpenAt:      time.Date(0, 1, 1, 10, 0, 0, 0, time.Local),
		ClosedAt:    time.Date(0, 1, 1, 23, 0, 0, 0, time.Local),
	},
	{
		Name:        "Gourmet Palace",
		Description: "Элитный ресторан с авторской кухней и винной картой.",
		Address:     "ул. Новый Арбат, 7",
		CardImg:     "/stores/13.jpg",
		Rating:      4.8,
		OpenAt:      time.Date(0, 1, 1, 10, 0, 0, 0, time.Local),
		ClosedAt:    time.Date(0, 1, 1, 23, 0, 0, 0, time.Local),
	},
	{
		Name:        "Saigon Corner",
		Description: "Вьетнамская кухня с традиционными супами и свежими роллами.",
		Address:     "ул. Пионерская, 9",
		CardImg:     "/stores/14.jpg",
		Rating:      4.5,
		OpenAt:      time.Date(0, 1, 1, 10, 0, 0, 0, time.Local),
		ClosedAt:    time.Date(0, 1, 1, 23, 0, 0, 0, time.Local),
	},
}

func fakeTag(dbPool *pgxpool.Pool) {
	query := `
	insert into tag (id, name)
	values ($1, $2);
	`

	var tags = []string{
		"Суши",
		"Индия",
		"Пицца",
		"Италия",
		"Фастфуд",
		"Самовывоз",
	}

	for _, tag := range tags {
		id := uuid.New()
		_, err := dbPool.Exec(context.Background(), query, id, tag)
		if err != nil {
			log.Fatal(err)
		}
	}
}

var cities = []string{
	"Москва",
	"Санкт-Петербург",
	"Самара",
	"Тула",
	"Нижний Новгород",
	"Екатеринбург",
}

func fakeCity(dbPool *pgxpool.Pool) {
	query := `
	insert into city (id, name)
	values ($1, $2);
	`

	for _, name := range cities {
		id := uuid.New()
		_, err := dbPool.Exec(context.Background(), query, id, name)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func FakeTagCity(dbPool *pgxpool.Pool) {
	fakeTag(dbPool)
	fakeCity(dbPool)
}

func FakeStores(dbPool *pgxpool.Pool) {
	for _, obj := range restaurants {
		city := cities[rand.IntN(len(cities))]
		query := "select id from city where name=$1"
		var cityId uuid.UUID
		_ = dbPool.QueryRow(context.Background(), query, city).Scan(&cityId)
		obj.CityId = cityId

		err := AppendStore(dbPool, obj)
		if err != nil {
			fmt.Println("ОШИБКА")
		}
	}
}
