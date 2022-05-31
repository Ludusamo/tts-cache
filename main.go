package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/julienschmidt/httprouter"
	"github.com/mailgun/groupcache/v2"
)

func getTTS(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	var res []byte
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*500)
	defer cancel()
	group := groupcache.GetGroup("tts")
	if group == nil {
		http.Error(w, "group tts not found", http.StatusNotFound)
	}
	err := group.Get(ctx, ps.ByName("id"), groupcache.AllocatingByteSliceSink(&res))
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	w.Write(res)
	w.Write([]byte{'\n'})
}

func main() {
	host := os.Getenv("HOST")
	peers := os.Getenv("PEERS")
	cacheSize, _ := strconv.Atoi(os.Getenv("CACHE_SIZE"))
	cacheTime, _ := strconv.Atoi(os.Getenv("CACHE_TIME"))
	flag.Parse()

	p := strings.Split(peers, ",")
	pool := groupcache.NewHTTPPool(fmt.Sprintf("http://%s:8000", host))
	pool.Set(p...)
	server := http.Server{
		Addr:    fmt.Sprintf(":8000"),
		Handler: pool,
	}
	go func() {
		log.Println("Serving...")
		if err := server.ListenAndServe(); err != nil {
			log.Fatal(err)
		}
	}()

	log.Println("IP:", host)

	groupcache.NewGroup("tts", int64(cacheSize), groupcache.GetterFunc(
		func(ctx context.Context, id string, dest groupcache.Sink) error {
			log.Println("getting id", id)
			return dest.SetBytes([]byte(id), time.Now().Add(time.Minute*time.Duration(cacheTime)))
		},
	))

	router := httprouter.New()
	router.GET("/tts/:id", getTTS)
	log.Fatal(http.ListenAndServe(":80", router))
}
