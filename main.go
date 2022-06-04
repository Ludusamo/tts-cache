package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/julienschmidt/httprouter"
	"github.com/mailgun/groupcache/v2"
)

func getTTS(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	decoder := json.NewDecoder(r.Body)
	var params TTSParams
	err := decoder.Decode(&params)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	var res []byte
	ctx := context.Background()
	group := groupcache.GetGroup("tts")
	if group == nil {
		http.Error(w, "group tts not found", http.StatusNotFound)
		return
	}
	key := params.GetKey()
	log.Println("fetching key:", key)
	err = group.Get(ctx, key, groupcache.AllocatingByteSliceSink(&res))
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	w.Write(res)
}

func main() {
	host := os.Getenv("HOST")
	peers := os.Getenv("PEERS")
	cacheSize, _ := strconv.Atoi(os.Getenv("CACHE_SIZE"))
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

	groupcache.NewGroup("tts", int64(cacheSize), groupcache.GetterFunc(TTSGetterFunc))

	router := httprouter.New()
	router.POST("/api/tts", getTTS)
	log.Fatal(http.ListenAndServe(":80", router))
}
