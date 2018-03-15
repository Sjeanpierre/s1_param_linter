package main

import (
	"github.com/google/go-github/github"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"os"
	"regexp"
)

type Response struct {
	Status  string
	Message string
}

type CommitSet struct {
	RepoName  string
	RepoOwner string
	CommitID  string
	ParamPaths
	Clean bool
}

type ParamPath struct {
	Product     string
	Landscape   string
	Environment string
	FileName    string
}

type ParamPaths []ParamPath

func main() {
	router := mux.NewRouter().StrictSlash(true)
	router.NotFoundHandler = http.HandlerFunc(notFoundHandler)
	router.HandleFunc("/event/", EventHandler).Methods("POST")
	loggedRouter := handlers.LoggingHandler(os.Stdout, router)
	log.Println("Started: Ready to serve")
	log.Fatal(http.ListenAndServe(":4545", loggedRouter))
}

func EventDetails(evt github.PushEvent) {
	ChangeSets := []CommitSet{}
	emptyParamPath := ParamPath{}
	for _, commit := range evt.Commits {
		filesModified := ParamPaths{}
		for _, fileName := range commit.Modified {
			paths := parseServicePath(fileName)
			if paths == emptyParamPath {
				continue
			}
			filesModified = append(filesModified, paths)
		}
		for _, fileName := range commit.Added {
			paths := parseServicePath(fileName)
			if paths == emptyParamPath {
				continue
			}
			filesModified = append(filesModified, paths)
		}
		modifications := CommitSet{
			RepoName:   *evt.Repo.Name,
			RepoOwner:  *evt.Repo.Owner.Name,
			CommitID:   *commit.ID,
			ParamPaths: filesModified,
		}
		checkCompliance(&modifications)
		ChangeSets = append(ChangeSets, modifications)
	}
	log.Printf("Changes %+v", ChangeSets)
	PostStatus(ChangeSets)
}

func checkCompliance(commit *CommitSet) {
	//ensure only one app is being modified in commit
	//multiple environments for the same app can be modified within commit
	//multiple apps can be modified in single, push, but different commits are needed
	//this will allow the further automation to target a single application at a time
	products := []string{}
	for _, path := range commit.ParamPaths {
		products = append(products, path.Product)
	}
	p := dedup(products)
	if len(p) > 1 {
		commit.Clean = false
		return
	}
	commit.Clean = true
	return
}

func dedup(Slice []string) []string {
	var returnSlice []string
	for _, value := range Slice {
		if !contains(returnSlice, value) {
			returnSlice = append(returnSlice, value)
		}
	}
	return returnSlice
}

func contains(Slice []string, searchVal string) bool {
	for _, value := range Slice {
		if value == searchVal {
			return true
		}
	}
	return false
}

//tested at https://play.golang.org/p/4XodrfKefXK
func parseServicePath(path string) (cp ParamPath) {
	re := regexp.MustCompile(`(?P<product>\w*)/(?P<environment>\w*)/(?P<landscape>\w*)/(?P<file_name>\w*).yml`)
	fieldNames := re.SubexpNames()
	capturedGroups := re.FindAllStringSubmatch(path, -1)
	if capturedGroups == nil {
		return cp
	}
	captured := capturedGroups[0]

	md := map[string]string{}
	for i, capturedGroup := range captured {
		md[fieldNames[i]] = capturedGroup
	}
	cp = ParamPath{Product: md["product"], Landscape: md["landscape"], Environment: md["environment"], FileName: md["file_name"]}
	return
}
