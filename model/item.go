package model

type Item struct {
	Name         string `csv:"name"`
	Description  string `csv:"description"`
	ImageURL     string `csv:"image_url"`
	Price        int    `csv:"price"`
	Rating       string `csv:"rating"`
	Store        string `csv:"store"`
	DetailedLink string `csv:"-"`
}
