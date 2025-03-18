package main

import (
	"fmt"
	"os"

	"github.com/projectdiscovery/yamldoc-go/encoder"
	"github.com/bio-routing/bio-rd/cmd/bio-rd/config"
)

func main() {
	FileDocs := []*encoder.FileDoc{
		config.GetconfigDoc(),
		config.GetpolicyDoc(),
		config.Getrouting_optionsDoc(),
		config.GetprotocolsDoc(),
		config.Getstatic_routeDoc(),
		config.GetbgpDoc(),
		config.GetisisDoc(),
	}

	for _, fd := range FileDocs {
		fc, err := fd.Encode()
		if err != nil {
			fmt.Printf("failed to encode the file doc: %v", err)
		}

		err = os.WriteFile("Documentation/user/config/"+fd.Name+".md", fc, 0600)
		if err != nil {
			fmt.Printf("unable to write doc file: %v", err)
		}
	}
}
