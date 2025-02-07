package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/owulveryck/gomcptest/host/openaiserver/chatengine"
	"github.com/owulveryck/gomcptest/host/openaiserver/chatengine/gcp"
)

func main() {
	fmt.Println("Server starting on port 8080")
	openAIHandler := chatengine.NewOpenAIV1WithToolHandler(gcp.NewChatSession())
	log.Fatal(http.ListenAndServe(":8080", openAIHandler))
}
