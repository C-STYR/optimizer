package main

import (
	"bytes"
	"database/sql"
	"image"
	"log"
	"math/rand"
	"net/http"
	"os"
	"path"
	"time"

	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"

	"github.com/C-STYR/optimizer/optimizerdb"
	"github.com/chai2010/webp"

	_ "github.com/mattn/go-sqlite3"

)

type WebpImage = bytes.Buffer

// creates slice of filenames from directory
func getImageNames(path string) []os.DirEntry {
	entries, err := os.ReadDir(path)
	if err != nil {
		log.Fatal(err)
	}
	return entries
}

// reads raw img file and outputs a slice of bytes
func loadImg(path string) []byte {
	data, err := os.ReadFile(path)
	if err != nil {
		log.Fatal(err)
	}
	return data
}

/* 
	parseImg() converts a slice of bytes into an image.Image
	bytes.NewReader() returns a *Reader
	image.Decode() decodes an image that has been encoded in a registered format
	it takes an io.Reader and returns Image, string, error (string is format name)
*/
func parseImg(rawImg []byte) image.Image {
	img, _, err := image.Decode(bytes.NewReader(rawImg))
	if err != nil {
		log.Fatal(err)
	}
	return img
}

// converts an image to WebP format
func imgToWebp(img image.Image) WebpImage {
	var buf bytes.Buffer
	if err := webp.Encode(&buf, img, &webp.Options{Lossless: true}); err != nil {
		log.Fatal(err)
	}
	return buf
}

func getImage()(string, WebpImage) {
	imgBase := "data"

	imgName := getImageNames(imgBase)

	index := rand.Intn(len(imgName))

	imgEntry := imgName[index]

	imgPath := path.Join(imgBase, imgEntry.Name())

	rawImg := loadImg(imgPath)

	parsedImg := parseImg(rawImg)

	webpImg := imgToWebp(parsedImg)

	return imgEntry.Name(), webpImg
}

func handleRoot(db *sql.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.Method != "GET" {
			return
		}

		name, img := getImage()

		optimizerdb.IncrementHitCount(db, name)

		log.Println("Serving", name)
		
		// writes the byte slice to the http reply
		w.Write(img.Bytes())

		return
	})
}

func main() {
	rand.Seed(time.Now().UnixNano()) //creates a unique seed

	log.Println("Using database 'optimizer.db'")
	db, err := sql.Open("sqlite3", "optimizer.db")
	if err != nil {
		log.Fatal(err)
	}

	optimizerdb.TryCreate(db)

	http.Handle("/", handleRoot(db))

	log.Println("Listening on localhost:8099")
	err = http.ListenAndServe("localhost:8099", nil)
	if err != nil {
		log.Fatalln(err)
	}
}