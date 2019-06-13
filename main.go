package main

import (
	"fmt"
	"github.com/bitly/go-simplejson"
	"github.com/deanishe/awgo"
	"github.com/mitchellh/go-homedir"
	"github.com/mozillazg/go-pinyin"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

type Bookmarks struct {
	name string
	url  string
}

func fetchChildren(contents map[string]interface{}) []Bookmarks {
	var bookmarks []Bookmarks
	for _, children := range contents["children"].([]interface{}) {
		childrenType := children.(map[string]interface{})["type"]
		switch childrenType {
		case "folder":
			bookmark := fetchChildren(children.(map[string]interface{}))
			bookmarks = append(bookmarks, bookmark...)
		case "url":
			var bookmark Bookmarks
			bookmark.name = children.(map[string]interface{})["name"].(string)
			bookmark.url = children.(map[string]interface{})["url"].(string)
			bookmarks = append(bookmarks, bookmark)
		}
	}
	return bookmarks
}

func pinyinFuzzyMatch(chnStr string, query string) bool {
	pinyinSlice := pinyin.Pinyin(chnStr, pinyin.NewArgs())
	if pinyinSlice == nil {
		return false
	}

	var pinyinStr string
	for _, p1 := range pinyinSlice {
		pinyinStr += p1[0]
	}
	//fmt.Println(pinyinStr)
	return strings.Contains(pinyinStr, query)
}

func enFuzzyMatch(enStr string, query string) bool {
	return strings.Contains(enStr, query)
}

var wf *aw.Workflow

func init() {
	// Create a new Workflow using default settings.
	// Critical settings are provided by Alfred via environment variables,
	// so this *will* die in flames if not run in an Alfred-like environment.
	wf = aw.New()
}

func run() {
	path, err := homedir.Expand("~/Library/Application Support/Google/Chrome/Default/Bookmarks")
	if err != nil {
		log.Fatal(err)
	}

	content, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatal(err)
	}

	res, err := simplejson.NewJson(content)

	if err != nil {
		fmt.Printf("%v\n", err)
		return
	}
	bookmarksBar, err := res.Get("roots").Get("bookmark_bar").Map()
	bookmarks := fetchChildren(bookmarksBar)

	for _, bookmark := range bookmarks {
		if len(os.Args) <= 1 {
			return
		}
		if pinyinFuzzyMatch(bookmark.name, os.Args[1]) || enFuzzyMatch(bookmark.name, os.Args[1]) || enFuzzyMatch(bookmark.url, os.Args[1]) {
			wf.NewItem(bookmark.name).Subtitle(bookmark.url).Valid(true).Arg(bookmark.url)
		}
	}
	wf.SendFeedback()
}

func main() {
	wf.Run(run)
}
