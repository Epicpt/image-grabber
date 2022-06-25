package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"sync"
)

var wg sync.WaitGroup

func checkError(err error) {
	if err != nil {
		panic(err)
	}
}

func removeDuplicateLink(sliceLink []string) []string {
	allKeys := make(map[string]bool)
	list := []string{}
	for _, ok := range sliceLink {
		if _, value := allKeys[ok]; !value {
			allKeys[ok] = true
			list = append(list, ok)
		}
	}
	return list
}

func downloadFile(ch chan string) {
	fullUrlFile := <-ch
	// создаем имя файла
	fileUrl, err := url.Parse(fullUrlFile)
	checkError(err)
	path := fileUrl.Path
	segments := strings.Split(path, "/")
	fileName := segments[len(segments)-1]
	// создаем пустой файла
	file, err := os.Create(fileName)
	checkError(err)
	defer file.Close()
	// получаем байты из Url адреса
	response, err := http.Get(fullUrlFile)
	checkError(err)
	defer response.Body.Close()
	if response.StatusCode != 200 { // проверяется что ответ ОК
		log.Fatal("Received non 200 response code")
	}
	// записываем байты в файл
	_, err = io.Copy(file, response.Body)
	checkError(err)

	fmt.Println(fileName)
	wg.Done()
}

func main() {
	response, err := http.Get("https://www.starwars.com/the-book-of-boba-fett-season-1-concept-art-gallery#")
	checkError(err)
	defer response.Body.Close()
	if response.StatusCode != 200 { // проверяется что ответ ОК
		log.Fatal("Received non 200 response code")
	}
	n, err := io.ReadAll(response.Body)
	checkError(err)

	linkRegexp := regexp.MustCompile("https://lumiere-a.akamaihd.net/v1/images/.{1,65}.jpeg")
	allLink := linkRegexp.FindAllString(string(n), -1)
	finalSliceLink := removeDuplicateLink(allLink)

	linkCh := make(chan string, 92)

	wg.Add(92)
	for _, val := range finalSliceLink {
		linkCh <- val
	}

	for i := 1; i <= 92; i++ {
		go downloadFile(linkCh)
	}
	wg.Wait()

}
