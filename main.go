package main

import (
	"crypto/tls"
	"flag"
	"log"
	"net/http"
	"runtime/pprof"
	"time"

	"ezcp.io/ezcp-server/routes"
	"golang.org/x/crypto/acme/autocert"

	"os"

	"ezcp.io/ezcp-server/db"
	"github.com/gorilla/mux"
	"github.com/rs/cors"
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

	// AllowedOrigin for GetToken
	AllowedOrigin string
)

func main() {
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

	var hostPort = flag.String("host", "localhost:8000", "host and port for http server")
	var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")
	var memprofile = flag.String("memprofile", "", "write memory profile to this file")
	var dbHost = flag.String("db", "localhost:27017", "host:port,host:port of mongodb servers")

	var ssl = flag.Bool("ssl", false, "Enable SSL support")
	flag.Parse()

	if hp := os.Getenv("HOST_PORT"); hp != "" {
		hostPort = &hp
	}

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

	os.Mkdir(routes.EZCPstorage, 0700) // it doesn't matter if it exists already

	r := mux.NewRouter()

	db, err := db.NewDB(*dbHost, db.BitgoToken(BitgoToken), BitgoWallet)
	if err != nil {
		panic(err)
	}
	defer db.Close()
	handler := routes.NewHandler(db, AllowedOrigin)

	r.HandleFunc("/token", handler.GetToken)
	r.HandleFunc("/token/{tx}", handler.GetTokenTx)
	r.HandleFunc("/upload/{token}", handler.Upload)
	r.HandleFunc("/download/{token}", handler.Download)
	r.HandleFunc("/bitcoin", handler.Bitcoin)

	withCors := cors.Default().Handler(r) // TODO LATER finer handling of allowed origins
	http.Handle("/", withCors)

	if *ssl {
		certManager := autocert.Manager{
			Prompt:     autocert.AcceptTOS,
			HostPolicy: autocert.HostWhitelist("ezcp.io"),
			Cache:      autocert.DirCache("certs"), // TODO LATER store in bold db or distributed db ?
			Email:      "chris@chris-hartwig.com",  // TODO env var
			ForceRSA:   false,
		}
		srv := &http.Server{
			Addr: ":https",
			TLSConfig: &tls.Config{
				GetCertificate: certManager.GetCertificate,
			},
		}
		log.Print("TLS Server starting at ", *hostPort)

		log.Fatal(srv.ListenAndServeTLS("", ""))
		return
	}
	log.Print("Server starting at ", *hostPort)

	srv := &http.Server{
		Addr: *hostPort,
	}
	log.Fatal(srv.ListenAndServe())
}
