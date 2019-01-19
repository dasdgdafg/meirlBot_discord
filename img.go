package main

import (
	"encoding/xml"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
)

type CuteImage struct {
	// no fields for now, cache stuff later
	// maybe allow config per channel
}

type response struct {
	Count string `xml:"count,attr"`
	Posts []post `xml:"post"`
}

type post struct {
	File string `xml:"file_url,attr"`
}

const baseUrl = "https://gelbooru.com/index.php?page=dapi&s=post&q=index&tags="

//var urlShortener = regexp.MustCompile("^(.*)/[^/^\\.]+(\\.[^/]+)$")

var baseStrings = []string{"(?i)me( irl)",
	"(?i)me( on the (?:left|right))",
	"(?i)me( being lewd)",
	"(?i)me( with tags) (.+)"}

var regexes = []*regexp.Regexp{regexp.MustCompile(baseStrings[0]),
	regexp.MustCompile(baseStrings[1]),
	regexp.MustCompile(baseStrings[2]),
	regexp.MustCompile(baseStrings[3])}

// these must match the order of the regexes
var tags = []string{"solo score:>5 rating:questionable",
	"multiple_girls score:>5 rating:questionable -large_breasts -1boy -multiple_boys",
	"solo score:>5 masturbation",
	""}

// these tags are always included in searches
var alwaysTags = "loli"

// to avoid looking up the count each time. it would be better to get these once and cache instead of hard coding
var counts = []int{10000, 3500, 1500, 0}

// returns (matching string, image url)
func (c CuteImage) getImageForMessage(msg string, nick string) (string, string, error) {
	for i, reg := range regexes {
		if reg.MatchString(msg) {
			matches := reg.FindStringSubmatch(msg)
			matchingString := matches[1]
			imageUrl := ""
			var err error
			// use matches[2] (user specified tags) if there are no tags
			if tags[i] == "" && len(matches) > 2 {
				// strip colors out of the tags
				tagString := matches[2]
				imageUrl, err = c.getImage(counts[i], tagString)
			} else {
				imageUrl, err = c.getImage(counts[i], tags[i])
			}
			return nick + matchingString, imageUrl, err
		}
	}
	log.Println("error determining image type for " + msg)
	return "", "", nil
}

func (c CuteImage) checkForMatch(msg string) bool {
	for _, reg := range regexes {
		if reg.MatchString(msg) {
			return true
		}
	}
	return false
}

func (c CuteImage) getImage(count int, tags string) (string, error) {
	// fetch the count if we don't have it
	if count < 1 {
		newC, err := c.getCount(tags)
		if err != nil {
			return "", err
		}
		if newC < 1 {
			return "", nil
		}
		count = newC
	}
	pid := rand.Intn(count)
	requestUrl := baseUrl + url.QueryEscape(tags+" "+alwaysTags) + "&limit=1&pid=" + strconv.Itoa(pid)
	log.Println("getting image from " + requestUrl)
	resp, err := http.Get(requestUrl)
	if err != nil {
		log.Println("error fetching image")
		log.Println(err)
		return "", err
	}
	defer resp.Body.Close()
	respBody := response{}
	err = xml.NewDecoder(resp.Body).Decode(&respBody)
	if err != nil {
		log.Println("error decoding response")
		log.Println(err)
		return "", err
	}
	if len(respBody.Posts) == 0 {
		log.Println("no images found")
		return "", nil
	}

	return respBody.Posts[0].File, nil
}

func (c CuteImage) getCount(tags string) (int, error) {
	requestUrl := baseUrl + url.QueryEscape(tags+" "+alwaysTags) + "&limit=0"
	log.Println("getting count from " + requestUrl)
	resp, err := http.Get(requestUrl)
	if err != nil {
		log.Println("error fetching count")
		log.Println(err)
		return 0, err
	}
	defer resp.Body.Close()
	respBody := response{}
	err = xml.NewDecoder(resp.Body).Decode(&respBody)
	if err != nil {
		log.Println("error decoding response")
		log.Println(err)
		return 0, err
	}
	result, _ := strconv.Atoi(respBody.Count)
	return result, nil
}
