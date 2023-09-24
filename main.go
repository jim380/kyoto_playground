package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"

	"github.com/kyoto-framework/kyoto/v2"
)

// docs: https://pkg.go.dev/github.com/kyoto-framework/kyoto/v2#hdr-Quick_start
type StateBlock struct {
	Height string
}

type BlockInfo struct {
	Block struct {
		Header header `json:"header"`
	}
}

type header struct {
	ChainID          string `json:"chain_id"`
	Height           string `json:"height"`
	Proposer_address string `json:"proposer_address"`
	LastTimestamp    string
	Timestamp        string `json:"time"`
}

/*
Component
  - Each component is a context receiver, which returns its state
  - Each component becomes a part of the page or top-level component,
    which executes component asynchronously and gets a state future object
  - Context holds common objects like http.ResponseWriter, *http.Request, etc
*/

func GetBlockInfo(ctx *kyoto.Context) (state StateBlock) {
	RESTAddr := "https://localhost:1317"
	route := "/cosmos/base/tendermint/v1beta1/blocks/latest"

	getLatestHeight := func() string {
		var blockInfo BlockInfo
		resp, err := HttpQuery(RESTAddr + route)
		if err != nil {
			log.Printf("Failed to query HTTP: %v", err)
			return ""
		}

		json.Unmarshal(resp, &blockInfo)
		if err != nil {
			log.Printf("Failed to unmarshal response: %v", err)
			return ""
		}

		return blockInfo.Block.Header.Height
	}

	/*
		Handle Actions
			- To call an action of parent component, use $ prefix in action name
			- To call an action of component by id, use <id:action> as an action name
		    - To push multiple component UI updates during a single action call,
		        call kyoto.ActionFlush(ctx, state) to initiate an update
	*/
	handled := kyoto.Action(ctx, "Reload Latest Height", func(args ...any) {
		// add logic here
		state.Height = getLatestHeight()
		log.Println("New block height was fetched:", state.Height)
	})
	// Prevent further execution if action handled
	if handled {
		return
	}
	// Default loading behavior if not handled
	state.Height = getLatestHeight()
	// Return
	return
}

type PIndexState struct {
	LatestHeight *kyoto.ComponentF[StateBlock]
}

/*
Page
  - A page is a top-level component, which attaches components and
    defines rendering
*/
func PIndex(ctx *kyoto.Context) (state PIndexState) {
	// Define rendering
	kyoto.Template(ctx, "page.index.html")

	// Attach components
	state.LatestHeight = kyoto.Use(ctx, GetBlockInfo)

	return
}

func HttpQuery(route string) ([]byte, error) {
	req, err := http.NewRequest("GET", route, nil)
	if err != nil {
		log.Printf("Failed to create request: %v", err)
		return nil, err
	}
	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("Failed to do request: %v", err)
		return nil, err
	}
	body, err := io.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		log.Printf("Failed to read response body: %v", err)
		return nil, err
	}

	return body, err
}

func main() {
	// Register page
	kyoto.HandlePage("/", PIndex)
	// Client
	kyoto.HandleAction(GetBlockInfo)
	// Serve
	kyoto.Serve(":8080")
}
