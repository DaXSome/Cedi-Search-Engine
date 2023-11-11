package main

type UrlQueue struct {
	URL     string `json:"url"`
	ItemTag struct {
		Attr        string `json:"attr"`
		ValuePrefix string `json:"value_prefix"`
	} `json:"item_tag"`
}
