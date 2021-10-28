package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/Krajiyah/new-world-api/internal"
	"github.com/PuerkitoBio/goquery"
	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

func main() {
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	db, err := internal.NewProdDB()
	checkError(err)
	checkError(db.AutoMigrate(&internal.Item{}))

	e := echo.New()
	e.GET("/ping", func(c echo.Context) error { return c.HTML(http.StatusOK, "pong") })
	e.GET("/item/:nameKey", handleGetItem(db, logger))

	port := os.Getenv("PORT")
	logger.WithField("port", port).Info("Running server")
	checkError(e.Start(":" + port))
}

func checkError(err error) {
	if err != nil {
		panic(err)
	}
}

func handleGetItem(db *gorm.DB, logger *logrus.Logger) func(echo.Context) error {
	return func(c echo.Context) error {
		nameKey := c.Param("nameKey")
		log := logger.WithContext(c.Request().Context()).WithField("nameKey", nameKey)

		item, err := getItemByNameViaDB(db, nameKey)
		if err == nil {
			log.WithField("item", item).Debug("got result (cached)")
			return c.JSON(http.StatusOK, item)
		}

		log.WithError(err).Warn("could not find item by name in db cache...looking on wiki")
		item, err = getItemByNameViaWiki(nameKey)
		if err == nil {
			log = log.WithField("item", item)
			log.Debug("got result")
			if err := db.Create(item).Error; err != nil {
				log.WithError(err).Error("could not create item in db cache")
				return c.HTML(http.StatusInternalServerError, "internal server error")
			}
			log.Debug("cached item in db")
			return c.JSON(http.StatusOK, item)
		}

		log.WithError(err).Warn("could not find item by name in wiki")
		return c.JSON(http.StatusNotFound, "item not found")
	}
}

func getItemByNameViaDB(db *gorm.DB, nameKey string) (*internal.Item, error) {
	item := &internal.Item{}
	err := db.Where("nameKey", nameKey).Find(item).Error
	return item, err
}

func getItemByNameViaWiki(nameKey string) (*internal.Item, error) {
	res, err := http.Get("https://newworld.fandom.com/wiki/" + nameKey)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return nil, fmt.Errorf("invalid status code: %d", res.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return nil, err
	}

	attrs := map[string]interface{}{}
	item := &internal.Item{NameKey: nameKey}
	item.Name = doc.Find("#firstHeading").Text()

	doc.Find(".pi-item .pi-data").Each(func(i int, s *goquery.Selection) {
		label := s.Find(".pi-data-label").Text()
		value := s.Find(".pi-data-value").Text()
		attrs[label] = value
	})

	attrsJson, err := internal.MapToJson(attrs)
	if err != nil {
		return nil, err
	}

	item.Attributes = attrsJson
	return item, nil
}
