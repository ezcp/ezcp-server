package main

import (
	"crypto/tls"
	"flag"
	"log"
	"net/http"
	"os"
	"runtime/pprof"
	"time"

	"strings"

	"ezcp.io/ezcp-server/db"
	"ezcp.io/ezcp-server/routes"
	"github.com/gorilla/mux"
	"golang.org/x/crypto/acme/autocert"
)

var (
	// Tag is set by Gitlab's CI build process
	Tag string

	// Build is set by Gitlab's CI build process
	Build string

	// BitgoWallet is set by Gitlab's CI build process
	BitgoWallet string

	// BitgoToken is set by Gitlab's CI build process
	BitgoToken string
)

func main() {
	allowedHosts := []string{"ezcp.io", "www.ezcp.io", "api0.ezcp.io",
		"api1.ezcp.io", "api2.ezcp.io", "api3.ezcp.io", "api4.ezcp.io",
		"api5.ezcp.io", "api6.ezcp.io", "api7.ezcp.io", "api8.ezcp.io",
		"api9.ezcp.io", "apia.ezcp.io", "apib.ezcp.io", "apic.ezcp.io",
		"apid.ezcp.io", "apie.ezcp.io", "apif.ezcp.io"}

	if Build != "" {
		if Tag == "" {
			log.Printf("ezcp-server build %s", Build)
		} else {
			log.Printf("ezcp-server %s - build %s", Tag, Build)
		}
	} else {
		log.Print("ezcp-server development version")
	}

	log.SetFlags(log.LUTC | log.LstdFlags)

	var purge = flag.Bool("purge", false, "purge old tokens")
	var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")
	var memprofile = flag.String("memprofile", "", "write memory profile to this file")
	var dbHost = flag.String("db", "localhost:27017", "host:port,host:port of mongodb servers")

	var ssl = flag.Bool("ssl", false, "Enable SSL support")
	flag.Parse()

	if envSSL := os.Getenv("SSL"); envSSL != "" {
		*ssl = true
	}

	if *cpuprofile != "" {
		after2min := time.After(time.Minute * 2)

		log.Print("Setting up CPU Prof")
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		go func() {
			<-after2min
			log.Print("Starting CPU prof output")
			pprof.StopCPUProfile()
			log.Print("Done CPU prof output")
		}()

	}
	if *memprofile != "" {
		after2min := time.After(time.Minute * 1)

		f, err := os.Create(*memprofile)
		if err != nil {
			log.Fatal(err)
		}

		go func() {
			<-after2min
			pprof.WriteHeapProfile(f)
			f.Close()
		}()
	}

	database, err := db.NewDB(*dbHost, db.BitgoToken(BitgoToken), BitgoWallet)
	if err != nil {
		panic(err)
	}
	defer database.Close()

	handler := routes.NewHandler(database)

	if *purge {
		log.Print("Purging old ezcp tokens")

		tokens, err := database.RemoveExpiredTokens()
		if err != nil {
			panic(err)
		}

		for _, eachToken := range tokens {
			path := db.GetFilePath(eachToken)
			_, err := os.Stat(path)
			if err != os.ErrNotExist {
				os.Remove(path)
			}
		}
		log.Print("Purging... done.")
		return
	}

	r := mux.NewRouter()
	r.HandleFunc("/upload/{token}", handler.Upload)
	r.HandleFunc("/download/{token}", handler.Download)

	r.HandleFunc("/bitcoin", handler.Bitcoin)
	r.HandleFunc("/token/{tx}", handler.GetTokenTx)

	r.HandleFunc("/linux", handler.DownloadOS("linux", "ezcp"))
	r.HandleFunc("/osx", handler.DownloadOS("darwin", "ezcp"))
	r.HandleFunc("/windows", handler.DownloadOS("windows", "ezcp.exe"))

	r.HandleFunc("/", handler.Root)

	http.Handle("/", Gzip(r))

	if *ssl {
		log.Print("Starting HTTPS server, on 0.0.0.0:443")
		go startSSL(allowedHosts, database)

		log.Print("Starting HTTP server, on 0.0.0.0:80")
		log.Fatal(http.ListenAndServe(":80", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			host := req.Host
			if last := strings.LastIndex(req.Host, ":"); last != -1 {
				host = req.Host[:last]
			}
			target := "https://" + host + req.URL.Path
			if len(req.URL.RawQuery) > 0 {
				target += "?" + req.URL.RawQuery
			}
			w.Header().Set("Strict-Transport-Security", `"max-age=31536000; includeSubDomains; preload"`)
			http.Redirect(w, req, target, http.StatusMovedPermanently)
		})))
		return
	}

	log.Print("Starting development server, on localhost:8000")
	srv := &http.Server{
		Addr: "localhost:8000",
	}
	log.Fatal(srv.ListenAndServe())
}

func startSSL(allowedHosts []string, db *db.DB) {
	certManager := autocert.Manager{
		Prompt:     autocert.AcceptTOS,
		HostPolicy: autocert.HostWhitelist(allowedHosts...),
		Cache:      db,
		Email:      "info@ezcp.io",
		ForceRSA:   false,
	}
	srv := &http.Server{
		Addr: ":https",
		TLSConfig: &tls.Config{
			GetCertificate: certManager.GetCertificate,
		},
	}
	log.Fatal(srv.ListenAndServeTLS("", ""))
}
