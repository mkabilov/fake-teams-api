package main

import (
	"encoding/json"
	"fmt"
	"hash/crc32"
	"io"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
)

const (
	membersCount = 10
	emailDomain  = "example.com"
)

var (
	nameFirstLetters = []byte("abcdefghijklmnopqrstuvwz")
	lastNames        = []string{
		"frerichs", "hengst", "kloth", "orth", "popp", "steigerwald", "kley", "voegele", "ostermeier",
		"schley", "eckel", "kaeser", "holzner", "klostermann", "ostertag", "weishaupt", "gotte", "feiler", "mast",
		"hielscher", "haertel", "moeller", "kramer", "stief", "kissel", "gottschlich", "drechsler", "lucas", "till",
		"koehne", "pfitzner", "sydow", "liedtke", "franz", "flick", "menz", "hertel", "kreuzer", "kleinert", "zeidler",
	}
)

type myHandler struct{}

type InfrastructureAccount struct {
	Id          string `json:"id"`
	Name        string `json:"name"`
	Provider    string `json:"provider"`
	Type        string `json:"type"`
	Description string `json:"description"`
	Owner       string `json:"owner"`
	OwnerDn     string `json:"owner_dn"`
	Disabled    bool   `json:"disabled"`
}

type Team struct {
	Dn           string   `json:"dn"`
	Id           string   `json:"id"`
	TeamName     string   `json:"id_name"`
	TeamId       string   `json:"team_id"`
	Type         string   `json:"type"`
	FullName     string   `json:"name"`
	Aliases      []string `json:"alias,omitempty"`
	Mails        []string `json:"mail"`
	Members      []string `json:"member"`
	CostCenter   string   `json:"cost_center,omitempty"`
	DeliveryLead string   `json:"delivery_lead,omitempty"`
	ParentTeamId string   `json:"parent_team_id,omitempty"`

	InfrastructureAccounts []InfrastructureAccount `json:"infrastructure-accounts"`
}

func main() {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt, syscall.SIGTERM) // Push signals into channel

	if len(os.Args) < 2 {
		fmt.Printf("Usage %s {port}\n", os.Args[0])
		os.Exit(1)
	}

	port, err := strconv.Atoi(os.Args[1])
	if err != nil {
		fmt.Printf("Port must be numeric\n")
		os.Exit(1)
	}

	server := http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: &myHandler{},
	}

	go func() {
		err = server.ListenAndServe()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}()

	fmt.Printf("Listening on %d\n", port)
	sig := <-sigs
	fmt.Printf("Got %+v signal. Shutting down\n", sig)
}

func members() []string {
	res := make([]string, membersCount)
	for i := 0; i < membersCount; i++ {
		name := nameFirstLetters[rand.Intn(len(nameFirstLetters))]
		res[i] = string(name) + lastNames[rand.Intn(len(lastNames))]
	}

	return res
}

func testTeam(teamName string, w http.ResponseWriter) {
	id := crc32.ChecksumIEEE([]byte(teamName))
	rand.Seed(int64(id))

	teamType := "official"
	if id%2 == 0 {
		teamType = "virtual"
	}

	team := Team{
		Id:       teamName,
		TeamId:   strconv.Itoa(int(id)),
		TeamName: strings.ToUpper(teamName),
		Members:  members(),
		FullName: strings.Title(teamName),
		Mails:    []string{fmt.Sprintf("%s@%s", teamName, emailDomain)},
		Type:     teamType,
	}

	m, err := json.Marshal(team)
	if err != nil {
		io.WriteString(w, fmt.Sprintf("Can't marshal: %s", err))
	}
	w.Header().Set("Content-Type", "application/json")

	w.Write(m)
}

func badRequest(w http.ResponseWriter) {
	w.WriteHeader(http.StatusBadRequest)
	w.Write([]byte("Bad request"))
}

func (*myHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("Request: %s\n", r.URL.Path)
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
