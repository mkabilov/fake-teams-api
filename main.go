package main

import (
	"encoding/json"
	"fmt"
	"hash/crc32"
	"io"
	"math/rand"
	"net/http"
	"strconv"
	"strings"

	"github.com/ikitiki/postgres-operator/pkg/util/teams"
)

var (
	NameFirstLetters = []byte("abcdefghijklmnopqrstuvwz")
	LastNames        = []string{
		"frerichs", "hengst", "kloth", "orth", "popp", "steigerwald", "kley", "voegele", "ostermeier",
		"schley", "eckel", "kaeser", "holzner", "klostermann", "ostertag", "weishaupt", "gotte", "feiler", "mast",
		"hielscher", "haertel", "moeller", "kramer", "stief", "kissel", "gottschlich", "drechsler", "lucas", "till",
		"koehne", "pfitzner", "sydow", "liedtke", "franz", "flick", "menz", "hertel", "kreuzer", "kleinert", "zeidler",
	}
)

type myHandler struct{}

func main() {
	server := http.Server{
		Addr:    ":8000",
		Handler: &myHandler{},
	}

	server.ListenAndServe()
}

func members() []string {
	res := make([]string, 0)
	for i := 0; i < 10; i++ {
		name := NameFirstLetters[rand.Intn(len(NameFirstLetters))]
		res = append(res, string(name)+LastNames[rand.Intn(len(LastNames))])
	}

	return res
}

func testTeam(teamName string, w http.ResponseWriter) {
	id := crc32.ChecksumIEEE([]byte(teamName))
	rand.Seed(int64(id))

	team := teams.Team{
		Id:       teamName,
		TeamId:   strconv.Itoa(int(id)),
		TeamName: strings.ToUpper(teamName),
		Members:  members(),
		FullName: strings.Title(teamName),
		Mails:    []string{fmt.Sprintf("%s@example.com", teamName)},
	}

	m, err := json.Marshal(team)
	if err != nil {
		io.WriteString(w, fmt.Sprintf("Can't marshal: %s", err))
	}
	w.Header().Set("Content-Type", "application/json")

	w.Write(m)
}

func badRequest(w http.ResponseWriter) {
	w.WriteHeader(400)
	w.Write([]byte("Bad request"))
}

func (*myHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !strings.HasPrefix(r.URL.Path, "/teams") {
		badRequest(w)
		return
	}
	var teamName string

	parts := strings.Split(r.URL.Path, "/")
	if len(parts) > 1 {
		teamName = parts[2]
	}

	if teamName == "" {
		badRequest(w)
		return
	}

	testTeam(teamName, w)
}
