package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/MarcusXavierr/xo-battle-back/internal/room"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
)

var upgrader = websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}

func main() {
	godotenv.Load()
	r := chi.NewRouter()
	allowedOrigins := strings.Split(os.Getenv("ALLOWED_ORIGINS"), ",")
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins: allowedOrigins,
		AllowedMethods: []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders: []string{"Accept", "Authorization", "Content-Type"},
	}))
	roomManager := room.NewRoomManager()
	go roomManager.RoomDeleter()
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("The system is alive"))
	})

	r.Post("/room/{id}", func(w http.ResponseWriter, r *http.Request) {
		name := chi.URLParam(r, "id")
		if name == "" {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("You must provide a room name"))
			return
		}
		if err := roomManager.CreateRoom(name); err != nil {
			w.WriteHeader(http.StatusConflict)
			w.Write([]byte("The room '" + name + "' already exists dipshit"))
			return
		}
		w.Write([]byte("Created room: " + name))
	})

	r.Get("/room/{id}/join", func(w http.ResponseWriter, r *http.Request) {
		roomName := chi.URLParam(r, "id")
		name := r.URL.Query().Get("name")
		kind := r.URL.Query().Get("player_type")

		if name == "" {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Missing name"))
			return
		}

		if lower := strings.ToLower(kind); lower != "" && lower != "x" && lower != "o" {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Invalid player type"))
			return
		}

		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Printf("Upgrade failed: %v\n", err)
			return
		}

		player := room.NewPlayer(conn, name)
		if err := roomManager.JoinRoom(roomName, player, kind); err != nil {
			log.Printf("Error joining room: %v", err)
			conn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("{\"type\": \"error\"} \"reason\": \"%s\"", err)))
			conn.Close()
			return
		}
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8888"
	}

	server := http.Server{
		Addr:    ":" + port,
		Handler: r,
	}

	go func() {
		log.Printf("Starting HTTP server on port %s", port)
		if err := server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("HTTP server closed with error: %v", err)
		}
		log.Println("Stopped serving HTTP connections")
	}()

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM)
	hold(sc)

	ctx, shutdownRelease := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownRelease()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatal("Server Shutdown:", err)
	}

	log.Println("Gracefully shut down the server")
}

func hold(sc chan os.Signal) {
	tries := 0
	slur := "Shut the fuck up bitch"
	for {
		value := <-sc
		if value == syscall.SIGTERM {
			break
		}

		if tries >= 3 {
			log.Println("Okay fuck it, I'm going to shut the fuck down")
			break
		}

		if tries >= 1 {
			log.Println(slur + " (" + strconv.Itoa(tries) + ")")
		} else {
			log.Println(slur)
		}

		tries++
	}
}
