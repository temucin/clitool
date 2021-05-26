package elastic

import (
	"clitool/utils/CmdRegistry"
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/elastic/go-elasticsearch/v8"
)

var ElasticCmd = CmdRegistry.Cmd{
	Name:    "elastic",
	RunCmd:  runCmd,
	FlagSet: ElasticFlagSet,
}

type HitSource struct {
	Target string `json:"member"`
}

func (h *HitSource) ToSlice() []string {
	ret := []string{}
	ret = append(ret, h.Target)
	return ret
}

type envelopeResponse struct {
	Took     int
	ScrollID string `json:"_scroll_id"`
	Hits     struct {
		Total struct {
			Value int
		}
		Hits []struct {
			ID         string          `json:"_id"`
			Source     HitSource       `json:"_source"`
			Highlights json.RawMessage `json:"highlight"`
			Sort       []interface{}   `json:"sort"`
		}
	}
}

var clusters map[string]string
var env string
var ElasticFlagSet flag.FlagSet
var index string

var sitClusters = map[string]string{
	"example": "https://example.us-east-1.es.amazonaws.com",
}

const (
	cmdUsage     = "Queries the specified elastic search cluster for data from targeted transactions"
	envUsage     = "Specifes which environment clusters to query."
	indexUsage   = "Specifies the index in the elasticSearch cluster from which to query"
	indexDefault = "default-index"
	query        = `{
			"query": {
			  "bool": {
				"filter": [
				  {
					"exists": {
					  "field": "target"
					}
				  },
				  {
					"exists": {
					  "field": "log_target"
					}
				  }
				]
			  }
			}
		  }`
)

func init() {
	ElasticFlagSet = *flag.NewFlagSet("elastic", flag.ContinueOnError)
	ElasticFlagSet.Usage = func() { fmt.Print(cmdUsage) }
	ElasticFlagSet.StringVar(&env, "env", "envSample", envUsage)
	ElasticFlagSet.StringVar(&env, "e", "envSample", "shortcut for environment")
	ElasticFlagSet.StringVar(&index, "index", indexDefault, indexUsage)
	CmdRegistry.RegisterCmd(ElasticCmd)
	CmdRegistry.RegisterFlagSet(ElasticFlagSet)
}

func cleanUp() {
	clusters = map[string]string{}
	env = ""
}

func validateFlagsAndArgs() int {
	if env == "envSample" {
		fmt.Println("Environment set to sit")
		clusters = sitClusters
	} else {
		fmt.Println("Error. Environment not set to valid value")
		return 1
	}

	fmt.Println("Using index", index)

	return 0
}

//RunCmd is the entrypoint into the elastic command execution
func runCmd() {
	ElasticFlagSet.Parse(CmdRegistry.CmdArgs())
	if validateFlagsAndArgs() != 0 {
		return
	}

	var wg *sync.WaitGroup = new(sync.WaitGroup)
	wg.Add(len(clusters))
	for i, v := range clusters {
		go executeQuery(i, v, wg)
	}
	wg.Wait()

}

func executeQuery(member string, clusterAddress string, wg *sync.WaitGroup) {
	start := time.Now()
	fmt.Println("Querying cluster", member)

	//Configure elasticsearch go client
	cfg := elasticsearch.Config{
		Addresses: []string{
			clusterAddress,
		},
	}
	es, err := elasticsearch.NewClient(cfg)
	if err != nil {
		log.Fatalf("Error creating the client: %s", err)
		return
	}

	//Build filter query for search command
	var b strings.Builder
	b.WriteString(query)
	read := strings.NewReader(b.String())

	//Search for 1000 rows and then set the scroll
	res, err := es.Search(
		es.Search.WithIndex(index),
		es.Search.WithSize(1000),
		es.Search.WithBody(read),
		es.Search.WithScroll(time.Minute),
		es.Search.WithPretty(),
	)
	if err != nil {
		log.Fatalf("Error getting data from index: %s", err)
	}
	// fmt.Println(res)

	//Decode initial results
	var r envelopeResponse
	if err := json.NewDecoder(res.Body).Decode(&r); err != nil {
		log.Fatalf("Error getting data from index: %s", err)
	}
	defer res.Body.Close()

	file, err := os.Create("output.csv")
	if err != nil {
		fmt.Println("Error creating output.csv", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()
	writer.Write([]string{"target"})

	//Write initial results
	for _, data := range r.Hits.Hits {
		// fmt.Println(data.Source.ToSlice())
		writer.Write(data.Source.ToSlice())
		// writeToFile(writer, data.Source.ToSlice())
	}

	//Scroll over remaining rows until there are none left
	scrollID := r.ScrollID
	for {
		//Execute scroll
		scroll_res, err := es.Scroll(
			es.Scroll.WithScroll(time.Minute),
			es.Scroll.WithScrollID(scrollID),
			es.Scroll.WithPretty(),
		)
		if err != nil {
			fmt.Println("Error executing scroll!", err)
		}
		if scroll_res.IsError() {
			log.Fatalf("Error with scroll! %s", scroll_res)
		}
		// fmt.Println(scroll_res)

		//Decode initial results
		var rs envelopeResponse
		if err := json.NewDecoder(scroll_res.Body).Decode(&rs); err != nil {
			log.Fatalf("Error reolving data: %s", err)
		}

		//End scrolling
		if len(rs.Hits.Hits) < 1 {
			fmt.Printf("\nQuery for %s completed. Time taken: %v", member, time.Since(start))
			break
		}

		//Write results
		for _, data := range rs.Hits.Hits {
			// fmt.Println(data.Source.ToSlice())
			writer.Write(data.Source.ToSlice())
			// writeToFile(writer, data.Source.ToSlice())
		}
		scrollID = rs.ScrollID
	}

	wg.Done()
}

var mu sync.Mutex

func writeToFile(writer *csv.Writer, data []string) {
	mu.Lock()
	defer mu.Unlock()
	writer.Write(data)
}
