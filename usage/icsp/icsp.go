package main

import (
	"fmt"
	"log"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/icsp"
)

func main() {
	apiClient := clients.New("")

	var icspBuilderName = "testicsp"
	// Define new ICSPBuilder
	icspbuilder := icsp.NewICSPBuilder(apiClient, icspBuilderName, "mysource", []string{"mymirror.io"})
	// Delete ImageContentSourcePolicy if exists. Delete succeeds when the policy is not found, so there is no need
	// to check existence first, which would also skip the reset on transient Get errors.
	if err := icspbuilder.Delete(); err != nil {
		log.Fatal(err)
	}

	// Create new ImageContentSourcePolicy from icspbuilder.
	_, err := icspbuilder.Create()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Initial ImageContentSourcePolicy: ", icspbuilder.Object.Spec.RepositoryDigestMirrors)

	// Add new source to the ImageContentSourcePolicy definition and update the object
	_, err = icspbuilder.WithRepositoryDigestMirror(
		"newsource",
		[]string{"Mirror1.io", "Mirror2.io", "Mirror3.io"}).Update()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Updated ImageContentSourcePolicy: ", icspbuilder.Object.Spec.RepositoryDigestMirrors)
	// Delete ImageContentSourcePolicy after demonstration
	err = icspbuilder.Delete()
	if err != nil {
		log.Fatal(err)
	}
}
